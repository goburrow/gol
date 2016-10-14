package gol

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"
)

const packageSeparator = '/'

// Level represents logging level.
type Level int

// Log levels
const (
	Uninitialized Level = iota
	All
	Trace
	Debug
	Info
	Warn
	Error
	Off
)

const (
	// RootLoggerName is the name of the root logger.
	RootLoggerName = "root"
)

var levelStrings = map[Level]string{
	All:   "ALL",
	Trace: "TRACE",
	Debug: "DEBUG",
	Info:  "INFO",
	Warn:  "WARN",
	Error: "ERROR",
	Off:   "OFF",
}

// LevelString returns the text for the level.
func LevelString(level Level) string {
	return levelStrings[level]
}

// LoggingEvent is the representation of logging events.
type LoggingEvent struct {
	// Name is the of the logger.
	Name string
	// Level is the current logger level.
	Level Level
	// Time is when the logging happens.
	Time time.Time

	Message bytes.Buffer
}

var eventPool = sync.Pool{}

func newLoggingEvent() *LoggingEvent {
	e := eventPool.Get()
	if e != nil {
		return e.(*LoggingEvent)
	}
	return &LoggingEvent{}
}

func releaseLoggingEvent(e *LoggingEvent) {
	e.Message.Reset()
	eventPool.Put(e)
}

// Appender appends contents to a Writer.
type Appender interface {
	Append(*LoggingEvent)
}

// DefaultAppender implements Appender interface.
type DefaultAppender struct {
	timeLayout string

	target io.Writer
}

// NewAppender allocates and returns a new DefaultAppender.
func NewAppender(target io.Writer) *DefaultAppender {
	return &DefaultAppender{
		timeLayout: "2006-01-02T15:04:05.000Z07:00", // ISO8601 with milliseconds.
		target:     target,
	}
}

// Append uses its encoder to send the event to the target.
func (appender *DefaultAppender) Append(event *LoggingEvent) {
	if appender.target == nil {
		return
	}
	// Use buffer from a logging event
	e := newLoggingEvent()
	defer releaseLoggingEvent(e)
	buf := &e.Message

	// Level (minimum 5 characters)
	level := LevelString(event.Level)
	n, _ := buf.WriteString(level)
	for n = 5 - n; n > 0; n-- {
		buf.WriteByte(' ')
	}

	// Time
	buf.WriteByte(' ')
	buf.WriteByte('[')
	var timeBuf [64]byte
	buf.Write(event.Time.AppendFormat(timeBuf[:0], appender.timeLayout))
	buf.WriteByte(']')

	// Logger name
	buf.WriteByte(' ')
	buf.WriteString(event.Name)
	buf.WriteByte(':')

	// Logging message in the end
	buf.WriteByte(' ')
	buf.Write(event.Message.Bytes())
	buf.WriteByte('\n')

	_, err := buf.WriteTo(appender.target)
	if err != nil {
		Print(err)
	}
}

// DefaultLogger implements Logger interface.
type DefaultLogger struct {
	name  string
	level Level

	appender Appender

	parent *DefaultLogger
}

// New allocates and returns a new DefaultLogger.
// This method should not be called directly in application, use
// LoggerFactory.GetLogger() instead as a DefaultLogger requires
// Appender from itself or its parent.
func New(name string, parent *DefaultLogger) *DefaultLogger {
	return &DefaultLogger{
		name:  name,
		level: Uninitialized,

		parent: parent,
	}
}

// Tracef logs message at Trace level.
func (logger *DefaultLogger) Tracef(format string, args ...interface{}) {
	logger.Printf(Trace, format, args)
}

// TraceEnabled checks if Trace level is enabled.
func (logger *DefaultLogger) TraceEnabled() bool {
	return logger.loggable(Trace)
}

// Debugf logs message at Debug level.
func (logger *DefaultLogger) Debugf(format string, args ...interface{}) {
	logger.Printf(Debug, format, args)
}

// DebugEnabled checks if Debug level is enabled.
func (logger *DefaultLogger) DebugEnabled() bool {
	return logger.loggable(Debug)
}

