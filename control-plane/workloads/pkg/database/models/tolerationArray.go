package models

import (
	"database/sql/driver"
	"encoding/json"
)

type TolerationArray []Toleration

func (c TolerationArray) Value() (driver.Value, error) {
	valueString, err := json.Marshal(c)
	return string(valueString), err
}

func (c *TolerationArray) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &c); err != nil {
		return err
	}
	return nil
}
