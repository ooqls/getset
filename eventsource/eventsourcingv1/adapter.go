package eventsourcingv1

import "github.com/google/uuid"

// Takes an event and applies it to a generic target
type Adapter[T any] interface {
	Apply(event Event, target *T) error
	GetEntityId(target T) uuid.UUID
}
