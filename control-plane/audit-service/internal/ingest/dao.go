/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package ingest

import (
	"context"

	"github.com/run-ai/runai/control-plane/audit-service/internal/database/models"
)

type Dao interface {
	Store(ctx context.Context, batch []*models.AuditLog) error
}
