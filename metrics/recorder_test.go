package metrics_test

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/istreamlabs/go-metrics/metrics"
)

// ExpectEqual compares two values and fails if they are not deeply equal.
func ExpectEqual(t *testing.T, expected, actual interface{}) {
	t.Helper()
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("Expected '%s' to be '%s'", actual, expected)
	}
}

type fakeTest struct{}

// Fatalf normally fails a test, but here we just panic, and it will be caught
// later to ensure that 1. It was called and 2. No subsequent code after the
// call runs, otherwise we might get other unrelated failures.
func (ft *fakeTest) Fatalf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}

// ExpectFailure will call a given function with a fake test and recording
// metrics client, then catch any panic which occurs from the fake test having
// its `Fatalf` method called.
func ExpectFailure(t metrics.TestFailer, msg string, handler func(*metrics.RecorderClient)) {
	defer func() {
		// Catch the panic from the test failer and
		if err := recover(); err == nil {
			t.Fatalf(msg)
		}
	}()
	handler(metrics.NewRecorderClient().WithTest(&fakeTest{}))
}

func ExampleRecorderClient() {
	recorder := metrics.NewRecorderClient()
	recorder.WithTags(map[string]string{
		"tag1": "value1",
		"tag2": "value2",
	}).Incr("requests.count")
	recorder.Incr("other")

	// When run in a unit test, you can invoke assertion methods.
	t := &testing.T{}
	recorder.WithTest(t).
		Expect("requests.count").
		Value(1).
		Tag("tag1", "value1").
		Tag("tag2", "value2")
	// Output:
}

func ExampleRecorderClient_output() {
	recorder := metrics.NewRecorderClient()
	recorder.WithTags(map[string]string{
		"tag1": "value1",
		"tag2": "value2",
	}).Incr("requests.count")
	recorder.Incr("other")

	// Converting calls to a string returns their serialized representation.
	for _, call := range recorder.GetCalls() {
		fmt.Println(call)
	}
	// Output: requests.count:1[tag1:value1 tag2:value2]
	// other:1[]
}

func ExampleMetricCall() {
	// First, create a recorder and emit a metric.
	recorder := metrics.NewRecorderClient()
	recorder.Incr("requests.count")

	// Then, print out some info about the `Incr` call.
	call := recorder.GetCalls()[0].(*metrics.MetricCall)
	fmt.Printf("'%s' value is '%v'", call.Name, call.Value)
	// Output: 'requests.count' value is '1'
}

func TestRecorderClient(t *testing.T) {
	var client metrics.Client
	client = metrics.NewRecorderClient().WithTest(t)

	client.Incr("one")
	client.Event(statsd.NewEvent("title", "desc"))
	client.Histogram("histo", 4.3)

	sub := client.WithTags(map[string]string{
		"tag1": "value1",
	})

	sub.Count("two", 2)

	sub2 := sub.WithTags(map[string]string{
		"tag1": "override",
		"tag2": "value2",
	})

	sub2.Timing("three", 3*time.Second)

	client.Decr("one")
	client.Gauge("memory", 1024)
	client.Histogram("histo", 123)
	client.Distribution("distro", 999)

	// Cast to access additional methods for testing.
	recorder := client.(*metrics.RecorderClient)

	if recorder.Length() < 4 {
		t.Fatal("Expected a length of at least 4")
	}

	recorder.Expect("one").Value(1)
	recorder.Expect("title").Text("desc")
	recorder.Expect("histo").Value(4.3)
	recorder.ExpectContains("two:2[tag1:value1]")
	recorder.Expect("distro").Value(999)
	recorder.
		Expect("three").
		Tag("tag1", "override").
		TagName("tag2")
	recorder.Expect("one").Value(-1)
	recorder.Expect("memory").Value(1024)
	recorder.Expect("histo").Value(123)
	recorder.Expect("distro").Value(999)

	recorder.If("*").Tag("tag1", "override2").Reject()

	client.Incr("one")
	recorder.Reset()
	if recorder.Length() > 0 {
		t.Fatal("Expected a length of zero after reset")
	}

	client.Close()

	recorder.Incr("one")
	recorder.Incr("two")
	calls := recorder.Expect("one").GetCalls()
	if len(calls) > 1 {
		t.Fatal("Expected query to return only one matching call")
	}
}

func TestRecorderAssertionNameFails(t *testing.T) {
	ExpectFailure(t, "Expecting wrong name should fail test",
		func(recorder *metrics.RecorderClient) {
			recorder.Incr("one")
			recorder.Expect("on2")
		})
}

