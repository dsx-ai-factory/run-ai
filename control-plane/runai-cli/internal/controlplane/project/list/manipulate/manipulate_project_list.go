/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	cliutils "github.com/run-ai/runai/control-plane/runai-cli/internal/utils"
	"k8s.io/utils/ptr"
)

func CreateAlignedProjects(projectsRaw Data, columns []ProjectColumn) ProjectsAlignedOutput {
	projectsAligned := make([]ProjectsAlignedRow, 0)
	for _, project := range projectsRaw.Projects {
		projectRow := &ProjectsAlignedRow{}
		for _, column := range columns {
			ColumnMap[column].extractFunc(ptr.To(project), projectRow)
		}
		projectsAligned = append(projectsAligned, *projectRow)
	}
	return ProjectsAlignedOutput{
		Projects:  projectsAligned,
		NextToken: projectsRaw.NextToken,
	}
}

func CreateTableData(projectsAligned []ProjectsAlignedRow, columns []ProjectColumn) [][]string {
	var data [][]string

	for _, project := range projectsAligned {
		var projectRow []string
		for _, column := range columns {
			projectRow = append(projectRow, ColumnMap[column].formatFunc(ptr.To(project)))
		}
		data = append(data, projectRow)
	}

	cliutils.MultiColumnSort(data, []cliutils.SortColumn{
		{Index: nameColumnIndex, Order: cliutils.Ascending},
	})

	return data
}
