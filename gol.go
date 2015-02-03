/*
Package gol provides a simple logging framework.
*/
package gol

import (
	"os"
)

// StaticLoggerFactory is the default logger factory that is used by GetLogger().
// It is made public so that users can override it their package's init().
var StaticLoggerFactory = NewLoggerFactory(os.Stdout)

// GetLogger returns Logger in the default logger factory.
func GetLogger(name string) Logger {
	return StaticLoggerFactory.GetLogger(name)
}
