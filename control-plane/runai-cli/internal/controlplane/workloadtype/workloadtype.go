/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloadtype

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/filter"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workloadtype/client"
	workloadtypelist "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workloadtype/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
)

func List(txn *transaction.Txn, cpUrl, token string, fb *filter.Builder, externalTypesOnly *bool) (*workloadtypelist.Output, error) {
	listClient, err := client.NewClient(cpUrl, token)
	if err != nil {
		return nil, err
	}
	service := workloadtypelist.NewService(listClient)
	return service.List(txn.GetCtx(), cpUrl, token, fb, externalTypesOnly)
}
