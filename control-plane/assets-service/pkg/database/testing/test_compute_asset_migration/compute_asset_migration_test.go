/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test_compute_asset_migration

import (
	"context"
	"fmt"
	"strings"
	ginkgoTesting "testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/backend_db"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dbmodel"
	dbTesting "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/testing"
	workloadsdb "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/workloads"
	authClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/utils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/integdb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	dbconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/config"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

const (
	uuid_tensorboard_env = "01ef611d-ce9b-4bff-a29f-6ba4bb4224dc"

	uuid_compute_gpu_memory_16g = "3ab6826a-2123-4b34-a686-64cdf47388ac"
	uuid_compute_gpu_memory_20g = "ecbb7a29-5e49-40f3-8145-15c2ffa4f710"
	uuid_compute_gpu_memory_40g = "f8f08816-53d9-47da-b3c7-2f003b175dc4"
	uuid_compute_gpu_memory_80g = "07e99ad0-aa2f-4dd2-9385-e4a4b2ddc207"

	uuid_compute_assets_a = "01234567-89ab-cdef-8765-64cdf47388ac"
	uuid_compute_assets_b = "876564cd-f473-88ac-0123-456789abcdef"

	uuid_templte_a1 = "7a3d9527-3679-48da-9b64-2120f7fd13f1"
	uuid_templte_a2 = "e026cbbb-2d6c-4910-8aff-918b079ade38"
	uuid_templte_a3 = "570df8a3-03d9-4a50-a903-d5572353ba63"
	uuid_templte_a4 = "ab12da03-03d9-4a50-a903-d5572353ba63"

	tenatAId = 1001
	tenatBId = 1002
)

// it's in a separate directory, because we want to run migrations up to a point,
// fill the workloads and assets tables with info,
// then run the new migration to test that it worked.

func TestComputeAssetMigration(t *ginkgoTesting.T) {
	config.MockConfig()
	RegisterFailHandler(Fail)
	config.Init()

	conf, cleanup, setupErr := integdb.SetupPostgresDB(&integdb.IntegdbConfig{Schema: "assets", MigrationDirPath: "file://../../migrations"})
	if setupErr != nil {
		t.Fatalf("failed to set up test postgres: %v", setupErr)
	}
	t.Cleanup(cleanup)
	// tests run against a single-user DB; use the owner role (not runaidb's limited
	// service session) so runtime AutoMigrate can create tables.
	conf.UseServiceSession = new(bool)
	if initErr := runaidb.Init(conf); initErr != nil {
		t.Fatalf("failed to init database: %v", initErr)
	}
	t.Cleanup(func() { _ = runaidb.Close() })

	backend_db.CreateTables(runaidb.DB().Actual())

	authClient.SetMockAuthorizationClient(auth.SimpleMockAuthService{})

	RunSpecs(t, "Test compute migration test")
}

var _ = Describe("Compute System Assets Migration", Ordered, func() {
	ctxTenantA := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenatAId))
	ctxTenantB := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenatBId))

	// This test is flaky in the ci, RC wasn't found, the allow smooth CI add 3 retries
	Context("Compute System Assets Migration", FlakeAttempts(3), func() {

		BeforeEach(func() {
			dbTesting.CleanTables()

			// The assets service can't do the migration flows from zero (migration flow depends on the old backend db)
			// new to preload the table with the 4 system scope assets
			injectComputeAssets(context.Background(), uuid_compute_gpu_memory_16g, "GPU Memory 16GB", string(api.ASSETKIND_COMPUTE), "system")
			injectComputeAssets(context.Background(), uuid_compute_gpu_memory_20g, "GPU Memory 20GB", string(api.ASSETKIND_COMPUTE), "system")
			injectComputeAssets(context.Background(), uuid_compute_gpu_memory_40g, "GPU Memory 40GB", string(api.ASSETKIND_COMPUTE), "system")
			injectComputeAssets(context.Background(), uuid_compute_gpu_memory_80g, "GPU Memory 80GB", string(api.ASSETKIND_COMPUTE), "system")

			// Let's preload it with more compute and environment assets for testing
			injectComputeAssets(ctxTenantA, uuid_compute_assets_a, "compute asset a", string(api.ASSETKIND_COMPUTE), "tenant")
			injectComputeAssets(ctxTenantB, uuid_compute_assets_b, "compute asset b", string(api.ASSETKIND_COMPUTE), "tenant")
			injectEnvironmentAssets(ctxTenantA, uuid_tensorboard_env, "tensorboard", string(api.ASSETKIND_ENVIRONMENT), "tenant")
			injectEnvironmentAssets(ctxTenantB, uuid_tensorboard_env, "tensorboard", string(api.ASSETKIND_ENVIRONMENT), "tenant")
		})

		It("Compute System Assets Migration", func() {
			usedUUIDs := map[string]struct{}{
				uuid_tensorboard_env:        struct{}{},
				uuid_compute_gpu_memory_16g: struct{}{},
				uuid_compute_gpu_memory_20g: struct{}{},
				uuid_compute_gpu_memory_40g: struct{}{},
				uuid_compute_gpu_memory_80g: struct{}{},
				uuid_compute_assets_a:       struct{}{},
				uuid_compute_assets_b:       struct{}{},
				uuid_templte_a1:             struct{}{},
				uuid_templte_a2:             struct{}{},
				uuid_templte_a3:             struct{}{},
			}

			cluster1 := uuid.NewString()

			// inject a tenant A's workload that uses gpu_memory_16g, it will need to be migrated to a new compute asset
			workloadId1, apiErr := injectWorkload(ctxTenantA, "ws1", cluster1, 1, "ns1", uuid_tensorboard_env, uuid_compute_gpu_memory_16g)
			Expect(apiErr).To(BeNil())
			// test that it contains uuid_compute_gpu_memory_16g as a compute asset
			computeId, err := getComputeAssetOfWorkload(ctxTenantA, workloadId1)
			Expect(err).To(BeNil())
			Expect(computeId).To(Equal(uuid_compute_gpu_memory_16g))

			// inject a tenant A's workload that uses simple compute asset shouldn't have any affect
			workloadId2, apiErr := injectWorkload(ctxTenantA, "ws2", cluster1, 1, "ns1", uuid_tensorboard_env, uuid_compute_assets_a)
			Expect(apiErr).To(BeNil())
			computeId, err = getComputeAssetOfWorkload(ctxTenantA, workloadId2)
			Expect(err).To(BeNil())
			Expect(computeId).To(Equal(uuid_compute_assets_a))

			// inject a tenant A's workload that uses gpu_memory_16g, it will need to be migrated to the same as ws1 compute asset
			workloadId3, apiErr := injectWorkload(ctxTenantA, "ws3", cluster1, 1, "ns1", uuid_tensorboard_env, uuid_compute_gpu_memory_16g)
			Expect(apiErr).To(BeNil())
			computeId, err = getComputeAssetOfWorkload(ctxTenantA, workloadId3)
			Expect(err).To(BeNil())
			Expect(computeId).To(Equal(uuid_compute_gpu_memory_16g))

			// inject a tenant A's workload that uses gpu_memory_20g, it will need to be migrated to a new compute asset
			workloadId4, apiErr := injectWorkload(ctxTenantA, "ws4", cluster1, 1, "ns1", uuid_tensorboard_env, uuid_compute_gpu_memory_20g)
			Expect(apiErr).To(BeNil())
			computeId, err = getComputeAssetOfWorkload(ctxTenantA, workloadId4)
			Expect(err).To(BeNil())
			Expect(computeId).To(Equal(uuid_compute_gpu_memory_20g))

			// inject a tenant B's workload that uses gpu_memory_20g, it will need to be migrated to a new compute asset
			workloadId5, apiErr := injectWorkload(ctxTenantB, "ws5", cluster1, 1, "ns1", uuid_tensorboard_env, uuid_compute_gpu_memory_20g)
			Expect(apiErr).To(BeNil())
			computeId, err = getComputeAssetOfWorkload(ctxTenantB, workloadId5)
			Expect(err).To(BeNil())
			Expect(computeId).To(Equal(uuid_compute_gpu_memory_20g))

			// inject a tenant A's workload that uses simple compute asset shouldn't have any affect
			workloadId6, apiErr := injectWorkload(ctxTenantB, "ws6", cluster1, 1, "ns1", uuid_tensorboard_env, uuid_compute_assets_b)
			Expect(apiErr).To(BeNil())
			computeId, err = getComputeAssetOfWorkload(ctxTenantB, workloadId6)
			Expect(err).To(BeNil())
			Expect(computeId).To(Equal(uuid_compute_assets_b))

			// inject templates for on tenant A
			injectTemplateContent(ctxTenantA, uuid_templte_a1, uuid_compute_gpu_memory_20g, "template1")
			injectTemplateContent(ctxTenantA, uuid_templte_a2, uuid_compute_gpu_memory_20g, "template2")
			injectTemplateContent(ctxTenantA, uuid_templte_a3, uuid_compute_assets_a, "template3")
			injectTemplateContent(ctxTenantA, uuid_templte_a4, uuid_compute_gpu_memory_40g, "template4")

			// The assets service can't do the migration flows from zero (migration flow depends on the old backend db)
			// therefore we need to skip the firsts migrations, by setting the migration from 20200002 to 20200003
			migrateTo(20200002, "up 20200003")

			// test that the old compute assets are gone
			computeId, err = getComputeAssetOfWorkload(ctxTenantA, uuid_compute_gpu_memory_16g)
			Expect(err.Error()).To(ContainSubstring("no compute asset found"))
			computeId, err = getComputeAssetOfWorkload(ctxTenantA, uuid_compute_gpu_memory_20g)
			Expect(err.Error()).To(ContainSubstring("no compute asset found"))
			computeId, err = getComputeAssetOfWorkload(ctxTenantA, uuid_compute_gpu_memory_40g)
			Expect(err.Error()).To(ContainSubstring("no compute asset found"))
			computeId, err = getComputeAssetOfWorkload(ctxTenantA, uuid_compute_gpu_memory_80g)
			Expect(err.Error()).To(ContainSubstring("no compute asset found"))

			// Check that workload1 uses a new compute asset that belongs to tenant A
			ws1_compute_asset, err := getComputeAssetOfWorkload(ctxTenantA, workloadId1)
			Expect(err).To(BeNil())
			Expect(ws1_compute_asset).ToNot(Equal(uuid_compute_gpu_memory_16g))
			Expect(containString(usedUUIDs, ws1_compute_asset)).To(BeFalse())
			usedUUIDs[ws1_compute_asset] = struct{}{}

			computeAsset, apiErr := dao.GetComputeAssetById(ctxTenantA, ws1_compute_asset)
			Expect(apiErr).To(BeNil())
			Expect(*computeAsset.Meta.TenantId).To(Equal(int32(tenatAId)))

			// Check that ws2 still uses the same compute asset
			ws2_compute_asset, err := getComputeAssetOfWorkload(ctxTenantA, workloadId2)
			Expect(err).To(BeNil())
			Expect(ws2_compute_asset).To(Equal(uuid_compute_assets_a))
			usedUUIDs[ws2_compute_asset] = struct{}{}

			// Check that ws3 uses the same compute asset as ws1
			ws3_compute_asset, err := getComputeAssetOfWorkload(ctxTenantA, workloadId3)
			Expect(err).To(BeNil())
			Expect(ws3_compute_asset).To(Equal(ws1_compute_asset))

			// Check that ws4 has a different compute asset than gpu_memory_20g and all the rest
			ws4_compute_asset, err := getComputeAssetOfWorkload(ctxTenantA, workloadId4)
			Expect(err).To(BeNil())
			Expect(containString(usedUUIDs, ws4_compute_asset)).To(BeFalse())

			computeAsset, apiErr = dao.GetComputeAssetById(ctxTenantA, ws4_compute_asset)
			Expect(apiErr).To(BeNil())
			Expect(*computeAsset.Meta.TenantId).To(Equal(int32(tenatAId)))
			usedUUIDs[ws4_compute_asset] = struct{}{}

			// Check that ws4 has a different compute asset than gpu_memory_20g and all the rest and belongs to tenant B
			ws5_compute_asset, err := getComputeAssetOfWorkload(ctxTenantA, workloadId5)
			Expect(err).To(BeNil())
			Expect(containString(usedUUIDs, ws5_compute_asset)).To(BeFalse())
			usedUUIDs[ws5_compute_asset] = struct{}{}

			computeAsset, apiErr = dao.GetComputeAssetById(ctxTenantB, ws5_compute_asset)
			Expect(apiErr).To(BeNil())
			Expect(*computeAsset.Meta.TenantId).To(Equal(int32(tenatBId)))

			// Check that ws6 still uses the same compute asset
			ws6_compute_asset, err := getComputeAssetOfWorkload(ctxTenantA, workloadId6)
			Expect(err).To(BeNil())
			Expect(ws6_compute_asset).To(Equal(uuid_compute_assets_b))

			// template 1 should have the same compute asset as ws4
			template1AssetsId, err := getAssetsIdOfTemplateContent(ctxTenantA, uuid_templte_a1)
			Expect(err).To(BeNil())
			Expect(template1AssetsId).To(Equal(ws4_compute_asset))

			// template 2 should have the same compute asset as ws4
			template2AssetsId, err := getAssetsIdOfTemplateContent(ctxTenantA, uuid_templte_a2)
			Expect(err).To(BeNil())
			Expect(template2AssetsId).To(Equal(ws4_compute_asset))

			// template 3 should keep it compute asset
			template3AssetsId, err := getAssetsIdOfTemplateContent(ctxTenantA, uuid_templte_a3)
			Expect(err).To(BeNil())
			Expect(template3AssetsId).To(Equal(uuid_compute_assets_a))

			// template 4 should have a new compute asset
			template4AssetsId, err := getAssetsIdOfTemplateContent(ctxTenantA, uuid_templte_a4)
			Expect(err).To(BeNil())
			Expect(containString(usedUUIDs, template4AssetsId)).To(BeFalse())
			usedUUIDs[template3AssetsId] = struct{}{}

			// What goes up, must come down, testing down migration
			migrateTo(20200003, "down 20200002")

			version, err := getSchemaVersion(context.Background())
			Expect(err).To(BeNil())
			Expect(version).To(Equal(20200002))
		})

	})
})

