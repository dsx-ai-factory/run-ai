/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFlags(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "CLI List Flags Tests")
}

var _ = Describe("RawFlagValidator", func() {
	Context("Valid flag combinations", func() {
		It("should allow raw flag with json", func() {
			output := &Output{}
			output.ofRaw = true
			output.ofJson = true

			err := output.RawFlagValidator()

			Expect(err).To(BeNil())
		})

		It("should allow raw flag with yaml", func() {
			output := &Output{}
			output.ofRaw = true
			output.ofYaml = true

			err := output.RawFlagValidator()

			Expect(err).To(BeNil())
		})

		It("should allow json flag alone", func() {
			output := &Output{}
			output.ofJson = true

			err := output.RawFlagValidator()

			Expect(err).To(BeNil())
		})

		It("should allow yaml flag alone", func() {
			output := &Output{}
			output.ofYaml = true

			err := output.RawFlagValidator()

			Expect(err).To(BeNil())
		})

		It("should allow table flag alone", func() {
			output := &Output{}
			output.ofTable = true

			err := output.RawFlagValidator()

			Expect(err).To(BeNil())
		})

		It("should allow empty output", func() {
			output := &Output{}

			err := output.RawFlagValidator()

			Expect(err).To(BeNil())
		})
	})

	Context("Invalid flag combinations", func() {
		It("should reject raw flag alone", func() {
			output := &Output{}
			output.ofRaw = true

			err := output.RawFlagValidator()

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal(errMsgSoleRaw))
		})

		It("should reject raw flag with table only", func() {
			output := &Output{}
			output.ofRaw = true
			output.ofTable = true

			err := output.RawFlagValidator()

			Expect(err).NotTo(BeNil())
			Expect(err.Error()).To(Equal(errMsgSoleRaw))
		})
	})
})
