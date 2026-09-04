/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package provider

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"

	backendclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/clients/backend"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/database/models"
	tmclient "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client/generated"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/internal/generated/pkg/api"

	ext_api "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/api"
	errtypes "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/common/errors"
	idmclienttypes "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
)

const (
	ClusterMetricsRemoteWriteLabel = "runai"
)

type ClusterOAuthClientGetter interface {
	GetClient(ctx context.Context, clusterUUID string, clientType idmclienttypes.ClientType) (idmclienttypes.Client, error)
}

type ClusterAccessControlCreator interface {
	CreateAccessControl(ctx context.Context, tennatID int32, clusterUUID, clusterName string) error
}

type ClusterControlPlaneConfigProvider struct {
	clusterOAuthClientGetter    ClusterOAuthClientGetter
	clusterAccessControlCreator ClusterAccessControlCreator

	tenantSettingsClient backendclient.TenantSettingsClient
	tmClient             *tmclient.APIClient
}

func NewClusterControlPlaneConfigProvider(
	clusterOAuthClientGetter ClusterOAuthClientGetter,
	clusterAccessControlCreator ClusterAccessControlCreator,
	tenantSettingsClient backendclient.TenantSettingsClient,
	tmClient *tmclient.APIClient,
) *ClusterControlPlaneConfigProvider {
	return &ClusterControlPlaneConfigProvider{
		clusterOAuthClientGetter:    clusterOAuthClientGetter,
		clusterAccessControlCreator: clusterAccessControlCreator,
		tenantSettingsClient:        tenantSettingsClient,
		tmClient:                    tmClient,
	}
}

func (p *ClusterControlPlaneConfigProvider) GetClusterControlPlaneConfig(ctx context.Context, tenantID int32, clusterUUID string) (*api.ClusterControlPlaneConfig, error) {
	tenantReq := p.tmClient.TenantsAPI.GetTenantByIdInternal(ctx, tenantID)
	tenant, resp, err := p.tmClient.TenantsAPI.GetTenantByIdInternalExecute(tenantReq)
	if err != nil {
		if resp != nil && resp.StatusCode == http.StatusNotFound {
			return nil, errors.Join(errtypes.ErrNotFound, fmt.Errorf("tenant %d not found: %w", tenantID, err))
		}
		return nil, fmt.Errorf("failed to get tenant %d: %w", tenantID, err)
	}

	cluster, apiErr := dao.GetClusterByIdAndTenantId(ctx, clusterUUID, tenantID, ext_api.VerbosityMetadata)
	if apiErr != nil {
		if apiErr.Code == http.StatusNotFound {
			return nil, errors.Join(errtypes.ErrNotFound, fmt.Errorf("cluster %s not found: %v", clusterUUID, apiErr))
		}
		return nil, fmt.Errorf("failed to get cluster %s: %v", clusterUUID, apiErr)
	}

	agentConfig, err := p.getAgentConfig(ctx, tenant, cluster)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent config: %w", err)
	}

	workloadsAuthConfig, err := p.getWorkloadsAuthConfig(ctx, tenant, cluster)
	if err != nil {
		return nil, fmt.Errorf("failed to get workloads auth config: %w", err)
	}

	metricsRemoteWrite, err := p.getClusterRemoteWriteConfig(ctx, tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics remote write config: %w", err)
	}

	metricsRemoteWrites, err := p.getAllConfiguredRemoteWrites(ctx, tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics remote write config: %w", err)
	}

	researcherServiceConfig, err := p.getResearcherServiceConfig(ctx, tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to get researcher service config: %w", err)
	}

	dockerConfigJson, err := p.getDockerConfigJson(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get docker config json: %w", err)
	}

	installation := runai_config.Get[*config.Installation]()
	env := runai_config.Get[*config.Environment]()
	return &api.ClusterControlPlaneConfig{
		FipsMode:  env.FipsMode,
		Agent:     agentConfig,
		Workloads: workloadsAuthConfig,
		// Backward compatibility until all clusters are 2.18+
		MetricsRemoteWrite:  metricsRemoteWrite,
		MetricsRemoteWrites: metricsRemoteWrites,
		ResearcherService:   researcherServiceConfig,
		ClusterName:         cluster.Name,
		TenantId:            cluster.TenantId,
		Environment:         env.EnvIdentifier,
		ReleaseName:         env.ReleaseName,
		InstallationType:    installation.InstallationType,
		DockerConfigJson:    dockerConfigJson,
	}, nil
}

func (p *ClusterControlPlaneConfigProvider) getAgentConfig(ctx context.Context, tenant *tmclient.Tenant, cluster *models.Cluster) (api.ClusterControlPlaneConfigAgent, error) {
	agentClient, err := p.getOrCreateClusterClient(ctx, tenant, cluster, idmclienttypes.CLIENTTYPE_AGENT)
	if err != nil {
		return api.ClusterControlPlaneConfigAgent{}, fmt.Errorf("failed to get or create agent client: %w", err)
	}

	return api.ClusterControlPlaneConfigAgent{
		Client: api.ClusterControlPlaneConfigAgentClient{
			Id:     agentClient.Id,
			Name:   agentClient.ClientId,
			Secret: agentClient.Secret,
		},
	}, nil
}

func (p *ClusterControlPlaneConfigProvider) getWorkloadsAuthConfig(ctx context.Context, tenant *tmclient.Tenant, cluster *models.Cluster) (api.ClusterControlPlaneConfigWorkloads, error) {
	workloadClient, err := p.getOrCreateClusterClient(ctx, tenant, cluster, idmclienttypes.CLIENTTYPE_WORKLOADS_AUTH)
	if err != nil {
		return api.ClusterControlPlaneConfigWorkloads{}, fmt.Errorf("failed to get or create workloads auth client: %w", err)
	}

	return api.ClusterControlPlaneConfigWorkloads{
		Client: api.ClusterControlPlaneConfigWorkloadsClient{
			Id:     workloadClient.Id,
			Name:   workloadClient.ClientId,
			Secret: workloadClient.Secret,
		},
	}, nil
}

