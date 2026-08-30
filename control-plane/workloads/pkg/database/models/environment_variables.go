package models

import (
	"database/sql/driver"
	"encoding/json"
)

type EnvironmentVariables map[string]string

func (c EnvironmentVariables) Value() (driver.Value, error) {
	valueString, err := json.Marshal(c)
	return string(valueString), err
}

func (c *EnvironmentVariables) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &c); err != nil {
		return err
	}
	return nil
}
