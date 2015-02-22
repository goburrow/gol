package misc

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/goburrow/gol"
)

// channelWriter is used for testing async appender
type channelWriter chan string

func (c channelWriter) Write(b []byte) (int, error) {
	c <- string(b)
	return len(b), nil
}

func TestAsyncAppender(t *testing.T) {
	c := make(chan string)

	appender := NewAsyncAppender(gol.NewAppender(channelWriter(c)))
	event := &gol.LoggingEvent{
		FormattedMessage: "run",
		Name:             "async",
		Level:            gol.LevelInfo,
		Time:             time.Now(),
	}
	appender.Append(event)
	select {
	case msg := <-c:
		if !strings.Contains(msg, "async: run") {
			t.Fatalf("unexpected message: %#v", msg)
		}
	}
}

func TestThresholdAppender(t *testing.T) {
	var buf bytes.Buffer

	appender := NewThresholdAppender(gol.NewAppender(&buf))
	appender.Threshold = gol.LevelInfo
	event := &gol.LoggingEvent{
		FormattedMessage: "append",
		Name:             "threshold",
		Level:            gol.LevelInfo,
		Time:             time.Now(),
	}
	appender.Append(event)
	msg := buf.String()
	if !strings.Contains(msg, "threshold: append") {
		t.Fatalf("unexpected message: %#v", msg)
	}
	buf.Reset()

	appender.Threshold = gol.LevelWarn
	appender.Append(event)
	msg = buf.String()
	if "" != msg {
		t.Fatalf("unexpected message: %#v", msg)
	}
}
