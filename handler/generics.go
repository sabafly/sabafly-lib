package handler

import (
	"github.com/disgoorg/log"
	"github.com/google/uuid"
)

type genericsHandler[T any] func(event *T) error

type generics[T any] struct {
	ID *uuid.UUID

	Check   Check[*T]
	Handler genericsHandler[T]
}

type genericsList[T any] struct {
	Map   map[uuid.UUID]generics[T]
	Array []generics[T]

	Logger log.Logger
}

func (g *genericsList[T]) Add(gen generics[T]) func() {
	if gen.ID != nil {
		g.Map[*gen.ID] = gen
		return func() {
			delete(g.Map, *gen.ID)
		}
	} else {
		g.Array = append(g.Array, gen)
		return nil
	}
}

func (g *genericsList[T]) Adds(gen ...generics[T]) {
	g.Array = append(g.Array, gen...)
}

func (g *genericsList[T]) handleEvent(event *T) {
	for _, gen := range g.Map {
		g.run(gen, event)
	}
	for _, gen := range g.Array {
		g.run(gen, event)
	}
}

func (g *genericsList[T]) run(generic generics[T], event *T) {
	if generic.Check != nil && generic.Check(event) {
		return
	}
	if err := generic.Handler(event); err != nil {
		g.Logger.Errorf("failed to handle event %T: %s", *event, err.Error())
		return
	}
}
