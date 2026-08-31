/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

type MyCredentialColumn string

const (
	ColumnKeyType        MyCredentialColumn = "type"
	ColumnKeyName        MyCredentialColumn = "name"
	ColumnKeyDescription MyCredentialColumn = "description"
)

const (
	DefaultWidth = 30
)
