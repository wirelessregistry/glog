package glog

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/DataDog/datadog-go/statsd"
)

var (
	logInterval int
	cMap        *counterMap
	pcMap       *counterMap
	ddCli       *statsd.Client
	ddAgent     string
	ddNamespace string
	ddTags      string
	printStack  bool
	mLogOnce    sync.Once
	hLogOnce    sync.Once
	ddOnce      sync.Once
)

const (
	separator = ";"
)

type counterMap struct {
	m  map[string]*dataCounter
	ll sync.RWMutex
}

func newCounterMap() *counterMap {
	return &counterMap{
		m: make(map[string]*dataCounter),
	}
}

func initAPM() {
	flag.IntVar(&logInterval, "logint", 60, "histogram log interval in seconds")
	flag.StringVar(&ddAgent, "ddAgent", "127.0.0.1:8125", "address of the datadog statsd agent as [HOST]:[PORT]")
	flag.StringVar(&ddNamespace, "ddNamespace", "wirelessregistry", "namespace to send metrics with")
	flag.StringVar(&ddTags, "ddTags", "", "tags to send metrics with")
	flag.BoolVar(&printStack, "printStack", false, "print the stack")

}

// SetLogInterval sets the log interval for histogram and memory consumption logging.
// Interval is provided as seconds. Defaults to 60s
func SetLogInterval(l int) {
	logInterval = l
}

// LogTimeTaken logs the elapsed from the time provided to the time the function is called.
func LogTimeTaken(name string, start time.Time) {
	elapsed := time.Since(start)
	Infof("%s took %s", name, elapsed)
}

// StartHistogramLogging launches the periodic logging of the counters.
// All counters are reset after reading their value.
func StartHistogramLogging() {
	hLogOnce.Do(func() {
		cMap = newCounterMap()
		pcMap = newCounterMap()
		initPeriodicLogger(time.Duration(logInterval)*time.Second, writeHistogram)
		initPeriodicLogger(time.Duration(logInterval)*time.Second, writePersistentCounter)
		Infof("printstack %v", printStack)
		if printStack {
			Info("Init printstack")
			initPeriodicLogger(time.Duration(logInterval)*time.Second, writeStack)
		}
	})
}

// StartDatadog initializes the connection to the datadog agent and launces the periodic send of metrics to datadog
func StartDatadog(agent string) {
	ddOnce.Do(func() {
		if agent != "" {
			ddAgent = agent
		}

		Infof("Initializing datadog agent at %s\n", ddAgent)
		c, err := statsd.New(ddAgent)
		if err != nil {
			Error(err)
		}
		// prefix every metric with the app name
		c.Namespace = ddNamespace + "."
		// send the EC2 availability zone as a tag with every metric
		c.Tags = append(c.Tags, ddTags)
		c.Tags = append(c.Tags, fmt.Sprintf("binary:%s", defaultBinaryName()))
		hostname, _ := os.Hostname()
		c.Tags = append(c.Tags, fmt.Sprintf("hostname:%s", hostname))
		Infof("Datadog agent ready with namespace '%s' and tags '%s'\n", c.Namespace, c.Tags)
		ddCli = c
	})
}

func defaultBinaryName() string {
	bn := os.Args[0]
	// cut path information from executable
	i := strings.LastIndex(bn, "/")
	if i > -1 {
		bn = bn[i+1:]
	}
	return bn
}

func initPeriodicLogger(t time.Duration, fn func()) chan struct{} {
	closeCh := make(chan struct{}, 1)

	go func() {
		t := time.NewTicker(t)
		for {
			select {
			case <-closeCh:
				t.Stop()
				return
			case <-t.C:
				fn()
			}
		}
	}()

	return closeCh
}

type dataCounter struct {
	i *int64
}

// IncCounter increments the referenced counter by the given value or creates it if it does not exist yet.
func IncCounter(name string, value int64) {
	IncTaggedCounter(name, []string{}, value)
}

