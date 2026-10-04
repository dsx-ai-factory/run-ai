/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package dao (file rct.go): DAO for the resource claim template asset kind.
// The RCT asset is stored across assets_metadata (shared) and assets_spec
// This file wires those two rows together with typed swagger models.

package dao

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbtables"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"
)

// CreateRct atomically creates a resource claim template asset (assets_metadata
// row + assets_spec row) and returns the typed asset with populated identity fields.
func CreateRct(ctx context.Context, request api.RctCreationRequest, nameValidationQuery *database.ParametrizedQuery) (*api.Rct, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	meta, err := createAssetInTransaction(ctx, tx, api.ASSETKIND_RESOURCE_CLAIM_TEMPLATE, request.Meta, request.Spec, nameValidationQuery)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_CREATIONS, api.ASSETKIND_RESOURCE_CLAIM_TEMPLATE).WithDetails(err)
	}
	committedTx = true

	return &api.Rct{
		Meta: *meta,
		Spec: request.Spec,
	}, nil
}

// RctOptions carries the optional flags of the RCT read paths. IncludeDeleted
// mirrors the credentials/PVC pattern: when the caller wants status_info the
// list must also return soft-deleted rows so the UI can render DELETING /
// DELETION_FAILED for RCTs that are mid-tombstone.
type RctOptions struct {
	IncludeDeleted bool
}

// ListRcts returns a paginated list of resource claim template assets. The
// offset/limit contract matches the swagger: fetch limit+1 rows so we can set
// RctListResponse.Next when the next page exists (i.e. when the returned page
// is full).
func ListRcts(ctx context.Context, offset, limit int32, options RctOptions) (*api.RctListResponse, *api.Error) {
	combinedList := make([]dbmodel.AssetSpecCombinedRec, 0)
	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_RESOURCE_CLAIM_TEMPLATE,
		},
	}

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, nil, assetsSpecJoin())
	if err != nil {
		return nil, err
	}
	applyAssetsSpecTenantFilter(ctx, result)

	if options.IncludeDeleted {
		result = result.Unscoped()
	}

	// Fetch limit+1 so we can detect "is there a next page" without a count query.
	fetchLimit := int(limit) + 1
	_, scanErr := result.
		Order(&query.AssetMeta.Id).
		Offset(int(offset)).
		Limit(fetchLimit).
		Scan(&combinedList)
	if scanErr != nil {
		return nil, api.NewError(http.StatusServiceUnavailable, "error occurred while attempting to list rct assets", scanErr)
	}

	response := api.RctListResponse{
		Entries: make([]api.Rct, 0, len(combinedList)),
	}

	// If we got more than limit rows, there is a next page: trim the extra row
	// and set the next offset. Otherwise leave Next as its zero value (no more pages).
	if int32(len(combinedList)) > limit {
		combinedList = combinedList[:limit]
		response.Next = offset + limit
	}

	for _, row := range combinedList {
		var rctSpec api.RctSpec
		if err := dbmodel.FromJsonBMap(row.Spec, &rctSpec); err != nil {
			return nil, database.Error(database.LIST, api.ASSETKIND_RESOURCE_CLAIM_TEMPLATE).WithDetails(err)
		}
		response.Entries = append(response.Entries, api.Rct{
			Meta: row.AssetMeta,
			Spec: rctSpec,
		})
	}

	return &response, nil
}

// GetRctById fetches a single resource claim template joined from
// assets_metadata + assets_spec, unmarshals the JSON spec into the typed
// RctSpec, and returns the typed asset.
func GetRctById(ctx context.Context, id string, options RctOptions) (*api.Rct, *api.Error) {
	var combinedRec dbmodel.AssetSpecCombinedRec
	query := dbmodel.AssetMetaRec{
		AssetMeta: dbmodel.AssetMeta{
			Kind: api.ASSETKIND_RESOURCE_CLAIM_TEMPLATE,
			Id:   id,
		},
	}

	result, err := GetAssetMeta(ctx, runaidb.DB().WithContext(ctx), query, nil, assetsSpecJoin())
	if err != nil {
		return nil, err
	}
	applyAssetsSpecTenantFilter(ctx, result)

	if options.IncludeDeleted {
		result = result.Unscoped()
	}

	scanRows, scanErr := result.Scan(&combinedRec)
	if scanErr != nil || scanRows == 0 {
		return nil, database.Error(database.GET, api.ASSETKIND_RESOURCE_CLAIM_TEMPLATE).WithDetailsFromDbResult(scanRows, scanErr)
	}

	var rctSpec api.RctSpec
	if err := dbmodel.FromJsonBMap(combinedRec.Spec, &rctSpec); err != nil {
		return nil, database.Error(database.GET, api.ASSETKIND_RESOURCE_CLAIM_TEMPLATE).WithDetails(err)
	}

	return &api.Rct{
		Meta: combinedRec.AssetMeta,
		Spec: rctSpec,
	}, nil
}

// UpdateRct updates the assets_metadata row and overwrites the JSON spec in
// assets_spec for a resource claim template, atomically.
func UpdateRct(ctx context.Context, id string, request api.RctUpdateRequest, nameValidationQuery *database.ParametrizedQuery) (*api.Rct, *api.Error) {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	result, _, err := updateRctInTransaction(ctx, tx, id, request, nameValidationQuery)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, database.Error(database.COMMIT_UPDATES, api.ASSETKIND_RESOURCE_CLAIM_TEMPLATE).WithDetails(err)
	}
	committedTx = true

	return result, nil
}

