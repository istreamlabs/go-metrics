package metrics_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
	"github.com/istreamlabs/go-metrics/metrics"
)

type withRater interface {
	WithRate(rate float64) metrics.Client
}

func ExampleDataDogClient() {
	// Create a new DataDog metrics client.
	datadog := metrics.NewDataDogClient("127.0.0.1:8125", "myprefix")
	datadog.WithTags(map[string]string{
		"tag": "value",
	}).Incr("requests.count")

	// Create a DataDog metrics client with a custom configured statsd.
	client, err := statsd.New("127.0.0.1:8125", statsd.WithNamespace("myprefix"))
	if err != nil {
		panic(err)
	}
	custom := metrics.NewDataDogClient("", "", metrics.WithStatsd(client))
	custom.WithTags(map[string]string{
		"tag": "value",
	}).Incr("requests.count")
}

func TestDataDogClient(t *testing.T) {
	// This connects to an address that's probably not running anything. The
	// stats essentially go into `/dev/null`. Right now the only thing this
	// ensures is that the functions can be called without crashing.
	// TODO: In the future, we should use a statsd mock here.
	var datadog metrics.Client = metrics.NewDataDogClient("127.0.0.1:8126", "testing", metrics.WithoutTelemetry())

	datadog.Incr("one")
	datadog.Event(statsd.NewEvent("title", "desc"))
	datadog.Timing("two", 2*time.Second)

	datadog.WithTags(map[string]string{
		"tag1": "value1",
	}).Incr("three")

	datadog.Decr("one")
	datadog.Gauge("memory", 1024)
	datadog.Histogram("histo", 123)
	datadog.Distribution("distro", 999)

	if rater, ok := datadog.(withRater); ok {
		ratedClient := rater.WithRate(0.5)
		ratedClient.Incr("rated")
	} else {
		t.Fatalf("Expected DataDog client to support sample rate")
	}

	// Test that tag overrides work.
	override := datadog.WithTags(map[string]string{
		"tag1": "value1",
		"tag2": "value2",
	}).WithTags(map[string]string{
		"tag1": "override",
		"tag3": "value3",
	})

	actual := override.(*metrics.DataDogClient).Tags()
	expected := []string{
		"tag1:override",
		"tag1:value1",
		"tag2:value2",
		"tag3:value3",
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("Expected %v to equal %v", actual, expected)
	}

	// Events should get tags assigned automatically.
	e := &statsd.Event{
		Title: "Test event",
	}

	datadog.WithTags(map[string]string{
		"tag1": "value1",
	}).Event(e)

	if !reflect.DeepEqual(e.Tags, []string{"tag1:value1"}) {
		t.Fatalf("Expected event to have tags '[tag1:value1]'. Found '%v'", e.Tags)
	}

	datadog.Close()
}

func TestDataDogCustom(t *testing.T) {
	client, err := statsd.New("127.0.0.1:8125", statsd.WithNamespace("myprefix"))
	if err != nil {
		panic(err)
	}
	// This should pass without error even though no address is passed in because
	// the client we created above does have an address.
	custom := metrics.NewDataDogClient("", "", metrics.WithStatsd(client))
	custom.WithTags(map[string]string{
		"tag": "value",
	}).Incr("requests.count")
	custom.Close()
}

func Benchmark_0Tags_100Emits(b *testing.B) {
	benchmarkClient(b, 0, 100, true, false, nil)
}

func BenchmarkTags_5Tags_100Emits(b *testing.B) {
	benchmarkClient(b, 5, 100, true, false, nil)
}

func BenchmarkTags_5Tags_100Emits_WithInline(b *testing.B) {
	benchmarkClient(b, 5, 100, true, true, nil)
}

func BenchmarkTags_10Tags_1000Emits(b *testing.B) {
	benchmarkClient(b, 10, 1000, true, false, nil)
}

func BenchmarkTags_10Tags_1000Emits_WithInline(b *testing.B) {
	benchmarkClient(b, 10, 1000, true, true, nil)
}

func BenchmarkTags_15Tags_100Emits(b *testing.B) {
	benchmarkClient(b, 15, 100, true, false, nil)
}

func BenchmarkTags_15Tags_100Emits_WithInline(b *testing.B) {
	benchmarkClient(b, 15, 100, true, true, nil)
}

func BenchmarkTags_15Tags_1000Emits_Incr(b *testing.B) {
	benchmarkClient(b, 15, 1000, false, false, nil)
}

func BenchmarkTags_15Tags_1000Emits_Incr_WithInline(b *testing.B) {
	benchmarkClient(b, 15, 1000, false, true, nil)
}

func BenchmarkTags_15Tags_1000Emits_Incr_NoAggr(b *testing.B) {
	client, _ := statsd.New("127.0.0.1:8126", statsd.WithoutClientSideAggregation())
	benchmarkClient(b, 15, 1000, false, false, client)
}

func benchmarkClient(b *testing.B, numTags, numMetrics int, histo bool, inlineTags bool, client *statsd.Client) {
	options := []metrics.Option{}
	if client != nil {
		options = append(options, metrics.WithStatsd(client))
	}
	var datadog metrics.Client = metrics.NewDataDogClient("127.0.0.1:8126", "testing", options...)
	defer datadog.Close()

	tags := map[string]string{}
	for i := 0; i < numTags; i++ {
		tags[fmt.Sprintf("tag-%v", i)] = fmt.Sprintf("value-%v", i)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cli := datadog.WithTags(tags)
		for m := 0; m < numMetrics; m++ {
			if histo {
				if inlineTags {
					cli.WithTags(map[string]string{"a": "b"}).Histogram("histo", 123)
				} else {
					cli.Histogram("histo", 123)
				}
			} else {
				if inlineTags {
					cli.WithTags(map[string]string{"a": "b"}).Incr("incr")
				} else {
					cli.Incr("incr")
				}
			}
		}
	}
}
