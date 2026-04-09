package tables

import (
	"fmt"

	"github.com/ooqls/getset/eventsource/eventsourcingv1"
	"github.com/ooqls/getset/eventsource/eventsourcingv1/events"
)

func GetCreateTableStmts(evs ...eventsourcingv1.EventSource) []string {
	allStmts := []string{}
	for _, ev := range evs {
		allStmts = append(allStmts, fmt.Sprintf(events.CreateEventsTableFmt, string(ev)))
	}
	return allStmts

}
