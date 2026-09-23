/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

type Configuration struct {
	UseMocks bool `mapstructure:"use_mocks"`
}

func (c *Configuration) SetDefaults() {
	c.UseMocks = false
}

type Service struct {
	RunaiVersion string `mapstructure:"runai_version"`
}

func (s *Service) SetDefaults() {
	s.RunaiVersion = "0.0.0"
}

type DistributionAssetsProfile string

const (
	// CC and CE are the currently agreed distribution asset profile identifiers.
	// They may be renamed when the final product terminology is defined.
	DistributionAssetsProfileCC DistributionAssetsProfile = "cc"
	DistributionAssetsProfileCE DistributionAssetsProfile = "ce"
)

type BootConfig struct {
	SupportEntitled           bool                      `mapstructure:"support_entitled"`
	DistributionAssetsProfile DistributionAssetsProfile `mapstructure:"distribution_assets_profile"`
	DocsBaseURL               string                    `mapstructure:"docs_base_url"`
	DocsCatalogVersion        string                    `mapstructure:"docs_catalog_version"`
	DocsVersioned             bool                      `mapstructure:"docs_versioned"`
	DocsWhatsNewURLTemplate   string                    `mapstructure:"docs_whats_new_url_template"`
	SupportURL                string                    `mapstructure:"support_url"`
	SupportLabel              string                    `mapstructure:"support_label"`
	VersionDisplayPrefix      string                    `mapstructure:"version_display_prefix"`
	ShowsClusterLocationStep  bool                      `mapstructure:"shows_cluster_location_step"`
	UnleashEnabled            bool                      `mapstructure:"unleash_enabled"`
	UnleashURL                string                    `mapstructure:"unleash_url"`
	UnleashClientKey          string                    `mapstructure:"unleash_client_key"`
	UnleashAppName            string                    `mapstructure:"unleash_app_name"`
	UnleashEnvironment        string                    `mapstructure:"unleash_environment"`
	UnleashProject            string                    `mapstructure:"unleash_project"`
}

func (b *BootConfig) SetDefaults() {
	b.DistributionAssetsProfile = DistributionAssetsProfileCC
	b.DocsBaseURL = "https://docs.nvidia.com/run-ai/self-hosted"
	b.DocsVersioned = true
	b.DocsCatalogVersion = "v2"
	b.DocsWhatsNewURLTemplate = "{docsBaseUrl}/documentation/getting-started/whats-new/whats-new"
	b.SupportLabel = "NVIDIA Run:ai support"
	b.ShowsClusterLocationStep = true
}

type Backend struct {
	BackendUrl       string `mapstructure:"backend_service_url"`
	TenantManagerUrl string `mapstructure:"tenant_manager_url"`
}

func (b *Backend) SetDefaults() {
	b.BackendUrl = "http://runai-backend-temp-backend-settings:8080"
	b.TenantManagerUrl = "http://runai-backend-tenants-manager:8080"
}