// IncCounter increments the referenced tagged counter by the given value or creates it if it does not exist yet.
func IncTaggedCounter(name string, tags []string, value int64) {
	if cMap == nil {
		Errorf("IncCounter() called before StartHistogramLogging()")
		return
	}
	modifyCounter(cMap, name, tags, value, incCounterValue)
}

// DecCounter decrements the referenced counter by the given value or creates it if it does not exist yet.
func DecCounter(name string, value int64) {
	DecTaggedCounter(name, []string{}, value)
}

// DecCounter decrements the referenced tagged counter by the given value or creates it if it does not exist yet.
func DecTaggedCounter(name string, tags []string, value int64) {
	if cMap == nil {
		Errorf("DecCounter() called before StartHistogramLogging()")
		return
	}
	modifyCounter(cMap, name, tags, value, decCounterValue)
}

/// SetCounter sets the referenced counter to the given value or creates it if it does not exist yet.
func SetCounter(name string, value int64) {
	SetTaggedCounter(name, []string{}, value)
}

// SetCounter sets the referenced tagged counter to the given value or creates it if it does not exist yet.
func SetTaggedCounter(name string, tags []string, value int64) {
	if cMap == nil {
		Errorf("SetCounter() called before StartHistogramLogging()")
		return
	}
	modifyCounter(cMap, name, tags, value, setCounterValue)
}

// IncCounter increments the referenced counter by the given value or creates it if it does not exist yet.
func IncPCounter(name string, value int64) {
	IncTaggedPCounter(name, []string{}, value)
}

// IncCounter increments the referenced tagged counter by the given value or creates it if it does not exist yet.
func IncTaggedPCounter(name string, tags []string, value int64) {
	if pcMap == nil {
		Errorf("IncCounter() called before StartHistogramLogging()")
		return
	}
	modifyCounter(pcMap, name, tags, value, incCounterValue)
}

// DecCounter decrements the referenced counter by the given value or creates it if it does not exist yet.
func DecPCounter(name string, value int64) {
	DecTaggedPCounter(name, []string{}, value)
}

// DecCounter decrements the referenced tagged counter by the given value or creates it if it does not exist yet.
func DecTaggedPCounter(name string, tags []string, value int64) {
	if pcMap == nil {
		Errorf("DecCounter() called before StartHistogramLogging()")
		return
	}
	modifyCounter(pcMap, name, tags, value, decCounterValue)
}

/// SetCounter sets the referenced counter to the given value or creates it if it does not exist yet.
func SetPCounter(name string, value int64) {
	SetTaggedPCounter(name, []string{}, value)
}

// SetCounter sets the referenced tagged counter to the given value or creates it if it does not exist yet.
func SetTaggedPCounter(name string, tags []string, value int64) {
	if cMap == nil {
		Errorf("SetCounter() called before StartHistogramLogging()")
		return
	}
	modifyCounter(pcMap, name, tags, value, setCounterValue)
}

func setCounterValue(c *dataCounter, value int64) {
	c.set(value)
}

func incCounterValue(c *dataCounter, value int64) {
	c.increment(value)
}

func decCounterValue(c *dataCounter, value int64) {
	c.decrement(value)
}

func modifyCounter(cm *counterMap, name string, tags []string, value int64, modify func(*dataCounter, int64)) {
	key := encodeCounterKey(name, tags)
	cm.ll.RLock()
	_, ok := cm.m[key]
	if !ok {
		cm.ll.RUnlock()
		cm.ll.Lock()
		_, ok := cm.m[key]
		if !ok {
			c := newDataCounter()
			modify(c, value)
			cm.m[key] = c
		} else {
			modify(cm.m[key], value)
		}
		cm.ll.Unlock()
	} else {
		modify(cm.m[key], value)
		cm.ll.RUnlock()
	}
}

