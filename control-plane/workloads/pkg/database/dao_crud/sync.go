/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao_crud

import (
	"context"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
)

const (
	SequenceLastValue     = "last_value"
	SequenceWorkloadsCrud = "crud_workloads_sequence"
)

func GetHighestUpdateCount(ctx context.Context) (int64, *api.Error) {
	var currentValue int64
	rowsAffected, err := runaidb.DB().WithContext(ctx).Table(SequenceWorkloadsCrud).Select(SequenceLastValue).Scan(&currentValue)
	if err != nil {
		return 0, api.NewError(http.StatusServiceUnavailable, "failed to read highest update count", err).WithDetailsFromDbResult(rowsAffected, err)
	}
	return currentValue, nil
}
