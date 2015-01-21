// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.
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
