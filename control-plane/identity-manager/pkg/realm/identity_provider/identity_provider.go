/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package identity_provider

import (
	"context"

	iamModels "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/models"
)

type IdpRepresentationer interface {
	GetRepresentation(ctx context.Context, data *models.IdpData, config *map[string]string) *iamModels.Idp
	GetMappersRepresentation(ctx context.Context, alias string, mappers map[string]string) []*iamModels.IdpMapper
}
