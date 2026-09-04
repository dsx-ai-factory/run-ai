/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package validators_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/helper/validators"
)

func TestSearchBaseCommand(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Validator Suite")
}

var _ = Describe("Validator Test", func() {
	Context("Name", func() {
		Context("Valid", func() {
			It("empty", func() {
				name := validators.NewNameValidator("")
				_, err := name.Execute()
				Expect(err).ToNot(HaveOccurred())
			})

			It("short", func() {
				name := validators.NewNameValidator("a-1")
				_, err := name.Execute()
				Expect(err).ToNot(HaveOccurred())
			})

			It("long (50) with hyphen", func() {
				name := validators.NewNameValidator("a-345678911234567892123456789312345678941234567895")
				_, err := name.Execute()
				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("Invalid", func() {
			It("end with hyphen", func() {
				name := validators.NewNameValidator("a-")
				msg, err := name.Execute()
				Expect(err).To(HaveOccurred())
				Expect(fmt.Sprintf(msg, err)).To(Equal("Error: invalid workload name. cannot end with hyphen."))
			})

			It("too long (51)", func() {
				name := validators.NewNameValidator("a-3456789112345678921234567893123456789412345678951")
				msg, err := name.Execute()
				Expect(err).To(HaveOccurred())
				Expect(fmt.Sprintf(msg, err)).To(Equal("Error: invalid workload name. exceeded max length (50)."))
			})

			It("capital letters", func() {
				name := validators.NewNameValidator("AWorkload")
				msg, err := name.Execute()
				Expect(err).To(HaveOccurred())
				Expect(fmt.Sprintf(msg, err)).To(Equal("Error: invalid workload name. string contain other characters other then lower case alphanumeric or hyphen."))
			})
		})
	})
})
