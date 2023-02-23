package metrics

import (
	"log"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
)

// DataDogClient is a dogstatsd metrics client implementation.
type DataDogClient struct {
	client *statsd.Client
	rate   float64
	tags   []string
}

// Options contains the configuration options for a client.
type Options struct {
	WithoutTelemetry bool
	Statsd           *statsd.Client
}

// Option is a client option. Can return an error if validation fails.
type Option func(*Options) error

// WithoutTelemetry turns off sending DataDog telemetry metrics.
func WithoutTelemetry() Option {
	return func(o *Options) error {
		o.WithoutTelemetry = true
		return nil
	}
}

func WithStatsd(s *statsd.Client) Option {
	return func(o *Options) error {
		o.Statsd = s
		return nil
	}
}

func resolveOptions(options []Option) (*Options, error) {
	o := &Options{
		WithoutTelemetry: false,
	}

	for _, option := range options {
		err := option(o)
		if err != nil {
			return nil, err
		}
	}
	return o, nil
}

// NewDataDogClient creates a new dogstatsd client pointing to `address` with
// the metrics prefix of `namespace`. For example, given a namespace of
// `foo.bar`, a call to `Incr('baz')` would emit a metric with the full name
// `foo.bar.baz` (note the period between the namespace and metric name).
func NewDataDogClient(address string, namespace string, options ...Option) *DataDogClient {
	o, err := resolveOptions(options)
	if err != nil {
		log.Panic(err)
	}

	var opts []statsd.Option
	if o.WithoutTelemetry {
		opts = append(opts, statsd.WithoutTelemetry())
	}
	if namespace != "" {
		opts = append(opts, statsd.WithNamespace(namespace))
	}

	var c *statsd.Client
	if o.Statsd != nil {
		c = o.Statsd
	} else {
		c, err = statsd.New(address, opts...)
		if err != nil {
			log.Panic(err)
		}
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
		tags:   c.tags, // clone isn't necessary since original slice is immutable
	}
}

// WithTags clones this client with additional tags. Duplicate tags overwrite
// the existing value.
func (c *DataDogClient) WithTags(tags map[string]string) Client {
	return &DataDogClient{
		client: c.client,
		rate:   c.rate,
		tags:   cloneTagsWithMap(c.tags, tags),
	}
}

// WithoutTelemetry clones this client with telemetry stats turned off. Underlying
// DataDog statsd client only supports turning off telemetry, which is on by default.
func (c *DataDogClient) WithoutTelemetry() Client {
	s, err := statsd.CloneWithExtraOptions(c.client, statsd.WithoutTelemetry())
	if err != nil {
		log.Panic(err)
	}
	return &DataDogClient{
		client: s,
		rate:   c.rate,
		tags:   c.tags, // clone isn't necessary since original slice is immutable
	}
}

// Close closes all client connections and flushes any buffered data.
func (c *DataDogClient) Close() error {
	return c.client.Close()
}

// Count adds some integer value to a metric.
func (c *DataDogClient) Count(name string, value int64) {
	c.client.Count(name, value, c.tags, c.rate)
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
	c.client.Gauge(name, value, c.tags, c.rate)
}

// Event tracks an event that may be relevant to other metrics.
func (c *DataDogClient) Event(e *statsd.Event) {
	if len(c.tags) > 0 {
		e.Tags = append(e.Tags, c.tags...)
	}

	c.client.Event(e)
}

// Timing tracks a duration.
func (c *DataDogClient) Timing(name string, value time.Duration) {
	c.client.Timing(name, value, c.tags, c.rate)
}

// Histogram sets a numeric value while tracking min/max/avg/p95/etc.
func (c *DataDogClient) Histogram(name string, value float64) {
	c.client.Histogram(name, value, c.tags, c.rate)
}

// Distribution tracks the statistical distribution of a set of values.
func (c *DataDogClient) Distribution(name string, value float64) {
	c.client.Distribution(name, value, c.tags, c.rate)
}
