/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package describe

import (
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/describe/displayer"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/describe/distributedinference"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/describe/inference"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/describe/nim"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
)

type Workload struct {
	General              displayer.General                `json:"general,omitempty"`
	Runtime              displayer.Runtime                `json:"workloadDetails,omitempty"`
	Compute              displayer.Compute                `json:"compute,omitempty"`
	Storage              *displayer.Storage               `json:"storage,omitempty"`
	MasterConfig         *displayer.WorkloadRuntimeConfig `json:"masterConfig,omitempty"`
	Networks             []displayer.Network              `json:"networks,omitempty"`
	NetworkTopology      *displayer.NetworkTopology       `json:"networkTopology,omitempty"`
	Events               []displayer.Event                `json:"events,omitempty"`
	PodsInfo             displayer.PodsInfo               `json:"podsInfo,omitempty"`
	Spec                 displayer.Spec                   `json:"spec,omitempty"`
	Issues               []displayer.Issue                `json:"issues,omitempty"`
	GPUAcceleration      *bool                            `json:"GPUAcceleration"`
	Inference            *inference.Fields                `json:"inference,omitempty"`
	DistributedInference *distributedinference.Fields     `json:"distributedInference,omitempty"`
	Nim                  *nim.Fields                      `json:"nim,omitempty"`
}

type Issues struct {
	Network  []displayer.Issue `json:"network"`
	Storage  []displayer.Issue `json:"storage"`
	Compute  []displayer.Issue `json:"compute"`
	Internal []displayer.Issue `json:"internal"`
}

func PhaseFromApi(phase openapi.Phase) displayer.Phase {
	phaseStr := displayer.Phase(phase)
	for _, allowed := range displayer.AllowedPhaseEnumValues {
		if allowed == phaseStr {
			return allowed
		}
	}
	return displayer.Unknown
}
