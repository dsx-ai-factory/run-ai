/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_crud

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/gpucompat"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	workloadstest "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/jsonutils/gojsonq"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

func TestWorkspaceMemoryAliasPolicyPipeline(t *testing.T) {
	for _, legacyDefault := range []bool{true, false} {
		for _, locked := range []bool{true, false} {
			for _, value := range []float64{.25, .5} {
				t.Run(fmt.Sprintf("legacy-default=%t/locked=%t/value=%v", legacyDefault, locked, value), func(t *testing.T) {
					defaultKey, inputKey := "gpuPortionRequest", "gpuMemoryPortionRequest"
					if !legacyDefault {
						defaultKey, inputKey = inputKey, defaultKey
					}
					var request api.WorkspaceCreationRequest
					body := fmt.Sprintf(`{"spec":{"image":"example/image","compute":{"gpuDevicesRequest":1,"gpuRequestType":"portion",%q:%v}}}`, inputKey, value)
					if err := json.Unmarshal([]byte(body), &request); err != nil {
						t.Fatal(err)
					}
					input, err := gojsonq.LoadJson(request.Spec)
					if err != nil {
						t.Fatal(err)
					}
					defaults := map[string]any{"compute": map[string]any{defaultKey: json.Number("0.25")}}
					rules := map[string]any{"compute": map[string]any{defaultKey: map[string]any{"canEdit": !locked}}}
					mutated, mutateErr := MutateWorkload(input, defaults, false, nil, nil)
					if mutateErr != nil {
						t.Fatal(mutateErr)
					}
					validationErr := validateSpecAgainstPolicy(Workspace, rules, defaults, mutated)
					if locked && value != .25 {
						if validationErr == nil || validationErr.Code != 400 {
							t.Fatalf("lock bypassed: %v", validationErr)
						}
						return
					}
					if validationErr != nil {
						t.Fatal(validationErr)
					}
					if defaults["compute"].(map[string]any)[defaultKey] != json.Number("0.25") {
						t.Fatal("policy default changed")
					}
					spec, _, rebuildErr := buildResponseSpecFromMap[workloadsApi.SupersetSpec, workloadsApi.MasterSpec](Workspace, mutated)
					if rebuildErr != nil {
						t.Fatal(rebuildErr)
					}
					compute := spec.Compute.Get()
					if compute.GpuMemoryPortionRequest.Get() == nil || *compute.GpuMemoryPortionRequest.Get() != value {
						t.Fatal("canonical value lost")
					}
					if compute.GpuPortionRequest.IsSet() {
						t.Fatal("deprecated value in canonical spec")
					}
					assertWorkspaceMemoryAliasRoundTrip(t, spec, value)
				})
			}
		}
	}
}

func assertWorkspaceMemoryAliasRoundTrip(t *testing.T, spec *workloadsApi.SupersetSpec, value float64) {
	t.Helper()
	compute := spec.Compute.Get()
	var sizing composer.GpuSizingInfo
	var resources corev1.ResourceRequirements
	fc := composer.FieldsComposer{}
	err := fc.ComposeGpuRequestsAndLimits(context.Background(), compute.GpuDevicesRequest.Get(), compute.GpuRequestType.Get(), compute.GpuMemoryPortionRequest.Get(), compute.GpuMemoryPortionLimit.Get(), compute.GpuMemoryRequest.Get(), compute.GpuMemoryLimit.Get(), composer.WholeGpusToK8sResources, &resources, nil, &sizing)()
	if err != nil {
		t.Fatal(err)
	}
	if sizing.PortionRequest == nil || *sizing.PortionRequest != value {
		t.Fatal("native allocation lost")
	}
	params := createWorkloadParams{Response: workloadsApi.SupersetWorkload{Spec: spec}}
	record := NewWorkloadFromWorkloadParams(&params, nil)
	stored, marshalErr := json.Marshal(gpucompat.LegacyWorkloadSpec(record.Spec))
	if marshalErr != nil {
		t.Fatal(marshalErr)
	}
	if strings.Contains(string(stored), "gpuMemoryPortion") || !strings.Contains(string(stored), "gpuPortionRequest") {
		t.Fatalf("DB format changed: %s", stored)
	}
	response := WorkloadFromDbRecord(record, workloadsApi.PHASE_CREATING)
	returned := response.Spec.Compute.Get()
	if returned.GetGpuPortionRequest() != value || returned.GetGpuMemoryPortionRequest() != value {
		t.Fatal("response aliases differ")
	}
	*returned.GpuPortionRequest.Get() = .9
	if record.Spec.Compute.Get().GetGpuMemoryPortionRequest() != value || spec.Compute.Get().GetGpuMemoryPortionRequest() != value {
		t.Fatal("response mutated stored/canonical values")
	}
}

