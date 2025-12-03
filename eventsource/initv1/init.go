package initv1

import (
	"context"

	"github.com/ooqls/getset/db/sqlx"
	"github.com/ooqls/getset/eventsource/eventsourcingv1"
	"github.com/ooqls/getset/eventsource/eventsourcingv1/tablesv1"
	"github.com/ooqls/getset/log"
	"go.uber.org/zap"
)

var l *zap.Logger = log.NewLogger("init")

func Init(ctx context.Context, sources ...eventsourcingv1.EventSource) {
	if len(sources) == 0 {
		return
	}

	l.Info("Seeding database with entity tables")
	sqlx.SeedSQLX(tablesv1.GetCreateTableStmts(sources...), []string{})

	for _, source := range sources {
		l.Info("Initialized entity",
			zap.String("event_source", string(source)),
		)
	}
}
