/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/orm"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"

	"github.com/google/uuid"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"gorm.io/gorm/clause"
)

const (
	// FictiveTenantId - we store a set of preset assets under a fictive tenant having max int32 as its id
	FictiveTenantId = math.MaxInt32

	// CurrentPresetsRevision - as indication for installing presets for a given tenant. For now there is no meaning
	// for the value, but in the future we may benefit a value that is numeric and not boolean
	CurrentPresetsRevision = 5 // Update to 5 to support chatbot-ui deprecation from 20250004 migration
)

// keep a set of all tenants for which preset assets clone has been attempted. this is in addition
// to tenant setting flag that we check when the service restarts, and this set is gone.
type GuardedTenantsSet struct {
	tenants datastructs.Set[int32]
	*sync.RWMutex
}

var presetsCloneAttemptedForTenant = GuardedTenantsSet{
	tenants: make(datastructs.Set[int32]),
	RWMutex: &sync.RWMutex{},
}

type AssetSupportingPresets interface {
	dbmodel.EnvironmentAssetRec |
		dbmodel.ComputeAssetRec |
		dbmodel.HostPathAssetRec |
		dbmodel.NFSAssetRec |
		dbmodel.GitAssetRec |
		dbmodel.S3AssetRec

	TableName() string  // we need to know the name of the specs table for this asset
	GetAssetId() string // a function to return the asset id from a record holding the spec
}

