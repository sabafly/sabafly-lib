package botlib

import (
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/sabafly/sabafly-lib/v2/emoji"
	"github.com/sabafly/sabafly-lib/v2/translate"

	"github.com/disgoorg/snowflake/v2"
	"github.com/sabafly/disgo/bot"
	"github.com/sabafly/disgo/discord"
	"github.com/sabafly/disgo/rest"
)

// 埋め込みの色、フッター、タイムスタンプを設定する
func SetEmbedProperties(embed discord.Embed) discord.Embed {
	now := time.Now()
	if embed.Color == 0 {
		embed.Color = Color
	}
	if embed.Footer == nil {
		embed.Footer = &discord.EmbedFooter{}
	}
	if embed.Footer.Text == "" {
		embed.Footer.Text = BotName
	}
	if embed.Timestamp == nil {
		embed.Timestamp = &now
	}
	return embed
}

// 埋め込みの色、フッター、タイムスタンプを設定する
func SetEmbedsProperties(embeds []discord.Embed) []discord.Embed {
	now := time.Now()
	for i := range embeds {
		if embeds[i].Color == 0 {
			embeds[i].Color = Color
		}
		if i == len(embeds)-1 {
			if embeds[i].Footer == nil {
				embeds[i].Footer = &discord.EmbedFooter{}
			}
			if embeds[i].Footer.Text == "" {
				embeds[i].Footer.Text = BotName
			}
			if embeds[i].Timestamp == nil {
				embeds[i].Timestamp = &now
			}
		}
	}
	return embeds
}

type responsibleInteraction interface {
	Locale() discord.Locale
	CreateMessage(discord.MessageCreate, ...rest.RequestOpt) error
}

func ReturnErr(interaction responsibleInteraction, err error, opts ...ReturnErrOption) error {
	cfg := new(ReturnErrCfg)
	for _, reo := range opts {
		reo(cfg)
	}
	embeds := ErrorTraceEmbed(interaction.Locale(), err)
	embeds = SetEmbedsProperties(embeds)
	if err2 := interaction.CreateMessage(discord.MessageCreate{
		Embeds: embeds,
		Flags: func() discord.MessageFlags {
			if cfg.Ephemeral {
				return discord.MessageFlagEphemeral
			} else {
				return 0
			}
		}(),
	}); err2 != nil {
		return fmt.Errorf("%w: %w", err, err2)
	}
	return err
}

type ReturnErrCfg struct {
	Ephemeral           bool   `json:"ephemeral"`
	TranslateData       []any  `json:"translate_data"`
	FallBackTitle       string `json:"fallback_title"`
	FallBackDescription string `json:"fallback_description"`
}

type ReturnErrOption func(*ReturnErrCfg)

func WithEphemeral(enabled bool) ReturnErrOption {
	return func(rec *ReturnErrCfg) {
		rec.Ephemeral = enabled
	}
}

func WithTranslateData(data ...any) ReturnErrOption {
	return func(rec *ReturnErrCfg) {
		rec.TranslateData = data
	}
}

func WithFallbackTitle(title string) ReturnErrOption {
	return func(rec *ReturnErrCfg) {
		rec.FallBackTitle = title
	}
}

func WithFallBackDescription(desc string) ReturnErrOption {
	return func(rec *ReturnErrCfg) {
		rec.FallBackDescription = desc
	}
}

func ReturnErrMessage(interaction responsibleInteraction, tr string, opts ...ReturnErrOption) error {
	embeds := ErrorMessageEmbed(interaction.Locale(), tr, opts...)
	embeds = SetEmbedsProperties(embeds)
	var flags discord.MessageFlags
	cfg := new(ReturnErrCfg)
	for _, reo := range opts {
		reo(cfg)
	}
	if cfg.Ephemeral {
		flags = discord.MessageFlagEphemeral
	}
	if err := interaction.CreateMessage(discord.MessageCreate{
		Embeds: embeds,
		Flags:  flags,
	}); err != nil {
		return err
	}
	return nil
}

