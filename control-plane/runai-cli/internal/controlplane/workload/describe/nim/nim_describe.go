/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package nim

import (
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	nimclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-manager-client/generated"
)

const (
	FieldServingPort          = "Serving Port"
	FieldGrpcPort             = "gRPC Port"
	FieldMetricsPort          = "Metrics Port"
	FieldExposeExternally     = "Expose Externally"
	FieldExposedProtocol      = "Exposed Protocol"
	FieldServingURL           = "Serving URL"
	FieldServiceType          = "Service Type"
	FieldModelStoreType       = "Type"
	FieldModelStoreClaimName  = "Claim Name"
	FieldModelStoreAccessMode = "Access Mode"
	FieldModelStoreCacheName  = "Cache Name"
	FieldModelStoreProfile    = "Profile"
	FieldMinReplicas          = "Min Replicas"
	FieldMaxReplicas          = "Max Replicas"
	FieldMetric               = "Metric"
	FieldMetricThreshold      = "Threshold"
	FieldScaleWindow          = "Scale Window"
	FieldMultiNode            = "Multi-Node"
	FieldWorkersPerReplica    = "Workers per Replica"
	FieldNgcAuthSecret        = "NGC Auth Secret"
	FieldReplicasRequested    = "Replicas (Requested)"
	FieldReplicasReady        = "Replicas (Ready)"

	TitleNimService           = "NIM Service Configuration"
	TitleServingConfiguration = "Serving Configuration"
	TitleModelStore           = "Model Store"
	TitleAutoscaling          = "Autoscaling"
)

