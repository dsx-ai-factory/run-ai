/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test_environment_assets_migration

import (
	"context"
	"fmt"
	"strings"
	ginkgoTesting "testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/backend_db"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/dao"
	dbTesting "github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/database/testing"
	"github.com/dsx-ai-factory/run-ai/control-plane/assets-service/pkg/query_types"
	authClient "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/auth"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/integdb"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	dbconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	tenantAId = 1001
	tenantBId = 1002
	tenantCId = 1003
)

func TestEnvironmentAssetsMigration(t *ginkgoTesting.T) {
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

	RunSpecs(t, "Test environment migration test")
}

var _ = Describe("Environment Assets Migration", Ordered, func() {
	ctxTenantA := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantAId))
	ctxTenantB := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantBId))
	ctxTenantC := context.WithValue(context.Background(), runaictx.CfTenantId, fmt.Sprintf("%d", tenantCId))

	// This test is flaky in the ci, RC wasn't found, the allow smooth CI add 3 retries
	// it maybe related to the fact the test is doing migration on the same db instance
	Context("environment assets migration", FlakeAttempts(3), func() {

		BeforeEach(func() {
			// In some runs the test can't access the schema_migrations table, making it flaky,
			// Calling the migrate function again to attempt mitigating the flakiness of the test
			// see ticket RUN-25205 for more info
			Expect(runaidb.Migrate(&runaidb.MigrationParams{Direction: dbconfig.MigrationDirectionUp})).To(BeNil())

			dbTesting.CleanTables()
		})

		It("replaces the legacy GPT-2 image without changing customized environments", func() {
			legacyEnvironmentID := "7d098107-131b-456f-a38c-9f4cc983b587"
			customEnvironmentID := "fd09b64d-ef9b-4bb0-a3af-16534020c0b7"
			legacyImage := "runai.jfrog.io/core-llm/" + "quickstart-inference:gpt2-cpu"
			replacementImage := "ghcr.io/huggingface/text-generation-inference:3.3.6-intel-cpu"
			customImage := "example.com/custom/gpt2:latest"

			db := runaidb.DB().Actual().WithContext(context.Background())
			result := db.Exec(`
				INSERT INTO assets_metadata (id, kind, scope, tenant_id, name, created_by, created_at, updated_by, updated_at) VALUES
					(?, 'environment', 'tenant', ?, 'gpt2', 'Run:ai', now(), 'Run:ai', now()),
					(?, 'environment', 'tenant', ?, 'gpt2', 'Run:ai', now(), 'Run:ai', now());`,
				legacyEnvironmentID, tenantAId,
				customEnvironmentID, tenantBId,
			)
			Expect(result.Error).To(BeNil())

			result = db.Exec(`
				INSERT INTO assets_environment (asset_id, image) VALUES
					(?, ?),
					(?, ?);`,
				legacyEnvironmentID, legacyImage,
				customEnvironmentID, customImage,
			)
			Expect(result.Error).To(BeNil())

			migrateTo(20250005, "up 20250006")

			var actualLegacyImage string
			result = db.Raw("SELECT image FROM assets_environment WHERE asset_id = ?", legacyEnvironmentID).Scan(&actualLegacyImage)
			Expect(result.Error).To(BeNil())
			Expect(actualLegacyImage).To(Equal(replacementImage))

			var actualLegacyCommand string
			result = db.Raw("SELECT command FROM assets_environment WHERE asset_id = ?", legacyEnvironmentID).Scan(&actualLegacyCommand)
			Expect(result.Error).To(BeNil())
			Expect(actualLegacyCommand).To(Equal("text-generation-launcher"))

			var actualCustomImage string
			result = db.Raw("SELECT image FROM assets_environment WHERE asset_id = ?", customEnvironmentID).Scan(&actualCustomImage)
			Expect(result.Error).To(BeNil())
			Expect(actualCustomImage).To(Equal(customImage))
		})

		It("replaces the Jupyter TensorBoard image and preserves customizations", func() {
			legacyEnvironmentID := "10a96a7b-e5cc-4dcc-ac38-ef206a06d4b4"
			customCommandEnvironmentID := "89247adf-c729-4550-81aa-fd3e17533046"
			customImageEnvironmentID := "57eb91a4-b44c-499d-92b7-95230a1b542e"
			legacyImage := "gcr.io/run-ai-demo/" + "jupyter-tensorboard"
			replacementImage := "nvcr.io/nvidia/pytorch:26.09-py3"
			replacementArgs := `-c "tensorboard --logdir \"${LOG_DIR:-./}\" --bind_all & exec jupyter lab --ip=0.0.0.0 --allow-root --ServerApp.token="`
			customImage := "example.com/custom/jupyter-tensorboard:latest"

			db := runaidb.DB().Actual().WithContext(context.Background())
			result := db.Exec(`
				INSERT INTO assets_metadata (id, kind, scope, tenant_id, name, created_by, created_at, updated_by, updated_at) VALUES
					(?, 'environment', 'tenant', ?, 'jupyter-tensorboard', 'Run:ai', now(), 'Run:ai', now()),
					(?, 'environment', 'tenant', ?, 'jupyter-tensorboard', 'Run:ai', now(), 'Run:ai', now()),
					(?, 'environment', 'tenant', ?, 'jupyter-tensorboard', 'Run:ai', now(), 'Run:ai', now());`,
				legacyEnvironmentID, tenantAId,
				customCommandEnvironmentID, tenantBId,
				customImageEnvironmentID, tenantCId,
			)
			Expect(result.Error).To(BeNil())

			result = db.Exec(`
				INSERT INTO assets_environment (asset_id, image, command, args) VALUES
					(?, ?, NULL, NULL),
					(?, ?, 'custom-command', 'custom-args'),
					(?, ?, NULL, NULL);`,
				legacyEnvironmentID, legacyImage,
				customCommandEnvironmentID, legacyImage,
				customImageEnvironmentID, customImage,
			)
			Expect(result.Error).To(BeNil())

			migrateTo(20250006, "up 20250007")

			type environmentSpec struct {
				Image   string
				Command *string
				Args    *string
			}

			var actualLegacy environmentSpec
			result = db.Raw("SELECT image, command, args FROM assets_environment WHERE asset_id = ?", legacyEnvironmentID).Scan(&actualLegacy)
			Expect(result.Error).To(BeNil())
			Expect(actualLegacy.Image).To(Equal(replacementImage))
			Expect(actualLegacy.Command).NotTo(BeNil())
			Expect(*actualLegacy.Command).To(Equal("bash"))
			Expect(actualLegacy.Args).NotTo(BeNil())
			Expect(*actualLegacy.Args).To(Equal(replacementArgs))

			var actualCustomCommand environmentSpec
			result = db.Raw("SELECT image, command, args FROM assets_environment WHERE asset_id = ?", customCommandEnvironmentID).Scan(&actualCustomCommand)
			Expect(result.Error).To(BeNil())
			Expect(actualCustomCommand.Image).To(Equal(replacementImage))
			Expect(actualCustomCommand.Command).NotTo(BeNil())
			Expect(*actualCustomCommand.Command).To(Equal("custom-command"))
			Expect(actualCustomCommand.Args).NotTo(BeNil())
			Expect(*actualCustomCommand.Args).To(Equal("custom-args"))

			var actualCustomImage environmentSpec
			result = db.Raw("SELECT image, command, args FROM assets_environment WHERE asset_id = ?", customImageEnvironmentID).Scan(&actualCustomImage)
			Expect(result.Error).To(BeNil())
			Expect(actualCustomImage.Image).To(Equal(customImage))
			Expect(actualCustomImage.Command).To(BeNil())
			Expect(actualCustomImage.Args).To(BeNil())
		})

		It("environment assets migration", func() {
			// inject the preset environments and verify they are exists
			injectPresetEnvironments()
			// list the environments for tenant A, the preset assets should be propagated to tenant A
			_, _ = dao.ListEnvironmentAssets(ctxTenantA, &query_types.QueryParams{})
			expectedEnvironments(ctxTenantA, []string{"tensorboard", "jupyter-lab"})

			// Migrate to the new version, the  new assets will be added
			migrateTo(20210007, "up 20210008")

			// Listing environment for tenant B, should have all 6 environments
			_, _ = dao.ListEnvironmentAssets(ctxTenantB, &query_types.QueryParams{})
			expectedEnvironments(ctxTenantB, []string{"tensorboard", "jupyter-lab", "tensorboard-tensorflow", "jupyter-scipy", "rstudio", "vscode"})

			// Down the migration, the new preset will be removed
			migrateTo(20210008, "down 20210007")

			_, _ = dao.ListEnvironmentAssets(ctxTenantC, &query_types.QueryParams{})
			expectedEnvironments(ctxTenantC, []string{"tensorboard", "jupyter-lab"})

		})
	})
})

