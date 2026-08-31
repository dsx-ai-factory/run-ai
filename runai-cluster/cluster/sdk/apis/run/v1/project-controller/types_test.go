/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package project_controller

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	runv1_internal "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1/runai-internal"
	"k8s.io/utils/ptr"
)

var _ = Describe("Project Controller", func() {
	var (
		projectController *ProjectController
		internal          *runv1_internal.Internal
		globalRegistry    string
		globalReplicaCount *int32
	)

	BeforeEach(func() {
		projectController = &ProjectController{}
		internal = &runv1_internal.Internal{}
		internal.SetDefaultsWhereNeeded()
		globalRegistry = "gcr.io/run-ai"
		globalReplicaCount = ptr.To[int32](2)
	})

	Describe("SetDefaultsWhereNeeded", func() {
		Context("IsOpenShift field", func() {
			When("KubernetesFlavor is OpenShift", func() {
				BeforeEach(func() {
					internal.Global.Platform = &runv1_internal.Platform{
						KubernetesFlavor: runv1_internal.KubernetesFlavorOpenShift,
						Version:          "4.11",
					}
				})
				It("should set IsOpenShift to true", func() {
					projectController.SetDefaultsWhereNeeded(internal, globalRegistry, globalReplicaCount)
					Expect(projectController.IsOpenShift).To(BeTrue())
				})
			})

			When("KubernetesFlavor is Vanilla", func() {
				BeforeEach(func() {
					internal.Global.Platform = &runv1_internal.Platform{
						KubernetesFlavor: runv1_internal.KubernetesFlavorVanilla,
					}
				})
				It("should set IsOpenShift to false", func() {
					projectController.SetDefaultsWhereNeeded(internal, globalRegistry, globalReplicaCount)
					Expect(projectController.IsOpenShift).To(BeFalse())
				})
			})

			When("KubernetesFlavor is EKS", func() {
				BeforeEach(func() {
					internal.Global.Platform = &runv1_internal.Platform{
						KubernetesFlavor: runv1_internal.KubernetesFlavorEKS,
					}
				})
				It("should set IsOpenShift to false", func() {
					projectController.SetDefaultsWhereNeeded(internal, globalRegistry, globalReplicaCount)
					Expect(projectController.IsOpenShift).To(BeFalse())
				})
			})

			When("KubernetesFlavor is GKE", func() {
				BeforeEach(func() {
					internal.Global.Platform = &runv1_internal.Platform{
						KubernetesFlavor: runv1_internal.KubernetesFlavorGKE,
					}
				})
				It("should set IsOpenShift to false", func() {
					projectController.SetDefaultsWhereNeeded(internal, globalRegistry, globalReplicaCount)
					Expect(projectController.IsOpenShift).To(BeFalse())
				})
			})

			When("KubernetesFlavor is AKS", func() {
				BeforeEach(func() {
					internal.Global.Platform = &runv1_internal.Platform{
						KubernetesFlavor: runv1_internal.KubernetesFlavorAKS,
					}
				})
				It("should set IsOpenShift to false", func() {
					projectController.SetDefaultsWhereNeeded(internal, globalRegistry, globalReplicaCount)
					Expect(projectController.IsOpenShift).To(BeFalse())
				})
			})

			When("KubernetesFlavor is Unknown", func() {
				BeforeEach(func() {
					internal.Global.Platform = &runv1_internal.Platform{
						KubernetesFlavor: runv1_internal.KubernetesFlavorUnknown,
					}
				})
				It("should set IsOpenShift to false", func() {
					projectController.SetDefaultsWhereNeeded(internal, globalRegistry, globalReplicaCount)
					Expect(projectController.IsOpenShift).To(BeFalse())
				})
			})
		})

		Context("Other defaults", func() {
			BeforeEach(func() {
				internal.Global.Platform = &runv1_internal.Platform{
					KubernetesFlavor: runv1_internal.KubernetesFlavorVanilla,
				}
			})

			It("should set Enabled to true by default", func() {
				projectController.SetDefaultsWhereNeeded(internal, globalRegistry, globalReplicaCount)
				Expect(projectController.Enabled).To(Equal(ptr.To(true)))
			})

			It("should set Replicas to globalReplicaCount", func() {
				projectController.SetDefaultsWhereNeeded(internal, globalRegistry, globalReplicaCount)
				Expect(projectController.Replicas).To(Equal(ptr.To[int32](2)))
			})

			It("should set CreateNamespaces to true by default", func() {
				projectController.SetDefaultsWhereNeeded(internal, globalRegistry, globalReplicaCount)
				Expect(projectController.CreateNamespaces).To(Equal(ptr.To(true)))
			})

			It("should initialize Image with correct registry", func() {
				projectController.SetDefaultsWhereNeeded(internal, globalRegistry, globalReplicaCount)
				Expect(projectController.Image).ToNot(BeNil())
				Expect(*projectController.Image.Registry).To(Equal(globalRegistry))
			})
		})
	})
})
