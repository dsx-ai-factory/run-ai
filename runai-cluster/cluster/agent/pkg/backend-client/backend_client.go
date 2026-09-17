/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package backendclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"

	cluster_service_client "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
	cluster_service_openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"

	auth_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	auth_openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	assets_client "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client"
	assets_openapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	org_unit_client "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client"
	org_unit_openapi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	policy_client "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client"
	policy_openapi "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
	workloads_openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	workloadsMgrApi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/generated/datavolumes_openapi"
	"github.com/rs/zerolog/log"
	"go.uber.org/zap"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/config"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/generated/backend_openapi"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/agent/pkg/generated/projects_openapi"
)

// BackendClient is the client for the runai backend API.
type BackendClient struct {
	Client               *backend_openapi.APIClient
	ProjectsClient       *projects_openapi.APIClient
	PoliciesClient       policy_client.IPolicyExternalClient
	WorkloadsClient      *workloads_openapi.APIClient
	WorkloadsMgrClient   *workloadsMgrApi.APIClient
	AuthClient           auth_client.IAuthorizationExternalClient
	DatavolumesClient    *datavolumes_openapi.APIClient
	OrgUnitClient        org_unit_client.IOrgUnitExternalClient
	AssetsServiceClient  assets_client.IAssetsExternalClient
	ClusterServiceClient cluster_service_client.IClusterExternalClient
}

type Options struct {
	UseDebugLogLevel bool
	HttpClient       *http.Client
}

var runaiBackendClient *BackendClient

func NewBackendClient(basePath string, options *Options) *BackendClient {
	runaiBackendClient = &BackendClient{
		Client:               GetBackendOpenapiClient(basePath, options),
		ProjectsClient:       GetProjectsOpenapiClient(basePath, options),
		PoliciesClient:       GetPoliciesOpenapiClient(basePath, options),
		WorkloadsClient:      GetWorkloadOpenapiClient(basePath, options),
		WorkloadsMgrClient:   GetWorkloadsMgrOpenapiClient(basePath, options),
		AssetsServiceClient:  GetAssetsOpenapiClient(basePath, options),
		AuthClient:           GetAuthOpenapiClient(basePath, options),
		DatavolumesClient:    GetDatavolumesOpenapiClient(basePath, options),
		OrgUnitClient:        GetOrgUnitsOpenapiClient(basePath, options),
		ClusterServiceClient: GetClusterServiceOpenapiClient(basePath, options),
	}

	//This disables the authentication of certificates in all http requests. This is insecure - do this for local-env only.
	if !config.Config().TLSRejectUnauthorized {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return runaiBackendClient
}

func (backendClient *BackendClient) Context() context.Context {
	return context.Background()
}

func (backendClient *BackendClient) HandleHttpResponse(res *http.Response, err error, baseMsg string) error {
	httpError := err
	if err != nil {
		if res != nil {
			return fmt.Errorf("Had error while sending http request.\n URL: %s.\n Status: %s.\n Error: %v\n", res.Request.URL, res.Status, err) //nolint:staticcheck
		}

		return err
	}

	if res != nil && res.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("Had error while sending http request.\n URL: %s.\n Status: %s\n", res.Request.URL, res.Status) //nolint:staticcheck
	}

	log.Debug().Msgf("%s successfully.", baseMsg)
	return httpError
}

func GetBackendOpenapiClient(basePath string, options *Options) *backend_openapi.APIClient {
	apiConfiguration := backend_openapi.NewConfiguration()
	apiConfiguration.Host = GetHostFromDomain(basePath)
	apiConfiguration.Servers = backend_openapi.ServerConfigurations{{URL: basePath}}
	apiConfiguration.Debug = options.UseDebugLogLevel
	if options.HttpClient != nil {
		apiConfiguration.HTTPClient = options.HttpClient
	}
	client := backend_openapi.NewAPIClient(apiConfiguration)
	return client
}

func GetPoliciesOpenapiClient(basePath string, options *Options) policy_client.IPolicyExternalClient {
	apiConfiguration := policy_openapi.NewConfiguration()
	apiConfiguration.Host = GetHostFromDomain(basePath)
	apiConfiguration.Servers = policy_openapi.ServerConfigurations{{URL: basePath}}
	apiConfiguration.Debug = options.UseDebugLogLevel
	if options.HttpClient != nil {
		apiConfiguration.HTTPClient = options.HttpClient
	}
	return policy_client.NewPoliciesExternalClientWithConfig(apiConfiguration)
}

