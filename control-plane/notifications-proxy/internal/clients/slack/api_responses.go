/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package slack

type CreateAppResponse struct {
	Ok          bool                 `json:"ok"`
	Error       string               `json:"error,omitempty"`
	AppId       string               `json:"app_id,omitempty"`
	Credentials CreateAppCredentials `json:"credentials,omitempty"`
}

type CreateAppCredentials struct {
	ClientId     string `json:"client_id,omitempty"`
	ClientSecret string `json:"client_secret,omitempty"`
}

type oauthAccessResponse struct {
	Ok         bool   `json:"ok,omitempty"`
	AppId      string `json:"app_id,omitempty"`
	AuthedUser struct {
		ID string `json:"id,omitempty"`
	} `json:"authed_user,omitempty"`
	Scope               string                  `json:"scope,omitempty"`
	TokenType           string                  `json:"token_type,omitempty"`
	AccessToken         string                  `json:"access_token,omitempty"`
	BotUserID           string                  `json:"bot_user_id,omitempty"`
	Team                oauthAccessResponseTeam `json:"team,omitempty"`
	Enterprise          interface{}             `json:"enterprise,omitempty"`
	IsEnterpriseInstall bool                    `json:"is_enterprise_install,omitempty"`
}

type oauthAccessResponseTeam struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}
