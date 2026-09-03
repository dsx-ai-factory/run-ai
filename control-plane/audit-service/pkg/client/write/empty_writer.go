/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package write

import (
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/audit-service/pkg/client/generated"
)

type EmptyWriter struct{}

func (d *EmptyWriter) Write(log openapi.AuditLog) error {
	return nil
}
