// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package spec

import (
	"context"
	"encoding/json"

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
		Entry("explicit memory", []string{"--gpu-type", "NVIDIA-H100-80GB-HBM3", "--gpu-memory-request", "27G"}, `{"gpuMemoryRequest":"27G"}`),
	)
})
