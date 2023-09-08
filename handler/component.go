package handler

import (
	"strings"

	"github.com/sabafly/sabafly-disgo/events"
)

type ComponentHandler func(event *events.ComponentInteractionCreate) error

type Component struct {
	Name      string
	Check     Check[*events.ComponentInteractionCreate]
	Checks    map[string]Check[*events.ComponentInteractionCreate]
	Handler   map[string]ComponentHandler
	Ephemeral map[string]bool
}

func (h *Handler) handleComponent(event *events.ComponentInteractionCreate) {
	if h.IsLogEvent {
		h.Logger.Infof("%s(%s) used %s component", event.Member().User.Tag(), event.Member().User.ID, event.Data.CustomID())
	}
	customID := event.Data.CustomID()
	h.Logger.Debugf("コンポーネントインタラクション呼び出し %s", customID)
	if !strings.HasPrefix(customID, "handler:") {
		return
	}

	var subName string
	if strings.Count(customID, ":") >= 2 {
		subName = strings.Split(customID, ":")[2]
	}

	componentName := strings.Split(customID, ":")[1]
	component, ok := h.Components[componentName]
	if !ok || component.Handler == nil {
		h.Logger.Errorf("No component handler for \"%s\" found", componentName)
	}

	if component.Check != nil && !component.Check(event) {
		return
	}

	if check, ok := component.Checks[subName]; ok && !check(event) {
		return
	}

	handler, ok := component.Handler[subName]
	if !ok {
		h.Logger.Debugf("不明なハンダラ %s", subName)
		err := event.DeferUpdateMessage()
		if err != nil {
			h.Logger.Errorf("Failed to handle unknown handler interaction for \"%s\" : %s", customID, err)
		}
		return
	}

	defer deferUpdateInteraction(event, component.Ephemeral != nil && component.Ephemeral[subName])
	if err := handler(event); err != nil {
		h.Logger.Errorf("Failed to handle component interaction for \"%s\" : %s", componentName, err)
	}
}
