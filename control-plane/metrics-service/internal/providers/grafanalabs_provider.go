/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package providers

import (
	"context"
	"errors"
	"fmt"

	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/grafana"
	"github.com/dsx-ai-factory/run-ai/control-plane/metrics-service/internal/metricstores"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

const (
	readMetricAPIRole   = "metrics:read"
	writeMetricAPIRole  = "metrics:write"
	readPrometheusPath  = "/api/prom"
	writePrometheusPath = "/api/prom/push"

	deleteVerifyMaxAttempts = 1
)

type GrafanaLabsClienter interface {
	CreateInstance(ctx context.Context, datasourceId string) (*grafana.Instance, error)
	CreateAccessPolicy(ctx context.Context, policyScope string, instance *grafana.Instance) (*grafana.Policy, error)
	CreateToken(ctx context.Context, region string, policy *grafana.Policy) (*grafana.Token, error)
	DeleteInstance(ctx context.Context, datasourceId string) error
	DeleteAccessPolicy(ctx context.Context, accessPolicyId string, region string) error
	DeleteToken(ctx context.Context, tokenId string, region string) error
}

type GrafanaLabsProvider struct {
	Client GrafanaLabsClienter
}

func NewGrafanaLabsProvider() *GrafanaLabsProvider {
	cfg := config.Get[*DatasourceProviderConfig]()
	return &GrafanaLabsProvider{Client: grafana.NewLabsClient(cfg.Auth.Token)}
}

func (glp *GrafanaLabsProvider) Create(ctx context.Context, datasourceId string) (*metricstores.Datasource, error) {
	instance, err := glp.Client.CreateInstance(ctx, datasourceId)
	if err != nil {
		return nil, fmt.Errorf("failed to create instance: %w", err)
	}

	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("Created grafana lab instance for datasource %s: %s", datasourceId, instance.Slug)

	readPolicy, err := glp.Client.CreateAccessPolicy(ctx, readMetricAPIRole, instance)
	if err != nil {
		return nil, fmt.Errorf("failed to create read access policy: %w", err)
	}

	readToken, err := glp.Client.CreateToken(ctx, instance.Region, readPolicy)
	if err != nil {
		return nil, fmt.Errorf("failed to create read token: %w", err)
	}

	writePolicy, err := glp.Client.CreateAccessPolicy(ctx, writeMetricAPIRole, instance)
	if err != nil {
		return nil, fmt.Errorf("failed to create write access policy: %w", err)
	}

	writeToken, err := glp.Client.CreateToken(ctx, instance.Region, writePolicy)
	if err != nil {
		return nil, fmt.Errorf("failed to create write token: %w", err)
	}

	return &metricstores.Datasource{
		Type: metricstores.DatasourceTypeGrafanaLabs,
		Read: metricstores.DatasourceEndpoint{
			Url: instance.Url + readPrometheusPath,
			Auth: &metricstores.DatasourceAuth{
				Type: metricstores.AuthTypeBasic,
				Basic: &metricstores.AuthBasic{
					User:     fmt.Sprintf("%d", instance.UserId),
					Password: readToken.Token,
				},
			},
		},
		Write: metricstores.DatasourceEndpoint{
			Url: instance.Url + writePrometheusPath,
			Auth: &metricstores.DatasourceAuth{
				Type: metricstores.AuthTypeBasic,
				Basic: &metricstores.AuthBasic{
					User:     fmt.Sprintf("%d", instance.UserId),
					Password: writeToken.Token,
				},
			},
		},
	}, nil
}

func (glp *GrafanaLabsProvider) Delete(ctx context.Context, datasourceId string) error {
	log := logging.WithCtx(runaictx.LoggerHook(ctx))
	log.Infof("deleting grafana labs instance for datasource %s", datasourceId)

	err := glp.Client.DeleteInstance(ctx, datasourceId)
	if errors.Is(err, grafana.ErrInstanceNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to delete instance: %w", err)
	}

	for attempt := 1; attempt <= deleteVerifyMaxAttempts; attempt++ {
		err := glp.Client.DeleteInstance(ctx, datasourceId)
		if errors.Is(err, grafana.ErrInstanceNotFound) {
			return nil
		}
		if err != nil {
			log.Warnf("verification attempt %d/%d for grafana labs instance %s failed: %v", attempt, deleteVerifyMaxAttempts, datasourceId, err)
		} else {
			log.Warnf("verification attempt %d/%d for grafana labs instance %s: instance still exists", attempt, deleteVerifyMaxAttempts, datasourceId)
		}
	}
	return nil
}
