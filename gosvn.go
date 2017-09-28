// Package svn provides golang svn api through svn command
package svn

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path"
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

// repo type
const (
	LocalRepo int = iota
	RemoteRepo
)

// SVN DIR
const (
	DefaultBranchesDir = "branches"
	DefaultTagsDir     = "tags"
	DefaultTrunkDir    = "trunk"
)

// ErrRepoTypeLocal Err when you want exec local repo only cmd, e.g. add
var ErrRepoTypeLocal = fmt.Errorf("cmd not available, local repo only")

// ErrRepoTypeRemote Err when you want exec remote repo only cmd, e.g. checkout
var ErrRepoTypeRemote = fmt.Errorf("cmd not available, remote repo only")

// Options svn options
type Options struct {
	SVNExecPath     string
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
	ConfigOption   string
	BranchesDir    string
	TagsDir        string
	TrunkDir       string
	Username       string
	Password       string
	NoBranchesTags bool
}

// SVN struct
type SVN struct {
	svnurl         *url.URL
	workDir        string
	svnExecPath    string
	timeout        time.Duration
	echo           bool
	env            []string
	targetBase     string
	globalArg      []string
	localRepo      bool
	branchesDir    string
	tagsDir        string
	trunkDir       string
	NoBranchesTags bool
}

// NewSVN new svn Instance
func NewSVN(svnurl string, opts *Options) (*SVN, error) {
	su, err := url.Parse(svnurl)
	if err != nil {
		return nil, err
	}
	basePath := su.Path
	if len(basePath) == 0 || basePath[len(basePath)-1] != '/' {
		basePath = basePath + "/"
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
	if opts.BranchesDir == "" {
		opts.BranchesDir = DefaultBranchesDir
	}
	if opts.TagsDir == "" {
		opts.TagsDir = DefaultTagsDir
	}
	if opts.TrunkDir == "" {
		opts.TrunkDir = DefaultTrunkDir
	}
	if opts.SVNExecPath == "" {
		opts.SVNExecPath = "svn"
	}
	return (&SVN{
		svnurl:         su,
		targetBase:     targetBase,
		svnExecPath:    opts.SVNExecPath,
		localRepo:      localRepo,
		workDir:        workDir,
		branchesDir:    opts.BranchesDir,
		tagsDir:        opts.TagsDir,
		trunkDir:       opts.TrunkDir,
		NoBranchesTags: opts.NoBranchesTags,
	}).initGlobalArg(opts), nil
}

// Kind return svn kind LocalRepo or RemoteRepo
func (s *SVN) Kind() int {
	if s.localRepo {
		return LocalRepo
	}
	return RemoteRepo
}

// local path only command

// Add file
func (s *SVN) Add(localPath string, opts map[string]interface{}) error {
	if !s.localRepo {
		return ErrRepoTypeLocal
	}
	_, err := s.execCMD(cmdAdd, localPath)
	return err
}

// Checkout repo to local
func (s *SVN) Checkout(remotePath, localPath string, args ...string) error {
	if s.localRepo {
		return ErrRepoTypeRemote
	}
	args = append(args, s.targetBase, localPath)
	_, err := s.execCMD(cmdCheckout, args...)
	return err
}

// Export repo
func (s *SVN) Export(remotePath, localPath string, args ...string) error {
	if s.localRepo {
		return ErrRepoTypeRemote
	}
	args = append(args, s.targetBase+remotePath, localPath)
	_, err := s.execCMD(cmdExport, args...)
	return err
}

// Commit file
func (s *SVN) Commit(path, msg string, opts map[string]interface{}) error {
	if !s.localRepo {
		return ErrRepoTypeLocal
	}
	_, err := s.execCMD(cmdCommit, path, "-m", msg)
	return err
}

// Cleanup Cleanup
func (s *SVN) Cleanup(localPath string) error {
	if !s.localRepo {
		return ErrRepoTypeLocal
	}
	_, err := s.execCMD(cmdCleanup, localPath)
	return err
}

// Copy Copy
func (s *SVN) Copy(src, dst, msg string) error {
	_, err := s.execCMD(cmdCopy, s.targetBase+src, s.targetBase+dst, "-m", msg)
	return err
}

// remote able command

// Blame file
func (s *SVN) Blame(path string) (br *BlameResp, err error) {
	br = &BlameResp{}
	err = s.execTOXML(br, cmdBlame, s.targetBase+path)
	return
}

// List Dir
func (s *SVN) List(path string) (lr *ListResp, err error) {
	lr = &ListResp{}
	err = s.execTOXML(lr, cmdList, s.targetBase+path)
	return
}

// Mkdir Mkdir
func (s *SVN) Mkdir(path string) error {
	_, err := s.execCMD(cmdMkdir, s.targetBase+path)
	return err
}

// Log Log
//  -r [--revision] ARG      : ARG (some commands also take ARG1:ARG2 range)
//                             A revision argument can be one of:
//                                NUMBER       revision number
//                                '{' DATE '}' revision at start of the date
//                                'HEAD'       latest in repository
//                                'BASE'       base rev of item's working copy
//                                'COMMITTED'  last commit at or before BASE
//                                'PREV'       revision just before COMMITTED
func (s *SVN) Log(path string, args ...string) (lor *LogResp, err error) {
	lor = &LogResp{}
	args = append(args, "-v", s.targetBase+path)
	// -v show detail log
	err = s.execTOXML(lor, cmdLog, args...)
	return
}

// Info Info
func (s *SVN) Info(path string) (ir *InfoResp, err error) {
	ir = &InfoResp{}
	err = s.execTOXML(ir, cmdInfo, s.targetBase+path)
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
	return s.listDir(s.branchesDir)
}

// NewBranch new brach from trunk
func (s *SVN) NewBranch(name, msg string) error {
	return s.Copy(s.trunkDir, path.Join(s.branchesDir, name), msg)
}

// BranchesDir BranchesDir
func (s *SVN) BranchesDir() string {
	return s.branchesDir
}

// Tags list all tag
func (s *SVN) Tags() ([]string, error) {
	return s.listDir(s.tagsDir)
}

// NewTag new tag from trunk
func (s *SVN) NewTag(name, msg string) error {
	return s.Copy(s.trunkDir, path.Join(s.tagsDir, name), msg)
}

// TagsDir TagsDir
func (s *SVN) TagsDir() string {
	return s.tagsDir
}

// TrunkDir Trunk Dir
func (s *SVN) TrunkDir() string {
	return s.trunkDir
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
	<-timer.C
	// FIXME deal kill error
	c.Process.Kill()
	*td = true
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
	} else if opts.Username != "" && opts.Password != "" {
		arg = append(arg, "--username", opts.Username)
		arg = append(arg, "--password", opts.Password)
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
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	ctx := context.TODO()
	if s.timeout != 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.timeout)
		defer cancel()
	}
	c := exec.CommandContext(ctx, s.svnExecPath, combinedArg...)
	c.Stdout = stdout
	c.Stderr = stderr
	c.Dir = s.workDir
	c.Env = s.env
	err := c.Run()
	fullcmd := fmt.Sprintf("%s %s", s.svnExecPath, strings.Join(combinedArg, " "))
	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("exec %s timeout", fullcmd)
	}
	if err != nil {
		return nil, NewError(fullcmd, err.Error(), stderr.String())
	}
	return ioutil.ReadAll(stdout)
}
