/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datasources

import (
	"context"

	"github.com/run-ai/runai/control-plane/metrics-service/internal/connections"
)

type ConnectionGetter interface {
	GetConnection(ctx context.Context, tenantId int32) (*connections.Connection, error)
}
