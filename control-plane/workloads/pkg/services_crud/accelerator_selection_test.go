/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"encoding/json"
	"testing"

	api "github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/stretchr/testify/require"
)

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
		nil, false, map[string]interface{}{"gpuType": model, "accelerator": map[string]interface{}{"resourceName": "amd.com/gpu", "deviceLabelKey": key}}, nil)
	require.Nil(t, policyErr)
	spec, _, buildErr := buildResponseSpecFromMap[workloadsApi.SupersetSpec, workloadsApi.MasterSpec](Training, effective)
	require.Nil(t, buildErr)
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
	warnings, err := resolveWorkloadAccelerators(context.Background(), nil, "cluster", &spec, nil)
	require.Nil(t, err)
	require.Empty(t, warnings)
	require.False(t, spec.HasGpuType())
	require.False(t, spec.HasAccelerator())
}

func TestAcceleratorRejectsExplicitEmptyModel(t *testing.T) {
	var spec workloadsApi.SupersetSpec
	spec.SetGpuType("")
	_, err := resolveWorkloadAccelerators(context.Background(), nil, "cluster", &spec, nil)
	require.NotNil(t, err)
	require.EqualValues(t, 400, err.Code)
}