func injectPresetEnvironments() {
	db := runaidb.DB().Actual().WithContext(context.Background())
	cmds := []string{
		`        INSERT INTO assets_metadata ("id", "kind", "scope", "tenant_id", "name", "created_by", "created_at", "updated_by", "updated_at") VALUES
            ('01ef611d-ce9b-4bff-a29f-6ba4bb4224dc', 'environment', 'tenant', '2147483647', 'tensorboard', 'test', '2025-01-01 15:00:00+00', 'test', '2025-01-01 15:00:00+00'),
            ('102d4a7b-2345-47d4-bcc1-52347eef1b4e', 'environment', 'tenant', '2147483647', 'jupyter-lab',     'test', '2025-01-01 15:00:00+00', 'test', '2025-01-01 15:00:00+00'); `,
		`INSERT INTO assets_environment ("asset_id", "command", "args", "image", "image_pull_policy", "connections", "uid_gid_source", "routing_method") VALUES
			('01ef611d-ce9b-4bff-a29f-6ba4bb4224dc', 'tensorboard', '--logdir ./ --host 0.0.0.0', 'tensorflow/tensorflow:2.20.0', 'IfNotPresent', '[{"name": "TensorBoard", "internalToolInfo": {"toolType": "tensorboard", "containerPort": 6006, "connectionType": "ExternalUrl", "externalUrlInfo": {}}}]',  'fromTheImage', ''),
			('102d4a7b-2345-47d4-bcc1-52347eef1b4e', 'start-notebook.sh', '--NotebookApp.token=''', 'jupyter/scipy-notebook', 'IfNotPresent', '[{"name": "Jupyter", "internalToolInfo": {"toolType": "jupyter-notebook", "containerPort": 8888, "connectionType": "ExternalUrl", "externalUrlInfo": {}}}]', 'fromTheImage', '');`,
	}
	for _, cmd := range cmds {
		result := db.Exec(cmd)
		Expect(result.Error).To(BeNil(), cmd)
	}
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

func migrateTo(start int, command string) {
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

func expectedEnvironments(ctx context.Context, expected []string) {
	actual, err := dao.ListEnvironmentAssets(ctx, &query_types.QueryParams{})
	Expect(err).To(BeNil())
	actualNames := make([]string, 0, len(actual.Entries))
	for _, env := range actual.Entries {
		actualNames = append(actualNames, env.Meta.Name)
	}
	Expect(actualNames).To(ConsistOf(expected))
}
