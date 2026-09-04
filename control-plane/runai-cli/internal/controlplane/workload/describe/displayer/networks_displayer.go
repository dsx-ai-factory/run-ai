/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package displayer

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/field"
)

// Type definitions
type Network struct {
	Name              string   `json:"name"`
	ToolType          string   `json:"toolType"`
	ConnectionType    string   `json:"connectionType"`
	Url               *string  `json:"url,omitempty"`
	AuthorizationType *string  `json:"authorizationType,omitempty"`
	AuthorizedUsers   []string `json:"authorizedUsers,omitempty"`
	AuthorizedGroups  []string `json:"authorizedGroups,omitempty"`
	Port              *int32   `json:"port,omitempty"`
}

// Displayer
type NetworksDisplayer struct {
	networks []Network
}

func NewNetworksSection(networks []Network) *Section {
	return &Section{
		Displayer: &NetworksDisplayer{
			networks: networks,
		},
	}
}

func (n *NetworksDisplayer) Subject() string {
	return TitleNetworks
}

func (n *NetworksDisplayer) Columns() []table.Column {
	return []table.Column{
		{Title: FieldName, Width: NetworkColumnWidth},
		{Title: FieldToolType, Width: NetworkColumnWidth},
		{Title: FieldConnectionType, Width: NetworkColumnWidth},
		{Title: FieldConnectionPort, Width: NetworkColumnWidth},
		{Title: FieldURL, Width: NetworkColumnWidth},
	}
}

func (n *NetworksDisplayer) Data() [][]string {
	var data [][]string
	for _, network := range n.networks {
		row := []string{
			network.Name,
			network.ToolType,
			network.ConnectionType,
			field.Int32OrEmpty(network.Port),
			field.StringOrEmpty(network.Url),
		}
		data = append(data, row)
	}
	return data
}

func (n *NetworksDisplayer) DisplayAfterTable() {
	headerPrinted := false
	columnPadding := GeneralAndRuntimeColumnPadding

	for _, network := range n.networks {
		if network.AuthorizationType == nil && len(network.AuthorizedUsers) == 0 && len(network.AuthorizedGroups) == 0 {
			continue
		}

		if !headerPrinted {
			fmt.Printf("\n  Access Control:\n") //nolint:forbidigo
			headerPrinted = true
		}

		portStr := ""
		if network.Port != nil {
			portStr = fmt.Sprintf(" (Port %d)", *network.Port)
		}
		fmt.Printf("  %s%s:\n", network.Name, portStr) //nolint:forbidigo

		if network.AuthorizationType != nil {
			fmt.Printf("    %-*s %s\n", columnPadding, "Authorization Type:", *network.AuthorizationType) //nolint:forbidigo
		}
		if len(network.AuthorizedUsers) > 0 {
			fmt.Printf("    %-*s %s\n", columnPadding, "Authorized Users / Service Accounts:", strings.Join(network.AuthorizedUsers, ", ")) //nolint:forbidigo
		}
		if len(network.AuthorizedGroups) > 0 {
			fmt.Printf("    %-*s %s\n", columnPadding, "Authorized Groups:", strings.Join(network.AuthorizedGroups, ", ")) //nolint:forbidigo
		}
	}
}