func updateRctInTransaction(ctx context.Context, tx orm.DB, id string, request api.RctUpdateRequest,
	nameValidationQuery *database.ParametrizedQuery,
) (*api.Rct, *dbmodel.AssetMeta, *api.Error) {
	assetMeta, rowsAffected, apiErr, dbErr := updateAssetMeta(ctx, tx, id, request.Meta, nameValidationQuery, nil)
	if apiErr != nil {
		return nil, nil, apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return nil, nil, database.Error(database.UPDATE, api.ASSETKIND_RESOURCE_CLAIM_TEMPLATE).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	specMap, marshalErr := dbmodel.ToJsonBMap(request.Spec)
	if marshalErr != nil {
		return nil, assetMeta, database.ErrorWithCode(database.UPDATE, api.ASSETKIND_RESOURCE_CLAIM_TEMPLATE, http.StatusBadRequest).WithDetails(marshalErr)
	}

	recId := dbmodel.AssetSpecRec{
		AssetId:  id,
		TenantId: auth.TenantIdOf(ctx),
	}
	update := dbmodel.AssetSpecUpdate{
		Spec: specMap,
	}

	saveRows, saveErr := tx.Model(&dbmodel.AssetSpecRec{}).Where(&recId).Updates(&update)
	if saveErr != nil || saveRows != 1 {
		return nil, assetMeta, database.Error(database.UPDATE, api.ASSETKIND_RESOURCE_CLAIM_TEMPLATE).WithDetailsFromDbResult(saveRows, saveErr)
	}

	if policyErr := deletePolicyComplianceOfAsset(ctx, tx, assetMeta.Id); policyErr != nil {
		return nil, assetMeta, policyErr
	}

	return &api.Rct{
		Meta: assetMeta.ToApiAssetMeta(),
		Spec: request.Spec,
	}, assetMeta, nil
}

// DeleteRct removes a resource claim template. Deletion type follows the
// existing asset pattern (see DeleteNFSAsset / DeleteComputeAsset): SoftDelete
// marks assets_metadata but keeps the spec (so agent sync can still see the
// asset while the delete propagates to the cluster); HardDelete removes the
// spec row plus policy compliance and workload asset references.
func DeleteRct(ctx context.Context, assetMeta *dbmodel.AssetMeta, deletionType DeletionType) *api.Error {
	tx := runaidb.DB().WithContext(ctx).Begin()
	committedTx := false
	defer orm.FinalizeTransaction(tx, &committedTx)

	err := deleteRctInTransaction(ctx, tx, assetMeta.Id, deletionType)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return database.Error(database.COMMIT_DELETIONS, api.ASSETKIND_RESOURCE_CLAIM_TEMPLATE).WithDetails(err)
	}
	committedTx = true

	return nil
}

func deleteRctInTransaction(ctx context.Context, tx orm.DB, id string, deletionType DeletionType) *api.Error {
	rowsAffected, apiErr, dbErr := deleteAssetMeta(ctx, tx, id, deletionType)
	if apiErr != nil {
		return apiErr
	}
	if dbErr != nil || rowsAffected == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_RESOURCE_CLAIM_TEMPLATE).WithDetailsFromDbResult(rowsAffected, dbErr)
	}

	if deletionType == SoftDelete {
		return nil
	}

	recId := dbmodel.AssetSpecRec{
		AssetId:  id,
		TenantId: auth.TenantIdOf(ctx),
	}

	delRows, delErr := tx.Where(&recId).Delete(&recId)
	if delErr != nil || delRows == 0 {
		return database.Error(database.DELETE, api.ASSETKIND_RESOURCE_CLAIM_TEMPLATE).WithDetailsFromDbResult(delRows, delErr)
	}

	if policyErr := deletePolicyComplianceOfAsset(ctx, tx, id); policyErr != nil {
		return policyErr
	}

	if err := deleteWorkloadAssetsByAssetId(ctx, tx, id); err != nil {
		return err
	}

	return nil
}

// assetsSpecJoin returns the qualified column reference that GetAssetMeta uses
// to inner-join the assets_spec table and select only its spec column. Using
// the qualified column form (instead of the whole table) keeps SELECT lean and
// avoids the duplicate tenant_id column that both assets_metadata and
// assets_spec expose.
func assetsSpecJoin() string {
	return fmt.Sprintf("%s.%s", dbtables.TableNameAssetsSpec, "spec")
}

// applyAssetsSpecTenantFilter narrows the joined assets_spec rows to the
// caller's tenant. assets_metadata is already tenant-filtered upstream by
// filterAssetMeta, and we write assets_spec.tenant_id to match at create
// time — but keeping the DB-side filter here means (a) a future code bug
// that writes a mismatched tenant_id can't leak rows across tenants and
// (b) Postgres can use the assets_spec_by_tenant index during the join.
func applyAssetsSpecTenantFilter(ctx context.Context, db orm.DB) {
	db.Where(fmt.Sprintf("%s.tenant_id = ?", dbtables.TableNameAssetsSpec), auth.TenantIdOf(ctx))
}
