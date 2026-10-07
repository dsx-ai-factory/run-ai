/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package realms

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
)

type RealmRepository interface {
	CreateRealm(ctx context.Context, realm *models.Realm) (string, error)
	DisableRealm(ctx context.Context, realmName string) error
	GetRealms(ctx context.Context) ([]*models.Realm, error)
	GetRealm(ctx context.Context, realm string) (*models.Realm, error)
	UpdateRealm(ctx context.Context, realmRepresentation models.Realm) error
	DeleteRealm(ctx context.Context, realmName string) error

	// realm settings
	UpdateAutoRedirectSSO(ctx context.Context, realm string, idpAlias *string, isEnabled bool) error
	SetRealmEmailScopeOptional(ctx context.Context, realm string) error
}
