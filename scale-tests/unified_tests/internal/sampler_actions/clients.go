/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package sampler_actions

import (
	"context"
	"fmt"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
	client_utils "github.com/dsx-ai-factory/run-ai/runai-application-packages/client-utils"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap/zapcore"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"time"

	clusterclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	idmclient "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
	orgunitclient "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
	workloadclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	runaiv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1"
	dwv2alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	twv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v1"
	twv2alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	"github.com/dsx-ai-factory/run-ai/scale-tests/unified_tests/internal/scale"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	kwokopv1beta1 "github.com/run-ai/kwok-operator/api/v1beta1"
	"golang.org/x/oauth2"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	k8sClient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	logLevelInfo zapcore.Level = 0
)

type Clients struct {
	KubeClient    k8sClient.Client
	DynamicClient dynamic.Interface

	ClusterServiceClient  *clusterclient.APIClient
	OrgUnitServiceClient  *orgunitclient.APIClient
	WorkloadServiceClient *workloadclient.APIClient
}
type TokenSource struct {
	tokenApi   idmclient.TokensAPI
	user       string
	password   string
	oauthToken *oauth2.Token
}

func getTokenExpiration(token *openapi.TokenResponse) (*time.Time, error) {
	claims := jwt.MapClaims{}
	_, _, err := jwt.NewParser().ParseUnverified(*token.AccessToken, claims)
	if err != nil {
		return nil, err
	}
	epoch := claims["exp"].(float64)
	timeObj := time.Unix(int64(epoch), 0)
	return ptr.To(timeObj), nil
}

func (ts *TokenSource) Token() (*oauth2.Token, error) {
	if ts.oauthToken == nil || ts.oauthToken.Expiry.Before(time.Now()) {
		request := idmclient.TokenRequest{
			Username:  ptr.To(ts.user),
			Password:  ptr.To(ts.password),
			GrantType: idmclient.TokenRequestGrantTypeEnum("password").Ptr(),
		}
		ctx := context.Background()
		ctx = context.WithValue(ctx, "user-agent", "runai-cli/0.10.0 sdk")
		token, _, err := ts.tokenApi.GrantToken(ctx).TokenRequest(request).Execute()
		if err != nil {
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}
		expiration, err := getTokenExpiration(token)
		if err != nil {
			return nil, fmt.Errorf("failed to get token expiration: %w", err)
		}
		ts.oauthToken = &oauth2.Token{
			AccessToken:  token.GetAccessToken(),
			RefreshToken: token.GetRefreshToken(),
			Expiry:       *expiration,
		}
	}
	return ts.oauthToken, nil
}

func InitKubeConfig(config *scale.ClusterConfig) *rest.Config {
	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&zap.Options{
		Level: logLevelInfo,
	})))
	conf := ctrl.GetConfigOrDie()
	conf.QPS = float32(config.KubeQPS)
	conf.Burst = 200

	return conf
}

func InitClients(ctx context.Context, config *scale.ClusterConfig, kubeConfig *rest.Config) (*Clients, error) {
	err := twv1.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, fmt.Errorf("failed to add workloadsv1 to the scheme: %w", err)
	}
	err = twv2alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, fmt.Errorf("failed to add workloadsv2alpha to the scheme: %w", err)
	}
	err = dwv2alpha1.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, fmt.Errorf("failed to add distributedworkloadsv2alpha1 to the scheme: %w", err)
	}
	err = runaiv1.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, fmt.Errorf("failed to add runaiv1 to the scheme: %w", err)
	}
	err = kwokopv1beta1.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, fmt.Errorf("failed to add kwokopv1beta1 to the scheme: %w", err)
	}
	err = monitoringv1.AddToScheme(scheme.Scheme)
	if err != nil {
		return nil, fmt.Errorf("failed to add monitoringv1 to the scheme: %w", err)
	}
	dynamicClient := dynamic.NewForConfigOrDie(kubeConfig)
	kubeClient, err := k8sClient.New(kubeConfig, k8sClient.Options{Scheme: scheme.Scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	tokenApi := newIdmClient(config.ControlPlaneURL).TokensAPI
	tokenSource := &TokenSource{
		tokenApi: tokenApi,
		user:     config.ControlPlaneUsername,
		password: config.ControlPlanePassword,
	}

	client := oauth2.NewClient(ctx, tokenSource)
	clusterServiceClient := clusterclient.NewAPIClient(&clusterclient.Configuration{
		HTTPClient: client,
		Servers:    []clusterclient.ServerConfiguration{{URL: config.ControlPlaneURL}},
	})
	orgUnitServiceClient := orgunitclient.NewAPIClient(&orgunitclient.Configuration{
		HTTPClient: client,
		Servers:    []orgunitclient.ServerConfiguration{{URL: config.ControlPlaneURL}},
	})

	workloadServiceClient := workloadclient.NewAPIClient(&workloadclient.Configuration{
		HTTPClient: client,
		Servers:    []workloadclient.ServerConfiguration{{URL: config.ControlPlaneURL}},
	})
	return &Clients{
		KubeClient:            kubeClient,
		DynamicClient:         dynamicClient,
		OrgUnitServiceClient:  orgUnitServiceClient,
		ClusterServiceClient:  clusterServiceClient,
		WorkloadServiceClient: workloadServiceClient,
	}, nil
}

func newIdmClient(baseURL string) *idmclient.APIClient {
	baseURL = client_utils.GetInternalUrl(baseURL)
	config := openapi.NewConfiguration()
	config.Servers = openapi.ServerConfigurations{{URL: baseURL}}
	return idmclient.NewAPIClient(config)
}
