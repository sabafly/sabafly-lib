/*
	Copyright (C) 2022-2023  sabafly

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/
package translate

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pelletier/go-toml/v2"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

var (
	defaultLang               = language.Japanese
	translations *i18n.Bundle = i18n.NewBundle(defaultLang)
)

func SetDefaultLanguage(lang language.Tag) {
	defaultLang = lang
}

var (
	Release        bool = false
	translate_path string
)

func LoadTranslations(dir_path string) (*i18n.Bundle, error) {
	bundle := i18n.NewBundle(defaultLang)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	fd, err := os.ReadDir(dir_path)
	if err != nil {
		panic(err)
	}
	translate_path = dir_path
	for _, de := range fd {
		_, err := bundle.LoadMessageFile(dir_path + "/" + de.Name())
		if err != nil {
			panic(err)
		}
	}
	translations = bundle
	return bundle, nil
}

func Message(locale discord.Locale, messageId string) (res string) {
	res = MessageWithFallBack(locale, messageId, "")
	return
}

func MessageWithFallBack(locale discord.Locale, messageId, fallback string) (res string) {
	res = TranslateWithFallBack(locale, messageId, map[string]any{}, fallback)
	return
}

func Translate(locale discord.Locale, messageId string, templateData any) (res string) {
	res = TranslateWithFallBack(locale, messageId, templateData, "")
	return
}

func TranslateWithFallBack(locale discord.Locale, messageId string, templateData any, fallback string) (res string) {
	res = Translates(locale, messageId, templateData, 2, fallback)
	return
}

func Translates(locale discord.Locale, messageId string, templateData any, pluralCount int, fallback string) string {
	messageId = strings.ReplaceAll(messageId, ".", "_")
	Localizer := i18n.NewLocalizer(translations, string(locale))
	res, err := Localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageId,
		TemplateData: templateData,
		PluralCount:  pluralCount,
	})
	if err != nil {
		Localizer = i18n.NewLocalizer(translations, "ja")
		res, err = Localizer.Localize(&i18n.LocalizeConfig{
			MessageID:    messageId,
			TemplateData: templateData,
			PluralCount:  pluralCount,
		})
		if err != nil {
			res = messageId
			if fallback != "" {
				res = fallback
				if !Release && translate_path != "" {
					file, err := os.OpenFile(translate_path+"/"+defaultLang.String()+".yaml", os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModeTemporary)
					if err != nil {
						return res
					}
					defer file.Close()
					_, _ = file.WriteString(fmt.Sprintf("%s: \"%s\"\n", messageId, fallback))
					buf, err := io.ReadAll(file)
					if err != nil {
						return res
					}
					_, _ = translations.ParseMessageFileBytes(buf, translate_path+"/"+defaultLang.String()+".yaml")
				}
			}
		}
	}
	return res
}

func MessageMap(key string, replace bool) *map[discord.Locale]string {
	res := map[discord.Locale]string{
		discord.LocaleEnglishUS:    Message(discord.LocaleEnglishUS, key),
		discord.LocaleEnglishGB:    Message(discord.LocaleEnglishGB, key),
		discord.LocaleBulgarian:    Message(discord.LocaleBulgarian, key),
		discord.LocaleChineseCN:    Message(discord.LocaleChineseCN, key),
		discord.LocaleChineseTW:    Message(discord.LocaleChineseTW, key),
		discord.LocaleCroatian:     Message(discord.LocaleCroatian, key),
		discord.LocaleCzech:        Message(discord.LocaleCzech, key),
		discord.LocaleDanish:       Message(discord.LocaleDanish, key),
		discord.LocaleDutch:        Message(discord.LocaleDutch, key),
		discord.LocaleFinnish:      Message(discord.LocaleFinnish, key),
		discord.LocaleFrench:       Message(discord.LocaleFrench, key),
		discord.LocaleGerman:       Message(discord.LocaleGerman, key),
		discord.LocaleGreek:        Message(discord.LocaleGreek, key),
		discord.LocaleHindi:        Message(discord.LocaleHindi, key),
		discord.LocaleHungarian:    Message(discord.LocaleHungarian, key),
		discord.LocaleIndonesian:   Message(discord.LocaleIndonesian, key),
		discord.LocaleItalian:      Message(discord.LocaleItalian, key),
		discord.LocaleJapanese:     Message(discord.LocaleJapanese, key),
		discord.LocaleKorean:       Message(discord.LocaleKorean, key),
		discord.LocaleLithuanian:   Message(discord.LocaleLithuanian, key),
		discord.LocaleNorwegian:    Message(discord.LocaleNorwegian, key),
		discord.LocalePolish:       Message(discord.LocalePolish, key),
		discord.LocalePortugueseBR: Message(discord.LocalePortugueseBR, key),
		discord.LocaleRomanian:     Message(discord.LocaleRomanian, key),
		discord.LocaleRussian:      Message(discord.LocaleRussian, key),
		discord.LocaleSpanishES:    Message(discord.LocaleSpanishES, key),
		discord.LocaleSwedish:      Message(discord.LocaleSwedish, key),
		discord.LocaleThai:         Message(discord.LocaleThai, key),
		discord.LocaleTurkish:      Message(discord.LocaleTurkish, key),
		discord.LocaleUkrainian:    Message(discord.LocaleUkrainian, key),
		discord.LocaleVietnamese:   Message(discord.LocaleVietnamese, key),
		discord.LocaleUnknown:      Message(discord.LocaleUnknown, key),
	}
	if replace {
		for l, v := range res {
			res[l] = strings.ReplaceAll(v, " ", "-")
		}
	}
	return &res
}

func ErrorEmbed(locale discord.Locale, key string, any ...any) (embed []*discord.Embed) {
	var trs string
	if len(any) != 0 {
		trs = Translate(locale, key, any[0])
	} else if key != "" {
		trs = Message(locale, key)
	}
	embed = append(embed, &discord.Embed{
		Title:       Message(locale, "error_message"),
		Description: trs,
		Color:       0xff0000,
	})
	return
}
