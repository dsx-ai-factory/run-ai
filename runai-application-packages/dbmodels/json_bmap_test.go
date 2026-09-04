/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dbmodels

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValue_ShouldReturnNull(t *testing.T) {
	var jsonStruct JsonBMap
	val, err := jsonStruct.Value()
	require.Nil(t, err)
	require.Equal(t, val.(string), "null")
}

func TestValue_MapShouldReturnJson(t *testing.T) {
	jsonStruct := JsonBMap{
		"num": 5,
		"str": "abc",
	}
	val, err := jsonStruct.Value()
	require.Nil(t, err)
	require.Contains(t, val.(string), "\"num\":5")
	require.Contains(t, val.(string), "\"str\":\"abc\"")
}

func TestValue_JsonShouldReturnCorrectMap(t *testing.T) {
	jsonStruct := JsonBMap{}
	value := "{ \"num\": 5, \"str\": \"abc\" }"
	err := jsonStruct.Scan([]byte(value))
	require.Nil(t, err)

	expected := JsonBMap{
		"num": json.Number("5"),
		"str": "abc",
	}
	require.Equal(t, jsonStruct, expected)
}

func TestValue_StrJsonShouldReturnCorrectMap(t *testing.T) {
	jsonStruct := JsonBMap{}
	value := "{ \"num\": 5, \"str\": \"abc\" }"
	err := jsonStruct.Scan(value)
	require.Nil(t, err)

	expected := JsonBMap{
		"num": json.Number("5"),
		"str": "abc",
	}
	require.Equal(t, jsonStruct, expected)
}

func TestValue_BadJsonShouldReturnError(t *testing.T) {
	jsonStruct := JsonBMap{}
	value := "{ \"num\": xx }"
	err := jsonStruct.Scan([]byte(value))
	require.Contains(t, err.Error(), "invalid character 'x'")
}
