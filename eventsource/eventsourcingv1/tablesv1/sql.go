package tablesv1

import (
	"fmt"

	"github.com/ooqls/getset/eventsource/eventsourcingv1"
	"github.com/ooqls/getset/eventsource/eventsourcingv1/eventsv1"
)

func GetCreateTableStmts(events ...eventsourcingv1.EventSource) []string {
	allStmts := []string{}
	for _, ev := range events {
		allStmts = append(allStmts, fmt.Sprintf(eventsv1.CreateEventsTableFmt, string(ev)))
	}

	return allStmts
}
