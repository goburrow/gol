package gol

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// Level represents logging level.
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

const (
	// Name of the root logger.
	RootLoggerName = "root"
	// ISO8601 with milliseconds.
	defaultTimeFormat = "2006-01-02T15:04:05.000-07:00"
	// See LoggingEvent for the order.
	defaultLayout = "%-5[3]s [%[4]s] %[2]s: %[1]s\n"
)

var levelStrings = map[Level]string{
	LevelAll:   "ALL",
	LevelTrace: "TRACE",
	LevelDebug: "DEBUG",
	LevelInfo:  "INFO",
	LevelWarn:  "WARN",
	LevelError: "ERROR",
	LevelOff:   "OFF",
}

// LevelString returns the text for the level.
func LevelString(level Level) string {
	return levelStrings[level]
}

// LoggingEvent is the representation of logging events.
type LoggingEvent struct {
	// FormattedMessage is the message formatted by Formatter.
	FormattedMessage string
	// #1: The 2 following properties construct formatted message.
	Format    string
	Arguments []interface{}
	// #2: Name of the logger.
	Name string
	// #3: Log level
	Level Level
	// #4: Time that the logging happens.
	Time time.Time
}

// Formatter constructs final message with given event.
type Formatter interface {
	Format(*LoggingEvent)
}

// Appender appends contents to a Writer.
type Appender interface {
	Append(*LoggingEvent)
}

// DefaultFormatter implements Formatter interface.
type DefaultFormatter struct {
	Layout     string
	TimeFormat string
}

// NewFormatter allocates and returns a new DefaultFormatter.
func NewFormatter() Formatter {
	return &DefaultFormatter{
		Layout:     defaultLayout,
		TimeFormat: defaultTimeFormat,
	}
}

func (formatter *DefaultFormatter) Format(event *LoggingEvent) {
	var msg string

	if len(event.Arguments) > 0 {
		msg = fmt.Sprintf(event.Format, event.Arguments...)
	} else {
		msg = event.Format
	}

	event.FormattedMessage = fmt.Sprintf(formatter.Layout,
		msg,
		event.Name,
		LevelString(event.Level),
		event.Time.Format(formatter.TimeFormat))
}

// DefaultAppender implements Appender interface.
type DefaultAppender struct {
	mu     sync.Mutex
	target io.Writer
}

// NewAppender allocates and returns a new DefaultAppender.
func NewAppender(target io.Writer) Appender {
	return &DefaultAppender{
		target: target,
	}
}

func (appender *DefaultAppender) Append(event *LoggingEvent) {
	appender.mu.Lock()
	defer appender.mu.Unlock()
	appender.target.Write([]byte(event.FormattedMessage))
}

func (appender *DefaultAppender) SetTarget(target io.Writer) {
	appender.mu.Lock()
	appender.target = target
	appender.mu.Unlock()
}

// DefaultLogger implements Logger interface.
type DefaultLogger struct {
	name   string
	parent *DefaultLogger

	level     Level
	formatter Formatter
	appender  Appender
}

// NewLogger allocates and returns a new DefaultLogger.
// This method should not be called directly in application, use
// LoggerFactory.GetLogger() instead as a DefaultLogger requires Appender and
// Formatter either from itself or its parent.
func NewLogger(name string) Logger {
	return &DefaultLogger{
		name:  name,
		level: LevelUninitialized,
	}
}

func (logger *DefaultLogger) Trace(format string, args ...interface{}) {
	logger.log(LevelTrace, format, args)
}

// TraceEnabled checks if Trace level is enabled.
func (logger *DefaultLogger) TraceEnabled() bool {
	return logger.loggable(LevelTrace)
}

func (logger *DefaultLogger) Debug(format string, args ...interface{}) {
	logger.log(LevelDebug, format, args)
}

// DebugEnabled checks if Debug level is enabled.
func (logger *DefaultLogger) DebugEnabled() bool {
	return logger.loggable(LevelDebug)
}

func (logger *DefaultLogger) Info(format string, args ...interface{}) {
	logger.log(LevelInfo, format, args)
}

// InfoEnabled checks if Info level is enabled.
func (logger *DefaultLogger) InfoEnabled() bool {
	return logger.loggable(LevelInfo)
}

