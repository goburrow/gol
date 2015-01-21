package gol

import (
	"bytes"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"
)

func getFileLine() (file string, line int) {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	} else {
		// Get file name only
		idx := strings.LastIndex(file, "/")
		if idx >= 0 {
			file = file[idx+1:]
		}
	}
	return
}

func assertEquals(t *testing.T, expected, actual interface{}) {
	file, line := getFileLine()

	if expected != actual {
		t.Logf("%s:%d: Expected: %+v (%T), actual: %+v (%T)\n", file, line,
			expected, expected, actual, actual)
		t.Fail()
	}
}

func assertContains(t *testing.T, s string, subs ...string) {
	file, line := getFileLine()

	for _, sub := range subs {
		if !strings.Contains(s, sub) {
			t.Logf("%s:%d: Could not find %s in: %s", file, line, sub, s)
			t.Fail()
		}
	}
}

func TestLayouter(t *testing.T) {
	layouter := NewLayouter()
	args := make([]interface{}, 2)
	args[0] = "#1"
	args[1] = "#2"

	event := LoggingEvent{
		Format:    "Arguments: %v, %v",
		Arguments: args,
		Name:      "My Logger",
		Level:     LevelTrace,
		Time:      time.Date(2015, time.April, 3, 2, 1, 0, 789000000, time.UTC),
	}
	content := layouter.Layout(&event)
	expected := "TRACE [2015-04-03T02:01:00.789+00:00] My Logger: Arguments: #1, #2\n"
	if expected != content {
		t.Fatalf("Unexpected content: %s", content)
	}
}

func TestAppender(t *testing.T) {
	var buf bytes.Buffer
	appender := NewAppender(&buf)

	appender.Write([]byte("something"))

	var buf2 bytes.Buffer
	appender.SetWriter(&buf2)
	appender.Write([]byte("something else"))
	assertEquals(t, "something", buf.String())
	assertEquals(t, "something else", buf2.String())
}

func TestLogger(t *testing.T) {
	var buf bytes.Buffer

	logger := NewLogger("MyLogger", nil)
	logger.SetLevel(LevelInfo)
	logger.SetLayouter(NewLayouter())
	logger.SetAppender(NewAppender(&buf))

	logger.Trace("Trace")
	logger.Debug("Debug %v %v", "a", 1)
	logger.Info("Info %v %v", "b", 2)
	logger.Error("Error %v", "c")
	logger.Warn("Warn")

	content := buf.String()
	lines := strings.Split(content, "\n")

	if !strings.HasPrefix(lines[0], "INFO  [") || !strings.HasSuffix(lines[0], "] MyLogger: Info b 2") {
		t.Fatalf("Unexpected content: %s", lines[0])
	}
	if !strings.HasPrefix(lines[1], "ERROR [") || !strings.HasSuffix(lines[1], "] MyLogger: Error c") {
		t.Fatalf("Unexpected content: %s", lines[1])
	}
	if !strings.HasPrefix(lines[2], "WARN  [") || !strings.HasSuffix(lines[2], "] MyLogger: Warn") {
		t.Fatalf("Unexpected content: %s", lines[2])
	}
}

func logAllLevels(logger Logger) {
	logger.Trace("Trace")
	logger.Debug("Debug")
	logger.Info("Info")
	logger.Warn("Warn")
	logger.Error("Error")
}

func TestRootLogger(t *testing.T) {
	logger := NewLogger("RootLogger", nil)

	// Should do nothing
	logAllLevels(logger)

	logger.SetLevel(LevelInfo)
	logAllLevels(logger)

	logger.SetLayouter(NewLayouter())
	logAllLevels(logger)

	logger.SetLayouter(nil)
	logger.SetAppender(NewAppender(nil))
	logAllLevels(logger)
}

