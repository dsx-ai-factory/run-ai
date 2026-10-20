/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/clients/policies"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models_crud"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	policyApi "github.com/dsx-ai-factory/run-ai/runai-clients/policy-service-client/generated"
	"github.com/stretchr/testify/require"
)

func TestGpuModePolicyPipeline(t *testing.T) {
	for _, tc := range []struct {
		name, request, asset, template, def, want string
		rules                                     map[string]interface{}
		fail                                      bool
	}{
		{name: "fallback"},
		{name: "administrator default", def: "sm-sharing", want: "sm-sharing"},
		{name: "editable explicit wins", request: "time-slicing", def: "sm-sharing", want: "time-slicing"},
		{name: "locked equal", request: "sm-sharing", def: "sm-sharing", rules: map[string]interface{}{"canEdit": false}, want: "sm-sharing"},
		{name: "locked conflict", request: "time-slicing", def: "sm-sharing", rules: map[string]interface{}{"canEdit": false}, fail: true},
		{name: "asset precedes template and policy", asset: "sm-sharing", template: "time-slicing", def: "time-slicing", want: "sm-sharing"},
		{name: "template precedes policy", template: "sm-sharing", def: "time-slicing", want: "sm-sharing"},
		{name: "request precedes asset", request: "time-slicing", asset: "sm-sharing", want: "time-slicing"},
		{name: "asset mode obeys locked policy", asset: "time-slicing", def: "sm-sharing", rules: map[string]interface{}{"canEdit": false}, fail: true},
		{name: "required omission", rules: map[string]interface{}{"required": true}, fail: true},
		{name: "required default", def: "sm-sharing", rules: map[string]interface{}{"required": true}, want: "sm-sharing"},
		{name: "explicit forbidden option", request: "time-slicing", rules: map[string]interface{}{"options": []interface{}{map[string]interface{}{"value": "sm-sharing"}}}, fail: true},
		{name: "explicit blocked", request: "sm-sharing", rules: map[string]interface{}{"blocked": true}, fail: true},
		{name: "optional omission with options", rules: map[string]interface{}{"options": []interface{}{map[string]interface{}{"value": "sm-sharing"}}}},
		{name: "optional omission with blocked", rules: map[string]interface{}{"blocked": true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			compute := map[string]interface{}{"gpuMemoryPortionRequest": json.Number("0.25")}
			if tc.request != "" {
				compute["gpuComputeMode"] = tc.request
			}
			input := map[string]interface{}{"compute": compute}
			defaults := map[string]interface{}{}
			if tc.def != "" {
				defaults["compute"] = map[string]interface{}{"gpuComputeMode": tc.def}
			}
			rules := map[string]interface{}{"compute": map[string]interface{}{"gpuComputeMode": tc.rules}}
			asset := map[string]interface{}{}
			if tc.asset != "" {
				asset["compute"] = map[string]interface{}{"gpuComputeMode": tc.asset}
			}
			template := map[string]interface{}{}
			if tc.template != "" {
				template["compute"] = map[string]interface{}{"gpuComputeMode": tc.template}
			}
			mutated, err := MutateWorkload(input, defaults, false, template, asset)
			require.Nil(t, err)
			require.NotContains(t, mutated["compute"], "computeMode")
			validationErr := validateSpecAgainstPolicy(Workspace, rules, defaults, mutated)
			if tc.fail {
				require.NotNil(t, validationErr)
				require.Equal(t, int32(400), validationErr.Code)
				return
			}
			require.Nil(t, validationErr)
			if tc.want == "" {
				require.NotContains(t, mutated["compute"], "gpuComputeMode")
			}
			spec, _, buildErr := buildResponseSpecFromMap[workloadsApi.SupersetSpec, workloadsApi.MasterSpec](Workspace, mutated)
			require.Nil(t, buildErr)
			params := createWorkloadParams{Response: workloadsApi.SupersetWorkload{Spec: spec}}
			stored := NewWorkloadFromWorkloadParams(&params, nil)
			response := WorkloadFromDbRecord(stored, workloadsApi.PHASE_CREATING)
			require.Equal(t, tc.want, string(response.Spec.Compute.Get().GetGpuComputeMode()))
			snapshot := NewWorkloadV2SnapshotFromWorkloadParams(&params, nil, "")
			require.Equal(t, tc.want, string(snapshot.Spec.Compute.Get().GetGpuComputeMode()))
			data, marshalErr := json.Marshal(response)
			require.NoError(t, marshalErr)
			require.NotContains(t, string(data), `"computeMode"`)
		})
	}
}

func TestGpuModeUpdateAndLegacyGuard(t *testing.T) {
	var full workloadsApi.SupersetSpec
	require.NoError(t, json.Unmarshal([]byte(`{"compute":{"gpuDevicesRequest":1,"gpuComputeMode":"sm-sharing"}}`), &full))
	require.NotNil(t, validateNativeGpuMode(&full))
	var update api.InferenceUpdateSpecSpec
	require.NoError(t, json.Unmarshal([]byte(`{"compute":{"gpuDevicesRequest":1}}`), &update))
	require.Nil(t, mergeUpdateSpecIntoFullSpec(context.Background(), &full, &update))
	require.False(t, full.Compute.Get().GpuComputeMode.IsSet())
	require.Nil(t, validateNativeGpuMode(&full))
	require.Nil(t, validateNativeGpuMode((*workloadsApi.MasterSpec)(nil)))
}

