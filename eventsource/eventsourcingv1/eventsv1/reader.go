package eventsv1

import (
	"context"
	"fmt"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/ooqls/getset/eventsource/eventsourcingv1"
)

var _ Reader = &SQLReader{}

type EventIterator func() bool

type Reader interface {
	Get(ctx context.Context, entityId uuid.UUID) (eventsourcingv1.Iterator[eventsourcingv1.Event], error)
	GetAll(ctx context.Context) (eventsourcingv1.Iterator[eventsourcingv1.Event], error)
	Count(ctx context.Context, entityId uuid.UUID) (int64, error)
}

type Options struct {
	cache cache.CacheInterface[[]eventsourcingv1.Event]
}

type opt func(o *Options)

func WithRedisCache() opt {
	return func(o *Options) {
		redisCached := eventsourcingv1.NewRedisCache[string]()
		cache := eventsourcingv1.NewJsonCache[[]eventsourcingv1.Event](redisCached)
		o.cache = cache
	}
}

func NewSQLReader(db *sqlx.DB, eventTable eventsourcingv1.EventSource, opts ...opt) *SQLReader {
	options := Options{}
	for _, opt := range opts {
		opt(&options)
	}
	return &SQLReader{
		eventTable: eventTable,
		db:         db,
		options:    options,
	}
}

type SQLReader struct {
	eventTable eventsourcingv1.EventSource
	db         *sqlx.DB
	options    Options
}

func (r *SQLReader) Get(ctx context.Context, entityId uuid.UUID) (eventsourcingv1.Iterator[eventsourcingv1.Event], error) {
	if r.options.cache != nil {
		cached, err := r.options.cache.Get(ctx, entityId)
		if err == nil {
			return eventsourcingv1.NewArrayEventIterator(cached), nil
		}
	}

	rows, err := SQLGetEvent(ctx, r.db, entityId, r.eventTable)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %v", err)
	}

	return eventsourcingv1.NewSQLEventIterator(rows), nil
}

func (r *SQLReader) GetAll(ctx context.Context) (eventsourcingv1.Iterator[eventsourcingv1.Event], error) {
	rows, err := r.db.QueryxContext(ctx, fmt.Sprintf(GetAllEventsTableFmt, r.eventTable))
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %v", err)
	}

	return eventsourcingv1.NewSQLEventIterator(rows), nil
}

func (r *SQLReader) Count(ctx context.Context, entityId uuid.UUID) (int64, error) {
	return SQLCountEvents(ctx, r.db, entityId, r.eventTable)
}
