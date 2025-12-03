package eventsv1

import (
	"context"

	"github.com/ooqls/getset/eventsource/eventsourcingv1"
	"go.uber.org/zap"
)

type EventApplicator[T any] struct {
	adapter eventsourcingv1.Adapter[T]
	r       Reader
}

func NewApplicator[T any](r Reader, adapter eventsourcingv1.Adapter[T]) *EventApplicator[T] {
	return &EventApplicator[T]{
		adapter: adapter,
		r:       r,
	}
}

func (a *EventApplicator[T]) Apply(ctx context.Context, ent *T) *ApplicatorError {
	next, err := a.r.Get(ctx, a.adapter.GetEntityId(*ent))
	if err != nil {
		return &ApplicatorError{err, nil}
	}

	appErr := ApplicatorError{}
	ev, err := next()
	for ev != nil {
		err = a.adapter.Apply(*ev, ent)
		if err != nil {
			l.Error("failed to apply event", zap.Error(err), zap.String("event", string(ev.Key)))
			appErr.Events = append(appErr.Events, *ev)
		}
		ev, err = next()
		if err != nil {
			return &ApplicatorError{err, nil}
		}
	}

	if len(appErr.Events) > 0 {
		return &appErr
	}

	return nil
}
