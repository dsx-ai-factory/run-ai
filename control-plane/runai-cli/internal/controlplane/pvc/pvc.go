/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pvc

import (
	projectservice "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/project"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
)

const (
	listTimer = "pvc_list_timer"
)

var wsClient Lister

func SetClient(wsClientSet Lister) {
	wsClient = wsClientSet
}

func List(txn *transaction.Txn, domain, token, clusterId, projectName string) (*Output, error) {
	var err error
	projId, err := projectservice.GetProjectId(txn, domain, token, clusterId, projectName)
	if err != nil {
		return nil, err
	}

	timerId := txn.AddTimer(listTimer)
	defer txn.StopTimer(timerId)

	if wsClient == nil {
		wsClient, err = NewObjTrackerApiClient(domain, token)
	}

	if err != nil {
		return nil, err
	}
	kotService := NewListerService(wsClient)
	return kotService.ListPvcsBrief(txn.GetCtx(), projId)
}
