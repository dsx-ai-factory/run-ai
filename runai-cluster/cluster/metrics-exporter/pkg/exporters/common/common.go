/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/scheduling/v2alpha2"
)

const (
	ProjectLabel = "project"
)

func GetDepartmentNameOfQueueFromLabels(queueLabels map[string]string) string {
	if departmentName, found := queueLabels[v2alpha2.DepartmentNameLabel]; found {
		return departmentName
	}
	return ""
}

func GetProjectNameOfQueueFromLabels(queueLabels map[string]string) string {
	if projectName, found := queueLabels[ProjectLabel]; found {
		return projectName
	}
	return ""
}
