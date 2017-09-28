package svn

import (
	"testing"
)

var sample = []struct {
	cmd     string
	stderr  string
	errCode []string
}{
	{
		"svn export --username dadmin --password dangerous svn://106.75.11.136:3690/testrepo/testdd test",
		`svn: E170013: Unable to connect to a repository at URL 'svn://106.75.11.136/testrepo/testdd'
svn: E170001: Authentication error from server: Username not found`,
		[]string{
			"E170013",
			"E170001",
		},
	},
}

func TestError(t *testing.T) {
	for _, es := range sample {
		err := NewError(es.cmd, "", es.stderr)
		t.Logf("%#v", err)
		for _, c := range es.errCode {
			if !err.HasErr(c) {
				t.Errorf("expect err has code %s", c)
			}
		}
	}
}
