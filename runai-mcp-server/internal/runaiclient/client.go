/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package runaiclient

import (
	"context"
	"net/http"
	"strings"
	"time"

	authz "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	bff "github.com/run-ai/runai/control-plane/bff-service/pkg/client/generated"
	clusterservice "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	identity "github.com/run-ai/runai/control-plane/identity-manager/pkg/client/generated"
	assets "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
	catalog "github.com/run-ai/runai/runai-clients/catalog-service-client/generated"
	datavolumes "github.com/run-ai/runai/runai-clients/datavolumes-service-client/generated"
	orgunit "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	policy "github.com/run-ai/runai/runai-clients/policy-service-client/generated"
	workloads "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	workloadsv2 "github.com/run-ai/runai/runai-clients/workloads-manager-client/generated"

	"github.com/run-ai/runai/runai-mcp-server/internal/auth"
)

const httpClientTimeout = 30 * time.Second

type Client struct {
	baseURL    string
	auth       auth.Provider
	httpClient *http.Client

	Authorization *authz.APIClient
	Assets        *assets.APIClient
	BFF           *bff.APIClient
	Catalog       *catalog.APIClient
	Cluster       *clusterservice.APIClient
	DataVolumes   *datavolumes.APIClient
	Identity      *identity.APIClient
	OrgUnit       *orgunit.APIClient
	Policy        *policy.APIClient
	Workloads     *workloads.APIClient
	WorkloadsV2   *workloadsv2.APIClient
}

func (c *Client) Token(ctx context.Context) (string, error) {
	return c.auth.Token(ctx)
}

func New(baseURL string, authProvider auth.Provider) *Client {
	baseURL = strings.TrimRight(baseURL, "/")
	httpClient := authProvider.HttpClient()
	httpClient.Timeout = httpClientTimeout

	bffCfg := bff.NewConfiguration()
	bffCfg.HTTPClient = httpClient
	bffCfg.Servers = bff.ServerConfigurations{{URL: baseURL}}

	orgUnitCfg := orgunit.NewConfiguration()
	orgUnitCfg.HTTPClient = httpClient
	orgUnitCfg.Servers = orgunit.ServerConfigurations{{URL: baseURL}}

	wlCfg := workloads.NewConfiguration()
	wlCfg.HTTPClient = httpClient
	wlCfg.Servers = workloads.ServerConfigurations{{URL: baseURL}}

	wlV2Cfg := workloadsv2.NewConfiguration()
	wlV2Cfg.HTTPClient = httpClient
	wlV2Cfg.Servers = workloadsv2.ServerConfigurations{{URL: baseURL}}

	clusterCfg := clusterservice.NewConfiguration()
	clusterCfg.HTTPClient = httpClient
	clusterCfg.Servers = clusterservice.ServerConfigurations{{URL: baseURL}}

	policyCfg := policy.NewConfiguration()
	policyCfg.HTTPClient = httpClient
	policyCfg.Servers = policy.ServerConfigurations{{URL: baseURL}}

	authzCfg := authz.NewConfiguration()
	authzCfg.HTTPClient = httpClient
	authzCfg.Servers = authz.ServerConfigurations{{URL: baseURL}}

	assetsCfg := assets.NewConfiguration()
	assetsCfg.HTTPClient = httpClient
	assetsCfg.Servers = assets.ServerConfigurations{{URL: baseURL}}

	identityCfg := identity.NewConfiguration()
	identityCfg.HTTPClient = httpClient
	identityCfg.Servers = identity.ServerConfigurations{{URL: baseURL}}

	catalogCfg := catalog.NewConfiguration()
	catalogCfg.HTTPClient = httpClient
	catalogCfg.Servers = catalog.ServerConfigurations{{URL: baseURL}}

	dataVolumesCfg := datavolumes.NewConfiguration()
	dataVolumesCfg.HTTPClient = httpClient
	dataVolumesCfg.Servers = datavolumes.ServerConfigurations{{URL: baseURL}}

	return &Client{
		baseURL:       baseURL,
		auth:          authProvider,
		httpClient:    httpClient,
		Authorization: authz.NewAPIClient(authzCfg),
		Assets:        assets.NewAPIClient(assetsCfg),
		BFF:           bff.NewAPIClient(bffCfg),
		Catalog:       catalog.NewAPIClient(catalogCfg),
		Cluster:       clusterservice.NewAPIClient(clusterCfg),
		DataVolumes:   datavolumes.NewAPIClient(dataVolumesCfg),
		Identity:      identity.NewAPIClient(identityCfg),
		OrgUnit:       orgunit.NewAPIClient(orgUnitCfg),
		Policy:        policy.NewAPIClient(policyCfg),
		Workloads:     workloads.NewAPIClient(wlCfg),
		WorkloadsV2:   workloadsv2.NewAPIClient(wlV2Cfg),
	}
}
