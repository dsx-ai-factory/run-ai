/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	listCommon "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
)

var (
	instance = config.GetInstance()
)

func SetOutput(output *listCommon.Output) error {
	outputDefault := instance.GetConfigValue(config.SettingsOutputType)
	return output.SetDefault(outputDefault)
}

func SetProject(project *string) {
	if *project == "" {
		*project = instance.GetConfigValue(config.ClusterProjectName)
	}
}
