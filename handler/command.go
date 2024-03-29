package handler

import (
	"time"

	"github.com/sabafly/sabafly-disgo/discord"
	"github.com/sabafly/sabafly-disgo/events"
	"github.com/sabafly/sabafly-disgo/rest"
)

type (
	CommandHandler      func(event *events.ApplicationCommandInteractionCreate) error
	AutocompleteHandler func(event *events.AutocompleteInteractionCreate) error
)

type Command struct {
	Create               discord.ApplicationCommandCreate
	Check                Check[*events.ApplicationCommandInteractionCreate]
	Checks               map[string]Check[*events.ApplicationCommandInteractionCreate]
	AutocompleteCheck    Check[*events.AutocompleteInteractionCreate]
	AutocompleteChecks   map[string]Check[*events.AutocompleteInteractionCreate]
	CommandHandlers      map[string]CommandHandler
	AutocompleteHandlers map[string]AutocompleteHandler
	Ephemeral            map[string]bool

	DevOnly bool
}

func (h *Handler) handleCommand(event *events.ApplicationCommandInteractionCreate) {
	if h.IsLogEvent {
		switch d := event.Data.(type) {
		case discord.SlashCommandInteractionData:
			h.Logger.Infof("%s(%s) used %s command type %d", event.User().Tag(), event.User().ID, d.CommandPath(), d.Type())
		case discord.UserCommandInteractionData:
			h.Logger.Infof("%s(%s) used %s command target %s type %d",
				event.User().Tag(), event.User().ID, d.CommandName(), d.TargetID(), d.Type(),
			)
		case discord.MessageCommandInteractionData:
			h.Logger.Infof("%s(%s) used %s command target %s type %d",
				event.User().Tag(), event.User().ID, d.CommandName(), d.TargetMessage().JumpURL(), d.Type(),
			)
		default:
			h.Logger.Infof("%s(%s) used %s command type %d", event.User().Tag(), event.User().ID, d.CommandName(), d.Type())
		}
	}
	name := event.Data.CommandName()
	h.Logger.Debugf("command created %s", name)
	cmd, ok := h.Commands[name]
	if !ok || cmd.CommandHandlers == nil {
		h.Logger.Errorf("No command or handler found for \"%s\"", name)
	}

	if cmd.Check != nil && !cmd.Check(event) {
		return
	}

	var path string
	if d, ok := event.Data.(discord.SlashCommandInteractionData); ok {
		path = buildCommandPath(d.SubCommandName, d.SubCommandGroupName)
	}

	if check, ok := cmd.Checks[path]; ok && !check(event) {
		return
	}

	handler, ok := cmd.CommandHandlers[path]
	if !ok {
		h.Logger.Warnf("No handler for command \"%s\" with path \"%s\" found", name, path)
		return
	}

	defer deferUpdateInteraction(event, cmd.Ephemeral != nil && cmd.Ephemeral[path])
	if err := handler(event); err != nil {
		h.Logger.Errorf("Failed to handle command \"%s\" with path \"%s\": %s", name, path, err)
	}
}

type deferCreateMessage interface {
	DeferCreateMessage(ephemeral bool, opts ...rest.RequestOpt) error
}

func deferUpdateInteraction(event deferCreateMessage, ephemeral bool) {
	time.Sleep(time.Millisecond * 2500)
	_ = event.DeferCreateMessage(ephemeral)
}

func (h *Handler) handleAutocomplete(event *events.AutocompleteInteractionCreate) {
	name := event.Data.CommandName
	cmd, ok := h.Commands[name]
	if !ok || cmd.AutocompleteHandlers == nil {
		h.Logger.Errorf("No autocomplete or handler found for \"%s\"", name)
	}

	if cmd.AutocompleteCheck != nil && !cmd.AutocompleteCheck(event) {
		return
	}

	path := buildCommandPath(event.Data.SubCommandName, event.Data.SubCommandGroupName)

	if check, ok := cmd.AutocompleteChecks[path]; ok && !check(event) {
		return
	}

	handler, ok := cmd.AutocompleteHandlers[path]
	if !ok {
		h.Logger.Warnf("No autocomplete handler for autocomplete \"%s\" with path \"%s\" found", name, path)
		return
	}

	if err := handler(event); err != nil {
		h.Logger.Errorf("Failed to handle autocomplete for autocomplete \"%s\" with path \"%s\": %s", name, path, err)
	}
}

func buildCommandPath(subcommand *string, subcommandGroup *string) string {
	var path string
	if subcommand != nil {
		path = *subcommand
	}
	if subcommandGroup != nil {
		path = *subcommandGroup + "/" + path
	}
	return path
}
