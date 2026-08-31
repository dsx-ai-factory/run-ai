/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package userinfo

import "github.com/run-ai/runai/runai-common-packages/auth/providers/oidc"

type Provider interface {
	Load(userIdentifier string) (*oidc.UserInfo, error)
}
