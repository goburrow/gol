package gol

import (
	"bytes"
	"errors"
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

func TestAppender(t *testing.T) {
	var buf bytes.Buffer
	appender := NewAppender(&buf)

	event := &LoggingEvent{
		Name:  "My Logger",
		Level: Trace,
		Time:  time.Date(2015, time.April, 3, 2, 1, 0, 789000000, time.UTC),
	}
	event.Message.WriteString("something")
	appender.Append(event)

	event.Message.Reset()
	event.Message.WriteString("something else")
	event.Level = Warn
	event.Time = event.Time.Add(1 * time.Minute)
	appender.Append(event)

	assertEquals(t, "TRACE [2015-04-03T02:01:00.789Z] My Logger: something\n"+
		"WARN  [2015-04-03T02:02:00.789Z] My Logger: something else\n", buf.String())
}

func TestAppenderWithTimeLayout(t *testing.T) {
	var buf bytes.Buffer
	appender := NewAppender(&buf)
	appender.timeLayout = time.RFC3339

	event := &LoggingEvent{
		Name:  "name",
		Level: Error,
		Time:  time.Date(2015, time.March, 21, 0, 0, 0, 789000000, time.FixedZone("Asia/Ho_Chi_Minh", 7*60*60)),
	}
	event.Message.WriteString("message")
	appender.Append(event)

	assertEquals(t, "ERROR [2015-03-21T00:00:00+07:00] name: message\n", buf.String())
}

type errorWriter struct {
}

func (*errorWriter) Write([]byte) (int, error) {
	return 0, errors.New("write")
}

func TestAppenderWithErrorEncoder(t *testing.T) {
	appender := NewAppender(&errorWriter{})

	event := &LoggingEvent{}
	appender.Append(event)
}

func TestLogger(t *testing.T) {
	var buf bytes.Buffer

	logger := New("MyLogger", nil)
	logger.SetLevel(Info)
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
	logger := New("RootLogger", nil)

	// Should do nothing
	logAllLevels(logger)

	logger.SetLevel(Info)
	logAllLevels(logger)

	logger.SetAppender(NewAppender(nil))
	logAllLevels(logger)
}

func TestLoggerLevel(t *testing.T) {
	var buf bytes.Buffer

	root := New(RootLoggerName, nil)
	root.SetLevel(All)
	root.SetAppender(NewAppender(&buf))

	logger := New("MyLogger", root)
	assertEquals(t, true, logger.TraceEnabled())
	assertEquals(t, true, logger.DebugEnabled())
	assertEquals(t, true, logger.InfoEnabled())
	assertEquals(t, true, logger.WarnEnabled())
	assertEquals(t, true, logger.ErrorEnabled())
	logAllLevels(logger)
	assertContains(t, buf.String(), "TRACE", "Trace", "DEBUG", "Debug", "INFO", "Info", "WARN", "Warn", "ERROR", "Error")

	buf.Reset()
	logger.SetLevel(Trace)
	assertEquals(t, true, logger.TraceEnabled())
	assertEquals(t, true, logger.DebugEnabled())
	assertEquals(t, true, logger.InfoEnabled())
	assertEquals(t, true, logger.WarnEnabled())
	assertEquals(t, true, logger.ErrorEnabled())
	logAllLevels(logger)
	assertContains(t, buf.String(), "TRACE", "Trace", "DEBUG", "Debug", "INFO", "Info", "WARN", "Warn", "ERROR", "Error")

	buf.Reset()
	logger.SetLevel(Debug)
	assertEquals(t, false, logger.TraceEnabled())
	assertEquals(t, true, logger.DebugEnabled())
	assertEquals(t, true, logger.InfoEnabled())
	assertEquals(t, true, logger.WarnEnabled())
	assertEquals(t, true, logger.ErrorEnabled())
	logAllLevels(logger)
	assertContains(t, buf.String(), "DEBUG", "Debug", "INFO", "Info", "WARN", "Warn", "ERROR", "Error")

	buf.Reset()
	logger.SetLevel(Info)
	assertEquals(t, false, logger.TraceEnabled())
	assertEquals(t, false, logger.DebugEnabled())
	assertEquals(t, true, logger.InfoEnabled())
	assertEquals(t, true, logger.WarnEnabled())
	assertEquals(t, true, logger.ErrorEnabled())
	logAllLevels(logger)
	assertContains(t, buf.String(), "INFO", "Info", "WARN", "Warn", "ERROR", "Error")

	buf.Reset()
	logger.SetLevel(Warn)
	assertEquals(t, false, logger.TraceEnabled())
	assertEquals(t, false, logger.DebugEnabled())
	assertEquals(t, false, logger.InfoEnabled())
	assertEquals(t, true, logger.WarnEnabled())
	assertEquals(t, true, logger.ErrorEnabled())
	logAllLevels(logger)
	assertContains(t, buf.String(), "WARN", "Warn", "ERROR", "Error")

	buf.Reset()
	logger.SetLevel(Error)
	assertEquals(t, false, logger.TraceEnabled())
	assertEquals(t, false, logger.DebugEnabled())
	assertEquals(t, false, logger.InfoEnabled())
	assertEquals(t, false, logger.WarnEnabled())
	assertEquals(t, true, logger.ErrorEnabled())
	logAllLevels(logger)
	assertContains(t, buf.String(), "ERROR", "Error")

	buf.Reset()
	logger.SetLevel(Off)
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
	factory := NewFactory(os.Stdout)
	logger := factory.GetLogger("abc").(*DefaultLogger)

	assertEquals(t, "abc", logger.name)

	logger = factory.GetLogger("def").(*DefaultLogger)
	assertEquals(t, "def", logger.name)

	logger = factory.GetLogger("").(*DefaultLogger)
	assertEquals(t, "root", logger.name)
}

func TestLoggerHierarchy(t *testing.T) {
	factory := NewFactory(os.Stdout)
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
	factory := NewFactory(os.Stdout)
	root := factory.GetLogger(RootLoggerName).(*DefaultLogger)
	a := factory.GetLogger("root/a").(*DefaultLogger)
	assertEquals(t, 2, len(factory.loggers))
	assertEquals(t, root, a.parent)
}
