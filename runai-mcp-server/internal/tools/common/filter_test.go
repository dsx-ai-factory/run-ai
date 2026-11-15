/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilterEq(t *testing.T) {
	assert.Equal(t, "clusterId==abc", FilterEq("clusterId", "abc"))
}

func TestValidateNoCommas(t *testing.T) {
	assert.NoError(t, ValidateNoCommas("clusterId", "abc"))
	assert.NoError(t, ValidateNoCommas("clusterId", ""))

	err := ValidateNoCommas("clusterId", "a,b")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "clusterId", "message names the offending field")
}

func TestValidateFilterValueRejectsFilterSeparators(t *testing.T) {
	for _, value := range []string{"a,b", "a==b", "a!=b", "a<=b", "a>=b", "a=@b", "a!@b", "a=^b", "a=$b", "a||b"} {
		t.Run(value, func(t *testing.T) {
			assert.Error(t, ValidateFilterValue("name", value))
		})
	}
	assert.NoError(t, ValidateFilterValue("name", "shared-cluster"))
}
