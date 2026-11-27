/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package environment

import (
	"fmt"
	"maps"
	"slices"
	"strings"
)

// Format renders environment variables as a bracketed, key-sorted list with aligned keys.
// A limit greater than zero appends a "Displaying limit/total" line when exceeded; zero or less means no limit.
func Format(environmentVariables map[string]string, columnPadding, limit int) string {
	if len(environmentVariables) == 0 {
		return "[]"
	}

	keys := slices.Sorted(maps.Keys(environmentVariables))
	maxKeyLength := 0
	for _, key := range keys {
		maxKeyLength = max(maxKeyLength, len(key))
	}

	indent := strings.Repeat(" ", columnPadding+1)

	var builder strings.Builder
	builder.WriteString("[")
	for i, key := range keys {
		if i > 0 {
			builder.WriteString(",\n")
			builder.WriteString(indent)
		}
		fmt.Fprintf(&builder, "%-*s: %s", maxKeyLength, key, environmentVariables[key])
	}
	builder.WriteString("]")

	if limit > 0 && len(keys) > limit {
		fmt.Fprintf(&builder, "\n%sDisplaying %d/%d environment variables", indent, limit, len(keys))
	}

	return builder.String()
}
