package services

import (
	"encoding/json"
	"testing"

	"github.com/run-ai/runai/control-plane/bff-service/internal/config"
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
