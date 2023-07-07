package permissions_test

import (
	"testing"

	"github.com/sabafly/sabafly-lib/v2/permissions"
)

func TestCheck(t *testing.T) {
	p := permissions.New().
		Add("test.test").
		Add("test.none")
	if !p.Check("test.test").And("test2.test").Or("test.none").Or("test3.test").Res() {
		t.Fail()
	}
}
