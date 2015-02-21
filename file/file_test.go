package file

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/goburrow/gol"
)

func TestFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(dir)

	file := filepath.Join(dir, "test.log")

	appender := NewAppender(file)
	err = appender.Start()
	if err != nil {
		t.Fatal(err)
	}
	defer appender.Stop()

	event := &gol.LoggingEvent{
		FormattedMessage: "message",
		Level:            gol.LevelInfo,
		Name:             "gol/file",
		Time:             time.Now(),
	}
	appender.Append(event)

	content, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "gol/file: message") {
		t.Fatalf("unexpected content: %s", string(content))
	}
}
