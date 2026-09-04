/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workload_submitter

import (
	"context"
	"fmt"
	"time"

	clusterclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	idmclient "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
	orgunitclient "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
	workloadclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
	"k8s.io/client-go/dynamic"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Clients struct {
	DynamicClient         dynamic.Interface
	ClusterServiceClient  *clusterclient.APIClient
	OrgUnitServiceClient  *orgunitclient.APIClient
	WorkloadServiceClient *workloadclient.APIClient
}

type tokenSource struct {
	tokenAPI idmclient.TokensAPI
	user     string
	password string
	token    *oauth2.Token
}

func (ts *tokenSource) Token() (*oauth2.Token, error) {
	if ts.token == nil || ts.token.Expiry.Before(time.Now()) {
		req := idmclient.TokenRequest{
			Username:  ptr.To(ts.user),
			Password:  ptr.To(ts.password),
			GrantType: idmclient.TokenRequestGrantTypeEnum("password").Ptr(),
		}
		ctx := context.Background()
		//nolint:staticcheck // idmclient reads this key by convention
		ctx = context.WithValue(ctx, "user-agent", "runai-cli/0.10.0 sdk")
		tokenResp, _, err := ts.tokenAPI.GrantToken(ctx).TokenRequest(req).Execute()
		if err != nil {
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}
		expiry, err := jwtExpiry(tokenResp.GetAccessToken())
		if err != nil {
			return nil, fmt.Errorf("failed to parse token expiry: %w", err)
		}
		ts.token = &oauth2.Token{
			AccessToken:  tokenResp.GetAccessToken(),
			RefreshToken: tokenResp.GetRefreshToken(),
			Expiry:       expiry,
		}
	}
	return ts.token, nil
}

func jwtExpiry(raw string) (time.Time, error) {
	claims := jwt.MapClaims{}
	if _, _, err := jwt.NewParser().ParseUnverified(raw, claims); err != nil {
		return time.Time{}, err
	}
	epoch, ok := claims["exp"].(float64)
	if !ok {
		return time.Time{}, fmt.Errorf("exp claim missing or not a number")
	}
	return time.Unix(int64(epoch), 0), nil
}

func InitClients(ctx context.Context, cfg *Config) (*Clients, error) {
	kubeConfig := ctrl.GetConfigOrDie()
	kubeConfig.QPS = 50
	kubeConfig.Burst = 100

	dynamicClient, err := dynamic.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	idmCfg := idmclient.NewConfiguration()
	idmCfg.Servers = idmclient.ServerConfigurations{{URL: cfg.ControlPlaneURL}}
	idmAPI := idmclient.NewAPIClient(idmCfg)

	ts := &tokenSource{
		tokenAPI: idmAPI.TokensAPI,
		user:     cfg.ControlPlaneUsername,
		password: cfg.ControlPlanePassword,
	}
	httpClient := oauth2.NewClient(ctx, ts)

	clusterClient := clusterclient.NewAPIClient(&clusterclient.Configuration{
		HTTPClient: httpClient,
		Servers:    []clusterclient.ServerConfiguration{{URL: cfg.ControlPlaneURL}},
	})
	orgUnitClient := orgunitclient.NewAPIClient(&orgunitclient.Configuration{
		HTTPClient: httpClient,
		Servers:    []orgunitclient.ServerConfiguration{{URL: cfg.ControlPlaneURL}},
	})
	workloadClient := workloadclient.NewAPIClient(&workloadclient.Configuration{
		HTTPClient: httpClient,
		Servers:    []workloadclient.ServerConfiguration{{URL: cfg.ControlPlaneURL}},
	})

	return &Clients{
		DynamicClient:         dynamicClient,
		ClusterServiceClient:  clusterClient,
		OrgUnitServiceClient:  orgUnitClient,
		WorkloadServiceClient: workloadClient,
	}, nil
}
