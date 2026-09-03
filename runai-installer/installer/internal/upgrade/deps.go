/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package upgrade

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/runai-installer/clients/runai/cluster"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/apiclient"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/helmclient"
	k8s "github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/utils/k8sclient"
)

type ClusterLookup interface {
	GetClusterInstallInfoByUUID(ctx context.Context, uuid, version, fqdn string) (*cluster.ClusterInstallationInfoResponse, int, error)
}

type HelmValuesReader func(releaseName, namespace string) (map[string]interface{}, error)

type Deps struct {
	HelmInstaller       step.HelmInstaller
	ClusterLookup       ClusterLookup
	HelmValuesReader    HelmValuesReader
	RunAIClientProvider step.RunAIAPIClientProvider
}

type clusterLookupImpl struct {
	fqdn           string
	caCertPath     string
	clientProvider step.RunAIAPIClientProvider
}

func (c *clusterLookupImpl) GetClusterInstallInfoByUUID(ctx context.Context, uuid, version, fqdn string) (*cluster.ClusterInstallationInfoResponse, int, error) {
	cl, err := c.clientProvider(c.fqdn, c.caCertPath)
	if err != nil {
		return nil, 0, err
	}
	return cl.GetClusterInstallInfoByUUID(ctx, uuid, version, fqdn)
}

func NewClusterLookup(fqdn, caCertPath string, clientProvider step.RunAIAPIClientProvider) ClusterLookup {
	return &clusterLookupImpl{fqdn: fqdn, caCertPath: caCertPath, clientProvider: clientProvider}
}

func ProductionDeps(k8sClient k8s.K8sClients, controlPlaneVersion string) Deps {
	return Deps{
		HelmInstaller:       step.NewHelmInstaller(),
		HelmValuesReader:    helmclient.GetReleaseValues,
		RunAIClientProvider: step.NewRunAIAPIClientProvider(apiclient.NewObjectGetter(k8sClient.GetClient()), controlPlaneVersion),
	}
}
