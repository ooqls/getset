package eventsv1

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/ooqls/getset/eventsource/eventsourcingv1"
)

var CreateEventsTableFmt = `CREATE TABLE IF NOT EXISTS %s (id SERIAL PRIMARY KEY, entity_id UUID, key TEXT NOT NULL, value JSONB NOT NULL, created TIMESTAMP DEFAULT CURRENT_TIMESTAMP );`
var InsertIntoEventsTableFmt = `INSERT INTO %s (entity_id, key, value) VALUES (:entity_id, :key, :value);`
var GetEventsTableFmt = `SELECT * FROM %s WHERE entity_id = $1 ORDER BY id;`
var GetAllEventsTableFmt = `SELECT * FROM %s ORDER BY id;`
var DeleteEventsTableFmt = `DELETE FROM %s WHERE entity_id = $1;`
var CountEventsTableFmt = `SELECT COUNT(*) FROM %s WHERE entity_id = $1;`

func SQLGetEvent(ctx context.Context, db *sqlx.DB, entityId uuid.UUID, source eventsourcingv1.EventSource) (*sqlx.Rows, error) {
	query := fmt.Sprintf(GetEventsTableFmt, source)
	rows, err := db.QueryxContext(ctx, query, entityId)
	return rows, err
}

func SQLInsertEvent(ctx context.Context, tx *sqlx.Tx, event eventsourcingv1.Event, source eventsourcingv1.EventSource) error {
	_, err := tx.NamedExecContext(ctx, fmt.Sprintf(InsertIntoEventsTableFmt, string(source)), &event)
	if err != nil {
		return err
	}
	return nil
}

func SQLDeleteEvents(ctx context.Context, tx *sqlx.Tx, entityId uuid.UUID, source eventsourcingv1.EventSource) error {
	_, err := tx.ExecContext(ctx, fmt.Sprintf(DeleteEventsTableFmt, source), entityId)
	if err != nil {
		return err
	}
	return nil
}

func SQLCountEvents(ctx context.Context, db *sqlx.DB, entityId uuid.UUID, source eventsourcingv1.EventSource) (int64, error) {
	var count int64
	query := fmt.Sprintf(CountEventsTableFmt, source)
	err := db.QueryRowContext(ctx, query, entityId).Scan(&count)
	return count, err
}
