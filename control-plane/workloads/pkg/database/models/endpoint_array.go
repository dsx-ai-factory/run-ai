package models

import (
	"database/sql/driver"
	"encoding/json"
)

type Endpoint struct {
	Host       string            `json:"host,omitempty"`
	Port       int32             `json:"port,omitempty"`
	Path       string            `json:"path,omitempty"`
	TlsEnabled *bool             `json:"tlsEnabled,omitempty"`
	Protocol   string            `json:"protocol,omitempty"`
	Url        string            `json:"url,omitempty"`
	SourceRef  EndpointSourceRef `json:"sourceRef,omitempty"`
}

type EndpointSourceRef struct {
	Kind string `json:"kind,omitempty"`
	Name string `json:"name,omitempty"`
}

type EndpointArray []Endpoint

func (e EndpointArray) Value() (driver.Value, error) {
	if e == nil {
		return nil, nil
	}
	valueString, err := json.Marshal(e)
	return string(valueString), err
}

func (e *EndpointArray) Scan(value interface{}) error {
	if err := json.Unmarshal(value.([]byte), &e); err != nil {
		return err
	}
	return nil
}
