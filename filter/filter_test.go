package filter

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/goburrow/gol"
)

func TestAppenderThreshold(t *testing.T) {
	var buf bytes.Buffer

	appender := NewAppender(gol.NewAppender(&buf))
	event := &gol.LoggingEvent{
		FormattedMessage: "append",
		Name:             "filter",
		Level:            gol.LevelInfo,
		Time:             time.Now(),
	}
	appender.SetThreshold(gol.LevelInfo)
	appender.Append(event)
	msg := buf.String()
	if !strings.Contains(msg, "filter: append") {
		t.Fatalf("unexpected message: %#v", msg)
	}
	buf.Reset()

	appender.SetThreshold(gol.LevelWarn)
	appender.Append(event)
	msg = buf.String()
	if "" != msg {
		t.Fatalf("unexpected message: %#v", msg)
	}
}

func TestAppenderIncludes(t *testing.T) {
	var buf bytes.Buffer

	appender := NewAppender(gol.NewAppender(&buf))
	event := &gol.LoggingEvent{
		FormattedMessage: "append",
		Name:             "filter",
		Level:            gol.LevelInfo,
		Time:             time.Now(),
	}
	appender.SetIncludes([]string{"filter"})
	appender.Append(event)
	msg := buf.String()
	if !strings.Contains(msg, "filter: append") {
		t.Fatalf("unexpected message: %#v", msg)
	}
	buf.Reset()

	appender.SetIncludes([]string{"filter3", "filter1", "filter2"})
	appender.Append(event)
	msg = buf.String()
	if "" != msg {
		t.Fatalf("unexpected message: %#v", msg)
	}
}

func TestAppenderExcludes(t *testing.T) {
	var buf bytes.Buffer

	appender := NewAppender(gol.NewAppender(&buf))
	event := &gol.LoggingEvent{
		FormattedMessage: "append",
		Name:             "filter",
		Level:            gol.LevelInfo,
		Time:             time.Now(),
	}
	appender.SetExcludes([]string{"filter"})
	appender.Append(event)
	msg := buf.String()
	if "" != msg {
		t.Fatalf("unexpected message: %#v", msg)
	}
	// Excludes overrule includes
	appender.SetIncludes([]string{"filter"})
	appender.Append(event)
	msg = buf.String()
	if "" != msg {
		t.Fatalf("unexpected message: %#v", msg)
	}
	appender.SetExcludes([]string{"filter3", "filter1", "filter2"})
	appender.SetIncludes(nil)
	appender.Append(event)
	msg = buf.String()
	if !strings.Contains(msg, "filter: append") {
		t.Fatalf("unexpected message: %#v", msg)
	}
}
