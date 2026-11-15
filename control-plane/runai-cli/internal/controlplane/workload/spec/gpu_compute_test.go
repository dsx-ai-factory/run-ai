/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package spec

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("GPU flags across submission and update flows", func() {
	DescribeTable("serializes compute values without introducing a mode default",
		func(create func() *GeneralSpec, update bool, generate func(*GeneralSpec) (any, error)) {
			gs := create()
			command := &cobra.Command{Use: "workload"}
			project := "project"
			if update {
				gs.SetUpdateFlags(command)
			} else {
				gs.SetFlags(command, &project)
			}
			Expect(command.ParseFlags([]string{"--gpu-memory-portion-request", "0.25", "--gpu-memory-portion-limit", "0.5"})).To(Succeed())
			assertCompute := func(expected string) {
				value, err := generate(gs)
				Expect(err).NotTo(HaveOccurred())
				data, err := json.Marshal(value)
				Expect(err).NotTo(HaveOccurred())
				var spec struct {
					Compute json.RawMessage `json:"compute"`
				}
				Expect(json.Unmarshal(data, &spec)).To(Succeed())
				Expect(spec.Compute).To(MatchJSON(expected))
			}
			assertCompute(`{"gpuMemoryPortionRequest":0.25,"gpuMemoryPortionLimit":0.5}`)
			Expect(command.ParseFlags([]string{"--gpu-compute-mode", "sm-sharing"})).To(Succeed())
			assertCompute(`{"gpuMemoryPortionRequest":0.25,"gpuMemoryPortionLimit":0.5,"gpuComputeMode":"sm-sharing"}`)
			Expect(command.ParseFlags([]string{"--gpu-compute-mode", "invalid"})).To(Succeed())
			_, err := generate(gs)
			Expect(err).To(MatchError(ContainSubstring("invalid value 'invalid' for GpuComputeMode")))
			gs.Clear()
			cpu := float64(1)
			gs.flags.Compute.CpuCoreRequest.Set(&cpu)
			assertCompute(`{"cpuCoreRequest":1}`)
		},
		Entry("workspace", NewWorkspaceSpec, false, func(gs *GeneralSpec) (any, error) {
			return gs.GenerateWorkspaceSpec(context.Background(), "cluster", nil)
		}),
		Entry("training", NewTrainingSpec, false, func(gs *GeneralSpec) (any, error) {
			return gs.GenerateTrainingSpec(context.Background(), "cluster", nil)
		}),
		Entry("inference submit", NewInferenceSpec, false, func(gs *GeneralSpec) (any, error) {
			return gs.GenerateInferenceSpec(context.Background(), "cluster", nil)
		}),
		Entry("inference update", NewInferenceSpec, true, func(gs *GeneralSpec) (any, error) {
			return gs.GenerateInferenceUpdateSpec(context.Background(), "cluster", nil)
		}),
		Entry("NIM submit", NewNimSpec, false, func(gs *GeneralSpec) (any, error) {
			return gs.GenerateNimServiceSpec(context.Background(), "cluster")
		}),
		Entry("NIM update", NewNimSpec, true, func(gs *GeneralSpec) (any, error) {
			return gs.GenerateNimServiceSpec(context.Background(), "cluster")
		}),
		Entry("legacy submit", func() *GeneralSpec { return &NewSubmitSpec().GeneralSpec }, false, func(gs *GeneralSpec) (any, error) {
			return gs.GenerateWorkspaceSpec(context.Background(), "cluster", nil)
		}),
	)
})

var _ = Describe("Distributed GPU memory aliases and modes", func() {
	DescribeTable("applies worker overrides without conflicting inherited aliases",
		func(args, leader, worker string) {
			gs := NewDistributedSpec(api.LWS)
			command := &cobra.Command{Use: "distributed"}
			project := "project"
			gs.SetFlags(command, &project)
			Expect(command.ParseFlags(strings.Fields(args))).To(Succeed())
			value, err := gs.GenerateDistributedInferenceSpec(context.Background(), "cluster", nil)
			Expect(err).NotTo(HaveOccurred())
			leaderCompute, err := json.Marshal(value.GetLeader().Compute)
			Expect(err).NotTo(HaveOccurred())
			workerCompute, err := json.Marshal(value.GetWorker().Compute)
			Expect(err).NotTo(HaveOccurred())
			Expect(leaderCompute).To(MatchJSON(leader))
			Expect(workerCompute).To(MatchJSON(worker))
			second, err := gs.GenerateDistributedInferenceSpec(context.Background(), "cluster", nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(second).To(Equal(value))
		},
		Entry("new worker names replace old aliases",
			"--gpu-portion-request 0.25 --gpu-portion-limit 0.5 --worker-gpu-memory-portion-request 0.5 --worker-gpu-memory-portion-limit 0.75",
			`{"gpuPortionRequest":0.25,"gpuPortionLimit":0.5}`, `{"gpuMemoryPortionRequest":0.5,"gpuMemoryPortionLimit":0.75}`),
		Entry("old worker names replace new aliases",
			"--gpu-memory-portion-request 0.25 --gpu-memory-portion-limit 0.5 --worker-gpu-portion-request 0.5 --worker-gpu-portion-limit 0.75",
			`{"gpuMemoryPortionRequest":0.25,"gpuMemoryPortionLimit":0.5}`, `{"gpuPortionRequest":0.5,"gpuPortionLimit":0.75}`),
		Entry("a worker mode alone creates a separate spec",
			"--gpu-memory-portion-request 0.25 --gpu-compute-mode time-slicing --worker-gpu-compute-mode sm-sharing",
			`{"gpuMemoryPortionRequest":0.25,"gpuComputeMode":"time-slicing"}`, `{"gpuMemoryPortionRequest":0.25,"gpuComputeMode":"sm-sharing"}`),
		Entry("worker-as-leader inherits new fields",
			"--workers 1 --worker-as-leader --gpu-memory-portion-request 0.25 --gpu-compute-mode sm-sharing",
			`{"gpuMemoryPortionRequest":0.25,"gpuComputeMode":"sm-sharing"}`, `{"gpuMemoryPortionRequest":0.25,"gpuComputeMode":"sm-sharing"}`),
	)

	It("keeps the MPI launcher CPU-only and rejects a requested master mode", func() {
		gs := NewDistributedSpec(api.MPI)
		command := &cobra.Command{Use: "mpi"}
		project := "project"
		gs.SetFlags(command, &project)
		Expect(command.ParseFlags(strings.Fields("--gpu-memory-portion-request 0.25 --gpu-memory-portion-limit 0.5 --gpu-compute-mode sm-sharing"))).To(Succeed())
		value, err := gs.GenerateDistributedSpec(context.Background(), "cluster", nil)
		Expect(err).NotTo(HaveOccurred())
		master, err := json.Marshal(value.GetMasterSpec().Compute)
		Expect(err).NotTo(HaveOccurred())
		Expect(master).To(MatchJSON(`null`))
		Expect(command.ParseFlags([]string{"--master-gpu-compute-mode", "sm-sharing"})).To(Succeed())
		Expect(gs.CheckDistributeMpiFlags()).To(HaveOccurred())
	})
})
