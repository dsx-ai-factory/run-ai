package models

import (
	"database/sql/driver"
	"encoding/json"
)

type EnvironmentArray []Environment

func (c EnvironmentArray) Value() (driver.Value, error) {
	valueString, err := json.Marshal(c)
	return string(valueString), err
}

func (c *EnvironmentArray) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &c); err != nil {
		return err
	}
	return nil
}
