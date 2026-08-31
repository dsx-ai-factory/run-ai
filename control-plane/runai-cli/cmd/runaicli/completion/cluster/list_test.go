/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package cluster

import (
	"github.com/run-ai/runai/control-plane/runai-cli/internal/cli/completion"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/cluster/list/manipulate"
	"github.com/spf13/cobra"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ColumnsFlagCompletion", func() {
	var cmd *cobra.Command

	BeforeEach(func() {
		cmd = &cobra.Command{}
	})

	Context("when toComplete is empty", func() {
		It("should return all available columns", func() {
			completions, directive := ColumnsFlagCompletion(cmd, []string{}, "")

			Expect(directive).To(Equal(cobra.ShellCompDirectiveNoSpace))
			Expect(completions).To(HaveLen(len(manipulate.ColumnMap)))
			// Verify it contains expected columns
			Expect(completions).To(ContainElements("Name", "URL", "Version", "Status", "KubernetesDistribution", "DistributionVersion"))
		})
	})

	Context("when toComplete has no comma", func() {
		It("should return all columns for empty input", func() {
			completions, directive := ColumnsFlagCompletion(cmd, []string{}, "")

			Expect(directive).To(Equal(cobra.ShellCompDirectiveNoSpace))
			Expect(completions).To(HaveLen(len(manipulate.ColumnMap)))
		})

		It("should filter columns by prefix", func() {
			completions, directive := ColumnsFlagCompletion(cmd, []string{}, "na")

			Expect(directive).To(Equal(cobra.ShellCompDirectiveNoSpace))
			Expect(completions).To(ContainElement("Name"))
			Expect(completions).NotTo(ContainElement("URL"))
			Expect(completions).NotTo(ContainElement("Status"))
		})

		It("should filter columns by prefix case insensitive", func() {
			completions, directive := ColumnsFlagCompletion(cmd, []string{}, "ur")

			Expect(directive).To(Equal(cobra.ShellCompDirectiveNoSpace))
			Expect(completions).To(ContainElement("URL"))
			Expect(completions).NotTo(ContainElement("Name"))
		})

		It("should return empty for non-matching prefix", func() {
			completions, directive := ColumnsFlagCompletion(cmd, []string{}, "xyz")

			Expect(directive).To(Equal(cobra.ShellCompDirectiveNoSpace))
			Expect(completions).To(BeEmpty())
		})
	})

	Context("when toComplete has comma-separated values", func() {
		It("should exclude already selected columns", func() {
			completions, directive := ColumnsFlagCompletion(cmd, []string{}, "name,")

			Expect(directive).To(Equal(cobra.ShellCompDirectiveNoSpace))
			Expect(completions).NotTo(ContainElement("name,Name"))
			Expect(completions).To(ContainElement("name,URL"))
			Expect(completions).To(ContainElement("name,Status"))
		})

		It("should handle multiple selected columns", func() {
			completions, directive := ColumnsFlagCompletion(cmd, []string{}, "name,url,")

			Expect(directive).To(Equal(cobra.ShellCompDirectiveNoSpace))
			Expect(completions).NotTo(ContainElement("name,url,Name"))
			Expect(completions).NotTo(ContainElement("name,url,URL"))
			Expect(completions).To(ContainElement("name,url,Status"))
			Expect(completions).To(ContainElement("name,url,Version"))
		})

		It("should filter remaining columns by prefix", func() {
			completions, directive := ColumnsFlagCompletion(cmd, []string{}, "name,st")

			Expect(directive).To(Equal(cobra.ShellCompDirectiveNoSpace))
			Expect(completions).To(ContainElement("name,Status"))
			Expect(completions).NotTo(ContainElement("name,URL"))
			Expect(completions).NotTo(ContainElement("name,Version"))
		})

		It("should handle case insensitive selected columns", func() {
			completions, directive := ColumnsFlagCompletion(cmd, []string{}, "Name,URL,")

			Expect(directive).To(Equal(cobra.ShellCompDirectiveNoSpace))
			Expect(completions).NotTo(ContainElement("Name,URL,Name"))
			Expect(completions).NotTo(ContainElement("Name,URL,URL"))
			Expect(completions).To(ContainElement("Name,URL,Status"))
		})

		It("should handle whitespace in column names", func() {
			completions, directive := ColumnsFlagCompletion(cmd, []string{}, " name , url ,")

			Expect(directive).To(Equal(cobra.ShellCompDirectiveNoSpace))
			Expect(completions).NotTo(ContainElement(" name , url ,Name"))
			Expect(completions).NotTo(ContainElement(" name , url ,URL"))
			Expect(completions).To(ContainElement(" name , url ,Status"))
		})
	})

	Context("when all columns are selected", func() {
		It("should return empty completions", func() {
			allColumns := "name,url,version,status,kubernetesdistribution,distributionversion,"
			completions, directive := ColumnsFlagCompletion(cmd, []string{}, allColumns)

			Expect(directive).To(Equal(cobra.ShellCompDirectiveNoSpace))
			Expect(completions).To(BeEmpty())
		})
	})
})

var _ = Describe("Helper Functions", func() {
	Context("parseColumnsForCompletion", func() {
		It("should parse empty string", func() {
			current, previous, split := completion.ParseColumnsForCompletion("")

			Expect(current).To(Equal(""))
			Expect(previous).To(Equal(""))
			Expect(split).To(Equal([]string{""}))
		})

		It("should parse single column", func() {
			current, previous, split := completion.ParseColumnsForCompletion("name")

			Expect(current).To(Equal("name"))
			Expect(previous).To(Equal(""))
			Expect(split).To(Equal([]string{"name"}))
		})

		It("should parse multiple columns", func() {
			current, previous, split := completion.ParseColumnsForCompletion("name,url,st")

			Expect(current).To(Equal("st"))
			Expect(previous).To(Equal("name,url,"))
			Expect(split).To(Equal([]string{"name", "url", "st"}))
		})

		It("should parse trailing comma", func() {
			current, previous, split := completion.ParseColumnsForCompletion("name,url,")

			Expect(current).To(Equal(""))
			Expect(previous).To(Equal("name,url,"))
			Expect(split).To(Equal([]string{"name", "url", ""}))
		})
	})

	Context("createSelectedColumnsMap", func() {
		It("should create map from column list", func() {
			columns := []string{"name", "URL", "Status"}
			result := completion.CreateSelectedColumnsMap(columns)

			Expect(result).To(HaveKey("name"))
			Expect(result).To(HaveKey("url"))    // lowercase
			Expect(result).To(HaveKey("status")) // lowercase
			Expect(result["name"]).To(BeTrue())
			Expect(result["url"]).To(BeTrue())
			Expect(result["status"]).To(BeTrue())
		})

		It("should ignore empty strings", func() {
			columns := []string{"name", "", "status"}
			result := completion.CreateSelectedColumnsMap(columns)

			Expect(result).To(HaveKey("name"))
			Expect(result).To(HaveKey("status"))
			Expect(result).NotTo(HaveKey(""))
			Expect(len(result)).To(Equal(2))
		})
	})
})
