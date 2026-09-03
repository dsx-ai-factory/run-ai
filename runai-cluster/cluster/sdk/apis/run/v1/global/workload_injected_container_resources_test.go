/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package global

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("WorkloadInjectedContainerResourceRequirements", func() {
	It("accepts valid requests and limits", func() {
		resources := &WorkloadInjectedContainerResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("100m"),
				corev1.ResourceMemory: resource.MustParse("128Mi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("200m"),
				corev1.ResourceMemory: resource.MustParse("256Mi"),
			},
		}

		Expect(resources.Validate()).To(Succeed())
	})

	It("rejects invalid resource names", func() {
		resources := &WorkloadInjectedContainerResourceRequirements{
			Requests: corev1.ResourceList{"invalid resource": resource.MustParse("1")},
		}

		Expect(resources.Validate()).To(MatchError(ContainSubstring("invalid resource name")))
	})

	It("rejects negative quantities", func() {
		resources := &WorkloadInjectedContainerResourceRequirements{
			Limits: corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("-1m")},
		}

		Expect(resources.Validate()).To(MatchError(ContainSubstring("must not be negative")))
	})

	It("rejects requests that exceed limits", func() {
		resources := &WorkloadInjectedContainerResourceRequirements{
			Requests: corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("256Mi")},
			Limits:   corev1.ResourceList{corev1.ResourceMemory: resource.MustParse("128Mi")},
		}

		Expect(resources.Validate()).To(MatchError(ContainSubstring("must not exceed")))
	})

	Describe("ResourceRequirementsFromStringMaps", func() {
		It("parses and validates resource maps", func() {
			resources, err := ResourceRequirementsFromStringMaps(
				map[string]string{"cpu": "200m"},
				map[string]string{"cpu": "100m"},
			)

			Expect(err).NotTo(HaveOccurred())
			Expect(resources.Limits.Cpu().Equal(resource.MustParse("200m"))).To(BeTrue())
			Expect(resources.Requests.Cpu().Equal(resource.MustParse("100m"))).To(BeTrue())
		})

		It("rejects invalid quantities and request-to-limit relationships", func() {
			_, err := ResourceRequirementsFromStringMaps(nil, map[string]string{"cpu": "invalid"})
			Expect(err).To(MatchError(ContainSubstring(`requests.cpu has invalid quantity "invalid"`)))

			_, err = ResourceRequirementsFromStringMaps(
				map[string]string{"memory": "128Mi"},
				map[string]string{"memory": "256Mi"},
			)
			Expect(err).To(MatchError(ContainSubstring("must not exceed")))
		})
	})
})
