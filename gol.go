/*
Package gol provides a simple logging framework.
*/
package gol

import (
	"fmt"
	"os"
	"runtime"
)

var (
	// staticLoggerFactory is the default logger factory that is used by GetLogger().
	staticLoggerFactory = NewLoggerFactory(os.Stdout)
	// debugMode allows Print to write results to standard error.
	debugMode = false
)

// GetLogger returns Logger in the default logger factory.
func GetLogger(name string) Logger {
	return staticLoggerFactory.GetLogger(name)
}

func SetDebugMode(val bool) {
	debugMode = val
}

// Print prints to standard error, used for debugging.
func Print(args ...interface{}) {
	if !debugMode {
		return
	}
	var (
		file string
		line int
		ok   bool
	)
	_, file, line, ok = runtime.Caller(1)
	if ok {
		fmt.Fprintf(os.Stderr, "%s:%d: ", file, line)
	}
	fmt.Fprintln(os.Stderr, args...)
}
