package metrics

import (
	"time"

	"github.com/DataDog/datadog-go/statsd"
)

// NullClient does nothing. Useful for tests when you do not care about metrics
// state or output.
type NullClient struct {
}

// NewNullClient creates a new null client.
func NewNullClient() *NullClient {
	return &NullClient{}
}

// WithTags clones this client with additional tags. Duplicate tags overwrite
// the existing value.
func (c *NullClient) WithTags(tags map[string]string) Client {
	return &NullClient{}
}

// WithRate clones this client with a given sample rate.
func (c *NullClient) WithRate(rate float64) Client {
	return &NullClient{}
}

// Close on a NullClient is a no-op
func (c *NullClient) Close() error {
	return nil
}

// Count adds some value to a metric.
func (c *NullClient) Count(name string, value int64) {
}

// Incr adds one to a metric.
func (c *NullClient) Incr(name string) {
}

// Decr subtracts one from a metric.
func (c *NullClient) Decr(name string) {
}

// Gauge sets a numeric value.
func (c *NullClient) Gauge(name string, value float64) {
}

// Event tracks an event that may be relevant to other metrics.
func (c *NullClient) Event(event *statsd.Event) {
}

// Timing tracks a duration.
func (c *NullClient) Timing(name string, value time.Duration) {
}

// Histogram sets a numeric value while tracking min/max/avg/p95/etc.
func (c *NullClient) Histogram(name string, value float64) {
}

// Distribution tracks the statistical distribution of a set of values.
func (c *NullClient) Distribution(name string, value float64) {
}