// clonePresetAssetsToTenant copy the set of preloaded assets to a given tenant.
func clonePresetAssetsToTenant(ctx context.Context, tenantId int32) {
	//
	//   check if we attempted presets loading for this tenant in the current life of this service
	//
	presetsCloneAttemptedForTenant.RLock()
	if presetsCloneAttemptedForTenant.tenants.Has(tenantId) {
		presetsCloneAttemptedForTenant.RUnlock()
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("preset assets clone already attempted for tenant %v", tenantId)
		return
	}
	presetsCloneAttemptedForTenant.tenants.Add(tenantId)
	presetsCloneAttemptedForTenant.RUnlock()

	//
	//  we are cloning multiple tables, we have to make sure it is all or nothing
	//
	tx := runaidb.DB().WithContext(ctx).Begin()
	committed := false
	defer orm.FinalizeTransaction(tx, &committed)

	//
	//  check if we already loaded presets for this tenant in the past
	//
	presetsRevision, revErr := getTenantPresetsRevision(ctx, tx, tenantId)
	if revErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to load presets revision for tenant %v: %v", tenantId, revErr)
	}
	if presetsRevision == CurrentPresetsRevision {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("already loaded presets for tenant %v", tenantId)
		return
	}

	//
	// list all the assets of the fictive tenant id, these are the presets we want to clone
	//
	var records []dbmodel.AssetMetaRec
	if findErr := tx.Where("tenant_id = ?", FictiveTenantId).Find(&records); findErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to list preset assets: %v", findErr)
		return
	}

	//
	// Keep a map b/w the asset id of any preset asset to its correlated id under the tenant we're cloning to.
	// we need it for copying the asset specs, in which we do not have tenant id.
	//
	presetAssetIdToTenantAssetId := make(map[string]string)

	//
	// Keep a list of preset asset ids for each kind of asset.
	//
	presetAssetIdPerKind := make(map[api.AssetKind][]string)
	for assetKind := range LoadPresetSpecsByKind {
		presetAssetIdPerKind[assetKind] = []string{}
	}

	//
	// Fetch all WorkloadSupportedTypes for the preset assets to clone them properly
	//
	presetAssetIds := make([]string, 0, len(records))
	for _, record := range records {
		presetAssetIds = append(presetAssetIds, record.Id)
	}
	var presetWorkloadSupportedTypes []dbmodel.WorkloadSupportedTypesRec
	if findErr := tx.Where("asset_id IN ?", presetAssetIds).Find(&presetWorkloadSupportedTypes); findErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to load preset workload supported types: %v", findErr)
		return
	}

	// Map preset asset ID to its workload supported types
	presetAssetIdToWorkloadTypes := make(map[string]*dbmodel.WorkloadSupportedTypes)
	for i := range presetWorkloadSupportedTypes {
		presetAssetIdToWorkloadTypes[presetWorkloadSupportedTypes[i].AssetId] = &presetWorkloadSupportedTypes[i].WorkloadSupportedTypes
	}

	//
	// Clone each metadata record, replacing tenantId from fictive to the id of the relevant tenant, and new uuid
	// for the asset id to prevent duplicate entries. keep track of this preset asset id -> tenant asset id mapping,
	// because we need the same ids for the specs of these assets.
	//
	for _, record := range records {
		newRecord := record
		newRecord.Id = uuid.New().String()
		newRecord.TenantId = convutils.Ref(tenantId)
		newRecord.CreatedAt = time.Now()
		newRecord.UpdatedAt = time.Now()

		_, createErr := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&newRecord)
		if createErr != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to clone preloaded asset %s of kind %s to tenant %v id %v: %v",
				record.Id, record.Kind, tenantId, newRecord.Id, createErr)
		} else {
			presetAssetIdToTenantAssetId[record.Id] = newRecord.Id
			presetAssetIdPerKind[record.Kind] = append(presetAssetIdPerKind[record.Kind], record.Id)
		}

		// Clone the workload supported types from the source preset asset
		if sourceWorkloadTypes, exists := presetAssetIdToWorkloadTypes[record.Id]; exists {
			recordOfWorkloadSupportedType := dbmodel.WorkloadSupportedTypesRec{
				WorkloadSupportedTypes: dbmodel.WorkloadSupportedTypes{
					AssetId:       newRecord.Id,
					Workspace:     sourceWorkloadTypes.Workspace,
					Training:      sourceWorkloadTypes.Training,
					Inference:     sourceWorkloadTypes.Inference,
					Distributed:   sourceWorkloadTypes.Distributed,
					DistFramework: sourceWorkloadTypes.DistFramework,
				},
			}

			if _, createErr := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&recordOfWorkloadSupportedType); createErr != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to clone preloaded workload supported type, with asset id %s , error: %v",
					recordOfWorkloadSupportedType.AssetId, createErr)
			}
		} else {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("no workload supported types found for preset asset %s", record.Id)
		}
	}

	//
	// Now that we're done with the metadata, clone the specs
	//
	for assetKind, ids := range presetAssetIdPerKind {
		if cloneErr := clonePresetSpecsToTenant(ctx, tx, tenantId, assetKind, ids, presetAssetIdToTenantAssetId); cloneErr != nil {
			return
		}
	}

	if commitErr := tx.Commit(); commitErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to clone preset assets to tenant %v: %v", tenantId, commitErr)
		return
	}
	committed = true

	//
	//   update the tenant settings flag to avoid loading the presets again for this tenant.
	//   we're doing it outside the transaction, as a separate update because we do not want to
	//   remove the presets only because we could not update the flag
	//
	if setErr := setTenantPresetsRevision(runaidb.DB().WithContext(ctx), tenantId, CurrentPresetsRevision); setErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("failed updating the preset settings flag for tenant %v: %v", tenantId, setErr)
		return
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("successfully cloned preset assets to tenant %v", tenantId)
}

// Generic functions to load/store specs of preset assets for each kind of asset.
// Specs require special handing, because the spec tables do not hold the tenant id, so
// we have to use asset ids from the metadata records that we cloned.
// We have two functions for this purpose
// - LoadPresetSpecsByKind -> loads the specs of all the relevant assets from the fictive tenant
// - StorePresetSpecs -> stores those specs with the same asset id as was created for this asset in the metadata

type (
	LoadPresetSpecsProto  func(ctx context.Context, tx orm.DB, assetKind api.AssetKind, ids []string, tenantId int32) (interface{}, error)
	StorePresetSpecsProto func(ctx context.Context, tx orm.DB, assetKind api.AssetKind, specs interface{}, tenantId int32) error
)

