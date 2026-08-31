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

	k8sobjtrack "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client/generated"

	cliVersion "github.com/run-ai/runai/control-plane/runai-cli/pkg/version"

	clusterservice "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	loginapi "github.com/run-ai/runai/control-plane/identity-manager/pkg/client/generated"
	assetsservice "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
	openapi "github.com/run-ai/runai/runai-clients/cli-exposer-client/generated"
	clusterapi "github.com/run-ai/runai/runai-clients/cluster-api-client/generated"
	datavolumesservice "github.com/run-ai/runai/runai-clients/datavolumes-service-client/generated"
	nimclient "github.com/run-ai/runai/runai-clients/nim-client/generated"
	orgunit "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	workloadservice "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	workloadmanager "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"
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

func CreateNimClientConfig(urlStr, token string) (*nimclient.Configuration, error) {
	serviceUrl, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	configuration := nimclient.NewConfiguration()
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
