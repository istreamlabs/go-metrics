package metrics

import (
	"fmt"
	"path"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/DataDog/datadog-go/statsd"
)

// Call describes either a metrics or event call. You can cast it to
// either `MetricCall` or `EventCall` to get at type-specific fields
// for custom checks. Conversion to a string results in a serialized
// representation that looks like one of the following, where the (RATE)
// field is only shown when not equal to 1.0:
//
//   // Serialized metric
//   NAME:VALUE(RATE)[TAG_NAME:TAG_VALUE TAG_NAME:TAG_VALUE ...]
//
//   // Serialized event
//   TITLE:TEXT[TAG_NAME:TAG_VALUE TAG_NAME:TAG_VALUE ...]
//
//   // Example of casting to get additional data
//   value := call.(*MetricCall).Value
//   title := call.(*EventCall).Event.Title
type Call fmt.Stringer

// TestFailer provides a small subset of methods that can be used to fail a
// test case. The built-in `testing.T` and `testing.B` structs implement
// these methods, as do other frameworks, e.g. `GinkgoT()` for Ginkgo.
type TestFailer interface {
	Fatalf(format string, args ...interface{})
}

// MetricCall tracks a single metrics call, value, and tags. All values are
// converted to `float64` from the `int`, `float64`, or `time.Duration` inputs.
type MetricCall struct {
	Name   string
	Value  float64
	Rate   float64
	TagMap map[string]string
}

// String returns a serialized representation of the metric.
func (m *MetricCall) String() string {
	tags := mapToStrings(m.TagMap)
	sort.Strings(tags)
	if m.Rate != 1.0 {
		return fmt.Sprintf("%s:%v(%v)%v", m.Name, m.Value, m.Rate, tags)
	}

	return fmt.Sprintf("%s:%v%v", m.Name, m.Value, tags)
}

// EventCall tracks a single event call and tags.
type EventCall struct {
	Event  *statsd.Event
	TagMap map[string]string
}

// String returns a serialized representation of the event.
func (e *EventCall) String() string {
	tags := mapToStrings(e.TagMap)
	sort.Strings(tags)
	return fmt.Sprintf("%s:%s%v", e.Event.Title, e.Event.Text, tags)
}

// stackInfo returns a string representation of the metrics call stack.
func stackInfo(info *callInfo) string {
	stack := make([]string, 0, len(info.Calls))
	for _, item := range info.Calls {
		stack = append(stack, item.String())
	}
	return strings.Join(stack, "\n")
}

type callInfo struct {
	Calls   []Call
	RWMutex sync.RWMutex
}

// RecorderClient records any metric that is sent, allowing you to make
// assertions about the metrics flowing out of a service. A shared call
// context is used that allows cloned clients to all write to the same
// call info store.
//
// Assertion methods for convenient testing
//
// These methods provide a fast way to write tests while providing useful
// and consistent output in the event of a test failure. For example:
//
//   func MyTest(t *testing.T) {
//     recorder := metrics.NewRecorderClient().WithTest(t)
//     recorder.Count("my.metric", 1)
//
//     recorder.Expect("my.metric").Value(1)
//   }
//
// It is possible to assert that a given metric, value, tag, etc does not
// exist and fail a test if it does:
//
//   func MyTest(t *testing.T) {
//     recorder := metrics.NewRecorderClient().WithTest(t)
//     recorder.Count("my.metric", 5)
//
//     // Since a value of 10 was set, this succeeds.
//     recorder.If("my.metric").Value(10).Reject()
//
//     // However, these two would fail the test.
//     recorder.If("my.metric").Reject()
//     recorder.If("my.metric").Value(5).Reject()
//   }
//
// Custom Checks
//
// The recorder provides access to individual call information so that
// custom checks can be written if needed. For example, to check that a given
// metric name was emitted twice with values in a particular order:
//
//   func MyTest(t *testing.T) {
//     recorder := metrics.NewRecorderClient().WithTest(t)
//     recorder.Count("foo", 1)
//     recorder.Count("foo", 2)
//
//     values := make([]float64, 2)
//     for _, call := range recorder.Expect("foo").GetCalls() {
//       incr = call.(*metrics.MetricCall)
//       values = append(values, incr.Value)
//     }
//     if !reflect.DeepEqual(values, []float64{1.0, 2.0}) {
//       recorder.Fatalf("Expected values '1, 2' in order.")
//     }
//   }
//
type RecorderClient struct {
	callInfo *callInfo
	test     TestFailer
	rate     float64
	tagMap   map[string]string
}

