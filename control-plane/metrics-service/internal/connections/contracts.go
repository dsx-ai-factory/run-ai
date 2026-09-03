/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package connections

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/generated/pkg/api"
)

type MetricStoreGetter interface {
	GetByTenantId(ctx context.Context, tenantId int32) (*api.MetricStore, error)
}
