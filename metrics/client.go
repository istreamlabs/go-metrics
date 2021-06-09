// Package metrics provides a generic interface for sending metrics from your
// script, application, or service. Metrics consist of a name, a numeric value,
// and optionally a number of tags that are made up of key/value pairs.
//
// Multiple implementations are provided to allow for production, local
// development, and testing use cases. Generally you will instantiate one
// of these clients and then write metrics to it:
//
//   import "github.com/istreamlabs/go-metrics/metrics"
//
//   client := metrics.NewDataDogClient("127.0.0.1:8125", "myprefix")
//
//   // Simple incrementing counter
//   client.Incr("requests.count")
//
//   // Tagging with counters
//   client.WithTags(map[string]string{
//     "tag": "value"
//   }).Incr("requests.count")
//
//   // Sample rate for high-throughput applications
//   client.WithRate(0.01).Incr("requests.count")
//
// Also provided are useful clients for testing, both for when you want
// to assert that certain metrics are emitted and a `NullClient` for when
// you want to ignore them.
//
//   func TestFoo(t *testing.T) {
//     client := metrics.NewRecorderClient().WithTest(t)
//
//     client.WithTags(map[string]string{
//       "tag": "value",
//     }).Count("requests.count", 1)
//
//     // Now, assert that the metric was emitted.
//     client.
//       Expect("requests.count").
//       Value(1).
//       Tag("tag", "value")
//   }
//
package metrics

import (
	"time"

	"github.com/DataDog/datadog-go/statsd"
)

// Client provides a generic interface to log metrics and events
type Client interface {
	// WithTags returns a new client with the given tags.
	WithTags(tags map[string]string) Client

	// WithRate returns a new client with the given sample rate.
	WithRate(rate float64) Client

	// Count/Incr/Decr set a numeric integer value.
	Count(name string, value int64)
	Incr(name string)
	Decr(name string)

	// Gauge sets a numeric floating point value.
	Gauge(name string, value float64)

	// Event creates a new event, which allows additional information to be
	// included when something worth calling out happens.
	Event(e *statsd.Event)

	// Timing creates a histogram of a duration.
	Timing(name string, value time.Duration)

	// Historgram creates a numeric floating point metric with min/max/avg/p95/etc.
	Histogram(name string, value float64)

	// Distribution tracks the statistical distribution of a set of values.
	Distribution(name string, value float64)

	// Close closes all client connections and flushes any buffered data.
	Close() error
}
