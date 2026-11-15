/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package testing

import (
	"context"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

var _ = Describe("legacy template scope cleanup", func() {
	DescribeTable("removes retained templates only within the deleted scope", func(scopeType api.ScopeType) {
		db := runaidb.DB().Actual()
		clusterID, otherClusterID := uuid.NewString(), uuid.NewString()
		var createdIDs []string
		DeferCleanup(func() {
			for _, table := range []struct{ name, idColumn string }{
				{"assets_template_content", "template_asset_id"},
				{"assets_template", "asset_id"},
				{"assets_supported_workload_types", "asset_id"},
				{"assets_environment", "asset_id"},
				{"assets_metadata", "id"},
			} {
				Expect(db.Table(table.name).Where(table.idColumn+" IN (?)", createdIDs).Delete(nil).Error).To(Succeed())
			}
		})

		sharedAssetID := uuid.NewString()
		createdIDs = append(createdIDs, sharedAssetID)
		sharedAsset := dbmodel.AssetMetaRec{AssetMeta: dbmodel.AssetMeta{
			Id: sharedAssetID, Name: sharedAssetID, Kind: api.ASSETKIND_ENVIRONMENT,
			Scope: api.SCOPE_TENANT, TenantId: convutils.Ref(int32(202)),
		}}
		Expect(db.Create(&sharedAsset).Error).To(Succeed())
		Expect(db.Exec("INSERT INTO assets_environment (asset_id) VALUES (?)", sharedAssetID).Error).To(Succeed())

		createTemplate := func(tenantID int32, cluster string, projectID int32, softDeleted bool) string {
			id := uuid.NewString()
			createdIDs = append(createdIDs, id)
			metadata := dbmodel.AssetMetaRec{AssetMeta: dbmodel.AssetMeta{
				Id: id, Name: id, Kind: "workload-template", Scope: api.SCOPE_PROJECT,
				TenantId: convutils.Ref(tenantID), ClusterId: convutils.Ref(cluster), ProjectId: convutils.Ref(projectID),
			}}
			if softDeleted {
				metadata.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
			}
			Expect(db.Create(&metadata).Error).To(Succeed())
			Expect(db.Exec("INSERT INTO assets_template (asset_id, specific_env) VALUES (?, ?)", id, "{}").Error).To(Succeed())
			Expect(db.Exec("INSERT INTO assets_template_content (template_asset_id, contained_asset_id) VALUES (?, ?)", id, sharedAssetID).Error).To(Succeed())
			Expect(db.Create(&dbmodel.WorkloadSupportedTypesRec{WorkloadSupportedTypes: dbmodel.WorkloadSupportedTypes{
				AssetId: id, Training: convutils.Ref(true),
			}}).Error).To(Succeed())
			return id
		}

		// Cross the batch boundary and include a soft-deleted record.
		deletedIDs := make([]string, 0, 102)
		for i := 0; i < 101; i++ {
			deletedIDs = append(deletedIDs, createTemplate(101, clusterID, 1, false))
		}
		deletedIDs = append(deletedIDs, createTemplate(101, clusterID, 1, true))
		otherProject := createTemplate(101, clusterID, 2, false)
		otherCluster := createTemplate(101, otherClusterID, 3, false)
		otherTenant := createTemplate(202, clusterID, 1, false)
		retainedIDs := []string{otherTenant}

		ctx := context.WithValue(context.Background(), runaictx.CfTenantId, "101")
		scopeID := "1"
		switch scopeType {
		case api.SCOPETYPE_PROJECT:
			retainedIDs = append(retainedIDs, otherProject, otherCluster)
		case api.SCOPETYPE_CLUSTER:
			scopeID = clusterID
			deletedIDs = append(deletedIDs, otherProject)
			retainedIDs = append(retainedIDs, otherCluster)
		case api.SCOPETYPE_TENANT:
			scopeID = "101"
			ctx = context.Background() // Tenant deletion carries the target in the event.
			deletedIDs = append(deletedIDs, otherProject, otherCluster)
		}

		Expect(dao.DeleteAssetsByScope(ctx, scopeType, scopeID)).To(Succeed())
		// Repeated deletion is harmless, including scopes that have no legacy templates left.
		Expect(dao.DeleteAssetsByScope(ctx, scopeType, scopeID)).To(Succeed())
		for _, table := range []struct{ name, idColumn string }{
			{"assets_template_content", "template_asset_id"},
			{"assets_template", "asset_id"},
			{"assets_supported_workload_types", "asset_id"},
			{"assets_metadata", "id"},
		} {
			var count int64
			Expect(db.Table(table.name).Where(table.idColumn+" IN (?)", deletedIDs).Count(&count).Error).To(Succeed())
			Expect(count).To(BeZero(), table.name)
			Expect(db.Table(table.name).Where(table.idColumn+" IN (?)", retainedIDs).Count(&count).Error).To(Succeed())
			Expect(count).To(Equal(int64(len(retainedIDs))), table.name)
		}
		for _, table := range []struct{ name, idColumn string }{
			{"assets_environment", "asset_id"},
			{"assets_metadata", "id"},
		} {
			var count int64
			Expect(db.Table(table.name).Where(table.idColumn+" = ?", sharedAssetID).Count(&count).Error).To(Succeed())
			Expect(count).To(Equal(int64(1)), "referenced asset must survive: "+table.name)
		}
	},
		Entry("project", api.SCOPETYPE_PROJECT),
		Entry("cluster", api.SCOPETYPE_CLUSTER),
		Entry("tenant", api.SCOPETYPE_TENANT),
	)
})
