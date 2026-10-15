/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gpucompat

import (
	"fmt"

	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
)

// NormalizeLegacyWorkloadCompute also supports callers that invoke composers
// directly, without passing through the HTTP source-normalization pipeline.
func NormalizeLegacyWorkloadCompute(source *workloadsApi.WorkspaceSpecSpecAllOfCompute) (*workloadsApi.WorkspaceSpecSpecAllOfCompute, error) {
	if source == nil {
		return nil, nil
	}
	result := *source
	for _, pair := range []struct {
		legacy, canonical *workloadsApi.NullableFloat64
		name              string
	}{
		{&result.GpuPortionRequest, &result.GpuMemoryPortionRequest, "Request"},
		{&result.GpuPortionLimit, &result.GpuMemoryPortionLimit, "Limit"},
	} {
		old, current := pair.legacy.Get(), pair.canonical.Get()
		if old != nil && current != nil && *old != *current {
			return nil, fmt.Errorf("compute.gpuPortion%s and compute.gpuMemoryPortion%s conflict: provide one alias or equal values", pair.name, pair.name)
		}
		if current == nil && pair.legacy.IsSet() {
			*pair.canonical = *pair.legacy
		}
		pair.legacy.Unset()
	}
	return &result, nil
}

// LegacyWorkloadSpec uses legacy JSON keys for existing DB records and composers.
func LegacyWorkloadSpec(source *workloadsApi.SupersetSpec) *workloadsApi.SupersetSpec {
	return projectWorkloadSpec(source, false)
}

func LegacyCompatibleWorkloadSpec(source *workloadsApi.SupersetSpec) *workloadsApi.SupersetSpec {
	return projectWorkloadSpec(source, true)
}

func LegacyMasterSpec(source *workloadsApi.MasterSpec) *workloadsApi.MasterSpec {
	return projectMasterSpec(source, false)
}

func LegacyCompatibleMasterSpec(source *workloadsApi.MasterSpec) *workloadsApi.MasterSpec {
	return projectMasterSpec(source, true)
}

func projectWorkloadSpec(source *workloadsApi.SupersetSpec, response bool) *workloadsApi.SupersetSpec {
	if source == nil {
		return nil
	}
	result := *source
	projectCompute(&result.Compute, response)
	for _, section := range []*workloadsApi.NullableDistributedInferenceLeaderWorkerSpec{&result.Leader, &result.Worker} {
		if section.Get() != nil {
			replica := *section.Get()
			projectCompute(&replica.Compute, response)
			section.Set(&replica)
		}
	}
	return &result
}

func projectMasterSpec(source *workloadsApi.MasterSpec, response bool) *workloadsApi.MasterSpec {
	if source == nil {
		return nil
	}
	result := *source
	projectCompute(&result.Compute, response)
	return &result
}

func projectCompute(section *workloadsApi.NullableWorkspaceSpecSpecAllOfCompute, response bool) {
	if section.Get() == nil {
		return
	}
	result := *section.Get()
	for i, pair := range []struct{ legacy, canonical *workloadsApi.NullableFloat64 }{
		{&result.GpuPortionRequest, &result.GpuMemoryPortionRequest},
		{&result.GpuPortionLimit, &result.GpuMemoryPortionLimit},
	} {
		value := pair.canonical
		if value.Get() == nil && pair.legacy.IsSet() {
			value = pair.legacy
		}
		if value.IsSet() {
			var number *float64
			if value.Get() != nil {
				n := *value.Get()
				number = &n
			}
			pair.legacy.Set(number)
			if response {
				if number != nil {
					n := *number
					pair.canonical.Set(&n)
				} else {
					pair.canonical.Set(nil)
				}
			}
		}
		if !response || (i == 0 && pair.canonical.Get() != nil && *pair.canonical.Get() > 1) {
			pair.canonical.Unset()
		}
	}
	section.Set(&result)
}
