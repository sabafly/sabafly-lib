package token_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/disgoorg/snowflake/v2"
	"github.com/sabafly/sabafly-lib/token"
)

func TestTokenGen(t *testing.T) {
	uid := snowflake.New(time.Now())
	fmt.Println(uid)
	tkn := token.NewToken(uid)
	fmt.Println(tkn)
	buf, err := tkn.MarshalJSON()
	if err != nil {
		t.Error(err)
		return
	}
	tkn2 := token.Token{}
	if err := tkn2.UnmarshalJSON(buf); err != nil {
		t.Error(err)
		return
	}
	fmt.Println(tkn2)
	if tkn.String() != tkn2.String() {
		t.Fail()
		return
	}
}

type TestObj struct {
	Token token.Token `json:"token"`
}

func TestTokenJSON(t *testing.T) {
	uid := snowflake.New(time.Now())
	fmt.Println(uid)
	tkn1 := token.NewToken(uid)
	fmt.Println(tkn1)
	obj1 := TestObj{
		Token: tkn1,
	}
	buf, err := json.Marshal(obj1)
	if err != nil {
		t.Error(err)
		return
	}
	obj2 := TestObj{}
	if err := json.Unmarshal(buf, &obj2); err != nil {
		t.Error(err)
		return
	}
	fmt.Println(obj2.Token)
	if obj1.Token.String() != obj2.Token.String() {
		t.Fail()
		return
	}
}
