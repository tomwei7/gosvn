package svn

import (
	"fmt"
	"strings"
)

// Err single svn err
type Err struct {
	Code    string
	Message string
}

// Error svn error
type Error struct {
	Cmd       string
	RawErrStr string
	Errs      []Err
	RawStr    string
}

// Error implement error interface
func (e Error) Error() string {
	return fmt.Sprintf("exec error cmd: %s\n, err: %s stderr:\n%s", e.Cmd, e.RawErrStr, e.RawStr)
}

// HasErr check err exists by code
func (e Error) HasErr(c string) bool {
	for i := range e.Errs {
		if e.Errs[i].Code == c {
			return true
		}
	}
	return false
}

// NewError new svn error
func NewError(cmd, err, stderr string) Error {
	lines := strings.Split(stderr, "\n")
	e := Error{Cmd: cmd, RawStr: stderr, RawErrStr: err, Errs: make([]Err, 0, len(lines))}
	for i := range lines {
		s := strings.SplitN(lines[i], ": ", 3)
		if len(s) == 3 {
			e.Errs = append(e.Errs, Err{Code: s[1], Message: s[2]})
		}
	}
	return e
}
