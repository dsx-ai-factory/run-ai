/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package idps

import (
	"context"
	"io"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
)

type IdpRepository interface {
	CreateIdp(ctx context.Context, realm string, idp models.Idp) (*string, error)
	GetIdps(ctx context.Context, realm string) ([]*models.Idp, error)
	GetIdp(ctx context.Context, realm string, idpAlias string) (*models.Idp, error)
	GetIdpAliasForExchangeToken(ctx context.Context, realm string) (*string, error)
	UpdateIdp(ctx context.Context, realm string, idp models.Idp) error
	DeleteIdp(ctx context.Context, realm string, idpAlias string) error
	IsIdpConfigured(ctx context.Context, realm string) (bool, error)
	GetIdpConfig(ctx context.Context, realm string, url string, providerId string) (map[string]string, error)
	GetIdpConfigFromFile(ctx context.Context, realm string, providerId string, fileName string, fileBody io.Reader) (map[string]string, error)
	SetBrowserRedirector(ctx context.Context, realm string, alias string) error
	GetSPSamlMetadataUrl(ctx context.Context, realm string, idpAlias string) string

	// idp mappers
	GetIdpMappers(ctx context.Context, realm, alias string) ([]*models.IdpMapper, error)
	UpdateIdpMappers(ctx context.Context, realm, alias string, mappers []*models.IdpMapper) error

	// idp settings
	SetMandatotyClaimMapper(ctx context.Context, realm, alias string, mapper *models.IdpMapper) error
	DeleteMandatoryClaimMapper(ctx context.Context, realm, alias, name string) error
}
