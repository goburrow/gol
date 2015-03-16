/*
Package log provides a bridge of default go log to gol.
*/
package log

import (
	"bytes"
	"log"

	"github.com/goburrow/gol"
)

func init() {
	log.SetFlags(log.Lshortfile)
	log.SetOutput(writerFunc(write))
}

type writerFunc func([]byte) (int, error)

func (f writerFunc) Write(b []byte) (int, error) {
	return f(b)
}

// write prints log message as level INFO.
func write(b []byte) (int, error) {
	logger := gol.GetLogger(getName(b))
	logger.Info(getMessage(b))
	return len(b), nil
}

// getName returns file name in the log message if a colon is found. Otherwise,
// gol.RootLoggerName will be used.
func getName(b []byte) string {
	// Take logger name  main.go:6: my message
	idx := bytes.IndexByte(b, byte(':'))
	if idx < 0 {
		return gol.RootLoggerName
	}
	return string(b[:idx])
}

// getMessage returns the message string stripping file name and line number and
// line termination.
func getMessage(b []byte) string {
	idx := bytes.Index(b, []byte{':', ' '})
	if idx < 0 {
		idx = -2
	}
	return string(b[idx+2 : len(b)-1])
}
