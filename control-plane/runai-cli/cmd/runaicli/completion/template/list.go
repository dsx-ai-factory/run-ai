/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package template

import (
	listCompletion "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cli/completion"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/template/list/manipulate"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var templateLowercaseToPascalCase = map[manipulate.TemplateColumn]string{
	manipulate.ColumnKeyName:                   "Name",
	manipulate.ColumnKeyScopeType:              "ScopeType",
	manipulate.ColumnKeyImage:                  "Image",
	manipulate.ColumnKeyGpuComputeRequest:      "GPUComputeRequest",
	manipulate.ColumnKeyWorkloadType:           "WorkloadType",
	manipulate.ColumnKeyMandatoryFieldsMissing: "MandatoryFieldsMissing",
	manipulate.ColumnKeyCreatedAt:              "CreatedAt",
}

func ColumnsFlagCompletion(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	currentColumns, previousColumns, allColumnsSplit := listCompletion.ParseColumnsForCompletion(toComplete)
	selectedColumnsMap := listCompletion.CreateSelectedColumnsMap(allColumnsSplit)
	sortedCompletion := listCompletion.CreateSortedColumnsCompletions(currentColumns, previousColumns, selectedColumnsMap, lo.Keys(manipulate.ColumnMap), templateLowercaseToPascalCase)
	return sortedCompletion, cobra.ShellCompDirectiveNoSpace
}
