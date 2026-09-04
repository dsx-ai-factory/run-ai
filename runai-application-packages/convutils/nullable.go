/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package convutils

// Nullable is a helper interface for the HasValue function, see the description of that function
// for further details
type Nullable interface {
	IsSet() bool
}

type NullableStruct[T any] interface {
	Nullable
	Get() *T
}

func HasValue[T any](v NullableStruct[T]) bool {
	if !v.IsSet() {
		return false
	}
	return v.Get() != nil
}
