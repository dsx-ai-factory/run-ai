/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pre_hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	idmclient "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client"
	semver "github.com/Masterminds/semver/v3"
)

const (
	versionPath                 = "/_version"
	defaultControlPlaneClientID = "runai-internal.installer"
)

var versionResponse struct {
	Version string `json:"version"`
}

func fetchBearerToken(cpURL, clientId, clientSecret string) (string, error) {
	if clientSecret == "" {
		return "", nil
	}
	if clientId == "" {
		clientId = defaultControlPlaneClientID
	}
	idmClient := idmclient.NewIdmExternalClient(cpURL)
	tokenResponse, err := idmClient.ClientCredentialsLogin(context.Background(), clientId, clientSecret)
	if err != nil {
		return "", fmt.Errorf("failed to obtain bearer token from control plane: %w", err)
	}
	return *tokenResponse.AccessToken, nil
}

func getControlPlaneVersion(url, bearerToken string) (string, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url+versionPath, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create GET request to control plane URL: %s", err)
	}
	if bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+bearerToken)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send GET request to control plane URL: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get control plane version: %s", resp.Status)
	}

	if err := json.NewDecoder(resp.Body).Decode(&versionResponse); err != nil {
		return "", fmt.Errorf("failed to decode control plane version response: %w", err)
	}
	if versionResponse.Version == "" {
		return "", fmt.Errorf("control plane version is empty")
	}
	return versionResponse.Version, nil
}

func compareVersions(controlPlaneVersion, clusterVersion string) error {

	if (strings.HasPrefix(controlPlaneVersion, "0.0.0")) || (strings.HasPrefix(clusterVersion, "0.0.0")) {
		// in case of development version
		return nil
	}

	cpVer, err := semver.NewVersion(controlPlaneVersion)
	if err != nil {
		return fmt.Errorf("failed to parse control plane version: %w", err)
	}
	clusterVer, err := semver.NewVersion(clusterVersion)
	if err != nil {
		return fmt.Errorf("failed to parse cluster version: %w", err)
	}

	if clusterVer.Major() > cpVer.Major() || (clusterVer.Major() == cpVer.Major() && clusterVer.Minor() > cpVer.Minor()) {
		return fmt.Errorf("cluster version (%s) is greater than control plane version (%s) (major.minor)", clusterVersion, controlPlaneVersion)
	}
	return nil
}

func parseVersions(semVersion string) (*semver.Version, error) {
	wrappedVer, err := semver.NewVersion(semVersion)
	if err != nil {
		return nil, fmt.Errorf("control plane version %s is not valid: %w", wrappedVer, err)
	}
	return wrappedVer, nil
}
