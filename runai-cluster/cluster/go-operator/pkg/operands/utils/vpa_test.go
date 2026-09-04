/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package utils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1/global"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	"k8s.io/utils/ptr"
)

var _ = Describe("BuildVPA", func() {
	var (
		name      = "test-vpa"
		namespace = "test-namespace"
		targetRef = autoscalingv1.CrossVersionObjectReference{
			Kind:       "Deployment",
			Name:       "test-deployment",
			APIVersion: "apps/v1",
		}
	)

	Context("when VPA should not be created", func() {
		It("should return nil when vpaSpec is nil", func() {
			result := BuildVPA(name, namespace, targetRef, nil)
			Expect(result).To(BeNil())
		})

		It("should return nil when vpaSpec.Enabled is nil", func() {
			vpaSpec := &global.VPASpec{}
			result := BuildVPA(name, namespace, targetRef, vpaSpec)
			Expect(result).To(BeNil())
		})

		It("should return nil when vpaSpec.Enabled is false", func() {
			vpaSpec := &global.VPASpec{
				Enabled: ptr.To(false),
			}
			result := BuildVPA(name, namespace, targetRef, vpaSpec)
			Expect(result).To(BeNil())
		})
	})

	Context("when VPA should be created", func() {
		It("should return a VPA with correct name and namespace", func() {
			vpaSpec := &global.VPASpec{
				Enabled: ptr.To(true),
			}
			result := BuildVPA(name, namespace, targetRef, vpaSpec)

			Expect(result).NotTo(BeNil())
			Expect(result.Name).To(Equal(name))
			Expect(result.Namespace).To(Equal(namespace))
		})

		It("should set the correct targetRef", func() {
			vpaSpec := &global.VPASpec{
				Enabled: ptr.To(true),
			}
			result := BuildVPA(name, namespace, targetRef, vpaSpec)

			Expect(result).NotTo(BeNil())
			Expect(result.Spec.TargetRef).NotTo(BeNil())
			Expect(result.Spec.TargetRef.Kind).To(Equal("Deployment"))
			Expect(result.Spec.TargetRef.Name).To(Equal("test-deployment"))
			Expect(result.Spec.TargetRef.APIVersion).To(Equal("apps/v1"))
		})

		It("should use default UpdatePolicy when UpdatePolicy is nil", func() {
			vpaSpec := &global.VPASpec{
				Enabled: ptr.To(true),
			}
			result := BuildVPA(name, namespace, targetRef, vpaSpec)

			Expect(result).NotTo(BeNil())
			Expect(result.Spec.UpdatePolicy).NotTo(BeNil())
			Expect(*result.Spec.UpdatePolicy.UpdateMode).To(Equal(vpav1.UpdateModeInPlaceOrRecreate))
		})

		It("should use default UpdateMode when UpdatePolicy has nil UpdateMode", func() {
			minReplicas := int32(2)
			vpaSpec := &global.VPASpec{
				Enabled: ptr.To(true),
				UpdatePolicy: &vpav1.PodUpdatePolicy{
					MinReplicas: &minReplicas,
				},
			}
			result := BuildVPA(name, namespace, targetRef, vpaSpec)

			Expect(result).NotTo(BeNil())
			Expect(result.Spec.UpdatePolicy).NotTo(BeNil())
			Expect(*result.Spec.UpdatePolicy.UpdateMode).To(Equal(vpav1.UpdateModeInPlaceOrRecreate))
			Expect(*result.Spec.UpdatePolicy.MinReplicas).To(Equal(int32(2)))
		})

		It("should set UpdatePolicy when provided", func() {
			updateMode := vpav1.UpdateMode("Auto")
			minReplicas := int32(3)
			vpaSpec := &global.VPASpec{
				Enabled: ptr.To(true),
				UpdatePolicy: &vpav1.PodUpdatePolicy{
					UpdateMode:  &updateMode,
					MinReplicas: &minReplicas,
				},
			}
			result := BuildVPA(name, namespace, targetRef, vpaSpec)

			Expect(result).NotTo(BeNil())
			Expect(result.Spec.UpdatePolicy).NotTo(BeNil())
			Expect(*result.Spec.UpdatePolicy.UpdateMode).To(Equal(vpav1.UpdateMode("Auto")))
			Expect(*result.Spec.UpdatePolicy.MinReplicas).To(Equal(int32(3)))
		})

		It("should not set ResourcePolicy when it is nil", func() {
			vpaSpec := &global.VPASpec{
				Enabled: ptr.To(true),
			}
			result := BuildVPA(name, namespace, targetRef, vpaSpec)

			Expect(result).NotTo(BeNil())
			Expect(result.Spec.ResourcePolicy).To(BeNil())
		})

		It("should set ResourcePolicy when provided", func() {
			containerPolicy := vpav1.ContainerResourcePolicy{
				ContainerName: "test-container",
				MinAllowed: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("100m"),
					v1.ResourceMemory: resource.MustParse("128Mi"),
				},
				MaxAllowed: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("2"),
					v1.ResourceMemory: resource.MustParse("4Gi"),
				},
			}
			vpaSpec := &global.VPASpec{
				Enabled: ptr.To(true),
				ResourcePolicy: &vpav1.PodResourcePolicy{
					ContainerPolicies: []vpav1.ContainerResourcePolicy{containerPolicy},
				},
			}
			result := BuildVPA(name, namespace, targetRef, vpaSpec)

			Expect(result).NotTo(BeNil())
			Expect(result.Spec.ResourcePolicy).NotTo(BeNil())
			Expect(result.Spec.ResourcePolicy.ContainerPolicies).To(HaveLen(1))
			Expect(result.Spec.ResourcePolicy.ContainerPolicies[0].ContainerName).To(Equal("test-container"))
			Expect(result.Spec.ResourcePolicy.ContainerPolicies[0].MinAllowed.Cpu().String()).To(Equal("100m"))
			Expect(result.Spec.ResourcePolicy.ContainerPolicies[0].MaxAllowed.Memory().String()).To(Equal("4Gi"))
		})

		It("should set all fields when fully configured", func() {
			updateMode := vpav1.UpdateModeInPlaceOrRecreate
			minReplicas := int32(2)
			containerPolicy := vpav1.ContainerResourcePolicy{
				ContainerName: "main",
				MinAllowed: v1.ResourceList{
					v1.ResourceCPU: resource.MustParse("50m"),
				},
			}
			vpaSpec := &global.VPASpec{
				Enabled: ptr.To(true),
				UpdatePolicy: &vpav1.PodUpdatePolicy{
					UpdateMode:  &updateMode,
					MinReplicas: &minReplicas,
				},
				ResourcePolicy: &vpav1.PodResourcePolicy{
					ContainerPolicies: []vpav1.ContainerResourcePolicy{containerPolicy},
				},
			}
			result := BuildVPA(name, namespace, targetRef, vpaSpec)

			Expect(result).NotTo(BeNil())
			Expect(result.Name).To(Equal(name))
			Expect(result.Namespace).To(Equal(namespace))
			Expect(result.Spec.TargetRef.Name).To(Equal("test-deployment"))
			Expect(*result.Spec.UpdatePolicy.UpdateMode).To(Equal(vpav1.UpdateModeInPlaceOrRecreate))
			Expect(*result.Spec.UpdatePolicy.UpdateMode).To(Equal(vpav1.UpdateMode("InPlaceOrRecreate")))
			Expect(*result.Spec.UpdatePolicy.MinReplicas).To(Equal(int32(2)))
			Expect(result.Spec.ResourcePolicy.ContainerPolicies[0].ContainerName).To(Equal("main"))
		})
	})
})
