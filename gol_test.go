package gol

import (
	"testing"
)

func TestGetLogger(t *testing.T) {
	logger := GetLogger("abc")

	if nil == logger {
		t.Fatalf("Unexpected logger: %v", logger)
	}
}