func TestWorkspaceMemoryAliasSourcePrecedence(t *testing.T) {
	input := map[string]any{"compute": map[string]any{"gpuMemoryPortionRequest": json.Number("0.5")}}
	assets := map[string]any{"compute": map[string]any{"gpuPortionRequest": json.Number("0.4")}}
	template := map[string]any{"compute": map[string]any{"gpuMemoryPortionRequest": json.Number("0.3")}}
	defaults := map[string]any{"compute": map[string]any{"gpuPortionRequest": json.Number("0.25")}}
	for _, tc := range []struct {
		name                    string
		input, assets, template map[string]any
		want                    float64
	}{
		{"request", input, assets, template, .5}, {"asset", nil, assets, template, .4}, {"template", nil, nil, template, .3}, {"policy", nil, nil, nil, .25},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := MutateWorkload(tc.input, defaults, false, tc.template, tc.assets)
			if err != nil {
				t.Fatal(err)
			}
			if fmt.Sprint(got["compute"].(map[string]any)["gpuMemoryPortionRequest"]) != fmt.Sprint(tc.want) {
				t.Fatalf("unexpected source precedence: %#v", got)
			}
		})
	}
}

func TestWorkspaceMemoryAliasConflicts(t *testing.T) {
	for _, source := range []string{"request", "policy", "template", "asset"} {
		t.Run(source, func(t *testing.T) {
			conflict := map[string]any{"compute": map[string]any{"gpuPortionRequest": .25, "gpuMemoryPortionRequest": json.Number("0.5")}}
			var input, defaults, template, assets map[string]any
			switch source {
			case "request":
				input = conflict
			case "policy":
				defaults = conflict
			case "template":
				template = conflict
			case "asset":
				assets = conflict
			}
			_, err := MutateWorkload(input, defaults, false, template, assets)
			if err == nil || !strings.Contains(err.Message, "conflict") {
				t.Fatalf("missing alias conflict: %v", err)
			}
		})
	}
}

