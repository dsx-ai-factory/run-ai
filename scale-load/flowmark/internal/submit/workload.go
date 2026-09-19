/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"maps"
	"slices"

	workloadclient "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
)

type WorkloadSpec struct {
	Name    string
	Project ProjectRef
	Image   string
	Command string
	GPUs    int32
	Labels  map[string]string
}

const (
	cpuCoreRequest   = 0.05
	cpuMemoryRequest = "1M"
)

func compute(gpus int32) workloadclient.NullableWorkspaceSpecSpecAllOfCompute {
	return *workloadclient.NewNullableWorkspaceSpecSpecAllOfCompute(&workloadclient.WorkspaceSpecSpecAllOfCompute{
		CpuCoreRequest:    *workloadclient.NewNullableFloat64(workloadclient.PtrFloat64(cpuCoreRequest)),
		CpuMemoryRequest:  *workloadclient.NewNullableString(workloadclient.PtrString(cpuMemoryRequest)),
		GpuDevicesRequest: *workloadclient.NewNullableInt32(workloadclient.PtrInt32(gpus)),
	})
}

func labels(values map[string]string) []workloadclient.Label {
	keys := slices.Sorted(maps.Keys(values))
	result := make([]workloadclient.Label, 0, len(keys))
	for _, key := range keys {
		result = append(result, workloadclient.Label{
			Name:  *workloadclient.NewNullableString(workloadclient.PtrString(key)),
			Value: *workloadclient.NewNullableString(workloadclient.PtrString(values[key])),
		})
	}
	return result
}

func nullableString(value string) workloadclient.NullableString {
	if value == "" {
		return workloadclient.NullableString{}
	}
	return *workloadclient.NewNullableString(workloadclient.PtrString(value))
}
