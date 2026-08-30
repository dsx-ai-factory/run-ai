package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type MetricStoreType string

const (
	MSTGrafanaLabs MetricStoreType = "grafanalabs"
	MSTMimir       MetricStoreType = "mimir"
	MSTThanos      MetricStoreType = "thanos"
)

type MetricStoreDetails struct {
	Type  MetricStoreType     `gorm:"not null"`
	Read  MetricStoreEndpoint `gorm:"not null;"`
	Write MetricStoreEndpoint `gorm:"not null;"`
}

func (m *MetricStoreDetails) Scan(value interface{}) error {
	if value == nil {
		*m = MetricStoreDetails{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan MetricStoreDetails: %v", value)
	}

	return json.Unmarshal(bytes, m)
}

// Value implements the driver.Valuer interface
func (m MetricStoreDetails) Value() (driver.Value, error) {
	return json.Marshal(m)
}
