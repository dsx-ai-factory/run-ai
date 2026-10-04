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
