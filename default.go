package gol

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// Level represents logging level
type Level int

const (
	LevelUninitialized Level = iota
	LevelAll
	LevelTrace
	LevelDebug
	LevelInfo
	LevelWarn
	LevelError
	LevelOff
)

var levelStrings = []string{
	LevelAll:   "ALL",
	LevelTrace: "TRACE",
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
	LevelOff:   "OFF",
}

const (
	RootLoggerName = "ROOT"
	// ISO8601 with milliseconds
	defaultTimeFormat = "2006-01-02T15:04:05.000-07:00"
	// See LoggingEvent for the order
	defaultFormat = "%-5[3]s [%[4]s] %[2]s: %[1]s\n"
)

// LoggingEvent is the representation of logging events
type LoggingEvent struct {
	// #1: The 2 following properties construct formatted message
	Format    string
	Arguments []interface{}
	// #2: Name of the logger
	Name string
	// #3: Log level
	Level Level
	// #4: Time that the logging happens
	Time time.Time
}

// Layouter constructs final message with given event
type Layouter interface {
	Layout(*LoggingEvent) string
}

// Appender appends contents to a Writer
type Appender interface {
	io.Writer
}

// DefaultLayouter implements Layouter interface
type DefaultLayouter struct {
	Format     string
	TimeFormat string
}

// NewLayouter allocates and returns a new DefaultLayouter
func NewLayouter() *DefaultLayouter {
	return &DefaultLayouter{
		Format:     defaultFormat,
		TimeFormat: defaultTimeFormat,
	}
}

func (layouter *DefaultLayouter) Layout(event *LoggingEvent) string {
	return fmt.Sprintf(layouter.Format,
		fmt.Sprintf(event.Format, event.Arguments...),
		event.Name,
		levelStrings[event.Level],
		event.Time.Format(layouter.TimeFormat))
}

// DefaultAppender implements Appender interface
type DefaultAppender struct {
	mu     sync.Mutex
	writer io.Writer
}

// NewAppender allocates and returns a new DefaultAppender
func NewAppender(writer io.Writer) *DefaultAppender {
	return &DefaultAppender{
		writer: writer,
	}
}

func (appender *DefaultAppender) Write(p []byte) (int, error) {
	appender.mu.Lock()
	defer appender.mu.Unlock()
	return appender.writer.Write(p)
}

func (appender *DefaultAppender) SetWriter(writer io.Writer) {
	appender.mu.Lock()
	appender.writer = writer
	appender.mu.Unlock()
}

// DefaultLogger implements Logger interface
type DefaultLogger struct {
	name   string
	parent *DefaultLogger

	level    Level
	layouter Layouter
	appender Appender
}

// NewLogger allocates and returns a new DefaultLogger
func NewLogger(name string, parent *DefaultLogger) *DefaultLogger {
	return &DefaultLogger{
		name:   name,
		parent: parent,
		level:  LevelUninitialized,
	}
}

func (logger *DefaultLogger) Trace(format string, args ...interface{}) {
	logger.log(LevelTrace, format, args)
}

// TraceEnabled checks if Trace level is enabled
func (logger *DefaultLogger) TraceEnabled() bool {
	return logger.logEnabled(LevelTrace)
}

func (logger *DefaultLogger) Debug(format string, args ...interface{}) {
	logger.log(LevelDebug, format, args)
}

// DebugEnabled checks if Debug level is enabled
func (logger *DefaultLogger) DebugEnabled() bool {
	return logger.logEnabled(LevelDebug)
}

func (logger *DefaultLogger) Info(format string, args ...interface{}) {
	logger.log(LevelInfo, format, args)
}

// InfoEnabled checks if Info level is enabled
func (logger *DefaultLogger) InfoEnabled() bool {
	return logger.logEnabled(LevelInfo)
}

func (logger *DefaultLogger) Warn(format string, args ...interface{}) {
	logger.log(LevelWarn, format, args)
}

// WarnEnabled checks if Warning level is enabled
func (logger *DefaultLogger) WarnEnabled() bool {
	return logger.logEnabled(LevelWarn)
}

func (logger *DefaultLogger) Error(format string, args ...interface{}) {
	logger.log(LevelError, format, args)
}

// ErrorEnabled checks if Error level is enabled
func (logger *DefaultLogger) ErrorEnabled() bool {
	return logger.logEnabled(LevelError)
}

// Level returns level of this logger or parent if not set
func (logger *DefaultLogger) Level() Level {
	if logger.level != LevelUninitialized {
		return logger.level
	}
	if logger.parent != nil {
		return logger.parent.Level()
	}
	return LevelOff
}

// SetLevel changes logging level of this logger
func (logger *DefaultLogger) SetLevel(level Level) {
	logger.level = level
}

// Layouter returns layouter of this logger or parent if not set
func (logger *DefaultLogger) Layouter() Layouter {
	if logger.layouter != nil {
		return logger.layouter
	}
	if logger.parent != nil {
		return logger.parent.Layouter()
	}
	return logger.layouter
}

// SetLayouter changes layouter of this logger
func (logger *DefaultLogger) SetLayouter(layouter Layouter) {
	logger.layouter = layouter
}

// Appender returns appender of this logger or parent if not set
func (logger *DefaultLogger) Appender() Appender {
	if logger.appender != nil {
		return logger.appender
	}
	if logger.parent != nil {
		return logger.parent.Appender()
	}
	return logger.appender
}

// SetAppender changes appender of this logger
func (logger *DefaultLogger) SetAppender(appender Appender) {
	logger.appender = appender
}

// logEnabled checks if the given logging level is enabled within this logger
func (logger *DefaultLogger) logEnabled(level Level) bool {
	return level >= logger.Level()
}

// log performs logging with given parameters
func (logger *DefaultLogger) log(level Level, format string, args []interface{}) {
	if !logger.logEnabled(level) {
		return
	}
	layouter := logger.Layouter()
	appender := logger.Appender()
	if layouter == nil || appender == nil {
		return
	}
	event := LoggingEvent{
		Time:      time.Now(),
		Name:      logger.name,
		Level:     level,
		Format:    format,
		Arguments: args,
	}
	record := layouter.Layout(&event)
	// Asynchronous
	appender.Write([]byte(record))
}

// DefaultLoggerFactory implements LoggerFactory interface
type DefaultLoggerFactory struct {
	root *DefaultLogger

	mu      sync.Mutex
	loggers map[string]Logger
}

// NewLoggerFactory allocates and returns new DefaultLoggerFactory
func NewLoggerFactory(writer io.Writer) *DefaultLoggerFactory {
	factory := &DefaultLoggerFactory{
		root:    NewLogger(RootLoggerName, nil),
		loggers: make(map[string]Logger),
	}
	factory.root.SetLevel(LevelDebug)
	factory.root.SetLayouter(NewLayouter())
	factory.root.SetAppender(NewAppender(writer))
	factory.loggers[RootLoggerName] = factory.root
	return factory
}

// GetLogger returns a new Logger or an existing one if the same name is found
func (factory *DefaultLoggerFactory) GetLogger(name string) Logger {
	factory.mu.Lock()
	defer factory.mu.Unlock()

	if name == "" {
		name = RootLoggerName
	}
	logger, ok := factory.loggers[name]
	if !ok {
		logger = NewLogger(name, factory.root)
		factory.loggers[name] = logger
	}
	return logger
}
