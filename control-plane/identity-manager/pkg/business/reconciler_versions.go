/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package business

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/models"
)

type ReconcilerVersionsDal interface {
	CreateReconcilerVersions(ctx context.Context, realm *models.ReconcilerVersions) (*models.ReconcilerVersions, error)
	UpdateServiceAccountsReconcilerVersions(ctx context.Context, tenantId string, version string) error
	UpdateAccessKeysReconcilerVersions(ctx context.Context, tenantId string, version string) error
	UpdateInternalServiceAccountsReconcilerVersions(ctx context.Context, tenantId string, version string) error
	GetReconcilerVersionsByTenantId(ctx context.Context, tenantId string) (*models.ReconcilerVersions, error)
	GetReconcilerVersions(ctx context.Context) (*[]models.ReconcilerVersions, error)
}
