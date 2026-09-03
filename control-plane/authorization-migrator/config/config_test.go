/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	dbconfig "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Suite")
}

var _ = Describe("Config", Ordered, func() {
	BeforeEach(func() {
		config.FlushAll()
		setAllRequiredEnvVars()
	})

	Context("when only required values are provided", func() {
		It("applies the default configuration", func() {
			Init()

			configData := config.Get[*ConfigData]()
			Expect(configData.Host).To(Equal("localhost"))
			Expect(configData.Port).To(Equal("5432"))
			Expect(configData.Name).To(Equal("test"))
			Expect(configData.SchemaSourceBackend).To(Equal("backend"))
			Expect(configData.SchemaSourceTenants).To(Equal("tenants_manager"))
			Expect(configData.SchemaTarget).To(Equal("authorization"))
			Expect(configData.MigrationVersion).To(Equal("v1"))
			Expect(configData.User).To(Equal("db-user"))
			Expect(configData.Password).To(Equal("db-password"))
			Expect(configData.SslMode).To(Equal(string(dbconfig.SslModeDisable)))
			Expect(configData.KeycloakHost).To(Equal("http://keycloak-http:80"))
			Expect(configData.KeycloakAdminUser).To(Equal("keycloak-admin-username"))
			Expect(configData.KeycloakAdminPassword).To(Equal("keycloak-admin-password"))
		})
	})

	Context("when custom values are provided", func() {
		It("uses the custom configuration", func() {
			GinkgoT().Setenv("DB_PORT", "1111")
			GinkgoT().Setenv("DB_SCHEMA_SOURCE_BACKEND", "custom_source")
			GinkgoT().Setenv("DB_SCHEMA_TARGET", "custom_target")
			GinkgoT().Setenv("DB_SCHEMA_SOURCE_TENANTS", "custom_tenants")
			GinkgoT().Setenv("DB_SSL_MODE", "allow")
			GinkgoT().Setenv("DB_MIGRATION_VERSION", "custom_version")
			GinkgoT().Setenv("KEYCLOAK_HOST", "http://custom-keycloak:80")

			Init()

			configData := config.Get[*ConfigData]()
			Expect(configData.Host).To(Equal("localhost"))
			Expect(configData.Port).To(Equal("1111"))
			Expect(configData.Name).To(Equal("test"))
			Expect(configData.SchemaSourceBackend).To(Equal("custom_source"))
			Expect(configData.SchemaTarget).To(Equal("custom_target"))
			Expect(configData.SchemaSourceTenants).To(Equal("custom_tenants"))
			Expect(configData.SslMode).To(Equal(string(dbconfig.SslModeAllow)))
			Expect(configData.MigrationVersion).To(Equal("custom_version"))
			Expect(configData.User).To(Equal("db-user"))
			Expect(configData.Password).To(Equal("db-password"))
			Expect(configData.KeycloakHost).To(Equal("http://custom-keycloak:80"))
			Expect(configData.KeycloakAdminUser).To(Equal("keycloak-admin-username"))
			Expect(configData.KeycloakAdminPassword).To(Equal("keycloak-admin-password"))
		})
	})

	Context("when the SSL mode is invalid", func() {
		It("falls back to the disabled mode", func() {
			GinkgoT().Setenv("DB_SSL_MODE", "wrong")

			Init()

			configData := config.Get[*ConfigData]()
			Expect(configData.SslMode).To(Equal(string(dbconfig.SslModeDisable)))
		})
	})
})

func setAllRequiredEnvVars() {
	GinkgoT().Setenv("SECRETS_MOUNT_PATH", "./secrets")
	GinkgoT().Setenv("DB_HOST", "localhost")
	GinkgoT().Setenv("DB_NAME", "test")
}
