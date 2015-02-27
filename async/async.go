/*
Package async provides an asynchronous appender for gol.
*/
package async

import (
	"sync"

	"github.com/goburrow/gol"
)

// Appender sends the logging event to all appenders asynchronously.
type Appender struct {
	appenders []gol.Appender
	wg        sync.WaitGroup
}

var _ (gol.Appender) = (*Appender)(nil)

// NewAppender allocates and returns a new Appender
func NewAppender(appender ...gol.Appender) *Appender {
	return &Appender{appenders: appender}
}

// Append calls all appenders
func (a *Appender) Append(e *gol.LoggingEvent) {
	for _, appender := range a.appenders {
		a.wg.Add(1)
		go a.appendTo(e, appender)
	}
}

// Start does not do anything at the moment.
func (a *Appender) Start() error {
	return nil
}

// Stop makes sure all appenders finish. It must be called when exiting.
func (a *Appender) Stop() error {
	a.wg.Wait()
	return nil
}

func (a *Appender) appendTo(e *gol.LoggingEvent, appender gol.Appender) {
	defer a.wg.Done()
	appender.Append(e)
}
