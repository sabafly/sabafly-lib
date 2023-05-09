package permissions_test

import (
	"testing"

	"github.com/sabafly/sabafly-lib/permissions"
)

func TestPerm(t *testing.T) {
	p := permissions.New()
	p.Add("test.perm")
	p.Add("test.yes.perm")
	p.Add("test2.yes")
	if !p.Has("test.perm") {
		t.Errorf("no test.perm %v", p)
	}
	p.Del("test.perm")
	if p.Has("test.perm") {
		t.Errorf("yes test.perm %s", p)
	}
	if !p.Has("test.yes.perm") {
		t.Errorf("no test.yes.perm %v", p)
	}
	if !p.Has("test2.yes") {
		t.Errorf("no test2.yes %v", p)
	}
	if !p.Has("test.*") {
		t.Errorf("no test.* %v", p)
	}
	if p.Has("test.no.perm") {
		t.Errorf("yes test.no.perm %v", p)
	}
}