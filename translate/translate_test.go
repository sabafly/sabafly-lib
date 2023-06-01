package translate_test

import (
	"testing"

	"github.com/disgoorg/disgo/discord"
	"github.com/sabafly/sabafly-lib/v2/translate"
)

func TestTranslate(t *testing.T) {
	if _, err := translate.LoadTranslations("./lang"); err != nil {
		t.Fatal(err)
	}
	if translate.Message(discord.Locale("ja"), "test") != "テスト" {
		t.Fail()
	}
	if translate.Message(discord.Locale("ja"), "not_exist") != "not_exist" {
		t.Fail()
	}
}
