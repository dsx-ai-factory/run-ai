/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/datavolume/client"
	projectservice "github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/project"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
)

var listClient Lister

func SetClient(lc Lister) {
	listClient = lc
}

func List(txn *transaction.Txn, cpUrl, token, projectName, clusterId string) (*Output, error) {
	var err error
	var projectID *string

	if listClient == nil {
		listClient, err = client.NewClient(cpUrl, token)
		if err != nil {
			return nil, err
		}
	}

	if projectName != "" {
		projectIdStr, err := projectservice.GetProjectId(txn, cpUrl, token, clusterId, projectName)
		if err != nil {
			return nil, err
		}
		projectID = &projectIdStr
	}

	datavolumeService := NewService(listClient)
	output, err := datavolumeService.List(txn.GetCtx(), projectID)
	if err != nil {
		return nil, err
	}

	return output, nil
}
