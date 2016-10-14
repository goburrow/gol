package syslog

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/goburrow/gol"
)

type bufNopCloser struct {
	bytes.Buffer
}

func (b *bufNopCloser) Close() error {
	return nil
}

func TestStubAppender(t *testing.T) {
	var buf bufNopCloser

	appender := NewAppender()
	appender.Tag = "gol"
	appender.Facility = LOG_LOCAL7
	appender.hostname = "localhost"
	appender.conn = &buf

	event := &gol.LoggingEvent{
		Level: gol.Debug,
		Name:  "gol/syslog",
		Time:  time.Date(2015, time.April, 3, 2, 1, 0, 789000000, time.Local),
	}
	event.Message.WriteString("message")

	appender.Append(event)
	msg := buf.String()
	if !strings.HasPrefix(msg, "<191>") {
		t.Fatalf("invalid priority %s", msg)
	}
	if !strings.HasSuffix(msg, "gol/syslog: message\n") {
		t.Fatalf("invalid message %s", msg)
	}
}

func TestAppender(t *testing.T) {
	appender := NewAppender()
	err := appender.Start()
	if err != nil {
		if strings.Contains(err.Error(), "syslog delivery error") {
			// Syslog is not supported.
			t.Skip(err)
		}
		t.Fatal(err)
	}
	defer appender.Stop()

	event := &gol.LoggingEvent{
		Level: gol.Info,
		Name:  "gol/syslog",
	}
	event.Message.WriteString("message")
	appender.Append(event)
}