func (logger *DefaultLogger) Warn(format string, args ...interface{}) {
	logger.log(LevelWarn, format, args)
}

// WarnEnabled checks if Warning level is enabled.
func (logger *DefaultLogger) WarnEnabled() bool {
	return logger.loggable(LevelWarn)
}

func (logger *DefaultLogger) Error(format string, args ...interface{}) {
	logger.log(LevelError, format, args)
}

// ErrorEnabled checks if Error level is enabled.
func (logger *DefaultLogger) ErrorEnabled() bool {
	return logger.loggable(LevelError)
}

// SetParent sets the parent of current logger.
func (logger *DefaultLogger) SetParent(parent *DefaultLogger) {
	if parent != logger {
		logger.parent = parent
	}
}

// Level returns level of this logger or parent if not set.
func (logger *DefaultLogger) Level() Level {
	if logger.level != LevelUninitialized {
		return logger.level
	}
	if logger.parent != nil {
		return logger.parent.Level()
	}
	return LevelOff
}

// SetLevel changes logging level of this logger.
func (logger *DefaultLogger) SetLevel(level Level) {
	logger.level = level
}

// Formatter returns formatter of this logger or parent if not set.
func (logger *DefaultLogger) Formatter() Formatter {
	if logger.formatter != nil {
		return logger.formatter
	}
	if logger.parent != nil {
		return logger.parent.Formatter()
	}
	return logger.formatter
}

// SetFormatter changes formatter of this logger.
func (logger *DefaultLogger) SetFormatter(formatter Formatter) {
	logger.formatter = formatter
}

// Appender returns appender of this logger or parent if not set.
func (logger *DefaultLogger) Appender() Appender {
	if logger.appender != nil {
		return logger.appender
	}
	if logger.parent != nil {
		return logger.parent.Appender()
	}
	return logger.appender
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
func (logger *DefaultLogger) log(level Level, format string, args []interface{}) {
	if !logger.loggable(level) {
		return
	}
	formatter := logger.Formatter()
	appender := logger.Appender()
	if formatter == nil || appender == nil {
		return
	}
	event := LoggingEvent{
		Time:      time.Now(),
		Name:      logger.name,
		Level:     level,
		Format:    format,
		Arguments: args,
	}
	formatter.Format(&event)
	appender.Append(&event)
}

// DefaultLoggerFactory implements LoggerFactory interface.
type DefaultLoggerFactory struct {
	root *DefaultLogger

	mu      sync.Mutex
	loggers map[string]*DefaultLogger
}

// NewLoggerFactory allocates and returns new DefaultLoggerFactory.
func NewLoggerFactory(writer io.Writer) LoggerFactory {
	factory := &DefaultLoggerFactory{
		root:    NewLogger(RootLoggerName).(*DefaultLogger),
		loggers: make(map[string]*DefaultLogger),
	}
	factory.root.SetLevel(LevelDebug)
	factory.root.SetFormatter(NewFormatter())
	factory.root.SetAppender(NewAppender(writer))
	return factory
}

// GetLogger returns a new Logger or an existing one if the same name is found.
func (factory *DefaultLoggerFactory) GetLogger(name string) Logger {
	if name == "" || name == RootLoggerName {
		return factory.root
	}
	factory.mu.Lock()
	defer factory.mu.Unlock()
	logger := factory.createLogger(name, factory.getParent(name))
	return logger
}

// getParent returns parent logger for given logger.
func (factory *DefaultLoggerFactory) getParent(name string) *DefaultLogger {
	parent := factory.root
	for i, c := range name {
		// Search for "." character
		if c == '.' {
			parentName := name[0:i]
			if parentName != "" {
				parent = factory.createLogger(parentName, parent)
			}
		}
	}
	return parent
}

// createLogger creates a new logger if not exist.
func (factory *DefaultLoggerFactory) createLogger(name string, parent *DefaultLogger) *DefaultLogger {
	logger, ok := factory.loggers[name]
	if !ok {
		logger = NewLogger(name).(*DefaultLogger)
		logger.SetParent(parent)
		factory.loggers[name] = logger
	}
	return logger
}
