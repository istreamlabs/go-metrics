# Metrics Library for Go

[![Travis](https://img.shields.io/travis/istreamlabs/go-metrics.svg)](https://travis-ci.org/istreamlabs/go-metrics) [![Codecov](https://img.shields.io/codecov/c/github/istreamlabs/go-metrics.svg)](https://codecov.io/gh/istreamlabs/go-metrics) [![Documentation](https://img.shields.io/badge/godoc-reference-blue.svg)](https://godoc.org/github.com/istreamlabs/go-metrics/metrics) [![GitHub tag](https://img.shields.io/github/tag/istreamlabs/go-metrics.svg)](https://github.com/istreamlabs/go-metrics/releases) [![License](https://img.shields.io/github/license/istreamlabs/go-metrics.svg)](https://github.com/istreamlabs/go-metrics/blob/master/LICENSE)

A library that provides an interface for sending metrics from your script, application, or service. Metrics consist of a name, value, and optionally some tags that are made up of key/value pairs.

Multiple implementations are provided to allow for production, local development, and testing use cases. The following clients are available:

Client           | Description
---------------- | -----------
`LoggerClient`   | Writes metrics into a log stream. Useful when running locally.
`DataDogClient`  | Writes metrics into DataDog. Useful for production.
`NullClient`     | Acts like a mock that does nothing. Useful for testing.
`RecorderClient` | Writes metrics into memory and provides a query interface. Useful for testing.

## Example Usage

Generally you will instantiate one of the above clients and then write metrics to it. First, install the library:

```sh
# If using dep, glide, godep, etc then use that tool. For example:
dep ensure github.com/istreamlabs/go-metrics

# Otherwise, just go get it:
go get github.com/istreamlabs/go-metrics/metrics
```

Then you can use it:

```go
import "github.com/istreamlabs/go-metrics/metrics"

var client metrics.Client
if os.Getenv("env") == "prod" {
  client = metrics.NewDataDogClient("127.0.0.1:8125", "myprefix")
} else {
  // Log to standard out instead of sending production metrics.
  client = metrics.NewLoggerClient(nil)
}

// Simple incrementing counter
client.Incr("requests.count")

// Tagging with counters
client.WithTags(map[string]string{
  "tag": "value"
}).Incr("requests.count")
```

The above code would result in `myprefix.requests.count` with a value of `1` showing up in DataDog if you have [`dogstatsd`](https://docs.datadoghq.com/guides/dogstatsd/) running locally and an environment variable `env` set to `prod`, otherwise it will print metrics to standard out. See the [`Client`](https://godoc.org/github.com/istreamlabs/go-metrics/metrics/#Client) interface for a list of available metrics methods.

Sometimes you wouldn't want to send a metric every single time a piece of code is executed. This is supported by setting a sample rate:

```go
// Sample rate for high-throughput applications
client.WithRate(0.01).Incr("requests.count")
```

Sample rates apply to metrics but not events. Any count-type metric (`Incr`, `Decr`, `Count`, and timing/histogram counts) will get multiplied to the full value, while gauges are sent unmodified. For example, when emitting a 10% sampled timing metric that takes an average of `200ms` to DataDog, you would see `1 call * (1/0.1 sample rate) = 10 calls` added to the histogram count while the average value remains `200ms` in the DataDog UI.

Also provided are useful clients for testing. For example, the following asserts that a metric with the given name, value, and tag was emitted during a test:

```go
func TestFoo(t *testing.T)
  client := metrics.NewRecorderClient().WithTest(t)

  client.WithTags(map[string]string{
    "tag": "value",
  }).Count("requests.count", 1)

  // Now, assert that the metric was emitted.
  client.
    Expect("requests.count").
    Value(1).
    Tag("tag", "value")
}
```

For more information and examples, see the [godocs](https://godoc.org/github.com/istreamlabs/go-metrics/metrics).

## License

Copyright &copy; 2017 iStreamPlanet Co., LLC

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License.
