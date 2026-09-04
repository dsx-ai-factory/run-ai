/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"

	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/common"

	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
	"gorm.io/gorm/clause"
)

// value of all is used to include all origins. this is done because empty value is reserved
// to existing (backward compatibility considerations).
const AllOrigins = "all"

func ListSecrets(ctx context.Context, filter *models.SecretFilters, options *models.ListOptions) ([]models.Secret, *api.Error) {
	if options == nil {
		options = &models.ListOptions{}
	}

	record := models.NewSecret(filter)
	dbErr := api.NewError(http.StatusServiceUnavailable, record.ErrorMessage("list"), nil)

	db := runaidb.DB().WithContext(ctx)

	query, err := filter.Apply(ctx, db.Model(record), options.IncludeAncestors, options.IncludeDescendants)
	if err != nil {
		return nil, err
	}

	if !options.IncludePrivate {
		query = query.Where(fmt.Sprintf("%s IS NULL OR %s = ?", models.SecretPrivateColumn, models.SecretPrivateColumn), false)
	}

	secrets := make([]models.Secret, 0)
	if err := query.Order(filter.SecretName).Find(&secrets); err != nil {
		dbErr = dbErr.WithDetails(err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, err)

		return nil, dbErr
	}

	return secrets, nil
}

func CreateSecret(ctx context.Context, secret *models.Secret) (*models.Secret, *api.Error) {
	dbErr := api.NewError(http.StatusServiceUnavailable, secret.ErrorMessage("insert"), nil)

	db := runaidb.DB()

	if err := upsertSecret(db, secret); err != nil {
		dbErr = dbErr.WithDetails(err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, err)

		return nil, dbErr
	}

	return secret, nil
}

func upsertSecret(db orm.DB, secret *models.Secret) error {
	record := secret.ScopeParams()
	record.SecretName = secret.SecretName

	err := db.Where(&record).First(&record)
	if err != nil && errors.Is(err, orm.ErrRecordNotFound) {
		_, err = db.Clauses(clause.OnConflict{DoNothing: true}).Create(secret)
	} else {
		secret.ID = record.ID
		_, err = db.Save(secret)
	}

	return err
}

func DeleteSecret(ctx context.Context, filters *models.SecretFilters) *api.Error {
	record := models.NewSecret(filters)
	dbErr := api.NewError(http.StatusServiceUnavailable, record.ErrorMessage("delete"), nil)

	db := runaidb.DB()

	if _, err := db.
		Where(filters).
		Delete(record); err != nil {
		dbErr = dbErr.WithDetails(err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, err)

		return dbErr
	}

	return nil
}

func ApplySecrets(ctx context.Context, filters *models.SecretFilters, secrets []api.Secret) (map[string]api.Secret, *api.Error) {
	record := models.NewSecret(filters)
	dbErr := api.NewError(http.StatusServiceUnavailable, record.ErrorMessage("apply"), nil)

	db := runaidb.DB()

	tx := db.Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	names := make([]string, 0, len(secrets))
	// Keep track on the ids of the secrets that were applied to relate the issues with the secrets
	secretIdMap := make(map[string]api.Secret, len(secrets))

	for _, secret := range secrets {
		toAdd := record.ScopeParams()
		toAdd.SecretName = ptr.To(secret.Name)
		toAdd.SecretType = ptr.To(string(secret.Type))
		toAdd.Origin = OriginPtrToStringPtr(secret.Origin)
		toAdd.Private = secret.Private

		names = append(names, secret.Name)

		if err := upsertSecret(tx, toAdd); err != nil {
			dbErr = dbErr.WithDetails(err)
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, err)

			return nil, dbErr
		}
		secretIdMap[toAdd.ID.String()] = secret
	}

	// Delete all records that are not in the list. for example
	// if I have in the table secret aaa bbb ccc and I receive from the sync bbb ccc ddd
	// I need to upsert ccc and ddd, but I also not to remove aaa (because it is no longer in the cluster)
	if _, err := tx.
		Where(filters).
		Where(fmt.Sprintf("%s not in ?", models.SecretNameColumn), names).
		Delete(record); err != nil {
		dbErr = dbErr.WithDetails(err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, err)

		return nil, dbErr
	}

	if err := tx.Commit(); err != nil {
		dbErr = dbErr.WithDetails(err)
		if dbErr.Code != http.StatusConflict {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, err)
			return nil, dbErr
		}
	}
	committed = true

	return secretIdMap, nil
}

func OriginPtrToStringPtr(origin *api.Origin) *string {
	if origin == nil || *origin == "" {
		return ptr.To(string(api.ORIGIN_EXISTING))
	} else {
		return ptr.To(string(*origin))
	}
}

func OriginFilterToStringPtr(originFilter string) *string {
	if originFilter == AllOrigins {
		return nil
	}

	var originToSearch string

	if originFilter == "" {
		// default should preserve previous behavior to show only existing. to show both customer-existing
		// and assets related, "all" should be specified"
		originToSearch = string(api.ORIGIN_EXISTING)
	} else {
		originToSearch = originFilter
	}
	return &originToSearch
}

func DeleteSecretsNotInCluster(ctx context.Context, filters *models.SecretFilters, secretIds []string) *api.Error {
	record := models.NewSecret(filters)
	dbErr := api.NewError(http.StatusServiceUnavailable, record.ErrorMessage("delete"), nil)

	db := runaidb.DB()

	tenantId, err := common.TenantIdFromContext(ctx)
	if err != nil {
		return err
	}

	query := db.Model(record).
		Where("tenant_id = ?", tenantId).
		Where(filters)

	if len(secretIds) != 0 {
		query = query.Not("id in ?", secretIds)
	}

	if _, dbError := query.Delete(record); dbError != nil {
		dbErr = dbErr.WithDetails(dbError)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, dbError)

		return dbErr
	}

	return nil
}

// DeleteProjectSecrets deletes secrets of a project
func DeleteProjectSecrets(ctx context.Context, tenantId int32, projectId string) *api.Error {
	return deleteEntitiesByScope(ctx, openapi.SCOPETYPE_PROJECT, tenantId, projectId, &models.Secret{})
}

// DeleteClusterSecrets deletes secrets within a cluster
func DeleteClusterSecrets(ctx context.Context, tenantId int32, clusterId string) *api.Error {
	return deleteClusterEntities(ctx, tenantId, clusterId, &models.Secret{})
}

// DeleteTenantSecrets deletes secrets within a tenant
func DeleteTenantSecrets(ctx context.Context, tenantId int32) *api.Error {
	return deleteTenantEntities(ctx, tenantId, &models.Secret{})
}
