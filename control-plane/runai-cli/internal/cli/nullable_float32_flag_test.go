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

var _ = Describe("NullableFloat32", func() {
	var nf *cli.NullableFloat32Flag

	BeforeEach(func() {
		nf = cli.NewNullableFloat32Flag(&openapi.NullableFloat32{})
	})

	Describe("Set", func() {
		It("should set a valid float value", func() {
			err := nf.Set("3.14")
			Expect(err).NotTo(HaveOccurred())
			Expect(nf.IsSet()).To(BeTrue())
			Expect(*nf.Value()).To(Equal(float32(3.14)))
		})

		It("should fail on invalid input", func() {
			err := nf.Set("invalid")
			Expect(err).To(HaveOccurred())
			Expect(nf.IsSet()).To(BeFalse())
		})
	})

	Describe("String", func() {
		It("should return an empty string when not set", func() {
			Expect(nf.String()).To(Equal(""))
		})

		It("should return the float value as a string when set", func() {
			err := nf.Set("4.56")
			if err != nil {
				Fail("Failed to set value")
			}
			Expect(nf.String()).To(Equal("4.560000"))
		})
	})

	Describe("IsSet", func() {
		It("should return false when value is not set", func() {
			Expect(nf.IsSet()).To(BeFalse())
		})

		It("should return true when value is set", func() {
			err := nf.Set("2.71")
			if err != nil {
				Fail("Failed to set value")
			}
			Expect(nf.IsSet()).To(BeTrue())
		})
	})
})
