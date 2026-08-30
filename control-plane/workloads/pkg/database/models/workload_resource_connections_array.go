package models

import (
	"database/sql/driver"
	"encoding/json"
)

type WorkloadResourceConnections struct {
	WorkloadId   string                   `json:"workloadId,omitempty"`
	ResourceRefs WorkloadResourceRefArray `json:"resourceRefs,omitempty"`
}

type WorkloadResourceConnectionsArray []WorkloadResourceConnections

func (r WorkloadResourceConnectionsArray) Value() (driver.Value, error) {
	valueString, err := json.Marshal(r)
	return string(valueString), err
}

func (r *WorkloadResourceConnectionsArray) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &r); err != nil {
		return err
	}
	return nil
}
