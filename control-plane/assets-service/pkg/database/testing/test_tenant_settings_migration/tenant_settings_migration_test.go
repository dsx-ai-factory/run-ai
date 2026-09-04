/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test_tenant_settings_migration

import (
	"context"
	"strings"
	ginkgoTesting "testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/integdb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	dbconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	migrationVersionBefore = 20250002
	migrationVersionThis   = 20250003

	notDirty = false
)

func TestTenantSettingsMigration(t *ginkgoTesting.T) {
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

	RunSpecs(t, "Tenant settings migration test")
}

var _ = Describe("Tenant Settings Migration", Ordered, func() {

	Context("with backend schema containing data", FlakeAttempts(3), func() {

		BeforeEach(func() {
			// Roll back to just before our migration so tenant_settings doesn't exist yet
			migrateTo(migrationVersionThis, "down 20250002")

			createBackendSchema()
			cleanBackendTable()
		})

		AfterEach(func() {
			runaidb.DB().Actual().Exec("DROP TABLE IF EXISTS tenant_settings")
			dropBackendSchema()
		})

		It("full up/down/idempotency sequence", func() {
			// --- Step 1: Seed backend.tenant_settings with initial data ---
			runaidb.DB().Actual().Exec(
				"INSERT INTO backend.tenant_settings (tenant_id, key, value, type) VALUES (1001, 'workspaces.preset_installed', '4', 'Number')")
			runaidb.DB().Actual().Exec(
				"INSERT INTO backend.tenant_settings (tenant_id, key, value, type) VALUES (1002, 'workspaces.preset_installed', '3', 'Number')")
			runaidb.DB().Actual().Exec(
				"INSERT INTO backend.tenant_settings (tenant_id, key, value, type) VALUES (1001, 'some.unrelated.key', 'hello', 'String')")

			// --- Step 2: UP #1 -- verify table created and data copied ---
			migrateTo(migrationVersionBefore, "up 20250003")

			schema := assetsSchemaName()
			Expect(tableExists(schema, "tenant_settings")).To(BeTrue(), "tenant_settings should exist in assets schema")
			Expect(getRevision(schema, 1001)).To(Equal("4"), "tenant 1001 revision should be copied")
			Expect(getRevision(schema, 1002)).To(Equal("3"), "tenant 1002 revision should be copied")
			Expect(keyExists(schema, 1001, "some.unrelated.key")).To(BeFalse(), "unrelated key should NOT be copied")

			// --- Step 3: UP #2 -- test idempotency ---
			migrateTo(migrationVersionBefore, "up 20250003")

			Expect(tableExists(schema, "tenant_settings")).To(BeTrue())
			Expect(getRevision(schema, 1001)).To(Equal("4"))
			Expect(getRevision(schema, 1002)).To(Equal("3"))
			Expect(rowCount("tenant_settings")).To(Equal(int64(2)), "should still have exactly 2 rows")

			// --- Step 4: Add a new row to the OLD backend table ---
			runaidb.DB().Actual().Exec(
				"INSERT INTO backend.tenant_settings (tenant_id, key, value, type) VALUES (1003, 'workspaces.preset_installed', '2', 'Number')")

			// --- Step 5: UP #3 -- verify new data is picked up ---
			migrateTo(migrationVersionBefore, "up 20250003")

			Expect(getRevision(schema, 1003)).To(Equal("2"), "tenant 1003 should be copied on re-run")
			Expect(getRevision(schema, 1001)).To(Equal("4"), "tenant 1001 should be unchanged")
			Expect(getRevision(schema, 1002)).To(Equal("3"), "tenant 1002 should be unchanged")
			Expect(rowCount("tenant_settings")).To(Equal(int64(3)))

			// --- Step 6: Add data to the NEW assets table ---
			runaidb.DB().Actual().Exec(
				"INSERT INTO tenant_settings (tenant_id, key, value, type) VALUES (1004, 'workspaces.preset_installed', '4', 'Number')")
			Expect(rowCount("tenant_settings")).To(Equal(int64(4)))

			// --- Step 7: DOWN #1 -- verify down migration works ---
			migrateTo(migrationVersionThis, "down 20250002")

			Expect(tableExists(schema, "tenant_settings")).To(BeFalse(), "assets tenant_settings should be dropped")

			// All 4 tenants should be in backend
			Expect(getRevision("backend", 1001)).To(Equal("4"))
			Expect(getRevision("backend", 1002)).To(Equal("3"))
			Expect(getRevision("backend", 1003)).To(Equal("2"))
			Expect(getRevision("backend", 1004)).To(Equal("4"), "tenant 1004 should be written back to backend")
			// The unrelated key should still be in backend (untouched)
			Expect(keyExists("backend", 1001, "some.unrelated.key")).To(BeTrue())

			// --- Step 8: DOWN #2 -- test idempotency ---
			migrateTo(migrationVersionThis, "down 20250002")

			Expect(tableExists(schema, "tenant_settings")).To(BeFalse(), "should still be dropped")
			Expect(getRevision("backend", 1001)).To(Equal("4"), "backend data should be unchanged")

			// --- Step 9: UP again -- verify clean round-trip ---
			migrateTo(migrationVersionBefore, "up 20250003")

			Expect(tableExists(schema, "tenant_settings")).To(BeTrue())
			Expect(getRevision(schema, 1001)).To(Equal("4"))
			Expect(getRevision(schema, 1002)).To(Equal("3"))
			Expect(getRevision(schema, 1003)).To(Equal("2"))
			Expect(getRevision(schema, 1004)).To(Equal("4"))
		})
	})

	Context("no backend schema", FlakeAttempts(3), func() {

		BeforeEach(func() {
			migrateTo(migrationVersionThis, "down 20250002")
			dropBackendSchema()
		})

		AfterEach(func() {
			runaidb.DB().Actual().Exec("DROP TABLE IF EXISTS tenant_settings")
		})

		It("should create empty table without errors", func() {
			migrateTo(migrationVersionBefore, "up 20250003")

			schema := assetsSchemaName()
			Expect(tableExists(schema, "tenant_settings")).To(BeTrue())
			Expect(rowCount("tenant_settings")).To(Equal(int64(0)))

			// Insert a row, then down should not fail even without backend schema
			runaidb.DB().Actual().Exec(
				"INSERT INTO tenant_settings (tenant_id, key, value, type) VALUES (9999, 'workspaces.preset_installed', '1', 'Number')")

			migrateTo(migrationVersionThis, "down 20250002")

			Expect(tableExists(schema, "tenant_settings")).To(BeFalse())
		})
	})

	Context("backend table exists but is empty", FlakeAttempts(3), func() {

		BeforeEach(func() {
			migrateTo(migrationVersionThis, "down 20250002")
			createBackendSchema()
			cleanBackendTable()
		})

		AfterEach(func() {
			runaidb.DB().Actual().Exec("DROP TABLE IF EXISTS tenant_settings")
			dropBackendSchema()
		})

		It("should create empty table without errors", func() {
			migrateTo(migrationVersionBefore, "up 20250003")

			schema := assetsSchemaName()
			Expect(tableExists(schema, "tenant_settings")).To(BeTrue())
			Expect(rowCount("tenant_settings")).To(Equal(int64(0)))
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

func tableExists(schema, tableName string) bool {
	var exists bool
	runaidb.DB().Actual().Raw(
		"SELECT to_regclass(? || '.' || ?) IS NOT NULL",
		schema, tableName,
	).Scan(&exists)
	return exists
}

func rowCount(tableName string) int64 {
	var count int64
	runaidb.DB().Actual().Raw("SELECT COUNT(*) FROM " + tableName).Scan(&count)
	return count
}

func getRevision(schema string, tenantId int) string {
	var value string
	runaidb.DB().Actual().Raw(
		"SELECT value FROM "+schema+".tenant_settings WHERE tenant_id = ? AND key = 'workspaces.preset_installed'",
		tenantId,
	).Scan(&value)
	return value
}

func keyExists(schema string, tenantId int, key string) bool {
	var count int64
	runaidb.DB().Actual().Raw(
		"SELECT COUNT(*) FROM "+schema+".tenant_settings WHERE tenant_id = ? AND key = ?",
		tenantId, key,
	).Scan(&count)
	return count > 0
}

func createBackendSchema() {
	runaidb.DB().Actual().Exec("CREATE SCHEMA IF NOT EXISTS backend")
	runaidb.DB().Actual().Exec(`CREATE TABLE IF NOT EXISTS backend.tenant_settings (
		tenant_id integer NOT NULL,
		key text NOT NULL,
		value text,
		type text,
		PRIMARY KEY (tenant_id, key)
	)`)
}

func dropBackendSchema() {
	runaidb.DB().Actual().Exec("DROP TABLE IF EXISTS backend.tenant_settings")
	runaidb.DB().Actual().Exec("DROP SCHEMA IF EXISTS backend")
}

func cleanBackendTable() {
	runaidb.DB().Actual().Exec("DELETE FROM backend.tenant_settings")
}

func assetsSchemaName() string {
	return "assets"
}
