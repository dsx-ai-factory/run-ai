/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"encoding/json"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/bff-service/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testBootConfig() *config.BootConfig {
	return &config.BootConfig{
		DistributionAssetsProfile: config.DistributionAssetsProfileCC,
		DocsBaseURL:               "https://docs.example.test/product",
		DocsVersioned:             true,
		DocsWhatsNewURLTemplate:   "{docsBaseUrl}/updates",
		SupportURL:                "https://support.example.test/contact",
		SupportLabel:              "Get help",
		ShowsClusterLocationStep:  true,
	}
}

func TestBootConfigBuildsVersionedDocs(t *testing.T) {
	result := NewBootConfigAPIService("2.25.9", testBootConfig()).resolve()

	assert.Equal(t, "cc", result.DistributionAssets.Profile)
	assert.Equal(t, "2.25.9", result.Version.Display)
	assert.Equal(t, "https://docs.example.test/product/2.25", result.Docs.BaseUrl)
	assert.Equal(t, "What's New in 2.25", result.Docs.WhatsNew.Label)
	assert.Equal(t, "https://docs.example.test/product/updates-2-25", result.Docs.WhatsNew.Url)
	assert.True(t, result.Capabilities.ShowsClusterLocationStep)
	assert.Nil(t, result.UnleashConfig)
}

func TestBootConfigUsesConfiguredDistributionAssetsProfile(t *testing.T) {
	bootConfig := testBootConfig()
	bootConfig.DistributionAssetsProfile = config.DistributionAssetsProfile(" ce ")

	result := NewBootConfigAPIService("2.25.9", bootConfig).resolve()

	assert.Equal(t, "ce", result.DistributionAssets.Profile)
}

func TestBootConfigUsesV2DefaultsWithoutSupportEntitlement(t *testing.T) {
	bootConfig := &config.BootConfig{}
	bootConfig.SetDefaults()

	result := NewBootConfigAPIService("3.0.4", bootConfig).resolve()

	assert.Equal(t, "cc", result.DistributionAssets.Profile)
	assert.Equal(t, "v2", result.Docs.CatalogVersion)
	assert.Equal(t, "https://docs.nvidia.com/run-ai/self-hosted/v3.0", result.Docs.BaseUrl)
	assert.Equal(t, "https://docs.nvidia.com/run-ai/self-hosted/v3.0/documentation/getting-started/whats-new/whats-new-3-0", result.Docs.WhatsNew.Url)
	assert.Empty(t, result.Support.Url)
}

func TestBootConfigKeepsV2DefaultsWithSupportEntitlement(t *testing.T) {
	bootConfig := &config.BootConfig{}
	bootConfig.SetDefaults()
	bootConfig.SupportEntitled = true

	result := NewBootConfigAPIService("3.0.4", bootConfig).resolve()
	payload, err := json.Marshal(result)
	require.NoError(t, err)

	assert.Equal(t, "ce", result.DistributionAssets.Profile)
	assert.Equal(t, "v2", result.Docs.CatalogVersion)
	assert.Equal(t, "https://docs.nvidia.com/run-ai/self-hosted/v3.0", result.Docs.BaseUrl)
	assert.Equal(t, "https://docs.nvidia.com/run-ai/self-hosted/v3.0/documentation/getting-started/whats-new/whats-new-3-0", result.Docs.WhatsNew.Url)
	assert.Equal(t, supportEntitledURL, result.Support.Url)
	assert.Equal(t, supportEntitledLabel, result.Support.Label)
	assert.NotContains(t, string(payload), "supportEntitled")
}

func TestBootConfigPreservesExplicitSaaSConfigWithoutSupportEntitlement(t *testing.T) {
	bootConfig := testBootConfig()
	bootConfig.SupportEntitled = false
	bootConfig.DistributionAssetsProfile = config.DistributionAssetsProfileCE
	bootConfig.DocsBaseURL = "https://docs.example.test/saas"
	bootConfig.DocsVersioned = false
	bootConfig.SupportURL = "https://support.example.test/saas"
	bootConfig.SupportLabel = "SaaS support"

	result := NewBootConfigAPIService("2.25.9-saas.7", bootConfig).resolve()

	assert.Equal(t, "ce", result.DistributionAssets.Profile)
	assert.Equal(t, "https://docs.example.test/saas", result.Docs.BaseUrl)
	assert.Equal(t, "https://support.example.test/saas", result.Support.Url)
	assert.Equal(t, "SaaS support", result.Support.Label)
}

