/*
Package filter provides an appender which has filter.
*/
package filter

import (
	"sort"

	"github.com/goburrow/gol"
)

// Appender is an logging appender which supports level threshold, inclusive or
// exclusive logger name.
type Appender struct {
	appender gol.Appender

	threshold gol.Level
	// Make sure includes and excludes are sorted as it relies on binary search
	// to check if the logger is in the list.
	includes []string
	excludes []string
}

var _ (gol.Appender) = (*Appender)(nil)

// NewAppender allocates and returns a new Appender
func NewAppender(a gol.Appender) *Appender {
	return &Appender{
		appender:  a,
		threshold: gol.All,
	}
}

// Append only send logging event to the assigned appender only when:
// - Logging level >= Threshold
// - Logger name is not in the excludes
// - Logger name is in the includes
func (a *Appender) Append(e *gol.LoggingEvent) {
	if e.Level < a.threshold {
		return
	}
	if len(a.excludes) > 0 {
		idx := sort.SearchStrings(a.excludes, e.Name)
		if idx < len(a.excludes) && a.excludes[idx] == e.Name {
			// Excluded
			return
		}
	}
	if len(a.includes) > 0 {
		idx := sort.SearchStrings(a.includes, e.Name)
		if !(idx < len(a.includes) && a.includes[idx] == e.Name) {
			// Not included
			return
		}
	}
	a.appender.Append(e)
}

// SetThreshold change logging threshold.
func (a *Appender) SetThreshold(t gol.Level) {
	a.threshold = t
}

// SetIncludes set inclusive logger names.
func (a *Appender) SetIncludes(includes ...string) {
	in := make([]string, len(includes))
	copy(in, includes)
	sort.Strings(in)
	a.includes = in
}

// SetExcludes set exclusive logger names.
func (a *Appender) SetExcludes(excludes ...string) {
	ex := make([]string, len(excludes))
	copy(ex, excludes)
	sort.Strings(ex)
	a.excludes = ex
}
