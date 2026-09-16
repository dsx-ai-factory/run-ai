/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package actions

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/runai-installer/installer/internal/orchestrator2"
)

type CheckControlPlaneAccess struct {
	ClusterName  string
	ControlPlane ControlPlaneRef
}

func (a CheckControlPlaneAccess) Kind() string { return "preflight.control_plane_access" }

func (a CheckControlPlaneAccess) Args() map[string]any {
	return map[string]any{"controlPlane": a.ControlPlane.Endpoint(), "cluster": a.ClusterName}
}

func (a CheckControlPlaneAccess) Describe() string {
	return fmt.Sprintf("verify the control plane at %s is reachable and accepts the supplied credentials", a.ControlPlane.Endpoint())
}

func (a CheckControlPlaneAccess) Execute(ctx context.Context, deps orchestrator2.Deps) (orchestrator2.ActionResult, error) {
	endpoint := a.ControlPlane.Endpoint()
	data := map[string]any{"controlPlaneUrl": endpoint}
	apiClient, err := a.ControlPlane.Client(deps)
	if err != nil {
		return orchestrator2.ActionResult{
			Status:      orchestrator2.StatusFail,
			Reason:      fmt.Sprintf("the control plane at %s is not reachable: %v", endpoint, err),
			Remediation: "check cluster.controlPlaneUrl, that the control plane is reachable from where the installer runs, and that its certificate is trusted (cluster.tls.caChainPath)",
			Data:        data,
		}, nil
	}
	clusters, statusCode, err := apiClient.GetClusters(ctx)
	if err != nil || statusCode != http.StatusOK {
		detail := fmt.Sprintf("status %d", statusCode)
		if err != nil {
			detail = err.Error()
		}
		return orchestrator2.ActionResult{
			Status:      orchestrator2.StatusFail,
			Reason:      fmt.Sprintf("the control plane at %s rejected the supplied credentials or denied listing clusters (%s)", endpoint, detail),
			Remediation: "check cluster.controlPlaneCredentials: the user or token must be able to view and create clusters in the control plane",
			Data:        data,
		}, nil
	}
	for _, c := range clusters {
		if c.Name == a.ClusterName {
			data["clusterUid"] = c.Uuid
			data["clusterRegistered"] = true
			return orchestrator2.ActionResult{Data: data}, nil
		}
	}
	data["clusterRegistered"] = false
	return orchestrator2.ActionResult{Data: data}, nil
}
