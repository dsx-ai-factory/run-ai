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

var _ = Describe("NullableInt32", func() {
	var (
		ni *cli.NullableInt32Flag
	)

	BeforeEach(func() {
		ni = cli.NewNullableInt32Flag(&openapi.NullableInt32{})
	})

	Describe("Set", func() {
		It("should set a valid integer value", func() {
			err := ni.Set("42")
			Expect(err).NotTo(HaveOccurred())
			Expect(ni.IsSet()).To(BeTrue())
			Expect(*ni.Value()).To(Equal(int32(42)))
		})

		It("should fail on invalid input", func() {
			err := ni.Set("invalid")
			Expect(err).To(HaveOccurred())
			Expect(ni.IsSet()).To(BeFalse())
		})
	})

	Describe("String", func() {
		It("should return an empty string when not set", func() {
			Expect(ni.String()).To(Equal(""))
		})

		It("should return the integer value as a string when set", func() {
			err := ni.Set("123")
			if err != nil {
				Fail("Failed to set value")
			}
			Expect(ni.String()).To(Equal("123"))
		})
	})

	Describe("IsSet", func() {
		It("should return false when value is not set", func() {
			Expect(ni.IsSet()).To(BeFalse())
		})

		It("should return true when value is set", func() {
			err := ni.Set("99")
			if err != nil {
				Fail("Failed to set value")
			}
			Expect(ni.IsSet()).To(BeTrue())
		})
	})
})