var LoadPresetSpecsByKind = map[api.AssetKind]LoadPresetSpecsProto{
	api.ASSETKIND_ENVIRONMENT: LoadPresetSpecs[dbmodel.EnvironmentAssetRec],
	api.ASSETKIND_COMPUTE:     LoadPresetSpecs[dbmodel.ComputeAssetRec],
	api.ASSETKIND_HOST_PATH:   LoadPresetSpecs[dbmodel.HostPathAssetRec],
	api.ASSETKIND_NFS:         LoadPresetSpecs[dbmodel.NFSAssetRec],
	api.ASSETKIND_GIT:         LoadPresetSpecs[dbmodel.GitAssetRec],
	api.ASSETKIND_S3:          LoadPresetSpecs[dbmodel.S3AssetRec],
}

var StorePresetSpecsByKind = map[api.AssetKind]StorePresetSpecsProto{
	api.ASSETKIND_ENVIRONMENT: StorePresetSpecs[dbmodel.EnvironmentAssetRec],
	api.ASSETKIND_COMPUTE:     StorePresetSpecs[dbmodel.ComputeAssetRec],
	api.ASSETKIND_HOST_PATH:   StorePresetSpecs[dbmodel.HostPathAssetRec],
	api.ASSETKIND_NFS:         StorePresetSpecs[dbmodel.NFSAssetRec],
	api.ASSETKIND_GIT:         StorePresetSpecs[dbmodel.GitAssetRec],
	api.ASSETKIND_S3:          StorePresetSpecs[dbmodel.S3AssetRec],
}

func clonePresetSpecsToTenant(ctx context.Context, tx orm.DB, tenantId int32, assetKind api.AssetKind, assetIds []string, presetAssetIdToTenantAssetId map[string]string) error {
	specs, loadErr := LoadPresetSpecsByKind[assetKind](ctx, tx, assetKind, assetIds, tenantId)
	if loadErr != nil {
		return loadErr
	}

	//
	// translate the asset ids of the preset specs to the same asset ids we created for them in assets_metadata
	//
	switch specsArr := specs.(type) {
	case []dbmodel.EnvironmentAssetRec:
		dbmodel.ReplaceEnvironmentSpecIds(specsArr, presetAssetIdToTenantAssetId)
	case []dbmodel.ComputeAssetRec:
		dbmodel.ReplaceComputeSpecIds(specsArr, presetAssetIdToTenantAssetId)
	case []dbmodel.HostPathAssetRec:
		dbmodel.ReplaceHostPathSpecIds(specsArr, presetAssetIdToTenantAssetId)
	case []dbmodel.NFSAssetRec:
		dbmodel.ReplaceNFSSpecIds(specsArr, presetAssetIdToTenantAssetId)
	case []dbmodel.GitAssetRec:
		dbmodel.ReplaceGitSpecIds(specsArr, presetAssetIdToTenantAssetId)
	case []dbmodel.S3AssetRec:
		dbmodel.ReplaceS3SpecIds(specsArr, presetAssetIdToTenantAssetId)
	}

	if storeErr := StorePresetSpecsByKind[assetKind](ctx, tx, assetKind, specs, tenantId); storeErr != nil {
		return storeErr
	}
	return nil
}

func LoadPresetSpecs[PK AssetSupportingPresets](ctx context.Context, tx orm.DB, assetKind api.AssetKind, ids []string, tenantId int32) (interface{}, error) {
	var rec PK
	specs := make([]PK, 0, len(ids))

	err := tx.Where(fmt.Sprintf("%s.asset_id in ?", rec.TableName()), ids).Find(&specs)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to find matching specs for preset assets of kind %s for tenant %v: %v", assetKind, tenantId, err)
		return nil, err
	}

	return specs, nil
}

func StorePresetSpecs[PK AssetSupportingPresets](ctx context.Context, tx orm.DB, assetKind api.AssetKind, specsInt interface{}, tenantId int32) error {
	specs := specsInt.([]PK)
	for _, spec := range specs {
		_, err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&spec)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to clone preloaded spec %s of kind %s to tenant %v: %v", spec.GetAssetId(), assetKind, tenantId, err)
			return err
		}
	}
	return nil
}
