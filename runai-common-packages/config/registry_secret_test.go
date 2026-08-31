/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type NestedConfig struct {
	Inner TestSecretConfig
}

func (t *NestedConfig) SetDefaults() {}

type NestedPtrConfig struct {
	Inner *TestSecretConfig `mapstructure:"inner"`
}

func (t *NestedPtrConfig) SetDefaults() {}

type TestSecretConfig struct {
	SecretVal1 string            `mapstructure:"secret_val1" config:"secret"`
	TestVal1   string            `mapstructure:"test_val1"`
	TestMap    map[string]string `mapstructure:"test_map"`
}

func (t *TestSecretConfig) SetDefaults() {
	t.TestVal1 = "test"
}

func TestGetSecret_WontLoadFromEnv(t *testing.T) {
	t.Setenv("SECRET_VAL1", "my_secret")
	t.Setenv("TEST_VAL1", "val")

	r := newRegistry()
	err := r.Register(&TestSecretConfig{})
	assert.Nil(t, err)
	conf := r.Get(&TestSecretConfig{}).(*TestSecretConfig)
	assert.Equal(t, "val", conf.TestVal1)
	assert.Empty(t, conf.SecretVal1)
}

func TestGetSecret_LoadsFromSecretConf(t *testing.T) {
	secretConf := `
secret_val1: my_secret
`
	t.Setenv("TEST_VAL1", "bla")

	f, err := createTempFile(secretConf)
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	registry := newRegistry()
	registry.SetSecretConfigFile(f.Name())

	err = registry.Register(&TestSecretConfig{})
	assert.Nil(t, err)

	result := registry.Get(&TestSecretConfig{}).(*TestSecretConfig)
	assert.Equal(t, "bla", result.TestVal1)
	assert.Equal(t, "my_secret", result.SecretVal1)
}

func TestGetSecret_ReloadsSecret(t *testing.T) {
	secretConf := `
secret_val1: my_secret
`

	f, err := createTempFile(secretConf)
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	registry := newRegistry()
	registry.SetSecretConfigFile(f.Name())

	err = registry.Register(&TestSecretConfig{})
	assert.Nil(t, err)

	result := registry.Get(&TestSecretConfig{}).(*TestSecretConfig)
	assert.Equal(t, "my_secret", result.SecretVal1)

	change := make(chan struct{}, 1)
	registry.OnConfigChange(func() {
		change <- struct{}{}
	})

	config_change := `
secret_val1: new_secret
`
	rewriteFile(t, f.Name(), config_change)

	<-change
	result = registry.Get(&TestSecretConfig{}).(*TestSecretConfig)

	assert.Equal(t, "new_secret", result.SecretVal1)

}

func TestGetSecret_WithRegularConfToo(t *testing.T) {
	secretConf := `
secret_val1: my_secret
`
	generalConf := `
secret_val1: attacker_secret
test_val1: bla
`

	sf, err := createTempFile(secretConf)
	assert.Nil(t, err)
	defer os.Remove(sf.Name())

	gf, err := createTempFile(generalConf)
	assert.Nil(t, err)
	defer os.Remove(gf.Name())

	registry := newRegistry()
	registry.SetSecretConfigFile(sf.Name())
	registry.SetConfigFiles(gf.Name())

	err = registry.Register(&TestSecretConfig{})
	assert.Nil(t, err)

	result := registry.Get(&TestSecretConfig{}).(*TestSecretConfig)
	assert.Equal(t, "bla", result.TestVal1)
	assert.Equal(t, "my_secret", result.SecretVal1)
}

func TestGetSecret_FilePathTakenFromEnv(t *testing.T) {
	secretConf := `
secret_val1: my_secret
`

	f, err := createTempFile(secretConf)
	assert.Nil(t, err)
	defer os.Remove(f.Name())

	t.Setenv(envSecretsConfigFilePath, f.Name())

	registry := newRegistry()
	registry.SetSecretConfigFile("")

	err = registry.Register(&TestSecretConfig{})
	assert.Nil(t, err)

	result := registry.Get(&TestSecretConfig{}).(*TestSecretConfig)
	assert.Equal(t, "my_secret", result.SecretVal1)
}

func TestGetSecret_WithoutConfig(t *testing.T) {
	registry := newRegistry()
	err := registry.SetSecretConfigFile("")
	require.Equal(t, err.Error(), "config file  was not found")
}

