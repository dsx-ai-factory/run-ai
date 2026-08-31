/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package page

type DuplicatorMapper interface {
	ShouldAdd(id string) bool
	ReachedLimit() bool
}
