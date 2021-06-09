package metrics

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"sort"
	"time"

	"github.com/DataDog/datadog-go/statsd"
	"github.com/mattn/go-isatty"
	"github.com/mgutz/ansi"
)

var (
	// Colors are from the ANSI 256 color pallette.
	// https://en.wikipedia.org/wiki/ANSI_escape_code#8-bit
	cname    = ansi.ColorFunc("208")
	cvalue   = ansi.ColorFunc("32")
	crate    = ansi.ColorFunc("106")
	csampled = ansi.ColorFunc("43")
	ctag     = ansi.ColorFunc("133")
)

// InfoLogger provides a method for logging info messages and is implemented
// by the standard `log` package as well as various other packages.
type InfoLogger interface {
	Printf(format string, args ...interface{})
}

// LoggerClient simple dumps metrics into the log. Useful when running
// locally for testing. Can be used with multiple different logging systems.
type LoggerClient struct {
	logger InfoLogger
	colors bool
	rate   float64
	tagMap map[string]string
}

// NewLoggerClient creates a new logging client. If `logger` is `nil` then it
// defaults to stdout using the built-in `log` package. It is equivalent to
// the following with added auto-detection for colorized output:
//
//   metrics.NewLoggerClient(log.New(os.Stdout, "", 0))
//
// You can use your own logger and enable colorized output manually via:
//
//   metrics.NewLoggerClient(myLog).Colorized()
func NewLoggerClient(logger InfoLogger) *LoggerClient {
	colors := false
	if logger == nil {
		logger = log.New(os.Stdout, "", 0)

		if isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd()) {
			colors = true
		}
	}

	client := &LoggerClient{
		logger: logger,
		colors: colors,
		rate:   1.0,
	}

	return client
}

// Colorized enables colored terminal output.
func (c *LoggerClient) Colorized() *LoggerClient {
	return &LoggerClient{
		logger: c.logger,
		rate:   c.rate,
		colors: true,
		tagMap: c.tagMap,
	}
}

// WithTags clones this client with additional tags. Duplicate tags overwrite
// the existing value.
func (c *LoggerClient) WithTags(tags map[string]string) Client {
	return &LoggerClient{
		logger: c.logger,
		rate:   c.rate,
		colors: c.colors,
		tagMap: combine(c.tagMap, tags),
	}
}

// WithRate clones this client with a given sample rate. Subsequent calls
// will be limited to logging metrics at this rate.
func (c *LoggerClient) WithRate(rate float64) Client {
	return &LoggerClient{
		logger: c.logger,
		rate:   rate,
		colors: c.colors,
		tagMap: c.tagMap,
	}
}

// print out the metric call, taking into account sample rate.
func (c *LoggerClient) print(t string, name string, value interface{}, sampled interface{}) {
	r := fmt.Sprintf("%v", c.rate)
	v := value
	s := sampled

	if c.colors {
		name = cname(name)
		r = crate(r)
		v = cvalue(fmt.Sprintf("%v", value))
		s = csampled(fmt.Sprintf("%v", sampled))
	}

	if c.rate == 1.0 {
		c.logger.Printf("%s %s:%v %v", t, name, v, c.getTags())
		return
	}

	if rand.Float64() < c.rate {
		if value == sampled {
			c.logger.Printf("%s %s:%v (%v) %v", t, name, v, r, c.getTags())
		} else {
			c.logger.Printf("%s %s:%v (%v * %v) %v", t, name, s, v, r, c.getTags())
		}
	}
}

func (c *LoggerClient) getTags() string {
	if !c.colors {
		return fmt.Sprintf("%v", c.tagMap)
	}

	keys := make([]string, 0, len(c.tagMap))
	for k := range c.tagMap {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	tags := ""
	for _, key := range keys {
		if tags != "" {
			tags += " "
		}

		tags += fmt.Sprintf("%s:%s", ctag(key), c.tagMap[key])
	}

	return "map[" + tags + "]"
}

// Close on LoggerClient is a no-op
func (c *LoggerClient) Close() error {
	return nil
}

// Count adds some value to a metric.
func (c *LoggerClient) Count(name string, value int64) {
	c.print("Count", name, value, float64(value)*c.rate)
}

// Incr adds one to a metric.
func (c *LoggerClient) Incr(name string) {
	c.Count(name, 1)
}

// Decr subtracts one from a metric.
func (c *LoggerClient) Decr(name string) {
	c.Count(name, -1)
}

// Gauge sets a numeric value.
func (c *LoggerClient) Gauge(name string, value float64) {
	c.print("Gauge", name, value, value)
}

// Event tracks an event that may be relevant to other metrics.
func (c *LoggerClient) Event(e *statsd.Event) {
	c.logger.Printf("Event %s\n%s %v", e.Title, e.Text, c.tagMap)
}

// Timing tracks a duration.
func (c *LoggerClient) Timing(name string, value time.Duration) {
	c.print("Timing", name, value, value)
}

// Histogram sets a numeric value while tracking min/max/avg/p95/etc.
func (c *LoggerClient) Histogram(name string, value float64) {
	c.print("Histogram", name, value, value)
}

// Distribution tracks the statistical distribution of a set of values.
func (c *LoggerClient) Distribution(name string, value float64) {
	c.print("Distribution", name, value, value)
}
