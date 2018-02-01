package glog

import (
	"sync"
	"testing"
	"time"
)

var (
	wg sync.WaitGroup
)

const (
	waitTime = 1200
)

func setAPMFlags() {
	logInterval = 1
}

func TestIncCounter(t *testing.T) {
	setFlags()
	setAPMFlags()
	defer logging.swap(logging.newBuffers())
	StartHistogramLogging()

	IncCounter("test1", 1)
	IncCounter("test2", 1)
	IncCounter("test1", 1)
	time.Sleep(waitTime * time.Millisecond)

	if !contains(infoLog, "test1: 2", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}

	if !contains(infoLog, "test2: 1", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}

	time.Sleep(1 * time.Second)
	if !contains(infoLog, "test1: 0", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}

	if !contains(infoLog, "test2: 0", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}
}

func TestIncTaggedCounter(t *testing.T) {
	setFlags()
	setAPMFlags()
	defer logging.swap(logging.newBuffers())
	StartHistogramLogging()

	IncTaggedCounter("test1", []string{"tag1:value1"}, 1)
	IncTaggedCounter("test2", []string{"tag1:value1"}, 1)
	IncTaggedCounter("test1", []string{"tag1:value1", "tag2:value2"}, 1)

	time.Sleep(waitTime * time.Millisecond)

	if !contains(infoLog, "test1#[tag1:value1]: ", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}

	if !contains(infoLog, "test2#[tag1:value1]: 1", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}

	if !contains(infoLog, "test1#[tag1:value1 tag2:value2]: 1", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}

}

func TestIncCounterConcurrent(t *testing.T) {
	setFlags()
	setAPMFlags()
	defer logging.swap(logging.newBuffers())
	StartHistogramLogging()

	wg.Add(2)
	go incCounterConcurrent("test1")
	go incCounterConcurrent("test2")
	wg.Wait()
	time.Sleep(waitTime * time.Millisecond)

	if !contains(infoLog, "test1: 1000", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}

	if !contains(infoLog, "test2: 1000", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}
	time.Sleep(1 * time.Second)
	if !contains(infoLog, "test1: 0", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}

	if !contains(infoLog, "test2: 0", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}
}

func incCounterConcurrent(name string) {
	defer wg.Done()
	for i := 0; i < 1000; i++ {
		IncCounter(name, 1)
	}
}

func TestDecCounter(t *testing.T) {
	setFlags()
	setAPMFlags()
	defer logging.swap(logging.newBuffers())
	StartHistogramLogging()

	IncCounter("test1", 2)
	DecCounter("test1", 1)
	DecCounter("test2", 1)

	time.Sleep(waitTime * time.Millisecond)

	if !contains(infoLog, "test1: 1", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}

	if !contains(infoLog, "test2: -1", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}

	time.Sleep(1 * time.Second)
	if !contains(infoLog, "test1: 0", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}

	if !contains(infoLog, "test2: 0", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}
}

func TestDecTaggedCounter(t *testing.T) {
	setFlags()
	setAPMFlags()
	defer logging.swap(logging.newBuffers())
	StartHistogramLogging()

	IncTaggedCounter("test1", []string{"tag1:value1"}, 2)
	DecTaggedCounter("test1", []string{"tag1:value1"}, 1)
	IncTaggedCounter("test2", []string{"tag1:value1"}, 1)
	IncTaggedCounter("test1", []string{"tag1:value1", "tag2:value2"}, 1)
	DecTaggedCounter("test1", []string{"tag1:value1", "tag2:value2"}, 1)

	time.Sleep(waitTime * time.Millisecond)

	if !contains(infoLog, "test1#[tag1:value1]: 1", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}

	if !contains(infoLog, "test2#[tag1:value1]: 1", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}

	if !contains(infoLog, "test1#[tag1:value1 tag2:value2]: 0", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}

}

func TestDecCounterConcurrent(t *testing.T) {
	setFlags()
	setAPMFlags()
	defer logging.swap(logging.newBuffers())
	StartHistogramLogging()
	SetCounter("test1", 2000)
	SetCounter("test2", 2000)
	wg.Add(2)
	go decCounterConcurrent("test1")
	go decCounterConcurrent("test2")
	wg.Wait()
	time.Sleep(waitTime * time.Millisecond)

	if !contains(infoLog, "test1: 1000", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}

	if !contains(infoLog, "test2: 1000", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}
	time.Sleep(1 * time.Second)
	if !contains(infoLog, "test1: 0", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}

	if !contains(infoLog, "test2: 0", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}
}

func decCounterConcurrent(name string) {
	defer wg.Done()
	for i := 0; i < 1000; i++ {
		DecCounter(name, 1)
	}
}

