package displayer

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/run-ai/runai/control-plane/runai-cli/pkg/ui/field"
)

const (
	TitleNetworkTopology = "Network Topology"

	FieldNetworkTopologyName = "Network topology name"
	FieldRequiredPlacement   = "Required Placement"
	FieldPreferredPlacement  = "Preferred Placement"
	FieldActualPlacement     = "Actual Placement"

	NetworkTopologyColumnWidth = 20
)

type NetworkTopology struct {
	Name               *string `json:"name,omitempty"`
	RequiredPlacement  *string `json:"requiredPlacement,omitempty"`
	PreferredPlacement *string `json:"preferredPlacement,omitempty"`
	ActualPlacement    *string `json:"actualPlacement,omitempty"`
}

type NetworkTopologyDisplayer struct {
	networkTopology *NetworkTopology
}

func NewNetworkTopologySection(networkTopology *NetworkTopology) *Section {
	return &Section{
		Displayer: &NetworkTopologyDisplayer{
			networkTopology: networkTopology,
		},
	}
}

func (n *NetworkTopologyDisplayer) Subject() string {
	return TitleNetworkTopology
}

func (n *NetworkTopologyDisplayer) Columns() []table.Column {
	return []table.Column{
		{Title: FieldNetworkTopologyName, Width: NetworkTopologyColumnWidth},
		{Title: FieldRequiredPlacement, Width: NetworkTopologyColumnWidth},
		{Title: FieldPreferredPlacement, Width: NetworkTopologyColumnWidth},
		{Title: FieldActualPlacement, Width: NetworkTopologyColumnWidth},
	}
}

func (n *NetworkTopologyDisplayer) Data() [][]string {
	if n.networkTopology == nil {
		return nil
	}
	return [][]string{
		{
			field.StringOrEmpty(n.networkTopology.Name),
			field.StringOrEmpty(n.networkTopology.RequiredPlacement),
			field.StringOrEmpty(n.networkTopology.PreferredPlacement),
			field.StringOrEmpty(n.networkTopology.ActualPlacement),
		},
	}
}

func FormatActualPlacement(level, value *string) *string {
	if level == nil || *level == "" {
		return nil
	}
	result := *level
	if value != nil && *value != "" {
		result = fmt.Sprintf("%s (%s)", *level, *value)
	}
	return &result
}
