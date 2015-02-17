package syslog

import (
	"testing"

	"github.com/goburrow/gol"
)

func TestSyslog(t *testing.T) {
	appender := NewAppender("test")
	err := appender.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer appender.Stop()

	event := &gol.LoggingEvent{
		FormattedMessage: "message",
		Level:            gol.LevelInfo,
	}
	appender.Append(event)
}
