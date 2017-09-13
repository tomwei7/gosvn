package svn

import (
	"io/ioutil"
	"log"
	"os"
	"path"
	"testing"
	"time"
)

const localSVNPath = "testrepo"

var svn *SVN
var localSVN *SVN
var svnurl string
var workDir string

func init() {
	svnurl = os.Getenv("TEST_SVNURL")
	workDir = os.Getenv("WORKDIR")
	var err error
	opts := &Options{Echo: os.Getenv("ECHO") != "", WorkDir: workDir, Timeout: 3 * time.Second, Username: os.Getenv("SVNUSER"), Password: os.Getenv("SVNPASSWD")}
	svn, err = NewSVN(svnurl, opts)
	if err != nil {
		log.Fatal(err)
	}
	localSVN, err = NewSVN(localSVNPath, opts)
	if err != nil {
		log.Fatal(err)
	}
	InitLocalRepo()
}

func InitLocalRepo() {
	if err := svn.Checkout("", localSVNPath); err != nil {
		log.Fatal(err)
	}
	if err := localSVN.Mkdir(DefaultBranchesDir); err != nil {
		log.Fatal(err)
	}
	if err := localSVN.Mkdir(DefaultTrunkDir); err != nil {
		log.Fatal(err)
	}
	if err := localSVN.Mkdir(DefaultTagsDir); err != nil {
		log.Fatal(err)
	}
}

func TestSVNAddCommit(t *testing.T) {
	fn := path.Join(workDir, localSVNPath, DefaultTrunkDir, "sample.txt")
	fp, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY, 0644)
	fp.WriteString("# SVN Test File\n")
	if err != nil {
		t.Fatal(err)
	}
	defer fp.Close()
	if err := localSVN.Add(fn, nil); err != nil {
		t.Fatal(err)
	}
	if err := localSVN.Commit(localSVNPath, "add sample.txt", nil); err != nil {
		t.Fatal(err)
	}
}

func TestSVNAddCommit2(t *testing.T) {
	fn := path.Join(workDir, localSVNPath, DefaultTrunkDir, "sample.txt")
	fp, err := os.OpenFile(fn, os.O_CREATE|os.O_WRONLY, 0644)
	fp.WriteString("# SVN Test File Line\n")
	if err != nil {
		t.Fatal(err)
	}
	defer fp.Close()
	if err := localSVN.Commit(localSVNPath, "update sample.txt", nil); err != nil {
		t.Fatal(err)
	}
}
func TestCleanUp(t *testing.T) {
	if err := localSVN.Cleanup(localSVNPath); err != nil {
		t.Error(err)
	}
}

func inArray(a []string, str string) bool {
	for _, s := range a {
		if str == s {
			return true
		}
	}
	return false
}

func TestNewBranch(t *testing.T) {
	bn := "develop"
	if err := svn.NewBranch(bn, "create branch develop"); err != nil {
		t.Fatal(err)
	}
	if branches, err := svn.Branches(); err != nil {
		t.Error(err)
	} else if !inArray(branches, bn) {
		t.Error("new branch fail, not exists")
	}
}
func TestNewTag(t *testing.T) {
	tn := "v0.1"
	if err := svn.NewTag(tn, "create tag v0.1"); err != nil {
		t.Fatal(err)
	}
	if tages, err := svn.Tags(); err != nil {
		t.Error(err)
	} else if !inArray(tages, tn) {
		t.Error("new tag fail, not exists")
	}
}

func TestLog(t *testing.T) {
	lr, err := svn.Log(path.Join(svn.trunkDir, "sample.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if len(lr.Logentrys) != 2 {
		t.Errorf("log error expect 2 logentry get %d", len(lr.Logentrys))
	}
}

func TestExport(t *testing.T) {
	err := svn.Export("trunk/sample.txt", "/tmp/sample.txt", "-r", "1")
	if err != nil {
		t.Fatal(err)
	}
	fp, err := os.Open("/tmp/sample.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer fp.Close()
	data, err := ioutil.ReadAll(fp)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "# SVN Test File\n" {
		t.Error("Export File Err")
	}
}
