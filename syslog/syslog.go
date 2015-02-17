/*
Package file provides logging to syslog.
*/
package syslog

import (
	"fmt"
	"log/syslog"
	"sync"

	"github.com/goburrow/gol"
)

const defaultLayout = "%[2]s: %[1]s"

type SyslogAppender struct {
	Network string
	RAddr   string
	Tag     string
	Layout  string

	mu     sync.Mutex
	writer *syslog.Writer
}

var _ gol.Appender = (*SyslogAppender)(nil)

func NewAppender(tag string) *SyslogAppender {
	return &SyslogAppender{
		Tag:    tag,
		Layout: defaultLayout,
	}
}

func (a *SyslogAppender) Append(event *gol.LoggingEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var err error
	if a.writer == nil {
		if err = a.connect(); err != nil {
			// FIXME: displaying error
			println(err.Error())
			return
		}
	}
	// FIXME: Syslog message is not standardized
	msg := fmt.Sprintf(a.Layout,
		event.FormattedMessage,
		event.Name)

	if event.Level >= gol.LevelError {
		err = a.writer.Err(msg)
	} else if event.Level >= gol.LevelWarn {
		err = a.writer.Warning(msg)
	} else if event.Level >= gol.LevelInfo {
		err = a.writer.Info(msg)
	} else {
		err = a.writer.Debug(msg)
	}
	if err != nil {
		println(err.Error())
	}
}

func (a *SyslogAppender) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.connect()
}

func (a *SyslogAppender) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.close()
}

// connect must be called with a.mu held.
func (a *SyslogAppender) connect() error {
	var err error
	a.writer, err = syslog.Dial(a.Network, a.RAddr, syslog.LOG_WARNING, a.Tag)
	return err
}

// close must be called with a.mu held.
func (a *SyslogAppender) close() error {
	if a.writer != nil {
		err := a.writer.Close()
		a.writer = nil
		return err
	}
	return nil
}
