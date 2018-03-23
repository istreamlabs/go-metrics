package metrics_test

import (
	"testing"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/istreamlabs/go-metrics/metrics"
)

func TestMultiClient(t *testing.T) {
	r1 := &LogRecorder{}
	r2 := &LogRecorder{}

	c1 := metrics.NewLoggerClient(r1)
	c2 := metrics.NewLoggerClient(r2)

	client := metrics.NewMultiClient(c1, c2)
	client.Incr("count")
	client.Decr("count")
	client.Count("count", 5)
	client.Gauge("gauge", 10)
	client.Histogram("histo", 1.25)
	client.Timing("timing", time.Duration(123))
	client.Event(&statsd.Event{})
	client.WithRate(1.2).Incr("rated")

	ExpectEqual(t, r1.messages, r2.messages)
}
