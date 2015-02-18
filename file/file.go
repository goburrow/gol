/*
Package file provides logging to file.
*/
package file

import (
	"os"
	"sync"

	"github.com/goburrow/gol"
)

type TriggeringPolicy interface {
	IsTriggering(*gol.LoggingEvent, *os.File) bool
}

type RollingPolicy interface {
	Rollover() error
}

type noTriggeringPolicy struct {
}

func (p *noTriggeringPolicy) IsTriggering(*gol.LoggingEvent, *os.File) bool {
	return false
}

type noRollingPolicy struct {
}

func (p *noRollingPolicy) Rollover() error {
	return nil
}

var (
	NoTriggeringPolicy (TriggeringPolicy) = &noTriggeringPolicy{}
	NoRollingPolicy    (RollingPolicy)    = &noRollingPolicy{}
)

type Appender struct {
	Encoder gol.Encoder

	TriggeringPolicy TriggeringPolicy
	RollingPolicy    RollingPolicy

	fileName string

	mu   sync.Mutex
	file *os.File
}

var _ (gol.Appender) = (*Appender)(nil)

func NewAppender(fileName string) *Appender {
	return &Appender{
		Encoder:          gol.NewEncoder(),
		TriggeringPolicy: NoTriggeringPolicy,
		RollingPolicy:    NoRollingPolicy,
		fileName:         fileName,
	}
}

func (a *Appender) Append(event *gol.LoggingEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var err error
	if a.file == nil {
		if err = a.open(); err != nil {
			println("gol/file:", err.Error())
			return
		}
	}
	a.Encoder.Encode(event, a.file)
	if a.TriggeringPolicy.IsTriggering(event, a.file) {
		if err = a.RollingPolicy.Rollover(); err != nil {
			println("gol/file:", err.Error())
		}
	}
}

func (a *Appender) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.file != nil {
		return nil
	}
	return a.open()
}

func (a *Appender) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.file != nil {
		err := a.file.Close()
		a.file = nil
		return err
	}
	return nil
}

func (a *Appender) open() error {
	var err error
	a.file, err = os.OpenFile(a.fileName, os.O_RDWR|os.O_CREATE, 0666)
	return err
}
