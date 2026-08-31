/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/business"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/models"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"gorm.io/gorm/clause"
)

const (
	RealmsTable = "realms"
)

func NewRealmDao() business.RealmDal {
	return &Realm{}
}

type Realm struct{}

func (r Realm) TableName() string {
	return RealmsTable
}

func (r Realm) CreateRealm(ctx context.Context, realm *models.Realm) (*models.Realm, error) {
	_, err := runaidb.DB().WithContext(ctx).Table(r.TableName()).Create(realm)
	if err != nil {
		return nil, utils.ParseGormError(err)
	}
	return realm, nil
}

func (r Realm) CreateRealms(ctx context.Context, realms *[]models.Realm) (int64, error) {
	rowsAffected, err := runaidb.DB().WithContext(ctx).Table(r.TableName()).Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(&realms, 500)
	if err != nil {
		return 0, utils.ParseGormError(err)
	}
	return rowsAffected, nil
}

func (r Realm) GetRealmByTenantId(ctx context.Context, tenantId string) (*models.Realm, error) {
	realm := &models.Realm{
		TenantId: tenantId,
	}
	return r.getRealm(ctx, realm, false)
}

func (r Realm) DeleteRealm(ctx context.Context, realmName string) (*uuid.UUID, error) {
	realm := &models.DeleteRealm{Realm: realmName}
	_, err := runaidb.DB().WithContext(ctx).Table(r.TableName()).Unscoped().Where(realm).Delete(realm)
	if err != nil {
		return nil, utils.ParseGormError(err)
	}
	return &realm.ID, nil
}

func (r Realm) DisableRealm(ctx context.Context, realmName string) (*uuid.UUID, error) {
	realm := &models.Realm{Realm: realmName}
	_, err := runaidb.DB().WithContext(ctx).Table(r.TableName()).Where(realm).Delete(realm)
	if err != nil {
		return nil, utils.ParseGormError(err)
	}
	return &realm.ID, nil
}

func (r Realm) GetRealmByName(ctx context.Context, name string) (*models.Realm, error) {
	realm := &models.Realm{
		Realm: name,
	}
	return r.getRealm(ctx, realm, false)
}

func (r Realm) getRealm(ctx context.Context, realm *models.Realm, isDeleted bool) (*models.Realm, error) {
	query := runaidb.DB().WithContext(ctx).Table(r.TableName()).Where(realm)
	if isDeleted {
		query = query.Unscoped()
	}
	err := query.First(realm)
	if err != nil {
		return nil, utils.ParseGormError(err)
	}
	return realm, nil
}

func (r Realm) GetRealms(ctx context.Context) (*[]models.Realm, error) {
	tenants := &[]models.Realm{}
	err := runaidb.DB().WithContext(ctx).Table(r.TableName()).Find(tenants)
	if err != nil {
		return nil, utils.ParseGormError(err)
	}
	return tenants, nil
}

func (r Realm) GetRealmByNameWithDeleted(ctx context.Context, name string) (*models.Realm, error) {
	realm := &models.Realm{
		Realm: name,
	}
	return r.getRealm(ctx, realm, true)
}
