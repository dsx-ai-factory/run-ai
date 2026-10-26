/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package workloadutils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	workloadsApi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
)

var _ = Describe("Synchronize Initialization Timeouts", func() {
	var responseSpec *workloadsApi.SupersetSpec

	BeforeEach(func() {
		responseSpec = &workloadsApi.SupersetSpec{}
	})

	Context("when ServingConfiguration InitializationTimeoutSeconds is set", func() {
		BeforeEach(func() {
			responseSpec.ServingConfiguration = *workloadsApi.NewNullableServingConfiguration(&workloadsApi.ServingConfiguration{
				InitializationTimeoutSeconds: *workloadsApi.NewNullableInt32(convutils.Ref(int32(10))),
			})
		})

		Context("and Autoscaling InitializationTimeoutSeconds exists", func() {
			BeforeEach(func() {
				responseSpec.Autoscaling = *workloadsApi.NewNullableAutoScaling(&workloadsApi.AutoScaling{})
			})

			It("should set Autoscaling timeout to ServingConfiguration timeout", func() {
				SynchronizeInitializationTimeouts(responseSpec)
				Expect(responseSpec.Autoscaling.IsSet()).To(BeTrue())
				Expect(responseSpec.Autoscaling.Get().InitializationTimeoutSeconds.IsSet()).To(BeTrue())
				Expect(*responseSpec.Autoscaling.Get().InitializationTimeoutSeconds.Get()).To(Equal(int32(10)))
			})
		})

		Context("and Autoscaling InitializationTimeoutSeconds is nil", func() {
			BeforeEach(func() {
				responseSpec.Autoscaling = workloadsApi.NullableAutoScaling{} // Already not a pointer
			})
			It("should create Autoscaling config and set its timeout", func() {
				SynchronizeInitializationTimeouts(responseSpec)
				Expect(responseSpec.Autoscaling.IsSet()).To(BeTrue())
				Expect(responseSpec.Autoscaling.Get().InitializationTimeoutSeconds.IsSet()).To(BeTrue())
				Expect(*responseSpec.Autoscaling.Get().InitializationTimeoutSeconds.Get()).To(Equal(int32(10)))
			})
		})
	})

	Context("when ServingConfiguration InitializationTimeoutSeconds is NOT set", func() {
		BeforeEach(func() {
			responseSpec.ServingConfiguration = *workloadsApi.NewNullableServingConfiguration(&workloadsApi.ServingConfiguration{})
		})

		Context("and Autoscaling timeout is set", func() {
			BeforeEach(func() {
				responseSpec.Autoscaling = *workloadsApi.NewNullableAutoScaling(&workloadsApi.AutoScaling{
					InitializationTimeoutSeconds: *workloadsApi.NewNullableInt32(convutils.Ref(int32(20))),
				})
			})

			Context("and ServingConfiguration exists", func() {
				It("should set ServingConfiguration timeout to Autoscaling timeout", func() {
					SynchronizeInitializationTimeouts(responseSpec)
					Expect(responseSpec.ServingConfiguration.IsSet()).To(BeTrue())
					Expect(responseSpec.ServingConfiguration.Get().InitializationTimeoutSeconds.IsSet()).To(BeTrue())
					Expect(*responseSpec.ServingConfiguration.Get().InitializationTimeoutSeconds.Get()).To(Equal(int32(20)))
				})
			})

			Context("and ServingConfiguration is nil initially (will be created by function)", func() {
				BeforeEach(func() {
					responseSpec.ServingConfiguration = workloadsApi.NullableServingConfiguration{} // Already not a pointer
				})
				It("should create ServingConfiguration and set its timeout", func() {
					SynchronizeInitializationTimeouts(responseSpec)
					Expect(responseSpec.ServingConfiguration.IsSet()).To(BeTrue())
					Expect(responseSpec.ServingConfiguration.Get().InitializationTimeoutSeconds.IsSet()).To(BeTrue())
					Expect(*responseSpec.ServingConfiguration.Get().InitializationTimeoutSeconds.Get()).To(Equal(int32(20)))
				})
			})
		})

		Context("and neither Autoscaling nor ServingConfiguration timeouts are set", func() {
			BeforeEach(func() {
				responseSpec.Autoscaling = *workloadsApi.NewNullableAutoScaling(&workloadsApi.AutoScaling{})
				responseSpec.ServingConfiguration = *workloadsApi.NewNullableServingConfiguration(&workloadsApi.ServingConfiguration{})
			})
			It("should not change anything", func() {
				SynchronizeInitializationTimeouts(responseSpec)
				Expect(responseSpec.Autoscaling.Get().InitializationTimeoutSeconds.IsSet()).To(BeFalse())
				Expect(responseSpec.ServingConfiguration.Get().InitializationTimeoutSeconds.IsSet()).To(BeFalse())
			})
		})
	})

	Context("when Autoscaling is nil", func() {
		BeforeEach(func() {
			responseSpec.Autoscaling = workloadsApi.NullableAutoScaling{} // Already not a pointer
		})
		Context("and ServingConfiguration timeout is set", func() {
			BeforeEach(func() {
				responseSpec.ServingConfiguration = *workloadsApi.NewNullableServingConfiguration(&workloadsApi.ServingConfiguration{
					InitializationTimeoutSeconds: *workloadsApi.NewNullableInt32(convutils.Ref(int32(30))),
				})
			})
			It("should create Autoscaling and set its timeout", func() {
				SynchronizeInitializationTimeouts(responseSpec)
				Expect(responseSpec.Autoscaling.IsSet()).To(BeTrue())
				Expect(responseSpec.Autoscaling.Get().InitializationTimeoutSeconds.IsSet()).To(BeTrue())
				Expect(*responseSpec.Autoscaling.Get().InitializationTimeoutSeconds.Get()).To(Equal(int32(30)))
			})
		})
	})

	Context("when ServingConfiguration is nil", func() {
		BeforeEach(func() {
			responseSpec.ServingConfiguration = workloadsApi.NullableServingConfiguration{} // Already not a pointer
		})
		Context("and Autoscaling timeout is set", func() {
			BeforeEach(func() {
				responseSpec.Autoscaling = *workloadsApi.NewNullableAutoScaling(&workloadsApi.AutoScaling{
					InitializationTimeoutSeconds: *workloadsApi.NewNullableInt32(convutils.Ref(int32(40))),
				})
			})
			It("should create ServingConfiguration and set its timeout", func() {
				SynchronizeInitializationTimeouts(responseSpec)
				Expect(responseSpec.ServingConfiguration.IsSet()).To(BeTrue())
				Expect(responseSpec.ServingConfiguration.Get().InitializationTimeoutSeconds.IsSet()).To(BeTrue())
				Expect(*responseSpec.ServingConfiguration.Get().InitializationTimeoutSeconds.Get()).To(Equal(int32(40)))
			})
		})
	})

	Context("when both Autoscaling and ServingConfiguration are nil initially", func() {
		BeforeEach(func() {
			responseSpec.Autoscaling = workloadsApi.NullableAutoScaling{}
			responseSpec.ServingConfiguration = workloadsApi.NullableServingConfiguration{}
		})
		It("should do nothing and not panic", func() {
			Expect(func() { SynchronizeInitializationTimeouts(responseSpec) }).ShouldNot(Panic())
			Expect(responseSpec.Autoscaling.IsSet()).To(BeFalse())
			Expect(responseSpec.ServingConfiguration.IsSet()).To(BeFalse())
		})
	})
})
