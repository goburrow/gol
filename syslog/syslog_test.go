package syslog

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/goburrow/gol"
)

func TestEncoder(t *testing.T) {
	encoder := NewEncoder()
	encoder.Hostname = "localhost"
	encoder.Tag = "gol"
	encoder.Facility = LOG_LOCAL7

	event := &gol.LoggingEvent{
		FormattedMessage: "message",
		Level:            gol.LevelDebug,
		Name:             "gol/syslog",
		Time:             time.Date(2015, time.April, 3, 2, 1, 0, 789000000, time.Local),
	}

	var buf bytes.Buffer
	if err := encoder.Encode(event, &buf); err != nil {
		t.Fatal(err)
	}
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
		FormattedMessage: "message",
		Level:            gol.LevelInfo,
		Name:             "gol/syslog",
	}
	appender.Append(event)
}
