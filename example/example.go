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
	exampleLogger.Info("Running app with arguments: %v.", os.Args)

	exampleLogger.Warn("Going to do nothing.")
	time.Sleep(1 * time.Second)

	// DefaultLogger is the internal implementation of Logger
	appLogger.(*gol.DefaultLogger).SetLevel(gol.LevelWarn)

	exampleLogger.Info("You won't see this message.")
	appLogger.Error("I %v! %[2]v %[2]v.", "quit", "bye")

	// Output:
	// INFO  [2015-01-14T12:43:35.546+10:00] app/example: Running app with arguments: [/go/bin/example].
	// WARN  [2015-01-14T12:43:35.546+10:00] app/example: Going to do nothing.
	// ERROR [2015-01-14T12:43:36.546+10:00] app: I quit! bye bye.
}
