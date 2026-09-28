/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloads

import (
	workloads "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"

	"github.com/dsx-ai-factory/run-ai/runai-mcp-server/internal/tools/common"
)

type measurement struct {
	Type   string             `json:"type"             jsonschema:"Metric type. Resource: GPU_UTILIZATION, GPU_MEMORY_USAGE_BYTES, CPU_USAGE_CORES, CPU_MEMORY_USAGE_BYTES, NVLINK_BANDWIDTH_TOTAL. NIM serving: NIM_NUM_REQUESTS_RUNNING, NIM_REQUEST_SUCCESS_TOTAL, NIM_REQUEST_FAILURE_TOTAL, NIM_GPU_CACHE_USAGE_PERC, NIM_E2E_REQUEST_LATENCY_SECONDS. Inference workloads also return THROUGHPUT and LATENCY. Per-pod GPU profiling (get_workload_pods only): GPU_SM_ACTIVITY_PER_GPU, GPU_SM_OCCUPANCY_PER_GPU, GPU_TENSOR_ACTIVITY_PER_GPU, GPU_MEMORY_BANDWIDTH_UTILIZATION_PER_GPU, GPU_NVLINK_TRANSMITTED_BANDWIDTH_PER_GPU, GPU_PCIE_TRANSMITTED_BANDWIDTH_PER_GPU"` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
	Labels map[string]string  `json:"labels,omitempty" jsonschema:"Optional metric labels forwarded verbatim from upstream"`
	Values []measurementValue `json:"values"           jsonschema:"Time-series samples for this metric, ordered as upstream returns them"`
}

type measurementValue struct {
	Value     string `json:"value"     jsonschema:"Sample value as a string, exactly as upstream returns it (no unit conversion)"`
	Timestamp string `json:"timestamp" jsonschema:"Sample timestamp (RFC3339)"`
}

type resourceCores struct {
	Request *float64 `json:"request,omitempty" jsonschema:"Floor — minimum the scheduler must reserve for the pod"`
	Limit   *float64 `json:"limit,omitempty"   jsonschema:"Ceiling — maximum the pod is allowed to consume"`
}

type resourceQuantity struct {
	Request string `json:"request,omitempty" jsonschema:"Floor as a Kubernetes quantity string (e.g. \"2Gi\")"`
	Limit   string `json:"limit,omitempty"   jsonschema:"Ceiling as a Kubernetes quantity string (e.g. \"4Gi\")"`
}

type extendedResource struct {
	Resource string `json:"resource,omitempty" jsonschema:"Extended resource name (e.g. \"nvidia.com/mig-1g.5gb\")"`
	Quantity string `json:"quantity,omitempty" jsonschema:"Requested quantity for this extended resource"`
}

type podRequestedResources struct {
	GpuRequestType    string             `json:"gpuRequestType,omitempty"    jsonschema:"How GPU is requested by the workload: \"portion\" | \"memory\""`
	Gpu               *resourceCores     `json:"gpu,omitempty"               jsonschema:"GPU request as a {request, limit} pair (units depend on gpuRequestType)"`
	GpuMemory         *resourceQuantity  `json:"gpuMemory,omitempty"         jsonschema:"GPU memory request as a {request, limit} pair of quantity strings"`
	Cpu               *resourceCores     `json:"cpu,omitempty"               jsonschema:"CPU cores request as a {request, limit} pair"`
	CpuMemory         *resourceQuantity  `json:"cpuMemory,omitempty"         jsonschema:"CPU memory request as a {request, limit} pair of quantity strings"`
	ExtendedResources []extendedResource `json:"extendedResources,omitempty" jsonschema:"Extra resources (e.g. NICs, MIG slices) requested with their quantity"`
}

type podAllocatedResources struct {
	Gpu               *float64           `json:"gpu,omitempty"               jsonschema:"GPUs actually granted to the pod (single number; floor/limit do NOT apply here)"`
	GpuMemory         string             `json:"gpuMemory,omitempty"         jsonschema:"GPU memory granted as a Kubernetes quantity string (e.g. \"2Gi\")"`
	Cpu               *float64           `json:"cpu,omitempty"               jsonschema:"CPU cores actually granted to the pod (single number)"`
	CpuMemory         string             `json:"cpuMemory,omitempty"         jsonschema:"CPU memory granted as a Kubernetes quantity string"`
	ExtendedResources []extendedResource `json:"extendedResources,omitempty" jsonschema:"Extra resources granted with their quantity"`
}

