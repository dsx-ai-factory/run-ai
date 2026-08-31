/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datasource

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/datasource/client"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/datasource/describe"
	projectservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/project"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
)

var lister describe.Lister

// SetClient sets the datasource lister client (for testing)
func SetClient(datasourceLister describe.Lister) {
	lister = datasourceLister
}

func GetId(txn *transaction.Txn, domain, token, dsName, dsType, clusterId, projectId string) (string, error) {
	if dsName == "" {
		return "", nil
	}

	dsClient, err := getClient(domain, token)
	if err != nil {
		return "", err
	}

	projectID, err := projectservice.ParseProjectIdInt32(projectId)
	if err != nil {
		return "", err
	}

	var clusterIDPtr *string
	if clusterId != "" {
		clusterIDPtr = &clusterId
	}

	var dsTypePtr *string
	if dsType != "" {
		dsTypePtr = &dsType
	}

	ds, err := describe.FindByName(txn.GetCtx(), dsClient, dsName, projectID, clusterIDPtr, dsTypePtr)
	if err != nil {
		return "", err
	}

	return ds.Meta.Id, nil
}

func getClient(domain, token string) (describe.Lister, error) {
	if lister != nil {
		return lister, nil
	}
	return client.NewClient(domain, token)
}
