package metrics_test

import (
	"testing"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/istreamlabs/go-metrics/metrics"
)

func ExampleNullClient() {
	client := metrics.NewNullClient()
	client.WithTags(map[string]string{
		"tag1": "value1",
		"tag2": "value2",
	}).Incr("requests.count")
	client.Incr("other")

	// Output:
}

func TestNullClientMethods(t *testing.T) {
	client := metrics.NewNullClient()

	client.Incr("count")
	client.Decr("count")
	client.Count("count", 5)

	client.Gauge("gauge", 10)

	client.Histogram("histo", 1.25)
	client.Timing("timing", time.Duration(123))
	client.Distribution("distro", 999)

	client.Event(&statsd.Event{})

	client.WithRate(1.2).Incr("rated")
	client.Close()
}
