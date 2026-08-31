/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package displayer

import (
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

// Builder
type SpecBuilder struct {
	spec Spec
}

func NewSpecBuilder() *SpecBuilder {
	return &SpecBuilder{spec: Spec{}}
}

func (b *SpecBuilder) Labels(labels []*openapi.Label) *SpecBuilder {
	b.spec.Labels = labels
	return b
}

func (b *SpecBuilder) MasterLabels(labels []*openapi.Label) *SpecBuilder {
	b.spec.MasterLabels = labels
	return b
}
func (b *SpecBuilder) Annotations(annotations []*openapi.Annotation) *SpecBuilder {
	b.spec.Annotations = annotations
	return b
}

func (b *SpecBuilder) SetRunaiCliCommandFromAnnotations(annotations *[]openapi.Annotation) *SpecBuilder {
	if annotations != nil {
		for _, ann := range *annotations {
			if ann.Name.Get() != nil && *ann.Name.Get() == RunaiSubmitCommandAnnotation {
				b.spec.RunaiCliCommand = *ann.Value.Get()
			}
		}
	}
	return b
}

func (b *SpecBuilder) AutoScaling(autoScaling openapi.AutoScaling) *SpecBuilder {
	b.spec.AutoScaling = autoScaling
	return b
}

func (b *SpecBuilder) AutoScalingPtr(autoScaling *openapi.AutoScaling) *SpecBuilder {
	if autoScaling != nil {
		b.spec.AutoScaling = *autoScaling
	}
	return b
}

func (b *SpecBuilder) RunaiCliCommand(command string) *SpecBuilder {
	b.spec.RunaiCliCommand = command
	return b
}

func (b *SpecBuilder) ServingConfiguration(config openapi.ServingConfiguration) *SpecBuilder {
	b.spec.ServingConfiguration = config
	return b
}

func (b *SpecBuilder) ServingConfigurationPtr(config *openapi.ServingConfiguration) *SpecBuilder {
	if config != nil {
		b.spec.ServingConfiguration = *config
	}
	return b
}

func (b *SpecBuilder) DistributedInferenceSpec(spec *openapi.DistributedInferenceSpecSpec) *SpecBuilder {
	b.spec.DistributedInferenceSpec = spec
	return b
}

func (b *SpecBuilder) ExposedUrls(urls []openapi.ExposedUrl) *SpecBuilder {
	b.spec.ExposedUrls = urls
	return b
}

func (b *SpecBuilder) ServingPort(servingPort *openapi.ServingPort) *SpecBuilder {
	b.spec.ServingPort = servingPort
	return b
}

func (b *SpecBuilder) Set(spec Spec) *SpecBuilder {
	b.spec = spec
	return b
}

func (b *SpecBuilder) Build() Spec {
	return b.spec
}
