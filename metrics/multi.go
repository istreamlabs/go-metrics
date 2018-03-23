package metrics

import (
	"time"

	"github.com/DataDog/datadog-go/statsd"
)

// MultiClient delegates to multiple clients.
type MultiClient struct {
	clients []Client
}

// NewMultiClient creates a new null client.
func NewMultiClient(c ...Client) *MultiClient {
	return &MultiClient{
		clients: c,
	}
}

// WithTags clones this client with additional tags. Duplicate tags overwrite
// the existing value.
func (c *MultiClient) WithTags(tags map[string]string) Client {
	mc := &MultiClient{}
	for _, client := range c.clients {
		mc.clients = append(mc.clients, client.WithTags(tags))
	}
	return mc
}

// WithRate clones this client with a given sample rate.
func (c *MultiClient) WithRate(rate float64) Client {
	mc := &MultiClient{}
	for _, client := range c.clients {
		mc.clients = append(mc.clients, client.WithRate(rate))
	}
	return mc
}

// Count adds some value to a metric.
func (c *MultiClient) Count(name string, value int64) {
	for _, client := range c.clients {
		client.Count(name, value)
	}
}

// Incr adds one to a metric.
func (c *MultiClient) Incr(name string) {
	for _, client := range c.clients {
		client.Incr(name)
	}
}

// Decr subtracts one from a metric.
func (c *MultiClient) Decr(name string) {
	for _, client := range c.clients {
		client.Decr(name)
	}
}

// Gauge sets a numeric value.
func (c *MultiClient) Gauge(name string, value float64) {
	for _, client := range c.clients {
		client.Gauge(name, value)
	}
}

// Event tracks an event that may be relevant to other metrics.
func (c *MultiClient) Event(event *statsd.Event) {
	for _, client := range c.clients {
		client.Event(event)
	}
}

// Timing tracks a duration.
func (c *MultiClient) Timing(name string, value time.Duration) {
	for _, client := range c.clients {
		client.Timing(name, value)
	}
}

// Histogram sets a numeric value while tracking min/max/avg/p95/etc.
func (c *MultiClient) Histogram(name string, value float64) {
	for _, client := range c.clients {
		client.Histogram(name, value)
	}
}