type SchemaMigration struct {
	Version int  `gorm:"column:version"`
	Dirty   bool `gorm:"column:dirty"`
}

func getSchemaVersion(ctx context.Context) (int, error) {
	var row SchemaMigration

	if err := runaidb.DB().Actual().WithContext(ctx).First(&row).Error; err != nil {
		return 0, err
	}

	return row.Version, nil
}

func setSchemaMigrations(ctx context.Context, version int, dirty bool) error {
	type SchemaMigration struct {
		Version int  `gorm:"column:version"`
		Dirty   bool `gorm:"column:dirty"`
	}

	// delete the migration status from the table
	if err := runaidb.DB().Actual().WithContext(ctx).Exec("DELETE FROM schema_migrations").Error; err != nil {
		return err
	}

	newRow := SchemaMigration{
		Version: version,
		Dirty:   dirty,
	}
	if err := runaidb.DB().Actual().Create(&newRow).Error; err != nil {
		return err
	}

	return nil
}

func injectComputeAssets(ctx context.Context, id string, name string, kind string, scope string) error {
	tenantId := auth.TenantIdOf(ctx)
	var result *gorm.DB
	if tenantId == 0 {
		sql := `
		INSERT INTO assets_metadata (id, name, scope, kind, created_by, created_at, updated_by, updated_at)
		VALUES (?, ?, ?, ?, 'test', NOW(), 'test', NOW());
	`
		result = runaidb.DB().Actual().WithContext(ctx).Exec(sql, id, name, scope, kind)
	} else {
		sql := `
		INSERT INTO assets_metadata (id, name, scope, kind, created_by, created_at, updated_by, updated_at, tenant_id)
		VALUES (?, ?, ?, ?, 'test', NOW(), 'test', NOW(), ?);
	`
		result = runaidb.DB().Actual().WithContext(ctx).Exec(sql, id, name, scope, kind, tenantId)
	}
	if result.Error != nil {
		return fmt.Errorf("failed to insert asset: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no rows affected")
	}

	sql := `INSERT INTO assets_compute (asset_id, gpu_request_type)
	VALUES (?, ?);`
	result = runaidb.DB().Actual().WithContext(ctx).Exec(sql, id, name)
	if result.Error != nil {
		return fmt.Errorf("failed to insert asset: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no rows affected")
	}
	return nil
}

