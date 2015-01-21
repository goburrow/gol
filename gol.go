package gol

import (
	"os"
)

var defaultLoggerFactory = NewLoggerFactory(os.Stdout)

// GetLogger returns Logger in the default logger factory
func GetLogger(name string) Logger {
	return defaultLoggerFactory.GetLogger(name)
}
