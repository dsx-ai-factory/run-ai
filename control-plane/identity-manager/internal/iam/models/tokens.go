/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package models

import "encoding/json"

const (
	GrantTypeRefreshToken  = "refresh_token"
	GrantTypeAppToken      = "app_token"
	GrantTypeExchangeToken = "exchange_token"
	GrantTypeUserPassword  = "password"
)

type Tokens struct {
	IdToken      string `json:"idToken,omitempty"`
	RefreshToken string `json:"refreshToken,omitempty"`
	AccessToken  string `json:"accessToken,omitempty"`
	ExpiresIn    int    `json:"expiresIn,omitempty"`
}

type UserInfo struct {
	Email  *string  `json:"sub,omitempty"`
	Uuid   *string  `json:"subject_uuid,omitempty"`
	Groups []string `json:"groups,omitempty"`
}

func NewUserInfo(user map[string]interface{}) *UserInfo {
	jsonData, err := json.Marshal(user)
	if err != nil {
		return nil
	}
	var userInfo UserInfo
	err = json.Unmarshal(jsonData, &userInfo)
	if err != nil {
		return nil
	}
	return &userInfo
}
