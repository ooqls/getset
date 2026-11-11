package initv1

import (
	"context"

	"github.com/ooqls/getset/db/bootstrap"
	"github.com/ooqls/getset/eventsource/eventsourcingv1"
	"github.com/ooqls/getset/eventsource/eventsourcingv1/tables"
	"github.com/ooqls/getset/log"
	"go.uber.org/zap"
)

var l *zap.Logger = log.NewLogger("init")

// Init prepares the source tables, returning the first bootstrap failure.
// Already completed statements are not rolled back.
func Init(ctx context.Context, sources ...eventsourcingv1.EventSource) error {
	if len(sources) == 0 {
		return nil
	}

	l.Info("Bootstrapping database with entity tables")
	if err := bootstrap.Run(ctx, bootstrap.Plan{
		Tables: tables.GetCreateTableStmts(sources...),
	}, bootstrap.SQLX("")); err != nil {
		return err
	}

	for _, source := range sources {
		l.Info("Initialized entity",
			zap.String("event_source", string(source)),
		)
	}
	return nil
}
