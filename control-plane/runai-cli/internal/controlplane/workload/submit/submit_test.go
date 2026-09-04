/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package submit

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SubmitService", func() {

	Context("determineNameAndPrefix", func() {
		It("should prioritize name-prefix flag", func() {
			name, usePrefix := determineNameAndPrefix("cli-name", "yaml-name", "prefix-name", "Pod")
			Expect(name).To(Equal("prefix-name"))
			Expect(usePrefix).To(BeTrue())
		})

		It("should use CLI argument when no name-prefix", func() {
			name, usePrefix := determineNameAndPrefix("cli-name", "yaml-name", "", "Pod")
			Expect(name).To(Equal("cli-name"))
			Expect(usePrefix).To(BeFalse())
		})

		It("should use YAML name when no CLI argument or name-prefix", func() {
			name, usePrefix := determineNameAndPrefix("", "yaml-name", "", "Pod")
			Expect(name).To(Equal(""))
			Expect(usePrefix).To(BeFalse())
		})

		It("should default to Kind with prefix when all are empty", func() {
			name, usePrefix := determineNameAndPrefix("", "", "", "PyTorchJob")
			Expect(name).To(Equal("PyTorchJob"))
			Expect(usePrefix).To(BeTrue())
		})

		It("should ignore YAML name when CLI argument is provided", func() {
			name, usePrefix := determineNameAndPrefix("cli-name", "yaml-name", "", "Pod")
			Expect(name).To(Equal("cli-name"))
			Expect(usePrefix).To(BeFalse())
		})

		It("should use name-prefix even when CLI argument and YAML name exist", func() {
			name, usePrefix := determineNameAndPrefix("cli-name", "yaml-name", "prefix-name", "Pod")
			Expect(name).To(Equal("prefix-name"))
			Expect(usePrefix).To(BeTrue())
		})
	})

	Context("resolveProject", func() {
		When("project flag is provided", func() {
			It("should return the project flag value", func() {
				project, err := resolveProject("my-project", "", "default")
				Expect(err).To(BeNil())
				Expect(project).To(Equal("my-project"))
			})

			It("should prioritize project flag over YAML namespace", func() {
				project, err := resolveProject("my-project", "yaml-namespace", "")
				Expect(err).To(BeNil())
				Expect(project).To(Equal("my-project"))
			})
		})

		When("YAML namespace is provided without project flag", func() {
			It("should return empty string to delegate to API", func() {
				project, err := resolveProject("", "yaml-namespace", "")
				Expect(err).To(BeNil())
				Expect(project).To(Equal(""))
			})
		})

		When("no project is specified anywhere", func() {
			It("should return an error with guidance when no default exists", func() {
				// With an empty instance, no default project is configured
				project, err := resolveProject("", "", "")
				Expect(err).NotTo(BeNil())
				Expect(project).To(Equal(""))
				Expect(err.Error()).To(ContainSubstring("project is missing"))
				Expect(err.Error()).To(ContainSubstring("runai project set"))
			})
		})
	})
})
