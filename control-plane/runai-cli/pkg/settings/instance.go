/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package settings

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Instance struct {
	options Options
}

type Options struct {
	ConfigPath string
	ConfigFile string
}

var (
	singleInstance *Instance
	lock           = &sync.Mutex{}
	settingsViper  *viper.Viper
)

func GetInstance() *Instance {
	if singleInstance == nil {
		lock.Lock()
		defer lock.Unlock()
		if singleInstance == nil {
			zap.L().Debug("creating Settings single instance now.")
			singleInstance = &Instance{}
		}
	}
	return singleInstance
}
func (li *Instance) Reload() {
	_ = settingsViper.ReadInConfig()
}
func (oi *Instance) Initializing(configPath string) error {
	lock.Lock()
	defer lock.Unlock()

	return oi.initConfiguration(configPath)
}

func (oi *Instance) initConfiguration(configPath string) error {
	configFile := consts.DefaultSettingsFile
	// Store minimal options for compatibility
	oi.options = Options{
		ConfigPath: configPath,
		ConfigFile: configFile,
	}

	zap.L().Debug("using Settings config", zap.String("path", configPath), zap.String("file", configFile))

	configFilePath := filepath.Join(configPath, configFile)
	err := CreateSettingsConfigFile(configPath, configFile)
	if err != nil {
		return fmt.Errorf("failed to check Settings config path %s: %w", configPath, err)
	}

	err = InitSettingsConfig(configFilePath)
	if err != nil {
		return err
	}

	return nil
}

func (oi *Instance) GetConfiguration() *Settings {
	configMap := settingsViper.AllSettings()
	encodedConfig, _ := json.MarshalIndent(configMap, "", " ")
	localConfig := Settings{}
	err := json.Unmarshal(encodedConfig, &localConfig)
	if err != nil {
		return nil
	}
	return &localConfig
}

func (oi *Instance) GetConfigValue(key string) string {
	return settingsViper.GetString(key)
}

func (oi *Instance) SetListFormatAlign(enabled bool) error {
	settingsViper.Set(ListsFormatAlign, enabled)
	return WriteSettings()
}

func (oi *Instance) GetListFormatAlign() bool {
	if !settingsViper.IsSet(ListsFormatAlign) {
		return DefaultListFormatAlign
	}
	return settingsViper.GetBool(ListsFormatAlign)
}

func (oi *Instance) SetWorkloadColumns(columns []string) error {
	settingsViper.Set(ListsWorkloadColumns, columns)
	return WriteSettings()
}

func (oi *Instance) GetWorkloadColumns() []string {
	if !settingsViper.IsSet(ListsWorkloadColumns) {
		return make([]string, 0)
	}
	return settingsViper.GetStringSlice(ListsWorkloadColumns)
}

func (oi *Instance) SetProjectColumns(columns []string) error {
	settingsViper.Set(ListsProjectColumns, columns)
	return WriteSettings()
}

func (oi *Instance) GetProjectColumns() []string {
	if !settingsViper.IsSet(ListsProjectColumns) {
		return make([]string, 0)
	}
	return settingsViper.GetStringSlice(ListsProjectColumns)
}

func (oi *Instance) SetClusterColumns(columns []string) error {
	settingsViper.Set(ListsClusterColumns, columns)
	return WriteSettings()
}

func (oi *Instance) GetClusterColumns() []string {
	if !settingsViper.IsSet(ListsClusterColumns) {
		return make([]string, 0)
	}
	return settingsViper.GetStringSlice(ListsClusterColumns)
}

func (oi *Instance) SetNodeColumns(columns []string) error {
	settingsViper.Set(ListsNodeColumns, columns)
	return WriteSettings()
}

func (oi *Instance) GetNodeColumns() []string {
	if !settingsViper.IsSet(ListsNodeColumns) {
		return make([]string, 0)
	}
	return settingsViper.GetStringSlice(ListsNodeColumns)
}

func (oi *Instance) SetTemplateColumns(columns []string) error {
	settingsViper.Set(ListsTemplateColumns, columns)
	return WriteSettings()
}

func (oi *Instance) GetTemplateColumns() []string {
	if !settingsViper.IsSet(ListsTemplateColumns) {
		return make([]string, 0)
	}
	return settingsViper.GetStringSlice(ListsTemplateColumns)
}

func (oi *Instance) SetEnvironmentColumns(columns []string) error {
	settingsViper.Set(ListsEnvironmentColumns, columns)
	return WriteSettings()
}

func (oi *Instance) GetEnvironmentColumns() []string {
	if !settingsViper.IsSet(ListsEnvironmentColumns) {
		return make([]string, 0)
	}
	return settingsViper.GetStringSlice(ListsEnvironmentColumns)
}

func (oi *Instance) SetComputeColumns(columns []string) error {
	settingsViper.Set(ListsComputeColumns, columns)
	return WriteSettings()
}

func (oi *Instance) GetComputeColumns() []string {
	if !settingsViper.IsSet(ListsComputeColumns) {
		return make([]string, 0)
	}
	return settingsViper.GetStringSlice(ListsComputeColumns)
}

func (oi *Instance) SetDatasourceColumns(columns []string) error {
	settingsViper.Set(ListsDatasourceColumns, columns)
	return WriteSettings()
}

func (oi *Instance) GetDatasourceColumns() []string {
	if !settingsViper.IsSet(ListsDatasourceColumns) {
		return make([]string, 0)
	}
	return settingsViper.GetStringSlice(ListsDatasourceColumns)
}

func (oi *Instance) SetCredentialColumns(columns []string) error {
	settingsViper.Set(ListsCredentialColumns, columns)
	return WriteSettings()
}

func (oi *Instance) GetCredentialColumns() []string {
	if !settingsViper.IsSet(ListsCredentialColumns) {
		return make([]string, 0)
	}
	return settingsViper.GetStringSlice(ListsCredentialColumns)
}

func (oi *Instance) SetDatavolumeColumns(columns []string) error {
	settingsViper.Set(ListsDatavolumeColumns, columns)
	return WriteSettings()
}

func (oi *Instance) GetDatavolumeColumns() []string {
	if !settingsViper.IsSet(ListsDatavolumeColumns) {
		return make([]string, 0)
	}
	return settingsViper.GetStringSlice(ListsDatavolumeColumns)
}

func (oi *Instance) SetWorkloadTypeColumns(columns []string) error {
	settingsViper.Set(ListsWorkloadTypeColumns, columns)
	return WriteSettings()
}

func (oi *Instance) GetWorkloadTypeColumns() []string {
	if !settingsViper.IsSet(ListsWorkloadTypeColumns) {
		return make([]string, 0)
	}
	return settingsViper.GetStringSlice(ListsWorkloadTypeColumns)
}

func (oi *Instance) SetAIApplicationColumns(columns []string) error {
	settingsViper.Set(ListsApplicationColumns, columns)
	return WriteSettings()
}

func (oi *Instance) GetAIApplicationColumns() []string {
	if !settingsViper.IsSet(ListsApplicationColumns) {
		return make([]string, 0)
	}
	return settingsViper.GetStringSlice(ListsApplicationColumns)
}
