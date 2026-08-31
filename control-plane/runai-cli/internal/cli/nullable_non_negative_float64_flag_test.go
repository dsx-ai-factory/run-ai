/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cli_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cli"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

var _ = Describe("NullableNonNegativeFloat64", func() {
	var nnf *cli.NullableNonNegativeFloat64Flag

	BeforeEach(func() {
		nnf = cli.NewNullableNonNegativeFloat64Flag(&openapi.NullableFloat64{})
	})

	Describe("Set", func() {
		It("should set a valid positive integer value", func() {
			err := nnf.Set("42")
			Expect(err).NotTo(HaveOccurred())
			Expect(nnf.IsSet()).To(BeTrue())
			Expect(*nnf.Value()).To(Equal(float64(42)))
		})

		It("should set a valid positive float value", func() {
			err := nnf.Set("3.14")
			Expect(err).NotTo(HaveOccurred())
			Expect(nnf.IsSet()).To(BeTrue())
			Expect(*nnf.Value()).To(Equal(float64(3.14)))
		})

		It("should set zero value", func() {
			err := nnf.Set("0")
			Expect(err).NotTo(HaveOccurred())
			Expect(nnf.IsSet()).To(BeTrue())
			Expect(*nnf.Value()).To(Equal(float64(0)))
		})

		It("should set zero float value", func() {
			err := nnf.Set("0.0")
			Expect(err).NotTo(HaveOccurred())
			Expect(nnf.IsSet()).To(BeTrue())
			Expect(*nnf.Value()).To(Equal(float64(0)))
		})

		It("should fail on negative value", func() {
			err := nnf.Set("-5")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("value cannot be negative"))
			Expect(nnf.IsSet()).To(BeFalse())
		})

		It("should fail on negative float value", func() {
			err := nnf.Set("-3.14")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("value cannot be negative"))
			Expect(nnf.IsSet()).To(BeFalse())
		})

		It("should fail on invalid input", func() {
			err := nnf.Set("invalid")
			Expect(err).To(HaveOccurred())
			Expect(nnf.IsSet()).To(BeFalse())
		})

		It("should fail on empty string", func() {
			err := nnf.Set("")
			Expect(err).To(HaveOccurred())
			Expect(nnf.IsSet()).To(BeFalse())
		})

		It("should handle large values", func() {
			err := nnf.Set("1000000.5")
			Expect(err).NotTo(HaveOccurred())
			Expect(nnf.IsSet()).To(BeTrue())
			Expect(*nnf.Value()).To(Equal(float64(1000000.5)))
		})

		It("should handle scientific notation", func() {
			err := nnf.Set("1.23e-4")
			Expect(err).NotTo(HaveOccurred())
			Expect(nnf.IsSet()).To(BeTrue())
			Expect(*nnf.Value()).To(Equal(float64(0.000123)))
		})
	})

	Describe("String", func() {
		It("should return an empty string when not set", func() {
			Expect(nnf.String()).To(Equal(""))
		})

		It("should return the integer value as a string when set", func() {
			err := nnf.Set("123")
			Expect(err).NotTo(HaveOccurred())
			Expect(nnf.String()).To(Equal("123"))
		})

		It("should return the float value as a string when set", func() {
			err := nnf.Set("123.456")
			Expect(err).NotTo(HaveOccurred())
			Expect(nnf.String()).To(Equal("123.456"))
		})

		It("should return zero as string", func() {
			err := nnf.Set("0")
			Expect(err).NotTo(HaveOccurred())
			Expect(nnf.String()).To(Equal("0"))
		})
	})

	Describe("IsSet", func() {
		It("should return false when value is not set", func() {
			Expect(nnf.IsSet()).To(BeFalse())
		})

		It("should return true when value is set", func() {
			err := nnf.Set("100.5")
			Expect(err).NotTo(HaveOccurred())
			Expect(nnf.IsSet()).To(BeTrue())
		})
	})

	Describe("Value", func() {
		It("should return nil when not set", func() {
			Expect(nnf.Value()).To(BeNil())
		})

		It("should return the correct value when set", func() {
			err := nnf.Set("999.99")
			Expect(err).NotTo(HaveOccurred())
			Expect(nnf.Value()).NotTo(BeNil())
			Expect(*nnf.Value()).To(Equal(float64(999.99)))
		})
	})

	Describe("Type", func() {
		It("should return 'positiveFloat'", func() {
			Expect(nnf.Type()).To(Equal("positiveFloat"))
		})
	})
})