// エラーメッセージ埋め込みを作成する
func ErrorMessageEmbed(locale discord.Locale, t string, opts ...ReturnErrOption) []discord.Embed {
	cfg := new(ReturnErrCfg)
	for _, reo := range opts {
		reo(cfg)
	}
	var td any
	if len(cfg.TranslateData) != 0 {
		td = cfg.TranslateData[0]
	}
	embeds := []discord.Embed{
		{
			Title:       translate.Message(locale, t+"_title", translate.WithFallback(cfg.FallBackTitle)),
			Description: translate.Translate(locale, t+"_message", td, translate.WithFallback(cfg.FallBackDescription)),
			Color:       0xff0000,
		},
	}
	embeds = SetEmbedsProperties(embeds)
	return embeds
}

// エラートレース埋め込みを作成する
func ErrorTraceEmbed(locale discord.Locale, err error) []discord.Embed {
	stack := debug.Stack()
	embeds := []discord.Embed{
		{
			Title:       "💥" + translate.Message(locale, "error_occurred_embed_message", translate.WithFallback("エラーが発生しました")),
			Description: fmt.Sprintf("%s\r```%s```", err, string(stack)),
			Color:       0xff0000,
		},
	}
	embeds = SetEmbedsProperties(embeds)
	return embeds
}

// 渡されたステータスの絵文字を返す
func StatusString(status discord.OnlineStatus) (str string) {
	switch status {
	case discord.OnlineStatusOnline:
		return "<:online:1055430359363354644>"
	case discord.OnlineStatusDND:
		return "<:dnd:1055434290629980220>"
	case discord.OnlineStatusIdle:
		return "<:idle:1055433789020586035> "
	case discord.OnlineStatusInvisible:
		return "<:offline:1055434315514785792>"
	case discord.OnlineStatusOffline:
		return "<:offline:1055434315514785792>"
	}
	return ""
}

// アクティビティ名をアクティビティの種類によって渡された言語に翻訳して返す
func ActivitiesNameString(locale discord.Locale, activity discord.Activity) (str string) {
	switch activity.Type {
	case discord.ActivityTypeGame:
		str = translate.Translate(locale, "activity_game_name", map[string]any{"Name": activity.Name})
	case discord.ActivityTypeStreaming:
		str = translate.Translate(locale, "activity_streaming_name", map[string]any{"Details": activity.Details, "URL": activity.URL})
	case discord.ActivityTypeListening:
		str = translate.Translate(locale, "activity_listening_name", map[string]any{"Name": activity.Name})
	case discord.ActivityTypeWatching:
		str = translate.Translate(locale, "activity_watching_name", map[string]any{"Name": activity.Name})
	case discord.ActivityTypeCustom:
		if activity.Emoji != nil {
			return
		}
		str = activity.Name
		if activity.Emoji.ID != nil && activity.Emoji.Name != nil {
			str = discord.EmojiMention(*activity.Emoji.ID, *activity.Emoji.Name) + " " + activity.Name
		}
	case discord.ActivityTypeCompeting:
		str = translate.Translate(locale, "activity_competing_name", map[string]any{"Name": activity.Name})
	}
	return str
}

func SendWebhook(client bot.Client, channelID snowflake.ID, data discord.WebhookMessageCreate) (st *discord.Message, err error) {
	webhooks, err := client.Rest().GetWebhooks(channelID)
	if err != nil {
		return nil, err
	}
	me, ok := client.Caches().SelfUser()
	if !ok {
		return nil, err
	}
	var token string
	var webhook discord.Webhook = nil
	for _, w := range webhooks {
		switch v := w.(type) {
		case discord.IncomingWebhook:
			if v.User.ID == me.User.ID {
				token = v.Token
				webhook = v
				if data.Username == "" {
					data.Username = me.Username
				}
				if data.AvatarURL == "" {
					data.AvatarURL = me.EffectiveAvatarURL(discord.WithFormat(discord.FileFormatPNG))
				}
				st, err = client.Rest().CreateWebhookMessage(webhook.ID(), token, data, true, snowflake.ID(0))
				if err != nil {
					return nil, err
				}
				return st, nil
			}
		}
	}
	if webhook == nil {
		var buf []byte
		if avatarURL := me.EffectiveAvatarURL(discord.WithFormat(discord.FileFormatPNG)); avatarURL != "" {
			resp, err := http.Get(avatarURL)
			if err != nil {
				return nil, fmt.Errorf("error on get: %w", err)
			}
			buf, err = io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("error on read all: %w", err)
			}
		}
		data, err := client.Rest().CreateWebhook(channelID, discord.WebhookCreate{
			Name:   BotName + "-webhook",
			Avatar: discord.NewIconRaw(discord.IconTypePNG, buf),
		})
		if err != nil {
			return nil, fmt.Errorf("error on create webhook: %w", err)
		}
		token = data.Token
		webhook = data
	}
	if data.Username == "" {
		data.Username = me.Username
	}
	if data.AvatarURL == "" {
		data.AvatarURL = me.EffectiveAvatarURL(discord.WithFormat(discord.FileFormatPNG))
	}
	st, err = client.Rest().CreateWebhookMessage(webhook.ID(), token, data, true, snowflake.ID(0))
	if err != nil {
		return nil, err
	}
	return st, nil
}

