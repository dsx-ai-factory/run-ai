/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package authn

import (
	idmclient "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client"
)

type AuthnManager struct {
	idmClient idmclient.IIdm
}

func NewAuthnManager(idmClient idmclient.IIdm) *AuthnManager {
	return &AuthnManager{
		idmClient: idmClient,
	}
}
