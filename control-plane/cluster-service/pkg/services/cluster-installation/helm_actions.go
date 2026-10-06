/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster_installation

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"gopkg.in/yaml.v3"
)

type VersionsObject struct {
	Entries map[string][]struct {
		Version string `yaml:"version"`
	} `yaml:"entries"`
}

type VersionsList []struct {
	Version string `yaml:"version"`
}

type HelmActions interface {
	GetVersionListFromHelm(ctx context.Context, url string) (VersionsList, error)
}

type HelmActionsImpl struct{}

func NewHelmActions() HelmActions {
	return &HelmActionsImpl{}
}

func (h *HelmActionsImpl) GetVersionListFromHelm(ctx context.Context, url string) (VersionsList, error) {
	body, err := h.getIndexYaml(ctx, url)
	if err != nil {
		return nil, err
	}

	var versionsObject VersionsObject
	err = yaml.Unmarshal(body, &versionsObject)
	if err != nil {
		return nil, err
	}
	versionsList, ok := versionsObject.Entries["runai-cluster"]
	if !ok {
		return nil, err
	}
	return versionsList, nil
}

func (h *HelmActionsImpl) getIndexYaml(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/index.yaml", url), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create fetch request for index.yaml - %s", err.Error())
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch index.yaml from registry - %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch index.yaml from registry, status code - %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading cluster version response body - %s", err.Error())
	}
	return body, nil
}
