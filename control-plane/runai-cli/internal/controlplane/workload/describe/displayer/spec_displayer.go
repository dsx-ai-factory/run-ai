/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package displayer

import (
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
)

// Type definitions
type Spec struct {
	Labels                   []*openapi.Label                      `json:"podLabels,omitempty" yaml:"podLabels,omitempty"`
	Annotations              []*openapi.Annotation                 `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	MasterLabels             []*openapi.Label                      `json:"masterPodLabels,omitempty" yaml:"masterPodLabels,omitempty"`
	AutoScaling              openapi.AutoScaling                   `json:"autoScaling,omitempty" yaml:"autoScaling,omitempty"`
	RunaiCliCommand          string                                `json:"runaiCliCommand,omitempty" yaml:"runaiCliCommand,omitempty"`
	ServingConfiguration     openapi.ServingConfiguration          `json:"servingConfiguration,omitempty" yaml:"servingConfiguration,omitempty"`
	DistributedInferenceSpec *openapi.DistributedInferenceSpecSpec `json:"distributedInferenceSpec,omitempty" yaml:"distributedInferenceSpec,omitempty"`
	ExposedUrls              []openapi.ExposedUrl                  `json:"exposedUrls,omitempty" yaml:"exposedUrls,omitempty"`
	ServingPort              *openapi.ServingPort                  `json:"servingPort,omitempty" yaml:"servingPort,omitempty"`
}
