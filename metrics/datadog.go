package metrics

import (
	"log"
	"time"

	"github.com/DataDog/datadog-go/statsd"
)

// DataDogClient is a dogstatsd metrics client implementation.
type DataDogClient struct {
	client *statsd.Client
	rate   float64
	tags   []string
}

// NewDataDogClient creates a new dogstatsd client pointing to `address` with
// the metrics prefix of `namespace`. For example, given a namespace of
// `foo.bar`, a call to `Incr('baz')` would emit a metric with the full name
// `foo.bar.baz` (note the period between the namespace and metric name).
func NewDataDogClient(address string, namespace string) *DataDogClient {
	c, err := statsd.New(address)
	if err != nil {
		log.Panic(err)
	}

	if namespace != "" {
		c.Namespace = namespace + "."
	}

	return &DataDogClient{
		client: c,
		rate:   1.0,
	}
}

// WithRate clones this client with a new sample rate.
func (c *DataDogClient) WithRate(rate float64) Client {
	return &DataDogClient{
		client: c.client,
		rate:   rate,
		tags:   c.tags,
	}
}

// WithTags clones this client with additional tags. Duplicate tags overwrite
// the existing value.
func (c *DataDogClient) WithTags(tags map[string]string) Client {
	return &DataDogClient{
		client: c.client,
		rate:   c.rate,
		tags:   cloneTags(nil, tags),
	}
}

func (c *DataDogClient) tagsList() []string {
	return c.tags
}

// Close closes all client connections and flushes any buffered data.
func (c *DataDogClient) Close() error {
	return c.client.Close()
}

// Count adds some integer value to a metric.
func (c *DataDogClient) Count(name string, value int64) {
	c.client.Count(name, value, c.tagsList(), c.rate)
}

// Incr adds one to a metric.
func (c *DataDogClient) Incr(name string) {
	c.Count(name, 1)
}

// Decr subtracts one from a metric.
func (c *DataDogClient) Decr(name string) {
	c.Count(name, -1)
}

// Gauge sets a numeric value.
func (c *DataDogClient) Gauge(name string, value float64) {
	c.client.Gauge(name, value, c.tagsList(), c.rate)
}

// Event tracks an event that may be relevant to other metrics.
func (c *DataDogClient) Event(e *statsd.Event) {
	if len(c.tags) > 0 {
		e.Tags = append(e.Tags, c.tagsList()...)
	}

	c.client.Event(e)
}

// Timing tracks a duration.
func (c *DataDogClient) Timing(name string, value time.Duration) {
	c.client.Timing(name, value, c.tagsList(), c.rate)
}

// Histogram sets a numeric value while tracking min/max/avg/p95/etc.
func (c *DataDogClient) Histogram(name string, value float64) {
	c.client.Histogram(name, value, c.tagsList(), c.rate)
}

// Distribution tracks the statistical distribution of a set of values.
func (c *DataDogClient) Distribution(name string, value float64) {
	c.client.Distribution(name, value, c.tagsList(), c.rate)
}
