/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package datastructs

// Set is a collection of unique comparable values. It allows you to check existence
// of a value in the set in O(1)
type Set[T comparable] map[T]bool

// NewSet is a constructor that can receive a set of values or an array
//
//	NewSet - empty set
//	NewSet(1, 3, 9) - set containing 1, 3, and 9
//	NewSet(array...) - set uniquely containing all elements of the array
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

// Add insert values to the set
func (s Set[T]) Add(values ...T) Set[T] {
	for _, value := range values {
		s[value] = false
	}
	return s
}

// Add delete values from the set
func (s Set[T]) Delete(values ...T) Set[T] {
	for _, value := range values {
		delete(s, value)
	}
	return s
}

// Has returns a boolean, for whether the value exists in the Set.
func (s Set[T]) Has(value T) bool {
	_, ok := s[value]
	return ok
}

func (s Set[T]) Keys() []T {
	result := make([]T, 0, len(s))

	for k := range s {
		result = append(result, k)
	}

	return result
}

func (s Set[T]) EqualTo(other Set[T]) bool {
	if len(s) != len(other) {
		return false
	}

	for k := range s {
		if !other.Has(k) {
			return false
		}
	}

	return true
}
