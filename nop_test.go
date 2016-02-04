package gol

import "testing"

func TestNOPLogger(t *testing.T) {
	if NOPLogger.TraceEnabled() ||
		NOPLogger.DebugEnabled() ||
		NOPLogger.InfoEnabled() ||
		NOPLogger.WarnEnabled() ||
		NOPLogger.ErrorEnabled() {
		t.Fatal("logger must not enabled")
	}
	NOPLogger.Tracef("trace")
	NOPLogger.Debugf("debug")
	NOPLogger.Infof("info")
	NOPLogger.Warnf("warn")
	NOPLogger.Errorf("error")
}
