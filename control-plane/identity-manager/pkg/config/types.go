/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

type Configuration struct {
	UseMocks         bool `mapstructure:"use_mocks"`
	AllowHTTPIdpUrls bool `mapstructure:"allow_http_idp_urls"`
}

func (c *Configuration) SetDefaults() {
	c.UseMocks = false
	c.AllowHTTPIdpUrls = false
}

type Iam struct {
	Url                                      string   `mapstructure:"keycloak_host"`
	AdminUsername                            string   `mapstructure:"keycloak_admin_username,omitempty" config:"secret"`
	AdminPassword                            string   `mapstructure:"keycloak_admin_password" config:"secret"`
	IgnoredRealms                            []string `mapstructure:"ignored_realms,omitempty"`
	TenantsManagerUrl                        string   `mapstructure:"tenant_mng_url"`
	ReconcilerInternalServiceAccountsVersion string   `mapstructure:"internal_service_accounts_version"`
	ReconcilerServiceAccountsVersion         string   `mapstructure:"service_accounts_version"`
	ReconcilerAccessKeysVersion              string   `mapstructure:"access_keys_version"`
	MountPath                                string   `mapstructure:"secrets_mount_path"`
}

func (i *Iam) SetDefaults() {
	i.MountPath = "/etc/secrets"
	i.Url = "http://keycloak-http"
	i.IgnoredRealms = []string{}
	i.TenantsManagerUrl = "http://runai-backend-tenants-manager:8080"
	i.ReconcilerServiceAccountsVersion = "v1"
	i.ReconcilerInternalServiceAccountsVersion = "v1"
	i.ReconcilerAccessKeysVersion = "v1"
}

type Notify struct {
	NotifyUrl                  string `mapstructure:"notify_url"`
	ConfigUrl                  string `mapstructure:"config_url"`
	Enabled                    bool   `mapstructure:"enabled"`
	WelcomeEmailTemplate       string `mapstructure:"welcome_email_template"`
	ResetPasswordEmailTemplate string `mapstructure:"reset_password_email_template"`
}

func (c *Notify) SetDefaults() {
	c.NotifyUrl = "http://runai-backend-notifications-service:5000"
	c.ConfigUrl = "http://runai-backend-notifications-service:8080"
	c.Enabled = true
	c.WelcomeEmailTemplate = "keycloak_welcome"
	c.ResetPasswordEmailTemplate = "keycloak_reset"
}

type TenantRealm struct {
	AppHost                 string `mapstructure:"app_url"`
	TenantDomainTemplate    string `mapstructure:"tenant_domain_template"`
	Env                     string `mapstructure:"environment"`
	DefaultClientId         string `mapstructure:"default_client_id"`
	ClientCachingExpiration int    `mapstructure:"client_caching_expiration"`
	Name                    string `mapstructure:"tenant_name"`
	IsMultiTenant           bool   `mapstructure:"multi_tenancy_enabled"`
	LoginTheme              string `mapstructure:"login_theme"`
	SmtpHost                string `mapstructure:"smtp_host"`
	SmtpPort                int    `mapstructure:"smtp_port"`
	SmtpFromEmail           string `mapstructure:"smtp_from_email"`
	HostHeader              string `mapstructure:"host_header"`
}

func (tr *TenantRealm) SetDefaults() {
	tr.AppHost = "https://app.run.ai"
	tr.TenantDomainTemplate = "https://{{.}}.run.ai"
	tr.Env = "development"
	tr.DefaultClientId = "runai-admin-ui"
	tr.Name = ""
	tr.IsMultiTenant = false
	tr.LoginTheme = "runai-theme"
	tr.ClientCachingExpiration = 60 * 24
	tr.SmtpHost = "runai-backend-notifications-service"
	tr.SmtpPort = 2525
	tr.SmtpFromEmail = "nvidia-runai@nvidia.com"
	tr.HostHeader = "Host"
}
