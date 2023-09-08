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
package handler

import (
	"github.com/disgoorg/log"
	"github.com/disgoorg/snowflake/v2"
	"github.com/google/uuid"
	"github.com/sabafly/sabafly-disgo/bot"
	"github.com/sabafly/sabafly-disgo/discord"
	"github.com/sabafly/sabafly-disgo/events"
)

var _ bot.EventListener = (*Handler)(nil)

func New(logger log.Logger) *Handler {
	return &Handler{
		Logger:     logger,
		Commands:   map[string]Command{},
		Components: map[string]Component{},
		Modals:     map[string]Modal{},
		Message:    map[uuid.UUID]Message{},
		Ready:      []func(*events.Ready){},
		MemberJoin: genericsList[events.GuildMemberJoin]{
			Map:   map[uuid.UUID]Generics[events.GuildMemberJoin]{},
			Array: []Generics[events.GuildMemberJoin]{},
		},
		MemberLeave: genericsList[events.GuildMemberLeave]{
			Map:   map[uuid.UUID]Generics[events.GuildMemberLeave]{},
			Array: []Generics[events.GuildMemberLeave]{},
		},
		MessageReactionAdd: genericsList[events.GuildMessageReactionAdd]{
			Map:   map[uuid.UUID]Generics[events.GuildMessageReactionAdd]{},
			Array: []Generics[events.GuildMessageReactionAdd]{},
		},
		MessageReactionRemove: genericsList[events.GuildMessageReactionRemove]{
			Map:   map[uuid.UUID]Generics[events.GuildMessageReactionRemove]{},
			Array: []Generics[events.GuildMessageReactionRemove]{},
		},

		ExcludeID: map[snowflake.ID]struct{}{},
	}
}

type Handler struct {
	Logger log.Logger

	Commands                   map[string]Command
	Components                 map[string]Component
	Modals                     map[string]Modal
	Message                    map[uuid.UUID]Message
	MessageUpdate              map[uuid.UUID]MessageUpdate
	MessageDelete              map[uuid.UUID]MessageDelete
	Ready                      []func(*events.Ready)
	MemberJoin                 genericsList[events.GuildMemberJoin]
	MemberLeave                genericsList[events.GuildMemberLeave]
	MemberUpdate               genericsList[events.GuildMemberUpdate]
	MessageReactionAdd         genericsList[events.GuildMessageReactionAdd]
	MessageReactionRemove      genericsList[events.GuildMessageReactionRemove]
	MessageReactionRemoveAll   genericsList[events.GuildMessageReactionRemoveAll]
	MessageReactionRemoveEmoji genericsList[events.GuildMessageReactionRemoveEmoji]
	Event                      []Event

	Static StaticHandler

	ExcludeID  map[snowflake.ID]struct{}
	DevGuildID []snowflake.ID
	IsDebug    bool
	ASync      bool
	IsLogEvent bool
}

type StaticHandler struct {
	Message       []Message
	MessageUpdate []MessageUpdate
	MessageDelete []MessageDelete
}

func (h *Handler) AddExclude(ids ...snowflake.ID) {
	for _, id := range ids {
		h.ExcludeID[id] = struct{}{}
	}
}

func (h *Handler) AddCommands(commands ...Command) {
	for _, command := range commands {
		h.Commands[command.Create.CommandName()] = command
	}
}

func (h *Handler) AddComponents(components ...Component) {
	for _, component := range components {
		h.Components[component.Name] = component
	}
}

func (h *Handler) AddComponent(component Component) func() {
	h.Components[component.Name] = component
	return func() {
		delete(h.Components, component.Name)
	}
}

func (h *Handler) AddModals(modals ...Modal) {
	for _, modal := range modals {
		h.Modals[modal.Name] = modal
	}
}

func (h *Handler) AddMessage(message Message) func() {
	if message.UUID != nil {
		h.Message[*message.UUID] = message
		return func() {
			delete(h.Message, *message.UUID)
		}
	} else {
		h.Static.Message = append(h.Static.Message, message)
		return nil
	}
}

func (h *Handler) AddMessages(messages ...Message) {
	h.Static.Message = append(h.Static.Message, messages...)
}

func (h *Handler) AddMessageUpdate(messageUpdate MessageUpdate) func() {
	if messageUpdate.UUID != nil {
		h.MessageUpdate[*messageUpdate.UUID] = messageUpdate
		return func() {
			delete(h.MessageUpdate, *messageUpdate.UUID)
		}
	} else {
		h.Static.MessageUpdate = append(h.Static.MessageUpdate, messageUpdate)
		return nil
	}
}

