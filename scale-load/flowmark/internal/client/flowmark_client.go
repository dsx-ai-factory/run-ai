/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	clusterclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	orgunitclient "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
	workloadclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	"golang.org/x/oauth2"
	"k8s.io/client-go/tools/clientcmd"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
)

const (
	kubeQPS      = 50
	kubeBurst    = 200
	loginTimeout = 30 * time.Second
	loginPath    = "/api/v1/token"
)

type Clients struct {
	Kube           k8sclient.Client
	ClusterService *clusterclient.APIClient
	OrgUnits       *orgunitclient.APIClient
	Workloads      *workloadclient.APIClient
}

func New(ctx context.Context, env core.Env) (*Clients, error) {
	baseURL, err := controlPlaneURL(env.ControlPlaneURL)
	if err != nil {
		return nil, err
	}

	source, err := credentials(ctx, env, baseURL)
	if err != nil {
		return nil, err
	}
	httpClient := oauth2.NewClient(ctx, source)

	kube, err := newKubeClient(env.KubeconfigPath)
	if err != nil {
		return nil, err
	}

	return &Clients{
		Kube: kube,
		ClusterService: clusterclient.NewAPIClient(&clusterclient.Configuration{
			HTTPClient: httpClient,
			Servers:    clusterclient.ServerConfigurations{{URL: baseURL}},
		}),
		OrgUnits: orgunitclient.NewAPIClient(&orgunitclient.Configuration{
			HTTPClient: httpClient,
			Servers:    orgunitclient.ServerConfigurations{{URL: baseURL}},
		}),
		Workloads: workloadclient.NewAPIClient(&workloadclient.Configuration{
			HTTPClient: httpClient,
			Servers:    workloadclient.ServerConfigurations{{URL: baseURL}},
		}),
	}, nil
}

func controlPlaneURL(raw string) (string, error) {
	if raw == "" {
		return "", errors.New("control plane URL is empty (set --cp-url)")
	}
	if !strings.HasPrefix(raw, "https://") {
		return "", fmt.Errorf("control plane URL %q must start with https://", raw)
	}
	return strings.TrimRight(raw, "/"), nil
}

func credentials(ctx context.Context, env core.Env, baseURL string) (oauth2.TokenSource, error) {
	if env.Token != "" {
		return oauth2.StaticTokenSource(&oauth2.Token{AccessToken: env.Token}), nil
	}
	if env.Username == "" || env.Password == "" {
		return nil, errors.New("missing control plane credentials: provide username and password, or a token")
	}

	loginClient := &http.Client{Timeout: loginTimeout}
	return newTokenSource(ctx, loginClient, baseURL+loginPath, env.Username, env.Password), nil
}

func newKubeClient(path string) (k8sclient.Client, error) {
	if path == "" {
		return nil, errors.New("kubeconfig path is empty (set --kubeconfig)")
	}

	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, fmt.Errorf("loading kubeconfig %q: %w", path, err)
	}
	config.QPS = kubeQPS
	config.Burst = kubeBurst

	kube, err := k8sclient.New(config, k8sclient.Options{})
	if err != nil {
		return nil, fmt.Errorf("creating kube client: %w", err)
	}
	return kube, nil
}
