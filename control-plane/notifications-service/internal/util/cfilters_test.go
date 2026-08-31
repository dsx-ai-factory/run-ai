/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package util

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testStruct struct {
	someId string
}

func testMap() map[string]testStruct {
	return map[string]testStruct{
		"1": {"id1"},
		"2": {"id2"},
		"3": {"id3"},
	}
}

func testSlice() []testStruct {
	return []testStruct{
		{"id1"},
		{"id2"},
		{"id3"},
	}
}

func TestFilterMap(t *testing.T) {
	filtered := FilterMap(testMap(), func(key string, val testStruct) bool {
		return val.someId == "id1"
	})

	require.Len(t, filtered, 1)
	require.Contains(t, filtered, "1")
}

func TestFilterMapByKeys(t *testing.T) {
	filtered := FilterMapByKeys(testMap(), "2")
	require.Len(t, filtered, 1)
	require.Contains(t, filtered, "2")
}

func TestFilterSlice(t *testing.T) {
	filtered := FilterSlice(testSlice(), func(idx int, val testStruct) bool {
		return idx > 1 || val.someId == "id2"
	})
	require.Len(t, filtered, 2)
	require.Equal(t, "id2", filtered[0].someId)
	require.Equal(t, "id3", filtered[1].someId)
}

func TestMigrateSlice(t *testing.T) {
	migrated, err := MigrateSlice(testSlice(), func(idx int, val testStruct) (testStruct, error) {
		if val.someId == "id2" {
			return testStruct{"updatedId"}, nil
		}
		return val, nil
	})
	require.NoError(t, err)
	require.Len(t, migrated, 3)
	require.Equal(t, "updatedId", migrated[1].someId)
}

func TestMigrateSliceToMap(t *testing.T) {
	migrated := MigrateSliceToMap(testSlice(), func(idx int, val testStruct) (string, testStruct, error) {
		return val.someId, val, nil
	})
	require.Len(t, migrated, 3)
	require.Equal(t, "id1", migrated["id1"].someId)
}