type podOutput struct {
	Name               string                 `json:"name"                       jsonschema:"Pod name"`
	ID                 string                 `json:"id"                         jsonschema:"Pod ID (UUID)"`
	NodeName           string                 `json:"nodeName,omitempty"         jsonschema:"Name of the node the pod runs on"`
	CurrentNodePool    string                 `json:"currentNodePool,omitempty"  jsonschema:"Node pool the pod is currently scheduled to"`
	K8sPhase           string                 `json:"k8sPhase"                   jsonschema:"Kubernetes pod phase (Pending | Running | Succeeded | Failed | Unknown)"`
	CreatedAt          string                 `json:"createdAt"                  jsonschema:"Pod creation timestamp (RFC3339)"`
	RevisionID         string                 `json:"revisionId,omitempty"       jsonschema:"Inference revision ID (UUID); set only for inference workload pods"`
	ClusterID          string                 `json:"clusterId,omitempty"        jsonschema:"Cluster ID the pod belongs to"`
	RequestedResources *podRequestedResources `json:"requestedResources,omitempty" jsonschema:"What the workload asked for: each dimension is a {request, limit} pair. Compare against allocatedResources to see whether the pod got what it asked for"` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
	AllocatedResources *podAllocatedResources `json:"allocatedResources,omitempty" jsonschema:"What the scheduler granted: each dimension is a single number / quantity (no request/limit pair)"`

	Measurements *[]measurement `json:"measurements,omitempty" jsonschema:"Latest per-pod measurements (10-min server-side window, numberOfSamples=2). Populated only by get_workload_pods — get_node_pods always omits this field. In get_workload_pods, also omitted when the metrics fetch failed or when this pod is returned as part of a revision (revision pods do not carry measurements); [] when the fetch succeeded with no samples"` //nolint:lll // long jsonschema tool description; cannot be wrapped in a struct tag
}

func mapPodOutput(p workloads.Pod) podOutput {
	out := podOutput{
		Name:               p.Name,
		ID:                 p.Id,
		ClusterID:          p.ClusterId,
		K8sPhase:           p.K8sPhase,
		CreatedAt:          common.TimestampString(&p.CreatedAt),
		RequestedResources: mapRequestedResources(p.RequestedResources.Get()),
		AllocatedResources: mapAllocatedResources(p.AllocatedResources.Get()),
	}
	if v := p.NodeName.Get(); v != nil {
		out.NodeName = *v
	}
	if v := p.CurrentNodePool.Get(); v != nil {
		out.CurrentNodePool = *v
	}
	if v := p.RevisionId.Get(); v != nil {
		out.RevisionID = *v
	}
	return out
}

func mapRequestedResources(src *workloads.PodRequestResources) *podRequestedResources {
	if src == nil {
		return nil
	}
	out := &podRequestedResources{
		Gpu:               mapResourceCores(src.Gpu.Get()),
		GpuMemory:         mapResourceQuantity(src.GpuMemory.Get()),
		Cpu:               mapResourceCores(src.Cpu.Get()),
		CpuMemory:         mapResourceQuantity(src.CpuMemory.Get()),
		ExtendedResources: mapExtendedResources(src.ExtendedResources),
	}
	if v := src.GpuRequestType.Get(); v != nil {
		out.GpuRequestType = string(*v)
	}
	return out
}

func mapAllocatedResources(src *workloads.AllocatedResources) *podAllocatedResources {
	if src == nil {
		return nil
	}
	out := &podAllocatedResources{
		Gpu:               src.Gpu.Get(),
		Cpu:               src.Cpu.Get(),
		ExtendedResources: mapExtendedResources(src.ExtendedResources),
	}
	if v := src.GpuMemory.Get(); v != nil {
		out.GpuMemory = *v
	}
	if v := src.CpuMemory.Get(); v != nil {
		out.CpuMemory = *v
	}
	return out
}

func mapResourceCores(src *workloads.RequestResourceCores) *resourceCores {
	if src == nil {
		return nil
	}
	return &resourceCores{
		Request: src.Request.Get(),
		Limit:   src.Limit.Get(),
	}
}

func mapResourceQuantity(src *workloads.RequestResourceQuantity) *resourceQuantity {
	if src == nil {
		return nil
	}
	out := &resourceQuantity{}
	if v := src.Request.Get(); v != nil {
		out.Request = *v
	}
	if v := src.Limit.Get(); v != nil {
		out.Limit = *v
	}
	return out
}

func mapExtendedResources(src []workloads.WorkloadsExtendedResource) []extendedResource {
	if len(src) == 0 {
		return nil
	}
	out := make([]extendedResource, 0, len(src))
	for _, r := range src {
		entry := extendedResource{}
		if v := r.Resource.Get(); v != nil {
			entry.Resource = *v
		}
		if v := r.Quantity.Get(); v != nil {
			entry.Quantity = *v
		}
		out = append(out, entry)
	}
	return out
}

// mapMeasurements converts the generated client's measurement series into our JSON output shape
func mapMeasurements(src []workloads.MeasurementResponse) []measurement {
	out := make([]measurement, 0, len(src))
	for _, m := range src {
		values := make([]measurementValue, 0, len(m.Values))
		for _, v := range m.Values {
			values = append(values, measurementValue{
				Value:     v.Value,
				Timestamp: common.TimestampString(v.Timestamp.Get()),
			})
		}
		out = append(out, measurement{
			Type:   m.Type,
			Labels: m.Labels,
			Values: values,
		})
	}
	return out
}
