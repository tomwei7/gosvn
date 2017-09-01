// Package svn provides golang svn api through svn command
package svn

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
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

// SVN DIR
const (
	BranchesDir = "/branches"
	TagsDir     = "/tags"
)

// ErrRepoType Err when you want exec local repo only cmd, e.g. add
var ErrRepoType = fmt.Errorf("cmd not available local repo only")

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
	Env                     []string
	EnvOverWrite            bool
	WorkDir                 string
	Timeout                 time.Duration
	ConfigDir               string //--config-dir read user configuration files from directory ARG
	// --config-option
	//set user configuration option in the format:
	//  FILE:SECTION:OPTION=[VALUE]
	//For example:
	//servers:global:http-library=serf
	ConfigOption string
}

// SVN struct
type SVN struct {
	svnurl      *url.URL
	workDir     string
	svnExecPath string
	timeout     time.Duration
	echo        bool
	env         []string
	targetBase  string
	globalArg   []string
	localRepo   bool
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
	localRepo := true
	var workDir string
	targetBase := basePath
	if !(su.Scheme == "file" || su.Scheme == "") {
		targetBase = fmt.Sprintf("%s://%s%s", su.Scheme, su.Host, basePath)
		localRepo = false
	} else {
		workDir = targetBase
	}
	return (&SVN{
		svnurl:      su,
		targetBase:  targetBase,
		svnExecPath: "svn",
		localRepo:   localRepo,
		workDir:     workDir,
	}).initGlobalArg(opts), nil
}

// local path only command

// Add file
func (s *SVN) Add(localPath string, opts map[string]interface{}) error {
	if !s.localRepo {
		return ErrRepoType
	}
	_, err := s.execCMD(CMDAdd, localPath)
	return err
}

// Commit file
func (s *SVN) Commit(localPath, msg string, opts map[string]interface{}) error {
	if !s.localRepo {
		return ErrRepoType
	}
	_, err := s.execCMD(CMDCommit, localPath, "-m", msg)
	return err
}

// Cleanup Cleanup
func (s *SVN) Cleanup(localPath string, opts map[string]interface{}) error {
	if !s.localRepo {
		return ErrRepoType
	}
	_, err := s.execCMD(CMDCleanup, localPath)
	return err
}

// remote able command

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

// Log Log
func (s *SVN) Log(path string) (lor *LogResp, err error) {
	lor = &LogResp{}
	// -v show detail log
	err = s.execTOXML(lor, CMDLog, "-v", s.targetBase+path)
	return
}

// Info Info
func (s *SVN) Info(path string) (ir *InfoResp, err error) {
	ir = &InfoResp{}
	err = s.execTOXML(ir, CMDInfo, s.targetBase+path)
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
	lr, err := s.List(path)
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

func (s *SVN) initGlobalArg(opts *Options) *SVN {
	if opts == nil {
		return s
	}
	arg := make([]string, 0)
	if s.svnurl.User != nil {
		if username := s.svnurl.User.Username(); username != "" {
			arg = append(arg, "--username", username)
			if password, ok := s.svnurl.User.Password(); ok {
				arg = append(arg, "--password", password)
			}
		}
	}
	if opts.ForceInteractiv {
		arg = append(arg, "--force-interactive")
	}
	if opts.NoAuthCache {
		arg = append(arg, "--no-auth-cache")
	}
	if opts.NonInteractive {
		arg = append(arg, "--non-interactive")
	}
	if opts.ConfigDir != "" {
		arg = append(arg, "--config-dir", opts.ConfigDir)
	}
	if opts.ConfigOption != "" {
		arg = append(arg, "--config-option", opts.ConfigOption)
	}
	switch opts.TrustServerCertFailures {
	case CAUnknownCa, CAOther, CAExpired, CACnMismatch, CANotYetValid:
		arg = append(arg, "--trust-server-cert-failures", string(opts.TrustServerCertFailures))
	}
	s.timeout = opts.Timeout
	s.globalArg = arg
	s.echo = opts.Echo
	s.env = os.Environ()
	if opts.WorkDir != "" {
		s.workDir = opts.WorkDir
	}
	if opts.EnvOverWrite {
		s.env = opts.Env
	} else {
		s.env = append(s.env, opts.Env...)
	}
	return s
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
	combinedArg := make([]string, len(arg)+len(s.globalArg)+1)
	combinedArg[0] = cmd
	copy(combinedArg[1:], s.globalArg)
	copy(combinedArg[1+len(s.globalArg):], arg)
	if s.echo {
		fmt.Printf("exec %s %s\n", s.svnExecPath, strings.Join(combinedArg, " "))
	}
	c := exec.Command(s.svnExecPath, combinedArg...)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	c.Stdout = stdout
	c.Stderr = stderr
	c.Dir = s.workDir
	c.Env = s.env
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
