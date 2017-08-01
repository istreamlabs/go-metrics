# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

- Add items here as needed.

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
