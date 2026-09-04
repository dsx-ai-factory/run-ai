/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gvk_compliance

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/resource"
)

var _ = Describe("quantity string parsing", func() {
	Context("ParseQuantityString", func() {
		It("parses a valid Kubernetes quantity string", func() {
			q, err := ParseQuantityString("512Mi")
			Expect(err).NotTo(HaveOccurred())
			Expect(q.String()).To(Equal("512Mi"))
		})

		It("rejects non-string input", func() {
			_, err := ParseQuantityString(42)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("expected string quantity"))
		})

		It("rejects an invalid quantity string", func() {
			_, err := ParseQuantityString("not-a-quantity")
			Expect(err).To(HaveOccurred())
		})
	})

	Context("ValidateNumericElementAgainstRange", func() {
		var min, max, step float32

		BeforeEach(func() {
			min = 2
			max = 10
			step = 2
		})

		It("accepts a value within min, max, and step", func() {
			Expect(ValidateNumericElementAgainstRange(float64(4), &min, &max, &step)).To(Succeed())
		})

		It("rejects a value below min", func() {
			Expect(ValidateNumericElementAgainstRange(float64(1), &min, &max, &step)).NotTo(Succeed())
		})

		It("rejects a value above max", func() {
			Expect(ValidateNumericElementAgainstRange(float64(11), &min, &max, &step)).NotTo(Succeed())
		})

		It("rejects a value off the step grid", func() {
			Expect(ValidateNumericElementAgainstRange(float64(5), &min, &max, &step)).NotTo(Succeed())
		})

		It("rejects a non-numeric element when step is set", func() {
			Expect(ValidateNumericElementAgainstRange("x", nil, nil, &step)).NotTo(Succeed())
		})
	})

	Context("ValidateQuantityValueAgainstRange", func() {
		var minQ, maxQ, stepQ resource.Quantity

		BeforeEach(func() {
			minQ = resource.MustParse("100Mi")
			maxQ = resource.MustParse("1Gi")
			stepQ = resource.MustParse("50Mi")
		})

		It("accepts a quantity within bounds on the step grid", func() {
			q := resource.MustParse("200Mi")
			Expect(ValidateQuantityValueAgainstRange(q, &minQ, &maxQ, &stepQ)).To(Succeed())
		})

		It("rejects a quantity below min", func() {
			q := resource.MustParse("50Mi")
			Expect(ValidateQuantityValueAgainstRange(q, &minQ, &maxQ, nil)).NotTo(Succeed())
		})

		It("rejects a quantity above max", func() {
			q := resource.MustParse("2Gi")
			Expect(ValidateQuantityValueAgainstRange(q, &minQ, &maxQ, nil)).NotTo(Succeed())
		})

		It("rejects a quantity off the step grid", func() {
			q := resource.MustParse("125Mi")
			Expect(ValidateQuantityValueAgainstRange(q, &minQ, &maxQ, &stepQ)).NotTo(Succeed())
		})
	})
})
