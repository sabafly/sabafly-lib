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
	"os"
	"strings"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/pelletier/go-toml/v2"
	"github.com/sabafly/disgo/discord"
	"golang.org/x/text/language"
	"gopkg.in/yaml.v2"
)

var (
	defaultLang     language.Tag    = language.Japanese
	bundle          *i18n.Bundle    = i18n.NewBundle(defaultLang)
	localizer       *i18n.Localizer = i18n.NewLocalizer(bundle)
	IsWriteFallBack bool            = false
	lang_path       string
)

func SetDefaultLanguage(lang language.Tag) {
	defaultLang = lang
}

type Cfg struct {
	Fallback         string
	FallbackLanguage discord.Locale
	TemplateData     any
	PluralCount      int
}

type Option func(*Cfg)

func WithFallback(fallback string) Option {
	return func(c *Cfg) {
		c.Fallback = fallback
	}
}

func WithFallBackLanguage(lang discord.Locale) Option {
	return func(c *Cfg) {
		c.FallbackLanguage = lang
	}
}

func WithTemplate(data any) Option {
	return func(c *Cfg) {
		c.TemplateData = data
	}
}

func WithPluralCount(count int) Option {
	return func(c *Cfg) {
		c.PluralCount = count
	}
}

func LoadTranslations(dir_path string) (*i18n.Bundle, error) {
	bundle = i18n.NewBundle(defaultLang)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	fd, err := os.ReadDir(dir_path)
	if err != nil {
		return nil, err
	}
	for _, de := range fd {
		_, err := bundle.LoadMessageFile(dir_path + "/" + de.Name())
		if err != nil {
			return nil, err
		}
	}
	lang_path = dir_path
	localizer = i18n.NewLocalizer(bundle, locales...)
	return bundle, nil
}

var (
	locales = []string{
		"en-US",
		"en-GB",
		"bg",
		"zh-CN",
		"zh-TW",
		"hr",
		"cs",
		"da",
		"nl",
		"fi",
		"fr",
		"de",
		"el",
		"hi",
		"hu",
		"id",
		"it",
		"ja",
		"ko",
		"lt",
		"no",
		"pl",
		"pt-BR",
		"ro",
		"ru",
		"es-ES",
		"sv-SE",
		"th",
		"tr",
		"uk",
		"vi",
	}
)

func Message(locale discord.Locale, messageId string, opts ...Option) (res string) {
	opt := new(Cfg)
	opt.FallbackLanguage = discord.LocaleJapanese
	for _, o := range opts {
		o(opt)
	}
	res, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageId,
		TemplateData: opt.TemplateData,
		PluralCount:  opt.PluralCount,
	})
	if err != nil {
		res = messageId
		if opt.Fallback != "" {
			res = opt.Fallback
			if IsWriteFallBack {
				f, err := os.Create(lang_path + "/" + locale.Code() + ".yaml")
				if err != nil {
					return res
				}
				defer f.Close()
				_, _ = f.WriteString(fmt.Sprintf("%s: %s\n", messageId, opt.Fallback))
			}
		}
	}
	return
}

// Deprecated: Use [github.com/sabafly/sabafly-lib/translate.WithTemplate] Option
func Translate(locale discord.Locale, messageId string, templateData any, opt ...Option) (res string) {
	return Message(locale, messageId, WithTemplate(templateData))
}

// Deprecated: Use [github.com/sabafly/sabafly-lib/translate.WithPluralCount] Option
func Translates(locale discord.Locale, messageId string, templateData any, pluralCount int, opts ...Option) string {
	return Message(locale, messageId, WithTemplate(templateData), WithPluralCount(pluralCount))
}

func MessageMap(key string, replace bool, opts ...Option) map[discord.Locale]string {
	res := map[discord.Locale]string{
		discord.LocaleEnglishUS:    Message(discord.LocaleEnglishUS, key, opts...),
		discord.LocaleEnglishGB:    Message(discord.LocaleEnglishGB, key, opts...),
		discord.LocaleBulgarian:    Message(discord.LocaleBulgarian, key, opts...),
		discord.LocaleChineseCN:    Message(discord.LocaleChineseCN, key, opts...),
		discord.LocaleChineseTW:    Message(discord.LocaleChineseTW, key, opts...),
		discord.LocaleCroatian:     Message(discord.LocaleCroatian, key, opts...),
		discord.LocaleCzech:        Message(discord.LocaleCzech, key, opts...),
		discord.LocaleDanish:       Message(discord.LocaleDanish, key, opts...),
		discord.LocaleDutch:        Message(discord.LocaleDutch, key, opts...),
		discord.LocaleFinnish:      Message(discord.LocaleFinnish, key, opts...),
		discord.LocaleFrench:       Message(discord.LocaleFrench, key, opts...),
		discord.LocaleGerman:       Message(discord.LocaleGerman, key, opts...),
		discord.LocaleGreek:        Message(discord.LocaleGreek, key, opts...),
		discord.LocaleHindi:        Message(discord.LocaleHindi, key, opts...),
		discord.LocaleHungarian:    Message(discord.LocaleHungarian, key, opts...),
		discord.LocaleIndonesian:   Message(discord.LocaleIndonesian, key, opts...),
		discord.LocaleItalian:      Message(discord.LocaleItalian, key, opts...),
		discord.LocaleJapanese:     Message(discord.LocaleJapanese, key, opts...),
		discord.LocaleKorean:       Message(discord.LocaleKorean, key, opts...),
		discord.LocaleLithuanian:   Message(discord.LocaleLithuanian, key, opts...),
		discord.LocaleNorwegian:    Message(discord.LocaleNorwegian, key, opts...),
		discord.LocalePolish:       Message(discord.LocalePolish, key, opts...),
		discord.LocalePortugueseBR: Message(discord.LocalePortugueseBR, key, opts...),
		discord.LocaleRomanian:     Message(discord.LocaleRomanian, key, opts...),
		discord.LocaleRussian:      Message(discord.LocaleRussian, key, opts...),
		discord.LocaleSpanishES:    Message(discord.LocaleSpanishES, key, opts...),
		discord.LocaleSwedish:      Message(discord.LocaleSwedish, key, opts...),
		discord.LocaleThai:         Message(discord.LocaleThai, key, opts...),
		discord.LocaleTurkish:      Message(discord.LocaleTurkish, key, opts...),
		discord.LocaleUkrainian:    Message(discord.LocaleUkrainian, key, opts...),
		discord.LocaleVietnamese:   Message(discord.LocaleVietnamese, key, opts...),
	}
	if replace {
		for l, v := range res {
			res[l] = strings.ReplaceAll(v, " ", "-")
		}
	}
	return res
}
