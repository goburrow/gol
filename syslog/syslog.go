/*
Package file provides logging to syslog.
*/
package syslog

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/goburrow/gol"
)

const (
	defaultLayout     = "<%[5]d>%[4]s %[6]s %[7]s[%[8]d]: %[2]s: %[1]s\n"
	defaultTimeLayout = time.RFC3339
	// Layout for local syslog does not have hostname and use time.Stamp
	localLayout     = "<%[5]d>%[4]s %[7]s[%[8]d]: %[2]s: %[1]s\n"
	localTimeLayout = time.Stamp

	dialTimeoutMs = 60000
)

type Facility int

// Facility
const (
	LOG_KERN Facility = iota
	LOG_USER
	LOG_MAIL
	LOG_DAEMON
	LOG_AUTH
	LOG_SYSLOG
	LOG_LPR
	LOG_NEWS
	LOG_UUCP
	LOG_CRON
	LOG_AUTHPRIV
	LOG_FTP
	_
	_
	_
	_
	LOG_LOCAL0
	LOG_LOCAL1
	LOG_LOCAL2
	LOG_LOCAL3
	LOG_LOCAL4
	LOG_LOCAL5
	LOG_LOCAL6
	LOG_LOCAL7
)

// Severity
const (
	sError = 3
	sWarn  = 4
	sInfo  = 6
	sDebug = 7
)

// Encoder provides additional arguments.
// #5: Priority,
// #6: Hostname,
// #7: Tag,
// #8: PID.
type Encoder struct {
	gol.DefaultEncoder

	Hostname string
	Tag      string
	Facility Facility
}

func NewEncoder() *Encoder {
	return &Encoder{
		DefaultEncoder: gol.DefaultEncoder{
			Layout:     defaultLayout,
			TimeLayout: defaultTimeLayout,
		},
	}
}

func (encoder *Encoder) Encode(event *gol.LoggingEvent, target io.Writer) error {
	priority := encoder.getPriority(event)
	timestamp := event.Time.Format(encoder.TimeLayout)

	var err error
	_, err = fmt.Fprintf(target, encoder.Layout,
		event.FormattedMessage,
		event.Name,
		gol.LevelString(event.Level),
		timestamp,
		priority,
		encoder.Hostname,
		encoder.Tag,
		os.Getpid(),
	)
	return err
}

func (encoder *Encoder) getPriority(event *gol.LoggingEvent) int {
	priority := int(encoder.Facility) * 8
	if event.Level >= gol.LevelError {
		priority += sError
	} else if event.Level >= gol.LevelWarn {
		priority += sWarn
	} else if event.Level >= gol.LevelInfo {
		priority += sInfo
	} else {
		priority += sDebug
	}
	return priority
}

// Appender sends logging to syslog server/daemon.
// All properties must be set before Start(), otherwise default values will be used.
type Appender struct {
	// Known networks are "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6",
	// "ip", "ip4", "ip6", "unix", "unixgram" and "unixpacket".
	// See tcp.Dial.
	Network  string
	Addr     string
	Facility Facility
	Tag      string
	Encoder  gol.Encoder

	mu   sync.Mutex
	conn net.Conn
	buf  bytes.Buffer
}

var _ gol.Appender = (*Appender)(nil)

func NewAppender() *Appender {
	return &Appender{
		Facility: LOG_LOCAL0,
	}
}

func (a *Appender) Append(event *gol.LoggingEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()

	var err error
	if a.conn == nil {
		if err = a.connect(); err != nil {
			gol.Print(err)
			return
		}
	}
	a.buf.Reset()
	if err = a.Encoder.Encode(event, &a.buf); err != nil {
		gol.Print(err)
		return
	}
	if _, err = a.conn.Write(a.buf.Bytes()); err != nil {
		gol.Print(err)
		return
	}
}

// Start connects to syslog server if not connected.
func (a *Appender) Start() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.conn != nil {
		return nil
	}
	return a.connect()
}

// Stop disconnects current connection.
func (a *Appender) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.conn != nil {
		err := a.conn.Close()
		a.conn = nil
		return err
	}
	return nil
}

// connect must be called with a.mu held.
func (a *Appender) connect() error {
	var (
		err      error
		local    bool
		conn     net.Conn
		hostname string
	)

	if a.Network == "" {
		conn, err = connectLocal(a.Network, a.Addr)
		if err != nil {
			return err
		}
		local = true
		hostname = "localhost"
	} else {
		timeout := time.Duration(dialTimeoutMs) * time.Millisecond
		dialer := net.Dialer{Timeout: timeout}
		conn, err = dialer.Dial(a.Network, a.Addr)
		if err != nil {
			return err
		}
		hostname = conn.LocalAddr().String()
	}
	a.conn = conn
	if a.Encoder == nil {
		a.Encoder = a.createEncoder(hostname, local)
	}
	return nil
}

func (a *Appender) createEncoder(hostname string, local bool) *Encoder {
	encoder := NewEncoder()
	encoder.Hostname = hostname
	encoder.Facility = a.Facility
	if a.Tag == "" {
		encoder.Tag = filepath.Base(os.Args[0])
	} else {
		encoder.Tag = a.Tag
	}
	if local {
		encoder.Layout = localLayout
		encoder.TimeLayout = localTimeLayout
	}
	return encoder
}
