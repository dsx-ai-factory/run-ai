package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type Condition struct {
	Type               string     `json:"type,omitempty"`
	Status             string     `json:"status,omitempty"`
	Message            string     `json:"message,omitempty"`
	Reason             string     `json:"reason,omitempty"`
	LastTransitionTime *time.Time `json:"last_transition_time"`
}

func (c Condition) Value() (driver.Value, error) {
	valueString, err := json.Marshal(c)
	return string(valueString), err
}

func (c *Condition) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &c); err != nil {
		return err
	}
	return nil
}
