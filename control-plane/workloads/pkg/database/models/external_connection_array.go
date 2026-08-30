package models

import (
	"database/sql/driver"
	"encoding/json"
)

type ExternalConnectionArray []Connection

func (c ExternalConnectionArray) Value() (driver.Value, error) {
	valueString, err := json.Marshal(c)
	return string(valueString), err
}

func (c *ExternalConnectionArray) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &c); err != nil {
		return err
	}
	return nil
}