func injectEnvironmentAssets(ctx context.Context, id string, name string, kind string, scope string) error {
	tenantId := auth.TenantIdOf(ctx)
	var result *gorm.DB
	if tenantId == 0 {
		sql := `
		INSERT INTO assets_metadata (id, name, scope, kind, created_by, created_at, updated_by, updated_at)
		VALUES (?, ?, ?, ?, 'test', NOW(), 'test', NOW());
	`
		result = runaidb.DB().Actual().WithContext(ctx).Exec(sql, id, name, scope, kind)
	} else {
		sql := `
		INSERT INTO assets_metadata (id, name, scope, kind, created_by, created_at, updated_by, updated_at, tenant_id)
		VALUES (?, ?, ?, ?, 'test', NOW(), 'test', NOW(), ?);
	`
		result = runaidb.DB().Actual().WithContext(ctx).Exec(sql, id, name, scope, kind, tenantId)
	}
	if result.Error != nil {
		return fmt.Errorf("failed to insert asset: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no rows affected")
	}
	return nil
}

func injectWorkload(ctx context.Context, name string, clusterId string, projectId int32, namespace string, environmentAssetId string, computeAssetId string) (string, *api.Error) {
	// Load the db with workloads
	workload1 := api.WorkloadCreationRequest{
		Name:      name,
		ClusterId: clusterId,
		ProjectId: projectId,
		Namespace: namespace,
		Assets: api.AssetsIds{
			Environment: environmentAssetId,
			Compute:     utils.Ref(computeAssetId),
		},
	}
	containedAssets := []dbmodel.AssetsContainedInWorkload{
		{AssetMetaRec: dbmodel.AssetMetaRec{AssetMeta: dbmodel.AssetMeta{Id: environmentAssetId, Kind: api.ASSETKIND_ENVIRONMENT}}},
		{AssetMetaRec: dbmodel.AssetMetaRec{AssetMeta: dbmodel.AssetMeta{Id: computeAssetId, Kind: api.ASSETKIND_COMPUTE}}},
	}

	id := uuid.NewString()
	_, apiErr := workloadsdb.CreateWorkload(ctx, "Workspace", &workload1, id, containedAssets, nil, nil, nil, nil)
	return id, apiErr
}