func (h *Handler) AddMessageUpdates(messageUpdates ...MessageUpdate) {
	h.Static.MessageUpdate = append(h.Static.MessageUpdate, messageUpdates...)
}

func (h *Handler) AddMessageDelete(messageDelete MessageDelete) func() {
	if messageDelete.UUID != nil {
		h.MessageDelete[*messageDelete.UUID] = messageDelete
		return func() {
			delete(h.MessageDelete, *messageDelete.UUID)
		}
	} else {
		h.Static.MessageDelete = append(h.Static.MessageDelete, messageDelete)
		return nil
	}
}

func (h *Handler) AddMessageDeletes(messageDeletes ...MessageDelete) {
	h.Static.MessageDelete = append(h.Static.MessageDelete, messageDeletes...)
}

func (h *Handler) AddEvent(events ...Event) {
	h.Event = append(h.Event, events...)
}

func (h *Handler) AddReady(ready func(*events.Ready)) {
	h.Ready = append(h.Ready, ready)
}

func (h *Handler) handleReady(e *events.Ready) {
	for _, v := range h.Ready {
		v(e)
	}
}

func (h *Handler) SyncCommands(client bot.Client, guildIDs ...snowflake.ID) {
	commands := []discord.ApplicationCommandCreate{}
	devCommands := []discord.ApplicationCommandCreate{}
	for _, command := range h.Commands {
		if command.DevOnly {
			devCommands = append(devCommands, command.Create)
		} else {
			commands = append(commands, command.Create)
		}
	}

	if len(devCommands) > 0 {
		for _, id := range h.DevGuildID {
			if _, err := client.Rest().SetGuildCommands(client.ApplicationID(), id, devCommands); err != nil {
				h.Logger.Errorf("Failed to sync %d commands: %s", id, err)
			}
			h.Logger.Infof("Synced %d guild %d commands", len(devCommands), id)
			cmd, err := client.Rest().GetGuildCommands(client.ApplicationID(), id, true)
			h.Logger.Debugf("%+v %s", *cmd[0].GuildID(), err)
		}
	}

	if len(guildIDs) == 0 {
		if _, err := client.Rest().SetGlobalCommands(client.ApplicationID(), commands); err != nil {
			h.Logger.Error("Failed to sync global commands: ", err)
			return
		}
		h.Logger.Infof("Synced %d global commands", len(commands))
		return
	}

	for _, guildID := range guildIDs {
		if _, err := client.Rest().SetGuildCommands(client.ApplicationID(), guildID, commands); err != nil {
			h.Logger.Errorf("Failed to sync commands for guild %d: %s", guildID, err)
			continue
		}
		h.Logger.Infof("Synced %d commands for guild %s", len(commands), guildID)
	}
}

func (h *Handler) OnEvent(event bot.Event) {
	if h.ASync {
		go func() {
			if !h.IsDebug {
				defer func() {
					if err := recover(); err != nil {
						h.Logger.Errorf("panic: %s", err)
					}
				}()
			}
			h.onEvent(event)
		}()
	} else {
		h.onEvent(event)
	}
}

func (h *Handler) onEvent(event bot.Event) {
	switch e := event.(type) {
	case *events.ApplicationCommandInteractionCreate:
		h.handleCommand(e)
	case *events.AutocompleteInteractionCreate:
		h.handleAutocomplete(e)
	case *events.ComponentInteractionCreate:
		h.handleComponent(e)
	case *events.ModalSubmitInteractionCreate:
		h.handleModal(e)
	case *events.GuildMessageCreate:
		h.handleMessage(e)
	case *events.GuildMessageDelete:
		h.handleMessageDelete(e)
	case *events.GuildMessageUpdate:
		h.handleMessageUpdate(e)
	case *events.Ready:
		h.handleReady(e)
	case *events.GuildMemberJoin:
		h.MemberJoin.handleEvent(e)
	case *events.GuildMemberLeave:
		h.MemberLeave.handleEvent(e)
	case *events.GuildMessageReactionAdd:
		h.MessageReactionAdd.handleEvent(e)
	case *events.GuildMessageReactionRemove:
		h.MessageReactionRemove.handleEvent(e)
	case *events.GuildMessageReactionRemoveAll:
		h.MessageReactionRemoveAll.handleEvent(e)
	case *events.GuildMessageReactionRemoveEmoji:
		h.MessageReactionRemoveEmoji.handleEvent(e)
	}
	h.handleEvent(event)
}