func TestWorkspaceMemoryAliasUpdateSemantics(t *testing.T) {
	for _, tc := range []struct {
		name, input    string
		request, limit *float64
		conflict       bool
	}{
		{name: "omitted", input: `{}`, request: gpuMemoryFloat(.25), limit: gpuMemoryFloat(.5)},
		{name: "null", input: `{"compute":null}`, request: gpuMemoryFloat(.25), limit: gpuMemoryFloat(.5)},
		{name: "empty section", input: `{"compute":{}}`},
		{name: "new replaces section", input: `{"compute":{"gpuMemoryPortionRequest":0.75}}`, request: gpuMemoryFloat(.75)},
		{name: "old replaces section", input: `{"compute":{"gpuPortionRequest":0.75}}`, request: gpuMemoryFloat(.75)},
		{name: "explicit zero", input: `{"compute":{"gpuMemoryPortionRequest":0}}`, request: gpuMemoryFloat(0)},
		{name: "conflict", input: `{"compute":{"gpuPortionRequest":0.25,"gpuMemoryPortionRequest":0.75}}`, conflict: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var original workloadsApi.SupersetSpec
			if err := json.Unmarshal([]byte(`{"compute":{"gpuPortionRequest":0.25,"gpuPortionLimit":0.5},"image":"image"}`), &original); err != nil {
				t.Fatal(err)
			}
			var update api.InferenceUpdateSpecSpec
			if err := json.Unmarshal([]byte(tc.input), &update); err != nil {
				t.Fatal(err)
			}
			err := copyUpdateSpecIntoFullSpec(context.Background(), &original, &update)
			if tc.conflict {
				if err == nil {
					t.Fatal("conflict accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			got := original.GetCompute()
			if !reflect.DeepEqual(got.GpuMemoryPortionRequest.Get(), tc.request) || !reflect.DeepEqual(got.GpuMemoryPortionLimit.Get(), tc.limit) {
				t.Fatalf("unexpected section merge: %#v", got)
			}
			if original.GetImage() != "image" {
				t.Fatal("lost unrelated field")
			}
		})
	}
}

func gpuMemoryFloat(value float64) *float64 { return &value }

func TestGPUMemoryTemplateProjectionAndReset(t *testing.T) {
	var input api.SupersetTemplate
	if err := json.Unmarshal([]byte(`{"spec":{"compute":{"gpuPortionRequest":0.25}},"combinedSpec":{"compute":{"gpuPortionRequest":0.25}}}`), &input); err != nil {
		t.Fatal(err)
	}
	canonical, err := normalizeTemplateGPUMemory(&input)
	if err != nil {
		t.Fatal(err)
	}
	if canonical.Spec.Compute.GpuMemoryPortionRequest == nil || canonical.Spec.Compute.GpuPortionRequest != nil {
		t.Fatal("template not canonical")
	}
	projected, err := legacyCompatibleTemplate(canonical)
	if err != nil {
		t.Fatal(err)
	}
	if *projected.Spec.Compute.GpuPortionRequest != .25 || *projected.CombinedSpec.Compute.GpuMemoryPortionRequest != .25 {
		t.Fatal("template response aliases differ")
	}
	request := &UnifiedRequest{template: canonical}
	if err := request.ClearFields(updateOptions{fieldsToClear: []string{"spec.compute.gpuPortionRequest"}}); err != nil {
		t.Fatal(err)
	}
	if request.template.Spec.Compute != nil && request.template.Spec.Compute.GpuMemoryPortionRequest != nil {
		t.Fatal("deprecated reset path missed canonical field")
	}
	if input.Spec.Compute.GpuPortionRequest == nil || *input.Spec.Compute.GpuPortionRequest != .25 {
		t.Fatal("template input was mutated")
	}
}

func TestLegacyWholeWorkspacePipeline(t *testing.T) {
	for _, source := range []string{"request", "policy", "template"} {
		t.Run(source, func(t *testing.T) {
			var input, defaults, template map[string]any
			values := map[string]any{"compute": map[string]any{"gpuPortionRequest": json.Number("4"), "gpuRequestType": "portion"}}
			switch source {
			case "request":
				input = values
			case "policy":
				defaults = values
			case "template":
				template = values
			}
			mutated, err := MutateWorkload(input, defaults, false, template, nil)
			require.Nil(t, err)
			rules := map[string]any{"compute": map[string]any{"gpuPortionRequest": map[string]any{"canEdit": false, "required": true, "min": 2.0, "max": 8.0}}}
			require.Nil(t, validateSpecAgainstPolicy(Workspace, rules, values, mutated))
			spec, _, buildErr := buildResponseSpecFromMap[workloadsApi.SupersetSpec, workloadsApi.MasterSpec](Workspace, mutated)
			require.Nil(t, buildErr)
			var sizing composer.GpuSizingInfo
			var resources corev1.ResourceRequirements
			compute := spec.Compute.Get()
			fc := composer.FieldsComposer{}
			require.Nil(t, fc.ComposeGpuRequestsAndLimits(context.Background(), compute.GpuDevicesRequest.Get(), compute.GpuRequestType.Get(), compute.GpuMemoryPortionRequest.Get(), nil, nil, nil, composer.WholeGpusToK8sResources, &resources, nil, &sizing)())
			quantity := resources.Limits[corev1.ResourceName("nvidia.com/gpu")]
			require.Equal(t, int64(4), quantity.Value())
			stored := NewWorkloadFromWorkloadParams(&createWorkloadParams{Response: workloadsApi.SupersetWorkload{Spec: spec}}, nil)
			response := WorkloadFromDbRecord(stored, workloadsApi.PHASE_CREATING)
			require.Equal(t, 4.0, response.Spec.Compute.Get().GetGpuPortionRequest())
			require.False(t, response.Spec.Compute.Get().GpuMemoryPortionRequest.IsSet())
			require.False(t, response.Spec.Compute.Get().GpuDevicesRequest.IsSet())
			payload, jsonErr := json.Marshal(response.Spec.Compute.Get())
			require.NoError(t, jsonErr)
			var resubmitted api.WorkspaceSpecSpecAllOfCompute
			require.NoError(t, json.Unmarshal(payload, &resubmitted))
			require.NoError(t, api.AssertWorkspaceSpecSpecAllOfComputeConstraints(resubmitted))
			resubmitMap, jsonErr := gojsonq.LoadJson(response.Spec)
			require.NoError(t, jsonErr)
			require.NoError(t, gpucompat.ValidateMemoryRequestInput(resubmitMap))
			require.Nil(t, validateSpecAgainstPolicy(Workspace, rules, values, resubmitMap))
			var update api.InferenceUpdateSpecSpec
			require.Nil(t, copyUpdateSpecIntoFullSpec(context.Background(), spec, &update))
			require.Equal(t, 4.0, spec.Compute.Get().GetGpuMemoryPortionRequest())
		})
	}
}

func TestMemoryRequestBoundsAtWorkloadAndTemplateInputs(t *testing.T) {
	for _, compute := range []string{`{"gpuMemoryPortionRequest":4}`, `{"gpuPortionRequest":4,"gpuMemoryPortionRequest":4}`} {
		t.Run(compute, func(t *testing.T) {
			var input api.WorkspaceSpecSpecAllOfCompute
			require.NoError(t, json.Unmarshal([]byte(compute), &input))
			require.Error(t, api.AssertWorkspaceSpecSpecAllOfComputeConstraints(input))
			var values map[string]any
			require.NoError(t, json.Unmarshal([]byte(`{"compute":`+compute+`}`), &values))
			err := createWorkload(context.Background(), nil, &createWorkloadParams{Spec: values}, Workspace)
			require.NotNil(t, err)
			require.Equal(t, int32(400), err.Code)
			var update api.InferenceUpdateSpecSpec
			require.NoError(t, json.Unmarshal([]byte(`{"compute":`+compute+`}`), &update))
			require.NotNil(t, copyUpdateSpecIntoFullSpec(context.Background(), &workloadsApi.SupersetSpec{}, &update))
			var template api.SupersetTemplate
			require.NoError(t, json.Unmarshal([]byte(`{"spec":{"compute":`+compute+`}}`), &template))
			ctx := workloadstest.SetContext("1", workloadstest.TestUser, []string{})
			_, templateErr := NewUnifiedRequest(ctx, api.WORKLOADTYPE_WORKSPACE, template)
			require.NotNil(t, templateErr)
			require.Equal(t, int32(400), templateErr.Code)
		})
	}
	var template api.SupersetTemplate
	require.NoError(t, json.Unmarshal([]byte(`{"spec":{"compute":{"gpuPortionRequest":4}}}`), &template))
	ctx := workloadstest.SetContext("1", workloadstest.TestUser, []string{})
	request, err := NewUnifiedRequest(ctx, api.WORKLOADTYPE_WORKSPACE, template)
	require.Nil(t, err)
	response, err := legacyCompatibleTemplate(request.template)
	require.Nil(t, err)
	require.Nil(t, response.Spec.Compute.GpuMemoryPortionRequest)
	require.Equal(t, 4.0, *response.Spec.Compute.GpuPortionRequest)
	_, err = NewUnifiedRequest(ctx, api.WORKLOADTYPE_WORKSPACE, response)
	require.Nil(t, err)
}

func TestLegacyWholeRequestPolicyRulesAndFractions(t *testing.T) {
	defaults := map[string]any{"compute": map[string]any{"gpuPortionRequest": json.Number("4")}}
	locked := map[string]any{"compute": map[string]any{"gpuPortionRequest": map[string]any{"canEdit": false}}}
	for _, input := range []map[string]any{
		{"compute": map[string]any{"gpuMemoryPortionRequest": json.Number("0.25")}},
		{"compute": map[string]any{"gpuPortionRequest": json.Number("2")}},
	} {
		mutated, err := MutateWorkload(input, defaults, false, nil, nil)
		require.Nil(t, err)
		require.NotNil(t, validateSpecAgainstPolicy(Workspace, locked, defaults, mutated))
	}
	required := map[string]any{"compute": map[string]any{"gpuPortionRequest": map[string]any{"required": true}}}
	deviceOnly := map[string]any{"compute": map[string]any{"gpuDevicesRequest": 4.0}}
	require.NotNil(t, validateSpecAgainstPolicy(Workspace, required, nil, deviceOnly))
	for _, value := range []float64{4.5, 2.1} {
		var resources corev1.ResourceRequirements
		fc := composer.FieldsComposer{}
		err := fc.ComposeGpuRequestsAndLimits(context.Background(), nil, nil, &value, nil, nil, nil, composer.WholeGpusToK8sResources, &resources, nil, nil)()
		require.NotNil(t, err, "fractional whole GPU request accepted")
	}
}

func TestUnifiedMemoryAliasPolicyResponse(t *testing.T) {
	for _, workloadType := range []api.WorkloadType{api.WORKLOADTYPE_WORKSPACE, api.WORKLOADTYPE_DISTRIBUTED, api.WORKLOADTYPE_DISTRIBUTED_INFERENCE} {
		for _, prefix := range []string{"gpuPortion", "gpuMemoryPortion"} {
			t.Run(string(workloadType)+"/"+prefix, func(t *testing.T) {
				defaults := map[string]any{"compute": map[string]any{prefix + "Request": .25, prefix + "Limit": .5}}
				rules := map[string]any{"compute": map[string]any{prefix + "Request": map[string]any{"canEdit": false}, prefix + "Limit": map[string]any{"max": .5}}}
				paths := []string{"worker"}
				switch workloadType {
				case api.WORKLOADTYPE_DISTRIBUTED:
					defaults = map[string]any{"worker": defaults, "master": defaults}
					rules = map[string]any{"worker": rules, "master": rules}
					paths = []string{"worker", "master"}
				case api.WORKLOADTYPE_DISTRIBUTED_INFERENCE:
					defaults = map[string]any{"leader": defaults, "worker": defaults}
					rules = map[string]any{"leader": rules, "worker": rules}
					paths = []string{"worker.leader", "worker.worker"}
				}
				before, err := json.Marshal([]any{defaults, rules})
				require.NoError(t, err)
				info, apiErr := (UnifiedWorkloadRequest{WorkloadType: workloadType}).composePolicyInfo(defaults, rules)
				require.Nil(t, apiErr)
				response, err := gojsonq.LoadJson(info)
				require.NoError(t, err)
				for _, path := range paths {
					parts := strings.Split(path, ".")
					for _, section := range []string{"defaults", "rules"} {
						fields := response[parts[0]].(map[string]any)[section].(map[string]any)
						if len(parts) > 1 {
							fields = fields[parts[1]].(map[string]any)
						}
						compute := fields["compute"].(map[string]any)
						for _, suffix := range []string{"Request", "Limit"} {
							require.Contains(t, compute, "gpuPortion"+suffix)
							require.Contains(t, compute, "gpuMemoryPortion"+suffix)
							require.Equal(t, compute["gpuPortion"+suffix], compute["gpuMemoryPortion"+suffix])
						}
					}
				}
				after, err := json.Marshal([]any{defaults, rules})
				require.NoError(t, err)
				require.JSONEq(t, string(before), string(after))
			})
		}
	}
}
