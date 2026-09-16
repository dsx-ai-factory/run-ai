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

// ControlPlaneRef locates the control plane a cluster registers with and says
// how the installer authenticates to it. With Credentials set it is an external
// control plane reached at URL with the user-supplied credentials; otherwise it
// is the control plane installed in the local cluster, reached at Domain with
// the admin credentials read from that cluster's secrets.
type ControlPlaneRef struct {
	URL         string
	Domain      string
	CACertPath  string
	Credentials *step.ControlPlaneCredentials
}

// Endpoint is the control-plane base URL API calls go to.
func (r ControlPlaneRef) Endpoint() string {
	if r.URL != "" {
		return r.URL
	}
	return ControlPlaneURL(r.Domain)
}

// Client returns the control-plane API client matching the ref's auth mode.
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

// clusterChartCredentialValues resolves the cluster's uid and control-plane
// credential from the control plane (registered by name) and renders the chart
// values that connect the cluster to it. controlPlaneURL is what the cluster
// chart is told to call; clusterURL is the address the control plane reaches
// the cluster on.
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
