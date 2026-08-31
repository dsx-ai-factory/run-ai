/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package apis

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/api"
)

type Workload struct {
	Type      api.WorkloadType
	Project   string
	Name      string
	Framework api.Framework
}

func NewWorkload(project string, workloadType api.WorkloadType, name string, framework api.Framework) *Workload {
	return &Workload{
		Type:      workloadType,
		Project:   project,
		Name:      name,
		Framework: framework,
	}
}

// NewWorkloadForClusterAPI creates a Workload for cluster API calls (exec, logs, port-forward).
// Name and framework are translated per api.Framework so the cluster URL is correct
// (e.g. NimService → external/external/ew-nimservice-<name>).
func NewWorkloadForClusterAPI(project string, workloadType api.WorkloadType, name string, framework api.Framework) *Workload {
	clusterName, clusterFramework := framework.ClusterAPIIdentifiers(name)
	return NewWorkload(project, workloadType, clusterName, clusterFramework)
}
