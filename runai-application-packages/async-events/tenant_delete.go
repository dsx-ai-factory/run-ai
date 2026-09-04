/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package async_events

type TenantDelete struct {
	TenantId string `json:"tenant_id"`
}

func NewTenantDelete(tenantId string) *TenantDelete {
	return &TenantDelete{
		TenantId: tenantId,
	}
}
