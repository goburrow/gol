/*
Package async provides an asynchronous appender for gol.
*/
package async

import (
	"sync"
	"time"

	"github.com/goburrow/gol"
)

// Appender sends the logging event to all appenders asynchronously.
// It implements gol.Appender.
type Appender struct {
	// drainTimeout is maximum duration before timing out flush a channel.
	drainTimeout time.Duration

	wg        sync.WaitGroup
	appenders []gol.Appender
	chans     []chan *gol.LoggingEvent

	// started in an indicator for this appender state.
	started bool
	// finish is used in Start and Stop
	finish chan struct{}
}

// NewAppender allocates and returns a new Appender.
// Start() must be called before writing data.
func NewAppender(appenders ...gol.Appender) *Appender {
	return NewAppenderWithBufSize(10, appenders...)
}

// NewAppenderWithBufSize returns an asynchronous appender with a buffer size
// of bufSize for each appender channel.
func NewAppenderWithBufSize(bufSize int, appenders ...gol.Appender) *Appender {
	a := &Appender{
		appenders:    appenders,
		drainTimeout: 10 * time.Second,
	}
	a.chans = make([]chan *gol.LoggingEvent, len(appenders))
	for i := range appenders {
		a.chans[i] = make(chan *gol.LoggingEvent, bufSize)
	}
	return a
}

// Append sends the event to all appenders.
func (a *Appender) Append(e *gol.LoggingEvent) {
	if !a.started {
		// Skip the event if appender is stopped.
		return
	}
	for _, c := range a.chans {
		// FIXME: This is still blocking if a channel buffer is full.
		c <- e
	}
}

// Start starts go routines for each writer.
func (a *Appender) Start() {
	if a.started {
		return
	}
	a.finish = make(chan struct{})
	a.wg.Add(len(a.chans))
	for i, c := range a.chans {
		go a.receive(c, a.appenders[i])
	}
	a.started = true
}

// Stop stops and waits until all go routines exited.
// Once Stop is called, this appender can not be started again.
func (a *Appender) Stop() {
	if !a.started {
		return
	}
	a.started = false
	close(a.finish)
	a.wg.Wait()
	// all channels can be closed now, but then users can not start this
	// appender again.
}

func (a *Appender) receive(c chan *gol.LoggingEvent, appender gol.Appender) {
	defer a.wg.Done()

	for {
		select {
		case <-a.finish:
			a.flush(c, appender)
			return
		case e := <-c:
			appender.Append(e)
		}
	}
}

// flush sends all pending data in the channel to writer or timeout after
// maximum of durationTimeout and the appender timeout.
func (a *Appender) flush(c chan *gol.LoggingEvent, appender gol.Appender) {
	timeout := time.After(a.drainTimeout)
	for {
		select {
		case <-timeout:
			// Timeout channel has higher priority.
			return
		case e := <-c:
			appender.Append(e)
			// Continue reading from the appender.
		default:
			// Channel is empty.
			return
		}
	}
}
