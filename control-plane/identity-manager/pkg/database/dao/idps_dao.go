/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/business"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
)

const (
	IdpsTable = "idps"
)

func NewIdpDao() business.IdpDal {
	return &Idps{}
}

type Idps struct{}

func (i Idps) TableName() string {
	return IdpsTable
}

// SaveIdpMetadata implements business.IdpDal.
func (i Idps) SaveIdpMetadata(ctx context.Context, idpMetadata *models.IdpMetadata) error {
	_, err := runaidb.DB().WithContext(ctx).Table(i.TableName()).Create(idpMetadata)
	if err != nil {
		return utils.ParseGormError(err)
	}
	return nil
}

// GetIdpMetadata implements business.IdpDal.
func (i Idps) GetIdpMetadata(ctx context.Context, tenantId string, idpAlias string) (*models.IdpMetadata, error) {
	resIdp := &models.IdpMetadata{}
	err := runaidb.DB().WithContext(ctx).Table(i.TableName()).Where(&models.IdpMetadata{
		TenantId: tenantId,
		Alias:    idpAlias,
	}).First(resIdp)
	if err != nil {
		return nil, utils.ParseGormError(err)
	}
	return resIdp, nil
}

// DeleteIdpMetadata implements business.IdpDal.
func (i Idps) DeleteIdpMetadata(ctx context.Context, tenantId string, idpAlias string) error {
	idpToDelete := &models.IdpMetadata{
		TenantId: tenantId,
		Alias:    idpAlias,
	}
	_, err := runaidb.DB().WithContext(ctx).Table(i.TableName()).Where(idpToDelete).Delete(idpToDelete)
	if err != nil {
		return utils.ParseGormError(err)
	}
	return nil
}

func (i Idps) UpdateIdpMetadata(ctx context.Context, newIdpMetadata *models.IdpMetadata) error {
	_, err := runaidb.DB().WithContext(ctx).Table(i.TableName()).Where(&models.IdpMetadata{
		TenantId: newIdpMetadata.TenantId,
		Alias:    newIdpMetadata.Alias,
	}).Updates(&models.IdpMetadata{
		ResourceName: newIdpMetadata.ResourceName,
		Resource:     newIdpMetadata.Resource,
		ResourceType: newIdpMetadata.ResourceType,
	})
	if err != nil {
		return utils.ParseGormError(err)
	}
	return nil
}
