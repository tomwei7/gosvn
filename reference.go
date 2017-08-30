// Package svn provides svn date reference
package svn

import (
	"encoding/xml"
	"time"
)

// NewBlameResp from give xml data
func NewBlameResp(data []byte) (*BlameResp, error) {
	br := &BlameResp{}
	err := xml.Unmarshal(data, br)
	return br, err
}

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
	LineNumber string `xml:"line-number,attr"`
	Commit     struct {
		Revision string    `xml:"revision,attr"`
		Auther   string    `xml:"author"`
		DateTime time.Time `xml:"date"`
	} `xml:"commit"`
}
