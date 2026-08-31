/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/run-ai/runai/control-plane/catalog-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/catalog-service/pkg/database/dao"
	"github.com/run-ai/runai/runai-common-packages/cache/v2"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"gopkg.in/yaml.v3"
)

const (
	helmRepoChartsCacheKeyPrefix = "helm_repo_charts"
	helmRepoFetchTimeout         = 30 * time.Second
)

type helmRepoChartsCache struct {
	Charts []api.HelmChart
}

type helmIndexFile struct {
	APIVersion string                          `yaml:"apiVersion"`
	Entries    map[string][]helmIndexChartMeta `yaml:"entries"`
}

type helmIndexChartMeta struct {
	Name        string    `yaml:"name"`
	Version     string    `yaml:"version"`
	AppVersion  string    `yaml:"appVersion"`
	Description string    `yaml:"description"`
	Created     time.Time `yaml:"created"`
}

func chartsCacheKey(repoId string, assetId string) string {
	if assetId == "" {
		return fmt.Sprintf("%s_%s", helmRepoChartsCacheKeyPrefix, repoId)
	}
	sum := sha256.Sum256([]byte(assetId))
	return fmt.Sprintf("%s_%s_%s", helmRepoChartsCacheKeyPrefix, repoId, hex.EncodeToString(sum[:]))
}

func (s *helmReposAPIService) getHelmRepoAPIKey(ctx context.Context, assetId string) (string, *api.Error) {
	if assetId == "" {
		return "", nil
	}
	if s.assetsClient == nil {
		return "", api.NewError(http.StatusInternalServerError, "assets client is not configured", nil)
	}
	ngcApiKeyInternal, err := s.assetsClient.GetNgcApiKeyByIdInternal(ctx, assetId)
	if err != nil {
		var details error
		if err.Details != nil {
			details = fmt.Errorf("%s", *err.Details)
		}
		return "", api.NewError(int(err.Code), err.Message, details)
	}
	return ngcApiKeyInternal.Key, nil
}

func (s *helmReposAPIService) BrowseHelmRepoCharts(ctx context.Context, helmRepoId string, assetId string) (api.ImplResponse, error) {
	helmRepo, apiErr := dao.GetHelmRepoById(ctx, helmRepoId)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: get helm repo for browse: %v", apiErr.Message, apiErr.Details)
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	if helmRepo.IsPrivate && assetId == "" {
		apiErr := api.NewValidationError("assetId is required for private helm repos")
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	apiKey, apiErr := s.getHelmRepoAPIKey(ctx, assetId)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: get helm repo api key: %v", apiErr.Message, apiErr.Details)
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	cacheKey := chartsCacheKey(helmRepoId, assetId)
	cachedResult, err := cache.Get[helmRepoChartsCache](ctx, cacheKey)
	if err == nil && cachedResult != nil {
		return api.Response(http.StatusOK, api.HelmChartListResponse{
			Charts: cachedResult.Charts,
		}), nil
	}

	charts, apiErr := fetchAndParseHelmIndex(ctx, helmRepo.Url, apiKey)
	if apiErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s: browse helm repo charts: %v", apiErr.Message, apiErr.Details)
		return api.Response(int(apiErr.Code), apiErr), nil
	}

	if cacheErr := cache.Set[helmRepoChartsCache](ctx, cacheKey, helmRepoChartsCache{Charts: charts}); cacheErr != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to cache helm repo charts: %v", cacheErr)
	}

	return api.Response(http.StatusOK, api.HelmChartListResponse{
		Charts: charts,
	}), nil
}

func isPrivateIP(ip net.IP) bool {
	privateRanges := []struct {
		network *net.IPNet
	}{
		{network: mustParseCIDR("10.0.0.0/8")},
		{network: mustParseCIDR("172.16.0.0/12")},
		{network: mustParseCIDR("192.168.0.0/16")},
		{network: mustParseCIDR("127.0.0.0/8")},
		{network: mustParseCIDR("169.254.0.0/16")},
		{network: mustParseCIDR("::1/128")},
		{network: mustParseCIDR("fc00::/7")},
		{network: mustParseCIDR("fe80::/10")},
	}
	for _, r := range privateRanges {
		if r.network.Contains(ip) {
			return true
		}
	}
	return false
}

