/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer_v2

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/common/cluster-capabilities/checker"
)

var _ = Describe("cluster-based composer options", func() {
	DescribeTable("gates ServiceSpec labels by cluster version",
		func(version string, expected bool) {
			capabilities, err := checker.NewClusterCapabilitiesCheckerForVersion(version)
			Expect(err).NotTo(HaveOccurred())
			options := composer.FieldsComposerOptions{}

			Expect(SetClusterDependentOptions(context.Background(), &options, capabilities)).To(Succeed())
			Expect(options.WorkloadV2ServiceSpecLabelsSupported).To(Equal(expected))
		},
		Entry("unsupported before 3.0", "2.99.99", false),
		Entry("supported at 3.0", "3.0.0", true),
	)

	DescribeTable("gates GPU memory-limit annotations by cluster version",
		func(version string, expected bool) {
			capabilities, err := checker.NewClusterCapabilitiesCheckerForVersion(version)
			Expect(err).NotTo(HaveOccurred())
			options := composer.FieldsComposerOptions{}

			Expect(SetClusterDependentOptions(context.Background(), &options, capabilities)).To(Succeed())
			Expect(options.GpuMemoryLimitAnnotationsSupported).To(Equal(expected))
		},
		Entry("falls back to env vars before 3.0", "2.99.99", false),
		Entry("emits new annotations from 3.0", "3.0.0", true),
	)
	DescribeTable("gates GPU memory requests independently from limits",
		func(version string, requests, limits bool) {
			capabilities, err := checker.NewClusterCapabilitiesCheckerForVersion(version)
			Expect(err).NotTo(HaveOccurred())
			options := composer.FieldsComposerOptions{}
			Expect(SetClusterDependentOptions(context.Background(), &options, capabilities)).To(Succeed())
			Expect(options.GpuMemoryRequestAnnotationsSupported).To(Equal(requests))
			Expect(options.GpuMemoryLimitAnnotationsSupported).To(Equal(limits))
		},
		Entry("older clusters", "2.99.99", false, false),
		Entry("existing limit support", "3.0.0", false, true),
		Entry("last version before request support", "3.0.22", false, true),
		Entry("first request-capable release", "3.0.23", true, true),
		Entry("later patch release", "3.0.24", true, true),
		Entry("later minor release", "3.1.0", true, true),
		Entry("development build", "0.0.0-dev", true, true),
	)
})
