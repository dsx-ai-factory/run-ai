/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/mycredential/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/mycredential/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/mycredential/list/manipulate"
)

func newTestEntry(credType, name, description string) client.MyCredentialEntry {
	return client.MyCredentialEntry{Type: credType, Name: name, Description: description}
}

var _ = Describe("Output", func() {
	var output *list.Output

	BeforeEach(func() {
		entries := []client.MyCredentialEntry{
			newTestEntry("dockerRegistry", "docker-hub", "Docker Hub registry"),
			newTestEntry("genericSecret", "aws-secret", "AWS credentials"),
			newTestEntry("ngcApiKey", "ngc-key", ""),
		}
		output = list.NewOutput(entries)
	})

	Context("Subject", func() {
		It("should return correct subject", func() {
			Expect(output.Subject()).To(Equal("my-credential"))
		})
	})

	Context("Columns", func() {
		It("should return Type plus default columns", func() {
			columns := output.Columns()
			defaults := manipulate.GetDefaultColumns()
			Expect(columns).To(HaveLen(len(defaults)))
			Expect(columns[0].Title).To(Equal("Type"))
			Expect(columns[1].Title).To(Equal("Name"))
			Expect(columns[2].Title).To(Equal("Description"))
		})

		It("should return columns with correct titles and widths", func() {
			columns := output.Columns()
			Expect(columns[0].Title).To(Equal("Type"))
			Expect(columns[0].Width).To(Equal(16))
			Expect(columns[1].Title).To(Equal("Name"))
			Expect(columns[1].Width).To(Equal(manipulate.DefaultWidth))
			Expect(columns[2].Title).To(Equal("Description"))
			Expect(columns[2].Width).To(Equal(manipulate.DefaultWidth))
		})
	})

	Context("Data", func() {
		It("should return rows with type, name, and description", func() {
			data := output.Data()
			Expect(data).To(HaveLen(3))
			Expect(data[0][0]).To(Equal("dockerRegistry"))
			Expect(data[0][1]).To(Equal("docker-hub"))
			Expect(data[1][0]).To(Equal("genericSecret"))
			Expect(data[1][1]).To(Equal("aws-secret"))
			Expect(data[2][0]).To(Equal("ngcApiKey"))
			Expect(data[2][1]).To(Equal("ngc-key"))
		})

		It("should preserve description alongside type and name", func() {
			data := output.Data()
			Expect(data[0]).To(Equal([]string{"dockerRegistry", "docker-hub", "Docker Hub registry"}))
			Expect(data[1]).To(Equal([]string{"genericSecret", "aws-secret", "AWS credentials"}))
			Expect(data[2]).To(Equal([]string{"ngcApiKey", "ngc-key", ""}))
		})

		It("should handle empty entries", func() {
			output := list.NewOutput([]client.MyCredentialEntry{})
			Expect(output.Data()).To(BeEmpty())
		})

		It("should handle nil entries", func() {
			output := list.NewOutput(nil)
			Expect(output.Data()).To(BeEmpty())
		})

		It("should handle single entry", func() {
			output := list.NewOutput(
				[]client.MyCredentialEntry{newTestEntry("genericSecret", "only-one", "single credential")},
			)
			data := output.Data()
			Expect(data).To(HaveLen(1))
			Expect(data[0]).To(Equal([]string{"genericSecret", "only-one", "single credential"}))
		})
	})

	Context("GenerateTable", func() {
		It("should generate table containing all credential data", func() {
			table, err := output.GenerateTable()
			Expect(err).ToNot(HaveOccurred())
			Expect(table).To(ContainSubstring("aws-secret"))
			Expect(table).To(ContainSubstring("docker-hub"))
			Expect(table).To(ContainSubstring("ngc-key"))
			Expect(table).To(ContainSubstring("AWS credentials"))
			Expect(table).To(ContainSubstring("Docker Hub registry"))
		})

		It("should include column headers in output", func() {
			table, err := output.GenerateTable()
			Expect(err).ToNot(HaveOccurred())
			Expect(table).To(ContainSubstring("Type"))
			Expect(table).To(ContainSubstring("Name"))
			Expect(table).To(ContainSubstring("Description"))
		})

		It("should not error on empty entries", func() {
			output := list.NewOutput([]client.MyCredentialEntry{})
			table, err := output.GenerateTable()
			Expect(err).ToNot(HaveOccurred())
			Expect(table).ToNot(BeEmpty())
		})
	})
})
