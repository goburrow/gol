/*
Package file provides logging to file.
*/
package file

import (
	"os"
	"sync"

	"github.com/goburrow/gol"
)

// Appender is a file appender with rolling policy.
type Appender struct {
	fileName string

	mu      sync.Mutex
	file    *os.File
	encoder gol.Encoder

	triggeringPolicy TriggeringPolicy
	rollingPolicy    RollingPolicy

	forceStopped bool
}

var _ (gol.Appender) = (*Appender)(nil)

// NewAppender allocates and returns a new Appender.
// Calling Start is only needed for catching errors.
func NewAppender(fileName string) *Appender {
	return &Appender{
		fileName:         fileName,
		encoder:          gol.NewEncoder(),
		triggeringPolicy: NoPolicy,
		rollingPolicy:    NoPolicy,
	}
}

func (a *Appender) Append(event *gol.LoggingEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var err error

	if a.file == nil {
		// Do not auto start once stopped.
		if a.forceStopped {
			return
		}
		if err = a.open(); err != nil {
			gol.Print(err)
			return
		}
	}
	if a.triggeringPolicy.IsTriggering(event, a.file) {
		if err = a.rollingPolicy.Rollover(a.file); err != nil {
			gol.Print(err)
			return
		}
	}
	if err = a.encoder.Encode(event, a.file); err != nil {
		gol.Print(err)
	}
}

// SetEncoder changes the encoder of this appender.
func (a *Appender) SetEncoder(encoder gol.Encoder) {
	a.mu.Lock()
	a.encoder = encoder
	a.mu.Unlock()
}

// SetTriggeringPolicy changes the triggering policy of this appender.
func (a *Appender) SetTriggeringPolicy(policy TriggeringPolicy) {
	a.mu.Lock()
	a.triggeringPolicy = policy
	a.mu.Unlock()
}

// SetRollingPolicy changes the rolling policy of this appender.
func (a *Appender) SetRollingPolicy(policy RollingPolicy) {
	a.mu.Lock()
	a.rollingPolicy = policy
	a.mu.Unlock()
}

func (a *Appender) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.forceStopped = false
	if a.file != nil {
		return nil
	}
	return a.open()
}

func (a *Appender) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.forceStopped = true
	if a.file != nil {
		err := a.file.Close()
		a.file = nil
		return err
	}
	return nil
}

// open must be called with a.mu held.
func (a *Appender) open() error {
	var err error
	a.file, err = os.OpenFile(a.fileName, os.O_RDWR|os.O_CREATE, 0666)
	return err
}
