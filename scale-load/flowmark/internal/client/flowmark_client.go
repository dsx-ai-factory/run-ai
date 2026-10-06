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
	"net/url"
	"strings"
	"time"

	clusterclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	orgunitclient "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	workloadclient "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"golang.org/x/oauth2"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/flowcontrol"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/dsx-ai-factory/run-ai/scale-load/flowmark/internal/core"
)

const (
	loginTimeout          = 30 * time.Second
	responseHeaderTimeout = time.Minute
	minTokenLifetime      = 5 * time.Minute
	loginPath             = "/api/v1/token"
)

type Clients struct {
	Identity       string
	TokenExpiry    time.Time
	Kube           k8sclient.WithWatch
	ClusterService *clusterclient.APIClient
	OrgUnits       *orgunitclient.APIClient
	Workloads      *workloadclient.APIClient

	http    *http.Client
	baseURL string
}

func New(ctx context.Context, target core.Target) (*Clients, error) {
	return NewWithLimits(ctx, target, DefaultLimits)
}

func NewWithLimits(ctx context.Context, target core.Target, limits Limits) (*Clients, error) {
	if limits.KubeRequestsPerSecond < 1 || limits.ControlPlaneRequestsPerSecond < 1 {
		return nil, fmt.Errorf("request limits must be at least 1 per second, got kube %d and control plane %d", limits.KubeRequestsPerSecond, limits.ControlPlaneRequestsPerSecond)
	}
	base, err := controlPlaneURL(target.ControlPlaneURL)
	if err != nil {
		return nil, err
	}

	source, err := credentials(ctx, target, base)
	if err != nil {
		return nil, err
	}
	token, err := source.Token()
	if err != nil {
		return nil, err
	}
	identity, err := tokenIdentity(token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("resolving the flowmark identity: %w", err)
	}
	var tokenExpiry time.Time
	if target.Token != "" {
		tokenExpiry = token.Expiry
	}
	httpClient := paced(apiClient(source), limits.ControlPlaneRequestsPerSecond)
	baseURL := base.String()

	kube, err := kubeWithLimit(target.KubeconfigPath, limits.KubeRequestsPerSecond)
	if err != nil {
		return nil, err
	}

	return &Clients{
		Identity:    identity,
		TokenExpiry: tokenExpiry,
		Kube:        kube,
		http:        httpClient,
		baseURL:     baseURL,
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

func controlPlaneURL(raw string) (*url.URL, error) {
	if raw == "" {
		return nil, errors.New("control plane URL is empty (set --cp-url)")
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		var urlErr *url.Error
		if errors.As(err, &urlErr) {
			err = urlErr.Err
		}
		return nil, fmt.Errorf("control plane URL is not a valid URL: %w", err)
	}
	redacted := parsed.Redacted()
	if parsed.Scheme != "https" {
		return nil, fmt.Errorf("control plane URL %q must use https", redacted)
	}
	if parsed.Host == "" {
		return nil, fmt.Errorf("control plane URL %q has no host", redacted)
	}
	if parsed.User != nil || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" {
		return nil, fmt.Errorf("control plane URL %q must not contain credentials, a query or a fragment", redacted)
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	parsed.RawPath = ""
	return parsed, nil
}

func apiClient(source oauth2.TokenSource) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ResponseHeaderTimeout = responseHeaderTimeout
	return &http.Client{
		Transport: &oauth2.Transport{Source: source, Base: transport},
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func credentials(ctx context.Context, target core.Target, base *url.URL) (oauth2.TokenSource, error) {
	if target.Token != "" {
		expiry, err := tokenExpiry(target.Token)
		if err != nil {
			return nil, fmt.Errorf("FLOWMARK_CP_TOKEN: %w", err)
		}
		remaining := time.Until(expiry)
		if remaining <= 0 {
			return nil, fmt.Errorf("FLOWMARK_CP_TOKEN expired %s ago (at %s)", (-remaining).Round(time.Second), expiry.UTC().Format(time.RFC3339))
		}
		if remaining < minTokenLifetime {
			return nil, fmt.Errorf("FLOWMARK_CP_TOKEN expires in %s (at %s); a run needs at least %s", remaining.Round(time.Second), expiry.UTC().Format(time.RFC3339), minTokenLifetime)
		}
		return oauth2.StaticTokenSource(&oauth2.Token{AccessToken: target.Token, Expiry: expiry}), nil
	}
	if target.Username == "" || target.Password == "" {
		return nil, errors.New("missing control plane credentials: provide username and password, or a token")
	}

	loginClient := &http.Client{
		Timeout: loginTimeout,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	return newTokenSource(ctx, loginClient, base.JoinPath(loginPath).String(), target.Username, target.Password), nil
}

func Kube(path string) (k8sclient.WithWatch, error) {
	return kubeWithLimit(path, DefaultLimits.KubeRequestsPerSecond)
}

func kubeWithLimit(path string, requestsPerSecond int) (k8sclient.WithWatch, error) {
	if path == "" {
		return nil, errors.New("kubeconfig path is empty (set --kubeconfig)")
	}

	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		return nil, fmt.Errorf("loading kubeconfig %q: %w", path, err)
	}
	config.QPS = float32(requestsPerSecond)
	config.Burst = burstOf(requestsPerSecond)
	config.RateLimiter = flowcontrol.NewTokenBucketRateLimiter(config.QPS, config.Burst)

	kube, err := k8sclient.NewWithWatch(config, k8sclient.Options{})
	if err != nil {
		return nil, fmt.Errorf("creating kube client: %w", err)
	}
	return kube, nil
}
