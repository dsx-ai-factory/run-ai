/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package update

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/common/constants"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("StatusUpdaterFactory", func() {

	It("returns InteractiveWorkloadStatusUpdaterFactory for InteractiveWorkload", func() {
		factory, err := GetStatusUpdaterFactory(constants.InteractiveWorkload)
		Expect(err).To(BeNil())
		Expect(factory).To(BeAssignableToTypeOf(&InteractiveWorkloadStatusUpdaterFactory{}))
	})

	It("returns TrainingWorkloadStatusUpdaterFactory for TrainingWorkload", func() {
		factory, err := GetStatusUpdaterFactory(constants.TrainingWorkload)
		Expect(err).To(BeNil())
		Expect(factory).To(BeAssignableToTypeOf(&TrainingWorkloadStatusUpdaterFactory{}))
	})

	It("returns DistributedWorkloadStatusUpdaterFactory for DistributedWorkload", func() {
		factory, err := GetStatusUpdaterFactory(constants.DistributedWorkload)
		Expect(err).To(BeNil())
		Expect(factory).To(BeAssignableToTypeOf(&DistributedWorkloadStatusUpdaterFactory{}))
	})

	It("returns error for unknown workload type", func() {
		_, err := GetStatusUpdaterFactory("unknown")
		Expect(err).ToNot(BeNil())
	})
	Context("factory", func() {
		var (
			k8sClient client.Client
		)

		BeforeEach(func() {
			k8sClient = fake.NewClientBuilder().Build()
		})

		It("returns InteractiveWorkloadStatusUpdater for InteractiveWorkload", func() {
			factory := &InteractiveWorkloadStatusUpdaterFactory{}
			updater := factory.CreateStatusUpdater(k8sClient)
			Expect(updater).To(BeAssignableToTypeOf(&WorkloadUpdater{}))
		})

		It("returns TrainingWorkloadStatusUpdater for TrainingWorkload", func() {
			factory := &TrainingWorkloadStatusUpdaterFactory{}
			updater := factory.CreateStatusUpdater(k8sClient)
			Expect(updater).To(BeAssignableToTypeOf(&WorkloadUpdater{}))
		})

		It("returns DistributedWorkloadStatusUpdater for DistributedWorkload", func() {
			factory := &DistributedWorkloadStatusUpdaterFactory{}
			updater := factory.CreateStatusUpdater(k8sClient)
			Expect(updater).To(BeAssignableToTypeOf(&WorkloadUpdater{}))
		})
	})

})
