# gol [![Build Status](https://travis-ci.org/goburrow/gol.svg)](https://travis-ci.org/goburrow/gol) [![Coverage Status](https://coveralls.io/repos/goburrow/gol/badge.svg)](https://coveralls.io/r/goburrow/gol)
Go logging made simple

## Introduction
gol (or golog) provides a generic logging API and a simple implementation which
supports logging level.

The [`Logger` interface](https://github.com/goburrow/gol/blob/master/api.go) is quite minimal and does not allow you to set level
directly but the `DefaultLogger`, its default implementation, does. In theory,
you can also create a Logger hierarchy with the `DefaultLogger`. I think it is
necessary for applications in production as we might need to change logging
level for components individually. However, I'm too lazy to support that in the
`DefaultLoggerFactory` at the moment. Actually, the reason is that I have not
decided what the behaviour if an user provides a "invalid" name, such as
".a^.^b." or "a...b" (Pull requests are welcomed).

## Example
See [example/example.go](https://github.com/goburrow/gol/blob/master/example/example.go)

```go
package main

import (
	"github.com/goburrow/gol"
	"os"
	"time"
)

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
