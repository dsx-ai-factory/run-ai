/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	"fmt"
	"strings"

	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	cliutils "github.com/run-ai/runai/control-plane/runai-cli/internal/utils"
)

// ClusterColumn represents a column in the cluster list
type ClusterColumn string

const (
	ColumnKeyName             ClusterColumn = "name"
	ColumnKeyURL              ClusterColumn = "url"
	ColumnKeyVersion          ClusterColumn = "version"
	ColumnKeyStatus           ClusterColumn = "status"
	ColumnKeyKubeDistribution ClusterColumn = "kubernetesdistribution"
	ColumnKeyKubeVersion      ClusterColumn = "distributionversion"
)

// FieldInformation represents a column in the cluster list with its field name and display title
type FieldInformation struct {
	columnTitle string
	extractFunc func(cluster *openapi.ClustersMinimalData, clusterRow *ClustersAlignedRow)
	formatFunc  func(clusterRow *ClustersAlignedRow) string
	columnWidth int
}

func (f FieldInformation) GetColumnWidth() int {
	return f.columnWidth
}

func (f FieldInformation) GetColumnTitle() string {
	return f.columnTitle
}

// ColumnMap maps column keys to their corresponding column definitions
var ColumnMap = map[ClusterColumn]FieldInformation{
	ColumnKeyName:             NameField,
	ColumnKeyURL:              URLField,
	ColumnKeyVersion:          VersionField,
	ColumnKeyStatus:           StatusField,
	ColumnKeyKubeDistribution: KubeDistributionField,
	ColumnKeyKubeVersion:      KubeVersionField,
}

var defaultColumns = []ClusterColumn{
	ColumnKeyName,
	ColumnKeyURL,
	ColumnKeyVersion,
	ColumnKeyStatus,
	ColumnKeyKubeDistribution,
	ColumnKeyKubeVersion,
}

func GetDefaultColumns() []ClusterColumn {
	return defaultColumns
}

// ParseClusterColumns parses and validates the provided column names
func ParseClusterColumns(columns []string) ([]ClusterColumn, error) {
	var parsedColumns []ClusterColumn

	for _, column := range columns {
		columnKey := ClusterColumn(strings.ToLower(strings.TrimSpace(column)))
		if _, exists := ColumnMap[columnKey]; !exists {
			return nil, fmt.Errorf("unknown column: %s", column)
		}
		parsedColumns = append(parsedColumns, columnKey)
	}

	return parsedColumns, nil
}

// GetFormattedData returns formatted data for display
func GetFormattedData(alignedRows []ClustersAlignedRow, columns []ClusterColumn) [][]string {
	var data [][]string

	for _, row := range alignedRows {
		var rowData []string

		for _, column := range columns {
			fieldInfo := ColumnMap[column]
			formatted := fieldInfo.formatFunc(&row)
			rowData = append(rowData, formatted)
		}

		data = append(data, rowData)
	}

	// Sort by name column if it exists
	nameIndex := -1
	for i, column := range columns {
		if column == ColumnKeyName {
			nameIndex = i
			break
		}
	}

	if nameIndex >= 0 {
		cliutils.MultiColumnSort(data, []cliutils.SortColumn{
			{Index: nameIndex, Order: cliutils.Ascending},
		})
	}

	return data
}

// GetColumnTitles returns the column titles for the given columns
func GetColumnTitles(columns []ClusterColumn) []string {
	var titles []string
	for _, column := range columns {
		fieldInfo := ColumnMap[column]
		titles = append(titles, fieldInfo.GetColumnTitle())
	}
	return titles
}

// CreateAlignedClusters creates aligned cluster data structure following workload pattern
func CreateAlignedClusters(clustersRaw Data, columns []ClusterColumn) ClustersAlignedOutput {
	clustersAligned := make([]ClustersAlignedRow, 0)
	for _, cluster := range clustersRaw.Clusters {
		clusterRow := ClustersAlignedRow{}
		for _, column := range columns {
			ColumnMap[column].extractFunc(&cluster, &clusterRow)
		}
		clustersAligned = append(clustersAligned, clusterRow)
	}
	return ClustersAlignedOutput{
		Clusters:  clustersAligned,
		NextToken: clustersRaw.NextToken,
	}
}

// CreateTableData creates table data from aligned rows (similar to workload pattern)
func CreateTableData(clustersAligned []ClustersAlignedRow, columns []ClusterColumn) [][]string {
	var data [][]string
	for _, cluster := range clustersAligned {
		var clusterRow []string
		for _, column := range columns {
			clusterRow = append(clusterRow, ColumnMap[column].formatFunc(&cluster))
		}
		data = append(data, clusterRow)
	}
	cliutils.MultiColumnSort(data, []cliutils.SortColumn{
		{Index: 0, Order: cliutils.Ascending}, // Sort by name column (index 0)
	})
	return data
}
