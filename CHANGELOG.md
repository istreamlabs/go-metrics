# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

- Add `WithRate(float64)` to the metrics interface and to all clients that implement
  the interface. All metrics calls support sample rates.
  - The `LoggerClient`:
    - Applies the sample rate when printing log messages. If the rate is `0.1` and you call `Incr()` ten times, expect about one message to have been printed out.
    - Displays the sample rate for counts if it is not `1.0`, e.g: `Count foo:0.2 (2 * 0.1) [tag1 tag2]`. This shows the sampled value, the passed value, and the sample rate.
    - Gauges, timings, and histograms will show the sample rate, but he value is left unmodified just like the DataDog implementation.
  - The `RecorderClient`:
    - Records all sample rates for metrics calls in `MetricCall.Rate`. No calls are excluded from the call list based on the sample rate, and the value recorded is the full value before multiplying by the sample rate.
    - Adds a `Rate(float64)` query method to filter by sampled metrics.
    - The following should work:

        ```go
        recorder := metrics.NewRecorderClient().WithTest(t)
        recorder.WithRate(0.1).Count("foo", 5)
        recorder.Expect("foo").Rate(0.1).Value(5)
        ```
- Add `Colorized()` method to `LoggerClient`, and automatically detect a TTY and enable color when `nil` is passed to the `NewLoggerClient` constructor.

## [1.2.0] - 2018-03-01

- Update build to test with Go 1.9, drop support for 1.7 since `dep` now
  requires Go 1.8+. Go 1.7 users can still use this library but must manage
  their own dependencies.
- Add `WithRate(rate float)` to the DataDog client to limit traffic sent to
  the `dogstatsd` daemon. The set rate will be applied to all calls made
  with the returned `Client`.
- Automatically assign tags to events.

## [1.1.0] - 2017-08-01

- Make `RecorderClient` goroutine-safe so that metrics can be written and
  checked concurrently. For example:

  ```go
  package main

  import (
    "sync"
    "github.com/istreamlabs/go-metrics/metrics"
  )

  func main() {
    client := metrics.NewRecorderClient()

    wg := sync.WaitGroup{}
    wg.Add(3)
    for i := 0; i < 3; i++ {
      go func() {
        client.Incr("concurrent.access")
        wg.Done()
      }()
    }
    wg.Wait()
  }
  ```

## [1.0.0] - 2017-07-26

- Make project public on GitHub.
