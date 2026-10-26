/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package composer_v2

import (
	"context"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies/composer"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
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
	DescribeTable("gates GPU compute mode by cluster version",
		func(version string, expected bool) {
			capabilities, err := checker.NewClusterCapabilitiesCheckerForVersion(version)
			Expect(err).NotTo(HaveOccurred())
			options := composer.FieldsComposerOptions{}
			Expect(SetClusterDependentOptions(context.Background(), &options, capabilities)).To(Succeed())
			fc := composer.FieldsComposer{Options: options}
			annotations := map[string]string{}
			composeErr := fc.ComposeGpuComputeMode(convutils.Ref("sm-sharing"), &composer.GpuSizingInfo{PortionRequest: convutils.Ref(0.25)}, "main", &annotations, composer.StringToMapString)()
			if expected {
				Expect(composeErr).To(BeNil())
				Expect(annotations).To(HaveKeyWithValue("nvidia.com/container.main.gpu-compute.mode", "sm-sharing"))
			} else {
				Expect(composeErr).NotTo(BeNil())
				Expect(composeErr.Code).To(Equal(int32(http.StatusBadRequest)))
				Expect(composeErr.Message).To(ContainSubstring("requires cluster version 3.0.14 or later"))
				Expect(annotations).To(BeEmpty())
			}
		},
		Entry("before the runtime is installed", "3.0.13", false),
		Entry("first supported release", "3.0.14", true),
		Entry("later release", "3.1.0", true),
		Entry("development build", "0.0.0-dev", true),
	)

})
