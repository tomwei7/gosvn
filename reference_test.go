package svn

import (
	"encoding/xml"
	"testing"
)

var testBlameXML = []byte(`<?xml version="1.0" encoding="UTF-8"?>
<blame>
  <target path="svn://unit-00/testrepo/branches/develop/test.md">
    <entry line-number="1">
      <commit revision="2">
        <author>test</author>
        <date>2017-08-30T06:26:05.592955Z</date>
      </commit>
    </entry>
    <entry line-number="2">
      <commit revision="6">
        <author>test</author>
        <date>2017-08-30T08:16:20.752100Z</date>
      </commit>
    </entry>
    <entry line-number="3">
      <commit revision="6">
        <author>test</author>
        <date>2017-08-30T08:16:20.752100Z</date>
      </commit>
    </entry>
  </target>
</blame>
`)

func TestNewBlameResp(t *testing.T) {
	br := &BlameResp{}
	if err := xml.Unmarshal(testBlameXML, br); err != nil {
		t.Errorf("Error %s", err.Error())
	}
	if len(br.BlameTarget.Entrys) != 3 {
		t.Errorf("expect 3 entry get %d", len(br.BlameTarget.Entrys))
	} else {
		entry := br.BlameTarget.Entrys[1]
		if entry.LineNumber != "2" || entry.Commit.Auther != "test" || entry.Commit.Revision != "6" {
			t.Errorf("Entry Err %+v", entry)
		}
	}
}
