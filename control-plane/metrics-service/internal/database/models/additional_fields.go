package models

import (
	"database/sql/driver"
	"encoding/json"
)

type AdditionalFields map[string]interface{}

func (c AdditionalFields) Value() (driver.Value, error) {
	if len(c) == 0 {
		return "{}", nil
	}

	valueString, err := json.Marshal(c)
	if err != nil {
		return "{}", err
	}

	return string(valueString), nil
}

func (c *AdditionalFields) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	valueBytes, ok := value.([]byte)
	if !ok || len(valueBytes) == 0 {
		return nil
	}

	if err := json.Unmarshal(valueBytes, &c); err != nil {
		return err
	}
	return nil
}
