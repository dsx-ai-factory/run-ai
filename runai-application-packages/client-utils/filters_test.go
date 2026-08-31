/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client_utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilters(t *testing.T) {
	builder := NewFilterBuilder()

	builder.AddFilter("field1", EqualOp, "abc")
	builder.AddFilter("field2", NotEqualOp, "123")
	builder.AddFilters("field3", ContainsOp, []string{
		"def",
		"ghi",
		"jkl",
	})

	filters := builder.Get()

	assert.Equal(t, len(filters), 3)

	assert.Equal(t, filters[0], "field1==abc")
	assert.Equal(t, filters[1], "field2!=123")
	assert.Equal(t, filters[2], "field3=@def||ghi||jkl")
}
