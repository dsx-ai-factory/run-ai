/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/extensions"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/extensions/distributed"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Distributed", func() {
	Context("isDistributedFrameworkSuspendable", func() {
		var testDeps extensions.Group

		BeforeEach(func() {
			testDeps = extensions.Group{
				distributed.MPI: extensions.NewExtension(distributed.MPI,
					true, map[extensions.CapabilityType]bool{distributed.Suspension: true}),
				distributed.Pytorch: extensions.NewExtension(distributed.Pytorch,
					true, map[extensions.CapabilityType]bool{distributed.Suspension: false}),
				distributed.Jax: extensions.NewExtension(distributed.Jax,
					true, map[extensions.CapabilityType]bool{distributed.Suspension: true}),
				distributed.Tensorflow: extensions.NewExtension(distributed.Tensorflow,
					false, nil),
			}
		})

		It("suspendable", func() {
			suspendable, err := isDistributedFrameworkSuspendable(testDeps, constants.KindMpiJob)
			Expect(err).NotTo(HaveOccurred())
			Expect(suspendable).To(BeTrue())
		})

		It("available not suspendable", func() {
			suspendable, err := isDistributedFrameworkSuspendable(testDeps, constants.KindPyTorchJob)
			Expect(err).NotTo(HaveOccurred())
			Expect(suspendable).To(BeFalse())
		})

		It("not available not suspendable", func() {
			suspendable, err := isDistributedFrameworkSuspendable(testDeps, constants.KindPyTorchJob)
			Expect(err).NotTo(HaveOccurred())
			Expect(suspendable).To(BeFalse())
		})
		It("extensions status unknown", func() {
			_, err := isDistributedFrameworkSuspendable(testDeps, constants.KindXGBoostJob)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("missing extension framework status for XGBoostJob"))
		})

		It("unknown framework", func() {
			suspendable, err := isDistributedFrameworkSuspendable(testDeps, "BlaJob")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unable to find the extension for the requested framework BlaJob"))
			Expect(suspendable).To(BeFalse())
		})

	})
})
