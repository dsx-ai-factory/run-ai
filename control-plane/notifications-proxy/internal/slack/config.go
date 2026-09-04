/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package slack

type SlackConfig struct {
	ClientId     string `mapstructure:"slack_client_id" config:"secret"`
	ClientSecret string `mapstructure:"slack_client_secret" config:"secret"`
	BaseUrl      string `mapstructure:"slack_base_url" validate:"required"`
}

func (sc *SlackConfig) SetDefaults() {
	sc.BaseUrl = "https://slack.com"
}

type ValidationResponse struct {
	Status  bool
	Message string
}
