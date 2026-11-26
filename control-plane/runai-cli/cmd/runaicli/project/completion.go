/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package project

import (
	orgapi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/projectresources"
	apiconfig "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/filter"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/clients"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/completion"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/list"
	projectservice "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/project"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

func GetProjectNamesForArg(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	fn := completion.GetArgCompletionFn(completion.Project, initProjectInstance, getProjectNames)
	return fn(cmd, args, toComplete)
}

func GetProjectNamesForFlag(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	fn := completion.GetFlagCompletionFn(completion.Project, initProjectInstance, getProjectNames)
	return fn(cmd, args, toComplete)
}

func getProjectNames(_ *cobra.Command) ([]string, error) {
	projectBriefs, err := getList(list.NewFilterPageOption())
	if err != nil {
		return nil, err
	}

	names := make([]string, 0)
	for _, brief := range projectBriefs.GetList() {
		names = append(names, brief.Name)
	}
	return names, nil
}

// GPUTypeCompletion suggests the model names advertised by the selected project’s pools.
func GPUTypeCompletion() completion.CompletionFn {
	return completion.GetFlagCompletionFn(completion.Project, initProjectInstance, func(cmd *cobra.Command) ([]string, error) {
		cpURL := instance.GetConfigValue(config.ControlPlaneUrl)
		token := instance.GetToken()
		clusterID := instance.GetConfigValue(config.ClusterUUID)
		projectID := instance.GetConfigValue(config.ClusterProjectID)
		if f := cmd.Flags().Lookup("project"); f != nil && f.Changed {
			var err error
			projectID, err = projectservice.GetProjectId(txn, cpURL, token, clusterID, f.Value.String())
			if err != nil {
				return nil, err
			}
		}
		if projectID == "" {
			return nil, nil
		}
		conf, err := apiconfig.CreateProjectConfig(cpURL, token)
		if err != nil {
			return nil, err
		}
		poolIDs, err := projectresources.GPUEnabledPoolIDs(txn.GetCtx(), orgapi.NewAPIClient(conf).ProjectsAPI, projectID)
		if err != nil || len(poolIDs) == 0 {
			return nil, err
		}
		cluster, err := clients.NewClusterService(cpURL, token)
		if err != nil {
			return nil, err
		}
		inventory, err := cluster.GetNodepoolMinimalList(txn.GetCtx(), filter.NewBuilder().WithClusterID(clusterID).StringList())
		if err != nil {
			return nil, err
		}
		eligible := sets.New(poolIDs...)
		models := sets.New[string]()
		for _, pool := range inventory.Nodepools {
			if !eligible.Has(pool.Id) {
				continue
			}
			for _, gpu := range pool.GpuTypes {
				if gpu.Type != "" {
					models.Insert(gpu.Type)
				}
			}
		}
		return sets.List(models), nil
	})
}
