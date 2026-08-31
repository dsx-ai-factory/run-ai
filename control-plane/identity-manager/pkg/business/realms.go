/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package business

import (
	"context"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/models"
)

//go:generate mockery --name RealmDal
type RealmDal interface {
	CreateRealm(ctx context.Context, realm *models.Realm) (*models.Realm, error)
	DeleteRealm(ctx context.Context, realmName string) (*uuid.UUID, error)
	DisableRealm(ctx context.Context, realmName string) (*uuid.UUID, error)
	CreateRealms(ctx context.Context, realm *[]models.Realm) (int64, error)
	GetRealmByTenantId(ctx context.Context, tenantId string) (*models.Realm, error)
	GetRealms(ctx context.Context) (*[]models.Realm, error)
	GetRealmByName(ctx context.Context, name string) (*models.Realm, error)
	GetRealmByNameWithDeleted(ctx context.Context, name string) (*models.Realm, error)
}
