package metrics_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/mgutz/ansi"

	"github.com/istreamlabs/go-metrics/metrics"
)

// LogRecorder dumps log messages into an array.
type LogRecorder struct {
	messages []string
}

// Printf acts like the standard `log.Printf` method, except that it writes
// into a string array instead of to stdout.
func (l *LogRecorder) Printf(format string, args ...interface{}) {
	l.messages = append(l.messages, fmt.Sprintf(format, args...))
}

func ExampleLoggerClient() {
	client := metrics.NewLoggerClient(nil)
	client.WithTags(map[string]string{
		"tag1": "value1",
	}).Incr("requests.count")
	// Output: Count requests.count:1 map[tag1:value1]
}

func TestLoggerClient(t *testing.T) {
	var client metrics.Client

	recorder := &LogRecorder{}
	client = metrics.NewLoggerClient(recorder)

	client.Incr("one")
	client.Event(statsd.NewEvent("title", "desc"))

	client.WithTags(map[string]string{
		"tag1": "value1",
	}).WithTags(map[string]string{
		"tag1": "override",
	}).Timing("two", 2*time.Second)

	client.Decr("one")
	client.Gauge("memory", 1024)
	client.Histogram("histo", 123)
	client.Distribution("distro", 999)
	client.Close()

	ExpectEqual(t, "Count one:1 map[]", recorder.messages[0])
	ExpectEqual(t, "Event title\ndesc map[]", recorder.messages[1])
	ExpectEqual(t, "Timing two:2s map[tag1:override]", recorder.messages[2])
	ExpectEqual(t, "Count one:-1 map[]", recorder.messages[3])
	ExpectEqual(t, "Gauge memory:1024 map[]", recorder.messages[4])
	ExpectEqual(t, "Histogram histo:123 map[]", recorder.messages[5])
	ExpectEqual(t, "Distribution distro:999 map[]", recorder.messages[6])

	// Make sure the call works, but since it is randomly sampled we have no
	// assertion to make.
	sampled := client.WithRate(0.8)
	sampled.Incr("sampled")
	sampled.Incr("sampled")
	sampled.Gauge("sampled-gauge", 123)

	// Test colorized output
	client.(*metrics.LoggerClient).Colorized().WithTags(map[string]string{
		"tag1": "val1",
		"tag2": "val2",
	}).Incr("colored")

	expected := fmt.Sprintf("Count %v:%v map[%v:val1 %v:val2]",
		ansi.Color("colored", "208"),
		ansi.Color("1", "32"),
		ansi.Color("tag1", "133"),
		ansi.Color("tag2", "133"))

	ExpectEqual(t, expected, recorder.messages[len(recorder.messages)-1])
}
