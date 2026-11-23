/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package spec

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	"github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

func TestAcceleratorFlags(t *testing.T) {
	for _, tc := range []struct {
		name               string
		args               []string
		accelerator, model string
	}{
		{name: "legacy"},
		{name: "NVIDIA model", args: []string{"--gpu-type", "H100"}, model: "H100"},
		{name: "any model on custom resource", args: []string{"--gpu-type", "any", "--accelerator", "vendor.example.com/fpga"}, accelerator: `{"resourceName":"vendor.example.com/fpga"}`, model: "any"},
		{name: "alternate vendor", args: []string{"--accelerator", "amd.com/gpu"}, accelerator: `{"resourceName":"amd.com/gpu"}`},
		{name: "explicit device label", args: []string{"--accelerator", "amd.com/gpu", "--device-label-key", "lab.example/model", "--gpu-type", "MI300X"}, accelerator: `{"resourceName":"amd.com/gpu","deviceLabelKey":"lab.example/model"}`, model: "MI300X"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			spec := NewTrainingSpec()
			cmd := &cobra.Command{Use: "training"}
			project := "project"
			spec.SetFlags(cmd, &project)
			g.Expect(cmd.ParseFlags(tc.args)).To(gomega.Succeed())
			result, err := spec.GenerateTrainingSpec(context.Background(), "cluster", nil)
			g.Expect(err).NotTo(gomega.HaveOccurred())
			g.Expect(result.GetGpuType()).To(gomega.Equal(tc.model))
			g.Expect(result.HasGpuType()).To(gomega.Equal(tc.model != ""))
			selection, present := result.GetAcceleratorOk()
			g.Expect(present).To(gomega.Equal(tc.accelerator != ""))
			if present {
				encoded, err := json.Marshal(selection)
				g.Expect(err).NotTo(gomega.HaveOccurred())
				g.Expect(encoded).To(gomega.MatchJSON(tc.accelerator))
			}
		})
	}
}

func TestRejectEmptyAcceleratorFlags(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
	}{
		{"empty resource with a model", []string{"--gpu-type", "H100", "--accelerator="}},
		{"empty model", []string{"--gpu-type="}},
		{"empty key", []string{"--device-label-key="}},
		{"blank resource", []string{"--accelerator", " "}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			cmd := &cobra.Command{Use: "submit"}
			project := "project"
			NewTrainingSpec().SetFlags(cmd, &project)
			g.Expect(cmd.ParseFlags(tc.args)).To(gomega.MatchError(gomega.ContainSubstring("must not be empty")))
		})
	}
}

func TestAcceleratorCPUOnlyTrainingMaster(t *testing.T) {
	for _, tc := range []struct {
		name        string
		framework   api.Framework
		masterFlags []string
	}{
		{"MPI launcher", api.MPI, []string{"--master-cpu-core-request", "1"}},
		{"CPU-only PyTorch master", api.PyTorch, []string{"--master-gpu-devices-request", "0"}},
		{"zero memory portion", api.PyTorch, []string{"--master-gpu-memory-portion-request", "0"}},
		{"another extended resource", api.PyTorch, []string{"--master-gpu-devices-request", "0", "--master-extended-resource", "example.com/network=1"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			spec := NewDistributedSpec(tc.framework)
			cmd := &cobra.Command{Use: "submit"}
			project := "project"
			spec.SetFlags(cmd, &project)
			flags := []string{"--gpu-devices-request", "1", "--gpu-type", "MI300X", "--accelerator", "amd.com/gpu"}
			g.Expect(cmd.ParseFlags(append(flags, tc.masterFlags...))).To(gomega.Succeed())
			job, err := spec.GenerateDistributedSpec(context.Background(), "cluster", nil)
			g.Expect(err).NotTo(gomega.HaveOccurred())
			g.Expect(job.Spec.GetGpuType()).To(gomega.Equal("MI300X"))
			g.Expect(job.GetMasterSpecSameAsWorker()).To(gomega.BeFalse())
			master := job.MasterSpec.Get()
			g.Expect(master).NotTo(gomega.BeNil())
			g.Expect(master.HasGpuType()).To(gomega.BeFalse())
			g.Expect(master.HasAccelerator()).To(gomega.BeFalse())
		})
	}
}

func TestMasterGPUSelectionUsesEffectiveCompute(t *testing.T) {
	for _, tc := range []struct {
		name, flag, value, computeJSON string
	}{
		{"memory portion", "--master-gpu-memory-portion-request", "0.5", `{"gpuMemoryPortionRequest":0.5}`},
		{"legacy portion", "--master-gpu-portion-request", "0.5", `{"gpuPortionRequest":0.5}`},
		{"whole GPUs", "--master-gpu-devices-request", "2", `{"gpuDevicesRequest":2}`},
		{"legacy whole GPUs", "--master-gpu-portion-request", "2", `{"gpuPortionRequest":2}`},
		{"generic resource", "--master-extended-resource", "nvidia.com/gpu=1", `{"gpuDevicesRequest":0,"extendedResources":[{"resource":"nvidia.com/gpu","quantity":"1"}]}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			s := NewDistributedSpec(api.PyTorch)
			cmd := &cobra.Command{Use: "submit"}
			project := "project"
			s.SetFlags(cmd, &project)
			g.Expect(cmd.ParseFlags([]string{"--gpu-type", "H100", "--accelerator", "nvidia.com/gpu", "--gpu-devices-request", "1",
				"--master-gpu-devices-request", "0", tc.flag, tc.value})).To(gomega.Succeed())
			workload, err := s.GenerateDistributedSpec(context.Background(), "cluster", nil)
			g.Expect(err).NotTo(gomega.HaveOccurred())
			master := workload.MasterSpec.Get()
			g.Expect(master).NotTo(gomega.BeNil())
			raw, err := json.Marshal(master.Compute.Get())
			g.Expect(err).NotTo(gomega.HaveOccurred())
			g.Expect(raw).To(gomega.MatchJSON(tc.computeJSON))
			g.Expect(master.GetGpuType()).To(gomega.Equal("H100"))
			g.Expect(master.Accelerator.Get()).NotTo(gomega.BeNil())
			g.Expect(master.Accelerator.Get().GetResourceName()).To(gomega.Equal("nvidia.com/gpu"))
		})
	}
}

func TestAcceleratorInferenceUpdate(t *testing.T) {
	g := gomega.NewWithT(t)
	spec := NewInferenceSpec()
	cmd := &cobra.Command{Use: "update"}
	spec.SetUpdateFlags(cmd)
	g.Expect(cmd.ParseFlags([]string{"--gpu-type", "any"})).To(gomega.Succeed())
	updated, err := spec.GenerateInferenceUpdateSpec(context.Background(), "cluster", nil)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(updated.GetGpuType()).To(gomega.Equal("any"))
	g.Expect(updated.HasAccelerator()).To(gomega.BeFalse())
}
