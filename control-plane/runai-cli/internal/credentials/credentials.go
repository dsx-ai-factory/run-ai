/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package credentials

import "encoding/json"

const (
	InteractiveAppName   = "Client ID"
	InteractiveAppSecret = "Client Secret"
)

type AccessKeyAuth struct {
	AppName   string `json:"client_id"`
	AppSecret string `json:"client_secret"`
}

func NewAccessKeyAuth(appName, appSecret string) *AccessKeyAuth {
	return &AccessKeyAuth{AppName: appName, AppSecret: appSecret}
}

func (aka *AccessKeyAuth) UnmarshalJSON(data []byte) error {
	type Alias struct {
		ClientID     string `json:"client_id"`
		AppName      string `json:"app_name"`
		ClientSecret string `json:"client_secret"`
		AppSecret    string `json:"app_secret"`
	}

	var temp Alias
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	if temp.ClientID != "" {
		aka.AppName = temp.ClientID
	} else {
		aka.AppName = temp.AppName
	}

	if temp.ClientSecret != "" {
		aka.AppSecret = temp.ClientSecret
	} else {
		aka.AppSecret = temp.AppSecret
	}

	return nil
}

func (aka *AccessKeyAuth) GetType() LoginType {
	return AccessKeyIDLogin
}

func (aka *AccessKeyAuth) GetID() string {
	return aka.AppName
}

func (aka *AccessKeyAuth) GetSecret() string {
	return aka.AppSecret
}
