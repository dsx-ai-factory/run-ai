/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package installconfig

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeConfig(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o600))
	return path
}

func TestLoad_ValidConfig(t *testing.T) {
	path := writeConfig(t, `
install:
  mode: full_installation
  connectivity: connected
registry:
  type: jfrog
  tokenPath: /etc/runai/jfrog.token
tls:
  certPath: /etc/runai/tls.crt
  keyPath: /etc/runai/tls.key
  caChainPath: /etc/runai/ca.crt
controlPlane:
  enabled: true
  version: 2.26.0
  domain: runai.example.com
cluster:
  enabled: true
  name: cluster01
prerequisites:
  skip: false
  gpuOperator:
    mode: install
    version: 24.9.1
    customHelmValues:
      driver.enabled: false
  ingress:
    mode: install
    class: nginx
`)

	cfg, err := Load(path)
	require.NoError(t, err)

	assert.Equal(t, "full_installation", cfg.Install.Mode)
	require.NotNil(t, cfg.Registry)
	assert.Equal(t, "jfrog", cfg.Registry.Type)
	assert.Equal(t, "/etc/runai/jfrog.token", cfg.Registry.TokenPath)
	require.NotNil(t, cfg.TLS)
	assert.Equal(t, "/etc/runai/tls.crt", cfg.TLS.CertPath)
	assert.Equal(t, "/etc/runai/tls.key", cfg.TLS.KeyPath)
	assert.Equal(t, "/etc/runai/ca.crt", cfg.TLS.CAChainPath)
	require.NotNil(t, cfg.ControlPlane)
	assert.True(t, cfg.ControlPlane.Enabled)
	assert.Equal(t, "runai.example.com", cfg.ControlPlane.Domain)
	require.NotNil(t, cfg.Cluster)
	assert.Equal(t, "cluster01", cfg.Cluster.Name)
	require.NotNil(t, cfg.Prerequisites)
	require.NotNil(t, cfg.Prerequisites.GPUOperator)
	assert.Equal(t, "24.9.1", cfg.Prerequisites.GPUOperator.Version)
}

func TestLoad_UnknownFieldRejected(t *testing.T) {
	path := writeConfig(t, `
install:
  mode: full_installation
  bogusField: nope
`)

	_, err := Load(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bogusField")
}

func TestLoad_WrongTypeRejected(t *testing.T) {
	path := writeConfig(t, `
install:
  mode: full_installation
controlPlane:
  enabled: "not-a-bool"
  domain: runai.example.com
`)

	_, err := Load(path)
	require.Error(t, err)
}

func TestLoad_SelfSignedTLSOnlyAllowedAtRoot(t *testing.T) {
	path := writeConfig(t, `
install:
  mode: full_installation
tls:
  generateSelfSigned: true
controlPlane:
  enabled: true
  domain: runai.example.com
  tls:
    generateSelfSigned: true
`)

	_, err := Load(path)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "generateSelfSigned")
}

func TestLoad_SelfSignedTLS(t *testing.T) {
	path := writeConfig(t, `
install:
  mode: runai_control_plane_only
tls:
  generateSelfSigned: true
controlPlane:
  enabled: true
  domain: runai.example.com
`)

	cfg, err := Load(path)
	require.NoError(t, err)
	require.NotNil(t, cfg.TLS)
	assert.True(t, cfg.TLS.GenerateSelfSigned)
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "does-not-exist.yaml"))
	require.Error(t, err)
}

func TestLoad_ControlPlaneCredentials(t *testing.T) {
	path := writeConfig(t, `
install:
  mode: runai_cluster_only
  connectivity: connected
registry:
  type: ngc
cluster:
  enabled: true
  name: cluster01
  domain: cluster.example.com
  controlPlaneUrl: https://cp.example.com
  controlPlaneCredentials:
    username: admin@example.com
    passwordPath: /etc/runai/cp-password
`)

	cfg, err := Load(path)
	require.NoError(t, err)

	require.NotNil(t, cfg.Cluster.ControlPlaneCredentials)
	assert.Equal(t, "admin@example.com", cfg.Cluster.ControlPlaneCredentials.Username)
	assert.Equal(t, "/etc/runai/cp-password", cfg.Cluster.ControlPlaneCredentials.PasswordPath)
	assert.Empty(t, cfg.Cluster.UID)
}
