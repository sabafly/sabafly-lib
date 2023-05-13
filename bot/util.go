package botlib

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"github.com/sabafly/sabafly-lib/translate"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
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

func ReturnErr(interaction responsibleInteraction, err error) error {
	embeds := ErrorTraceEmbed(interaction.Locale(), err)
	embeds = SetEmbedsProperties(embeds)
	if err := interaction.CreateMessage(discord.MessageCreate{
		Embeds: embeds,
	}); err != nil {
		return err
	}
	return err
}

func ReturnErrMessage(interaction responsibleInteraction, tr, fallback_title, fallback_description string, data ...any) error {
	return ReturnErrMessageEphemeral(interaction, tr, fallback_title, fallback_description, false, data...)
}

func ReturnErrMessageEphemeral(interaction responsibleInteraction, tr, fallback_title, fallback_description string, ephemeral bool, data ...any) error {
	embeds := ErrorMessageEmbed(interaction.Locale(), tr, fallback_title, fallback_description, data...)
	embeds = SetEmbedsProperties(embeds)
	var flags discord.MessageFlags
	if ephemeral {
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
func ErrorMessageEmbed(locale discord.Locale, t, fallback_title, fallback_description string, data ...any) []discord.Embed {
	var td any
	if len(data) != 0 {
		td = data[0]
	}
	embeds := []discord.Embed{
		{
			Title:       translate.Message(locale, t+"_title", translate.WithFallback(fallback_title)),
			Description: translate.Translate(locale, t+"_message", td, translate.WithFallback(fallback_description)),
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

// エラーが発生したことを返すレスポンスを作成する
func ErrorRespond(locale discord.Locale, err error) discord.MessageCreate {
	return discord.MessageCreate{
		Embeds: ErrorTraceEmbed(locale, err),
	}
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
		str = discord.EmojiMention(*activity.Emoji.ID, activity.Emoji.Name) + " " + activity.Name
	case discord.ActivityTypeCompeting:
		str = translate.Translate(locale, "activity_competing_name", map[string]any{"Name": activity.Name})
	}
	return str
}

func MessageLogDetails(m []MessageLog) (day, week, all int, channelID snowflake.ID) {
	var inDay, inWeek []MessageLog
	channelCount := map[snowflake.ID]int{}
	for _, ml := range m {
		channelCount[ml.ChannelID]++
		timestamp := ml.ID.Time()
		if timestamp.After(time.Now().Add(-time.Hour * 24 * 7)) {
			inWeek = append(inWeek, ml)
		}
	}
	for _, ml := range inWeek {
		timestamp := ml.ID.Time()
		if timestamp.After(time.Now().Add(-time.Hour * 24)) {
			inDay = append(inDay, ml)
		}
	}
	count := []struct {
		ID    snowflake.ID
		Count int
	}{}
	for k, v := range channelCount {
		count = append(count, struct {
			ID    snowflake.ID
			Count int
		}{ID: k, Count: v})
	}
	sort.Slice(count, func(i, j int) bool {
		return count[i].Count > count[j].Count
	})
	if len(count) != 0 {
		channelID = count[0].ID
	}
	return len(inDay), len(inWeek), len(m), channelID
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
		if w.Type() != discord.WebhookTypeIncoming {
			continue
		}
		buf, err := w.MarshalJSON()
		if err != nil {
			continue
		}
		data := discord.IncomingWebhook{}
		err = json.Unmarshal(buf, &data)
		if err != nil {
			return nil, err
		}
		if data.User.ID == client.ID() {
			token = data.Token
			webhook = data
		}
	}
	if webhook == nil {
		var buf []byte
		if avatarURL := me.EffectiveAvatarURL(discord.WithFormat(discord.ImageFormatPNG)); avatarURL != "" {
			resp, err := http.Get(avatarURL)
			if err != nil {
				return nil, err
			}
			buf, err = io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}
		}
		data, err := client.Rest().CreateWebhook(channelID, discord.WebhookCreate{
			Name:   BotName + "-webhook",
			Avatar: discord.NewIconRaw(discord.IconTypePNG, buf),
		})
		if err != nil {
			return nil, err
		}
		token = data.Token
		webhook = data
	}
	if data.Username == "" {
		data.Username = me.Username
	}
	if data.AvatarURL == "" {
		data.AvatarURL = me.EffectiveAvatarURL(discord.WithFormat(discord.ImageFormatPNG))
	}
	st, err = client.Rest().CreateWebhookMessage(webhook.ID(), token, data, true, snowflake.ID(0))
	if err != nil {
		return nil, err
	}
	return st, nil
}

var EmojiRegex = regexp.MustCompile(`<(a|):[A-z0-9_~]+:[0-9]{18,20}>`)

func GetCustomEmojis(str string) []discord.Emoji {
	var toReturn []discord.Emoji
	emojis := EmojiRegex.FindAllString(str, -1)
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
	emoji := discord.ComponentEmoji{
		Name: str,
	}
	if !EmojiRegex.MatchString(str) {
		return emoji
	}
	emojis := GetCustomEmojis(str)
	if len(emojis) < 1 {
		return emoji
	}
	emoji = discord.ComponentEmoji{
		ID:       emojis[0].ID,
		Name:     emojis[0].Name,
		Animated: emojis[0].Animated,
	}
	return emoji
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
