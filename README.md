# gol [![Build Status](https://travis-ci.org/goburrow/gol.svg)](https://travis-ci.org/goburrow/gol) [![Coverage Status](https://coveralls.io/repos/goburrow/gol/badge.svg)](https://coveralls.io/r/goburrow/gol)
Go logging made simple

## Introduction
gol (or golog) provides a generic logging API and a simple implementation which
supports logging level and hierarchy.

The [`Logger` interface](https://github.com/goburrow/gol/blob/master/api.go)
is quite minimal and does not allow you to set level directly but
the `DefaultLogger`, its default implementation, does.
You can also create a Logger hierarchy with the `DefaultLogger`.
A logger "a.b" will inherit logging level, layouter and appender from logger "a"
unless its own properties are set.

## Example
See [example/example.go](https://github.com/goburrow/gol/blob/master/example/example.go)

```go
package main

import (
	"github.com/goburrow/gol"
	"os"
	"time"
)

func init() {
    // Override the default logger if needed, e.g.
    // gol.StaticLoggerFactory = gol.NewLoggerFactory(os.Stderr)
}

func main() {
	// Get logger with name "example"
	logger := gol.GetLogger("example")
	logger.Info("Running app with arguments: %v.", os.Args)

	logger.Warn("Going to do nothing.")
	time.Sleep(1 * time.Second)

	// Root logger is what other loggers inherit from
	rootLogger := gol.GetLogger(gol.RootLoggerName)
	// DefaultLogger is an internal implementation of Logger
	rootLogger.(*gol.DefaultLogger).SetLevel(gol.LevelWarn)

	logger.Info("You won't see this message.")
	rootLogger.Error("I %v! %[2]v %[2]v.", "quit", "bye")

	// Output:
	// INFO  [2015-01-14T12:43:35.546+10:00] example: Running app with arguments: [/go/bin/example].
	// WARN  [2015-01-14T12:43:35.546+10:00] example: Going to do nothing.
	// ERROR [2015-01-14T12:43:36.546+10:00] ROOT: I quit! bye bye.
}
```
