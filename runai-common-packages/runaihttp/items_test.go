/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaihttp

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetNumberOfItems_StringSlice(t *testing.T) {
	arr := []string{"test", "test2"}
	length, err := getNumberOfItems(arr)
	require.NoError(t, err)
	require.Equal(t, 2, length)
}

func TestGetNumberOfItems_StringArray(t *testing.T) {
	arr := [2]string{"test", "test2"}
	length, err := getNumberOfItems(arr)
	require.NoError(t, err)
	require.Equal(t, 2, length)
}

func TestGetNumberOfItems_StructSlice(t *testing.T) {
	type TestStruct struct {
		Field string
	}
	arr := []TestStruct{{Field: "test"}, {Field: "test2"}}
	length, err := getNumberOfItems(arr)
	require.NoError(t, err)
	require.Equal(t, 2, length)
}

func TestGetNumberOfItems_ZeroLength(t *testing.T) {
	arr := []string{}
	length, err := getNumberOfItems(arr)
	require.NoError(t, err)
	require.Equal(t, 0, length)
}

func TestGetNumberOfItems_InvalidType(t *testing.T) {
	arr := "test"
	length, err := getNumberOfItems(arr)
	require.Error(t, err)
	require.Equal(t, -1, length)
}

func TestGetNumberOfItems_Map(t *testing.T) {
	arr := map[string]string{"test": "test2"}
	length, err := getNumberOfItems(arr)
	require.NoError(t, err)
	require.Equal(t, 1, length)
}

func TestGetNumberOfItems_Map_ZeroLength(t *testing.T) {
	arr := map[string]string{}
	length, err := getNumberOfItems(arr)
	require.NoError(t, err)
	require.Equal(t, 0, length)
}

func TestGetNumberOfItems_Map_Structs(t *testing.T) {
	type TestStruct struct {
		Field string
	}
	arr := map[string]TestStruct{"test": {Field: "test2"}}
	length, err := getNumberOfItems(arr)
	require.NoError(t, err)
	require.Equal(t, 1, length)
}

func TestGetNumberOfItems_Map_StringKey_ZeroLength(t *testing.T) {
	arr := map[string]string{}
	length, err := getNumberOfItems(arr)
	require.NoError(t, err)
	require.Equal(t, 0, length)
}

func TestAssertMinItems_Ok(t *testing.T) {
	arr := []string{"test", "test2"}
	require.NoError(t, AssertMinItems(arr, "array", 1))
}

func TestAssertMinItems_TooSmall(t *testing.T) {
	arr := []string{"test"}
	require.EqualError(t, AssertMinItems(arr, "array", 2), "value for array must have at least 2 items")
}

func TestAssertMinItems_NotSlice(t *testing.T) {
	arr := "test"
	require.NoError(t, AssertMinItems(arr, "array", 10))
}

func TestAssertMinItems_Exact(t *testing.T) {
	arr := []string{"test", "test2"}
	require.NoError(t, AssertMinItems(arr, "array", 2))
}

func TestAssertMaxItems_Ok(t *testing.T) {
	arr := []string{}
	require.NoError(t, AssertMaxItems(arr, "array", 1))
}

func TestAssertMaxItems_TooLarge(t *testing.T) {
	arr := []string{"test", "test2"}
	require.EqualError(t, AssertMaxItems(arr, "array", 1), "value for array must have at most 1 items")
}

func TestAssertMaxItems_NotSlice(t *testing.T) {
	arr := "test"
	require.NoError(t, AssertMaxItems(arr, "array", 10))
}

func TestAssertMaxItems_Exact(t *testing.T) {
	arr := []string{}
	require.NoError(t, AssertMaxItems(arr, "array", 0))
}

func TestAssertMinItems_Map(t *testing.T) {
	arr := map[string]string{"test": "test2"}
	require.NoError(t, AssertMinItems(arr, "map", 1))
}

func TestAssertMinItems_Map_ZeroLength(t *testing.T) {
	arr := map[string]string{}
	require.NoError(t, AssertMinItems(arr, "map", 0))
}

func TestAssertMinItems_Map_Structs(t *testing.T) {
	type TestStruct struct {
		Field string
	}
	arr := map[string]TestStruct{"test": {Field: "test2"}}
	require.NoError(t, AssertMinItems(arr, "map", 1))
}

func TestAssertMinItems_Map_StringKey_ZeroLength(t *testing.T) {
	arr := map[string]string{}
	require.NoError(t, AssertMinItems(arr, "map", 0))
}

func TestAssertMaxItems_Map(t *testing.T) {
	arr := map[string]string{"test": "test2"}
	require.NoError(t, AssertMaxItems(arr, "map", 1))
}

func TestAssertMaxItems_Map_ZeroLength(t *testing.T) {
	arr := map[string]string{}
	require.NoError(t, AssertMaxItems(arr, "map", 0))
}

func TestAssertMaxItems_Map_TooLarge(t *testing.T) {
	arr := map[string]string{"test": "test2", "test2": "test3"}
	require.EqualError(t, AssertMaxItems(arr, "map", 1), "value for map must have at most 1 items")
}

func TestAssertMinItems_Map_TooSmall(t *testing.T) {
	arr := map[string]string{}
	require.EqualError(t, AssertMinItems(arr, "map", 1), "value for map must have at least 1 items")
}