func TestLoggerLevel(t *testing.T) {
	var buf bytes.Buffer

	root := NewLogger("ROOT", nil)
	root.SetLevel(LevelAll)
	root.SetLayouter(NewLayouter())
	root.SetAppender(NewAppender(&buf))

	logger := NewLogger("MyLogger", root)
	assertEquals(t, true, logger.TraceEnabled())
	assertEquals(t, true, logger.DebugEnabled())
	assertEquals(t, true, logger.InfoEnabled())
	assertEquals(t, true, logger.WarnEnabled())
	assertEquals(t, true, logger.ErrorEnabled())
	logAllLevels(logger)
	assertContains(t, buf.String(), "TRACE", "Trace", "DEBUG", "Debug", "INFO", "Info", "WARN", "Warn", "ERROR", "Error")

	buf.Reset()
	logger.SetLevel(LevelTrace)
	assertEquals(t, true, logger.TraceEnabled())
	assertEquals(t, true, logger.DebugEnabled())
	assertEquals(t, true, logger.InfoEnabled())
	assertEquals(t, true, logger.WarnEnabled())
	assertEquals(t, true, logger.ErrorEnabled())
	logAllLevels(logger)
	assertContains(t, buf.String(), "TRACE", "Trace", "DEBUG", "Debug", "INFO", "Info", "WARN", "Warn", "ERROR", "Error")

	buf.Reset()
	logger.SetLevel(LevelDebug)
	assertEquals(t, false, logger.TraceEnabled())
	assertEquals(t, true, logger.DebugEnabled())
	assertEquals(t, true, logger.InfoEnabled())
	assertEquals(t, true, logger.WarnEnabled())
	assertEquals(t, true, logger.ErrorEnabled())
	logAllLevels(logger)
	assertContains(t, buf.String(), "DEBUG", "Debug", "INFO", "Info", "WARN", "Warn", "ERROR", "Error")

	buf.Reset()
	logger.SetLevel(LevelInfo)
	assertEquals(t, false, logger.TraceEnabled())
	assertEquals(t, false, logger.DebugEnabled())
	assertEquals(t, true, logger.InfoEnabled())
	assertEquals(t, true, logger.WarnEnabled())
	assertEquals(t, true, logger.ErrorEnabled())
	logAllLevels(logger)
	assertContains(t, buf.String(), "INFO", "Info", "WARN", "Warn", "ERROR", "Error")

	buf.Reset()
	logger.SetLevel(LevelWarn)
	assertEquals(t, false, logger.TraceEnabled())
	assertEquals(t, false, logger.DebugEnabled())
	assertEquals(t, false, logger.InfoEnabled())
	assertEquals(t, true, logger.WarnEnabled())
	assertEquals(t, true, logger.ErrorEnabled())
	logAllLevels(logger)
	assertContains(t, buf.String(), "WARN", "Warn", "ERROR", "Error")

	buf.Reset()
	logger.SetLevel(LevelError)
	assertEquals(t, false, logger.TraceEnabled())
	assertEquals(t, false, logger.DebugEnabled())
	assertEquals(t, false, logger.InfoEnabled())
	assertEquals(t, false, logger.WarnEnabled())
	assertEquals(t, true, logger.ErrorEnabled())
	logAllLevels(logger)
	assertContains(t, buf.String(), "ERROR", "Error")

	buf.Reset()
	logger.SetLevel(LevelOff)
	assertEquals(t, false, logger.TraceEnabled())
	assertEquals(t, false, logger.DebugEnabled())
	assertEquals(t, false, logger.InfoEnabled())
	assertEquals(t, false, logger.WarnEnabled())
	assertEquals(t, false, logger.ErrorEnabled())
	if "" != buf.String() {
		t.Fatalf("Content should be empty: %s", buf.String())
	}
}

func TestLoggerFactory(t *testing.T) {
	var factory LoggerFactory = NewLoggerFactory(os.Stdout)
	logger := factory.GetLogger("abc")

	assertEquals(t, "abc", logger.(*DefaultLogger).name)

	logger = factory.GetLogger("def")
	assertEquals(t, "def", logger.(*DefaultLogger).name)

	logger = factory.GetLogger("")
	assertEquals(t, "ROOT", logger.(*DefaultLogger).name)
}
