/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package migration

const (
	SuspendPatchString = `{"spec":{"active":{"value": false }}}`
)

var (
	paginationLimit int64 = 500
)

func setLimit(l int64) {
	paginationLimit = l
}
