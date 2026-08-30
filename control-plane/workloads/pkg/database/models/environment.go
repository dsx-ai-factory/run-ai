package models

import (
	"database/sql/driver"
	"encoding/json"
)

type Environment struct {
	Connections []Connection `json:"connection,omitempty"`
	Name        string       `json:"name,omitempty"`
	Id          string       `json:"id,omitempty"`
	ReplicaType *string      `json:"replica_type"`
}

func (c Environment) Value() (driver.Value, error) {
	valueString, err := json.Marshal(c)
	return string(valueString), err
}

func (c *Environment) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &c); err != nil {
		return err
	}
	return nil
}