func TestLogMemoryConsumption(t *testing.T) {
	setFlags()
	setAPMFlags()
	defer logging.swap(logging.newBuffers())
	SetLogInterval(1)
	StartHistogramLogging()
	time.Sleep(waitTime * time.Millisecond)

	if !contains(infoLog, "memusage.alloc:", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}

}

func TestEncodeDecodeCounterKey(t *testing.T) {
	// test no tags
	expectedName := "namespace.process.metric"
	expectedTags := []string{}
	name, tags := decodeCounterKey(encodeCounterKey(expectedName, expectedTags))
	if name != expectedName {
		t.Errorf("Expected name %s got %s", expectedName, name)
	}
	if len(tags) != 0 {
		t.Errorf("Expected empty tags got %s", tags)
	}

	// test single tags
	expectedName = "namespace.process.metric"
	expectedTags = []string{"tagname:value"}

	name, tags = decodeCounterKey(encodeCounterKey(expectedName, expectedTags))
	if name != expectedName {
		t.Errorf("Expected name %s got %s", expectedName, name)
	}
	for i, tag := range tags {
		if tag != expectedTags[i] {
			t.Errorf("Expected tag %s got %s", expectedTags, tag)
		}
	}
	// test multiple tags
	expectedName = "namespace.process.metric"
	expectedTags = []string{"tagname1:value1", "tagname2:value2"}

	name, tags = decodeCounterKey(encodeCounterKey(expectedName, expectedTags))
	if name != expectedName {
		t.Errorf("Expected name %s got %s", expectedName, name)
	}
	for i, tag := range tags {
		if tag != expectedTags[i] {
			t.Errorf("Expected tag %s got %s", expectedTags, tag)
		}
	}

}

func TestIncPCounter(t *testing.T) {
	setFlags()
	setAPMFlags()
	defer logging.swap(logging.newBuffers())
	defer ResetPersistentCounter()
	StartHistogramLogging()

	IncPCounter("test1", 1)
	IncPCounter("test2", 1)
	IncPCounter("test1", 1)
	time.Sleep(waitTime * time.Millisecond)

	if !contains(infoLog, "test1: 2", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}

	if !contains(infoLog, "test2: 1", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}

	IncPCounter("test1", 1)
	IncPCounter("test2", 1)
	IncPCounter("test1", 1)

	time.Sleep(1 * time.Second)
	if !contains(infoLog, "test1: 4", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}

	if !contains(infoLog, "test2: 2", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}
}

func TestIncPTaggedCounter(t *testing.T) {
	setFlags()
	setAPMFlags()
	defer logging.swap(logging.newBuffers())
	defer ResetPersistentCounter()
	StartHistogramLogging()

	IncTaggedPCounter("test1", []string{"tag1:value1"}, 1)
	IncTaggedPCounter("test2", []string{"tag1:value1"}, 1)
	IncTaggedPCounter("test1", []string{"tag1:value1", "tag2:value2"}, 1)

	time.Sleep(waitTime * time.Millisecond)

	if !contains(infoLog, "test1#[tag1:value1]: ", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}

	if !contains(infoLog, "test2#[tag1:value1]: 1", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}

	if !contains(infoLog, "test1#[tag1:value1 tag2:value2]: 1", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}

	IncTaggedPCounter("test1", []string{"tag1:value1"}, 1)
	IncTaggedPCounter("test2", []string{"tag1:value1"}, 1)
	IncTaggedPCounter("test1", []string{"tag1:value1", "tag2:value2"}, 1)

	time.Sleep(waitTime * time.Millisecond)

	if !contains(infoLog, "test1#[tag1:value1]: ", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}

	if !contains(infoLog, "test2#[tag1:value1]: 2", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}

	if !contains(infoLog, "test1#[tag1:value1 tag2:value2]: 2", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}

}

func TestIncPCounterConcurrent(t *testing.T) {
	setFlags()
	setAPMFlags()
	defer logging.swap(logging.newBuffers())
	defer ResetPersistentCounter()
	StartHistogramLogging()

	wg.Add(2)
	go incPCounterConcurrent("test1")
	go incPCounterConcurrent("test2")
	wg.Wait()
	time.Sleep(waitTime * time.Millisecond)

	if !contains(infoLog, "test1: 1000", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}

	if !contains(infoLog, "test2: 1000", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}
	time.Sleep(1 * time.Second)
	wg.Add(2)
	go incPCounterConcurrent("test1")
	go incPCounterConcurrent("test2")
	wg.Wait()
	time.Sleep(waitTime * time.Millisecond)

	if !contains(infoLog, "test1: 2000", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}

	if !contains(infoLog, "test2: 2000", t) {
		t.Errorf("Info has wrong character: %q", contents(infoLog))
	}
}

func incPCounterConcurrent(name string) {
	defer wg.Done()
	for i := 0; i < 1000; i++ {
		IncPCounter(name, 1)
	}
}
