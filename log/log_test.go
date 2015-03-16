package log

import (
	"bytes"
	"log"
	"strings"
	"testing"

	"github.com/goburrow/gol"
)

func TestLogPrint(t *testing.T) {
	var buf bytes.Buffer
	gol.GetLogger(gol.RootLoggerName).(*gol.DefaultLogger).SetAppender(gol.NewAppender(&buf))

	log.Print("my message")
	if !strings.HasSuffix(buf.String(), "log_test.go: my message\n") {
		t.Fatalf("unexpected message % x", buf.Bytes())
	}
}

func TestLogPrintf(t *testing.T) {
	var buf bytes.Buffer
	gol.GetLogger(gol.RootLoggerName).(*gol.DefaultLogger).SetAppender(gol.NewAppender(&buf))

	log.Printf("%v: %v", "x", 1)
	if !strings.HasSuffix(buf.String(), "log_test.go: x: 1\n") {
		t.Fatalf("unexpected message % x", buf.Bytes())
	}
}
