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
type Appender struct {
	// DrainTimeout is maximum duration before timing out flush a channel.
	DrainTimeout time.Duration

	wg        sync.WaitGroup
	appenders []gol.Appender
	chans     []chan *gol.LoggingEvent

	mu     sync.Mutex
	finish chan struct{}

	forceStopped bool
}

var _ (gol.Appender) = (*Appender)(nil)

// NewAppender allocates and returns a new Appender.
func NewAppender(bufferSize int, appenders ...gol.Appender) *Appender {
	a := &Appender{
		appenders: appenders,
		chans:     make([]chan *gol.LoggingEvent, len(appenders)),

		DrainTimeout: 10 * time.Second,
	}
	for i, _ := range appenders {
		a.chans[i] = make(chan *gol.LoggingEvent, bufferSize)
	}
	return a
}

// Append sends the event to all appenders.
func (a *Appender) Append(e *gol.LoggingEvent) {
	// Need to use mutex here to make sure messages are in correct order
	// for all channels.
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.finish == nil {
		if a.forceStopped {
			// Skip the event if appender is stopped.
			return
		}
		a.start()
	}
	for _, c := range a.chans {
		// FIXME: This is still blocking if a channel buffer is full.
		c <- e
	}
}

// Start does not do anything at the moment.
func (a *Appender) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.forceStopped = false
	if a.finish != nil {
		return nil // started
	}
	a.start()
	return nil
}

func (a *Appender) start() {
	a.finish = make(chan struct{})
	a.wg.Add(len(a.chans))
	for i, c := range a.chans {
		go a.receive(c, a.appenders[i])
	}
}

// Stop makes sure all appenders finish. It must be called when exiting.
func (a *Appender) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.forceStopped = true
	if a.finish == nil {
		return nil // not started
	}
	close(a.finish)
	a.wg.Wait()
	a.finish = nil
	return nil
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

func (a *Appender) flush(c chan *gol.LoggingEvent, appender gol.Appender) {
	timeout := time.After(a.DrainTimeout)
	for {
		select {
		case <-timeout:
			return
		case e := <-c:
			appender.Append(e)
		default:
			// Channel is empty.
			return
		}
	}
}
