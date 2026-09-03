/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package reflection

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

type b struct {
	findme string `mapstructure:"sometag"`
}

type a struct {
	str     string
	aPtr    *b
	aStruct b
}

func TestNewPathExplorer(t *testing.T) {
	explorer := NewPathExplorer(func(field reflect.StructField) bool {
		return field.Name == "findme"
	}, a{
		str:     "test",
		aPtr:    &b{findme: "test"},
		aStruct: b{findme: "test"},
	})

	require.Len(t, explorer.Paths, 2)
	require.Equal(t, "aPtr.sometag", explorer.Paths[0].GetFullPath())
	require.Equal(t, "aStruct.sometag", explorer.Paths[1].GetFullPath())
}