func TestBootConfigNormalizesUnknownDistributionAssetsProfile(t *testing.T) {
	for _, profile := range []string{"", "unknown", "../ce"} {
		t.Run(profile, func(t *testing.T) {
			bootConfig := testBootConfig()
			bootConfig.DistributionAssetsProfile = config.DistributionAssetsProfile(profile)

			result := NewBootConfigAPIService("2.25.9", bootConfig).resolve()

			assert.Equal(t, "cc", result.DistributionAssets.Profile)
		})
	}
}

func TestBootConfigTrimsVersion(t *testing.T) {
	result := NewBootConfigAPIService(" 2.25.9 ", testBootConfig()).resolve()

	assert.Equal(t, "2.25.9", result.Version.Display)
	assert.Equal(t, "https://docs.example.test/product/2.25", result.Docs.BaseUrl)
	assert.Equal(t, "What's New in 2.25", result.Docs.WhatsNew.Label)
}

func TestBootConfigUsesGenericWhatsNewForNonReleaseVersions(t *testing.T) {
	for _, version := range []string{"master", "0.0.0", "not-a-version"} {
		t.Run(version, func(t *testing.T) {
			result := NewBootConfigAPIService(version, testBootConfig()).resolve()
			assert.Equal(t, "What's New", result.Docs.WhatsNew.Label)
			assert.Equal(t, "https://docs.example.test/product/updates", result.Docs.WhatsNew.Url)
		})
	}
}

func TestHasUsableDocsVersionAcceptsAdditionalCaptureGroups(t *testing.T) {
	assert.True(t, hasUsableDocsVersion([]string{"2.25.9", "2", "25", "9"}))
	assert.False(t, hasUsableDocsVersion(nil))
	assert.False(t, hasUsableDocsVersion([]string{"0.0.0", "0", "0"}))
}

func TestBootConfigOmitsSupportURLWhenNotConfigured(t *testing.T) {
	bootConfig := testBootConfig()
	bootConfig.SupportURL = ""

	result := NewBootConfigAPIService("2.25.9", bootConfig).resolve()
	payload, err := json.Marshal(result)
	require.NoError(t, err)

	var decoded map[string]any
	require.NoError(t, json.Unmarshal(payload, &decoded))
	support, ok := decoded["support"].(map[string]any)
	require.True(t, ok)
	assert.NotContains(t, support, "url")
	assert.Equal(t, "Get help", support["label"])
}

func TestBootConfigAppliesDeploymentOverrides(t *testing.T) {
	bootConfig := testBootConfig()
	bootConfig.DocsBaseURL = "/product-docs"
	bootConfig.DocsVersioned = false
	bootConfig.DocsWhatsNewURLTemplate = "{docsBaseUrl}/updates"
	bootConfig.VersionDisplayPrefix = "42-"
	bootConfig.ShowsClusterLocationStep = false
	bootConfig.UnleashEnabled = true
	bootConfig.UnleashURL = "https://unleash.example/api/frontend"
	bootConfig.UnleashClientKey = "browser-public-key"
	bootConfig.UnleashAppName = "browser-app"
	bootConfig.UnleashEnvironment = "development"
	bootConfig.UnleashProject = "example-project"

	result := NewBootConfigAPIService("v2.25.9", bootConfig).resolve()

	assert.Equal(t, "422259", result.Version.Display)
	assert.Equal(t, "What's New", result.Docs.WhatsNew.Label)
	assert.Equal(t, "/product-docs/updates", result.Docs.WhatsNew.Url)
	assert.False(t, result.Capabilities.ShowsClusterLocationStep)
	require.NotNil(t, result.UnleashConfig)
	assert.Equal(t, "browser-public-key", result.UnleashConfig.ClientKey)
	assert.Equal(t, "example-project", result.UnleashConfig.Project)
}

