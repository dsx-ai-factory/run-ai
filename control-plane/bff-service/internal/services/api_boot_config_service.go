/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/bff-service/internal/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/bff-service/internal/generated/pkg/api"
)

var releaseVersionPattern = regexp.MustCompile(`^v?(\d+)\.(\d+)\.\d+(?:-saas\.\d+)?$`)
var nonDigitPattern = regexp.MustCompile(`\D`)

type BootConfigAPIService struct {
	runaiVersion string
	config       *config.BootConfig
}

func NewBootConfigAPIService(runaiVersion string, bootConfig *config.BootConfig) *BootConfigAPIService {
	return &BootConfigAPIService{runaiVersion: runaiVersion, config: bootConfig}
}

func (s *BootConfigAPIService) GetBootConfig(_ context.Context) (api.ImplResponse, error) {
	return api.Response(http.StatusOK, s.resolve()), nil
}

func (s *BootConfigAPIService) resolve() api.BootConfig {
	trimmedVersion := strings.TrimSpace(s.runaiVersion)
	docsBaseURL := strings.TrimRight(s.config.DocsBaseURL, "/")
	whatsNewBaseURL := docsBaseURL
	versionMatches := releaseVersionPattern.FindStringSubmatch(trimmedVersion)
	hasDocsVersion := hasUsableDocsVersion(versionMatches)
	if s.config.DocsVersioned && hasDocsVersion {
		docsBaseURL += "/" + versionMatches[1] + "." + versionMatches[2]
	}

	whatsNewURL := strings.ReplaceAll(s.config.DocsWhatsNewURLTemplate, "{docsBaseUrl}", whatsNewBaseURL)
	whatsNewLabel := "What's New"
	if s.config.DocsVersioned && hasDocsVersion {
		whatsNewURL += "-" + versionMatches[1] + "-" + versionMatches[2]
		whatsNewLabel += " in " + versionMatches[1] + "." + versionMatches[2]
	}

	displayVersion := trimmedVersion
	if s.config.VersionDisplayPrefix != "" {
		displayVersion = nonDigitPattern.ReplaceAllString(s.config.VersionDisplayPrefix+trimmedVersion, "")
	}

	result := api.BootConfig{
		DistributionAssets: api.BootConfigDistributionAssets{Profile: s.resolveDistributionAssetsProfile()},
		Version:            api.BootConfigVersion{Display: displayVersion},
		Docs: api.BootConfigDocs{
			BaseUrl: docsBaseURL,
			WhatsNew: api.BootConfigWhatsNew{
				Label: whatsNewLabel,
				Url:   whatsNewURL,
			},
		},
		Support: api.BootConfigSupport{
			Url:   s.config.SupportURL,
			Label: s.config.SupportLabel,
		},
		Capabilities: api.BootConfigCapabilities{
			ShowsClusterLocationStep: s.config.ShowsClusterLocationStep,
		},
	}

	if s.hasCompleteUnleashConfig() {
		unleashConfig := api.BootConfigUnleashConfig{
			Url:         s.config.UnleashURL,
			ClientKey:   s.config.UnleashClientKey,
			AppName:     s.config.UnleashAppName,
			Environment: s.config.UnleashEnvironment,
		}
		if strings.TrimSpace(s.config.UnleashProject) != "" {
			unleashConfig.Project = s.config.UnleashProject
		}
		result.UnleashConfig = &unleashConfig
	}

	return result
}

func (s *BootConfigAPIService) resolveDistributionAssetsProfile() string {
	if strings.EqualFold(
		strings.TrimSpace(string(s.config.DistributionAssetsProfile)),
		string(config.DistributionAssetsProfileCE),
	) {
		return string(config.DistributionAssetsProfileCE)
	}

	return string(config.DistributionAssetsProfileCC)
}

func hasUsableDocsVersion(versionMatches []string) bool {
	return len(versionMatches) >= 3 && (versionMatches[1] != "0" || versionMatches[2] != "0")
}

func (s *BootConfigAPIService) hasCompleteUnleashConfig() bool {
	return s.config.UnleashEnabled &&
		strings.TrimSpace(s.config.UnleashURL) != "" &&
		strings.TrimSpace(s.config.UnleashClientKey) != "" &&
		strings.TrimSpace(s.config.UnleashAppName) != "" &&
		strings.TrimSpace(s.config.UnleashEnvironment) != ""
}
