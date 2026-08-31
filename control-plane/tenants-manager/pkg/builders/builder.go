/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package builders

import (
	"context"

	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/tenants-manager/pkg/domain/models"
)

type TenantInfo struct {
	Name         string
	Id           uint
	MimirId      string
	Email        string
	Password     *string
	ContractType *string
	Role         authorization_api.AllowedInternalCreateRoles
	// If MetricStore is nil, it is set by the metric-store builder; otherwise, the provided value is used as-is.
	MetricStore *models.MetricStoreDetails
}

//go:generate mockery
type Interface interface {
	Name() string
	Construct(ctx context.Context, tenant *TenantInfo, builderResponse map[string]interface{}) error
	Destruct(ctx context.Context, tenant *TenantInfo) error
}
