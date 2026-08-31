/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaicloak

type Tokens struct {
	AccessToken  string `json:"accessToken,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
	IdToken      string `json:"idToken,omitempty"`
}

type tokenExchangeResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	NotBeforePolicy  int    `json:"not-before-policy"`
	SessionState     string `json:"session_state"`
	Scope            string `json:"scope"`
	IssuedTokenType  string `json:"issued_token_type"`
}

type executionConfig struct {
	DefaultProvider string `json:"defaultProvider,omitempty"`
}

type executionConfigRequest struct {
	Alias  string          `json:"alias,omitempty"`
	Config executionConfig `json:"config,omitempty"`
}
