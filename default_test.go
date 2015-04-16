package gol

import (
	"bytes"
	"errors"
	"io"
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

func TestFormatter(t *testing.T) {
	formatter := NewFormatter()
	args := make([]interface{}, 2)
	args[0] = "#1"
	args[1] = "#2"

	event := LoggingEvent{
		Format:    "Arguments: %v, %v",
		Arguments: args,
	}
	event.FormattedMessage = formatter.Format(&event)
	assertEquals(t, "Arguments: #1, #2", event.FormattedMessage)

	// Empty argument
	event.Arguments = nil
	event.FormattedMessage = formatter.Format(&event)
	assertEquals(t, "Arguments: %v, %v", event.FormattedMessage)
}

func TestEncoder(t *testing.T) {
	encoder := NewEncoder()
	event := &LoggingEvent{
		FormattedMessage: "message",
		Name:             "name",
		Level:            LevelError,
		Time:             time.Date(2015, time.March, 21, 0, 0, 0, 789000000, time.FixedZone("Asia/Ho_Chi_Minh", 7*60*60)),
	}

	var buf bytes.Buffer
	encoder.Encode(event, &buf)
	assertEquals(t, "ERROR [2015-03-21T00:00:00.789+07:00] name: message\n", buf.String())

	buf.Reset()
	encoder.Layout = "%[4]s %[2]s: (%[3]s) %[1]s\n"
	encoder.TimeLayout = time.RFC3339
	encoder.Encode(event, &buf)
	assertEquals(t, "2015-03-21T00:00:00+07:00 name: (ERROR) message\n", buf.String())
}

func TestAppender(t *testing.T) {
	var buf bytes.Buffer
	appender := NewAppender(&buf)

	event := &LoggingEvent{
		FormattedMessage: "something",
		Name:             "My Logger",
		Level:            LevelTrace,
		Time:             time.Date(2015, time.April, 3, 2, 1, 0, 789000000, time.UTC),
	}
	appender.Append(event)

	event.FormattedMessage = "something else"
	event.Time = event.Time.Add(1 * time.Minute)

	var buf2 bytes.Buffer
	encoder := NewEncoder()
	encoder.Layout = "%[4]s %[2]s: %[1]s\n"

	appender.SetTarget(&buf2)
	appender.SetEncoder(encoder)

	appender.Append(event)
	assertEquals(t, "TRACE [2015-04-03T02:01:00.789Z] My Logger: something\n", buf.String())
	assertEquals(t, "2015-04-03T02:02:00.789Z My Logger: something else\n", buf2.String())
}

type errorEncoder struct {
}

func (*errorEncoder) Encode(*LoggingEvent, io.Writer) error {
	return errors.New("encode")
}

func TestAppenderWithErrorEncoder(t *testing.T) {
	var buf bytes.Buffer
	appender := NewAppender(&buf)
	appender.SetEncoder(&errorEncoder{})

	event := &LoggingEvent{
		FormattedMessage: "something",
	}
	appender.Append(event)
	assertEquals(t, "", buf.String())
}

func TestLogger(t *testing.T) {
	var buf bytes.Buffer

	logger := NewLogger("MyLogger")
	logger.SetLevel(LevelInfo)
	logger.SetFormatter(NewFormatter())
	logger.SetAppender(NewAppender(&buf))

	logger.Tracef("Trace")
	logger.Debugf("Debug %v %v", "a", 1)
	logger.Infof("Info %v %v", "b", 2)
	logger.Errorf("Error %v", "c")
	logger.Warnf("Warn")

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
	logger.Tracef("Trace")
	logger.Debugf("Debug")
	logger.Infof("Info")
	logger.Warnf("Warn")
	logger.Errorf("Error")
}

func TestRootLogger(t *testing.T) {
	logger := NewLogger("RootLogger")

	// Should do nothing
	logAllLevels(logger)

	logger.SetLevel(LevelInfo)
	logAllLevels(logger)

	logger.SetFormatter(NewFormatter())
	logAllLevels(logger)

	logger.SetFormatter(nil)
	logger.SetAppender(NewAppender(nil))
	logAllLevels(logger)
}

func TestLoggerLevel(t *testing.T) {
	var buf bytes.Buffer

	root := NewLogger(RootLoggerName)
	root.SetLevel(LevelAll)
	root.SetFormatter(NewFormatter())
	root.SetAppender(NewAppender(&buf))

	logger := NewLogger("MyLogger")
	logger.SetParent(root)
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
	factory := NewLoggerFactory(os.Stdout)
	logger := factory.GetLogger("abc").(*DefaultLogger)

	assertEquals(t, "abc", logger.name)

	logger = factory.GetLogger("def").(*DefaultLogger)
	assertEquals(t, "def", logger.name)

	logger = factory.GetLogger("").(*DefaultLogger)
	assertEquals(t, "root", logger.name)
}

func TestLoggerHierarchy(t *testing.T) {
	factory := NewLoggerFactory(os.Stdout)
	root := factory.GetLogger("").(*DefaultLogger)
	if root.parent != nil {
		t.Fatal("Parent of root logger must be nil")
	}

	a := factory.GetLogger("aaa").(*DefaultLogger)
	assertEquals(t, root, a.parent)

	c := factory.GetLogger("aaa/bb/c.go").(*DefaultLogger)
	b := factory.GetLogger("aaa/bb").(*DefaultLogger)
	assertEquals(t, a, b.parent)
	assertEquals(t, b, c.parent)
}

func TestHierarchyWithRootLogger(t *testing.T) {
	factory := NewLoggerFactory(os.Stdout)
	root := factory.GetLogger(RootLoggerName).(*DefaultLogger)
	a := factory.GetLogger("root/a").(*DefaultLogger)
	assertEquals(t, 2, len(factory.(*DefaultLoggerFactory).loggers))
	assertEquals(t, root, a.parent)
}
