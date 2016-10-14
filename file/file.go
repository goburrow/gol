/*
Package file provides logging to file.
*/
package file

import (
	"sync"

	"github.com/goburrow/gol"
	"github.com/goburrow/gol/file/rotation"
)

// Appender is a file appender with rolling policy.
type Appender struct {
	mu   sync.Mutex
	file *rotation.File

	appender gol.Appender
}

var _ (gol.Appender) = (*Appender)(nil)

// NewAppender allocates and returns a new Appender.
// Calling Start is only needed for catching errors.
func NewAppender(fileName string) *Appender {
	file := rotation.NewFile(fileName)

	return &Appender{
		file:     file,
		appender: gol.NewAppender(file),
	}
}

// Append encodes the given logging event to file.
func (a *Appender) Append(event *gol.LoggingEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.file.IsOpenned() {
		a.appender.Append(event)
	}
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

// Start opens the log file.
func (a *Appender) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.file.IsOpenned() {
		return nil
	}
	return a.file.Open()
}

// Stop closes the log file.
func (a *Appender) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.file.Close()
}
