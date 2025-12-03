package eventsourcingv1

import (
	"github.com/jmoiron/sqlx"
	"github.com/ooqls/getset/log"
	"go.uber.org/zap"
)

var l *zap.Logger = log.NewLogger("sql")

type Iterator[T any] func() (*T, error)

func NewArrayIterator[T any](arr []T) Iterator[T] {
	i := 0
	return func() (*T, error) {
		if i < len(arr) {
			v := arr[i]
			i++
			return &v, nil
		}
		return nil, nil
	}
}

func NewArrayEventIterator(arr []Event) Iterator[Event] {
	i := 0
	return func() (*Event, error) {
		if i < len(arr) {
			v := arr[i]
			i++
			return &v, nil
		}
		return nil, nil
	}
}

func NewSQLEventIterator(rows *sqlx.Rows) Iterator[Event] {
	next := func() (*Event, error) {
		var v Event
		if rows.Next() {
			if err := rows.StructScan(&v); err != nil {
				l.Error("Failed to scan row", zap.Error(err))
				return nil, err
			}
			return &v, nil
		} else {
			err := rows.Close()
			if err != nil {
				l.Error("Error closing rows", zap.Error(err))
			}
		}

		return nil, nil
	}

	return next
}
