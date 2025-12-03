package eventsv1

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/ooqls/getset/eventsource/eventsourcingv1"
	"github.com/ooqls/getset/log"
	"go.uber.org/zap"
)

var l *zap.Logger = log.NewLogger("eventsv1")
var _ Writer = &SQLWriter{}

type wOpt struct {
	key   string
	value interface{}
}

func WithTransaction(tx *sqlx.Tx) wOpt {
	return wOpt{
		key:   "tx",
		value: tx,
	}
}

type Writer interface {
	// appends the given events to the event store
	Append(ctx context.Context, events ...eventsourcingv1.Event) error
	// deletes all the events for the given entity
	Del(ctx context.Context, entityId uuid.UUID, opts ...wOpt) error
}

func NewSQLWriter(db *sqlx.DB, eventTable eventsourcingv1.EventSource) *SQLWriter {
	return &SQLWriter{
		eventTable: eventTable,
		db:         db,
	}
}

type SQLWriter struct {
	db         *sqlx.DB
	eventTable eventsourcingv1.EventSource
}

func (w *SQLWriter) Append(ctx context.Context, events ...eventsourcingv1.Event) error {
	l.Debug("appending events", zap.Int("count", len(events)))

	tx, err := w.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	for _, event := range events {
		if err := SQLInsertEvent(ctx, tx, event, w.eventTable); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to insert event: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (w *SQLWriter) Del(ctx context.Context, entityId uuid.UUID, opts ...wOpt) error {
	l.Debug("deleting events", zap.String("entity_id", entityId.String()))

	var tx *sqlx.Tx
	var err error
	for _, opt := range opts {
		if opt.key == "tx" {
			tx = opt.value.(*sqlx.Tx)
		}
	}

	if tx == nil {
		tx, err = w.db.BeginTxx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
	}
	return SQLDeleteEvents(ctx, tx, entityId, w.eventTable)

}
