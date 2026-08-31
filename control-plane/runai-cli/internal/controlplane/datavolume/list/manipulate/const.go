/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

type DatavolumeColumn string

const (
	ColumnKeyName      DatavolumeColumn = "name"
	ColumnKeyStatus    DatavolumeColumn = "status"
	ColumnKeyScope     DatavolumeColumn = "scope"
	ColumnKeyOriginPvc DatavolumeColumn = "originpvc"
)

const (
	DefaultWidth = 20
)
