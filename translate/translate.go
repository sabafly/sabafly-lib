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

type Cfg struct {
	Fallback         string
	FallbackLanguage discord.Locale
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

func LoadTranslations(dir_path string) (*i18n.Bundle, error) {
	bundle := i18n.NewBundle(defaultLang)
	bundle.RegisterUnmarshalFunc("yaml", yaml.Unmarshal)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	fd, err := os.ReadDir(dir_path)
	if err != nil {
		panic(err)
	}
	for _, de := range fd {
		_, err := bundle.LoadMessageFile(dir_path + "/" + de.Name())
		if err != nil {
			panic(err)
		}
	}
	translations = bundle
	return bundle, nil
}

func Message(locale discord.Locale, messageId string, opts ...Option) (res string) {
	res = Translate(locale, messageId, map[string]any{}, opts...)
	return
}

// Deprecated: Use Message() with WithFallback()
func MessageWithFallBack(locale discord.Locale, messageId, fallback string) (res string) {
	res = Translate(locale, messageId, map[string]any{}, WithFallback(fallback))
	return
}

func Translate(locale discord.Locale, messageId string, templateData any, opt ...Option) (res string) {
	res = Translates(locale, messageId, templateData, 2, opt...)
	return
}

// Deprecated: Use Translate() with WithFallback()
func TranslateWithFallBack(locale discord.Locale, messageId string, templateData any, fallback string) (res string) {
	res = Translates(locale, messageId, templateData, 2, WithFallback(fallback))
	return
}

func Translates(locale discord.Locale, messageId string, templateData any, pluralCount int, opts ...Option) string {
	localizer := i18n.NewLocalizer(translations, string(locale))
	res, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageId,
		TemplateData: templateData,
		PluralCount:  pluralCount,
	})
	opt := new(Cfg)
	opt.FallbackLanguage = discord.LocaleJapanese
	for _, o := range opts {
		o(opt)
	}
	if err != nil {
		localizer = i18n.NewLocalizer(translations, string(opt.FallbackLanguage))
		res, err = localizer.Localize(&i18n.LocalizeConfig{
			MessageID:    messageId,
			TemplateData: templateData,
			PluralCount:  pluralCount,
		})
		if err != nil {
			res = messageId
			if opt.Fallback != "" {
				res = opt.Fallback
			}
		}
	}
	return res
}

func MessageMap(key string, replace bool, opts ...Option) *map[discord.Locale]string {
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
		discord.LocaleUnknown:      Message(discord.LocaleUnknown, key, opts...),
	}
	if replace {
		for l, v := range res {
			res[l] = strings.ReplaceAll(v, " ", "-")
		}
	}
	return &res
}
