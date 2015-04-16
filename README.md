# gol [![Build Status](https://travis-ci.org/goburrow/gol.svg)](https://travis-ci.org/goburrow/gol) [![GoDoc](https://godoc.org/github.com/goburrow/gol?status.svg)](https://godoc.org/github.com/goburrow/gol) [![Coverage Status](https://coveralls.io/repos/goburrow/gol/badge.svg?branch=master)](https://coveralls.io/r/goburrow/gol?branch=master)
Go logging made simple

## Introduction
gol (or golog) provides a generic logging API and a simple implementation which
supports logging level and hierarchy.

The [`Logger` interface](https://github.com/goburrow/gol/blob/master/api.go)
is kept minimal and does not allow you to set level directly but
the `DefaultLogger`, its default implementation, does.
You can also create a Logger hierarchy with the `DefaultLogger`.
For example, logger `a/b/c` will inherit logging level and appender from logger `a/b`
unless its own properties are set.

## Example
See [example/example.go](https://github.com/goburrow/gol/blob/master/example/example.go)

```go
package main

import (
	"os"
	"time"

	"github.com/goburrow/gol"
)

var exampleLogger, appLogger gol.Logger

func init() {
	// Get logger with name "app/example"
	exampleLogger = gol.GetLogger("app/example")
	// Logger "app" is the parent of the logger "app/example"
	appLogger = gol.GetLogger("app")
}

func main() {
	exampleLogger.Infof("Running app with arguments: %v.", os.Args)

	exampleLogger.Warnf("Going to do nothing.")
	time.Sleep(1 * time.Second)

	// DefaultLogger is the internal implementation of Logger
	appLogger.(*gol.DefaultLogger).SetLevel(gol.LevelWarn)

	exampleLogger.Infof("You won't see this message.")
	appLogger.Errorf("I %v! %[2]v %[2]v.", "quit", "bye")

	// Output:
	// INFO  [2015-01-14T12:43:35.546+10:00] app/example: Running app with arguments: [/go/bin/example].
	// WARN  [2015-01-14T12:43:35.546+10:00] app/example: Going to do nothing.
	// ERROR [2015-01-14T12:43:36.546+10:00] app: I quit! bye bye.
}
```
