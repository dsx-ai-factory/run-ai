/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package prometheus

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/prometheus/client_golang/api"
	promv1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/config"
	"github.com/prometheus/common/model"
	"k8s.io/klog/v2"
)

const (
	prometheusMetricTimeoutSeconds = 5
	serviceAccountTokenFile        = "/var/run/secrets/kubernetes.io/serviceaccount/token" //nolint:gosec
	httpClientName                 = "runtime-rules-prom-client"
	httpsPrefix                    = "https"
)

type PromClient struct {
	client api.Client
}

func NewPromClient(serviceUrl string) (*PromClient, error) {
	zap.S().Infof("Initializing prometheus client with url: <%v>", serviceUrl)
	client, err := createClient(serviceUrl)
	if err != nil {
		return nil, err
	}

	return &PromClient{client: client}, nil
}

func (pc *PromClient) Query(ctx context.Context, query string) (result model.Value, err error) {
	v1api := promv1.NewAPI(pc.client)
	ctx, cancel := context.WithTimeout(ctx, prometheusMetricTimeoutSeconds*time.Second)
	defer cancel()

	result, warnings, _, err := v1api.Query(ctx, query, time.Now())
	if err != nil {
		klog.Error("Error querying prometheus: ", err.Error())
		return nil, err
	}
	if len(warnings) > 0 {
		klog.Warning("Warnings while querying prometheus: ", warnings)
	}

	return result, nil
}

func createClient(serviceUrl string) (api.Client, error) {
	var clientConfig api.Config
	if strings.HasPrefix(serviceUrl, httpsPrefix) {
		httpsConfig, err := createHttpsClientConfig(serviceUrl)
		if err != nil {
			return nil, err
		}
		clientConfig = *httpsConfig
	} else {
		clientConfig = api.Config{
			Address: serviceUrl,
		}
	}
	return api.NewClient(clientConfig)
}

func createHttpsClientConfig(serviceUrl string) (*api.Config, error) {
	tlsConfig := config.TLSConfig{
		InsecureSkipVerify: true,
	}

	httpClientConfig := config.HTTPClientConfig{
		BearerTokenFile: serviceAccountTokenFile,
		TLSConfig:       tlsConfig,
	}

	httpClient, err := config.NewClientFromConfig(httpClientConfig, httpClientName)
	if err != nil {
		return nil, err
	}

	clientConfig := api.Config{
		Address: serviceUrl,
		Client:  httpClient,
	}
	return &clientConfig, nil
}