func GetWebhook(client bot.Client, channelID snowflake.ID) (id snowflake.ID, token string, err error) {
	webhooks, err := client.Rest().GetWebhooks(channelID)
	if err != nil {
		return 0, "", err
	}
	me, ok := client.Caches().SelfUser()
	if !ok {
		return 0, "", err
	}
	var webhook discord.Webhook = nil
	for _, w := range webhooks {
		switch v := w.(type) {
		case discord.IncomingWebhook:
			if v.User.ID == me.User.ID {
				token = v.Token
				webhook = v
				return webhook.ID(), token, nil
			}
		}
	}
	if webhook == nil {
		var buf []byte
		if avatarURL := me.EffectiveAvatarURL(discord.WithFormat(discord.FileFormatPNG)); avatarURL != "" {
			resp, err := http.Get(avatarURL)
			if err != nil {
				return 0, "", fmt.Errorf("error on get: %w", err)
			}
			buf, err = io.ReadAll(resp.Body)
			if err != nil {
				return 0, "", fmt.Errorf("error on read all: %w", err)
			}
		}
		data, err := client.Rest().CreateWebhook(channelID, discord.WebhookCreate{
			Name:   BotName + "-webhook",
			Avatar: discord.NewIconRaw(discord.IconTypePNG, buf),
		})
		if err != nil {
			return 0, "", fmt.Errorf("error on create webhook: %w", err)
		}
		token = data.Token
		webhook = data
	}
	return webhook.ID(), token, nil
}

func GetCustomEmojis(str string) []discord.Emoji {
	var toReturn []discord.Emoji
	emojis := emoji.DiscordEmoji.FindAllString(str, -1)
	if len(emojis) < 1 {
		return toReturn
	}
	for _, em := range emojis {
		parts := strings.Split(em, ":")
		toReturn = append(toReturn, discord.Emoji{
			ID:       snowflake.MustParse(parts[2][:len(parts[2])-1]),
			Name:     parts[1],
			Animated: strings.HasPrefix(em, "<a:"),
		})
	}
	return toReturn
}

func ParseComponentEmoji(str string) discord.ComponentEmoji {
	e := discord.ComponentEmoji{
		Name: str,
	}
	if !emoji.MatchString(str) {
		return e
	}
	emojis := GetCustomEmojis(str)
	if len(emojis) < 1 {
		return e
	}
	e = discord.ComponentEmoji{
		ID:       emojis[0].ID,
		Name:     emojis[0].Name,
		Animated: emojis[0].Animated,
	}
	return e
}

func Number2Emoji(n int) string {
	return string(rune('🇦' - 1 + n))
}

func FormatComponentEmoji(e discord.ComponentEmoji) string {
	var zeroID snowflake.ID
	if e.ID == zeroID {
		return e.Name
	}
	if e.Animated {
		return fmt.Sprintf("<a:%s:%d>", e.Name, e.ID)
	} else {
		return fmt.Sprintf("<:%s:%d>", e.Name, e.ID)
	}
}

func ReactionComponentEmoji(e discord.ComponentEmoji) string {
	var zeroID snowflake.ID
	if e.ID == zeroID {
		return e.Name
	}
	return fmt.Sprintf("%s:%d", e.Name, e.ID)
}

func GetHighestRolePosition(role map[snowflake.ID]discord.Role) (int, snowflake.ID) {
	var max int
	var id snowflake.ID
	for i, r := range role {
		if max < r.Position {
			max = r.Position
			id = i
		}
	}
	return max, id
}
