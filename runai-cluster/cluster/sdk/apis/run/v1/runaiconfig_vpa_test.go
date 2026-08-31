/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/global"
	wc "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/workload-controller"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/utils/ptr"
)

var _ = Describe("RunaiConfig VPA", func() {
	Context("VPA propagation", func() {
		When("global VPA is set", func() {
			var rc *RunaiConfig

			BeforeEach(func() {
				updateMode := vpav1.UpdateModeInPlaceOrRecreate
				rc = &RunaiConfig{
					Spec: RunaiConfigSpec{
						Global: &global.Global{
							VPA: &global.VPASpec{
								Enabled: ptr.To(true),
								UpdatePolicy: &vpav1.PodUpdatePolicy{
									UpdateMode: &updateMode,
								},
							},
						},
					},
				}
			})

			It("should propagate global VPA to workload-controller", func() {
				rc.SetDefaultsWhereNeeded()

				Expect(rc.Spec.WorkloadController.VPA).NotTo(BeNil())
				Expect(rc.Spec.WorkloadController.VPA.Enabled).NotTo(BeNil())
				Expect(*rc.Spec.WorkloadController.VPA.Enabled).To(BeTrue())
				Expect(rc.Spec.WorkloadController.VPA.UpdatePolicy).NotTo(BeNil())
				Expect(*rc.Spec.WorkloadController.VPA.UpdatePolicy.UpdateMode).To(Equal(vpav1.UpdateModeInPlaceOrRecreate))
			})
		})

		When("group VPA is set", func() {
			var rc *RunaiConfig

			BeforeEach(func() {
				updateMode := vpav1.UpdateMode("Auto")
				rc = &RunaiConfig{
					Spec: RunaiConfigSpec{
						Global: &global.Global{
							WorkloadServices: &global.WorkloadServices{
								VPA: &global.VPASpec{
									Enabled: ptr.To(true),
									UpdatePolicy: &vpav1.PodUpdatePolicy{
										UpdateMode: &updateMode,
									},
								},
							},
						},
					},
				}
			})

			It("should propagate group VPA to workload-controller", func() {
				rc.SetDefaultsWhereNeeded()

				Expect(rc.Spec.WorkloadController.VPA).NotTo(BeNil())
				Expect(rc.Spec.WorkloadController.VPA.Enabled).NotTo(BeNil())
				Expect(*rc.Spec.WorkloadController.VPA.Enabled).To(BeTrue())
				Expect(rc.Spec.WorkloadController.VPA.UpdatePolicy).NotTo(BeNil())
				Expect(*rc.Spec.WorkloadController.VPA.UpdatePolicy.UpdateMode).To(Equal(vpav1.UpdateMode("Auto")))
			})
		})

		When("individual VPA is set", func() {
			var rc *RunaiConfig

			BeforeEach(func() {
				updateMode := vpav1.UpdateMode("Off")
				rc = &RunaiConfig{
					Spec: RunaiConfigSpec{
						WorkloadController: &wc.WorkloadController{
							VPA: &global.VPASpec{
								Enabled: ptr.To(false),
								UpdatePolicy: &vpav1.PodUpdatePolicy{
									UpdateMode: &updateMode,
								},
							},
						},
					},
				}
			})

			It("should keep individual VPA settings", func() {
				rc.SetDefaultsWhereNeeded()

				Expect(rc.Spec.WorkloadController.VPA).NotTo(BeNil())
				Expect(rc.Spec.WorkloadController.VPA.Enabled).NotTo(BeNil())
				Expect(*rc.Spec.WorkloadController.VPA.Enabled).To(BeFalse())
				Expect(rc.Spec.WorkloadController.VPA.UpdatePolicy).NotTo(BeNil())
				Expect(*rc.Spec.WorkloadController.VPA.UpdatePolicy.UpdateMode).To(Equal(vpav1.UpdateMode("Off")))
			})
		})

		When("group VPA overrides global VPA", func() {
			var rc *RunaiConfig

			BeforeEach(func() {
				globalMode := vpav1.UpdateModeInPlaceOrRecreate
				groupMode := vpav1.UpdateMode("Auto")
				rc = &RunaiConfig{
					Spec: RunaiConfigSpec{
						Global: &global.Global{
							VPA: &global.VPASpec{
								Enabled: ptr.To(true),
								UpdatePolicy: &vpav1.PodUpdatePolicy{
									UpdateMode: &globalMode,
								},
							},
							WorkloadServices: &global.WorkloadServices{
								VPA: &global.VPASpec{
									Enabled: ptr.To(false),
									UpdatePolicy: &vpav1.PodUpdatePolicy{
										UpdateMode: &groupMode,
									},
								},
							},
						},
					},
				}
			})

			It("should use group VPA over global VPA", func() {
				rc.SetDefaultsWhereNeeded()

				Expect(rc.Spec.WorkloadController.VPA).NotTo(BeNil())
				Expect(*rc.Spec.WorkloadController.VPA.Enabled).To(BeFalse())
				Expect(*rc.Spec.WorkloadController.VPA.UpdatePolicy.UpdateMode).To(Equal(vpav1.UpdateMode("Auto")))
			})
		})

		When("individual VPA overrides group VPA", func() {
			var rc *RunaiConfig

			BeforeEach(func() {
				groupMode := vpav1.UpdateMode("Auto")
				individualMode := vpav1.UpdateMode("Off")
				rc = &RunaiConfig{
					Spec: RunaiConfigSpec{
						Global: &global.Global{
							WorkloadServices: &global.WorkloadServices{
								VPA: &global.VPASpec{
									Enabled: ptr.To(true),
									UpdatePolicy: &vpav1.PodUpdatePolicy{
										UpdateMode: &groupMode,
									},
								},
							},
						},
						WorkloadController: &wc.WorkloadController{
							VPA: &global.VPASpec{
								Enabled: ptr.To(false),
								UpdatePolicy: &vpav1.PodUpdatePolicy{
									UpdateMode: &individualMode,
								},
							},
						},
					},
				}
			})

			It("should use individual VPA over group VPA", func() {
				rc.SetDefaultsWhereNeeded()

				Expect(rc.Spec.WorkloadController.VPA).NotTo(BeNil())
				Expect(*rc.Spec.WorkloadController.VPA.Enabled).To(BeFalse())
				Expect(*rc.Spec.WorkloadController.VPA.UpdatePolicy.UpdateMode).To(Equal(vpav1.UpdateMode("Off")))
			})
		})

		When("individual VPA overrides global VPA", func() {
			var rc *RunaiConfig

			BeforeEach(func() {
				globalMode := vpav1.UpdateModeInPlaceOrRecreate
				individualMode := vpav1.UpdateMode("Off")
				rc = &RunaiConfig{
					Spec: RunaiConfigSpec{
						Global: &global.Global{
							VPA: &global.VPASpec{
								Enabled: ptr.To(true),
								UpdatePolicy: &vpav1.PodUpdatePolicy{
									UpdateMode: &globalMode,
								},
							},
						},
						WorkloadController: &wc.WorkloadController{
							VPA: &global.VPASpec{
								Enabled: ptr.To(false),
								UpdatePolicy: &vpav1.PodUpdatePolicy{
									UpdateMode: &individualMode,
								},
							},
						},
					},
				}
			})

			It("should use individual VPA over global VPA", func() {
				rc.SetDefaultsWhereNeeded()

				Expect(rc.Spec.WorkloadController.VPA).NotTo(BeNil())
				Expect(*rc.Spec.WorkloadController.VPA.Enabled).To(BeFalse())
				Expect(*rc.Spec.WorkloadController.VPA.UpdatePolicy.UpdateMode).To(Equal(vpav1.UpdateMode("Off")))
			})
		})

		When("individual VPA overrides both group and global VPA", func() {
			var rc *RunaiConfig

			BeforeEach(func() {
				globalMode := vpav1.UpdateModeInPlaceOrRecreate
				groupMode := vpav1.UpdateMode("Auto")
				individualMode := vpav1.UpdateMode("Initial")
				rc = &RunaiConfig{
					Spec: RunaiConfigSpec{
						Global: &global.Global{
							VPA: &global.VPASpec{
								Enabled: ptr.To(true),
								UpdatePolicy: &vpav1.PodUpdatePolicy{
									UpdateMode: &globalMode,
								},
							},
							WorkloadServices: &global.WorkloadServices{
								VPA: &global.VPASpec{
									Enabled: ptr.To(true),
									UpdatePolicy: &vpav1.PodUpdatePolicy{
										UpdateMode: &groupMode,
									},
								},
							},
						},
						WorkloadController: &wc.WorkloadController{
							VPA: &global.VPASpec{
								Enabled: ptr.To(false),
								UpdatePolicy: &vpav1.PodUpdatePolicy{
									UpdateMode: &individualMode,
								},
							},
						},
					},
				}
			})

			It("should use individual VPA over both group and global VPA", func() {
				rc.SetDefaultsWhereNeeded()

				Expect(rc.Spec.WorkloadController.VPA).NotTo(BeNil())
				Expect(*rc.Spec.WorkloadController.VPA.Enabled).To(BeFalse())
				Expect(*rc.Spec.WorkloadController.VPA.UpdatePolicy.UpdateMode).To(Equal(vpav1.UpdateMode("Initial")))
			})
		})
	})
})
