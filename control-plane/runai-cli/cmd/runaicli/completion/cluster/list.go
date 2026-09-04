/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cli/completion"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/cluster/list/manipulate"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

// lowercaseToPascalCase maps internal column keys to display names for autocompletion
var lowercaseToPascalCase = map[manipulate.ClusterColumn]string{
	manipulate.ColumnKeyName:             "Name",
	manipulate.ColumnKeyURL:              "URL",
	manipulate.ColumnKeyVersion:          "Version",
	manipulate.ColumnKeyStatus:           "Status",
	manipulate.ColumnKeyKubeDistribution: "KubernetesDistribution",
	manipulate.ColumnKeyKubeVersion:      "DistributionVersion",
}

// ColumnsFlagCompletion provides autocomplete for the --columns flag with multi-value support
func ColumnsFlagCompletion(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	currentColumns, previousColumns, columnsList := completion.ParseColumnsForCompletion(toComplete)
	selectedColumnsMap := completion.CreateSelectedColumnsMap(columnsList)
	sortedCompletion := completion.CreateSortedColumnsCompletions(currentColumns, previousColumns, selectedColumnsMap, lo.Keys(manipulate.ColumnMap), lowercaseToPascalCase)
	return sortedCompletion, cobra.ShellCompDirectiveNoSpace
}
