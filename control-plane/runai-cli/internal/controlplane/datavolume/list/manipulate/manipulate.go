/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	"fmt"
	"strings"

	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/datavolumes-service-client/generated"
)

func ParseDatavolumeColumns(columnNames []string) ([]DatavolumeColumn, error) {
	var columns []DatavolumeColumn
	for _, name := range columnNames {
		normalizedName := strings.ToLower(strings.TrimSpace(name))
		column := DatavolumeColumn(normalizedName)

		if _, exists := ColumnMap[column]; !exists {
			return nil, fmt.Errorf("invalid column name: %s", name)
		}
		columns = append(columns, column)
	}
	return columns, nil
}

func BriefDatavolumeFromEntry(dv openapi.DatavolumeNoSharedScopes) BriefDatavolume {
	return BriefDatavolume{
		Name:      dv.Name,
		Scope:     extractScope(&dv),
		OriginPvc: dv.OriginPvcName,
		Status:    datavolumeStatusString(datavolumePhase(&dv)),
	}
}

func CreateAlignedDatavolumes(datavolumes []openapi.DatavolumeNoSharedScopes, columns []DatavolumeColumn) DatavolumesAlignedData {
	var aligned []DatavolumesAlignedRow

	for _, dv := range datavolumes {
		row := DatavolumesAlignedRow{}
		for _, col := range columns {
			fieldInfo := ColumnMap[col]
			fieldInfo.extractFunc(&dv, &row)
		}
		aligned = append(aligned, row)
	}

	return DatavolumesAlignedData{Datavolumes: aligned}
}

func CreateTableData(datavolumes []DatavolumesAlignedRow, columns []DatavolumeColumn) [][]string {
	var data [][]string

	for _, dvRow := range datavolumes {
		var row []string
		for _, col := range columns {
			fieldInfo := ColumnMap[col]
			row = append(row, fieldInfo.formatFunc(&dvRow))
		}
		data = append(data, row)
	}

	return data
}