type Fields struct {
	Image            *string  `json:"image,omitempty"`
	NgcAuthSecret    *string  `json:"ngcAuthSecret,omitempty"`
	Category         *string  `json:"category,omitempty"`
	PriorityClass    *string  `json:"priorityClass,omitempty"`
	Preemptibility   *string  `json:"preemptibility,omitempty"`
	ImagePullPolicy  *string  `json:"imagePullPolicy,omitempty"`
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`

	ServingPort      *int32  `json:"servingPort,omitempty"`
	GrpcPort         *int32  `json:"grpcPort,omitempty"`
	MetricsPort      *int32  `json:"metricsPort,omitempty"`
	ExposeExternally *bool   `json:"exposeExternally,omitempty"`
	ExposedProtocol  *string `json:"exposedProtocol,omitempty"`
	ServingURL       *string `json:"servingUrl,omitempty"`
	ServiceType      *string `json:"serviceType,omitempty"`

	GpuDevicesRequest *int32   `json:"gpuDevicesRequest,omitempty"`
	GpuPortionRequest *float64 `json:"gpuPortionRequest,omitempty"`
	GpuPortionLimit   *float64 `json:"gpuPortionLimit,omitempty"`
	GpuComputeMode    *string  `json:"gpuComputeMode,omitempty" yaml:"gpucomputemode,omitempty"`
	GpuMemoryRequest  *string  `json:"gpuMemoryRequest,omitempty"`
	GpuMemoryLimit    *string  `json:"gpuMemoryLimit,omitempty"`
	CpuCoreRequest    *float64 `json:"cpuCoreRequest,omitempty"`
	CpuCoreLimit      *float64 `json:"cpuCoreLimit,omitempty"`
	CpuMemoryRequest  *string  `json:"cpuMemoryRequest,omitempty"`
	CpuMemoryLimit    *string  `json:"cpuMemoryLimit,omitempty"`

	ModelStoreType       *string `json:"modelStoreType,omitempty"`
	ModelStoreClaimName  *string `json:"modelStoreClaimName,omitempty"`
	ModelStoreAccessMode *string `json:"modelStoreAccessMode,omitempty"`
	ModelStoreCacheName  *string `json:"modelStoreCacheName,omitempty"`
	ModelStoreProfile    *string `json:"modelStoreProfile,omitempty"`

	AutoscalingEnabled *bool   `json:"autoscalingEnabled,omitempty"`
	MinReplicas        *int32  `json:"minReplicas,omitempty"`
	MaxReplicas        *int32  `json:"maxReplicas,omitempty"`
	Metric             *string `json:"metric,omitempty"`
	MetricThreshold    *int32  `json:"metricThreshold,omitempty"`
	ScaleWindowSeconds *int32  `json:"scaleWindowSeconds,omitempty"`

	MultiNodeEnabled *bool  `json:"multiNodeEnabled,omitempty"`
	Workers          *int32 `json:"workers,omitempty"`

	Tolerations []nimclient.Toleration `json:"tolerations,omitempty"`

	ReadinessProbePath         *string `json:"readinessProbePath,omitempty"`
	ReadinessProbePort         *int32  `json:"readinessProbePort,omitempty"`
	ReadinessProbeInitialDelay *int32  `json:"readinessProbeInitialDelay,omitempty"`
	ReadinessProbePeriod       *int32  `json:"readinessProbePeriod,omitempty"`
	ReadinessProbeTimeout      *int32  `json:"readinessProbeTimeout,omitempty"`
	ReadinessProbeSuccess      *int32  `json:"readinessProbeSuccess,omitempty"`
	ReadinessProbeFailure      *int32  `json:"readinessProbeFailure,omitempty"`

	RunAsUser  *int64 `json:"runAsUser,omitempty"`
	RunAsGroup *int64 `json:"runAsGroup,omitempty"`

	EnvironmentVariables map[string]string `json:"environmentVariables,omitempty"`

	Labels      []*openapi.Label      `json:"labels,omitempty"`
	Annotations []*openapi.Annotation `json:"annotations,omitempty"`

	ReplicasRequested *int32 `json:"replicasRequested,omitempty"`
	ReplicasReady     *int32 `json:"replicasReady,omitempty"`
}

type Builder struct {
	fields Fields
}

func NewBuilder() *Builder {
	return &Builder{
		fields: Fields{},
	}
}

func (b *Builder) Build() *Fields {
	return &b.fields
}

func (b *Builder) Image(v string)              { b.fields.Image = &v }
func (b *Builder) NgcAuthSecret(v string)      { b.fields.NgcAuthSecret = &v }
func (b *Builder) Category(v string)           { b.fields.Category = &v }
func (b *Builder) PriorityClass(v string)      { b.fields.PriorityClass = &v }
func (b *Builder) Preemptibility(v string)     { b.fields.Preemptibility = &v }
func (b *Builder) ImagePullPolicy(v string)    { b.fields.ImagePullPolicy = &v }
func (b *Builder) ImagePullSecrets(v []string) { b.fields.ImagePullSecrets = v }

func (b *Builder) ServingPort(v int32)      { b.fields.ServingPort = &v }
func (b *Builder) GrpcPort(v int32)         { b.fields.GrpcPort = &v }
func (b *Builder) MetricsPort(v int32)      { b.fields.MetricsPort = &v }
func (b *Builder) ExposeExternally(v bool)  { b.fields.ExposeExternally = &v }
func (b *Builder) ExposedProtocol(v string) { b.fields.ExposedProtocol = &v }
func (b *Builder) ServingURL(v string)      { b.fields.ServingURL = &v }
func (b *Builder) ServiceType(v string)     { b.fields.ServiceType = &v }

func (b *Builder) GpuDevicesRequest(v int32)   { b.fields.GpuDevicesRequest = &v }
func (b *Builder) GpuPortionRequest(v float64) { b.fields.GpuPortionRequest = &v }
func (b *Builder) GpuPortionLimit(v float64)   { b.fields.GpuPortionLimit = &v }
func (b *Builder) GpuComputeMode(v string)     { b.fields.GpuComputeMode = &v }
func (b *Builder) GpuMemoryRequest(v string)   { b.fields.GpuMemoryRequest = &v }
func (b *Builder) GpuMemoryLimit(v string)     { b.fields.GpuMemoryLimit = &v }
func (b *Builder) CpuCoreRequest(v float64)    { b.fields.CpuCoreRequest = &v }
func (b *Builder) CpuCoreLimit(v float64)      { b.fields.CpuCoreLimit = &v }
func (b *Builder) CpuMemoryRequest(v string)   { b.fields.CpuMemoryRequest = &v }
func (b *Builder) CpuMemoryLimit(v string)     { b.fields.CpuMemoryLimit = &v }

func (b *Builder) ModelStoreType(v string)       { b.fields.ModelStoreType = &v }
func (b *Builder) ModelStoreClaimName(v string)  { b.fields.ModelStoreClaimName = &v }
func (b *Builder) ModelStoreAccessMode(v string) { b.fields.ModelStoreAccessMode = &v }
func (b *Builder) ModelStoreCacheName(v string)  { b.fields.ModelStoreCacheName = &v }
func (b *Builder) ModelStoreProfile(v string)    { b.fields.ModelStoreProfile = &v }

func (b *Builder) AutoscalingEnabled(v bool)  { b.fields.AutoscalingEnabled = &v }
func (b *Builder) MinReplicas(v int32)        { b.fields.MinReplicas = &v }
func (b *Builder) MaxReplicas(v int32)        { b.fields.MaxReplicas = &v }
func (b *Builder) Metric(v string)            { b.fields.Metric = &v }
func (b *Builder) MetricThreshold(v int32)    { b.fields.MetricThreshold = &v }
func (b *Builder) ScaleWindowSeconds(v int32) { b.fields.ScaleWindowSeconds = &v }

func (b *Builder) MultiNodeEnabled(v bool) { b.fields.MultiNodeEnabled = &v }
func (b *Builder) Workers(v int32)         { b.fields.Workers = &v }

func (b *Builder) Tolerations(v []nimclient.Toleration) { b.fields.Tolerations = v }

func (b *Builder) ReadinessProbePath(v string)        { b.fields.ReadinessProbePath = &v }
func (b *Builder) ReadinessProbePort(v int32)         { b.fields.ReadinessProbePort = &v }
func (b *Builder) ReadinessProbeInitialDelay(v int32) { b.fields.ReadinessProbeInitialDelay = &v }
func (b *Builder) ReadinessProbePeriod(v int32)       { b.fields.ReadinessProbePeriod = &v }
func (b *Builder) ReadinessProbeTimeout(v int32)      { b.fields.ReadinessProbeTimeout = &v }
func (b *Builder) ReadinessProbeSuccess(v int32)      { b.fields.ReadinessProbeSuccess = &v }
func (b *Builder) ReadinessProbeFailure(v int32)      { b.fields.ReadinessProbeFailure = &v }

func (b *Builder) RunAsUser(v int64)  { b.fields.RunAsUser = &v }
func (b *Builder) RunAsGroup(v int64) { b.fields.RunAsGroup = &v }

func (b *Builder) EnvironmentVariables(v map[string]string) { b.fields.EnvironmentVariables = v }
func (b *Builder) Labels(v []*openapi.Label)                { b.fields.Labels = v }
func (b *Builder) Annotations(v []*openapi.Annotation)      { b.fields.Annotations = v }

func BuildFromSpec(spec *nimclient.NimServiceSpec) *Fields {
	if spec == nil {
		return &Fields{}
	}

	builder := NewBuilder()

	if spec.Image.IsSet() {
		builder.Image(spec.GetImage())
	}
	if spec.NgcAuthSecret.IsSet() {
		builder.NgcAuthSecret(spec.GetNgcAuthSecret())
	}
	if spec.Category.IsSet() {
		builder.Category(spec.GetCategory())
	}
	if spec.PriorityClass.IsSet() {
		builder.PriorityClass(spec.GetPriorityClass())
	}
	if spec.Preemptibility.IsSet() {
		builder.Preemptibility(string(spec.GetPreemptibility()))
	}
	if spec.ImagePullPolicy.IsSet() {
		builder.ImagePullPolicy(string(spec.GetImagePullPolicy()))
	}
	if len(spec.ImagePullSecrets) > 0 {
		secrets := make([]string, len(spec.ImagePullSecrets))
		for i, s := range spec.ImagePullSecrets {
			if s.Name != nil {
				secrets[i] = *s.Name
			}
		}
		builder.ImagePullSecrets(secrets)
	}

	if spec.ServingPort.IsSet() {
		servingPort := spec.GetServingPort()
		if servingPort.Port.IsSet() {
			builder.ServingPort(servingPort.GetPort())
		}
		if servingPort.GrpcPort.IsSet() {
			builder.GrpcPort(servingPort.GetGrpcPort())
		}
		if servingPort.MetricsPort.IsSet() {
			builder.MetricsPort(servingPort.GetMetricsPort())
		}
		if servingPort.ExposeExternally.IsSet() {
			builder.ExposeExternally(servingPort.GetExposeExternally())
		}
		if servingPort.ExposedProtocol.IsSet() {
			builder.ExposedProtocol(string(servingPort.GetExposedProtocol()))
		}
		if servingPort.ExposedUrl.IsSet() {
			builder.ServingURL(servingPort.GetExposedUrl())
		}
		if servingPort.ServiceType.IsSet() {
			builder.ServiceType(string(servingPort.GetServiceType()))
		}
	}

	if spec.Compute.IsSet() {
		compute := spec.GetCompute()
		if compute.GpuDevicesRequest.IsSet() {
			builder.GpuDevicesRequest(compute.GetGpuDevicesRequest())
		}
		if value, _ := compute.GetGpuMemoryPortionRequestOk(); value != nil {
			builder.GpuPortionRequest(*value)
		} else if compute.GpuPortionRequest.IsSet() {
			builder.GpuPortionRequest(compute.GetGpuPortionRequest())
		}
		if value, _ := compute.GetGpuMemoryPortionLimitOk(); value != nil {
			builder.GpuPortionLimit(*value)
		} else if compute.GpuPortionLimit.IsSet() {
			builder.GpuPortionLimit(compute.GetGpuPortionLimit())
		}
		if mode, _ := compute.GetGpuComputeModeOk(); mode != nil {
			builder.GpuComputeMode(string(*mode))
		}
		if compute.GpuMemoryRequest.IsSet() {
			builder.GpuMemoryRequest(compute.GetGpuMemoryRequest())
		}
		if compute.GpuMemoryLimit.IsSet() {
			builder.GpuMemoryLimit(compute.GetGpuMemoryLimit())
		}
		if compute.CpuCoreRequest.IsSet() {
			builder.CpuCoreRequest(compute.GetCpuCoreRequest())
		}
		if compute.CpuCoreLimit.IsSet() {
			builder.CpuCoreLimit(compute.GetCpuCoreLimit())
		}
		if compute.CpuMemoryRequest.IsSet() {
			builder.CpuMemoryRequest(compute.GetCpuMemoryRequest())
		}
		if compute.CpuMemoryLimit.IsSet() {
			builder.CpuMemoryLimit(compute.GetCpuMemoryLimit())
		}
	}

	if spec.ModelStore.IsSet() {
		modelStore := spec.GetModelStore()
		if modelStore.Pvc.IsSet() {
			pvc := modelStore.GetPvc()
			builder.ModelStoreType("PVC")
			if pvc.ClaimName.IsSet() {
				builder.ModelStoreClaimName(pvc.GetClaimName())
			}
			if pvc.ClaimInfo.IsSet() {
				claimInfo := pvc.GetClaimInfo()
				if claimInfo.AccessModes.IsSet() {
					accessModes := claimInfo.GetAccessModes()
					if accessModes.GetReadWriteOnce() {
						builder.ModelStoreAccessMode("ReadWriteOnce")
					} else if accessModes.GetReadWriteMany() {
						builder.ModelStoreAccessMode("ReadWriteMany")
					} else if accessModes.GetReadOnlyMany() {
						builder.ModelStoreAccessMode("ReadOnlyMany")
					}
				}
			}
		} else if modelStore.NimCache.IsSet() {
			cache := modelStore.GetNimCache()
			builder.ModelStoreType("NIM Cache")
			if cache.Name.IsSet() {
				builder.ModelStoreCacheName(cache.GetName())
			}
			if cache.Profile.IsSet() {
				builder.ModelStoreProfile(cache.GetProfile())
			}
		}
	}

	if spec.Autoscaling.IsSet() {
		autoscaling := spec.GetAutoscaling()
		builder.AutoscalingEnabled(true)
		if autoscaling.MinReplicas.IsSet() {
			builder.MinReplicas(autoscaling.GetMinReplicas())
		}
		if autoscaling.MaxReplicas.IsSet() {
			builder.MaxReplicas(autoscaling.GetMaxReplicas())
		}
		if autoscaling.Metric.IsSet() {
			builder.Metric(autoscaling.GetMetric())
		}
		if autoscaling.MetricThreshold.IsSet() {
			builder.MetricThreshold(autoscaling.GetMetricThreshold())
		}
		if autoscaling.ScaleWindowSeconds.IsSet() {
			builder.ScaleWindowSeconds(autoscaling.GetScaleWindowSeconds())
		}
	} else {
		builder.AutoscalingEnabled(false)
	}

	if spec.MultiNode.IsSet() {
		multiNode := spec.GetMultiNode()
		if multiNode.Workers.IsSet() {
			builder.MultiNodeEnabled(true)
			builder.Workers(multiNode.GetWorkers())
		} else {
			builder.MultiNodeEnabled(false)
		}
	} else {
		builder.MultiNodeEnabled(false)
	}

	if len(spec.Tolerations) > 0 {
		builder.Tolerations(spec.Tolerations)
	}

	if spec.Probes.IsSet() {
		probes := spec.GetProbes()
		if probes.Readiness.IsSet() {
			readiness := probes.GetReadiness()
			if readiness.Handler.IsSet() {
				handler := readiness.GetHandler()
				if handler.HasHttpGet() {
					httpGet := handler.GetHttpGet()
					if httpGet.Path.IsSet() {
						builder.ReadinessProbePath(httpGet.GetPath())
					}
					if httpGet.Port.IsSet() {
						builder.ReadinessProbePort(httpGet.GetPort())
					}
				}
			}
			if readiness.InitialDelaySeconds.IsSet() {
				builder.ReadinessProbeInitialDelay(readiness.GetInitialDelaySeconds())
			}
			if readiness.PeriodSeconds.IsSet() {
				builder.ReadinessProbePeriod(readiness.GetPeriodSeconds())
			}
			if readiness.TimeoutSeconds.IsSet() {
				builder.ReadinessProbeTimeout(readiness.GetTimeoutSeconds())
			}
			if readiness.SuccessThreshold.IsSet() {
				builder.ReadinessProbeSuccess(readiness.GetSuccessThreshold())
			}
			if readiness.FailureThreshold.IsSet() {
				builder.ReadinessProbeFailure(readiness.GetFailureThreshold())
			}
		}
	}

	if spec.Security.IsSet() {
		security := spec.GetSecurity()
		if security.RunAsUid.IsSet() {
			builder.RunAsUser(security.GetRunAsUid())
		}
		if security.RunAsGid.IsSet() {
			builder.RunAsGroup(security.GetRunAsGid())
		}
	}

	if len(spec.EnvironmentVariables) > 0 {
		envs := make(map[string]string)
		for _, ev := range spec.EnvironmentVariables {
			value := ""
			if ev.Value.IsSet() {
				value = ev.GetValue()
			}
			envs[ev.GetName()] = value
		}
		builder.EnvironmentVariables(envs)
	}

	if len(spec.Labels) > 0 {
		labels := make([]*openapi.Label, len(spec.Labels))
		for i, l := range spec.Labels {
			name := l.GetName()
			value := ""
			if l.Value.IsSet() {
				value = l.GetValue()
			}
			labels[i] = &openapi.Label{
				Name:  *openapi.NewNullableString(&name),
				Value: *openapi.NewNullableString(&value),
			}
		}
		builder.Labels(labels)
	}
	if len(spec.Annotations) > 0 {
		annotations := make([]*openapi.Annotation, len(spec.Annotations))
		for i, a := range spec.Annotations {
			name := a.GetName()
			value := ""
			if a.Value.IsSet() {
				value = a.GetValue()
			}
			annotations[i] = &openapi.Annotation{
				Name:  *openapi.NewNullableString(&name),
				Value: *openapi.NewNullableString(&value),
			}
		}
		builder.Annotations(annotations)
	}

	return builder.Build()
}

func SetFields(target **Fields, fields *Fields) {
	*target = fields
}
