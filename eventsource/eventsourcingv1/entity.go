package eventsourcingv1

import "github.com/google/uuid"

type Entity interface {
	GetId() uuid.UUID
	Apply(Event) error
}
