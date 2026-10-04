/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	AirgapBaseURLEnv      = "E2E_AIRGAP_BASE_URL"
	HarborDomainEnv       = "E2E_HARBOR_DOMAIN"
	DefaultHarborDomain   = "example.com"
	harborProject         = "runai"
	airgapDownloadTimeout = 30 * time.Minute
)

// AirgapBaseURL returns the configured air-gap tarball base URL without a
// trailing slash, or "" when none is set.
func AirgapBaseURL() string {
	return strings.TrimSuffix(strings.TrimSpace(os.Getenv(AirgapBaseURLEnv)), "/")
}

func HarborUser() string {
	return os.Getenv("E2E_HARBOR_USERNAME")
}

func HarborPass() string {
	return os.Getenv("E2E_HARBOR_PASSWORD")
}

func HarborDomain() string {
	if domain := strings.TrimSpace(os.Getenv(HarborDomainEnv)); domain != "" {
		return domain
	}
	return DefaultHarborDomain
}

func HarborRegistryURL(clusterName, kubernetesDistribution string) string {
	if kubernetesDistribution == "openshift" {
		return fmt.Sprintf("%s-harbor.apps.%s.%s/%s", clusterName, clusterName, HarborDomain(), harborProject)
	}
	return fmt.Sprintf("%s-harbor.%s/%s", clusterName, HarborDomain(), harborProject)
}

// extractRegistryAuth returns the basic-auth value the pull secret holds for
// the host of baseURL. Docker config keys may be written with or without a
// scheme, so both forms are accepted.
func extractRegistryAuth(dockerConfigBase64, baseURL string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(dockerConfigBase64))
	if err != nil {
		return "", fmt.Errorf("decode docker config: %w", err)
	}

	var cfg struct {
		Auths map[string]struct {
			Auth string `json:"auth"`
		} `json:"auths"`
	}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return "", fmt.Errorf("parse docker config: %w", err)
	}

	u, err := url.Parse(baseURL)
	if err != nil || u.Host == "" {
		return "", fmt.Errorf("parse %s %q: %w", AirgapBaseURLEnv, baseURL, err)
	}
	for _, key := range []string{u.Scheme + "://" + u.Host, u.Host} {
		if entry, ok := cfg.Auths[key]; ok {
			return entry.Auth, nil
		}
	}
	return "", fmt.Errorf("no auth entry for %s in docker config", u.Host)
}

func DownloadAirgapPackage(pullSecret, version, destDir string) (string, error) {
	baseURL := AirgapBaseURL()
	if baseURL == "" {
		return "", fmt.Errorf("%s is not set", AirgapBaseURLEnv)
	}
	basicAuth, err := extractRegistryAuth(pullSecret, baseURL)
	if err != nil {
		return "", fmt.Errorf("extract registry credentials: %w", err)
	}

	filename := fmt.Sprintf("runai-airgapped-package-%s.tar.gz", version)
	downloadURL := fmt.Sprintf("%s/%s", baseURL, filename)
	destPath := filepath.Join(destDir, filename)

	client := &http.Client{Timeout: airgapDownloadTimeout}

	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Basic "+basicAuth)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("download airgap package: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download %s failed with status %d", downloadURL, resp.StatusCode)
	}

	f, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	return destPath, nil
}
