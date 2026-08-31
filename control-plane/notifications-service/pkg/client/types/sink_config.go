/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package types

type SinkConfig struct {
	Id string `json:"id,omitempty" gorm:"not null;"`

	Type string `json:"type,omitempty" gorm:"not null;"`

	Config SinkConfigDetails `json:"config,omitempty" gorm:"serializer:json; type:jsonb;not null;"`
}

type SinkConfigDetails struct {
	TemplatesFolder     string   `mapstructure:"templates_folder,omitempty" json:"templates_folder,omitempty"`
	TemplateFile        string   `mapstructure:"template_file,omitempty" json:"template_file,omitempty"`
	DirectNotifications *bool    `mapstructure:"direct_notifications,omitempty" json:"direct_notifications,omitempty"`
	Recipients          []string `mapstructure:"recipients,omitempty" json:"recipients,omitempty"`

	EmailSinkConfig `mapstructure:",squash"`
	SlackSinkConfig `mapstructure:",squash"`
}

type EmailSinkConfig struct {
	LogoUrl         string `mapstructure:"logo_url,omitempty" json:"logo_url,omitempty"`
	FromDisplayName string `mapstructure:"from_display_name,omitempty" json:"from_display_name,omitempty"`
	From            string `mapstructure:"from,omitempty" json:"from,omitempty"`
	Subject         string `mapstructure:"subject,omitempty" json:"subject,omitempty"`
	SmtpHost        string `mapstructure:"smtp_host,omitempty" json:"smtp_host,omitempty"`
	SmtpPort        int    `mapstructure:"smtp_port,omitempty" json:"smtp_port,omitempty"`
	SmtpTlsEnabled  *bool  `mapstructure:"smtp_tls_enabled,omitempty" json:"smtp_tls_enabled,omitempty"`
	User            string `mapstructure:"user,omitempty" json:"user,omitempty"`
	Password        string `mapstructure:"password,omitempty" json:"password,omitempty" secret:"true"`
	AuthType        string `mapstructure:"auth_type,omitempty" json:"auth_type,omitempty"`
	UseAttachments  *bool  `mapstructure:"use_attachments,omitempty" json:"use_attachments,omitempty"`
}

type SlackSinkConfig struct {
	ApiToken string `mapstructure:"api_token,omitempty" json:"api_token,omitempty" secret:"true"`
}
