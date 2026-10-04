/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package oidc

type Configuration struct {
	Issuer           string `json:"issuer"`
	UserInfoEndpoint string `json:"userinfo_endpoint"`
	JWKSURI          string `json:"jwks_uri"`
}

type UserInfo struct {
	Groups []string `json:"groups"`
}
