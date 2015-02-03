package gol

// Logger specifies how logging in application is done.
type Logger interface {
	Trace(string, ...interface{})
	TraceEnabled() bool
	Debug(string, ...interface{})
	DebugEnabled() bool
	Info(string, ...interface{})
	InfoEnabled() bool
	Warn(string, ...interface{})
	WarnEnabled() bool
	Error(string, ...interface{})
	ErrorEnabled() bool
}

// LoggerFactory produces Logger.
type LoggerFactory interface {
	GetLogger(name string) Logger
}
