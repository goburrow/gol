// +build linux darwin

package syslog

import (
	"errors"
	"net"
	"time"
)

// connectLocal is taken from log/syslog
func connectLocal(_ string, _ string) (net.Conn, error) {
	logTypes := []string{"unixgram", "unix"}
	logPaths := []string{"/dev/log", "/var/run/syslog"}

	timeout := time.Duration(dialTimeoutMs) * time.Millisecond
	dialer := net.Dialer{Timeout: timeout}

	for _, network := range logTypes {
		for _, path := range logPaths {
			conn, err := dialer.Dial(network, path)
			if err == nil {
				return conn, nil
			}
		}
	}
	return nil, errors.New("gol: unix syslog delivery error")
}
