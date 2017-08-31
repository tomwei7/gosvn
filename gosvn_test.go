package svn

import (
	"log"
	"os"
	"testing"
)

var svn *SVN
var svnurl string

func init() {
	svnurl = os.Getenv("TEST_SVNURL")
	var err error
	svn, err = NewSVN(svnurl, &Options{Echo: true})
	if err != nil {
		log.Fatal(err)
	}
}

func TestBlame(t *testing.T) {
	_, err := svn.Blame("/trunk/test.md")
	if err != nil {
		t.Error(err)
	}
}

func TestList(t *testing.T) {
	ret, err := svn.List("/")
	if err != nil {
		t.Error(err)
	}
	if len(ret.Files) != 3 {
		t.Errorf("%+v", *ret)
	}
}

func TestBranches(t *testing.T) {
	ret, err := svn.Branches()
	if err != nil {
		t.Error(err)
	}
	if len(ret) != 2 {
		t.Errorf("%+v", ret)
	}
}

func TestTags(t *testing.T) {
	ret, err := svn.Tags()
	if err != nil {
		t.Error(err)
	}
	if len(ret) != 2 {
		t.Errorf("%+v", ret)
	}
}

func TestLog(t *testing.T) {
	ret, err := svn.Log("/trunk/testlog.txt")
	if err != nil {
		t.Error(err)
	}
	if len(ret.Logentrys) != 2 {
		t.Errorf("%+v", ret)
	}
}

func TestInfo(t *testing.T) {
	ret, err := svn.Info("/trunk/testlog.txt")
	if err != nil {
		t.Error(err)
	}
	t.Log("%v", ret)
}
