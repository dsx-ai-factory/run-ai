/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import (
	"context"
	"fmt"

	runaicluster "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
)

// EnsureCluster registers the cluster with the control plane by name
// (get-or-create). ControlPlane.Domain is the cluster's own domain, which for a
// same-cluster install is also the control plane's; ControlPlane.URL and
// Credentials point at an external control plane instead.
type EnsureCluster struct {
	ClusterName  string
	ControlPlane ControlPlaneRef
}

// NewEnsureCluster is the same-cluster form: the control plane lives at domain
// and is reached with the admin credentials read from the local cluster.
func NewEnsureCluster(clusterName, domain, caCertPath string) EnsureCluster {
	return EnsureCluster{ClusterName: clusterName, ControlPlane: ControlPlaneRef{Domain: domain, CACertPath: caCertPath}}
}

func (a EnsureCluster) Kind() string { return "runai.ensure_cluster" }

func (a EnsureCluster) Args() map[string]any {
	return map[string]any{
		"cluster":      a.ClusterName,
		"domain":       a.ControlPlane.Domain,
		"controlPlane": a.ControlPlane.Endpoint(),
	}
}

func (a EnsureCluster) Describe() string {
	return fmt.Sprintf("ensure cluster %q is registered with control plane at %s", a.ClusterName, a.ControlPlane.Endpoint())
}

// clusterDomain is the address the control plane records for the cluster: the
// cluster's own URL, which for a same-cluster install is the control plane URL.
func (a EnsureCluster) clusterDomain() string {
	return ControlPlaneURL(a.ControlPlane.Domain)
}

func (a EnsureCluster) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	apiClient, err := a.ControlPlane.Client(deps)
	if err != nil {
		return orchestrator2.ActionResult{}, err
	}
	uuid, err := FindClusterUUID(ctx, apiClient, a.ClusterName)
	if err != nil {
		return orchestrator2.ActionResult{}, err
	}
	if uuid != "" {
		return orchestrator2.ActionResult{Data: map[string]any{"clusterUid": uuid, "created": false}}, nil
	}
	domain := a.clusterDomain()
	created, statusCode, err := apiClient.CreateCluster(ctx, runaicluster.ClusterCreationRequest{Name: a.ClusterName, Domain: &domain})
	if err != nil {
		return orchestrator2.ActionResult{}, err
	}
	if statusCode < 200 || statusCode >= 300 {
		return orchestrator2.ActionResult{}, fmt.Errorf("create cluster %q: unexpected status %d", a.ClusterName, statusCode)
	}
	data := map[string]any{"created": true}
	if created != nil {
		data["clusterUid"] = created.Uuid
	}
	return orchestrator2.ActionResult{Data: data}, nil
}

// Detect reports whether the control plane already knows the cluster by name.
func (a EnsureCluster) Detect(ctx context.Context, deps orchestrator2.Deps) (bool, error) {
	apiClient, err := a.ControlPlane.Client(deps)
	if err != nil {
		return false, err
	}
	uuid, err := FindClusterUUID(ctx, apiClient, a.ClusterName)
	if err != nil {
		return false, err
	}
	return uuid != "", nil
}
