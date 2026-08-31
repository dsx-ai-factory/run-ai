/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package distributedinference

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/workload/describe/displayer"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	"k8s.io/utils/ptr"
)

type Builder struct {
	fields Fields
}

func NewBuilder() *Builder {
	return &Builder{
		fields: Fields{},
	}
}

func (b *Builder) RestartPolicy(val string) *Builder {
	b.fields.RestartPolicy = &val
	return b
}

func (b *Builder) StartupPolicy(val string) *Builder {
	b.fields.StartupPolicy = &val
	return b
}

func (b *Builder) Replicas(val *int32) *Builder {
	b.fields.Replicas = val
	return b
}

func (b *Builder) Workers(val *int32) *Builder {
	b.fields.Workers = val
	return b
}

func (b *Builder) ServingPort(val *int32) *Builder {
	b.fields.ServingPort = val
	return b
}

func (b *Builder) AuthorizationType(val string) *Builder {
	b.fields.AuthorizationType = &val
	return b
}

func (b *Builder) AuthorizedUsers(val []string) *Builder {
	b.fields.AuthorizedUsers = val
	return b
}

func (b *Builder) AuthorizedGroups(val []string) *Builder {
	b.fields.AuthorizedGroups = val
	return b
}

func (b *Builder) Protocol(val string) *Builder {
	b.fields.Protocol = &val
	return b
}

func (b *Builder) ExposeExternally(val *bool) *Builder {
	b.fields.ExposeExternally = val
	return b
}

func (b *Builder) ServingURL(val string) *Builder {
	b.fields.ServingURL = &val
	return b
}

func (b *Builder) ReplicasReady(val *int32) *Builder {
	b.fields.ReplicasReady = val
	return b
}

func (b *Builder) Leader(val *displayer.WorkloadRuntimeConfig) *Builder {
	b.fields.Leader = val
	return b
}

func (b *Builder) Worker(val *displayer.WorkloadRuntimeConfig) *Builder {
	b.fields.Worker = val
	return b
}

func (b *Builder) Build() *Fields {
	return &b.fields
}

func BuildFromSpec(spec *openapi.DistributedInferenceSpecSpec) *Fields {
	if spec == nil {
		return nil
	}

	builder := NewBuilder()

	if spec.RestartPolicy.IsSet() {
		builder.RestartPolicy(string(spec.GetRestartPolicy()))
	}

	if spec.StartupPolicy.IsSet() {
		builder.StartupPolicy(string(spec.GetStartupPolicy()))
	}

	if spec.Replicas.IsSet() {
		replicas := spec.GetReplicas()
		builder.Replicas(&replicas)
	}

	if spec.Workers.IsSet() {
		workers := spec.GetWorkers()
		builder.Workers(&workers)
	}

	if spec.ServingPort.IsSet() {
		servingPort := spec.GetServingPort()

		if servingPort.Port.IsSet() {
			port := servingPort.GetPort()
			builder.ServingPort(&port)
		}

		if servingPort.AuthorizationType.IsSet() {
			builder.AuthorizationType(string(servingPort.GetAuthorizationType()))
		}

		if len(servingPort.AuthorizedUsers) > 0 {
			builder.AuthorizedUsers(servingPort.AuthorizedUsers)
		}

		if len(servingPort.AuthorizedGroups) > 0 {
			builder.AuthorizedGroups(servingPort.AuthorizedGroups)
		}

		if servingPort.Protocol.IsSet() {
			builder.Protocol(string(servingPort.GetProtocol()))
		}

		if servingPort.ExposeExternally.IsSet() {
			exposeExternally := servingPort.GetExposeExternally()
			builder.ExposeExternally(&exposeExternally)
		}
	}

	if spec.Leader.IsSet() {
		leader := spec.GetLeader()
		builder.Leader(BuildWorkloadRuntimeConfigFromSpec(&leader))
	}

	if spec.Worker.IsSet() {
		worker := spec.GetWorker()
		builder.Worker(BuildWorkloadRuntimeConfigFromSpec(&worker))
	}

	return builder.Build()
}

func SetFields(target **Fields, fields *Fields) {
	*target = fields
}

func BuildAndSetFields(target **Fields, spec *openapi.DistributedInferenceSpecSpec, additionalFields map[string]interface{}, urls []string) {
	if spec == nil {
		return
	}
	fields := BuildFromSpec(spec)

	if additionalFields != nil {
		if readyReplicas, ok := additionalFields["readyReplicas"]; ok {
			if val, ok := readyReplicas.(float64); ok {
				fields.ReplicasReady = ptr.To(int32(val))
			}
		}
	}

	if len(urls) > 0 {
		fields.ServingURL = &urls[0]
	}

	SetFields(target, fields)
}
