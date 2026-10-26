/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package business

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/models"
)

type IdpDal interface {
	SaveIdpMetadata(ctx context.Context, idpMetadata *models.IdpMetadata) error
	UpdateIdpMetadata(ctx context.Context, newIdpMetadata *models.IdpMetadata) error
	GetIdpMetadata(ctx context.Context, tenantId string, idpAlias string) (*models.IdpMetadata, error)
	DeleteIdpMetadata(ctx context.Context, tenantId string, idpAlias string) error
}
