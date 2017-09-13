// Package svn provides svn date reference
package svn

import (
	"encoding/xml"
	"time"
)

// CMD for svn
const (
	cmdAdd        = "add"
	cmdAuth       = "auth"
	cmdBlame      = "blame"
	cmdCat        = "cat"
	cmdChangelist = "changelist"
	cmdCheckout   = "checkout"
	cmdCleanup    = "cleanup"
	cmdCommit     = "commit"
	cmdCopy       = "copy"
	cmdDelete     = "delete"
	cmdDiff       = "diff"
	cmdExport     = "export"
	cmdHelp       = "help"
	cmdImport     = "import"
	cmdInfo       = "info"
	cmdList       = "list"
	cmdLock       = "lock"
	cmdLog        = "log"
	cmdMerge      = "merge"
	cmdMergeinfo  = "mergeinfo"
	cmdMkdir      = "mkdir"
	cmdMove       = "move"
	cmdPatch      = "patch"
	cmdPropdel    = "propdel"
	cmdPropedit   = "propedit"
	cmdPropget    = "propget"
	cmdProplist   = "proplist"
	cmdPropset    = "propset"
	cmdRelocate   = "relocate"
	cmdResolve    = "resolve"
	cmdResolved   = "resolved"
	cmdRevert     = "revert"
	cmdStatus     = "status"
	cmdSwitch     = "switch"
	cmdUnlock     = "unlock"
	cmdUpdate     = "update"
	cmdUpgrade    = "upgrade"
)

// Kind
const (
	KindFile = "file"
	KindDir  = "dir"
)

// BlameResp Blame
type BlameResp struct {
	XMLName     xml.Name `xml:"blame"`
	BlameTarget Target   `xml:"target"`
}

// Target recode info for each line
type Target struct {
	Path   string  `xml:"path,attr"`
	Entrys []Entry `xml:"entry"`
}

// Entry LineInfo
type Entry struct {
	LineNumber string  `xml:"line-number,attr"`
	Commit     CommitT `xml:"commit"`
}

// CommitT Commit Type
type CommitT struct {
	Revision string    `xml:"revision,attr"`
	Auther   string    `xml:"author"`
	DateTime time.Time `xml:"date"`
}

// File snv file type
type File struct {
	Kind   string  `xml:"kind,attr"`
	Commit CommitT `xml:"commit"`
	Name   string  `xml:"name"`
}

// ListResp ListResp
type ListResp struct {
	XMLName xml.Name `xml:"lists"`
	Files   []File   `xml:"list>entry"`
}

// LogResp svn log struct
type LogResp struct {
	XMLName   xml.Name   `xml:"log"`
	Logentrys []Logentry `xml:"logentry"`
}

// Logentry svn logentry
type Logentry struct {
	Revision string    `xml:"revision,attr"`
	Author   string    `xml:"author"`
	DateTime time.Time `xml:"date"`
	Msg      string    `xml:"msg"`
	Paths    []Path    `xml:"paths>path"`
}

// Path svn path
type Path struct {
	Action   string `xml:"action,attr"`
	PropMods string `xml:"prop-mods,attr"`
	TextMods string `xml:"text-mods,attr"`
	Kind     string `xml:"kind,attr"`
	Value    string `xml:",chardata"`
}

// InfoResp InfoResp
type InfoResp struct {
	XMLName xml.Name `xml:"info"`
	Info    InfoT    `xml:"entry"`
}

// InfoT svn info
type InfoT struct {
	Kind        string `xml:"kind,attr"`
	Path        string `xml:"path,attr"`
	Revision    string `xml:"revision,attr"`
	URL         string `xml:"url"`
	RelativeURL string `xml:"relative-url"`
	Repository  struct {
		Root string `xml:"root"`
		uuid string `xml:"uuid"`
	} `xml:"repository"`
	Commit CommitT `xml:"commit"`
}
