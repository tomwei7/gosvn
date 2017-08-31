// Package svn provides golang svn api through svn command
package svn

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/url"
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
	CMDList  = "list"
)

// SVN DIR
const (
	BranchesDir = "/branches"
	TagsDir     = "/tags"
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
	svnurl      *url.URL
	workdir     string
	svnExecPath string
	timeout     time.Duration
	echo        bool
	env         []string
	targetBase  string
}

// NewSVN new svn Instance
func NewSVN(svnurl string, opts *Options) (*SVN, error) {
	su, err := url.Parse(svnurl)
	if err != nil {
		return nil, err
	}
	basePath := su.Path
	if len(basePath) > 0 && basePath[len(basePath)-1] == '/' {
		basePath = basePath[:len(basePath)-1]
	}
	return &SVN{
		svnurl:      su,
		targetBase:  fmt.Sprintf("%s://%s%s", su.Scheme, su.Host, basePath),
		svnExecPath: "svn",
	}, nil
}

// Blame file
func (s *SVN) Blame(path string) (br *BlameResp, err error) {
	br = &BlameResp{}
	err = s.execTOXML(br, CMDBlame, s.targetBase+path)
	return
}

// List Dir
func (s *SVN) List(path string) (lr *ListResp, err error) {
	lr = &ListResp{}
	err = s.execTOXML(lr, CMDList, s.targetBase+path)
	return
}

// Branches list all branches
// if you want to use this branches and tags api, the svn repo must have this directory struct
//.
//├── branches
//│   └── develop
//├── tags
//│   └── v0.1
//└── trunk
//    └── test.md
func (s *SVN) Branches() ([]string, error) {
	return s.listDir(BranchesDir)
}

// Tags list all tag
func (s *SVN) Tags() ([]string, error) {
	return s.listDir(TagsDir)
}

func (s *SVN) listDir(path string) ([]string, error) {
	lr, err := s.List(BranchesDir)
	if err != nil {
		return nil, err
	}
	dirs := make([]string, 0, len(lr.Files))
	for i := range lr.Files {
		if lr.Files[i].Kind == KindDir {
			dirs = append(dirs, lr.Files[i].Name)
		}
	}
	return dirs, nil
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
	if username := s.svnurl.User.Username(); username != "" {
		arg = append(arg, "--username", username)
		if password, ok := s.svnurl.User.Password(); ok {
			arg = append(arg, "--password", password)
		}
	}
	return arg
}

func (s *SVN) execTOXML(v interface{}, cmd string, arg ...string) error {
	carg := make([]string, len(arg)+1)
	carg[0] = "--xml"
	copy(carg[1:], arg)
	data, err := s.execCMD(cmd, carg...)
	if err != nil {
		return err
	}
	return xml.Unmarshal(data, v)
}

func (s *SVN) execCMD(cmd string, arg ...string) ([]byte, error) {
	garg := s.globalArg()
	combinedArg := make([]string, len(arg)+len(garg)+1)
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
	c.Dir = s.workdir
	var td bool
	s.setTimeout(c, &td)
	if err := c.Run(); err != nil || td {
		fullcmd := fmt.Sprintf("%s %s", s.svnExecPath, strings.Join(combinedArg, " "))
		if td {
			return nil, fmt.Errorf("cmd: %s timeout", fullcmd)
		}
		return nil, fmt.Errorf("exec error cmd: %s\n, err: %s stderr:\n%s", fullcmd, err.Error(), stderr.String())
	}
	return ioutil.ReadAll(stdout)
}
