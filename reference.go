// Package svn provides svn date reference
package svn

import (
	"encoding/xml"
	"time"
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
