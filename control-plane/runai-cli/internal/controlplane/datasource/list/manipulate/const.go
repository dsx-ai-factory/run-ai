/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

type DatasourceColumn string

const (
	ColumnKeyName  DatasourceColumn = "name"
	ColumnKeyType  DatasourceColumn = "type"
	ColumnKeyScope DatasourceColumn = "scope"
)

const (
	DefaultWidth = 20
)
