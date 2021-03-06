package gol

// Logger specifies how logging in application is done.
type Logger interface {
	Tracef(string, ...interface{})
	TraceEnabled() bool
	Debugf(string, ...interface{})
	DebugEnabled() bool
	Infof(string, ...interface{})
	InfoEnabled() bool
	Warnf(string, ...interface{})
	WarnEnabled() bool
	Errorf(string, ...interface{})
	ErrorEnabled() bool
}

// Factory produces Logger.
type Factory interface {
	GetLogger(name string) Logger
}
