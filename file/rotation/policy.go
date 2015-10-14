package rotation

import (
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	timeFormat = "2006-01-02"
)

var (
	// NoPolicy never triggers and does nothing when rolling.
	NoPolicy       = &noPolicy{}
	errSamePattern = errors.New("archived file pattern same to current active file")
)

// TriggeringPolicy controls how rollingPolicy is activated.
type TriggeringPolicy interface {
	IsTriggering(*os.File, []byte) bool
}

// RollingPolicy rolls over the log file.
type RollingPolicy interface {
	Rollover(*os.File) error
}

// TriggerTimer returns trigger time
type TriggerTimer interface {
	TriggerTime() time.Time
}

type triggerTimerFunc func() time.Time

func (f triggerTimerFunc) TriggerTime() time.Time {
	return f()
}

// noPolicy is a TriggeringPolicy and RollingPolicy which does nothing.
type noPolicy struct {
}

func (*noPolicy) IsTriggering(*os.File, []byte) bool {
	return false
}

func (*noPolicy) Rollover(*os.File) error {
	return nil
}

// TimeTriggeringPolicy triggers when day changed.
// TODO: able to specify daily, weekly or monthly.
type TimeTriggeringPolicy struct {
	mu           sync.Mutex
	isTriggering bool
	triggerTime  time.Time

	timer  *time.Timer
	finish chan struct{}

	currentTime func() time.Time
}

var _ TriggerTimer = (*TimeTriggeringPolicy)(nil)

// NewTimeTriggeringPolicy allocates and returns a TimeTriggeringPolicy.
func NewTimeTriggeringPolicy() *TimeTriggeringPolicy {
	return &TimeTriggeringPolicy{
		currentTime: time.Now,
	}
}

// IsTriggering is called in Appender so it's only happens when an logging event
// happens.
func (p *TimeTriggeringPolicy) IsTriggering(*os.File, []byte) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	// isTriggering is set by timer function
	if p.isTriggering {
		// Toggle it
		p.isTriggering = false
		return true
	}
	return false
}

// Start starts timer with current local time.
func (p *TimeTriggeringPolicy) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.finish = make(chan struct{})
	p.startTimer()
	return nil
}

// Stop stops running timer.
func (p *TimeTriggeringPolicy) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.timer != nil {
		p.timer.Stop()
		close(p.finish)
		p.timer = nil
	}
	return nil
}

// startTimer must be called with p.mu held.
func (p *TimeTriggeringPolicy) startTimer() {
	now := p.currentTime()
	// Next day
	tmr := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 1E3, time.Local)
	p.timer = time.NewTimer(tmr.Sub(now))

	go p.checkTriggering()
}

// TriggerTime returns the time trigger event was raised.
func (p *TimeTriggeringPolicy) TriggerTime() time.Time {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.triggerTime
}

// checkTriggering must be called in a go routine.
func (p *TimeTriggeringPolicy) checkTriggering() {
	timer := p.timer
	select {
	case <-p.finish:
		return
	case tm := <-timer.C:
		p.mu.Lock()
		defer p.mu.Unlock()

		p.isTriggering = true
		p.triggerTime = tm
		p.startTimer()
	}
}

// TimeRollingPolicy allows the roll over to be based on time.
// TODO: able to specify format.
type TimeRollingPolicy struct {
	FilePattern string
	FileCount   int

	TriggerTimer TriggerTimer
}

// NewTimeRollingPolicy allocates and returns a new TimeRollingPolicy.
func NewTimeRollingPolicy() *TimeRollingPolicy {
	return &TimeRollingPolicy{
		// TODO: Associate with a TimeTriggeringPolicy
		TriggerTimer: triggerTimerFunc(time.Now),
	}
}

// Rollover rolls the log file.
func (p *TimeRollingPolicy) Rollover(f *os.File) error {
	var pattern string
	if p.FilePattern != "" {
		pattern = p.FilePattern
	} else {
		pattern = defaultFilePattern(f.Name())
	}

	// Previous day is used here actually as we want log file is generated with
	// the day before the event happenned.
	triggerTime := p.TriggerTimer.TriggerTime().AddDate(0, 0, -1)
	// Remove history
	if p.FileCount > 0 {
		timestamp := triggerTime.AddDate(0, 0, -p.FileCount).Format(timeFormat)
		name := fmt.Sprintf(pattern, timestamp)
		if fileExists(name) {
			if err := os.Remove(name); err != nil {
				return err
			}
		}
	}

	// Archive current file
	name := fmt.Sprintf(pattern, triggerTime.Format(timeFormat))
	target, err := os.Create(name)
	if err != nil {
		return err
	}
	defer target.Close()
	if err = archiveFile(target, f); err != nil {
		return err
	}
	// Clear content
	if err = f.Truncate(0); err != nil {
		return err
	}
	_, err = f.Seek(0, os.SEEK_SET)
	return err
}

// defaultFilePattern returns pattern generated from the given file name.
// e.g. ../file.log => ../file-%s.log.gz
func defaultFilePattern(name string) string {
	var buf bytes.Buffer

	ext := filepath.Ext(name)
	buf.WriteString(name[:len(name)-len(ext)])
	buf.WriteString("-%s")
	buf.WriteString(ext)
	buf.WriteString(".gz")
	return buf.String()
}

// Check if a file exists.
func fileExists(file string) bool {
	st, err := os.Stat(file)
	if err == nil && !st.IsDir() {
		return true
	}
	return false
}

// archiveFile compresses src if dst extention is .gz. Otherwise, it will copy
// src to dst.
func archiveFile(dst *os.File, src *os.File) error {
	var err error
	_, err = src.Seek(0, os.SEEK_SET)
	if err != nil {
		return err
	}
	// Compress if necessary
	switch filepath.Ext(dst.Name()) {
	case ".gz":
		err = gzipCopy(dst, src)
	default:
		// FIXME: move file only
		_, err = io.Copy(dst, src)
	}
	if err != nil {
		// Need to seek back to end of file
		src.Seek(0, os.SEEK_END)
	}
	return err
}

func gzipCopy(dst io.Writer, src io.Reader) error {
	w := gzip.NewWriter(dst)
	defer w.Close()

	var err error
	_, err = io.Copy(w, src)
	return err
}
