package permissions_test

import (
	"fmt"
	"testing"

	"github.com/sabafly/sabafly-lib/v2/permissions"
)

func TestPerm(t *testing.T) {
	p := permissions.New().
		Add("test.perm").
		Add("test.perm.test").
		Add("test.yes.no").
		Add("test.yes.perm").
		Add("test2.yes").
		Add("test3.*")
	fmt.Println(p.List())
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
	if !p.Has("test3.test.yes") {
		t.Errorf("no test3.test.yes %v", p)
	}
	fmt.Println(p.List())
}
