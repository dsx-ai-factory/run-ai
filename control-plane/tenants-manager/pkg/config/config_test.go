/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/stretchr/testify/assert"
)

func TestInit_Defaults(t *testing.T) {
	secretsDir := initConfigForTesting(t)

	err := Init(secretsDir)
	assert.NoError(t, err)

	generalConf := config.Get[*General]()
	assert.Empty(t, generalConf.ReleaseName)

	serverConf := config.Get[*Server]()
	assert.Equal(t, 8080, serverConf.ServicePort)

	authConf := config.Get[*Auth]()
	assert.Empty(t, authConf.ManagementRealm)
	assert.Nil(t, authConf.ManagementGroups)
	assert.Equal(t, "http://identity-manager:8080", authConf.IdentityServiceUrl)
	assert.Equal(t, "https://localhost.com", authConf.ServiceUrl)

	tenantConf := config.Get[*Tenant]()
	assert.Empty(t, tenantConf.Name)
	assert.Empty(t, tenantConf.Username)
	assert.Empty(t, tenantConf.Password)
	assert.Empty(t, tenantConf.MngName)
	assert.Empty(t, tenantConf.MngUsername)
	assert.Empty(t, tenantConf.MngPassword)
}

func TestInit_CustomTenant_NoDomainTemplate(t *testing.T) {
	secretsDir := initConfigForTesting(t)

	t.Setenv("TENANT_DOMAIN_TEMPLATE", "")

	err := Init(secretsDir)
	assert.Equal(t, err.Error(), "failed to register tenant configuration: failed to validate config struct Tenant, error: Key: 'Tenant.DomainTemplate' Error:Field validation for 'DomainTemplate' failed on the 'required' tag")
}

func TestInit_CustomTenant_NoUsernamePass(t *testing.T) {
	secretsDir := initConfigForTesting(t)

	t.Setenv("TENANT_NAME", "name-only")

	err := Init(secretsDir)
	assert.Contains(t, err.Error(), "Key: 'Tenant.Username' Error:Field validation for 'Username' failed on the 'required_with' tag")
	assert.Contains(t, err.Error(), "Key: 'Tenant.Password' Error:Field validation for 'Password' failed on the 'required_with' tag")
}

func TestInit_CustomTenant_NoMngUsernamePass(t *testing.T) {
	secretsDir := initConfigForTesting(t)

	t.Setenv("TENANT_MNG_NAME", "mng-name-only")

	err := Init(secretsDir)
	assert.Contains(t, err.Error(), "Key: 'Tenant.MngUsername' Error:Field validation for 'MngUsername' failed on the 'required_with' tag")
	assert.Contains(t, err.Error(), "Key: 'Tenant.MngPassword' Error:Field validation for 'MngPassword' failed on the 'required_with' tag")
}

func TestInit_CustomTenant(t *testing.T) {
	secretsDir := initConfigForTesting(t)

	t.Setenv("TENANT_DOMAIN_TEMPLATE", "template")
	t.Setenv("TENANT_NAME", "name")
	t.Setenv("ADMIN_USERNAME", "user")
	t.Setenv("TENANT_MNG_NAME", "mng-name")

	secrets := map[string]string{
		"ADMIN_PASSWORD":      "pass",
		"TENANT_MNG_USER":     "mng-user",
		"TENANT_MNG_PASSWORD": "mng-pass",
	}
	err := writeSecretsToDir(secretsDir, secrets)
	assert.NoError(t, err)

	err = Init(secretsDir)
	assert.NoError(t, err)

	tenantConf := config.Get[*Tenant]()
	assert.Equal(t, "template", tenantConf.DomainTemplate)
	assert.Equal(t, "name", tenantConf.Name)
	assert.Equal(t, "user", tenantConf.Username)
	assert.Equal(t, "pass", tenantConf.Password)
	assert.Equal(t, "mng-name", tenantConf.MngName)
	assert.Equal(t, "mng-user", tenantConf.MngUsername)
	assert.Equal(t, "mng-pass", tenantConf.MngPassword)
}

func TestInit_CustomGeneral(t *testing.T) {
	secretsDir := initConfigForTesting(t)

	t.Setenv("RELEASE_NAME", "name")

	err := Init(secretsDir)
	assert.NoError(t, err)

	generalConf := config.Get[*General]()
	assert.Equal(t, "name", generalConf.ReleaseName)
}

func TestInit_CustomServer(t *testing.T) {
	secretsDir := initConfigForTesting(t)

	t.Setenv("SERVICE_PORT", "1234")

	err := Init(secretsDir)
	assert.NoError(t, err)

	serverConf := config.Get[*Server]()
	assert.Equal(t, 1234, serverConf.ServicePort)
}

func TestInit_CustomAuth(t *testing.T) {
	secretsDir := initConfigForTesting(t)

	t.Setenv("IDENTITY_MNG_URL", "mng-url")
	t.Setenv("AUTH_SERVICE_URL", "service-url")

	err := Init(secretsDir)
	assert.NoError(t, err)

	authConf := config.Get[*Auth]()
	assert.Equal(t, "mng-url", authConf.IdentityServiceUrl)
	assert.Equal(t, "service-url", authConf.ServiceUrl)
}

func createSecretsDir(secrets map[string]string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "secrets")
	if err != nil {
		return "", err
	}

	err = writeSecretsToDir(tmpDir, secrets)
	if err != nil {
		return "", err
	}
	return tmpDir, nil
}

func writeSecretsToDir(secretDir string, secrets map[string]string) error {
	for key, value := range secrets {
		secretPath := filepath.Join(secretDir, key)
		if err := os.WriteFile(secretPath, []byte(value), 0o600); err != nil {
			return err
		}
	}
	return nil
}

func initConfigForTesting(t *testing.T) string {
	config.FlushAll()

	t.Setenv("TENANT_DOMAIN_TEMPLATE", "template")

	secretDir, err := createSecretsDir(map[string]string{})
	assert.NoError(t, err)

	return secretDir
}