func TestBootConfigOmitsIncompleteUnleashConfig(t *testing.T) {
	bootConfig := testBootConfig()
	bootConfig.UnleashEnabled = true
	bootConfig.UnleashURL = "https://unleash.example/api/frontend"
	bootConfig.UnleashAppName = "browser-app"
	bootConfig.UnleashEnvironment = "development"

	result := NewBootConfigAPIService("2.25.9", bootConfig).resolve()

	assert.Nil(t, result.UnleashConfig)
}

func TestBootConfigCatalogVersioning(t *testing.T) {
	for _, tc := range []struct {
		name, catalog, version, baseSuffix, updatesSuffix string
		versioned                                         bool
	}{
		{name: "v1 release", catalog: "v1", version: "3.0.4", versioned: true, baseSuffix: "/3.0", updatesSuffix: "/updates-3-0"},
		{name: "v2 release", catalog: "v2", version: "v3.0.4", versioned: true, baseSuffix: "/v3.0", updatesSuffix: "/v3.0/updates-3-0"},
		{name: "v2 development", catalog: "v2", version: "0.0.0-build.42", versioned: true, updatesSuffix: "/updates"},
		{name: "v2 unversioned", catalog: "v2", version: "3.0.4", updatesSuffix: "/updates"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, entitled := range []bool{false, true} {
				bootConfig := testBootConfig()
				bootConfig.DocsCatalogVersion = tc.catalog
				bootConfig.DocsVersioned = tc.versioned
				bootConfig.SupportEntitled = entitled
				result := NewBootConfigAPIService(tc.version, bootConfig).resolve()
				assert.Equal(t, tc.catalog, result.Docs.CatalogVersion)
				assert.Equal(t, "https://docs.example.test/product"+tc.baseSuffix, result.Docs.BaseUrl)
				assert.Equal(t, "https://docs.example.test/product"+tc.updatesSuffix, result.Docs.WhatsNew.Url)
			}
		})
	}
}

func TestBootConfigCatalogDefaults(t *testing.T) {
	for _, catalog := range []string{"", "v1", "unsupported"} {
		t.Run(catalog, func(t *testing.T) {
			bootConfig := testBootConfig()
			bootConfig.DocsCatalogVersion = catalog
			result := NewBootConfigAPIService("3.0.1", bootConfig).resolve()
			assert.Equal(t, "v1", result.Docs.CatalogVersion)
		})
	}
}

func TestBootConfigV2PreservesConfiguredTemplateAndRelativeBase(t *testing.T) {
	bootConfig := testBootConfig()
	bootConfig.DocsCatalogVersion = "v2"
	bootConfig.DocsBaseURL = "/internal-docs/"
	bootConfig.DocsWhatsNewURLTemplate = "{docsBaseUrl}/documentation/getting-started/whats-new/whats-new"

	result := NewBootConfigAPIService("3.0.1", bootConfig).resolve()
	assert.Equal(t, "/internal-docs/v3.0", result.Docs.BaseUrl)
	assert.Equal(t, "/internal-docs/v3.0/documentation/getting-started/whats-new/whats-new-3-0", result.Docs.WhatsNew.Url)
}

func TestBootConfigPreservesV1DeploymentOverride(t *testing.T) {
	bootConfig := &config.BootConfig{}
	bootConfig.SetDefaults()
	bootConfig.DistributionAssetsProfile = config.DistributionAssetsProfileCE
	bootConfig.DocsCatalogVersion = "v1"
	bootConfig.DocsBaseURL = "https://run-ai-docs.nvidia.com/saas"
	bootConfig.DocsVersioned = false
	bootConfig.DocsWhatsNewURLTemplate = "{docsBaseUrl}/getting-started/whats-new-for-nvidia-run-ai-saas"

	result := NewBootConfigAPIService("3.0.4", bootConfig).resolve()

	assert.Equal(t, "v1", result.Docs.CatalogVersion)
	assert.Equal(t, "https://run-ai-docs.nvidia.com/saas", result.Docs.BaseUrl)
	assert.Equal(t, "https://run-ai-docs.nvidia.com/saas/getting-started/whats-new-for-nvidia-run-ai-saas", result.Docs.WhatsNew.Url)
}
