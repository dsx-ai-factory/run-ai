/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import "strings"

// RemoveKeys returns a recursive copy of a JSON-like value without keys that
// case-insensitively match any entry in keys. The input is not modified.
func RemoveKeys(value any, keys []string) any {
	switch value := value.(type) {
	case map[string]any:
		clean := make(map[string]any, len(value))
		for key, nested := range value {
			if containsFold(keys, key) {
				continue
			}
			clean[key] = RemoveKeys(nested, keys)
		}
		return clean
	case []any:
		clean := make([]any, len(value))
		for i, nested := range value {
			clean[i] = RemoveKeys(nested, keys)
		}
		return clean
	case []map[string]any:
		clean := make([]map[string]any, len(value))
		for i, nested := range value {
			clean[i] = RemoveKeys(nested, keys).(map[string]any)
		}
		return clean
	default:
		return value
	}
}

func containsFold(values []string, target string) bool {
	for _, value := range values {
		if strings.EqualFold(value, target) {
			return true
		}
	}
	return false
}
