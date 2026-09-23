/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
)

type ControlPlaneRef struct {
	URL         string
	Domain      string
	CACertPath  string
	Credentials *step.ControlPlaneCredentials
}

func (r ControlPlaneRef) Endpoint() string {
	if r.URL != "" {
		return r.URL
	}
	return ControlPlaneURL(r.Domain)
}

func (r ControlPlaneRef) Client(deps orchestrator2.Deps) (step.RunAIAPIClient, error) {
	if r.Credentials != nil {
		if deps.RunAIClientWithCredentials == nil {
			return nil, errors.New("control-plane credentials are not supported by this installer build")
		}
		return deps.RunAIClientWithCredentials(r.Endpoint(), r.CACertPath, *r.Credentials)
	}
	if deps.RunAIClientProvider == nil {
		return nil, errors.New("no control-plane client provider configured")
	}
	return deps.RunAIClientProvider(r.Endpoint(), r.CACertPath)
}

func clusterChartCredentialValues(ctx context.Context, apiClient step.RunAIAPIClient, clusterName, version, controlPlaneURL, clusterURL string) ([]string, error) {
	uuid, err := FindClusterUUID(ctx, apiClient, clusterName)
	if err != nil {
		return nil, err
	}
	if uuid == "" {
		return nil, fmt.Errorf("cluster %q not found in control plane", clusterName)
	}
	installInfo, statusCode, err := apiClient.GetClusterInstallInfoByUUID(ctx, uuid, version, clusterURL)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("get cluster install info: unexpected status %d", statusCode)
	}
	return []string{
		"controlPlane.url=" + controlPlaneURL,
		step.ClusterAuthValue(installInfo),
		"cluster.uid=" + uuid,
		"cluster.url=" + clusterURL,
	}, nil
}
