/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

type Set[T comparable] map[T]bool

func NewSet[T comparable](values ...T) Set[T] {
	set := make(Set[T], len(values))
	set.Add(values...)
	return set
}

func (s Set[T]) Clone() Set[T] {
	result := make(Set[T], len(s))
	for k := range s {
		result.Add(k)
	}
	return result
}

// Add inserts a value to the set, without retaining insertion order.
func (s Set[T]) Add(values ...T) {
	for _, value := range values {
		s[value] = false
	}
}

// Has returns a boolean, for whether the value exists in the Set.
func (s Set[T]) Has(value T) bool {
	_, ok := s[value]
	return ok
}
