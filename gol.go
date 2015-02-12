/*
Package gol provides a simple logging framework.
*/
package gol

import (
	"os"
)

// staticLoggerFactory is the default logger factory that is used by GetLogger().
var staticLoggerFactory = NewLoggerFactory(os.Stdout)

// SetLoggerFactory changes the default staticLoggerFactory.
// This method should only be called in package's init().
func SetLoggerFactory(f LoggerFactory) {
	staticLoggerFactory = f
}

// GetLogger returns Logger in the default logger factory.
func GetLogger(name string) Logger {
	return staticLoggerFactory.GetLogger(name)
}
