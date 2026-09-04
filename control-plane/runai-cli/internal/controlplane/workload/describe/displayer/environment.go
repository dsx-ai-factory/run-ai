/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package displayer

import (
	"fmt"
	"slices"
	"strings"

	"github.com/samber/lo"
)

func GetEnvironmentVariablesDisplayForMap(environmentVariables map[string]string, columnPadding int) string {
	if len(environmentVariables) == 0 {
		return ""
	}
	return GetEnvironmentVariablesDisplay(&environmentVariables, columnPadding)
}

// TODO(Roee G.): This function is duplicated in template/describe/display_template_describe.go
// Consolidate into a shared utility package to avoid code duplication and ensure consistent behavior
func GetEnvironmentVariablesDisplay(environmentVariables *map[string]string, columnPadding int) string {
	if environmentVariables == nil || len(*environmentVariables) == 0 {
		return "[]"
	}

	var envVarsBuilder strings.Builder
	envVarsBuilder.WriteString("[")

	maxKeyLength := 0
	for key := range *environmentVariables {
		if len(key) > maxKeyLength {
			maxKeyLength = len(key)
		}
	}

	keys := lo.Keys(*environmentVariables)
	slices.Sort(keys)

	valueColumnIndent := strings.Repeat(" ", columnPadding+1)

	for i, key := range keys {
		keyValuePart := fmt.Sprintf("%-*s: %s", maxKeyLength, key, (*environmentVariables)[key])
		if i == 0 {
			envVarsBuilder.WriteString(keyValuePart)
		} else {
			envVarsBuilder.WriteString(",\n")
			envVarsBuilder.WriteString(valueColumnIndent)
			envVarsBuilder.WriteString(keyValuePart)
		}
	}

	envVarsBuilder.WriteString("]")

	if len(keys) > MaxDisplayEnvVarsCount {
		truncationIndent := strings.Repeat(" ", columnPadding+1)
		envVarsBuilder.WriteString(fmt.Sprintf("\n%sDisplaying %d/%d environment variables", truncationIndent, MaxDisplayEnvVarsCount, len(keys)))
	}

	return envVarsBuilder.String()
}
