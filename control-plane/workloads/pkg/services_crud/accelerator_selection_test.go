/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"encoding/json"
	"testing"

	orgapi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	api "github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/stretchr/testify/require"
)

func TestAcceleratorProjectSelection(t *testing.T) {
	const nvidia = "nvidia.com/gpu"
	const amd = "amd.com/gpu"
	const nvidiaLabel = "nvidia.com/gpu.product"
	const amdLabel = "hardware.example.com/gpu-model"
	mixed := &orgapi.ProjectAccelerators{Items: []orgapi.AcceleratorOption{
		{ResourceName: nvidia, Devices: []orgapi.AcceleratorDevice{{Key: nvidiaLabel, Value: "H100"}}},
		{ResourceName: amd, Devices: []orgapi.AcceleratorDevice{{Key: amdLabel, Value: "MI300X"}}},
	}}
	tests := []struct {
		name, resource, model, key, wantKey, warning string
		wantError                                    bool
		inventory                                    *orgapi.ProjectAccelerators
	}{
		{name: "default NVIDIA H100", model: "H100", inventory: mixed},
		{name: "AMD any model", resource: amd, inventory: mixed},
		{name: "discover AMD key", resource: amd, model: "MI300X", wantKey: amdLabel, inventory: mixed},
		{name: "AMD does not inherit NVIDIA device", resource: amd, model: "H100", wantKey: amdLabel, warning: "GPU_TYPE_NOT_FOUND_IN_PROJECT_POOLS", inventory: mixed},
		{name: "explicit missing AMD selector", resource: amd, model: "MI400X", key: amdLabel, warning: "GPU_TYPE_NOT_FOUND_IN_PROJECT_POOLS", inventory: mixed},
		{name: "known vendor and absent model", model: "A100", warning: "GPU_TYPE_NOT_FOUND_IN_PROJECT_POOLS", inventory: mixed},
		{name: "lookup outage with default label", model: "H100", warning: "GPU_INVENTORY_UNAVAILABLE"},
		{name: "outage with explicit selector", resource: amd, model: "MI300X", key: amdLabel, warning: "GPU_INVENTORY_UNAVAILABLE"},
		{name: "outage with unknown label", resource: amd, model: "MI300X", wantError: true},
		{name: "invalid syntax", model: "invalid/value", wantError: true, inventory: mixed},
		{name: "reject CPU resource", resource: "cpu", wantError: true, inventory: mixed},
		{name: "ambiguous key", resource: amd, model: "MI300X", wantError: true, inventory: &orgapi.ProjectAccelerators{Items: []orgapi.AcceleratorOption{{ResourceName: amd, Devices: []orgapi.AcceleratorDevice{{Key: amdLabel, Value: "MI300X"}, {Key: "example.com/model", Value: "MI300X"}}}}}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var spec workloadsApi.SupersetSpec
			var selection *workloadsApi.AcceleratorSelection
			if tc.resource != "" || tc.key != "" {
				selection = workloadsApi.NewAcceleratorSelection()
				if tc.resource != "" {
					selection.SetResourceName(tc.resource)
				}
				if tc.key != "" {
					selection.SetDeviceLabelKey(tc.key)
				}
			}
			warning, err := resolveAccelerator("spec", &spec, selection, tc.model, tc.inventory)
			if tc.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tc.warning != "" {
				require.NotNil(t, warning)
				require.Equal(t, tc.warning, warning.Code)
			} else {
				require.Nil(t, warning)
			}
			if tc.wantKey != "" {
				got := spec.GetAccelerator()
				require.Equal(t, tc.wantKey, got.GetDeviceLabelKey())
			}
		})
	}
}

func TestAcceleratorLegacySubmission(t *testing.T) {
	var spec workloadsApi.SupersetSpec
	require.False(t, hasGpuSelection(nil))
	spec.SetGpuType("H100")
	require.NotNil(t, validateNativeGpuMode(&spec))
	var worker workloadsApi.DistributedInferenceLeaderWorkerSpec
	selection := workloadsApi.NewAcceleratorSelection()
	selection.SetResourceName("amd.com/gpu")
	worker.SetAccelerator(*selection)
	spec.UnsetGpuType()
	spec.SetWorker(worker)
	require.NotNil(t, validateNativeGpuMode(&spec))
}

func TestAcceleratorSavedSelection(t *testing.T) {
	const key = "example.com/model"
	const model = "MI300X"
	effective, policyErr := MutateWorkload(
		map[string]interface{}{"compute": map[string]interface{}{"gpuDevicesRequest": 1}},
		nil, false, map[string]interface{}{"gpuType": model, "accelerator": map[string]interface{}{"resourceName": "amd.com/gpu"}}, nil)
	require.Nil(t, policyErr)
	spec, _, buildErr := buildResponseSpecFromMap[workloadsApi.SupersetSpec, workloadsApi.MasterSpec](Training, effective)
	require.Nil(t, buildErr)
	lookup := &orgapi.ProjectAccelerators{Items: []orgapi.AcceleratorOption{{ResourceName: "amd.com/gpu", Devices: []orgapi.AcceleratorDevice{{Key: key, Value: model}}}}}
	warning, err := resolveAccelerator("spec", spec, spec.Accelerator.Get(), spec.GetGpuType(), lookup)
	require.NoError(t, err)
	require.Nil(t, warning)
	snapshot := NewWorkloadV2SnapshotFromWorkloadParams(&createWorkloadParams{Response: workloadsApi.SupersetWorkload{Spec: spec}}, nil, "research")
	copied := workloadFromV2Snapshot(snapshot, workloadsApi.WORKLOADDESIREDPHASE_RUNNING)
	require.Equal(t, model, copied.Spec.GetGpuType())
	require.Equal(t, key, copied.Spec.Accelerator.Get().GetDeviceLabelKey())
	require.Nil(t, mergeUpdateSpecIntoFullSpec(context.Background(), copied.Spec, map[string]interface{}{"image": "training:2"}))
	require.Equal(t, key, copied.Spec.Accelerator.Get().GetDeviceLabelKey())
	encoded, err := json.Marshal(copied.Spec)
	require.NoError(t, err)
	var request api.TrainingSpecSpec
	require.NoError(t, json.Unmarshal(encoded, &request))
	require.Equal(t, model, *request.GpuType)
	require.Equal(t, key, *request.Accelerator.DeviceLabelKey)
}

func TestAcceleratorClearWithoutNewInventoryDependency(t *testing.T) {
	spec := workloadsApi.SupersetSpec{}
	spec.SetGpuType("H100")
	require.Nil(t, mergeUpdateSpecIntoFullSpec(context.Background(), &spec, map[string]interface{}{"gpuType": "any"}))
	warnings, err := resolveWorkloadAccelerators(context.Background(), "project", &spec, nil)
	require.Nil(t, err)
	require.Empty(t, warnings)
	require.False(t, spec.HasGpuType())
	require.False(t, spec.HasAccelerator())
}

func TestAcceleratorRejectsExplicitEmptyModel(t *testing.T) {
	var spec workloadsApi.SupersetSpec
	spec.SetGpuType("")
	_, err := resolveWorkloadAccelerators(context.Background(), "project", &spec, nil)
	require.NotNil(t, err)
	require.EqualValues(t, 400, err.Code)
}
