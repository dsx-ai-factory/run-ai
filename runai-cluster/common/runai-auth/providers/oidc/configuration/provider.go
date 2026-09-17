/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package configuration

import "github.com/dsx-ai-factory/run-ai/runai-cluster/common/runai-auth/providers/oidc"

type Provider interface {
	Configuration() (*oidc.Configuration, error)
	UserInfo(userIdentifier string) (*oidc.UserInfo, error)
}
