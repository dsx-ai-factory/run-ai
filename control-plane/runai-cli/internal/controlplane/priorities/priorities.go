/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package priorities

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
)

var wpClient Lister

func SetClient(wpc Lister) {
	wpClient = wpc
}

func List(txn *transaction.Txn, cpUrl, token, clusterId string) (*Output, error) {
	var err error

	if wpClient == nil {
		wpClient, err = NewWorkloadPrioritiesServiceClient(cpUrl, token)
		if err != nil {
			return nil, err
		}
	}

	prioritiesService := NewPrioritiesService(wpClient)
	return prioritiesService.GetWorkloadClusterPrioritiesBrief(txn.GetCtx(), clusterId)
}
