/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"context"
	"fmt"

	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/backend_db"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	"gorm.io/gorm"
)

func forceCleanTables(db *gorm.DB, objects ...interface{}) {
	for _, o := range objects {
		db.Unscoped().Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&o)
	}
}

func cleanTables(db *gorm.DB, objects ...interface{}) {
	for _, o := range objects {
		db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&o)
	}
}

func CleanTables() {
	forceCleanTables(
		runaidb.DB().Actual().WithContext(context.Background()),
		dbmodel.TemplateAssetRec{},
		dbmodel.TemplateContentRec{},

		dbmodel.ComputeAssetRec{},
		dbmodel.EnvironmentAssetRec{},

		dbmodel.S3AssetRec{},
		dbmodel.GitAssetRec{},
		dbmodel.NFSAssetRec{},
		dbmodel.PVCAssetRec{},
		dbmodel.HostPathAssetRec{},
		dbmodel.ConfigMapAssetRec{},
		dbmodel.SecretAssetRec{},

		dbmodel.AccessKeyRecord{},
		dbmodel.PasswordRecord{},
		dbmodel.DockerRegistryRecord{},
		dbmodel.GenericSecretRecord{},

		dbmodel.WorkloadSupportedTypesRec{},
		dbmodel.AssetMetaRec{},

		dbmodel.RegistryRecord{},

		dbmodel.PolicyComplianceRecord{},

		dbmodel.ModelAssetRec{},

		dbmodel.LastUsedByWorkloadRecord{},

		dbmodel.SyncStatusRec{},

		dbmodel.AssetNameLockRecord{},

		dbmodel.TenantSettingsRec{},
	)

	gormDb := runaidb.DB().Actual().WithContext(context.Background())
	backend_db.DeleteAll(gormDb)
}

func CleanWorkloadRelatedTables() {
	cleanTables(runaidb.DB().Actual().WithContext(context.Background()),
		&dbmodel.WorkloadsAssetsRec{},
		&dbmodel.WorkloadRec{},
	)
}

// ChangeTenantIdForAssets changes tenant id of assets from given value to another. this is needed for testing.
func ChangeTenantIdForAssets(ctx context.Context, fromTenantId int32, toTenantId int32) error {
	db := runaidb.DB().WithContext(ctx)
	_, err := db.Model(dbmodel.AssetMetaRec{}).Where("tenant_id = ?", fromTenantId).Update("tenant_id", toTenantId)
	return err
}

func RestartUpdateCounter() {
	runaidb.DB().Actual().Exec(fmt.Sprintf("ALTER SEQUENCE %s RESTART WITH 1", "assets_metadata_update_count"))
}
