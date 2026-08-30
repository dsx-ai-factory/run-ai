package models

import (
	"database/sql/driver"
	"encoding/json"
)

type ConditionArray []Condition

func (c ConditionArray) Value() (driver.Value, error) {
	valueString, err := json.Marshal(c)
	return string(valueString), err
}

func (c *ConditionArray) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &c); err != nil {
		return err
	}
	return nil
}