// NewRecorderClient creates a new recording metrics client.
func NewRecorderClient() *RecorderClient {
	return &RecorderClient{
		callInfo: &callInfo{},
		rate:     1.0,
	}
}

// WithTags clones this client with additional tags. Duplicate tags overwrite
// the existing value.
func (c *RecorderClient) WithTags(tags map[string]string) Client {
	return &RecorderClient{
		callInfo: c.callInfo,
		test:     c.test,
		rate:     c.rate,
		tagMap:   combine(c.tagMap, tags),
	}
}

// WithRate clones this client with a new sample rate.
func (c *RecorderClient) WithRate(rate float64) Client {
	return &RecorderClient{
		callInfo: c.callInfo,
		test:     c.test,
		rate:     rate,
		tagMap:   combine(map[string]string{}, c.tagMap),
	}
}

// WithTest returns a recorder client linked with a given test instance.
func (c *RecorderClient) WithTest(test TestFailer) *RecorderClient {
	return &RecorderClient{
		callInfo: c.callInfo,
		test:     test,
		rate:     c.rate,
		tagMap:   c.tagMap,
	}
}

// logCall will record a single metrics call.
func (c *RecorderClient) logCall(name string, value interface{}) {
	tagMapCopy := make(map[string]string, len(c.tagMap))
	for k, v := range c.tagMap {
		tagMapCopy[k] = v
	}
	c.callInfo.RWMutex.Lock()
	defer c.callInfo.RWMutex.Unlock()
	c.callInfo.Calls = append(c.callInfo.Calls, &MetricCall{
		Name:   name,
		Value:  toFloat64(value),
		Rate:   c.rate,
		TagMap: tagMapCopy,
	})
}

// Count adds some value to a metric.
func (c *RecorderClient) Count(name string, value int64) {
	// Normally this would be stored as an integer, but instead we assert that
	// it can be cast to an int, cast it, and then store it as a float so that
	// assertions below are simpler.
	c.logCall(name, value)
}

// Incr adds one to a metric.
func (c *RecorderClient) Incr(name string) {
	c.Count(name, 1)
}

// Decr subtracts one from a metric.
func (c *RecorderClient) Decr(name string) {
	c.Count(name, -1)
}

// Gauge sets a numeric value.
func (c *RecorderClient) Gauge(name string, value float64) {
	c.logCall(name, value)
}

// Event tracks an event that may be relevant to other metrics.
func (c *RecorderClient) Event(e *statsd.Event) {
	var tagMapCopy map[string]string
	for k, v := range c.tagMap {
		tagMapCopy[k] = v
	}
	c.callInfo.RWMutex.Lock()
	defer c.callInfo.RWMutex.Unlock()
	c.callInfo.Calls = append(c.callInfo.Calls, &EventCall{
		Event:  e,
		TagMap: tagMapCopy,
	})
}

// Timing tracks a duration.
func (c *RecorderClient) Timing(name string, value time.Duration) {
	c.logCall(name, value)
}

// Histogram sets a numeric value while tracking min/max/avg/p95/etc.
func (c *RecorderClient) Histogram(name string, value float64) {
	c.logCall(name, value)
}

// Reset will clear the call info context, which is useful between test runs.
func (c *RecorderClient) Reset() {
	c.callInfo.RWMutex.Lock()
	defer c.callInfo.RWMutex.Unlock()
	c.callInfo.Calls = make([]Call, 0)
}

// Length returns the number of calls in the call info context. It is a
// shorthand for `len(recorder.GetCalls())`.
func (c *RecorderClient) Length() int {
	c.callInfo.RWMutex.RLock()
	defer c.callInfo.RWMutex.RUnlock()
	return len(c.callInfo.Calls)
}

