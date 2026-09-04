/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/credential/client"
	projectservice "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/project"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
)

var listClient Lister

func SetClient(lc Lister) {
	listClient = lc
}

func List(txn *transaction.Txn, cpUrl, token, projectName, clusterId, credType string) (*Output, error) {
	var err error
	var projectID *int32
	var clusterIDPtr *string
	var credTypePtr *string

	if listClient == nil {
		listClient, err = client.NewClient(cpUrl, token)
		if err != nil {
			return nil, err
		}
	}

	if clusterId != "" {
		clusterIDPtr = &clusterId
	}

	if credType != "" {
		credTypePtr = &credType
	}

	if projectName != "" {
		projectIdStr, err := projectservice.GetProjectId(txn, cpUrl, token, clusterId, projectName)
		if err != nil {
			return nil, err
		}
		projectID, err = projectservice.ParseProjectIdInt32(projectIdStr)
		if err != nil {
			return nil, err
		}
	}

	credentialService := NewService(listClient)
	output, err := credentialService.List(txn.GetCtx(), projectID, clusterIDPtr, credTypePtr)
	if err != nil {
		return nil, err
	}

	return output, nil
}