func TestRecorderAssertionNameWithTagsFails(t *testing.T) {
	ExpectFailure(t, "Expecting wrong tag name should fail test",
		func(recorder *metrics.RecorderClient) {
			recorder.WithTags(map[string]string{"foo": "1"}).Incr("one")
			recorder.Expect("one").Tag("bar", "1")
		})
}

func TestRecorderAssertionNameWithTagNameFails(t *testing.T) {
	ExpectFailure(t, "Expecting wrong tag name should fail test",
		func(recorder *metrics.RecorderClient) {
			recorder.WithTags(map[string]string{"foo": "1"}).Incr("one")
			recorder.Expect("one").TagName("bar")
		})
}

func TestRecorderAssertionContainsFails(t *testing.T) {
	ExpectFailure(t,
		"Expecting value not contained within serialized metric should fail",
		func(recorder *metrics.RecorderClient) {
			recorder.Incr("one")
			recorder.ExpectContains("b")
		})
}

func TestRecorderEmptyStackFails(t *testing.T) {
	ExpectFailure(t, "Empty metrics stack assertions should fail",
		func(recorder *metrics.RecorderClient) {
			recorder.Expect("test")
		})
}

func TestRecorderExpectingMetricWhenEventFails(t *testing.T) {
	ExpectFailure(t, "Expecting metrics when next item is event should fail",
		func(recorder *metrics.RecorderClient) {
			recorder.Event(statsd.NewEvent("test", ""))
			recorder.Expect("test").Value(1)
		})
}

func TestRecorderExpectEmpty(t *testing.T) {
	recorder := metrics.NewRecorderClient().WithTest(t)
	recorder.ExpectEmpty()

	ExpectFailure(t, "Expecting empty assertion to fail when metrics present",
		func(recorder *metrics.RecorderClient) {
			recorder.Incr("foo")
			recorder.ExpectEmpty()
		})
}

func TestRecorderRejection(t *testing.T) {
	recorder := metrics.NewRecorderClient().WithTest(t)
	recorder.WithTags(map[string]string{
		"tag1": "1",
		"tag2": "2",
	}).Incr("foo")
	recorder.Incr("bar")

	// Should not fail test if the condition does not pass.
	recorder.If("baz").Reject()

	// The `foo` metric is present, but only once. This would only fail the test
	// if it were present two or more times.
	recorder.If("foo").MinTimes(2).Reject()

	// Opposite of rejection is to use `accept`.
	recorder.If("foo").Accept()

	// Ensure rejection actually fails the test when the condition passes.
	ExpectFailure(t, "Expecting metric ID rejection to fail when present",
		func(r *metrics.RecorderClient) {
			r.Incr("foo")
			r.If("foo").Reject()
		})

	ExpectFailure(t, "Expecting metric ID acceptance to fail when missing",
		func(r *metrics.RecorderClient) {
			r.Incr("foo")
			r.If("bar").Accept()
		})
}

func TestRecorderMinTimes(t *testing.T) {
	ExpectFailure(t, "Expecting at least two metrics when only one matches should fail",
		func(recorder *metrics.RecorderClient) {
			recorder.Count("foo", 5.0)
			recorder.Count("foo", 2)
			recorder.Incr("bar")

			// We expect two, but only one matches both the name AND number.
			recorder.Expect("foo").Value(5.0).MinTimes(2)
		})

	ExpectFailure(t, "Expecting at least one metric when the stack is empty should fail",
		func(recorder *metrics.RecorderClient) {
			recorder.Expect("*")
		})
}

func TestRecorderAssertionRequiresFailer(t *testing.T) {
	client := metrics.NewRecorderClient()

	defer func() {
		// Catch the panic from the test failer and
		if err := recover(); err == nil {
			t.Fatalf("Assertion without test failer should panic")
		}
	}()

	client.Expect("foo")
}

func TestRecorderConcurrency(t *testing.T) {
	client := metrics.NewRecorderClient().WithTest(t)

	// Write metrics concurrently, then wait for them to all complete.
	wg := sync.WaitGroup{}
	wg.Add(3)
	for i := 0; i < 3; i++ {
		go func() {
			client.Incr("test.concurrency")
			wg.Done()
		}()
	}
	wg.Wait()

	// Since there are three goroutines above, we should get three metrics.
	client.Expect("test.concurrency").MinTimes(3)
}

func TestRecorderWithRate(t *testing.T) {
	recorder := metrics.NewRecorderClient().WithTest(t)

	recorder.WithRate(0.1).Incr("sampled")

	recorder.If("sampled").Rate(1.0).Reject()
	recorder.Expect("sampled").Rate(0.1)
}
