package async

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

func TestAppender(t *testing.T) {
	c := make(chan string)

	appender := NewAppender(gol.NewAppender(channelWriter(c)))
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

func TestAppenderStop(t *testing.T) {
	buffers := [...]bytes.Buffer{
		bytes.Buffer{},
		bytes.Buffer{},
		bytes.Buffer{},
	}

	appender := NewAppender(
		gol.NewAppender(&buffers[0]),
		gol.NewAppender(&buffers[1]),
		gol.NewAppender(&buffers[2]),
	)
	appender.Start()
	event := &gol.LoggingEvent{
		FormattedMessage: "run",
		Name:             "async",
		Level:            gol.LevelInfo,
		Time:             time.Now(),
	}
	appender.Append(event)
	appender.Stop()
	for i, _ := range buffers {
		if !strings.Contains(buffers[i].String(), "async: run") {
			t.Fatalf("unexpected message: %#v", buffers[i].String())
		}
	}
}
