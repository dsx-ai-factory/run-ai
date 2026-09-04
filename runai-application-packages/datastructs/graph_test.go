/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datastructs

import (
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestBasicCycle(t *testing.T) {
	g := NewDependenciesGraph[string]()

	g.AddDependency("a", "b")
	g.AddDependency("b", "a")

	_, err := g.FindOrder()
	assert.NotEqual(t, nil, err)
	assert.Contains(t, err.Error(), "cycle detected:")
}

func TestCycle(t *testing.T) {
	g := NewDependenciesGraph[string]()

	g.AddDependency("a", "b")
	g.AddDependency("c", "d")
	g.AddDependency("d", "e")
	g.AddDependency("e", "c")

	_, err := g.FindOrder()
	assert.NotEqual(t, nil, err)
	assert.Contains(t, err.Error(), "cycle detected:")
}

func TestFourIsolated(t *testing.T) {
	g := NewDependenciesGraph[string]()

	g.AddDependency("a", "b")
	g.AddDependency("x", "y")

	order, err := g.FindOrder()
	assert.Equal(t, nil, err)
	assert.Equal(t, 4, len(order))

	sort.Strings(order)
	assert.Equal(t, []string{"a", "b", "x", "y"}, order)
}

func TestLongPath(t *testing.T) {
	g := NewDependenciesGraph[string]()

	g.AddDependency("a", "b")
	g.AddDependency("b", "c")
	g.AddDependency("c", "d")
	g.AddDependency("d", "e")

	order, err := g.FindOrder()
	assert.Equal(t, nil, err)
	assert.Equal(t, []string{"e", "d", "c", "b", "a"}, order)
}

func TestMultiplePaths(t *testing.T) {
	g := NewDependenciesGraph[string]()

	g.AddDependency("a", "b")
	g.AddDependency("b", "d")
	g.AddDependency("b", "c")

	order, err := g.FindOrder()
	assert.Equal(t, nil, err)

	assert.True(t,
		assert.ObjectsAreEqual([]string{"d", "c", "b", "a"}, order) ||
			assert.ObjectsAreEqual([]string{"c", "d", "b", "a"}, order),
		"Order should be either [d c b a] or [c d b a]",
	)
}
