/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package credential_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/credential"
)

var _ = Describe("Credential List Command", func() {
	Describe("Command Structure", func() {
		It("should have correct aliases", func() {
			Expect(credential.List.Aliases).To(ContainElement("ls"))
		})

		It("should accept no arguments", func() {
			Expect(credential.List.Args).NotTo(BeNil())
		})

		It("should have correct use string", func() {
			Expect(credential.List.Use).To(Equal("list"))
		})
	})

	Describe("Flags", func() {
		It("should have project flag with shorthand", func() {
			flag := credential.List.Flags().Lookup("project")
			Expect(flag).NotTo(BeNil())
			Expect(flag.Shorthand).To(Equal("p"))
		})

		It("should have all flag with shorthand", func() {
			flag := credential.List.Flags().Lookup("all")
			Expect(flag).NotTo(BeNil())
			Expect(flag.Shorthand).To(Equal("A"))
		})

		It("should have type flag", func() {
			flag := credential.List.Flags().Lookup("type")
			Expect(flag).NotTo(BeNil())
		})

		It("should have type flag as optional (not required)", func() {
			flag := credential.List.Flags().Lookup("type")
			Expect(flag).NotTo(BeNil())
			annotations := flag.Annotations
			Expect(annotations).NotTo(HaveKey("cobra_annotation_bash_completion_one_required_flag"))
		})

		It("should allow parsing type flag", func() {
			err := credential.List.Flags().Parse([]string{"--type", "genericSecret"})
			Expect(err).To(BeNil())
		})

		It("should allow running without type flag", func() {
			err := credential.List.Flags().Parse([]string{"-A"})
			Expect(err).To(BeNil())
		})

		It("should allow parsing dockerRegistry type", func() {
			err := credential.List.Flags().Parse([]string{"--type", "dockerRegistry"})
			Expect(err).To(BeNil())
		})

		It("should allow parsing ngcApiKey type", func() {
			err := credential.List.Flags().Parse([]string{"--type", "ngcApiKey"})
			Expect(err).To(BeNil())
		})

		It("should allow parsing password type", func() {
			err := credential.List.Flags().Parse([]string{"--type", "password"})
			Expect(err).To(BeNil())
		})

		It("should allow parsing accessKey type", func() {
			err := credential.List.Flags().Parse([]string{"--type", "accessKey"})
			Expect(err).To(BeNil())
		})
	})
})
