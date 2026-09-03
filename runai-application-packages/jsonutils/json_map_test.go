/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package jsonutils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEmptySection(t *testing.T) {
	actual := make(map[string]interface{})
	newSection := AddSection(actual, "first")

	expected := map[string]interface{}{
		"first": make(map[string]interface{}),
	}

	assert.Equal(t, expected, actual)

	newSectionAsMap := MapMe(newSection)
	newSectionAsMap["second"] = "child-of-first"

	expected["first"] = map[string]interface{}{
		"second": "child-of-first",
	}

	assert.Equal(t, expected, actual)
}

func TestExistingSection(t *testing.T) {
	actual := map[string]interface{}{
		"first": "1st",
		"2nd":   "2nd",
		"3rd": map[string]interface{}{
			"3.1": "3rd.1st",
		},
	}

	existingSection := AddSection(actual, "3rd")

	expectedSection := map[string]interface{}{
		"3.1": "3rd.1st",
	}

	assert.Equal(t, existingSection, expectedSection)
}
