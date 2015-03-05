/*
Package file provides logging to file.
*/
package file

import (
	"os"
	"sync"

	"github.com/goburrow/gol"
	"github.com/goburrow/gol/file/rotation"
)

const (
	openFlag = os.O_RDWR | os.O_CREATE | os.O_APPEND
	openMode = 0644
)

// Appender is a file appender with rolling policy.
type Appender struct {
	mu      sync.Mutex
	file    *rotation.File
	encoder gol.Encoder

	forceStopped bool
}

var _ (gol.Appender) = (*Appender)(nil)

// NewAppender allocates and returns a new Appender.
// Calling Start is only needed for catching errors.
func NewAppender(fileName string) *Appender {
	return &Appender{
		file:    rotation.NewFile(fileName),
		encoder: gol.NewEncoder(),
	}
}

func (a *Appender) Append(event *gol.LoggingEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var err error

	if !a.file.IsOpenned() {
		// Do not auto start once stopped.
		if a.forceStopped {
			return
		}
		if err = a.open(); err != nil {
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
func (a *Appender) SetTriggeringPolicy(policy rotation.TriggeringPolicy) {
	a.mu.Lock()
	a.file.SetTriggeringPolicy(policy)
	a.mu.Unlock()
}

// SetRollingPolicy changes the rolling policy of this appender.
func (a *Appender) SetRollingPolicy(policy rotation.RollingPolicy) {
	a.mu.Lock()
	a.file.SetRollingPolicy(policy)
	a.mu.Unlock()
}

func (a *Appender) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.forceStopped = false
	if a.file.IsOpenned() {
		return nil
	}
	return a.open()
}

func (a *Appender) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.forceStopped = true
	return a.file.Close()
}

// open must be called with a.mu held.
func (a *Appender) open() error {
	return a.file.Open(openFlag, openMode)
}
