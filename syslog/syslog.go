/*
Package syslog provides logging to syslog.
*/
package syslog

import (
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
	// #5: Priority,
	// #6: Hostname,
	// #7: Tag,
	// #8: PID.
	defaultLayout     = "<%[5]d>%[4]s %[6]s %[7]s[%[8]d]: %[2]s: %[1]s\n"
	defaultTimeLayout = time.RFC3339
	// Layout for local syslog does not have hostname and use time.Stamp
	localLayout     = "<%[5]d>%[4]s %[7]s[%[8]d]: %[2]s: %[1]s\n"
	localTimeLayout = time.Stamp

	dialTimeoutMs = 60000
)

// Facility is the syslog facility.
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

	hostname   string
	layout     string
	timeLayout string

	mu   sync.Mutex
	conn io.WriteCloser
}

var _ gol.Appender = (*Appender)(nil)

// NewAppender allocates and returns a new Appender.
func NewAppender() *Appender {
	return &Appender{
		Facility: LOG_LOCAL0,

		layout:     defaultLayout,
		timeLayout: defaultTimeLayout,
	}
}

// Append encodes the given logging event and sends to syslog connection.
func (a *Appender) Append(event *gol.LoggingEvent) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.conn == nil {
		return
	}

	priority := a.getPriority(event)
	timestamp := event.Time.Format(a.timeLayout)

	_, err := fmt.Fprintf(a.conn, a.layout,
		&event.Message,
		event.Name,
		gol.LevelString(event.Level),
		timestamp,
		priority,
		a.hostname,
		a.Tag,
		os.Getpid(),
	)
	if err != nil {
		gol.Print(err)
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
	if a.Network == "" {
		conn, err := connectLocal(a.Network, a.Addr)
		if err != nil {
			return err
		}
		a.hostname = "localhost"
		a.layout = localLayout
		a.timeLayout = localTimeLayout
		a.conn = conn
	} else {
		timeout := time.Duration(dialTimeoutMs) * time.Millisecond
		dialer := net.Dialer{Timeout: timeout}
		conn, err := dialer.Dial(a.Network, a.Addr)
		if err != nil {
			return err
		}
		a.hostname = conn.LocalAddr().String()
		a.layout = defaultLayout
		a.timeLayout = defaultTimeLayout
		a.conn = conn
	}
	if a.Tag == "" {
		a.Tag = filepath.Base(os.Args[0])
	}
	return nil
}

func (a *Appender) getPriority(event *gol.LoggingEvent) int {
	priority := int(a.Facility) * 8
	if event.Level >= gol.Error {
		priority += sError
	} else if event.Level >= gol.Warn {
		priority += sWarn
	} else if event.Level >= gol.Info {
		priority += sInfo
	} else {
		priority += sDebug
	}
	return priority
}
