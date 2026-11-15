/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gpucompat

import (
	"encoding/json"
	"testing"

	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
)

func TestNullableResponseMemoryAliases(t *testing.T) {
	if LegacyCompatibleWorkloadSpec(&workloadsApi.SupersetSpec{}).Compute.IsSet() {
		t.Fatal("fabricated compute section")
	}
	for _, key := range []string{"gpuPortionRequest", "gpuMemoryPortionRequest", "gpuPortionLimit", "gpuMemoryPortionLimit"} {
		t.Run(key, func(t *testing.T) {
			var spec workloadsApi.SupersetSpec
			if err := json.Unmarshal([]byte(`{"compute":{"`+key+`":null}}`), &spec); err != nil {
				t.Fatal(err)
			}
			response := LegacyCompatibleWorkloadSpec(&spec).GetCompute()
			old, current := response.GpuPortionRequest, response.GpuMemoryPortionRequest
			if key == "gpuPortionLimit" || key == "gpuMemoryPortionLimit" {
				old, current = response.GpuPortionLimit, response.GpuMemoryPortionLimit
			}
			if !old.IsSet() || !current.IsSet() || old.Get() != nil || current.Get() != nil {
				t.Fatal("null alias presence diverged")
			}
		})
	}
}

func TestWholeGPUResponseSections(t *testing.T) {
	var spec workloadsApi.SupersetSpec
	if err := json.Unmarshal([]byte(`{"compute":{"gpuPortionRequest":4},"leader":{"compute":{"gpuPortionRequest":4}},"worker":{"compute":{"gpuPortionRequest":4}}}`), &spec); err != nil {
		t.Fatal(err)
	}
	response := LegacyCompatibleWorkloadSpec(&spec)
	master := workloadsApi.MasterSpec{Compute: spec.Compute}
	masterResponse := LegacyCompatibleMasterSpec(&master)
	for name, section := range map[string]*workloadsApi.WorkspaceSpecSpecAllOfCompute{
		"compute": response.Compute.Get(), "leader": response.Leader.Get().Compute.Get(), "worker": response.Worker.Get().Compute.Get(), "master": masterResponse.Compute.Get(),
	} {
		t.Run(name, func(t *testing.T) {
			if section.GpuMemoryPortionRequest.IsSet() || section.GetGpuPortionRequest() != 4 || section.GpuDevicesRequest.IsSet() {
				t.Fatalf("invalid whole GPU response: %#v", section)
			}
		})
	}
	*response.Compute.Get().GpuPortionRequest.Get() = 2
	if spec.Compute.Get().GetGpuPortionRequest() != 4 {
		t.Fatal("response mutated stored value")
	}
}
