/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package spec

import (
	"context"
	"encoding/json"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("Distributed master GPU requests", func() {
	var (
		gs      *GeneralSpec
		command *cobra.Command
	)

	BeforeEach(func() {
		gs = NewDistributedSpec(api.PyTorch)
		command = &cobra.Command{Use: "pytorch"}
		project := "project"
		gs.SetFlags(command, &project)
	})

	DescribeTable("serializes the requested worker and master compute resources",
		func(args []string, workerCompute, masterCompute string, sameAsWorker bool) {
			Expect(command.ParseFlags(args)).To(Succeed())
			distributedSpec, err := gs.GenerateDistributedSpec(context.Background(), "cluster-id", nil)
			Expect(err).NotTo(HaveOccurred())

			payload, err := json.Marshal(distributedSpec)
			Expect(err).NotTo(HaveOccurred())
			var request struct {
				Spec struct {
					Compute json.RawMessage `json:"compute"`
				} `json:"spec"`
				MasterSpec *struct {
					Compute json.RawMessage `json:"compute"`
				} `json:"masterSpec"`
				MasterSpecSameAsWorker bool `json:"masterSpecSameAsWorker"`
			}
			Expect(json.Unmarshal(payload, &request)).To(Succeed())
			Expect(request.Spec.Compute).To(MatchJSON(workerCompute))
			Expect(request.MasterSpecSameAsWorker).To(Equal(sameAsWorker))
			if masterCompute == "" {
				Expect(request.MasterSpec).To(BeNil())
			} else {
				Expect(request.MasterSpec).NotTo(BeNil())
				Expect(request.MasterSpec.Compute).To(MatchJSON(masterCompute))
			}
		},
		Entry("omitted master GPU request inherits the worker spec",
			[]string{"--gpu-devices-request", "1"},
			`{"gpuDevicesRequest":1}`, "", true),
		Entry("new master names replace inherited old aliases",
			[]string{"--gpu-portion-request", "0.25", "--gpu-portion-limit", "0.5", "--master-gpu-memory-portion-request", "0.5", "--master-gpu-memory-portion-limit", "0.75"},
			`{"gpuPortionRequest":0.25,"gpuPortionLimit":0.5}`, `{"gpuMemoryPortionRequest":0.5,"gpuMemoryPortionLimit":0.75}`, false),
		Entry("old master names replace inherited new aliases",
			[]string{"--gpu-memory-portion-request", "0.25", "--gpu-memory-portion-limit", "0.5", "--master-gpu-portion-request", "0.5", "--master-gpu-portion-limit", "0.75"},
			`{"gpuMemoryPortionRequest":0.25,"gpuMemoryPortionLimit":0.5}`, `{"gpuPortionRequest":0.5,"gpuPortionLimit":0.75}`, false),
		Entry("a mode override alone creates the master spec",
			[]string{"--gpu-memory-portion-request", "0.25", "--gpu-compute-mode", "time-slicing", "--master-gpu-compute-mode", "sm-sharing"},
			`{"gpuMemoryPortionRequest":0.25,"gpuComputeMode":"time-slicing"}`, `{"gpuMemoryPortionRequest":0.25,"gpuComputeMode":"sm-sharing"}`, false),
		Entry("a CPU-only master clears inherited new GPU fields",
			[]string{"--gpu-memory-portion-request", "0.25", "--gpu-memory-portion-limit", "0.5", "--gpu-compute-mode", "sm-sharing", "--master-gpu-devices-request", "0"},
			`{"gpuMemoryPortionRequest":0.25,"gpuMemoryPortionLimit":0.5,"gpuComputeMode":"sm-sharing"}`, `{"gpuDevicesRequest":0}`, false),
		Entry("contradictory explicit master aliases remain available for API validation",
			[]string{"--gpu-memory-portion-request", "0.25", "--master-gpu-portion-request", "0.5", "--master-gpu-memory-portion-request", "0.75"},
			`{"gpuMemoryPortionRequest":0.25}`, `{"gpuPortionRequest":0.5,"gpuMemoryPortionRequest":0.75}`, false),
		Entry("a CPU override preserves the inherited GPU request",
			[]string{"--gpu-devices-request", "1", "--master-cpu-core-request", "1"},
			`{"gpuDevicesRequest":1}`, `{"gpuDevicesRequest":1,"cpuCoreRequest":1}`, false),
		Entry("an explicit zero creates a CPU-only master",
			[]string{"--gpu-devices-request", "1", "--master-gpu-devices-request", "0"},
			`{"gpuDevicesRequest":1}`, `{"gpuDevicesRequest":0}`, false),
		Entry("an explicit master zero survives an omitted worker GPU request",
			[]string{"--cpu-core-request", "1", "--master-gpu-devices-request", "0"},
			`{"cpuCoreRequest":1}`, `{"cpuCoreRequest":1,"gpuDevicesRequest":0}`, false),
		Entry("an explicit zero survives another master override",
			[]string{"--gpu-devices-request", "1", "--master-gpu-devices-request", "0", "--master-cpu-core-request", "1"},
			`{"gpuDevicesRequest":1}`, `{"gpuDevicesRequest":0,"cpuCoreRequest":1}`, false),
		Entry("a positive master GPU request overrides the worker request",
			[]string{"--gpu-devices-request", "1", "--master-gpu-devices-request", "2"},
			`{"gpuDevicesRequest":1}`, `{"gpuDevicesRequest":2}`, false),
		Entry("an explicit master device count survives a new memory fraction",
			[]string{"--master-gpu-devices-request", "2", "--master-gpu-memory-portion-request", "0.5"},
			`null`, `{"gpuDevicesRequest":2,"gpuMemoryPortionRequest":0.5}`, false),
		Entry("an explicit master device count survives a legacy fraction",
			[]string{"--master-gpu-devices-request", "2", "--master-gpu-portion-request", "0.5"},
			`null`, `{"gpuDevicesRequest":2,"gpuPortionRequest":0.5}`, false),
		Entry("a full new memory fraction preserves the explicit device count",
			[]string{"--master-gpu-memory-portion-request", "1", "--master-gpu-devices-request", "2"},
			`null`, `{"gpuDevicesRequest":2,"gpuMemoryPortionRequest":1}`, false),
		Entry("a full legacy fraction preserves the explicit device count",
			[]string{"--master-gpu-portion-request", "1", "--master-gpu-devices-request", "2"},
			`null`, `{"gpuDevicesRequest":2,"gpuPortionRequest":1}`, false),
		Entry("a legacy whole-GPU portion keeps its existing precedence",
			[]string{"--master-gpu-devices-request", "2", "--master-gpu-portion-request", "4"},
			`null`, `{"gpuPortionRequest":4}`, false),
		Entry("explicit master devices and new fractions replace inherited worker settings",
			[]string{
				"--gpu-devices-request", "4", "--gpu-portion-request", "0.25", "--gpu-portion-limit", "0.5", "--gpu-compute-mode", "time-slicing",
				"--master-gpu-devices-request", "2", "--master-gpu-memory-portion-request", "0.5", "--master-gpu-memory-portion-limit", "0.75", "--master-gpu-compute-mode", "sm-sharing",
			},
			`{"gpuDevicesRequest":4,"gpuPortionRequest":0.25,"gpuPortionLimit":0.5,"gpuComputeMode":"time-slicing"}`,
			`{"gpuDevicesRequest":2,"gpuMemoryPortionRequest":0.5,"gpuMemoryPortionLimit":0.75,"gpuComputeMode":"sm-sharing"}`, false),
		Entry("a new master fraction alone does not inherit the worker device count",
			[]string{"--gpu-devices-request", "4", "--master-gpu-memory-portion-request", "0.5"},
			`{"gpuDevicesRequest":4}`, `{"gpuMemoryPortionRequest":0.5}`, false),
		Entry("an explicit zero new fraction remains available for API validation",
			[]string{"--master-gpu-devices-request", "2", "--master-gpu-memory-portion-request", "0"},
			`null`, `{"gpuDevicesRequest":2,"gpuMemoryPortionRequest":0}`, false),
		Entry("an explicit new master fraction retains precedence over a zero device request",
			[]string{"--master-gpu-devices-request", "0", "--master-gpu-memory-portion-request", "0.5"},
			`null`, `{"gpuMemoryPortionRequest":0.5}`, false),
		Entry("no master suppresses an explicit master GPU request",
			[]string{"--gpu-devices-request", "1", "--master-gpu-devices-request", "0", "--no-master"},
			`{"gpuDevicesRequest":1}`, "", false),
		Entry("a CPU-only master does not inherit worker GPU fractions or limits",
			[]string{"--gpu-devices-request", "1", "--gpu-request-type", "portion", "--gpu-portion-request", "0.5", "--gpu-portion-limit", "0.75", "--master-gpu-devices-request", "0"},
			`{"gpuDevicesRequest":1,"gpuRequestType":"portion","gpuPortionRequest":0.5,"gpuPortionLimit":0.75}`, `{"gpuDevicesRequest":0}`, false),
		Entry("a CPU-only master does not inherit worker GPU memory or limits",
			[]string{"--gpu-devices-request", "1", "--gpu-request-type", "memory", "--gpu-memory-request", "1G", "--gpu-memory-limit", "2G", "--master-gpu-devices-request", "0"},
			`{"gpuDevicesRequest":1,"gpuRequestType":"memory","gpuMemoryRequest":"1G","gpuMemoryLimit":"2G"}`, `{"gpuDevicesRequest":0}`, false),
		Entry("a fractional master override does not serialize a zero device count",
			[]string{"--gpu-devices-request", "1", "--master-gpu-portion-request", "0.5", "--master-gpu-portion-limit", "0.75"},
			`{"gpuDevicesRequest":1}`, `{"gpuPortionRequest":0.5,"gpuPortionLimit":0.75}`, false),
		Entry("an explicit positive master portion retains precedence over a zero device request",
			[]string{"--gpu-devices-request", "1", "--master-gpu-devices-request", "0", "--master-gpu-portion-request", "0.5", "--master-gpu-portion-limit", "0.75"},
			`{"gpuDevicesRequest":1}`, `{"gpuPortionRequest":0.5,"gpuPortionLimit":0.75}`, false),
	)

	DescribeTable("ignores a null master GPU device request",
		func(masterArgs []string, masterCompute string) {
			args := append([]string{"--gpu-devices-request", "1"}, masterArgs...)
			Expect(command.ParseFlags(args)).To(Succeed())
			gs.flags.Compute.MasterGpuDeviceRequest.Set(nil)

			distributedSpec, err := gs.GenerateDistributedSpec(context.Background(), "cluster-id", nil)
			Expect(err).NotTo(HaveOccurred())
			worker := distributedSpec.GetSpec()
			payload, err := json.Marshal(worker.GetCompute())
			Expect(err).NotTo(HaveOccurred())
			Expect(payload).To(MatchJSON(`{"gpuDevicesRequest":1}`))

			if masterCompute == "" {
				Expect(distributedSpec.GetMasterSpecSameAsWorker()).To(BeTrue())
				Expect(distributedSpec.HasMasterSpec()).To(BeFalse())
			} else {
				Expect(distributedSpec.GetMasterSpecSameAsWorker()).To(BeFalse())
				Expect(distributedSpec.HasMasterSpec()).To(BeTrue())
				master := distributedSpec.GetMasterSpec()
				payload, err = json.Marshal(master.GetCompute())
				Expect(err).NotTo(HaveOccurred())
				Expect(payload).To(MatchJSON(masterCompute))
			}
		},
		Entry("without another master override", []string{}, ""),
		Entry("with a master CPU override", []string{"--master-cpu-core-request", "1"}, `{"gpuDevicesRequest":1,"cpuCoreRequest":1}`),
		Entry("with a master fraction", []string{"--master-gpu-memory-portion-request", "0.5"}, `{"gpuMemoryPortionRequest":0.5}`),
	)

	DescribeTable("preserves the CPU-only MPI master with a separate command",
		func(masterArgs []string) {
			gs = NewDistributedSpec(api.MPI)
			command = &cobra.Command{Use: "mpi"}
			project := "project"
			gs.SetFlags(command, &project)
			args := append([]string{"--gpu-devices-request", "1", "--master-command", "sleep"}, masterArgs...)
			Expect(command.ParseFlags(args)).To(Succeed())
			distributedSpec, err := gs.GenerateDistributedSpec(context.Background(), "cluster-id", nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(distributedSpec.GetMasterSpecSameAsWorker()).To(BeFalse())
			Expect(distributedSpec.HasMasterSpec()).To(BeTrue())

			master := distributedSpec.GetMasterSpec()
			payload, err := json.Marshal(master.GetCompute())
			Expect(err).NotTo(HaveOccurred())
			Expect(payload).To(MatchJSON(`{}`))
			worker := distributedSpec.GetSpec()
			payload, err = json.Marshal(worker.GetCompute())
			Expect(err).NotTo(HaveOccurred())
			Expect(payload).To(MatchJSON(`{"gpuDevicesRequest":1}`))
		},
		Entry("when master devices are omitted", []string{}),
		Entry("when master devices are explicitly zero", []string{"--master-gpu-devices-request", "0"}),
	)

	It("rejects a negative master GPU device request", func() {
		Expect(command.ParseFlags([]string{"--master-gpu-devices-request", "-1"})).To(HaveOccurred())
	})
})
