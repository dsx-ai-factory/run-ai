/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package distributedinference

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/describe/displayer"
)

type Fields struct {
	RestartPolicy          *string                          `json:"restartPolicy,omitempty"`
	StartupPolicy          *string                          `json:"startupPolicy,omitempty"`
	Replicas               *int32                           `json:"replicas,omitempty"`
	Workers                *int32                           `json:"workers,omitempty"`
	ServingPort            *int32                           `json:"servingPort,omitempty"`
	ServingURL             *string                          `json:"servingUrl,omitempty"`
	AuthorizationType      *string                          `json:"authorizationType,omitempty"`
	AuthorizedUsers        []string                         `json:"authorizedUsers,omitempty"`
	AuthorizedGroups       []string                         `json:"authorizedGroups,omitempty"`
	Protocol               *string                          `json:"protocol,omitempty"`
	ExposeExternally       *bool                            `json:"exposeExternally,omitempty"`
	ClusterLocalAccessOnly *bool                            `json:"clusterLocalAccessOnly,omitempty"`
	ReplicasReady          *int32                           `json:"replicasReady,omitempty"`
	Leader                 *displayer.WorkloadRuntimeConfig `json:"leader,omitempty"`
	Worker                 *displayer.WorkloadRuntimeConfig `json:"worker,omitempty"`
}
