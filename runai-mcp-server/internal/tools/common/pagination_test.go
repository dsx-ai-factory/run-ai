/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidatePagination(t *testing.T) {
	cases := []struct {
		name          string
		limit, offset int32
		wantErr       bool
	}{
		{"zero ok", 0, 0, false},
		{"within bounds", 50, 10, false},
		{"max ok", MaxLimit, 0, false},
		{"negative limit", -1, 0, true},
		{"over max", MaxLimit + 1, 0, true},
		{"negative offset", 50, -1, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePagination(tc.limit, tc.offset)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestEffectiveLimit(t *testing.T) {
	assert.Equal(t, DefaultLimit, EffectiveLimit(0), "unset resolves to default")
	assert.Equal(t, int32(25), EffectiveLimit(25), "explicit limit is kept")
}

func TestPaginate(t *testing.T) {
	items := []int{0, 1, 2, 3, 4}

	t.Run("first page has more", func(t *testing.T) {
		page, hasMore, next := Paginate(items, 0, 2)
		assert.Equal(t, []int{0, 1}, page)
		assert.True(t, hasMore)
		assert.Equal(t, int32(2), next)
	})

	t.Run("last partial page", func(t *testing.T) {
		page, hasMore, next := Paginate(items, 4, 2)
		assert.Equal(t, []int{4}, page)
		assert.False(t, hasMore)
		assert.Zero(t, next)
	})

	t.Run("offset past end is empty and non-nil", func(t *testing.T) {
		page, hasMore, next := Paginate(items, 99, 2)
		assert.NotNil(t, page)
		assert.Empty(t, page)
		assert.False(t, hasMore)
		assert.Zero(t, next)
	})

	t.Run("exact boundary has no more", func(t *testing.T) {
		_, hasMore, next := Paginate(items, 0, 5)
		assert.False(t, hasMore)
		assert.Zero(t, next)
	})
}
