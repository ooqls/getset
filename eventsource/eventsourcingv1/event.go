package eventsourcingv1

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type EventSource string

type EventKey string

type Event struct {
	Id       int64      `db:"id" json:"id"`
	EntityId uuid.UUID  `db:"entity_id" json:"entity_id"`
	Key      EventKey   `db:"key" json:"key"`
	Value    EventData  `db:"value" json:"value"`
	Created  *time.Time `db:"created" json:"created"`
}

type EventData map[string]interface{}

func (data *EventData) Scan(value interface{}) error {
	*data = make(EventData)

	if byteValue, ok := value.([]byte); ok {
		if err := json.Unmarshal(byteValue, data); err != nil {
			return err
		}
	}

	return nil
}

func (data EventData) Value() (driver.Value, error) {
	return json.Marshal(data)
}
