/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package node

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/completion"
	nodeservice "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/node"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

func GetNodeNamesForFlag(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	fn := completion.GetFlagCompletionFn(completion.Node, initNodeInstance, getNodeNames)
	return fn(cmd, args, toComplete)
}

func getNodeNames(_ *cobra.Command) ([]string, error) {
	nodesHandler, err := executeNodeList()
	if err != nil {
		return nil, err
	}
	nodes := nodesHandler.GetList()
	names := make([]string, 0)
	for _, brief := range nodes {
		names = append(names, brief.Name)
	}
	return names, nil
}

func GetGPUTypesForFlag(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return completion.GetFlagCompletionFn(completion.GPUType, initNodeInstance, getGPUTypes)(cmd, args, toComplete)
}
func getGPUTypes(_ *cobra.Command) ([]string, error) {
	service, err := nodeservice.GetNodeLister(instance.GetConfigValue(config.ControlPlaneUrl), instance.GetToken())
	if err != nil {
		return nil, err
	}
	return service.GPUTypes(instance.GetTxn().GetCtx(), instance.GetConfigValue(config.ClusterUUID))
}
