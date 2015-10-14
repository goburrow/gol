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
		Level:            gol.Info,
		Name:             "gol/file",
		Time:             time.Now(),
	}
	appender.Append(event)

	data, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.HasSuffix(content, "gol/file: message\n") {
		t.Fatalf("unexpected content: %s", content)
	}
}

func TestExistedFile(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	_, err = f.Write([]byte("test\n"))
	f.Close()
	if err != nil {
		t.Fatal(err)
	}

	appender := NewAppender(f.Name())
	event := &gol.LoggingEvent{
		FormattedMessage: "message",
		Level:            gol.Info,
		Name:             "gol/file",
		Time:             time.Now(),
	}
	appender.Append(event)
	appender.Stop()

	data, err := ioutil.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.HasPrefix(content, "test\n") || !strings.HasSuffix(content, "gol/file: message\n") {
		t.Fatalf("unexpected content: %s", content)
	}
}
