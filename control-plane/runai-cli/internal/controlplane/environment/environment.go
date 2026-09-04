/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package environment

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/asset"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/environment/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/environment/describe"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/spec/spectype"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
)

var lister describe.Lister

// SetClient sets the environment lister client (for testing)
func SetClient(envLister describe.Lister) {
	lister = envLister
}

// GetId fetches environment asset ID by name and validates workload type support
func GetId(txn *transaction.Txn, domain, token, envName, clusterId, projectId string, workloadType spectype.Workload) (string, error) {
	if envName == "" {
		return "", nil
	}

	envClient, err := getClient(domain, token)
	if err != nil {
		return "", err
	}

	params, err := asset.ParseResolveParams(projectId, clusterId)
	if err != nil {
		return "", err
	}

	env, err := describe.FindByName(txn.GetCtx(), envClient, envName, params.ProjectID, params.ClusterID, workloadType.APIString())
	if err != nil {
		return "", err
	}

	return env.Meta.Id, nil
}

func getClient(domain, token string) (describe.Lister, error) {
	if lister != nil {
		return lister, nil
	}
	return client.NewClient(domain, token)
}
