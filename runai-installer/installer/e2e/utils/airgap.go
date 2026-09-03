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
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	JfrogAirgapBaseURL = "https://runai.jfrog.io/artifactory/runai-airgapped-prod"
	harborProject      = "runai"

	jfrogRegistryKey      = "https://runai.jfrog.io"
	airgapDownloadTimeout = 30 * time.Minute
)

func HarborUser() string {
	return os.Getenv("E2E_HARBOR_USERNAME")
}

func HarborPass() string {
	return os.Getenv("E2E_HARBOR_PASSWORD")
}

func HarborRegistryURL(clusterName, kubernetesDistribution string) string {
	if kubernetesDistribution == "openshift" {
		return fmt.Sprintf("%s-harbor.apps.%s.runailabs.com/%s", clusterName, clusterName, harborProject)
	}
	return fmt.Sprintf("%s-harbor.runailabs.com/%s", clusterName, harborProject)
}

func extractJfrogAuth(dockerConfigBase64 string) (string, error) {
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

	entry, ok := cfg.Auths[jfrogRegistryKey]
	if !ok {
		return "", fmt.Errorf("no auth entry for %s in docker config", jfrogRegistryKey)
	}
	return entry.Auth, nil
}

func DownloadAirgapPackage(jfrogToken, version, destDir string) (string, error) {
	basicAuth, err := extractJfrogAuth(jfrogToken)
	if err != nil {
		return "", fmt.Errorf("extract jfrog credentials: %w", err)
	}

	filename := fmt.Sprintf("runai-airgapped-package-%s.tar.gz", version)
	url := fmt.Sprintf("%s/%s", JfrogAirgapBaseURL, filename)
	destPath := filepath.Join(destDir, filename)

	client := &http.Client{Timeout: airgapDownloadTimeout}

	req, err := http.NewRequest("GET", url, nil)
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
		return "", fmt.Errorf("download %s failed with status %d", url, resp.StatusCode)
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
