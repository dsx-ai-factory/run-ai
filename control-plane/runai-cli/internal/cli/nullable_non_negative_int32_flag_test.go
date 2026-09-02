/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cli_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cli"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
)

var _ = Describe("NullableNonNegativeInt32", func() {
	var nni *cli.NullableNonNegativeInt32Flag

	BeforeEach(func() {
		nni = cli.NewNullableNonNegativeInt32Flag(&openapi.NullableInt32{})
	})

	Describe("Set", func() {
		It("should set a valid positive integer value", func() {
			err := nni.Set("42")
			Expect(err).NotTo(HaveOccurred())
			Expect(nni.IsSet()).To(BeTrue())
			Expect(*nni.Value()).To(Equal(int32(42)))
		})

		It("should set zero value", func() {
			err := nni.Set("0")
			Expect(err).NotTo(HaveOccurred())
			Expect(nni.IsSet()).To(BeTrue())
			Expect(*nni.Value()).To(Equal(int32(0)))
		})

		It("should fail on negative value", func() {
			err := nni.Set("-5")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("value cannot be negative"))
			Expect(nni.IsSet()).To(BeFalse())
		})

		It("should fail on float input", func() {
			err := nni.Set("3.14")
			Expect(err).To(HaveOccurred())
			Expect(nni.IsSet()).To(BeFalse())
		})

		It("should fail on invalid input", func() {
			err := nni.Set("invalid")
			Expect(err).To(HaveOccurred())
			Expect(nni.IsSet()).To(BeFalse())
		})

		It("should fail on empty string", func() {
			err := nni.Set("")
			Expect(err).To(HaveOccurred())
			Expect(nni.IsSet()).To(BeFalse())
		})

		It("should handle maximum int32 value", func() {
			err := nni.Set("2147483647")
			Expect(err).NotTo(HaveOccurred())
			Expect(nni.IsSet()).To(BeTrue())
			Expect(*nni.Value()).To(Equal(int32(2147483647)))
		})

		It("should fail on value exceeding int32 range", func() {
			err := nni.Set("2147483648")
			Expect(err).To(HaveOccurred())
			Expect(nni.IsSet()).To(BeFalse())
		})
	})

	Describe("String", func() {
		It("should return an empty string when not set", func() {
			Expect(nni.String()).To(Equal(""))
		})

		It("should return the integer value as a string when set", func() {
			err := nni.Set("123")
			Expect(err).NotTo(HaveOccurred())
			Expect(nni.String()).To(Equal("123"))
		})

		It("should return zero as string", func() {
			err := nni.Set("0")
			Expect(err).NotTo(HaveOccurred())
			Expect(nni.String()).To(Equal("0"))
		})
	})

	Describe("IsSet", func() {
		It("should return false when value is not set", func() {
			Expect(nni.IsSet()).To(BeFalse())
		})

		It("should return true when value is set", func() {
			err := nni.Set("100")
			Expect(err).NotTo(HaveOccurred())
			Expect(nni.IsSet()).To(BeTrue())
		})
	})

	Describe("Value", func() {
		It("should return nil when not set", func() {
			Expect(nni.Value()).To(BeNil())
		})

		It("should return the correct value when set", func() {
			err := nni.Set("999")
			Expect(err).NotTo(HaveOccurred())
			Expect(nni.Value()).NotTo(BeNil())
			Expect(*nni.Value()).To(Equal(int32(999)))
		})
	})

	Describe("Type", func() {
		It("should return 'int32'", func() {
			Expect(nni.Type()).To(Equal("positiveInt"))
		})
	})
})
