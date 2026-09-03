/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package jsonutils

import (
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

type NoJson struct {
	Foo bool
	Bar int
}

type TwoFields struct {
	Foo bool `json:"foo"`
	Bar int  `json:"bar"`
}

func TestCrackerNoJson(t *testing.T) {
	noJsonCracker := NewJsonCracker(NoJson{})
	assert.Equal(t, []string{}, noJsonCracker.AllNames())
	assert.Equal(t, "", noJsonCracker.NameOf(&noJsonCracker.Struct().Foo))
}

func TestCrackerStruct(t *testing.T) {
	noJsonCracker := NewJsonCracker(TwoFields{})
	allNames := noJsonCracker.AllNames()
	sort.Strings(allNames)
	assert.Equal(t, []string{"bar", "foo"}, allNames)
	assert.Equal(t, "bar", noJsonCracker.NameOf(&noJsonCracker.Struct().Bar))
	assert.Equal(t, "foo", noJsonCracker.NameOf(&noJsonCracker.Struct().Foo))
}
