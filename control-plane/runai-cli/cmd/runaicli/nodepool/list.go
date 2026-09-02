/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nodepool

import (
	"context"
	"sync"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/clients"

	commonconfig "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/config"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"

	listCommon "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/nodepool"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

var (
	output listCommon.Output
)

var nodepoolService *nodepool.BackendNodepoolService
var getNodepoolService sync.Once

func SetBackendNodepoolService(bns *nodepool.BackendNodepoolService) {
	nodepoolService = bns
}

func GetNodepoolService(domain, token string) (*nodepool.BackendNodepoolService, error) {
	var err error
	getNodepoolService.Do(func() {
		if nodepoolService != nil {
			return
		}
		nodepoolService, err = getBackendNodepoolService(domain, token)
	})
	if err != nil {
		return nil, err
	}
	return nodepoolService, nil
}

// list represents the list command
var list = &cobra.Command{
	Use:   "list",
	Short: "List node pool",
	Args:  cobra.NoArgs,

	PreRunE: func(cmd *cobra.Command, args []string) error {
		txn = instance.GetTxn()
		err := commonconfig.SetOutput(&output)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.ErrorOutputType, err)
		}

		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		var (
			printOut    string
			briefOutput *nodepool.Output
			err         error
		)
		ctx := txn.GetCtx()
		cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
		token := instance.GetToken()
		clusterUuid := instance.GetConfigValue(config.ClusterUUID)

		briefOutput, err = ExecuteNodepoolListBrief(ctx, cpUrl, token, clusterUuid)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListNodePools, err)
		}

		printOut, err = output.Generate(briefOutput)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToListNodePools, err)
		}
		cmd.Println(printOut)
		return nil
	},
}

func init() {
	Cmd.AddCommand(list)

	listCommon.FlagOutputType(list, &output)
}

func getBackendNodepoolService(cpUrl, token string) (*nodepool.BackendNodepoolService, error) {
	client, err := clients.NewClusterService(cpUrl, token)
	if err != nil {
		return nil, err
	}

	return nodepool.NewBackendNodepoolService(client, client), nil
}

func ExecuteNodepoolList(ctx context.Context, cpUrl, token, clusterUuid string) (*openapi.GetNodepools200Response, *openapi.Nodes, api.NodeTelemetriesByType, error) {
	nodeService, err := GetNodepoolService(cpUrl, token)
	if err != nil {
		return nil, nil, nil, err
	}
	return nodeService.GetNodepoolList(ctx, clusterUuid)
}

func ExecuteNodepoolListBrief(ctx context.Context, cpUrl, token, clusterUuid string) (*nodepool.Output, error) {
	nodepools, nodes, telemetry, err := ExecuteNodepoolList(ctx, cpUrl, token, clusterUuid)
	if err != nil {
		return nil, err
	}
	return nodepool.NewOutput(nodepools, nodes, telemetry)
}
