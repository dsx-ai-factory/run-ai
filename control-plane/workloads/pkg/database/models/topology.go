package models

import (
	"database/sql/driver"
	"encoding/json"
)

type Topology struct {
	Name               string `json:"name,omitempty"`
	PreferredPlacement string `json:"preferredPlacement,omitempty"`
	RequiredPlacement  string `json:"requiredPlacement,omitempty"`
}

func (c Topology) Value() (driver.Value, error) {
	valueString, err := json.Marshal(c)
	return string(valueString), err
}

func (c *Topology) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &c); err != nil {
		return err
	}
	return nil
}
