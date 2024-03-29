package handler

import (
	"strings"

	"github.com/sabafly/sabafly-disgo/events"
)

type ModalHandler func(event *events.ModalSubmitInteractionCreate) error

type Modal struct {
	Name      string
	Check     Check[*events.ModalSubmitInteractionCreate]
	Checks    map[string]Check[*events.ModalSubmitInteractionCreate]
	Handler   map[string]ModalHandler
	Ephemeral map[string]bool
}

func (h *Handler) handleModal(event *events.ModalSubmitInteractionCreate) {
	if h.IsLogEvent {
		h.Logger.Infof("%s(%s) used %d modal", event.Member().User.Tag(), event.Member().User.ID, event.Data.CustomID)
	}
	customID := event.Data.CustomID
	h.Logger.Debugf("モーダル提出インタラクション呼び出し %s", customID)
	if !strings.HasPrefix(customID, "handler:") {
		return
	}

	var subName string
	if strings.Count(customID, ":") >= 2 {
		subName = strings.Split(customID, ":")[2]
	}

	modalName := strings.Split(customID, ":")[1]
	modal, ok := h.Modals[modalName]
	if !ok || modal.Handler == nil {
		h.Logger.Errorf("No modal handler for \"%s\" found", modalName)
	}

	if modal.Check != nil && !modal.Check(event) {
		return
	}

	if check, ok := modal.Checks[modalName]; ok && !check(event) {
		return
	}

	handler, ok := modal.Handler[subName]
	if !ok {
		h.Logger.Debugf("不明なハンダラ %s", subName)
		return
	}
	defer deferUpdateInteraction(event, modal.Ephemeral != nil && modal.Ephemeral[subName])
	if err := handler(event); err != nil {
		h.Logger.Errorf("Failed to handle modal interaction for \"%s\" : %s", modalName, err)
	}
}
