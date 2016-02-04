package gol

var (
	NOPLogger Logger = (*nopLogger)(nil)
)

type nopLogger struct{}

func (*nopLogger) Tracef(string, ...interface{}) {}

func (*nopLogger) TraceEnabled() bool {
	return false
}

func (*nopLogger) Debugf(string, ...interface{}) {}

func (*nopLogger) DebugEnabled() bool {
	return false
}

func (*nopLogger) Infof(string, ...interface{}) {}

func (*nopLogger) InfoEnabled() bool {
	return false
}

func (*nopLogger) Warnf(string, ...interface{}) {}

func (*nopLogger) WarnEnabled() bool {
	return false
}

func (*nopLogger) Errorf(string, ...interface{}) {}

func (*nopLogger) ErrorEnabled() bool {
	return false
}
