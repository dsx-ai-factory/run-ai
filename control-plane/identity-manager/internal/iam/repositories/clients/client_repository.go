/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package clients

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
)

type ClientRepository interface {
	CreateClient(ctx context.Context, realm string, client *models.Client) (string, error)
	UpdateClient(ctx context.Context, realm string, client *models.Client) error
	GetClient(ctx context.Context, realm string, ID string) (*models.Client, error)
	UpdateRedirectUris(ctx context.Context, realm string, ID string, redirectUris []string) error
	DeleteClient(ctx context.Context, realm string, ID string) error
	GetClients(ctx context.Context, realm string, filter *models.ClientFilter) ([]models.Client, error)
	GetClientSecret(ctx context.Context, realm string, ID string) (string, error)
	GetClientByClientId(ctx context.Context, realm string, clientId string) (*models.Client, error)
	CreateExchangePolicy(ctx context.Context, realm string, idpAlias, targetClientId string) error
	GetClientsCount(ctx context.Context, realm string, filter *models.ClientFilter) (int, error)
	RegenerateClientSecret(ctx context.Context, realm string, ID string) (string, error)
	ToggleClient(ctx context.Context, realm string, ID string, enabled bool) error
	UpdateClientStatus(ctx context.Context, realm string, ID string, isEnabled bool) error
	GetAccessKeyClientById(ctx context.Context, realm string, createdBy string, id string) (*models.Client, error)
	SetEmailScopeOptional(ctx context.Context, realm string, idOfClient string) error
	SetEmailScopeDefault(ctx context.Context, realm string, idOfClient string) error
	UpdatePostLogoutRedirectUris(ctx context.Context, realm string, clientId string, customUri string) error
	DeleteAccessKeysByEmail(ctx context.Context, realm string, email string) error
}