func TestNestedConfig_InnerSecretRemovesValue(t *testing.T) {
	registry := newRegistry()
	registry.FlushAll()
	configFile := `
inner:
    secret_val1: bla
`

	f, err := createTempFile(configFile)
	assert.Nil(t, err)
	err = registry.SetConfigFiles(f.Name())
	assert.Nil(t, err)

	registry.Register(&NestedConfig{})
	config := registry.Get(&NestedConfig{}).(*NestedConfig)

	assert.Equal(t, "", config.Inner.SecretVal1)
}

func TestNestedConfig_InnerSecretSetsValue(t *testing.T) {
	registry := newRegistry()
	registry.FlushAll()
	configFile := `
inner:
    secret_val1: bla
`

	f, err := createTempFile(configFile)
	assert.Nil(t, err)
	err = registry.SetSecretConfigFile(f.Name())
	assert.Nil(t, err)

	registry.Register(&NestedConfig{})
	config := registry.Get(&NestedConfig{}).(*NestedConfig)

	assert.Equal(t, "bla", config.Inner.SecretVal1)
}

func TestSetSecretsFolder_RootLoadsByFileName(t *testing.T) {
	secretContent := `my_secret`
	td, err := os.MkdirTemp(os.TempDir(), "*")
	assert.NoError(t, err)
	fpath := filepath.Join(td, "test_secret.secret_val1")
	_, err = createSecretFile(fpath, secretContent)
	assert.NoError(t, err)
	defer os.Remove(fpath)

	registry := newRegistry()
	err = registry.SetKbSecretsFolder(td)
	require.Nil(t, err)

	registry.RegisterRoot(&TestSecretConfig{})
	result := registry.Get(&TestSecretConfig{}).(*TestSecretConfig)

	assert.Equal(t, "my_secret", result.SecretVal1)
}

func TestSetSecretsFolder_LoadsByFileName(t *testing.T) {
	secretContent := `my_secret`
	td, err := os.MkdirTemp(os.TempDir(), "*")
	fpath := filepath.Join(td, "secret_val1")
	assert.NoError(t, err)
	_, err = createSecretFile(fpath, secretContent)
	assert.NoError(t, err)
	defer os.Remove(fpath)

	registry := newRegistry()
	err = registry.SetKbSecretsFolder(td)
	require.Nil(t, err)

	registry.Register(&TestSecretConfig{})
	result := registry.Get(&TestSecretConfig{}).(*TestSecretConfig)

	assert.Equal(t, "my_secret", result.SecretVal1)
}

func TestSetSecretsFolder_ReloadsChangedSecret(t *testing.T) {
	td := t.TempDir()
	fpath := filepath.Join(td, "secret_val1")
	require.NoError(t, os.WriteFile(fpath, []byte("my_secret"), 0600))

	registry := newRegistry()
	require.NoError(t, registry.SetKbSecretsFolder(td))
	require.NoError(t, registry.Register(&TestSecretConfig{}))
	assert.Equal(t, "my_secret", registry.Get(&TestSecretConfig{}).(*TestSecretConfig).SecretVal1)

	change := make(chan struct{}, 1)
	registry.OnConfigChange(func() {
		select {
		case change <- struct{}{}:
		default:
		}
	})

	require.NoError(t, os.WriteFile(fpath, []byte("new_secret"), 0600))
	waitForConfigChange(t, change)

	assert.Equal(t, "new_secret", registry.Get(&TestSecretConfig{}).(*TestSecretConfig).SecretVal1)
}

func TestSetSecretsFolder_ReloadsKubernetesAtomicUpdate(t *testing.T) {
	td := t.TempDir()
	require.NoError(t, createDataAndSymlinkStructure(td))

	registry := newRegistry()
	require.NoError(t, registry.SetKbSecretsFolder(td))
	require.NoError(t, registry.Register(&TestSecretConfig{}))
	assert.Equal(t, "content_of_SECRET_VAL1", registry.Get(&TestSecretConfig{}).(*TestSecretConfig).SecretVal1)

	change := make(chan struct{}, 1)
	registry.OnConfigChange(func() {
		select {
		case change <- struct{}{}:
		default:
		}
	})

	newDataDir := filepath.Join(td, "..2026_07_23_16_00_00.0000000000")
	require.NoError(t, os.Mkdir(newDataDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(newDataDir, "SECRET_VAL1"), []byte("rotated_secret"), 0600))
	require.NoError(t, os.Symlink(newDataDir, filepath.Join(td, "..data_tmp")))
	require.NoError(t, os.Rename(filepath.Join(td, "..data_tmp"), filepath.Join(td, "..data")))

	waitForConfigChange(t, change)

	assert.Equal(t, "rotated_secret", registry.Get(&TestSecretConfig{}).(*TestSecretConfig).SecretVal1)
}

