package async

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/goburrow/gol"
)

var _ (gol.Appender) = (*Appender)(nil)

// channelWriter is used for testing async appender
type channelWriter chan string

func (c channelWriter) Write(b []byte) (int, error) {
	c <- string(b)
	return len(b), nil
}

func TestAppenderAsync(t *testing.T) {
	c := make(chan string)

	appender := NewAppender(gol.NewAppender(channelWriter(c)))
	appender.Start()
	defer appender.Stop()
	event := &gol.LoggingEvent{
		FormattedMessage: "run",
		Name:             "async",
		Level:            gol.Info,
		Time:             time.Now(),
	}
	appender.Append(event)
	select {
	case msg := <-c:
		if !strings.Contains(msg, "async: run") {
			t.Fatalf("unexpected message: %#v", msg)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("not received after 1 second")
	}
}

// slowWriter is a io.Writer which appends data with delay.
type slowWriter struct {
	d time.Duration
	s []string
}

func (s *slowWriter) Write(b []byte) (int, error) {
	time.Sleep(s.d)
	s.s = append(s.s, string(b))
	return len(b), nil
}

func TestAppenderStop(t *testing.T) {
	writers := [...]*slowWriter{
		&slowWriter{20 * time.Millisecond, nil},
		&slowWriter{100 * time.Millisecond, nil},
		&slowWriter{50 * time.Millisecond, nil},
	}

	size := 5
	appender := NewAppenderWithBufSize(size,
		gol.NewAppender(writers[0]),
		gol.NewAppender(writers[1]),
		gol.NewAppender(writers[2]),
	)
	appender.Start()
	event := &gol.LoggingEvent{
		FormattedMessage: "run",
		Name:             "async",
		Level:            gol.Info,
		Time:             time.Now(),
	}
	for i := 0; i < size; i++ {
		appender.Append(event)
	}
	appender.Stop()
	for _, w := range writers {
		if size != len(w.s) {
			t.Fatalf("unexpected message count: %#v", len(w.s))
		}
		if !strings.Contains(w.s[0], "async: run") {
			t.Fatalf("unexpected message: %#v", w.s[0])
		}
	}
}

func TestAppenderLifeCycle(t *testing.T) {
	var buf bytes.Buffer
	size := 5
	appender := NewAppenderWithBufSize(size, gol.NewAppender(&buf))

	event := &gol.LoggingEvent{
		FormattedMessage: "run",
		Name:             "async",
		Level:            gol.Info,
		Time:             time.Now(),
	}

	appender.Start()
	for i := 0; i < size; i++ {
		appender.Append(event)
	}
	appender.Stop()
	if strings.Count(buf.String(), "async: run") != size {
		t.Fatalf("unexpected message: %v", buf.String())
	}
	appender.Start()
	for i := 0; i < size; i++ {
		appender.Append(event)
	}
	appender.Stop()
	if strings.Count(buf.String(), "async: run") != 2*size {
		t.Fatalf("unexpected message: %v", buf.String())
	}
}
