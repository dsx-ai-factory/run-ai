/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package smtp

type AuthType string

const (
	AuthLogin AuthType = "auth_login"
	AuthPlain AuthType = "auth_plain"
	AuthNone  AuthType = "auth_none"
)

type SmtpConfig struct {
	SmtpHost       string   `mapstructure:"smtp_host" json:"smtp_host"`
	SmtpPort       int      `mapstructure:"smtp_port" json:"smtp_port"`
	SmtpTlsEnabled bool     `mapstructure:"smtp_tls_enabled" json:"smtp_tls_enabled"`
	User           string   `mapstructure:"user" json:"user"`
	Password       string   `mapstructure:"password" json:"password" secret:"true"`
	AuthType       AuthType `mapstructure:"auth_type" json:"auth_type"`
}
