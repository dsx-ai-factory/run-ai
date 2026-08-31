/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package structs

type OrgUnitKey struct {
	Scope string
	Id    string
}

type OrgUnitMap map[OrgUnitKey]OrgUnitKey
