/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/cluster/client"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/cluster/list"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/cluster/list/manipulate"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/cluster/service"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
)

const (
	TimerList = "cluster_list_timer"
)

var clusterClient service.ClusterClient

func SetClient(projClientSet service.ClusterClient) {
	clusterClient = projClientSet
}

func GetClusterList(txn *transaction.Txn, domain, token string) (*list.Output, error) {
	var err error

	ptnId := txn.AddTimer(TimerList)
	defer txn.StopTimer(ptnId)

	clusterService, err := getClusterService(domain, token)
	if err != nil {
		return nil, err
	}
	return clusterService.GetClusterList(txn.GetCtx())
}

func GetClusterByUuid(txn *transaction.Txn, domain, token, uuid string) (*manipulate.BriefCluster, error) {
	var err error

	ptnId := txn.AddTimer(TimerList)
	defer txn.StopTimer(ptnId)

	clusterService, err := getClusterService(domain, token)
	if err != nil {
		return nil, err
	}
	return clusterService.GetClusterByUuid(txn.GetCtx(), uuid)
}

func GetClusterByName(txn *transaction.Txn, domain, token, name string) (*manipulate.BriefCluster, error) {
	ptnId := txn.AddTimer(TimerList)
	defer txn.StopTimer(ptnId)

	clusterService, err := getClusterService(domain, token)
	if err != nil {
		return nil, err
	}

	return clusterService.GetClusterByName(txn.GetCtx(), name)
}

func getClusterService(domain, token string) (*service.BackendClusterService, error) {
	var err error
	if clusterClient == nil {
		clusterClient, err = client.NewBackendClusterClient(domain, token)
	}

	if err != nil {
		return nil, err
	}

	return service.NewBackendClusterService(clusterClient), nil
}
