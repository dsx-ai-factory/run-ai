/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package structs

type contextKey string

func (c contextKey) String() string {
	return string(c)
}

var (
	PermittedScopesKey = contextKey("permitted-scopes")
)
