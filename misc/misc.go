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

// ThresholdAppender has lowest level of event to append to the parent appender.
type ThresholdAppender struct {
	Threshold gol.Level
	appender  gol.Appender
}

var _ (gol.Appender) = (*ThresholdAppender)(nil)

// NewThresholdAppender allocates and returns a new ThresholdAppender
func NewThresholdAppender(a gol.Appender) *ThresholdAppender {
	return &ThresholdAppender{
		Threshold: gol.LevelAll,
		appender:  a,
	}
}

// Append calls appender
func (a *ThresholdAppender) Append(event *gol.LoggingEvent) {
	if event.Level >= a.Threshold {
		a.appender.Append(event)
	}
}