// encodeCounterKey encodes the name and tags as a byte in a string.
// The first byte gives the position of the tags in the array.
// If the first byte is 0, the tags are empty.
func encodeCounterKey(name string, tags []string) string {
	var key []byte
	if len(tags) == 0 {
		key = append(key, 0)
		key = append(key, ([]byte(name))...)
	} else {
		key = append(key, byte(len(name)+1))
		key = append(key, ([]byte(name))...)
		for i, tag := range tags {
			if i != 0 {
				key = append(key, ([]byte(separator))...)
			}
			key = append(key, ([]byte(tag))...)
		}

	}

	return string(key[:])
}

// decodeCounterKey extraces the name and tags from the key.
// if the first byte is 0, the tags are considered being empty
func decodeCounterKey(key string) (string, []string) {
	bkey := []byte(key)
	tagPos := int(bkey[0])
	if tagPos == 0 {
		return string(key[1:]), []string{}
	}
	tags := strings.Split(string(key[tagPos:]), separator)
	return string(key[1:tagPos]), tags

}

func writeHistogram() {
	for key, c := range cMap.m {
		val := c.reset()
		name, tags := decodeCounterKey(key)
		writeMetric(name, tags, val)
	}
	writeMemoryConsumption()
}

func writePersistentCounter() {
	for key, c := range pcMap.m {
		val := c.read()
		name, tags := decodeCounterKey(key)
		writeMetric(name, tags, val)
	}
}

func writeStack() {
	Info("write printstack")
	pprof.Lookup("goroutine").WriteTo(os.Stderr, 1)
}

func ResetPersistentCounter() {
	for _, c := range pcMap.m {
		c.reset()
	}
}

func writeMemoryConsumption() {
	var mem runtime.MemStats

	runtime.ReadMemStats(&mem)

	writeMetric("memusage.alloc", []string{}, int64(mem.Alloc))
	writeMetric("memusage.stackinuse", []string{}, int64(mem.StackInuse))
	writeMetric("memusage.heapalloc", []string{}, int64(mem.HeapAlloc))
	writeMetric("memusage.heapinuse", []string{}, int64(mem.HeapInuse))
	writeMetric("memusage.numgoroutine", []string{}, int64(runtime.NumGoroutine()))
}

func writeMetric(name string, tags []string, val int64) {
	logMetric(name, tags, val)
	sendMetric(name, tags, val)
}

func logMetric(name string, tags []string, val int64) {
	if len(tags) == 0 {
		Infof("%s: %d\n", name, val)
	} else {
		Infof("%s#%v: %d\n", name, tags, val)
	}

}

func sendMetric(name string, tags []string, val int64) {
	if ddCli == nil {
		return
	}
	ddCli.Gauge(name, float64(val), tags, 1)
}

func newDataCounter() *dataCounter {
	return &dataCounter{
		i: new(int64),
	}
}

func (d *dataCounter) increment(val int64) {
	atomic.AddInt64(d.i, val)
}

func (d *dataCounter) decrement(val int64) {
	atomic.AddInt64(d.i, -val)
}

func (d *dataCounter) set(val int64) {
	atomic.StoreInt64(d.i, val)
}

func (d *dataCounter) read() int64 {
	return atomic.LoadInt64(d.i)
}

func (d *dataCounter) reset() int64 {
	current := atomic.LoadInt64(d.i)
	atomic.StoreInt64(d.i, 0)

	return current
}

type Trace struct {
	key   string
	tags  []string
	start time.Time
}

func StartTrace(key string, tags []string) *Trace {
	return &Trace{
		key:   key,
		tags:  tags,
		start: time.Now(),
	}
}

func (t *Trace) Stop() {

	delta := time.Since(t.start).Nanoseconds() / 1000000

	IncTaggedCounter(t.key+"processed", t.tags, 1)
	IncTaggedCounter(t.key+"processingtime", t.tags, delta)
}