func TestGpuModesForReplicas(t *testing.T) {
	input := map[string]interface{}{
		"worker": map[string]interface{}{"compute": map[string]interface{}{"gpuMemoryPortionRequest": 0.25, "gpuComputeMode": "sm-sharing"}},
		"master": map[string]interface{}{"compute": map[string]interface{}{"gpuDevicesRequest": 1, "gpuComputeMode": "time-slicing"}},
	}
	for _, same := range []bool{false, true} {
		t.Run(fmt.Sprintf("masterSameAsWorker=%v", same), func(t *testing.T) {
			request := input
			if same {
				request = map[string]interface{}{"worker": input["worker"], "master": input["worker"]}
			}
			mutated, err := MutateWorkload(request, nil, same, nil, nil)
			require.Nil(t, err)
			worker, master, buildErr := buildResponseSpecFromMap[workloadsApi.SupersetSpec, workloadsApi.MasterSpec](Distributed, mutated)
			require.Nil(t, buildErr)
			require.Equal(t, workloadsApi.GPUCOMPUTEMODE_SM_SHARING, worker.Compute.Get().GetGpuComputeMode())
			expected := workloadsApi.GPUCOMPUTEMODE_TIME_SLICING
			if same {
				expected = workloadsApi.GPUCOMPUTEMODE_SM_SHARING
			}
			require.Equal(t, expected, master.Compute.Get().GetGpuComputeMode())
			if same {
				require.Nil(t, preValidateMpiGpuFields(context.Background(), worker, true, master))
				require.False(t, master.Compute.Get().GpuDevicesRequest.IsSet())
				require.Equal(t, workloadsApi.GPUCOMPUTEMODE_SM_SHARING, worker.Compute.Get().GetGpuComputeMode())
			}
		})
	}
}

func TestInferenceUpdateModePolicy(t *testing.T) {
	originalPolicies := policies.Get()
	t.Cleanup(func() { policies.SetMock(originalPolicies) })
	for _, tc := range []struct {
		name, mode, defaultMode string
		rule                    map[string]interface{}
		fail                    bool
	}{
		{name: "mode default", defaultMode: "sm-sharing"},
		{name: "optional options", rule: map[string]interface{}{"options": []interface{}{map[string]interface{}{"value": "sm-sharing"}}}},
		{name: "blocked omission", rule: map[string]interface{}{"blocked": true}},
		{name: "required omission", rule: map[string]interface{}{"required": true}, fail: true},
		{name: "required default", defaultMode: "sm-sharing", rule: map[string]interface{}{"required": true}},
		{name: "locked conflict", mode: "time-slicing", defaultMode: "sm-sharing", rule: map[string]interface{}{"canEdit": false}, fail: true},
		{name: "locked equal", mode: "sm-sharing", defaultMode: "sm-sharing", rule: map[string]interface{}{"canEdit": false}},
	} {
		for _, computeUpdated := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/computeUpdated=%v", tc.name, computeUpdated), func(t *testing.T) {
				defaults := map[string]interface{}{"cpuCoreRequest": 8}
				if tc.defaultMode != "" {
					defaults["gpuComputeMode"] = tc.defaultMode
				}
				var policy policyApi.InferencePolicyDefaultsAndRulesV2
				require.NoError(t, convutils.CopySpecs(&policy, map[string]interface{}{
					"defaults": map[string]interface{}{"compute": defaults},
					"rules":    map[string]interface{}{"compute": map[string]interface{}{"gpuComputeMode": tc.rule}},
				}))
				policies.SetMock(policies.SimpleMockClientFactory{SimpleMockClient: policies.SimpleMockClient{EffectiveInferencePolicyV2: &policy}})
				var spec workloadsApi.SupersetSpec
				require.NoError(t, json.Unmarshal([]byte(`{"image":"test-image","compute":{"gpuDevicesRequest":1}}`), &spec))
				if tc.mode != "" {
					spec.Compute.Get().SetGpuComputeMode(workloadsApi.GpuComputeMode(tc.mode))
				}
				var update api.InferenceUpdateSpecSpec
				if computeUpdated {
					require.NoError(t, convutils.CopySpecs(&update.Compute, spec.Compute.Get()))
				}
				require.Nil(t, mergeUpdateSpecIntoFullSpec(context.Background(), &spec, &update))
				before, marshalErr := json.Marshal(spec)
				require.NoError(t, marshalErr)
				workload := &models_crud.Workload{ProjectId: "1", WorkloadType: Inference, Spec: &spec}
				_, err := validatePolicy(context.Background(), workload)
				after, marshalErr := json.Marshal(workload.Spec)
				require.NoError(t, marshalErr)
				require.JSONEq(t, string(before), string(after))
				if tc.fail {
					require.NotNil(t, err)
					require.Equal(t, int32(400), err.Code)
					return
				}
				require.Nil(t, err)
				want := tc.mode
				require.Equal(t, want, string(workload.Spec.Compute.Get().GetGpuComputeMode()))
				require.Equal(t, want != "", workload.Spec.Compute.Get().GpuComputeMode.IsSet())
				require.False(t, workload.Spec.Compute.Get().CpuCoreRequest.IsSet())
				require.Equal(t, "test-image", workload.Spec.GetImage())
			})
		}
	}
}
