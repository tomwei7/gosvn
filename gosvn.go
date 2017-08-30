// Package svn provides golang svn api through svn command
// Available subcommands:
// [ ]  add
// [ ]  auth
// [ ]  blame (praise, annotate, ann)
// [ ]  cat
// [ ]  changelist (cl)
// [ ]  checkout (co)
// [ ]  cleanup
// [ ]  commit (ci)
// [ ]  copy (cp)
// [ ]  delete (del, remove, rm)
// [ ]  diff (di)
// [ ]  export
// [ ]  help (?, h)
// [ ]  import
// [ ]  info
// [ ]  list (ls)
// [ ]  lock
// [ ]  log
// [ ]  merge
// [ ]  mergeinfo
// [ ]  mkdir
// [ ]  move (mv, rename, ren)
// [ ]  patch
// [ ]  propdel (pdel, pd)
// [ ]  propedit (pedit, pe)
// [ ]  propget (pget, pg)
// [ ]  proplist (plist, pl)
// [ ]  propset (pset, ps)
// [ ]  relocate
// [ ]  resolve
// [ ]  resolved
// [ ]  revert
// [ ]  status (stat, st)
// [ ]  switch (sw)
// [ ]  unlock
// [ ]  update (up)
// [ ]  upgrade
package svn

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"time"
)

// CAFailArg CAFailArg
type CAFailArg string

// arg for TrustServerCertFailures
const (
	CAUnknownCa   CAFailArg = "unknown-ca"
	CACnMismatch  CAFailArg = "cn-mismatch"
	CAExpired     CAFailArg = "expired"
	CANotYetValid CAFailArg = "not-yet-valid"
	CAOther       CAFailArg = "other"
)

// CMD for svn
const (
	CMDBlame = "blame"
)

// Options svn options
type Options struct {
	NoAuthCache     bool // do not cache authentication tokens
	NonInteractive  bool // do no interactive prompting (default is to prompt only if standard input is a terminal device)
	ForceInteractiv bool // do interactive prompting even if standard input is not a terminal device

	//with --non-interactive, accept SSL server
	// certificates with failures; ARG is comma-separated
	// list of 'unknown-ca' (Unknown Authority),
	// 'cn-mismatch' (Hostname mismatch), 'expired'
	// (Expired certificate), 'not-yet-valid' (Not yet
	// valid certificate) and 'other' (all other not
	// separately classified certificate errors)
	TrustServerCertFailures CAFailArg
	Echo                    bool
}

// SVN struct
type SVN struct {
	username     string
	password     string
	rawurl       string
	target       string
	workdir      string
	svnExecPath  string
	timeout      time.Duration
	authRequired bool
	echo         bool
}

// NewSVN new svn Instance
func NewSVN(rawurl string, opts *Options) (*SVN, error) {
	return &SVN{}, nil
}

// Blame file
func (s *SVN) Blame(path string) (*BlameResp, error) {
}

// kill child process when timeout
func (s *SVN) setTimeout(c *exec.Cmd, td *bool) {
	if s.timeout == 0 {
		return
	}
	timer := time.NewTimer(s.timeout)
	for _ = range timer.C {
		if c.ProcessState.Exited() {
			// FIXME deal kill error
			c.Process.Kill()
			*td = true
		}
	}
}

func (s *SVN) globalArg() []string {
	arg := make([]string, 0)
	if s.authRequired {
		arg = append(arg, "--username", s.username, "--password", s.password)
	}
	return arg
}

func (s *SVN) execCMD(cmd string, arg ...string) ([]byte, error) {
	garg := s.globalArg()
	combinedArg := make([]string, 1, len(arg)+len(garg)+1)
	combinedArg[0] = cmd
	copy(combinedArg[1:], garg)
	copy(combinedArg[1+len(garg):], arg)
	if s.echo {
		fmt.Printf("exec %s %s\n", s.svnExecPath, strings.Join(combinedArg, " "))
	}
	c := exec.Command(s.svnExecPath, combinedArg...)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	c.Stdout = stdout
	c.Stderr = stderr
	var td bool
	s.setTimeout(c, &td)
	if err := c.Run(); err != nil || td {
		fullcmd := fmt.Sprintf("%s %s %s", s.svnExecPath, cmd, strings.Join(arg, " "))
		if td {
			return nil, fmt.Errorf("cmd: %s timeout", fullcmd)
		}
		return nil, fmt.Errorf("exec error cmd: %s\n, err: %s stderr:\n%s", fullcmd, err.Error(), stderr.String())
	}
	return ioutil.ReadAll(stdout)
}
