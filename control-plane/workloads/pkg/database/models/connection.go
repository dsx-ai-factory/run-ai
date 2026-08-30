package models

import (
	"database/sql/driver"
	"encoding/json"
)

type Connection struct {
	Name                string   `json:"name,omitempty"`
	Type                string   `json:"type,omitempty"`
	Url                 string   `json:"url,omitempty"`
	AuthorizationType   string   `json:"authorizationType,omitempty"`
	AuthorizedUsers     []string `json:"authorizedUsers,omitempty"`
	AuthorizedGroups    []string `json:"authorizedGroups,omitempty"`
	ConnectionType      string   `json:"connectionType,omitempty"`
	ContainerPort       int32    `json:"containerPort,omitempty"`
	CorrelationName     string   `json:"correlationName,omitempty" gorm:"not null;"`
	ExposedUrlSubDomain *string  `json:"exposedUrlSubDomain,omitempty" gorm:""`
	ExposedUrlPath      *string  `json:"exposedUrlPath,omitempty" gorm:""`
	PortResourceName    *string  `json:"portResourceName,omitempty" gorm:""`
	ServiceResourceName *string  `json:"serviceResourceName,omitempty" gorm:""`
}

func (c Connection) Value() (driver.Value, error) {
	valueString, err := json.Marshal(c)
	return string(valueString), err
}

func (c *Connection) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &c); err != nil {
		return err
	}
	return nil
}