// Infof logs message at Info level.
func (logger *DefaultLogger) Infof(format string, args ...interface{}) {
	logger.Printf(Info, format, args)
}

// InfoEnabled checks if Info level is enabled.
func (logger *DefaultLogger) InfoEnabled() bool {
	return logger.loggable(Info)
}

// Warnf logs message at Warning level.
func (logger *DefaultLogger) Warnf(format string, args ...interface{}) {
	logger.Printf(Warn, format, args)
}

// WarnEnabled checks if Warning level is enabled.
func (logger *DefaultLogger) WarnEnabled() bool {
	return logger.loggable(Warn)
}

// Errorf logs message at Error level.
func (logger *DefaultLogger) Errorf(format string, args ...interface{}) {
	logger.Printf(Error, format, args)
}

// ErrorEnabled checks if Error level is enabled.
func (logger *DefaultLogger) ErrorEnabled() bool {
	return logger.loggable(Error)
}

// Level returns level of this logger or parent if not set.
func (logger *DefaultLogger) Level() Level {
	for logger != nil {
		if logger.level != Uninitialized {
			return logger.level
		}
		logger = logger.parent
	}
	return Off
}

// SetLevel changes logging level of this logger.
func (logger *DefaultLogger) SetLevel(level Level) {
	logger.level = level
}

// Appender returns appender of this logger or parent if not set.
func (logger *DefaultLogger) Appender() Appender {
	for logger != nil {
		if logger.appender != nil {
			return logger.appender
		}
		logger = logger.parent
	}
	return nil
}

// SetAppender changes appender of this logger.
func (logger *DefaultLogger) SetAppender(appender Appender) {
	logger.appender = appender
}

// loggable checks if the given logging level is enabled within this logger.
func (logger *DefaultLogger) loggable(level Level) bool {
	return level >= logger.Level()
}

// log performs logging with given parameters.
func (logger *DefaultLogger) Printf(level Level, format string, args []interface{}) {
	if !logger.loggable(level) {
		return
	}
	appender := logger.Appender()
	if appender == nil {
		return
	}
	event := newLoggingEvent()
	defer releaseLoggingEvent(event)

	event.Time = time.Now()
	event.Name = logger.name
	event.Level = level
	fmt.Fprintf(&event.Message, format, args...)

	appender.Append(event)
}

// DefaultFactory implements Factory interface.
type DefaultFactory struct {
	root *DefaultLogger

	mu      sync.RWMutex
	loggers map[string]*DefaultLogger
}

// NewFactory allocates and returns new DefaultFactory.
func NewFactory(writer io.Writer) *DefaultFactory {
	rootLogger := New(RootLoggerName, nil)
	rootLogger.SetLevel(Info)
	rootLogger.SetAppender(NewAppender(writer))

	return &DefaultFactory{
		root: rootLogger,
		loggers: map[string]*DefaultLogger{
			RootLoggerName: rootLogger,
		},
	}
}

// GetLogger returns a new Logger or an existing one if the same name is found.
func (factory *DefaultFactory) GetLogger(name string) Logger {
	if name == "" {
		return factory.root
	}
	factory.mu.RLock()
	logger, ok := factory.loggers[name]
	factory.mu.RUnlock()
	if !ok {
		factory.mu.Lock()
		logger = factory.createLogger(name, factory.getParent(name))
		factory.mu.Unlock()
	}
	return logger
}

// getParent returns parent logger for given logger.
func (factory *DefaultFactory) getParent(name string) *DefaultLogger {
	parent := factory.root
	for i, c := range name {
		// Search for package separator character
		if c == packageSeparator {
			parentName := name[0:i]
			if parentName != "" {
				parent = factory.createLogger(parentName, parent)
			}
		}
	}
	return parent
}

// createLogger creates a new logger if not exist.
func (factory *DefaultFactory) createLogger(name string, parent *DefaultLogger) *DefaultLogger {
	logger, ok := factory.loggers[name]
	if !ok {
		logger = New(name, parent)
		factory.loggers[name] = logger
	}
	return logger
}
