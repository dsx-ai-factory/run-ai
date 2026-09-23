/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package userinfo

import "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-api/internal/auth/oidc"

type Provider interface {
	Load(userIdentifier string) (*oidc.UserInfo, error)
}