func TestSetSecretsFolder_FolderPathTakenFromEnv(t *testing.T) {
	secretContent := `my_secret`
	td, err := os.MkdirTemp(os.TempDir(), "*")
	fpath := filepath.Join(td, "secret_val1")
	assert.NoError(t, err)
	_, err = createSecretFile(fpath, secretContent)
	assert.NoError(t, err)
	defer os.Remove(fpath)

	t.Setenv(envKbSecretsFolderPath, td)
	registry := newRegistry()
	err = registry.SetKbSecretsFolder("")
	require.Nil(t, err)

	registry.Register(&TestSecretConfig{})
	result := registry.Get(&TestSecretConfig{}).(*TestSecretConfig)

	assert.Equal(t, "my_secret", result.SecretVal1)
}

func TestSetSecretsFolder_SymlinkFilesAreNotIgnored(t *testing.T) {
	baseDir, err := os.MkdirTemp("", "test_symlinks_with_data_*")
	require.NoError(t, err)
	defer os.RemoveAll(baseDir)

	err = createDataAndSymlinkStructure(baseDir)
	require.NoError(t, err)

	t.Setenv(envKbSecretsFolderPath, baseDir)
	registry := newRegistry()
	err = registry.SetKbSecretsFolder("")
	require.Nil(t, err)

	registry.Register(&TestSecretConfig{})
	conf := registry.Get(&TestSecretConfig{}).(*TestSecretConfig)

	require.Equal(t, conf.SecretVal1, "content_of_SECRET_VAL1")
}

func TestSetOverrides_OverridesConfiguredSecret(t *testing.T) {
	secretConf := "secret_val1: original_secret\n"
	f, err := createTempFile(secretConf)
	require.NoError(t, err)
	defer os.Remove(f.Name())

	registry := newRegistry()
	require.NoError(t, registry.SetSecretConfigFile(f.Name()))
	require.NoError(t, registry.Register(&TestSecretConfig{}))

	result := registry.Get(&TestSecretConfig{}).(*TestSecretConfig)
	assert.Equal(t, "original_secret", result.SecretVal1)

	require.NoError(t, registry.SetOverrides(&TestSecretConfig{SecretVal1: "override_secret"}))

	result = registry.Get(&TestSecretConfig{}).(*TestSecretConfig)
	assert.Equal(t, "override_secret", result.SecretVal1)
}

func TestSetOverrides_NilPtrFieldSetFromOverride(t *testing.T) {
	registry := newRegistry()
	require.NoError(t, registry.Register(&NestedPtrConfig{}))

	result := registry.Get(&NestedPtrConfig{}).(*NestedPtrConfig)
	require.Nil(t, result.Inner)

	// dst.Inner is nil — the override must Set it, not recurse into nil.
	require.NoError(t, registry.SetOverrides(&NestedPtrConfig{
		Inner: &TestSecretConfig{TestVal1: "from_override"},
	}))

	result = registry.Get(&NestedPtrConfig{}).(*NestedPtrConfig)
	require.NotNil(t, result.Inner)
	assert.Equal(t, "from_override", result.Inner.TestVal1)
}

func TestSetOverrides_PartialNestedStructPreservesExistingValues(t *testing.T) {
	f, err := createTempFile("inner:\n  test_val1: original\n")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	registry := newRegistry()
	require.NoError(t, registry.SetConfigFiles(f.Name()))
	require.NoError(t, registry.Register(&NestedConfig{}))

	result := registry.Get(&NestedConfig{}).(*NestedConfig)
	require.Equal(t, "original", result.Inner.TestVal1)

	require.NoError(t, registry.SetOverrides(&NestedConfig{
		Inner: TestSecretConfig{SecretVal1: "new_secret"},
	}))

	result = registry.Get(&NestedConfig{}).(*NestedConfig)
	assert.Equal(t, "original", result.Inner.TestVal1)
	assert.Equal(t, "new_secret", result.Inner.SecretVal1)
}

