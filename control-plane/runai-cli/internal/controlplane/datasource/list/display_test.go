/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list_test

import (
	listCommon "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/datasource/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/datasource/list/manipulate"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Datasource List Display", func() {
	var (
		output      *list.Output
		datasources []openapi.DatasourceListResponseEntry
	)

	BeforeEach(func() {
		ds1 := newTestDatasource("nfs-data", openapi.ASSETKIND_NFS, "Cluster")
		ds2 := newTestDatasource("s3-models", openapi.ASSETKIND_S3, "Project")
		datasources = []openapi.DatasourceListResponseEntry{ds1, ds2}

		response := &openapi.DatasourceListResponse{
			Entries: datasources,
		}
		var err error
		output, err = list.ParseListResponse(response)
		Expect(err).ToNot(HaveOccurred())
	})

	Context("ParseListResponse", func() {
		It("should parse response successfully", func() {
			Expect(output).ToNot(BeNil())
		})

		It("should return error for nil response", func() {
			output, err := list.ParseListResponse(nil)
			Expect(err).To(HaveOccurred())
			Expect(output).To(BeNil())
		})
	})

	Context("SetRaw", func() {
		It("should set raw flag", func() {
			output.SetRaw(true)
			// No direct way to verify, but ensure no panic
		})
	})

	Context("SetColumns", func() {
		It("should set valid columns", func() {
			err := output.SetColumns([]string{"name", "type", "scope"})
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return error for invalid columns", func() {
			err := output.SetColumns([]string{"invalid"})
			Expect(err).To(HaveOccurred())
		})
	})

	Context("Subject", func() {
		It("should return correct subject", func() {
			subject := output.Subject()
			Expect(subject).To(Equal("datasource"))
		})
	})

	Context("Columns", func() {
		It("should return table columns", func() {
			err := output.SetColumns([]string{"name", "type"})
			Expect(err).ToNot(HaveOccurred())

			columns := output.Columns()
			Expect(columns).To(HaveLen(2))
			Expect(columns[0].Title).To(Equal("Name"))
			Expect(columns[1].Title).To(Equal("Type"))
		})
	})

	Context("Data", func() {
		It("should return table data", func() {
			err := output.SetColumns([]string{"name", "type", "scope"})
			Expect(err).ToNot(HaveOccurred())

			data := output.Data()
			Expect(data).To(HaveLen(2))
			Expect(data[0][0]).To(Equal("nfs-data"))
			Expect(data[0][1]).To(Equal("Nfs"))
			Expect(data[0][2]).To(Equal("Cluster"))

			Expect(data[1][0]).To(Equal("s3-models"))
			Expect(data[1][1]).To(Equal("S3"))
			Expect(data[1][2]).To(Equal("Project"))
		})
	})

	Context("GenerateJson", func() {
		It("should generate JSON output", func() {
			err := output.SetColumns([]string{"name", "type"})
			Expect(err).ToNot(HaveOccurred())

			json, err := output.GenerateJson()
			Expect(err).ToNot(HaveOccurred())
			Expect(json).To(ContainSubstring("nfs-data"))
			Expect(json).To(ContainSubstring("s3-models"))
		})

		It("should generate raw JSON output", func() {
			output.SetRaw(true)
			json, err := output.GenerateJson()
			Expect(err).ToNot(HaveOccurred())
			Expect(json).To(ContainSubstring("nfs-data"))
		})
	})

	Context("GenerateYaml", func() {
		It("should generate YAML output", func() {
			err := output.SetColumns([]string{"name", "type"})
			Expect(err).ToNot(HaveOccurred())

			yaml, err := output.GenerateYaml()
			Expect(err).ToNot(HaveOccurred())
			Expect(yaml).To(ContainSubstring("nfs-data"))
			Expect(yaml).To(ContainSubstring("s3-models"))
		})

		It("should generate raw YAML output", func() {
			output.SetRaw(true)
			yaml, err := output.GenerateYaml()
			Expect(err).ToNot(HaveOccurred())
			Expect(yaml).To(ContainSubstring("nfs-data"))
		})
	})

	Context("HandleColumnsFlags", func() {
		It("should return nil when set-default-columns is false", func() {
			mockOutput := &listCommon.Output{}
			err := list.HandleColumnsFlags(*mockOutput, nil, "test")
			Expect(err).ToNot(HaveOccurred())
		})
	})
})

var _ = Describe("Default Columns", func() {
	It("should have correct default columns", func() {
		defaults := manipulate.GetDefaultColumns()
		Expect(defaults).To(HaveLen(3))
		Expect(defaults).To(ContainElement(manipulate.ColumnKeyName))
		Expect(defaults).To(ContainElement(manipulate.ColumnKeyType))
		Expect(defaults).To(ContainElement(manipulate.ColumnKeyScope))
	})
})
