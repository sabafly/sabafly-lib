package handler

import (
	"github.com/google/uuid"
	"github.com/sabafly/disgo/bot"
)

type RawHandler func(event bot.Event) error

type Event struct {
	ID *uuid.UUID

	Check   Check[bot.Event]
	Handler RawHandler
}

func (h *Handler) handleEvent(event bot.Event) {
	for _, e := range h.Event {
		if e.Check != nil && !e.Check(event) {
			continue
		}

		if err := e.Handler(event); err != nil {
			h.Logger.Errorf("Failed to handle raw event: %v", event)
		}
	}
}
