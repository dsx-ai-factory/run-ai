/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package convutils

// Ref returns a pointer to v, this is useful for referring const and return value of a function
func Ref[T any](v T) *T {
	return &v
}

// Deref returns a pointer de-reference, or default value in case it nil
func Deref[T any](p *T, defValue T) T {
	if p == nil {
		return defValue
	}
	return *p
}

func Must[T any](obj T, err error) T {
	if err != nil {
		panic(err)
	}
	return obj
}

// ValueOrDefault return the input value unless it is the zero value, than it returns the default value
func ValueOrDefault[T comparable](value T, defValue T) T {
	var zeroT T
	if value != zeroT {
		return value
	}
	return defValue
}
