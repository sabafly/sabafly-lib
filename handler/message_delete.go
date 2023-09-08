package handler

import (
	"github.com/disgoorg/snowflake/v2"
	"github.com/google/uuid"
	"github.com/sabafly/sabafly-disgo/events"
)

type (
	MessageDeleteHandler func(event *events.GuildMessageDelete) error
)

type MessageDelete struct {
	UUID      *uuid.UUID
	ChannelID *snowflake.ID
	AuthorID  *snowflake.ID
	Check     Check[*events.GuildMessageDelete]
	Handler   MessageDeleteHandler
}

func (h *Handler) handleMessageDelete(event *events.GuildMessageDelete) {
	if _, ok := h.ExcludeID[event.ChannelID]; ok {
		return
	}
	h.Logger.Debugf("メッセージ作成 %d", event.ChannelID)
	for _, m := range h.Static.MessageDelete {
		h.run_message_delete(m, event)
	}
	for _, m := range h.MessageDelete {
		h.run_message_delete(m, event)
	}
}

func (h *Handler) run_message_delete(m MessageDelete, event *events.GuildMessageDelete) {
	if m.ChannelID != nil && *m.ChannelID != event.ChannelID {
		h.Logger.Debug("チャンネルが違います")
		return
	}
	if m.AuthorID != nil && *m.AuthorID != event.Message.Author.ID {
		h.Logger.Debugf("送信者が違います %d %d", *m.AuthorID, event.Message.Author.ID)
		return
	}
	if m.Check != nil && !m.Check(event) {
		return
	}
	if err := m.Handler(event); err != nil {
		h.Logger.Errorf("Failed to handle message \"%d\" in \"%s\", %s: %s", event.MessageID, event.GuildID, event.ChannelID, err.Error())
	}
}
