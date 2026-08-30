package models

import (
	"database/sql/driver"
	"encoding/json"
)

type Label struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

func (c Label) Value() (driver.Value, error) {
	valueString, err := json.Marshal(c)
	return string(valueString), err
}

func (c *Label) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &c); err != nil {
		return err
	}
	return nil
}
