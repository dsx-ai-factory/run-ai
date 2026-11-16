/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package config

import (
	"fmt"
	"net/http"
	"net/url"
	"runtime"

	catalog "github.com/dsx-ai-factory/run-ai/runai-clients/catalog-service-client/generated"

	k8sobjtrack "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"

	cliVersion "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/version"

	clusterservice "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	loginapi "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
	orgunit "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	workloadservice "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	assetsservice "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/cli-exposer-client/generated"
	clusterapi "github.com/dsx-ai-factory/run-ai/runai-clients/cluster-api-client/generated"
	datavolumesservice "github.com/dsx-ai-factory/run-ai/runai-clients/datavolumes-service-client/generated"
	workloadmanager "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
)

const (
	AuthorizationHeader = "Authorization"
	BearerTokenStr      = "Bearer %s"
	UserAgentStr        = "runai-cli/%s sdk/%s go/%s %s/%s"
)

func CreateClusterConfig(urlStr, token string) (*clusterservice.Configuration, error) {
	serviceUrl, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	configuration := clusterservice.NewConfiguration()
	configuration.Scheme = serviceUrl.Scheme
	configuration.Host = serviceUrl.Host
	configuration.UserAgent = GetUserAgent()
	configuration.AddDefaultHeader(AuthorizationHeader, fmt.Sprintf(BearerTokenStr, token))
	return configuration, nil
}

func CreateDepartmentConfig(urlStr, token string) (*orgunit.Configuration, error) {
	serviceUrl, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	configuration := orgunit.NewConfiguration()
	configuration.Scheme = serviceUrl.Scheme
	configuration.Host = serviceUrl.Host
	configuration.UserAgent = GetUserAgent()
	configuration.AddDefaultHeader(AuthorizationHeader, fmt.Sprintf(BearerTokenStr, token))
	return configuration, nil
}

func CreateProjectConfig(urlStr, token string) (*orgunit.Configuration, error) {
	serviceUrl, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	configuration := orgunit.NewConfiguration()
	configuration.Scheme = serviceUrl.Scheme
	configuration.Host = serviceUrl.Host
	configuration.UserAgent = GetUserAgent()
	configuration.AddDefaultHeader(AuthorizationHeader, fmt.Sprintf(BearerTokenStr, token))
	return configuration, nil
}

func CreateWorkloadServiceConfig(urlStr, token string) (*workloadservice.Configuration, error) {
	serviceUrl, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	configuration := workloadservice.NewConfiguration()
	configuration.Scheme = serviceUrl.Scheme
	configuration.Host = serviceUrl.Host
	configuration.UserAgent = GetUserAgent()
	configuration.AddDefaultHeader(AuthorizationHeader, fmt.Sprintf(BearerTokenStr, token))
	return configuration, nil
}

func CreateLoginClientConfig(urlStr string) (*loginapi.Configuration, error) {
	serviceUrl, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	configuration := loginapi.NewConfiguration()
	configuration.Scheme = serviceUrl.Scheme
	configuration.Host = serviceUrl.Host
	configuration.UserAgent = GetUserAgent()
	nonFollowClient := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	configuration.HTTPClient = &nonFollowClient
	return configuration, nil
}

func CreateExposerConfig(urlStr, token string) (*openapi.Configuration, error) {
	serviceUrl, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	configuration := openapi.NewConfiguration()
	configuration.Scheme = serviceUrl.Scheme
	configuration.Host = serviceUrl.Host
	configuration.UserAgent = GetUserAgent()
	configuration.AddDefaultHeader(AuthorizationHeader, fmt.Sprintf(BearerTokenStr, token))
	return configuration, nil
}

func CreateClusterApiConfig(urlStr, token string) (*clusterapi.Configuration, error) {
	serviceUrl, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	configuration := clusterapi.NewConfiguration()
	configuration.Scheme = serviceUrl.Scheme
	configuration.Host = serviceUrl.Host
	configuration.UserAgent = GetUserAgent()
	configuration.AddDefaultHeader(AuthorizationHeader, fmt.Sprintf(BearerTokenStr, token))
	return configuration, nil
}

func CreateObjTrackerApiConfig(urlStr, token string) (*k8sobjtrack.Configuration, error) {
	serviceUrl, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	configuration := k8sobjtrack.NewConfiguration()
	configuration.Scheme = serviceUrl.Scheme
	configuration.Host = serviceUrl.Host
	configuration.UserAgent = GetUserAgent()
	configuration.AddDefaultHeader(AuthorizationHeader, fmt.Sprintf(BearerTokenStr, token))
	return configuration, nil
}

func CreateWorkloadManagerServiceConfig(urlStr, token string) (*workloadmanager.Configuration, error) {
	serviceUrl, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	configuration := workloadmanager.NewConfiguration()
	configuration.Scheme = serviceUrl.Scheme
	configuration.Host = serviceUrl.Host
	configuration.Servers = workloadmanager.ServerConfigurations{{URL: urlStr}}
	configuration.UserAgent = GetUserAgent()
	configuration.AddDefaultHeader(AuthorizationHeader, fmt.Sprintf(BearerTokenStr, token))
	return configuration, nil
}

func CreateAssetsServiceConfig(urlStr, token string) (*assetsservice.Configuration, error) {
	serviceUrl, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	configuration := assetsservice.NewConfiguration()
	configuration.Scheme = serviceUrl.Scheme
	configuration.Host = serviceUrl.Host
	configuration.UserAgent = GetUserAgent()
	configuration.AddDefaultHeader(AuthorizationHeader, fmt.Sprintf(BearerTokenStr, token))
	return configuration, nil
}

func CreateDatavolumesServiceConfig(urlStr, token string) (*datavolumesservice.Configuration, error) {
	serviceUrl, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	configuration := datavolumesservice.NewConfiguration()
	configuration.Scheme = serviceUrl.Scheme
	configuration.Host = serviceUrl.Host
	configuration.UserAgent = GetUserAgent()
	configuration.AddDefaultHeader(AuthorizationHeader, fmt.Sprintf(BearerTokenStr, token))
	return configuration, nil
}

func GetUserAgent() string {
	currentVersion := cliVersion.GetVersion()
	return fmt.Sprintf(UserAgentStr, currentVersion.ControlPlaneVersion, currentVersion.Version,
		currentVersion.GoVersion, runtime.GOOS, runtime.GOARCH)
}

func CreateCatalogConfig(urlStr, token string) (*catalog.Configuration, error) {
	serviceURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	configuration := catalog.NewConfiguration()
	configuration.Scheme = serviceURL.Scheme
	configuration.Host = serviceURL.Host
	configuration.UserAgent = GetUserAgent()
	configuration.AddDefaultHeader(AuthorizationHeader, fmt.Sprintf(BearerTokenStr, token))
	return configuration, nil
}
