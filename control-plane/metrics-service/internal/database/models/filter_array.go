package models

import (
	"database/sql/driver"
	"encoding/json"
)

type FilterArray []Filter

func (c FilterArray) Value() (driver.Value, error) {
	valueString, err := json.Marshal(c)
	return string(valueString), err
}

func (c *FilterArray) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &c); err != nil {
		return err
	}
	return nil
}
