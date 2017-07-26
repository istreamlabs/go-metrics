# Metrics Library for Go

[![Travis](https://img.shields.io/travis/istreamlabs/go-metrics.svg)](https://travis-ci.org/istreamlabs/go-metrics) [![Codecov](https://img.shields.io/codecov/c/github/istreamlabs/go-metrics.svg)](https://codecov.io/gh/istreamlabs/go-metrics) [![GitHub tag](https://img.shields.io/github/tag/istreamlabs/go-metrics.svg)](https://github.com/istreamlabs/go-metrics/releases) [![License](https://img.shields.io/github/license/istreamlabs/go-metrics.svg)](https://github.com/istreamlabs/go-metrics/blob/master/LICENSE)

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
go get github.com/istreamlabs/go-metrics
```

Then you can use it:

```go
import "github.com/istreamlabs/go-metrics"

client := metrics.NewDataDogClient("127.0.0.1:8125", "myprefix")

// Simple incrementing counter
client.Incr("requests.count")

// Tagging with counters
client.WithTags(map[string]string{
  "tag": "value"
}).Incr("requests.count")
```

The above code would result in `myprefix.requests.count` with a value of `1` showing up in DataDog if you have [`dogstatsd`](https://docs.datadoghq.com/guides/dogstatsd/) running locally. See the [`Client`](https://godoc.org/github.com/istreamlabs/go-metrics/#Client) interface for a list of available metrics methods.

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

For more information and examples, see the [godocs](https://godoc.org/github.com/istreamlabs/go-metrics/).

## License

Copyright &copy; 2017 iStreamPlanet Co., LLC

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License.