func mustParseCIDR(s string) *net.IPNet {
	_, network, err := net.ParseCIDR(s)
	if err != nil {
		panic(fmt.Sprintf("invalid CIDR %q: %v", s, err))
	}
	return network
}

func validateHelmRepoURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("malformed URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("unsupported scheme %q, only http and https are allowed", u.Scheme)
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("empty host in URL")
	}
	if ip := net.ParseIP(host); ip != nil && isPrivateIP(ip) {
		return fmt.Errorf("URL must not point to a private/internal address (%s)", host)
	}
	return nil
}

func resolvedHostIsPrivate(ctx context.Context, rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	host := u.Hostname()
	ips, err := net.DefaultResolver.LookupHost(ctx, host)
	if err != nil {
		return fmt.Errorf("failed to resolve host %q: %w", host, err)
	}
	for _, ipStr := range ips {
		if ip := net.ParseIP(ipStr); ip != nil && isPrivateIP(ip) {
			return fmt.Errorf("host %q resolves to private/internal address %s", host, ipStr)
		}
	}
	return nil
}

func hardenedHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			if err := validateHelmRepoURL(req.URL.String()); err != nil {
				return fmt.Errorf("redirect blocked: %w", err)
			}
			if err := resolvedHostIsPrivate(req.Context(), req.URL.String()); err != nil {
				return fmt.Errorf("redirect blocked: %w", err)
			}
			return nil
		},
	}
}

func fetchAndParseHelmIndex(ctx context.Context, repoUrl string, apiKey string) ([]api.HelmChart, *api.Error) {
	indexUrl := strings.TrimRight(repoUrl, "/") + "/index.yaml"

	if err := validateHelmRepoURL(indexUrl); err != nil {
		return nil, api.NewValidationError(fmt.Sprintf("invalid helm repo URL: %v", err))
	}
	if err := resolvedHostIsPrivate(ctx, indexUrl); err != nil {
		return nil, api.NewError(http.StatusBadRequest, "helm repo URL resolves to a blocked address", err)
	}

	client := hardenedHTTPClient(helmRepoFetchTimeout)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, indexUrl, nil)
	if err != nil {
		return nil, api.NewError(http.StatusInternalServerError, fmt.Sprintf("failed to create request for %s", indexUrl), err)
	}
	if apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, api.NewError(http.StatusBadGateway, fmt.Sprintf("failed to fetch helm repo index from %s", indexUrl), err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, api.NewError(http.StatusBadGateway, fmt.Sprintf("helm repo returned status %d for %s", resp.StatusCode, indexUrl), nil)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, api.NewError(http.StatusInternalServerError, "failed to read helm repo index response", err)
	}

	var indexFile helmIndexFile
	if err := yaml.Unmarshal(body, &indexFile); err != nil {
		return nil, api.NewError(http.StatusBadGateway, "failed to parse helm repo index.yaml", err)
	}

	charts := make([]api.HelmChart, 0, len(indexFile.Entries))
	for chartName, versions := range indexFile.Entries {
		apiVersions := make([]api.HelmChartVersion, 0, len(versions))
		for _, v := range versions {
			apiVersions = append(apiVersions, api.HelmChartVersion{
				Version:     v.Version,
				AppVersion:  v.AppVersion,
				Description: v.Description,
				Created:     v.Created,
			})
		}
		charts = append(charts, api.HelmChart{
			Name:     chartName,
			Versions: apiVersions,
		})
	}

	sort.Slice(charts, func(i, j int) bool {
		return charts[i].Name < charts[j].Name
	})

	return charts, nil
}
