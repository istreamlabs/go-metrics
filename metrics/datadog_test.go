package metrics_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/istreamlabs/go-metrics/metrics"
)

type withRater interface {
	WithRate(rate float64) metrics.Client
}

func ExampleDataDogClient() {
	datadog := metrics.NewDataDogClient("127.0.0.1:8125", "myprefix")
	datadog.WithTags(map[string]string{
		"tag": "value",
	}).Incr("requests.count")
}

func TestDataDogClient(t *testing.T) {
	// This connects to an address that's probably not running anything. The
	// stats essentially go into `/dev/null`. Right now the only thing this
	// ensures is that the functions can be called without crashing.
	// TODO: In the future, we should use a statsd mock here.
	var datadog metrics.Client
	datadog = metrics.NewDataDogClient("127.0.0.1:8126", "testing")

	datadog.Incr("one")
	datadog.Event(statsd.NewEvent("title", "desc"))
	datadog.Timing("two", 2*time.Second)

	datadog.WithTags(map[string]string{
		"tag1": "value1",
	}).Incr("three")

	datadog.Decr("one")
	datadog.Gauge("memory", 1024)
	datadog.Histogram("histo", 123)

	if rater, ok := datadog.(withRater); ok {
		ratedClient := rater.WithRate(0.5)
		ratedClient.Incr("rated")
	} else {
		t.Fatalf("Expected DataDog client to support sample rate")
	}

	// Test that tag overrides work.
	override := datadog.WithTags(map[string]string{
		"tag1": "value1",
	}).WithTags(map[string]string{
		"tag1": "override",
		"tag2": "value2",
	})

	actual := override.(*metrics.DataDogClient).TagMap()
	expected := map[string]string{
		"tag1": "override",
		"tag2": "value2",
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("Expected %v to equal %v", actual, expected)
	}
}