func TestSetOverrides_AppliesSecretField(t *testing.T) {
	registry := newRegistry()
	require.NoError(t, registry.Register(&TestSecretConfig{}))

	err := registry.SetOverrides(&TestSecretConfig{SecretVal1: "override_secret"})
	require.NoError(t, err)

	result := registry.Get(&TestSecretConfig{}).(*TestSecretConfig)
	assert.Equal(t, "override_secret", result.SecretVal1)
}

func TestSetOverrides_SecretSurvivesReload(t *testing.T) {
	conf := "test_val1: original\n"
	f, err := createTempFile(conf)
	require.NoError(t, err)
	defer os.Remove(f.Name())

	registry := newRegistry()
	require.NoError(t, registry.SetConfigFiles(f.Name()))
	require.NoError(t, registry.Register(&TestSecretConfig{}))
	require.NoError(t, registry.SetOverrides(&TestSecretConfig{SecretVal1: "override_secret"}))

	change := make(chan struct{}, 1)
	registry.OnConfigChange(func() {
		select {
		case change <- struct{}{}:
		default:
		}
	})

	require.NoError(t, os.WriteFile(f.Name(), []byte("test_val1: updated"), 0600))
	waitForConfigChange(t, change)

	result := registry.Get(&TestSecretConfig{}).(*TestSecretConfig)
	assert.Equal(t, "updated", result.TestVal1)
	assert.Equal(t, "override_secret", result.SecretVal1)
}

func TestSetOverrides_SecretFromRegularConfigIgnored(t *testing.T) {
	conf := `test_val1: original`
	f, err := createTempFile(conf)
	require.NoError(t, err)
	defer os.Remove(f.Name())

	registry := newRegistry()
	require.NoError(t, registry.SetConfigFiles(f.Name()))
	require.NoError(t, registry.Register(&TestSecretConfig{}))

	change := make(chan struct{}, 1)
	registry.OnConfigChange(func() {
		select {
		case change <- struct{}{}:
		default:
		}
	})

	require.NoError(t, os.WriteFile(f.Name(), []byte("test_val1: updated\nsecret_val1: updated_secret"), 0600))
	waitForConfigChange(t, change)

	result := registry.Get(&TestSecretConfig{}).(*TestSecretConfig)
	assert.Equal(t, "updated", result.TestVal1)
	assert.Empty(t, result.SecretVal1)
}

func waitForConfigChange(t *testing.T, change <-chan struct{}) {
	t.Helper()
	select {
	case <-change:
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for config reload")
	}
}

func createDataAndSymlinkStructure(baseDir string) error {
	dataDirWithTimestamp := filepath.Join(baseDir, "..2025_01_13_09_41_52.1466878844")
	err := os.MkdirAll(dataDirWithTimestamp, 0755)
	if err != nil {
		return fmt.Errorf("failed to create ..data directory: %w", err)
	}

	fileName := "SECRET_VAL1"
	filePath := filepath.Join(dataDirWithTimestamp, fileName)
	err = os.WriteFile(filePath, []byte(fmt.Sprintf("content_of_%s", fileName)), 0644)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", filePath, err)
	}

	dataDir := filepath.Join(baseDir, "..data")
	err = os.Symlink(dataDirWithTimestamp, dataDir)
	if err != nil {
		return fmt.Errorf("failed to create symlink %s -> %s: %w", dataDirWithTimestamp, dataDirWithTimestamp, err)
	}

	symlinkPath := filepath.Join(baseDir, fileName)
	targetPath := filepath.Join(dataDir, fileName)
	err = os.Symlink(targetPath, symlinkPath)
	if err != nil {
		return fmt.Errorf("failed to create symlink %s -> %s: %w", symlinkPath, targetPath, err)
	}

	return nil
}

func createSecretFile(filePath string, data string) (*os.File, error) {
	tempFile, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}

	if _, err := tempFile.Write([]byte(data)); err != nil {
		return nil, err
	}
	if err := tempFile.Close(); err != nil {
		return nil, err
	}
	return tempFile, nil
}
