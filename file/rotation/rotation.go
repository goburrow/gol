/*
Package rotation provides triggering and rolling policy for system files.
*/
package rotation

import (
	"io"
	"os"
)

const (
	openFlag = os.O_RDWR | os.O_CREATE | os.O_APPEND
	openMode = 0644
)

// File is a wrapper for file writer which supports rotation. Operations on
// File are not thread-safe.
type File struct {
	name string
	file *os.File

	triggeringPolicy TriggeringPolicy
	rollingPolicy    RollingPolicy
}

// NewFile creates a new file writer with the given name. Open() must be called
// before writing.
func NewFile(name string) *File {
	return &File{
		name:             name,
		triggeringPolicy: NoPolicy,
		rollingPolicy:    NoPolicy,
	}
}

var _ io.Writer = (*File)(nil)

// Write must be called after Open().
func (f *File) Write(b []byte) (int, error) {
	if f.triggeringPolicy.IsTriggering(f.file, b) {
		if err := f.rollingPolicy.Rollover(f.file); err != nil {
			return 0, err
		}
	}
	return f.file.Write(b)
}

// Open opens the file.
func (f *File) Open() error {
	return f.OpenFile(openFlag, openMode)
}

// OpenFile opens the file with the given flag and permission.
func (f *File) OpenFile(flag int, perm os.FileMode) error {
	file, err := os.OpenFile(f.name, flag, perm)
	if err != nil {
		return err
	}
	f.file = file
	return nil
}

// IsOpenned checks whether file is openned.
func (f *File) IsOpenned() bool {
	return f.file != nil
}

// Close closes the openning file.
func (f *File) Close() error {
	if f.file == nil {
		return nil
	}
	err := f.file.Close()
	f.file = nil
	return err
}

// SetTriggeringPolicy sets a new TriggeringPolicy.
func (f *File) SetTriggeringPolicy(p TriggeringPolicy) {
	f.triggeringPolicy = p
}

// SetRollingPolicy sets a new RollingPolicy.
func (f *File) SetRollingPolicy(p RollingPolicy) {
	f.rollingPolicy = p
}
