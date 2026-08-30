package models

import (
	"database/sql/driver"
	"encoding/json"
)

type Datasource struct {
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
	Id   string `json:"id,omitempty"`
}

func (c Datasource) Value() (driver.Value, error) {
	valueString, err := json.Marshal(c)
	return string(valueString), err
}

func (c *Datasource) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &c); err != nil {
		return err
	}
	return nil
}
