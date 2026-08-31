/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test_deprecate_chatbot_ui_migration

import (
	"context"
	"fmt"
	"strings"
	ginkgoTesting "testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/config"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/backend_db"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/assets-service/pkg/query_types"
	authClient "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	"github.com/run-ai/runai/control-plane/workloads-and-policies/auth"
	"github.com/run-ai/runai/runai-common-packages/integdb"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/run-ai/runai/runai-common-packages/runaidb"
	dbconfig "github.com/run-ai/runai/runai-common-packages/runaidb/config"
)

const (
	migrationVersionBefore = 20250003
	migrationVersionThis   = 20250004

	notDirty = false

	fictiveTenantId = 2147483647
	tenantAId       = 1001
	tenantBId       = 1002

	chatbotUIEnvId = "715a51c3-93ce-4838-a6c6-4d248e6b9ba0"
)

func TestDeprecateChatbotUIMigration(t *ginkgoTesting.T) {
	config.MockConfig()
	config.Init()
	RegisterFailHandler(Fail)
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

	RunSpecs(t, "Deprecate chatbot-ui migration test")
}

var _ = Describe("Deprecate chatbot-ui Migration", Ordered, func() {
	ctxTenantA := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantAId))
	ctxTenantB := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantBId))

	Context("deprecate chatbot-ui migration", FlakeAttempts(3), func() {

		BeforeEach(func() {
			Expect(runaidb.Migrate(&runaidb.MigrationParams{Direction: dbconfig.MigrationDirectionUp})).To(BeNil())

			// Roll back our migration — this restores chatbot-ui to the fictive tenant
			// The fictive tenant now has: chatbot-ui + huggingface-chat-ui + openwebui
			migrateTo(migrationVersionThis, "down 20250003")

			// Clean only test tenant data (not fictive tenant presets)
			cleanTestTenantData()

			// Reset preset clone tracking so cloning triggers fresh
			dao.ForceTenantCloneForTesting(int32(tenantAId))
			dao.ForceTenantCloneForTesting(int32(tenantBId))
		})

		It("existing tenant keeps chatbot-ui as legacy, new tenant does not get it", func() {
			// Trigger preset cloning for tenant A — should get chatbot-ui (before deprecation)
			_, _ = dao.ListEnvironmentAssets(ctxTenantA, &query_types.QueryParams{})

			// Verify tenant A has chatbot-ui before migration
			Expect(tenantHasEnvironment(ctxTenantA, "chatbot-ui")).To(BeTrue(), "tenant A should have chatbot-ui before migration")
			Expect(tenantHasEnvironment(ctxTenantA, "huggingface-chat-ui")).To(BeTrue(), "tenant A should have huggingface-chat-ui")
			Expect(tenantHasEnvironment(ctxTenantA, "openwebui")).To(BeTrue(), "tenant A should have openwebui")

			// Run the deprecation migration UP
			migrateTo(migrationVersionBefore, "up 20250004")

			// Verify chatbot-ui is removed from fictive tenant
			Expect(fictiveTenantHasChatbotUI()).To(BeFalse(), "chatbot-ui should be removed from fictive tenant")

			// Verify tenant A still has chatbot-ui but with Legacy description
			Expect(tenantHasEnvironment(ctxTenantA, "chatbot-ui")).To(BeTrue(), "tenant A should still have chatbot-ui after migration")
			desc := getEnvironmentDescription(tenantAId, "chatbot-ui")
			Expect(desc).To(HavePrefix("(Legacy)"), "tenant A chatbot-ui should have Legacy prefix")

			// Trigger preset cloning for tenant B (new tenant after migration)
			dao.ForceTenantCloneForTesting(int32(tenantBId))
			_, _ = dao.ListEnvironmentAssets(ctxTenantB, &query_types.QueryParams{})

			// Verify tenant B does NOT have chatbot-ui
			Expect(tenantHasEnvironment(ctxTenantB, "chatbot-ui")).To(BeFalse(), "new tenant B should NOT get chatbot-ui")

			// Verify tenant B still gets the new chat UIs
			Expect(tenantHasEnvironment(ctxTenantB, "huggingface-chat-ui")).To(BeTrue(), "tenant B should have huggingface-chat-ui")
			Expect(tenantHasEnvironment(ctxTenantB, "openwebui")).To(BeTrue(), "tenant B should have openwebui")

			// Run DOWN migration
			migrateTo(migrationVersionThis, "down 20250003")

			// Verify chatbot-ui is restored in fictive tenant
			Expect(fictiveTenantHasChatbotUI()).To(BeTrue(), "chatbot-ui should be restored in fictive tenant after rollback")

			// Verify tenant A description is reverted
			desc = getEnvironmentDescription(tenantAId, "chatbot-ui")
			Expect(desc).ToNot(HavePrefix("(Legacy)"), "tenant A chatbot-ui should not have Legacy prefix after rollback")
		})
	})
})

// --- Test utility functions ---

func migrateTo(start int, command string) {
	migErr := setSchemaMigrations(context.Background(), start, notDirty)
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

func setSchemaMigrations(ctx context.Context, version int, dirty bool) error {
	type SchemaMigration struct {
		Version int  `gorm:"column:version"`
		Dirty   bool `gorm:"column:dirty"`
	}

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

func cleanTestTenantData() {
	// Delete only test tenant data, preserving fictive tenant presets
	for _, tenantId := range []int{tenantAId, tenantBId} {
		runaidb.DB().Actual().Exec("DELETE FROM assets_supported_workload_types WHERE asset_id IN (SELECT id FROM assets_metadata WHERE tenant_id = ?)", tenantId)
		runaidb.DB().Actual().Exec("DELETE FROM assets_environment WHERE asset_id IN (SELECT id FROM assets_metadata WHERE tenant_id = ?)", tenantId)
		runaidb.DB().Actual().Exec("DELETE FROM assets_metadata WHERE tenant_id = ?", tenantId)
	}
	runaidb.DB().Actual().Exec("DELETE FROM tenant_settings WHERE tenant_id IN (?, ?)", tenantAId, tenantBId)
}

func tenantHasEnvironment(ctx context.Context, name string) bool {
	actual, err := dao.ListEnvironmentAssets(ctx, &query_types.QueryParams{})
	if err != nil {
		return false
	}
	for _, env := range actual.Entries {
		if env.Meta.Name == name {
			return true
		}
	}
	return false
}

func fictiveTenantHasChatbotUI() bool {
	var count int64
	runaidb.DB().Actual().Raw(
		"SELECT COUNT(*) FROM assets_metadata WHERE id = ? AND tenant_id = ?",
		chatbotUIEnvId, fictiveTenantId,
	).Scan(&count)
	return count > 0
}

func getEnvironmentDescription(tenantId int, name string) string {
	var description string
	runaidb.DB().Actual().Raw(
		"SELECT COALESCE(description, '') FROM assets_metadata WHERE name = ? AND kind = 'environment' AND tenant_id = ?",
		name, tenantId,
	).Scan(&description)
	return description
}
