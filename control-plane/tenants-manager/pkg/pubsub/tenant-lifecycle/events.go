/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tenant_lifecycle

import (
	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/domain/models"
)

type TenantProvision struct {
	TenantId     uint                                         `json:"tenant_id"`
	TenantName   string                                       `json:"tenant_name"`
	MimirId      string                                       `json:"mimir_id"`
	Email        string                                       `json:"email"`
	Password     *string                                      `json:"password,omitempty"`
	ContractType *string                                      `json:"contract_type,omitempty"`
	Role         authorization_api.AllowedInternalCreateRoles `json:"role"`
	MetricStore  *models.MetricStoreDetails                   `json:"metric_store,omitempty"`
}

type TenantDestruct struct {
	TenantId   uint   `json:"tenant_id"`
	TenantName string `json:"tenant_name"`
	MimirId    string `json:"mimir_id"`
	Email      string `json:"email"`
}