func (p *ClusterControlPlaneConfigProvider) getClusterRemoteWriteConfig(ctx context.Context, tenant *tmclient.Tenant) (api.ClusterControlPlaneConfigMetricsRemoteWrite, error) {
	metricStore, apiErr := p.tenantSettingsClient.GetTenantMetricStore(ctx, tenant.Id)
	if apiErr != nil {
		return api.ClusterControlPlaneConfigMetricsRemoteWrite{}, fmt.Errorf("failed to get tenant metric store: %v", apiErr)
	}

	details := metricStore.MetricStoreDetails.Get()
	if details == nil {
		return api.ClusterControlPlaneConfigMetricsRemoteWrite{}, nil
	}

	return api.ClusterControlPlaneConfigMetricsRemoteWrite{
		Url:   details.Write.Url,
		Label: ClusterMetricsRemoteWriteLabel,
		// backwards compatibility until all clusters are 2.18+
		OrgID:   ptr.Deref(tenant.MimirId, ""),
		Headers: ptr.Deref(details.Write.Headers, nil),
		Credentials: func() api.ClusterControlPlaneConfigMetricsRemoteWriteCredentials {
			if !details.Write.GetAuth().Basic.IsSet() {
				return api.ClusterControlPlaneConfigMetricsRemoteWriteCredentials{}
			}
			return api.ClusterControlPlaneConfigMetricsRemoteWriteCredentials{
				Username: details.Write.GetAuth().Basic.Get().Username,
				Password: details.Write.GetAuth().Basic.Get().Password,
			}
		}(),
	}, nil
}

func (p *ClusterControlPlaneConfigProvider) getResearcherServiceConfig(_ context.Context, tenant *tmclient.Tenant) (api.ClusterControlPlaneConfigResearcherService, error) {
	authConfig := runai_config.Get[*config.Auth]()
	return api.ClusterControlPlaneConfigResearcherService{
		JwksURL: fmt.Sprintf("https://%s/auth/realms/%s/protocol/openid-connect/certs", authConfig.KeycloakExternalEndpoint, tenant.Name),
		OauthProxy: api.ClusterControlPlaneConfigResearcherServiceOauthProxy{ // Required for backwards compatibility with clusters up to 2.18
			ClientId:      "dummy",
			ClientSecret:  "dummy",
			OidcIssuerUrl: fmt.Sprintf("https://%s/auth/realms/%s", authConfig.KeycloakExternalEndpoint, tenant.Name),
		},
	}, nil
}

// getDockerConfigJson returns the base64-encoded static image pull secret
// supplied via Helm (config). Self-hosted only: a single static value shared by
// all clusters, no per-tenant lookup. Returns "" when disabled; errors when
// enabled but no secret is configured (a misconfiguration).
func (p *ClusterControlPlaneConfigProvider) getDockerConfigJson(ctx context.Context) (string, error) {
	ipsConfig := runai_config.Get[*config.ImagePullSecret]()
	if !ipsConfig.Enabled {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Image pull secret is not enabled")
		return "", nil
	}
	if ipsConfig.StaticImagePullSecret == "" {
		return "", fmt.Errorf("image pull secret is enabled but no static image pull secret is configured")
	}
	return base64.StdEncoding.EncodeToString([]byte(ipsConfig.StaticImagePullSecret)), nil
}

func (p *ClusterControlPlaneConfigProvider) getOrCreateClusterClient(ctx context.Context, tenant *tmclient.Tenant, cluster *models.Cluster, clientType idmclienttypes.ClientType) (idmclienttypes.Client, error) {
	client, err := p.clusterOAuthClientGetter.GetClient(ctx, cluster.ID.String(), clientType)
	if err != nil {
		if !errors.Is(err, errtypes.ErrNotFound) {
			return idmclienttypes.Client{}, fmt.Errorf("failed to get client: %w", err)
		}

		err := p.clusterAccessControlCreator.CreateAccessControl(ctx, tenant.Id, cluster.ID.String(), cluster.Name)
		if err != nil {
			return idmclienttypes.Client{}, fmt.Errorf("failed to create access control: %w", err)
		}

		client, err = p.clusterOAuthClientGetter.GetClient(ctx, cluster.ID.String(), clientType)
		if err != nil {
			return idmclienttypes.Client{}, fmt.Errorf("failed to get client after creating access control: %w", err)
		}
	}

	return client, nil
}

func (p *ClusterControlPlaneConfigProvider) getAllConfiguredRemoteWrites(ctx context.Context, tenant *tmclient.Tenant) (remoteWrites []api.ClusterControlPlaneConfigMetricsRemoteWrite, err error) {
	mainRemoteWrite, err := p.getClusterRemoteWriteConfig(ctx, tenant)
	if err != nil {
		return nil, err
	}
	remoteWrites = append(remoteWrites, mainRemoteWrite)
	appMetricStore := runai_config.Get[*config.ApplicationMetricStore]()
	if appMetricStore.Url != "" {
		remoteWrites = append(remoteWrites, api.ClusterControlPlaneConfigMetricsRemoteWrite{
			Url:     appMetricStore.Url,
			Headers: appMetricStore.Headers,
			Label:   appMetricStore.Label,
			Credentials: api.ClusterControlPlaneConfigMetricsRemoteWriteCredentials{
				Username: appMetricStore.Username,
				Password: appMetricStore.Password,
			},
		})
	}
	return remoteWrites, nil
}
