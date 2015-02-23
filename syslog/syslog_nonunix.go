// +build !linux,!darwin

package syslog

import (
	"errors"
	"net"
)

// connectLocal is taken from log/syslog
func connectLocal(_ string, _ string) (net.Conn, error) {
	return nil, errors.New("syslog: connecting to local syslog daemon is not implemented")
}
