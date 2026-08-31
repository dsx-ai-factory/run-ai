/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
)

// validatePolicyQuantityRangeBounds is exercised from post_validate when a set/default
// is checked against quantityRange peers; the same bounds are already validated per
// enforcement via validateQuantityRangeEnforcement, so these tests pin the helper directly.
var _ = Describe("validatePolicyQuantityRangeBounds", func() {
	It("rejects min greater than max", func() {
		minQ, err := resource.ParseQuantity("2Gi")
		Expect(err).NotTo(HaveOccurred())
		maxQ, err := resource.ParseQuantity("1Gi")
		Expect(err).NotTo(HaveOccurred())
		err = validatePolicyQuantityRangeBounds(&minQ, &maxQ, nil)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("less than or equal"))
		Expect(err.Error()).To(ContainSubstring("quantityRange.min"))
	})

	It("rejects non-positive step", func() {
		minQ, err := resource.ParseQuantity("0")
		Expect(err).NotTo(HaveOccurred())
		maxQ, err := resource.ParseQuantity("1Gi")
		Expect(err).NotTo(HaveOccurred())
		stepQ, err := resource.ParseQuantity("0")
		Expect(err).NotTo(HaveOccurred())
		err = validatePolicyQuantityRangeBounds(&minQ, &maxQ, &stepQ)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("greater than 0"))
	})

	It("rejects step wider than max-min when all three are set", func() {
		minQ, err := resource.ParseQuantity("100Mi")
		Expect(err).NotTo(HaveOccurred())
		maxQ, err := resource.ParseQuantity("200Mi")
		Expect(err).NotTo(HaveOccurred())
		stepQ, err := resource.ParseQuantity("500Mi")
		Expect(err).NotTo(HaveOccurred())
		err = validatePolicyQuantityRangeBounds(&minQ, &maxQ, &stepQ)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("less than or equal to quantityRange.max-quantityRange.min"))
	})

	It("accepts coherent min, max, and step", func() {
		minQ, err := resource.ParseQuantity("0")
		Expect(err).NotTo(HaveOccurred())
		maxQ, err := resource.ParseQuantity("1Gi")
		Expect(err).NotTo(HaveOccurred())
		stepQ, err := resource.ParseQuantity("256Mi")
		Expect(err).NotTo(HaveOccurred())
		Expect(validatePolicyQuantityRangeBounds(&minQ, &maxQ, &stepQ)).To(BeNil())
	})
})
