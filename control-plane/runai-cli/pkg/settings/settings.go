/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package settings

import (
	runaios "github.com/run-ai/runai/control-plane/runai-cli/internal/os"
	"github.com/spf13/viper"
)

const (
	outputSection        = "output"
	ListsSection         = "output.lists"
	ListsFormatAlign     = "output.lists.format_align"
	ListsWorkload        = "output.lists.workload"
	ListsWorkloadColumns = "output.lists.workload.columns"

	ListsProject        = "output.lists.project"
	ListsProjectColumns = "output.lists.project.columns"

	ListsCluster        = "output.lists.cluster"
	ListsClusterColumns = "output.lists.cluster.columns"

	ListsNode        = "output.lists.node"
	ListsNodeColumns = "output.lists.node.columns"

	ListsTemplate        = "output.lists.template"
	ListsTemplateColumns = "output.lists.template.columns"

	ListsEnvironment        = "output.lists.environment"
	ListsEnvironmentColumns = "output.lists.environment.columns"

	ListsCompute        = "output.lists.compute"
	ListsComputeColumns = "output.lists.compute.columns"

	ListsDatasource        = "output.lists.datasource"
	ListsDatasourceColumns = "output.lists.datasource.columns"

	ListsCredential        = "output.lists.credential"
	ListsCredentialColumns = "output.lists.credential.columns"

	ListsDatavolume        = "output.lists.datavolume"
	ListsDatavolumeColumns = "output.lists.datavolume.columns"

	ListsWorkloadType        = "output.lists.workloadtype"
	ListsWorkloadTypeColumns = "output.lists.workloadtype.columns"

	ListsApplication        = "output.lists.application"
	ListsApplicationColumns = "output.lists.application.columns"
)

// Settings default values

const DefaultListFormatAlign = false

type Lists struct {
	FormatAlign bool     `json:"format_align" yaml:"format_align"`
	Workload    Workload `json:"workload,omitempty" yaml:"workload,omitempty"`

	Project Project `json:"project,omitempty" yaml:"project,omitempty"`

	Cluster Cluster `json:"cluster,omitempty" yaml:"cluster,omitempty"`

	Node Node `json:"node,omitempty" yaml:"node,omitempty"`

	Template Template `json:"template,omitempty" yaml:"template,omitempty"`

	Compute Compute `json:"compute,omitempty" yaml:"compute,omitempty"`
}

type Workload struct {
	Columns []string `json:"columns,omitempty" yaml:"columns,omitempty"`
}

type Project struct {
	Columns []string `json:"columns,omitempty" yaml:"columns,omitempty"`
}

type Cluster struct {
	Columns []string `json:"columns,omitempty" yaml:"columns,omitempty"`
}

type Node struct {
	Columns []string `json:"columns,omitempty" yaml:"columns,omitempty"`
}

type Template struct {
	Columns []string `json:"columns,omitempty" yaml:"columns,omitempty"`
}

type Compute struct {
	Columns []string `json:"columns,omitempty" yaml:"columns,omitempty"`
}

type Environment struct {
	Columns []string `json:"columns,omitempty" yaml:"columns,omitempty"`
}

type Datavolume struct {
	Columns []string `json:"columns,omitempty" yaml:"columns,omitempty"`
}

type Output struct {
	Lists Lists `json:"lists" yaml:"lists"`
}
type Settings struct {
	Output Output `json:"output" yaml:"output"`
}

var settingsStringMap = map[string]any{
	ListsFormatAlign:     DefaultListFormatAlign,
	ListsWorkloadColumns: make([]string, 0),

	ListsProjectColumns: make([]string, 0),

	ListsClusterColumns: make([]string, 0),

	ListsNodeColumns: make([]string, 0),

	ListsTemplateColumns: make([]string, 0),

	ListsEnvironmentColumns: make([]string, 0),

	ListsComputeColumns: make([]string, 0),

	ListsDatasourceColumns: make([]string, 0),

	ListsDatavolumeColumns: make([]string, 0),

	ListsCredentialColumns: make([]string, 0),

	ListsWorkloadTypeColumns: make([]string, 0),

	ListsApplicationColumns: make([]string, 0),
}

func init() {
	settingsViper = viper.New() // Separate instance!
}

func InitSettingsConfig(outputFilePath string) error {
	settingsViper.SetConfigFile(outputFilePath)
	err := settingsViper.ReadInConfig()
	if err != nil {
		return err
	}

	defaultSettingsSet()
	return err
}

func defaultSettingsSet() {
	for key, value := range settingsStringMap {
		settingsViper.SetDefault(key, value)
	}
}

func WriteSettings() error {
	return settingsViper.WriteConfig()
}

// CreateSettingsConfigFile creates the settings config file using the config instance's path
func CreateSettingsConfigFile(configPath string, configFile string) error {
	return runaios.CreateConfigFileWithViper(configPath, configFile, settingsViper)
}
