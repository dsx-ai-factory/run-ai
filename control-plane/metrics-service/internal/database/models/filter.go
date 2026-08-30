package models

import (
	"database/sql/driver"
	"encoding/json"

	"github.com/lib/pq"
)

type Filter struct {
	Key    string         `json:"key"`
	Values pq.StringArray `gorm:"type:text[]" json:"values"`
}

func (c Filter) Value() (driver.Value, error) {
	valueString, err := json.Marshal(c)
	return string(valueString), err
}

func (c *Filter) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &c); err != nil {
		return err
	}
	return nil
}
