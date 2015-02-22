package misc

import (
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