// Fatalf fails whatever test is attached to this recorder and additionally
// appends the current metrics call stack and calling information to the
// output message to help with debugging.
func (c *RecorderClient) Fatalf(format string, args ...interface{}) {
	if c.test == nil {
		panic("No test associated with metrics recorder, you must call `recorder.WithTest(t)`")
	}
	// blacklist contains a set of fully qualified function name components that
	// we will filter out to keep the call stack concise.
	blacklist := []string{
		"github.com/istreamlabs/go-metrics/metrics.",
		"testing.tRunner",
		"runtime.goexit",
	}
	buf := "\nFrom call stack:\n"
	callers := make([]uintptr, 10)
	runtime.Callers(1, callers)
	frames := runtime.CallersFrames(callers)
PRINT_FRAMES:
	for {
		frame, more := frames.Next()
		for _, component := range blacklist {
			if strings.Contains(frame.Function, component) {
				continue PRINT_FRAMES
			}
		}
		if frame.Function != "" {
			blurb := getBlurb(frame.File, frame.Line)
			buf += fmt.Sprintf("%s %s:%d\n\t%s\n", frame.Function, path.Base(frame.File), frame.Line, blurb)
		}
		if !more {
			break
		}
	}

	args = append(args, stackInfo(c.callInfo), buf)
	c.test.Fatalf(format+" Current metrics stack:\n%s%s", args...)
}

// GetCalls returns a slice of all recorded calls.
func (c *RecorderClient) GetCalls() []Call {
	return c.callInfo.Calls
}

// ExpectEmpty asserts that no metrics have been emitted.
func (c *RecorderClient) ExpectEmpty() {
	c.callInfo.RWMutex.RLock()
	defer c.callInfo.RWMutex.RUnlock()
	if len(c.callInfo.Calls) > 0 {
		c.Fatalf("Expected empty metrics call stack.")
	}
}

// callsCopy creates a shallow copy of the calls list.
func (c *RecorderClient) callsCopy() []Call {
	c.callInfo.RWMutex.RLock()
	defer c.callInfo.RWMutex.RUnlock()
	calls := make([]Call, len(c.callInfo.Calls))
	for i := 0; i < len(c.callInfo.Calls); i++ {
		calls[i] = c.callInfo.Calls[i]
	}
	return calls
}

// Expect finds metrics (by name) or events (by title) and returns the
// matching calls. A wildcard `*` character will match any ID. This method does
// *not* remove the call from the recorded call list.
//
//   // Get a metric by its name.
//   recorder.Expect("my.metric")
//
//   // Get an event by its title.
//   recorder.Expect("my.event")
func (c *RecorderClient) Expect(id string) Query {
	return (&query{
		calls:    c.callsCopy(),
		test:     c,
		minCalls: 1,
		checkMin: true,
	}).ID(id)
}

// ExpectContains finds metrics or events that contain the `component` in their
// serialized representations. It does *not* remove the call from the recorded
// call list.
//
//   recorder.Incr("foo1")
//   recorder.Incr("foo2")
//
//   // The following matches both calls above.
//   recorder.ExpectContains("foo")
//
// See `Call.String()` for the serialization format.
func (c *RecorderClient) ExpectContains(component string) Query {
	return (&query{
		calls:    c.callsCopy(),
		test:     c,
		minCalls: 1,
		checkMin: true,
	}).Contains(component)
}

// If acts like `Expect`, but doesn't check for the minimum number of calls
// after each query operation. It is to be used with query methods like
// `Accept` and `Reject`. For example:
//
//   // Create a metric with the good tag.
//   recorder.WithTags(map[string]string{"good": "tag"}).Incr("my.metric")
//
//   // This will not fail, because the bad tag is not found.
//   recorder.If("my.metric").Tag("bad", "tag").Reject()
//
//   // This will fail because the metric is found.
//   recorder.If("my.metric").Reject()
//
//   // The following are equivalent, but the first is preferred.
//   recorder.Expect("my.metric")
//   recorder.If("my.metric").Accept()
func (c *RecorderClient) If(id string) Query {
	return (&query{
		calls:    c.callsCopy(),
		test:     c,
		minCalls: 1,
		checkMin: false,
	}).ID(id)
}
