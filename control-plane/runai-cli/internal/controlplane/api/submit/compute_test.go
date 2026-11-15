/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/spec/spectype"
	"github.com/spf13/pflag"

	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

func TestGenerateComputeSpecGpuDevicesRequest(t *testing.T) {
	for _, tc := range []struct {
		name    string
		devices *int32
		want    string
	}{
		{name: "omitted", want: `{"cpuCoreRequest":1}`},
		{name: "explicit zero", devices: ptr.To(int32(0)), want: `{"cpuCoreRequest":1,"gpuDevicesRequest":0}`},
		{name: "positive", devices: ptr.To(int32(2)), want: `{"cpuCoreRequest":1,"gpuDevicesRequest":2}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			compute := &Compute{}
			compute.CpuCoreRequest.Set(ptr.To(float64(1)))
			if tc.devices != nil {
				compute.GpuDeviceRequest.Set(tc.devices)
			}

			spec, err := compute.GenerateComputeSpec()
			g.Expect(err).NotTo(HaveOccurred())
			payload, err := json.Marshal(spec)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(payload).To(MatchJSON(tc.want))
		})
	}
}

func TestGpuComputeFlags(t *testing.T) {
	for _, tc := range []struct {
		name, args, want string
	}{
		{"omitted", "", `null`},
		{"new names", "--gpu-memory-portion-request 0.25 --gpu-memory-portion-limit 0.5", `{"gpuMemoryPortionRequest":0.25,"gpuMemoryPortionLimit":0.5}`},
		{"old names", "--gpu-portion-request 0.25 --gpu-portion-limit 0.5", `{"gpuPortionRequest":0.25,"gpuPortionLimit":0.5}`},
		{"legacy whole GPU request", "--gpu-portion-request 4", `{"gpuPortionRequest":4}`},
		{"equal aliases", "--gpu-portion-request 0.25 --gpu-memory-portion-request 0.25", `{"gpuPortionRequest":0.25,"gpuMemoryPortionRequest":0.25}`},
		{"conflicts reach API validation", "--gpu-portion-request 0.25 --gpu-memory-portion-request 0.5", `{"gpuPortionRequest":0.25,"gpuMemoryPortionRequest":0.5}`},
		{"flag order does not hide conflicts", "--gpu-memory-portion-request 0.5 --gpu-portion-request 0.25", `{"gpuPortionRequest":0.25,"gpuMemoryPortionRequest":0.5}`},
		{"explicit zero", "--gpu-memory-portion-request 0", `{"gpuMemoryPortionRequest":0}`},
		{"SM sharing", "--gpu-compute-mode sm-sharing", `{"gpuComputeMode":"sm-sharing"}`},
		{"explicit time slicing", "--gpu-compute-mode time-slicing", `{"gpuComputeMode":"time-slicing"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			compute := &Compute{}
			flags := pflag.NewFlagSet("compute", pflag.ContinueOnError)
			ComputerFlagSet(flags, compute, spectype.Workspace, api.Runai, false)
			g.Expect(flags.Parse(strings.Fields(tc.args))).To(Succeed())
			spec, err := compute.GenerateComputeSpec()
			g.Expect(err).NotTo(HaveOccurred())
			payload, err := json.Marshal(spec)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(payload).To(MatchJSON(tc.want))
			g.Expect(flags.Lookup("gpu-portion-request").Deprecated).To(ContainSubstring("--gpu-memory-portion-request"))
		})
	}
}
