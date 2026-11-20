/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package displayer

import "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/ui/environment"

func GetEnvironmentVariablesDisplayForMap(environmentVariables map[string]string, columnPadding int) string {
	if len(environmentVariables) == 0 {
		return ""
	}
	return GetEnvironmentVariablesDisplay(&environmentVariables, columnPadding)
}

func GetEnvironmentVariablesDisplay(environmentVariables *map[string]string, columnPadding int) string {
	if environmentVariables == nil {
		return "[]"
	}
	return environment.Format(*environmentVariables, columnPadding, MaxDisplayEnvVarsCount)
}
