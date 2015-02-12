package gol

import (
	"testing"
)

type stubFactory struct {
}

func (f *stubFactory) GetLogger(name string) Logger {
	return NewLogger("test." + name)
}

func TestGetLogger(t *testing.T) {
	logger := GetLogger("abc")
	if nil == logger {
		t.Fatal("Default logger factory does not exist")
	}
}

func TestSetLoggerFactory(t *testing.T) {
	SetLoggerFactory(&stubFactory{})
	logger := GetLogger("go")
	if "test.go" != logger.(*DefaultLogger).name {
		t.Fatalf("Unexpected logger %#v", logger)
	}
}
