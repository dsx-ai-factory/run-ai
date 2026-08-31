/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package application

import (
	listCompletion "github.com/run-ai/runai/control-plane/runai-cli/internal/cli/completion"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/application/list/manipulate"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var applicationLowercaseToPascalCase = map[manipulate.AIApplicationColumn]string{
	manipulate.ColumnKeyName:            "Name",
	manipulate.ColumnKeyType:            "Type",
	manipulate.ColumnKeyStatus:          "Status",
	manipulate.ColumnKeyProject:         "Project",
	manipulate.ColumnKeyGPUComputeReq:   "GPUComputeReq",
	manipulate.ColumnKeyGPUComputeAlloc: "GPUComputeAlloc",
	manipulate.ColumnKeyGPUMemoryReq:    "GPUMemoryReq",
	manipulate.ColumnKeyGPUMemoryAlloc:  "GPUMemoryAlloc",
	manipulate.ColumnKeyCPUComputeReq:   "CPUComputeReq",
	manipulate.ColumnKeyCPUComputeAlloc: "CPUComputeAlloc",
	manipulate.ColumnKeyCPUMemoryReq:    "CPUMemoryReq",
	manipulate.ColumnKeyCPUMemoryAlloc:  "CPUMemoryAlloc",
}

func ColumnsFlagCompletion(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	currentColumns, previousColumns, allColumnsSplit := listCompletion.ParseColumnsForCompletion(toComplete)
	selectedColumnsMap := listCompletion.CreateSelectedColumnsMap(allColumnsSplit)
	sortedCompletion := listCompletion.CreateSortedColumnsCompletions(currentColumns, previousColumns, selectedColumnsMap, lo.Keys(manipulate.ColumnMap), applicationLowercaseToPascalCase)
	return sortedCompletion, cobra.ShellCompDirectiveNoSpace
}
