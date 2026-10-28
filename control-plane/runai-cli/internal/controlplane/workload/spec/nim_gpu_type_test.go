// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"context"
	"encoding/json"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("NIM GPU type submission", func() {
	DescribeTable("preserves allocation omission and explicit allocation",
		func(args []string, expectedCompute string) {
			gs := NewNimSpec()
			command := &cobra.Command{Use: "nim"}
			project := "project"
			gs.SetFlags(command, &project)
			Expect(command.ParseFlags(args)).To(Succeed())
			spec, err := gs.GenerateNimServiceSpec(context.Background(), "cluster")
			Expect(err).NotTo(HaveOccurred())
			data, err := json.Marshal(spec)
			Expect(err).NotTo(HaveOccurred())
			var wire map[string]json.RawMessage
			Expect(json.Unmarshal(data, &wire)).To(Succeed())
			Expect(wire["gpuType"]).To(MatchJSON(`"NVIDIA-H100-80GB-HBM3"`))
			if expectedCompute == "" {
				Expect(wire).NotTo(HaveKey("compute"))
			} else {
				Expect(wire["compute"]).To(MatchJSON(expectedCompute))
			}
		},
		Entry("automatic sizing", []string{"--gpu-type", "NVIDIA-H100-80GB-HBM3"}, ""),
		Entry("explicit count", []string{"--gpu-type", "NVIDIA-H100-80GB-HBM3", "--gpu-devices-request", "2"}, `{"gpuDevicesRequest":2}`),
		Entry("count and fraction", []string{"--gpu-type", "NVIDIA-H100-80GB-HBM3", "-g", "2", "--gpu-memory-portion-request", "0.5"}, `{"gpuDevicesRequest":2,"gpuMemoryPortionRequest":0.5}`),
		Entry("count and memory", []string{"--gpu-type", "NVIDIA-H100-80GB-HBM3", "-g", "2", "--gpu-memory-request", "10G"}, `{"gpuDevicesRequest":2,"gpuMemoryRequest":"10G"}`),
		Entry("explicit memory", []string{"--gpu-type", "NVIDIA-H100-80GB-HBM3", "--gpu-memory-request", "27G"}, `{"gpuMemoryRequest":"27G"}`),
	)
})

var _ = Describe("NIM GPU inventory lookup", func() {
	It("prefers an exact label even when another differs only in case", func() {
		gs := NewNimSpec()
		gs.flags.Nim.GpuType = "NVIDIA-A100"
		Expect(gs.ResolveNimGPUType(func() ([]string, error) { return []string{"nvidia-a100", "NVIDIA-A100"}, nil })).To(Succeed())
		Expect(gs.flags.Nim.GpuType).To(Equal("NVIDIA-A100"))
	})
	It("propagates inventory errors without changing the supplied label", func() {
		gs := NewNimSpec()
		gs.flags.Nim.GpuType = "nvidia-a100"
		failure := errors.New("inventory unavailable")
		err := gs.ResolveNimGPUType(func() ([]string, error) { return nil, failure })
		Expect(errors.Is(err, failure)).To(BeTrue())
		Expect(gs.flags.Nim.GpuType).To(Equal("nvidia-a100"))
	})

	It("resolves mixed case to the exact inventory label without changing allocation", func() {
		gs := NewNimSpec()
		cmd := &cobra.Command{}
		project := "p"
		gs.SetFlags(cmd, &project)
		Expect(cmd.ParseFlags([]string{"--gpu-type", "nViDiA-a100-sxm4-40gb", "-g", "2"})).To(Succeed())
		Expect(gs.ResolveNimGPUType(func() ([]string, error) { return []string{"NVIDIA-A100-SXM4-40GB"}, nil })).To(Succeed())
		result, err := gs.GenerateNimServiceSpec(context.Background(), "cluster")
		Expect(err).NotTo(HaveOccurred())
		Expect(result.GetGpuType()).To(Equal("NVIDIA-A100-SXM4-40GB"))
		compute := result.GetCompute()
		Expect(compute.GetGpuDevicesRequest()).To(Equal(int32(2)))
	})
	It("does not fetch inventory when the flag is omitted", func() {
		gs := NewNimSpec()
		Expect(gs.ResolveNimGPUType(func() ([]string, error) { Fail("unexpected inventory lookup"); return nil, nil })).To(Succeed())
	})
	It("does not invent aliases for absent hardware", func() {
		gs := NewNimSpec()
		gs.flags.Nim.GpuType = "absent"
		Expect(gs.ResolveNimGPUType(func() ([]string, error) { return []string{"NVIDIA-A100"}, nil })).To(Succeed())
		Expect(gs.flags.Nim.GpuType).To(Equal("absent"))
	})
	It("rejects an ambiguous case-insensitive match", func() {
		gs := NewNimSpec()
		gs.flags.Nim.GpuType = "NVIDIA-a100"
		Expect(gs.ResolveNimGPUType(func() ([]string, error) { return []string{"NVIDIA-A100", "nvidia-a100"}, nil })).NotTo(Succeed())
	})
})
