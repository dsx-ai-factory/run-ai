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

var _ = Describe("Distributed specification reuse", func() {
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

	DescribeTable("preserves worker resources across repeated generation",
		func(args []string, workerCompute string) {
			Expect(command.ParseFlags(args)).To(Succeed())

			first, err := gs.GenerateDistributedSpec(context.Background(), "cluster-id", nil)
			Expect(err).NotTo(HaveOccurred())
			firstPayload, err := json.Marshal(first)
			Expect(err).NotTo(HaveOccurred())
			computePayload, err := json.Marshal(first.GetSpec().Compute)
			Expect(err).NotTo(HaveOccurred())
			Expect(computePayload).To(MatchJSON(workerCompute))

			second, err := gs.GenerateDistributedSpec(context.Background(), "cluster-id", nil)
			Expect(err).NotTo(HaveOccurred())
			secondPayload, err := json.Marshal(second)
			Expect(err).NotTo(HaveOccurred())
			Expect(secondPayload).To(MatchJSON(firstPayload))
		},
		Entry("when the master requests zero GPUs",
			[]string{"--gpu-devices-request", "1", "--master-gpu-devices-request", "0"},
			`{"gpuDevicesRequest":1}`),
		Entry("when the master overrides both devices and the fraction",
			[]string{"--gpu-devices-request", "4", "--gpu-memory-portion-request", "0.25", "--master-gpu-devices-request", "2", "--master-gpu-memory-portion-request", "0.5"},
			`{"gpuDevicesRequest":4,"gpuMemoryPortionRequest":0.25}`),
		Entry("when worker compute is omitted and the master requests GPUs",
			[]string{"--master-gpu-devices-request", "2"},
			`null`),
	)

	DescribeTable("restores flags after master generation fails",
		func(setInvalidMasterFlag func(*GeneralSpec), expectedError string) {
			Expect(command.ParseFlags([]string{"--gpu-devices-request", "1", "--master-gpu-devices-request", "0"})).To(Succeed())
			setInvalidMasterFlag(gs)
			originalFlags, err := json.Marshal(gs.flags)
			Expect(err).NotTo(HaveOccurred())

			_, err = gs.GenerateDistributedSpec(context.Background(), "cluster-id", nil)
			Expect(err).To(MatchError(expectedError))
			restoredFlags, err := json.Marshal(gs.flags)
			Expect(err).NotTo(HaveOccurred())
			Expect(restoredFlags).To(MatchJSON(originalFlags))

			gs.flags.Container.MasterEnvironmentVariables = nil
			gs.flags.Compute.MasterExtendedResources = nil
			retried, err := gs.GenerateDistributedSpec(context.Background(), "cluster-id", nil)
			Expect(err).NotTo(HaveOccurred())
			computePayload, err := json.Marshal(retried.GetSpec().Compute)
			Expect(err).NotTo(HaveOccurred())
			Expect(computePayload).To(MatchJSON(`{"gpuDevicesRequest":1}`))
		},
		Entry("while applying master environment overrides",
			func(gs *GeneralSpec) { gs.flags.Container.MasterEnvironmentVariables = []string{"MISSING=-"} },
			"couldn't find environment to remove, env: MISSING"),
		Entry("while building master compute resources",
			func(gs *GeneralSpec) { gs.flags.Compute.MasterExtendedResources = []string{"invalid"} },
			"value (invalid) does not contain `=` sign"),
	)
})
