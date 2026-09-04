/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package finalizers

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetRemoveFinalizerFactory", func() {
	It("returns InteractiveWorkloadRemoveFinalizerFactory for InteractiveWorkload", func() {
		factory, err := GetRemoveFinalizerFactory(constants.InteractiveWorkload)
		Expect(err).To(BeNil())
		Expect(factory).To(BeAssignableToTypeOf(&InteractiveWorkloadRemoveFinalizerFactory{}))
	})

	It("returns TrainingWorkloadRemoveFinalizerFactory for TrainingWorkload", func() {
		factory, err := GetRemoveFinalizerFactory(constants.TrainingWorkload)
		Expect(err).To(BeNil())
		Expect(factory).To(BeAssignableToTypeOf(&TrainingWorkloadRemoveFinalizerFactory{}))
	})

	It("returns DistributedWorkloadRemoveFinalizerFactory for DistributedWorkload", func() {
		factory, err := GetRemoveFinalizerFactory(constants.DistributedWorkload)
		Expect(err).To(BeNil())
		Expect(factory).To(BeAssignableToTypeOf(&DistributedWorkloadRemoveFinalizerFactory{}))
	})

	It("returns error for unknown workload type", func() {
		_, err := GetRemoveFinalizerFactory("unknown")
		Expect(err).ToNot(BeNil())
	})
})