func injectTemplateContent(ctx context.Context, template_id string, asset_id string, name string) error {
	var result *gorm.DB
	tenantId := auth.TenantIdOf(ctx)
	sql := `
		INSERT INTO assets_metadata (id, name, scope, kind, created_by, created_at, updated_by, updated_at, tenant_id)
		VALUES (?, ?, 'tenant', ?, 'test', NOW(), 'test', NOW(), ?);
	`
	result = runaidb.DB().Actual().WithContext(ctx).Exec(sql, template_id, name, string(api.ASSETKIND_WORKLOAD_TEMPLATE), tenantId)

	if result.Error != nil {
		return fmt.Errorf("failed to insert asset: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no rows affected")
	}

	// inject to the template content table
	sql = `
		INSERT INTO assets_template_content (template_asset_id, contained_asset_id)
		VALUES (?, ?);
	`
	result = runaidb.DB().Actual().WithContext(ctx).Exec(sql, template_id, asset_id)
	if result.Error != nil {
		return fmt.Errorf("failed to insert asset: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no rows affected")
	}
	return nil
}

type AssetsTemplateContent struct {
	ContainedAssetId string `gorm:"column:contained_asset_id"`
	TemplateAssetId  string `gorm:"column:template_asset_id"`
}

func (AssetsTemplateContent) TableName() string {
	return "assets_template_content"
}
func getAssetsIdOfTemplateContent(ctx context.Context, template_id string) (string, error) {

	var row AssetsTemplateContent

	if err := runaidb.DB().Actual().WithContext(ctx).Where("template_asset_id = ?", template_id).First(&row).Error; err != nil {
		return "", err
	}

	return row.ContainedAssetId, nil
}

