/*
Package file provides logging to syslog.
*/
package syslog

import (
	"bytes"
	"log/syslog"
	"sync"

	"github.com/goburrow/gol"
)

const defaultLayout = "%[2]s: %[1]s"

type Appender struct {
	Encoder gol.Encoder

	Network string
	RAddr   string
	Tag     string

	mu     sync.Mutex
	target *syslog.Writer
}

var _ gol.Appender = (*Appender)(nil)

func NewAppender(tag string) *Appender {
	encoder := gol.NewEncoder()
	encoder.Layout = defaultLayout
	return &Appender{
		Encoder: encoder,
		Tag:     tag,
	}
}

func (a *Appender) Append(event *gol.LoggingEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var err error
	if a.target == nil {
		if err := a.connect(); err != nil {
			// FIXME: displaying error
			println("gol/syslog:", err.Error())
			return
		}
	}
	var buf bytes.Buffer
	a.Encoder.Encode(event, &buf)

	if event.Level >= gol.LevelError {
		err = a.target.Err(buf.String())
	} else if event.Level >= gol.LevelWarn {
		err = a.target.Warning(buf.String())
	} else if event.Level >= gol.LevelInfo {
		err = a.target.Info(buf.String())
	} else {
		err = a.target.Debug(buf.String())
	}
	if err != nil {
		println("gol/syslog:", err.Error())
	}
}

// Start connects to syslog server if not connected.
func (a *Appender) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.target != nil {
		return nil
	}
	return a.connect()
}

// Stop disconnects current connection.
func (a *Appender) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.target != nil {
		err := a.target.Close()
		a.target = nil
		return err
	}
	return nil
}

// connect must be called with a.mu held.
func (a *Appender) connect() error {
	var err error
	a.target, err = syslog.Dial(a.Network, a.RAddr, syslog.LOG_WARNING, a.Tag)
	return err
}
