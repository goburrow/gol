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

	appender := NewAppender(1, gol.NewAppender(channelWriter(c)))
	err := appender.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer appender.Stop()
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

	appender := NewAppender(1,
		gol.NewAppender(&buffers[0]),
		gol.NewAppender(&buffers[1]),
		gol.NewAppender(&buffers[2]),
	)
	err := appender.Start()
	if err != nil {
		t.Fatal(err)
	}
	event := &gol.LoggingEvent{
		FormattedMessage: "run",
		Name:             "async",
		Level:            gol.LevelInfo,
		Time:             time.Now(),
	}
	appender.Append(event)
	err = appender.Stop()
	if err != nil {
		t.Fatal(err)
	}
	for i, _ := range buffers {
		if !strings.Contains(buffers[i].String(), "async: run") {
			t.Fatalf("unexpected message: %#v", buffers[i].String())
		}
	}
}

func TestAppenderLifeCycle(t *testing.T) {
	var buf bytes.Buffer
	size := 5
	appender := NewAppender(size, gol.NewAppender(&buf))

	event := &gol.LoggingEvent{
		FormattedMessage: "run",
		Name:             "async",
		Level:            gol.LevelInfo,
		Time:             time.Now(),
	}

	for i := 0; i < size; i++ {
		appender.Append(event)
	}
	err := appender.Stop()
	if err != nil {
		t.Fatal(err)
	}
	err = appender.Start()
	if err != nil {
		t.Fatal(err)
	}
	err = appender.Start()
	if err != nil {
		t.Fatal(err)
	}
	err = appender.Stop()
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(buf.String(), "async: run") != size {
		t.Fatalf("unexpected message: %v", buf.String())
	}
}
