/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package externalworkloads

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
)

var _ = Describe("equalExternalWorkloadSpec storage refs", func() {
	base := func() *v1alpha1.ExternalWorkloadSpec {
		return &v1alpha1.ExternalWorkloadSpec{
			StorageRefs: []v1alpha1.StorageRef{
				{Kind: v1alpha1.StorageRefKindPVC, Name: "data"},
				{Kind: v1alpha1.StorageRefKindSecret, Name: "regcred"},
			},
		}
	}

	It("treats identical storage refs as equal", func() {
		Expect(equalExternalWorkloadSpec(base(), base())).To(BeTrue())
	})

	It("treats an added storage ref as not equal", func() {
		right := base()
		right.StorageRefs = append(right.StorageRefs, v1alpha1.StorageRef{Kind: v1alpha1.StorageRefKindConfigMap, Name: "cfg"})
		Expect(equalExternalWorkloadSpec(base(), right)).To(BeFalse())
	})

	It("treats a changed storage ref name as not equal", func() {
		right := base()
		right.StorageRefs[0].Name = "other"
		Expect(equalExternalWorkloadSpec(base(), right)).To(BeFalse())
	})

	It("treats removed storage refs as not equal", func() {
		right := base()
		right.StorageRefs = nil
		Expect(equalExternalWorkloadSpec(base(), right)).To(BeFalse())
	})
})
