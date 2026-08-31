/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloadtype

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cli/completion"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workloadtype/list/manipulate"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

// For outputting the autocomplete options in pascal case
var lowercaseToPascalCase = map[manipulate.WorkloadTypeColumn]string{
	manipulate.ColumnKeyName:           "Name",
	manipulate.ColumnKeyCategory:       "Category",
	manipulate.ColumnKeyPriority:       "Priority",
	manipulate.ColumnKeyPreemptibility: "Preemptibility",
	manipulate.ColumnKeyGroup:          "Group",
	manipulate.ColumnKeySource:         "Source",
}

// ColumnsFlagCompletion provides autocomplete for the --columns flag with multi-value support
func ColumnsFlagCompletion(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	currentColumns, previousColumns, allColumnsSplit := completion.ParseColumnsForCompletion(toComplete)
	selectedColumnsMap := completion.CreateSelectedColumnsMap(allColumnsSplit)
	sortedCompletion := completion.CreateSortedColumnsCompletions(currentColumns, previousColumns, selectedColumnsMap, lo.Keys(manipulate.ColumnMap), lowercaseToPascalCase)
	return sortedCompletion, cobra.ShellCompDirectiveNoSpace
}
