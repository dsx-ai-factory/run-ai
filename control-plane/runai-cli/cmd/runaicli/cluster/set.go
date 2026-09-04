/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster

import (
	"errors"
	"fmt"

	argshandler "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/args"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/attributes"

	clusterservice "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/cluster"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cmdtxn"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/consts"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/cluster/list/manipulate"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/spf13/cobra"
)

const (
	clusterIdAttr             = "id"
	ErrorShouldBeOneAttribute = "need to set either cluster ID or name"
)

var (
	clusterId   string
	clusterName string
)

var set = &cobra.Command{
	Use:   "set [CLUSTER_NAME]",
	Short: consts.ClusterSetShortMsg,
	Args:  cobra.MaximumNArgs(1),

	PreRunE: func(cmd *cobra.Command, args []string) error {
		clusterName = argshandler.GetNameFromArgs(&args)
		err := attributes.MutuallyExclusiveAtLeastOne(clusterId, clusterName)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd,
				fmt.Sprintf(consts.ErrorClusterSetClusterIdWithName, clusterIdAttr),
				err)
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {
		var err error
		var cluster *manipulate.BriefCluster

		cpUrl := instance.GetConfigValue(config.ControlPlaneUrl)
		token := instance.GetToken()

		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToSetCluster, err)
		}

		if clusterId != "" {
			cluster, err = clusterservice.GetClusterByUuid(txn, cpUrl, token, clusterId)
		} else if clusterName != "" {
			cluster, err = clusterservice.GetClusterByName(txn, cpUrl, token, clusterName)
		} else {
			// should be blocked my cobra
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToSetCluster, errors.New(ErrorShouldBeOneAttribute))
		}

		if cluster == nil || err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToSetCluster, err)
		}

		err = instance.SetCluster(cluster.Uuid, cluster.Name, cluster.Domain)
		if err != nil {
			return cmdtxn.CmdErrorPrintAndSet(cmd, consts.FailedToSetCluster, err)
		}
		cmd.Println(fmt.Sprintf(auth.SetClusterProperties, cluster.Name))
		return nil
	},
}

func init() {
	Cmd.AddCommand(set)
	set.Flags().StringVar(&clusterId, clusterIdAttr, "", "set by cluster ID")
}

func Clear() {
	clusterId = ""
	clusterName = ""
}
