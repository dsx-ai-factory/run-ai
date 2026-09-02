/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package api_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api"
)

var _ = Describe("WorkloadType", func() {
	Describe("CLIType method", func() {
		DescribeTable("returns correct CLI type",
			func(wt api.WorkloadType, expected string) {
				Expect(wt.CLIType()).To(Equal(expected))
			},
			Entry("Workspace", api.Workspace, "workspace"),
			Entry("Inference", api.Inference, "inference"),
			Entry("Training", api.Training, "training"),
			Entry("Distributed", api.Distributed, "training"),
			Entry("Undefined", api.Undefined, "unknown"),
		)
	})

	Describe("String method", func() {
		DescribeTable("returns correct string representation",
			func(wt api.WorkloadType, expected string) {
				Expect(wt.String()).To(Equal(expected))
			},
			Entry("Workspace", api.Workspace, "workspace"),
			Entry("Inference", api.Inference, "inference"),
			Entry("Training", api.Training, "training"),
			Entry("Distributed", api.Distributed, "distributed"),
			Entry("Undefined", api.Undefined, "unknown"),
		)
	})

	Describe("GetWorkloadTypes function", func() {
		It("returns all workload types", func() {
			types := api.GetWorkloadTypes()
			Expect(types).To(ConsistOf("workspace", "training", "inference"))
		})
	})

	Describe("QueryParam method", func() {
		DescribeTable("returns correct query parameter",
			func(wt api.WorkloadType, expected string) {
				Expect(wt.QueryParam()).To(Equal(expected))
			},
			Entry("Workspace", api.Workspace, "Workspace"),
			Entry("Inference", api.Inference, "Inference"),
			Entry("Training", api.Training, "Training||Distributed"),
			Entry("Distributed", api.Distributed, "Distributed"),
			Entry("Undefined", api.Undefined, "Unknown"),
		)
	})

	Describe("ParseWorkloadType function", func() {
		DescribeTable("parses workload type correctly",
			func(input string, expectedType api.WorkloadType, expectError bool) {
				wt, err := api.ParseWorkloadType(input)
				if expectError {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("unknown workload type"))
					Expect(wt).To(Equal(api.WorkloadType(-1)))
				} else {
					Expect(err).NotTo(HaveOccurred())
					Expect(wt).To(Equal(expectedType))
				}
			},
			Entry("workspace", "workspace", api.Workspace, false),
			Entry("inference", "inference", api.Inference, false),
			Entry("inference nim", "inference nim", api.Inference, false),
			Entry("training", "training", api.Training, false),
			Entry("unknown", "unknown", api.WorkloadType(-1), true),
			Entry("empty string", "", api.WorkloadType(-1), true),
		)
	})
})
