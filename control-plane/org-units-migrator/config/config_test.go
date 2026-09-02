/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	dao_utils "github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = Describe("Config tests", Ordered, func() {
	BeforeEach(func() {
		os.Setenv(dao_utils.EnvSecretsMountPath, GinkgoT().TempDir())
	})

	It("Default config", func() {
		config.FlushAll()
		Init()
		configData := config.Get[*ConfigData]()
		Expect(configData.SchemaSource).To(Equal("backend"))
		Expect(configData.SchemaTarget).To(Equal("org_unit_service"))
		Expect(configData.MigrationVersion).To(Equal("v1"))
		Expect(configData.ShouldMarkCompleted).To(BeFalse())
		Expect(configData.TenantsToMigrate).To(BeEmpty())
		Expect(configData.TenantsToNotMigrate).To(BeEmpty())
	})

	It("Custom config", func() {
		config.FlushAll()
		GinkgoT().Setenv("DB_SCHEMA_SOURCE", "custom_source")
		GinkgoT().Setenv("DB_SCHEMA_TARGET", "custom_target")
		GinkgoT().Setenv("DB_MIGRATION_VERSION", "custom_version")
		GinkgoT().Setenv("SHOULD_MARK_COMPLETED", "true")
		GinkgoT().Setenv("TENANTS_TO_MIGRATE", "1,2")
		GinkgoT().Setenv("TENANTS_TO_NOT_MIGRATE", "3,4")
		Init()
		configData := config.Get[*ConfigData]()
		Expect(configData.SchemaSource).To(Equal("custom_source"))
		Expect(configData.SchemaTarget).To(Equal("custom_target"))
		Expect(configData.MigrationVersion).To(Equal("custom_version"))
		Expect(configData.ShouldMarkCompleted).To(BeTrue())
		Expect(configData.TenantsToMigrate).To(Equal([]string{"1", "2"}))
		Expect(configData.TenantsToNotMigrate).To(Equal([]string{"3", "4"}))
	})

	It("DB config", func() {
		config.FlushAll()
		Init()
		Expect(config.Get[*dao_utils.DataBaseConfig]().MigrationCommand).To(Equal("None"))
	})
})
