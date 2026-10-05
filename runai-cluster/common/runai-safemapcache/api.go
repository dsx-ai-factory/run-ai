/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package safemapcache

type Cache[K any, V any] interface {
	Store(key K, value V)
	Load(key K) (value V, found bool, valid bool)
	Delete(key K)
	Clear()
	Destroy()
	Active() bool
}
