/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package jsonutils

// AddSection adds a json section to a json map, in case it does not exist
// if it exists, it returns the existing one.
func AddSection(aMap map[string]interface{}, section string) map[string]interface{} {
	existingSection, existing := aMap[section]
	if existing {
		return MapMe(existingSection)
	}

	newSection := make(map[string]interface{})
	aMap[section] = newSection
	return newSection
}

func MapMe(src interface{}) map[string]interface{} {
	switch realType := src.(type) {
	case map[string]interface{}:
		return realType
	default:
		return make(map[string]interface{})
	}
}
