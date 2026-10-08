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
	"gorm.io/gorm"
)

const (
	tenantAId = 1001
	tenantBId = 1002
	tenantCId = 1003
	tenantDId = 1004
	tenantEId = 1005
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
			gcrEnvironmentID := "7d098107-131b-456f-a38c-9f4cc983b587"
			jfrogEnvironmentID := "fd09b64d-ef9b-4bb0-a3af-16534020c0b7"
			customSettingsEnvironmentID := "31dd81b6-f53e-46f4-a86c-4868743a05b7"
			customImageEnvironmentID := "9840b53e-c2ce-40d6-a27f-e3cdfa2c25f7"
			gcrImage := "gcr.io/run-ai-demo/" + "quickstart-inference:gpt2-cpu"
			jfrogImage := "runai.jfrog.io/core-llm/" + "quickstart-inference:gpt2-cpu"
			replacementImage := "ghcr.io/huggingface/text-generation-inference:3.3.6-intel-cpu"
			customImage := "example.com/custom/gpt2:latest"
			customEnvironmentVariables := `[{"name":"CUSTOM_VARIABLE","value":"custom-value"}]`
			customConnections := `[{"name":"Custom","internalToolInfo":{"toolType":"custom","containerPort":1234,"connectionType":"ServingPort","servingPortInfo":{"protocol":"http"}}}]`
			replacementConnections := `[{"name":"ServingPort","internalToolInfo":{"toolType":"custom","containerPort":3000,"connectionType":"ServingPort","servingPortInfo":{"protocol":"http"}}}]`

			db := runaidb.DB().Actual().WithContext(context.Background())
			result := db.Exec(`
				INSERT INTO assets_metadata (id, kind, scope, tenant_id, name, created_by, created_at, updated_by, updated_at) VALUES
					(?, 'environment', 'tenant', ?, 'gpt2', 'Run:ai', now(), 'Run:ai', now()),
					(?, 'environment', 'tenant', ?, 'gpt2', 'Run:ai', now(), 'Run:ai', now()),
					(?, 'environment', 'tenant', ?, 'gpt2', 'Run:ai', now(), 'Run:ai', now()),
					(?, 'environment', 'tenant', ?, 'gpt2', 'Run:ai', now(), 'Run:ai', now());`,
				gcrEnvironmentID, tenantAId,
				jfrogEnvironmentID, tenantBId,
				customSettingsEnvironmentID, tenantCId,
				customImageEnvironmentID, tenantDId,
			)
			Expect(result.Error).To(BeNil())

			result = db.Exec(`
				INSERT INTO assets_environment (asset_id, image, command, args, environment_variables, connections) VALUES
					(?, ?, NULL, NULL, '[]'::jsonb, '[]'::jsonb),
					(?, ?, NULL, NULL, '[]'::jsonb, '[]'::jsonb),
					(?, ?, 'custom-command', 'custom-args', ?::jsonb, ?::jsonb),
					(?, ?, NULL, NULL, '[]'::jsonb, '[]'::jsonb);`,
				gcrEnvironmentID, gcrImage,
				jfrogEnvironmentID, jfrogImage,
				customSettingsEnvironmentID, jfrogImage, customEnvironmentVariables, customConnections,
				customImageEnvironmentID, customImage,
			)
			Expect(result.Error).To(BeNil())

			migrateTo(30000001, "up 30000002")

			type environmentSpec struct {
				Image                string
				Command              *string
				Args                 *string
				EnvironmentVariables string
				Connections          string
			}

			for _, environmentID := range []string{gcrEnvironmentID, jfrogEnvironmentID} {
				var actual environmentSpec
				result = db.Raw(`SELECT image, command, args, environment_variables::text AS environment_variables,
					connections::text AS connections FROM assets_environment WHERE asset_id = ?`, environmentID).Scan(&actual)
				Expect(result.Error).To(BeNil())
				Expect(actual.Image).To(Equal(replacementImage))
				Expect(actual.Command).NotTo(BeNil())
				Expect(*actual.Command).To(Equal("text-generation-launcher"))
				Expect(actual.Args).NotTo(BeNil())
				Expect(*actual.Args).To(Equal("--model-id gpt2 --port 3000"))
				Expect(actual.EnvironmentVariables).To(MatchJSON(`[]`))
				Expect(actual.Connections).To(MatchJSON(replacementConnections))
			}

			var actualCustomSettings environmentSpec
			result = db.Raw(`SELECT image, command, args, environment_variables::text AS environment_variables,
				connections::text AS connections FROM assets_environment WHERE asset_id = ?`, customSettingsEnvironmentID).Scan(&actualCustomSettings)
			Expect(result.Error).To(BeNil())
			Expect(actualCustomSettings.Image).To(Equal(replacementImage))
			Expect(actualCustomSettings.Command).NotTo(BeNil())
			Expect(*actualCustomSettings.Command).To(Equal("custom-command"))
			Expect(actualCustomSettings.Args).NotTo(BeNil())
			Expect(*actualCustomSettings.Args).To(Equal("custom-args"))
			Expect(actualCustomSettings.EnvironmentVariables).To(MatchJSON(customEnvironmentVariables))
			Expect(actualCustomSettings.Connections).To(MatchJSON(customConnections))

			var actualCustomImage string
			result = db.Raw("SELECT image FROM assets_environment WHERE asset_id = ?", customImageEnvironmentID).Scan(&actualCustomImage)
			Expect(result.Error).To(BeNil())
			Expect(actualCustomImage).To(Equal(customImage))
		})

		It("replaces the Jupyter TensorBoard image and preserves customizations", func() {
			legacyEnvironmentID := "10a96a7b-e5cc-4dcc-ac38-ef206a06d4b4"
			customCommandEnvironmentID := "89247adf-c729-4550-81aa-fd3e17533046"
			customImageEnvironmentID := "57eb91a4-b44c-499d-92b7-95230a1b542e"
			missingSpecEnvironmentID := "6d9a2313-f447-40af-bd61-56bafa249e52"
			customMissingSpecEnvironmentID := "41153629-3970-47b3-9c9f-c2976cfe20c2"
			legacyImage := "gcr.io/run-ai-demo/" + "jupyter-tensorboard"
			replacementImage := "nvcr.io/nvidia/pytorch:26.09-py3"
			replacementArgs := `-c "tensorboard --logdir \"${LOG_DIR:-./}\" --bind_all & exec jupyter lab --ip=0.0.0.0 --allow-root --ServerApp.token="`
			customImage := "example.com/custom/jupyter-tensorboard:latest"

			db := runaidb.DB().Actual().WithContext(context.Background())
			result := db.Exec(`
				INSERT INTO assets_metadata (id, kind, scope, tenant_id, name, created_by, created_at, updated_by, updated_at) VALUES
					(?, 'environment', 'tenant', ?, 'jupyter-tensorboard', 'Run:ai', now(), 'Run:ai', now()),
					(?, 'environment', 'tenant', ?, 'jupyter-tensorboard', 'Run:ai', now(), 'Run:ai', now()),
					(?, 'environment', 'tenant', ?, 'jupyter-tensorboard', 'Run:ai', now(), 'Run:ai', now()),
					(?, 'environment', 'tenant', ?, 'jupyter-tensorboard', 'Run:ai', now(), 'Run:ai', now()),
					(?, 'environment', 'tenant', ?, 'jupyter-tensorboard', 'Run:ai', now(), 'Run:ai', now());`,
				legacyEnvironmentID, tenantAId,
				customCommandEnvironmentID, tenantBId,
				customImageEnvironmentID, tenantCId,
				missingSpecEnvironmentID, tenantDId,
				customMissingSpecEnvironmentID, tenantEId,
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

			migrateTo(30000002, "up 30000003")

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

			var actualInsertedSpec environmentSpec
			result = db.Raw("SELECT image, command, args FROM assets_environment WHERE asset_id = ?", missingSpecEnvironmentID).Scan(&actualInsertedSpec)
			Expect(result.Error).To(BeNil())
			Expect(actualInsertedSpec.Image).To(Equal(replacementImage))
			Expect(actualInsertedSpec.Command).NotTo(BeNil())
			Expect(*actualInsertedSpec.Command).To(Equal("bash"))
			Expect(actualInsertedSpec.Args).NotTo(BeNil())
			Expect(*actualInsertedSpec.Args).To(Equal(replacementArgs))

			var customMissingSpecCount int64
			result = db.Raw("SELECT COUNT(*) FROM assets_environment WHERE asset_id = ?", customMissingSpecEnvironmentID).Scan(&customMissingSpecCount)
			Expect(result.Error).To(BeNil())
			Expect(customMissingSpecCount).To(BeZero())
		})

		It("replaces every legacy vLLM image without changing custom images", func() {
			gcrVLLMEnvironmentID := "c6a8d1c6-79a0-4d32-a35d-f66db4894871"
			v055VLLMEnvironmentID := "92108029-12d3-49e4-9b49-05d8994693c5"
			v064VLLMEnvironmentID := "9c1af535-8390-41dc-8e84-25ade989fc4e"
			customVLLMEnvironmentID := "79bb39a1-b5b4-45b5-8558-24fa5fc0bd74"
			gcrVLLMImage := "gcr.io/run-ai-demo/" + "llm"
			v055VLLMImage := "runai.jfrog.io/core-llm/" + "runai-vllm:v0.5.5-0.5.0"
			v064VLLMImage := "runai.jfrog.io/core-llm/" + "runai-vllm:v0.6.4-0.10.0"
			customVLLMImage := "example.com/custom/vllm:latest"
			replacementVLLMImage := "vllm/vllm-openai:v0.6.4"
			replacementVLLMArgs := `-c "if [ -n \"$RUNAI_MODEL_NAME\" ]; then exec python3 -m vllm.entrypoints.openai.api_server --model \"$RUNAI_MODEL\" --served-model-name \"$RUNAI_MODEL_NAME\" --host 0.0.0.0 --port 8000; else exec python3 -m vllm.entrypoints.openai.api_server --model \"$RUNAI_MODEL\" --host 0.0.0.0 --port 8000; fi"`

			db := runaidb.DB().Actual().WithContext(context.Background())
			result := db.Exec(`
				INSERT INTO assets_metadata (id, kind, scope, tenant_id, name, created_by, created_at, updated_by, updated_at) VALUES
					(?, 'environment', 'tenant', ?, 'llm-server', 'Run:ai', now(), 'Run:ai', now()),
					(?, 'environment', 'tenant', ?, 'llm-server', 'Run:ai', now(), 'Run:ai', now()),
					(?, 'environment', 'tenant', ?, 'llm-server', 'Run:ai', now(), 'Run:ai', now()),
					(?, 'environment', 'tenant', ?, 'llm-server', 'Run:ai', now(), 'Run:ai', now());`,
				gcrVLLMEnvironmentID, tenantAId,
				v055VLLMEnvironmentID, tenantBId,
				v064VLLMEnvironmentID, tenantCId,
				customVLLMEnvironmentID, tenantDId,
			)
			Expect(result.Error).To(BeNil())

			result = db.Exec(`
				INSERT INTO assets_environment (asset_id, image, command, args) VALUES
					(?, ?, NULL, NULL),
					(?, ?, NULL, NULL),
					(?, ?, NULL, NULL),
					(?, ?, NULL, NULL);`,
				gcrVLLMEnvironmentID, gcrVLLMImage,
				v055VLLMEnvironmentID, v055VLLMImage,
				v064VLLMEnvironmentID, v064VLLMImage,
				customVLLMEnvironmentID, customVLLMImage,
			)
			Expect(result.Error).To(BeNil())

			migrateTo(30000003, "up 30000004")

			type environmentSpec struct {
				Image   string
				Command *string
				Args    *string
			}

			for _, environmentID := range []string{gcrVLLMEnvironmentID, v055VLLMEnvironmentID, v064VLLMEnvironmentID} {
				var actualVLLM environmentSpec
				result = db.Raw("SELECT image, command, args FROM assets_environment WHERE asset_id = ?", environmentID).Scan(&actualVLLM)
				Expect(result.Error).To(BeNil())
				Expect(actualVLLM.Image).To(Equal(replacementVLLMImage))
				Expect(actualVLLM.Command).NotTo(BeNil())
				Expect(*actualVLLM.Command).To(Equal("bash"))
				Expect(actualVLLM.Args).NotTo(BeNil())
				Expect(*actualVLLM.Args).To(Equal(replacementVLLMArgs))
			}

			var actualCustomVLLMImage string
			result = db.Raw("SELECT image FROM assets_environment WHERE asset_id = ?", customVLLMEnvironmentID).Scan(&actualCustomVLLMImage)
			Expect(result.Error).To(BeNil())
			Expect(actualCustomVLLMImage).To(Equal(customVLLMImage))
		})

		It("keeps public-image assets when migrating down and then up again", func() {
			gpt2EnvironmentID := "0acd7192-a3e1-4b31-bd7e-1e0c0d583348"
			jupyterEnvironmentID := "a5ec966f-d449-4c82-b4f2-66362f145fa5"
			vllmEnvironmentID := "54892437-c334-4d51-89e6-c51d948594db"
			expectedImages := map[string]string{
				gpt2EnvironmentID:    "ghcr.io/huggingface/text-generation-inference:3.3.6-intel-cpu",
				jupyterEnvironmentID: "nvcr.io/nvidia/pytorch:26.09-py3",
				vllmEnvironmentID:    "vllm/vllm-openai:v0.6.4",
			}

			db := runaidb.DB().Actual().WithContext(context.Background())
			result := db.Exec(`
				INSERT INTO assets_metadata (id, kind, scope, tenant_id, name, created_by, created_at, updated_by, updated_at) VALUES
					(?, 'environment', 'tenant', ?, 'gpt2', 'Run:ai', now(), 'Run:ai', now()),
					(?, 'environment', 'tenant', ?, 'jupyter-tensorboard', 'Run:ai', now(), 'Run:ai', now()),
					(?, 'environment', 'tenant', ?, 'llm-server', 'Run:ai', now(), 'Run:ai', now());`,
				gpt2EnvironmentID, tenantAId,
				jupyterEnvironmentID, tenantBId,
				vllmEnvironmentID, tenantCId,
			)
			Expect(result.Error).To(BeNil())

			result = db.Exec(`
				INSERT INTO assets_environment (asset_id, image) VALUES
					(?, ?),
					(?, ?),
					(?, ?);`,
				gpt2EnvironmentID, expectedImages[gpt2EnvironmentID],
				jupyterEnvironmentID, expectedImages[jupyterEnvironmentID],
				vllmEnvironmentID, expectedImages[vllmEnvironmentID],
			)
			Expect(result.Error).To(BeNil())

			migrateTo(30000004, "down 30000001")
			assertEnvironmentImages(db, expectedImages)

			migrateTo(30000001, "up 30000004")
			assertEnvironmentImages(db, expectedImages)
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

func assertEnvironmentImages(db *gorm.DB, expectedImages map[string]string) {
	for environmentID, expectedImage := range expectedImages {
		var actualImage string
		result := db.Raw("SELECT image FROM assets_environment WHERE asset_id = ?", environmentID).Scan(&actualImage)
		Expect(result.Error).To(BeNil())
		Expect(actualImage).To(Equal(expectedImage))
	}
}

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
