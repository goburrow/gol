package gol

import (
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

type stubFactory struct {
}

func (f *stubFactory) GetLogger(name string) Logger {
	return New("test."+name, nil)
}

func TestGetLogger(t *testing.T) {
	logger := GetLogger("abc")
	if nil == logger {
		t.Fatal("Default logger factory does not exist")
	}
}

func TestDebugMode(t *testing.T) {
	f, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		f.Close()
		os.Remove(f.Name())
	}()
	oldStderr := os.Stderr
	defer func() {
		os.Stderr = oldStderr
	}()

	os.Stderr = f
	SetDebugMode(true)
	Print(1, "2", errors.New("test"))
	if err = f.Sync(); err != nil {
		t.Fatal(err)
	}

	content, err := ioutil.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "1 2 test") {
		t.Fatal(string(content))
	}
}
