/*
Package misc contains various extensions for gol.
*/
package misc

import "github.com/goburrow/gol"

// AsyncAppender is an appender that runs asynchrously.
type AsyncAppender struct {
	appender gol.Appender
}

var _ (gol.Appender) = (*AsyncAppender)(nil)

// NewAsyncAppender allocates and returns a new AsyncAppender
func NewAsyncAppender(a gol.Appender) *AsyncAppender {
	return &AsyncAppender{a}
}

// Append calls appender
func (a *AsyncAppender) Append(event *gol.LoggingEvent) {
	go a.appender.Append(event)
}