func GetOrgUnitsOpenapiClient(basePath string, options *Options) org_unit_client.IOrgUnitExternalClient {
	apiConfiguration := org_unit_openapi.NewConfiguration()
	apiConfiguration.Host = GetHostFromDomain(basePath)
	apiConfiguration.Servers = org_unit_openapi.ServerConfigurations{{URL: basePath}}
	apiConfiguration.Debug = options.UseDebugLogLevel
	if options.HttpClient != nil {
		apiConfiguration.HTTPClient = options.HttpClient
	}
	return org_unit_client.NewOrgUnitExternalClientWithConfig(apiConfiguration)
}

func GetAssetsOpenapiClient(basePath string, options *Options) assets_client.IAssetsExternalClient {
	apiConfiguration := assets_openapi.NewConfiguration()
	apiConfiguration.Host = GetHostFromDomain(basePath)
	apiConfiguration.Servers = assets_openapi.ServerConfigurations{{URL: basePath}}
	apiConfiguration.Debug = options.UseDebugLogLevel
	if options.HttpClient != nil {
		apiConfiguration.HTTPClient = options.HttpClient
	}
	client := assets_client.NewExternalClientWithConfig(apiConfiguration)
	return client
}

func GetWorkloadOpenapiClient(basePath string, options *Options) *workloads_openapi.APIClient {
	apiConfiguration := workloads_openapi.NewConfiguration()
	apiConfiguration.Host = GetHostFromDomain(basePath)
	apiConfiguration.Servers = workloads_openapi.ServerConfigurations{{URL: basePath}}
	apiConfiguration.Debug = options.UseDebugLogLevel
	if options.HttpClient != nil {
		apiConfiguration.HTTPClient = options.HttpClient
	}
	client := workloads_openapi.NewAPIClient(apiConfiguration)
	return client
}

func GetWorkloadsMgrOpenapiClient(basePath string, options *Options) *workloadsMgrApi.APIClient {
	apiConfiguration := workloadsMgrApi.NewConfiguration()
	apiConfiguration.Host = GetHostFromDomain(basePath)
	apiConfiguration.Servers = workloadsMgrApi.ServerConfigurations{{URL: basePath}}
	apiConfiguration.Debug = options.UseDebugLogLevel
	if options.HttpClient != nil {
		apiConfiguration.HTTPClient = options.HttpClient
	}
	client := workloadsMgrApi.NewAPIClient(apiConfiguration)
	return client
}

func GetClusterServiceOpenapiClient(basePath string, options *Options) cluster_service_client.IClusterExternalClient {
	apiConfiguration := cluster_service_openapi.NewConfiguration()
	apiConfiguration.Host = GetHostFromDomain(basePath)
	apiConfiguration.Servers = cluster_service_openapi.ServerConfigurations{{URL: basePath}}
	apiConfiguration.Debug = options.UseDebugLogLevel
	if options.HttpClient != nil {
		apiConfiguration.HTTPClient = options.HttpClient
	}
	return cluster_service_client.NewClusterServiceExternalClientWithConfig(apiConfiguration)
}

func GetProjectsOpenapiClient(basePath string, options *Options) *projects_openapi.APIClient {
	apiConfiguration := projects_openapi.NewConfiguration()
	apiConfiguration.Host = GetHostFromDomain(basePath)
	apiConfiguration.Servers = projects_openapi.ServerConfigurations{{URL: basePath}}
	apiConfiguration.Debug = options.UseDebugLogLevel
	if options.HttpClient != nil {
		apiConfiguration.HTTPClient = options.HttpClient
	}
	client := projects_openapi.NewAPIClient(apiConfiguration)
	return client
}

func GetAuthOpenapiClient(basePath string, options *Options) auth_client.IAuthorizationExternalClient {
	apiConfiguration := auth_openapi.NewConfiguration()
	apiConfiguration.Host = GetHostFromDomain(basePath)
	apiConfiguration.Servers = auth_openapi.ServerConfigurations{{URL: basePath}}
	apiConfiguration.Debug = options.UseDebugLogLevel
	if options.HttpClient != nil {
		apiConfiguration.HTTPClient = options.HttpClient
	}
	client := auth_client.NewAuthorizationExternalClientWithConfig(apiConfiguration)
	return client
}

func GetDatavolumesOpenapiClient(basePath string, options *Options) *datavolumes_openapi.APIClient {
	apiConfiguration := datavolumes_openapi.NewConfiguration()
	apiConfiguration.Host = GetHostFromDomain(basePath)
	apiConfiguration.Servers = datavolumes_openapi.ServerConfigurations{{URL: basePath}}
	apiConfiguration.Debug = options.UseDebugLogLevel
	if options.HttpClient != nil {
		apiConfiguration.HTTPClient = options.HttpClient
	}
	client := datavolumes_openapi.NewAPIClient(apiConfiguration)
	return client
}

func GetHostFromDomain(runaiHost string) string {
	u, err := url.Parse(runaiHost)
	if err != nil {
		zap.S().Error("failed to parse runaiHost", err)
		return ""
	}
	return u.Host
}
