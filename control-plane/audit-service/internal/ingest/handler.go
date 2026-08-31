/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package ingest

import (
	"github.com/run-ai/runai/control-plane/audit-service/internal/database/models"
)

type Handler interface {
	Stop() error
	Handle(log *models.AuditLog) error
	Next() Handler
}
