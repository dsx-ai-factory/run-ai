/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package compute

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/asset"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/compute/client"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/compute/describe"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
)

var lister describe.Lister

// SetClient sets the compute lister client (for testing)
func SetClient(computeLister describe.Lister) {
	lister = computeLister
}

func GetId(txn *transaction.Txn, domain, token, computeName, clusterId, projectId string) (string, error) {
	if computeName == "" {
		return "", nil
	}

	computeClient, err := getClient(domain, token)
	if err != nil {
		return "", err
	}

	params, err := asset.ParseResolveParams(projectId, clusterId)
	if err != nil {
		return "", err
	}

	compute, err := describe.FindByName(txn.GetCtx(), computeClient, computeName, params.ProjectID, params.ClusterID)
	if err != nil {
		return "", err
	}

	return compute.Meta.Id, nil
}

func getClient(domain, token string) (describe.Lister, error) {
	if lister != nil {
		return lister, nil
	}
	return client.NewClient(domain, token)
}
