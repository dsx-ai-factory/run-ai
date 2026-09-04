/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"
	"net/http"

	authorizationClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/k8s-objects-tracker/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
)

// scoper is an interface to interact with a db entity that has a scope (name suggested by ai)
type scoper interface {
	// TableName returns the model table name
	TableName() string
	// TenantIdColumn returns the model tenant id column
	TenantIdColumn() string
	// ScopeTypeColumn returns the model scope type column
	ScopeTypeColumn() string
	// ScopeIdColumn returns the model scope id column
	ScopeIdColumn() string
}

// clusterer in is an interface to interact with a db entity that has a cluster id (name suggested by ai)
type clueterer interface {
	// TableName returns the model table name
	TableName() string
	// TenantIdColumn returns the model tenant id column
	TenantIdColumn() string
	// ClusterIdColumn returns the model cluster id column
	ClusterIdColumn() string
}

// deleteEntitiesByScope deletes all rows that of a scope type with the target id
func deleteEntitiesByScope(ctx context.Context, scopeType authorizationClient.ScopeType, tenantId int32, targetId string, scoper scoper) *api.Error {
	dbErr := api.NewError(http.StatusServiceUnavailable, "Failed to delete project entities", nil)

	db := runaidb.DB().WithContext(ctx)

	rows, err := db.Table(scoper.TableName()).
		Where(fmt.Sprintf("%s = ?", scoper.TenantIdColumn()), tenantId).
		Where(fmt.Sprintf("%s = ?", scoper.ScopeTypeColumn()), scopeType).
		Where(fmt.Sprintf("%s = ?", scoper.ScopeIdColumn()), targetId).
		Delete(scoper)

	if err != nil {
		dbErr = dbErr.WithDetailsFromDbResult(rows, err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, err)

		return dbErr
	}

	return nil
}

// deleteClusterEntities deletes all rows that has the target cluster id
// this function is similar to deleteEntitiesByScope, so why not unified them.
// the different is that this function deletes all rows that are related to the cluster (including sibling departments and projects)
// and doesn't use the scope type column as a filter, while deleteEntitiesByScope deletes only the scope and it is sibling
func deleteClusterEntities(ctx context.Context, tenantId int32, clusterId string, clusterer clueterer) *api.Error {
	dbErr := api.NewError(http.StatusServiceUnavailable, "Failed to delete cluster entities", nil)

	db := runaidb.DB().WithContext(ctx)

	_, err := db.Table(clusterer.TableName()).
		Where(fmt.Sprintf("%s = ?", clusterer.TenantIdColumn()), tenantId).
		Where(fmt.Sprintf("%s = ?", clusterer.ClusterIdColumn()), clusterId).
		Delete(clusterer)

	if err != nil {
		dbErr = dbErr.WithDetails(err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, err)

		return dbErr
	}

	return nil
}

// deleteTenantEntities deletes all rows that has the target tenant id
func deleteTenantEntities(ctx context.Context, tenantId int32, scoper scoper) *api.Error {
	dbErr := api.NewError(http.StatusServiceUnavailable, "Failed to delete tenant entities", nil)
	db := runaidb.DB().WithContext(ctx)

	_, err := db.Table(scoper.TableName()).
		Where(fmt.Sprintf("%s = ?", scoper.TenantIdColumn()), tenantId).
		Delete(scoper)

	if err != nil {
		dbErr = dbErr.WithDetails(err)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: %v", dbErr, err)

		return dbErr
	}

	return nil
}
