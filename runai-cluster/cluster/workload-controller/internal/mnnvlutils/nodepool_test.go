/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package mnnvlutils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	kaiv1alpha1 "github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("NodePool MNNVL Functions", func() {
	const accelLabel = "vendor.com/clique"

	kaiPool := func(name string, annotations map[string]string) kaiv1alpha1.NodePool {
		return kaiv1alpha1.NodePool{
			ObjectMeta: metav1.ObjectMeta{Name: name, Annotations: annotations},
		}
	}

	detected := func(value string) map[string]string {
		return map[string]string{
			kaiv1alpha1.AnnotationGPUNetworkAccelerationDetected: value,
			kaiv1alpha1.AnnotationGPUNetworkAccelerationLabelKey: accelLabel,
		}
	}

	Describe("isNodePoolValidForMNNVL", func() {
		DescribeTable("reads the detection annotation",
			func(annotations map[string]string, expected bool) {
				Expect(isNodePoolValidForMNNVL(kaiPool("np1", annotations))).To(Equal(expected))
			},
			Entry("detected", detected("true"), true),
			Entry("not detected", detected("false"), false),
			Entry("empty value", detected(""), false),
			// Acceleration relaxes MNNVL placement, so an unreadable annotation must not
			// be taken as detected — that would schedule onto unsupported nodes.
			Entry("unparseable value", detected("yes-please"), false),
			Entry("no annotations at all", nil, false),
		)

		// The legacy CRD schema defaulted the label key, so a pool with detection but no
		// key annotation is still valid rather than being rejected for an empty key.
		It("accepts a detected pool with no label-key annotation", func() {
			pool := kaiPool("np1", map[string]string{
				kaiv1alpha1.AnnotationGPUNetworkAccelerationDetected: "true",
			})

			Expect(isNodePoolValidForMNNVL(pool)).To(BeTrue())
		})
	})

	DescribeTable("getUnsupportedNodePools",
		func(nodePools []kaiv1alpha1.NodePool, expectedUnsupported []string) {
			unsupported := getUnsupportedNodePools(nodePools)

			if len(expectedUnsupported) == 0 {
				Expect(unsupported).To(HaveLen(0))
			} else {
				Expect(unsupported).To(ContainElements(expectedUnsupported))
			}
		},
		Entry("mixed nodepools",
			[]kaiv1alpha1.NodePool{
				kaiPool("supported-pool", detected("true")),
				kaiPool("unsupported-pool", detected("false")),
			},
			[]string{"unsupported-pool"},
		),
		Entry("all supported",
			[]kaiv1alpha1.NodePool{
				kaiPool("supported-pool-1", detected("true")),
				kaiPool("supported-pool-2", detected("true")),
			},
			nil,
		),
		Entry("all unsupported",
			[]kaiv1alpha1.NodePool{
				kaiPool("unsupported-pool-1", detected("false")),
				kaiPool("unsupported-pool-2", nil),
			},
			[]string{"unsupported-pool-1", "unsupported-pool-2"},
		),
		Entry("empty nodepools", []kaiv1alpha1.NodePool{}, nil),
		Entry("nil nodepools", nil, nil),
	)
})