func getComputeAssetOfWorkload(ctx context.Context, workloadId string) (string, error) {
	assets, apiErr := workloadsdb.GetWorkloadAssetsByWorkloadId(ctx, []string{"Workspace"}, workloadId)
	Expect(apiErr).To(BeNil())
	Expect(assets).ToNot(BeNil())

	for _, asset := range assets {
		if asset.WorkloadsAssetsRec.AssetKind == api.ASSETKIND_COMPUTE {
			return asset.WorkloadsAssetsRec.AssetId, nil
		}
	}

	return "", fmt.Errorf("no compute asset found")
}

func migrateTo(start int, command string) {
	// To test the 20200003 set the migration status to 20200002, and ran a the command to migrate to 20200003
	migErr := setSchemaMigrations(context.Background(), start, false)
	Expect(migErr).To(BeNil())
	fields := strings.Fields(command)
	params := &runaidb.MigrationParams{}
	switch fields[0] {
	case "up", "upmost":
		params.Direction = dbconfig.MigrationDirectionUp
	case "down", "downmost":
		params.Direction = dbconfig.MigrationDirectionDown
	case "force":
		params.Direction = dbconfig.MigrationDirectionForce
	}
	if len(fields) == 2 {
		params.TargetVersion = &fields[1]
	}
	Expect(runaidb.Migrate(params)).To(BeNil())
}

func containString(m map[string]struct{}, s string) bool {
	_, ok := m[s]
	return ok
}
