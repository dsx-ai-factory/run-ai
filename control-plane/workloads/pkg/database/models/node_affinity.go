package models

import (
	"database/sql/driver"
	"encoding/json"
)

type NodeAffinity struct {
	Required  *NodeSelector          `json:"required,omitempty"`
	Preferred *PreferredNodeSelector `json:"preferred,omitempty"`
}

type NodeSelector struct {
	NodeSelectorTerms []NodeSelectorTerm `json:"nodeSelectorTerms"`
}

type PreferredNodeSelector struct {
	NodeSelectorTerms []PreferredNodeSelectorTerm `json:"nodeSelectorTerms"`
}

type PreferredNodeSelectorTerm struct {
	Weight     int32            `json:"weight"`
	Preference NodeSelectorTerm `json:"preference"`
}

type NodeSelectorTerm struct {
	MatchExpressions []NodeSelectorRequirement `json:"matchExpressions,omitempty"`
}

type NodeSelectorRequirement struct {
	Key      string   `json:"key"`
	Operator string   `json:"operator"`
	Values   []string `json:"values,omitempty"`
}

func (c NodeAffinity) Value() (driver.Value, error) {
	valueString, err := json.Marshal(c)
	return string(valueString), err
}

func (c *NodeAffinity) Scan(value interface{}) error {
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
