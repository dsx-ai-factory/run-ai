/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package harness

import (
	"context"
	"fmt"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-installer/clients/runai/cluster"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/step"
)

// FakeRunAIAPIClient is an in-memory implementation of step.RunAIAPIClient.
// CP/cluster readiness is derived from helm calls recorded by the bound
// FakeHelmInstaller so that GetClusters/GetClusterByUUID behave realistically
// during a full install flow.
type FakeRunAIAPIClient struct {
	ClusterUUID    string
	BootstrapToken string
	ClientSecret   string
	Helm           *FakeHelmInstaller
	ClusterCreated bool
	ProviderCalls  []RunAIClientProviderCall
	// CPReady marks the control plane as already present (e.g. an external,
	// pre-existing control plane). It lets cluster-only installs—where the CP
	// chart is not installed in this run—behave against a ready control plane.
	CPReady bool
}

type RunAIClientProviderCall struct {
	URL        string
	CACertPath string
}

func (f *FakeRunAIAPIClient) cpInstalled() bool {
	return f.CPReady || len(f.Helm.CallsByRelease(step.CPReleaseName)) > 0
}

func (f *FakeRunAIAPIClient) clusterInstalled() bool {
	return len(f.Helm.CallsByRelease(step.ClusterReleaseName)) > 0
}

func (f *FakeRunAIAPIClient) GetClusters(_ context.Context) ([]cluster.DisplayedCluster, int, error) {
	if !f.cpInstalled() {
		return nil, 0, fmt.Errorf("CP not ready")
	}
	if !f.ClusterCreated {
		return []cluster.DisplayedCluster{}, 200, nil
	}
	return []cluster.DisplayedCluster{{Name: "test-cluster", Uuid: f.ClusterUUID}}, 200, nil
}

func (f *FakeRunAIAPIClient) CreateCluster(_ context.Context, req cluster.ClusterCreationRequest) (*cluster.DisplayedCluster, int, error) {
	f.ClusterCreated = true
	return &cluster.DisplayedCluster{Name: req.Name, Uuid: f.ClusterUUID}, 200, nil
}

func (f *FakeRunAIAPIClient) GetClusterByUUID(_ context.Context, uuid string) (*cluster.DisplayedCluster, int, error) {
	if !f.clusterInstalled() {
		return nil, 0, fmt.Errorf("cluster not ready")
	}
	t := time.Now()
	dc := cluster.NewDisplayedCluster(uuid, 1, "test-cluster", t)
	dc.SetLastLiveness(t)
	return dc, 200, nil
}

func (f *FakeRunAIAPIClient) GetClusterInstallInfoByUUID(_ context.Context, _, _, _ string) (*cluster.ClusterInstallationInfoResponse, int, error) {
	installInfo := cluster.NewClusterInstallationInfoResponse("", "", "", f.ClientSecret)
	installInfo.BootstrapToken = f.BootstrapToken
	return installInfo, 200, nil
}

// NewFakeRunAIClientProvider builds a FakeRunAIAPIClient and returns both
// the client (so tests can assert on its state) and a provider closure that
// installer code can use as a step.RunAIAPIClientProvider.
func NewFakeRunAIClientProvider(helm *FakeHelmInstaller) (*FakeRunAIAPIClient, step.RunAIAPIClientProvider) {
	client := &FakeRunAIAPIClient{
		ClusterUUID:  "uuid-airgap-test",
		ClientSecret: "s3cr3t-airgap",
		Helm:         helm,
	}
	return client, func(url, caCertPath string) (step.RunAIAPIClient, error) {
		client.ProviderCalls = append(client.ProviderCalls, RunAIClientProviderCall{URL: url, CACertPath: caCertPath})
		return client, nil
	}
}

// FakeClusterLookup implements upgrade.ClusterLookup for upgrade tests.
type FakeClusterLookup struct {
	InstallInfo *cluster.ClusterInstallationInfoResponse
}

func (f *FakeClusterLookup) GetClusterInstallInfoByUUID(_ context.Context, _, _, _ string) (*cluster.ClusterInstallationInfoResponse, int, error) {
	return f.InstallInfo, 200, nil
}

// NewFakeClusterLookup returns a FakeClusterLookup pre-seeded with a
// stock install-info response that carries the well-known test client secret.
func NewFakeClusterLookup() *FakeClusterLookup {
	return &FakeClusterLookup{
		InstallInfo: cluster.NewClusterInstallationInfoResponse("", "", "", "s3cr3t"),
	}
}
