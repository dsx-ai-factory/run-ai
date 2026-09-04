/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

var _ = Describe("ParseAIApplicationColumns", func() {
	Context("when columns slice is empty", func() {
		It("should return nil for empty slice", func() {
			result, err := ParseAIApplicationColumns([]string{})
			Expect(err).To(BeNil())
			Expect(result).To(BeNil())
		})

		It("should return nil for nil input", func() {
			result, err := ParseAIApplicationColumns(nil)
			Expect(err).To(BeNil())
			Expect(result).To(BeNil())
		})
	})

	Context("when columns contain valid values", func() {
		It("should parse single valid column", func() {
			result, err := ParseAIApplicationColumns([]string{"name"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))
			Expect(result[0]).To(Equal(ColumnKeyName))
		})

		It("should parse multiple valid columns", func() {
			result, err := ParseAIApplicationColumns([]string{"name", "type", "status", "project"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(4))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyType, ColumnKeyStatus, ColumnKeyProject))
		})

		It("should handle case insensitive input", func() {
			result, err := ParseAIApplicationColumns([]string{"NAME", "Type", "STATUS"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyType, ColumnKeyStatus))
		})

		It("should trim whitespace", func() {
			result, err := ParseAIApplicationColumns([]string{"  name  ", " type ", "status"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyType, ColumnKeyStatus))
		})

		It("should parse all available columns", func() {
			allColumnNames := []string{
				"name", "type", "status", "project",
				"gpucomputereq", "gpucomputealloc",
				"gpumemoryreq", "gpumemoryalloc",
				"cpucomputereq", "cpucomputealloc",
				"cpumemoryreq", "cpumemoryalloc",
			}
			result, err := ParseAIApplicationColumns(allColumnNames)
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(len(allColumnNames)))
		})
	})

	Context("when columns contain invalid values", func() {
		It("should return error for single invalid column", func() {
			result, err := ParseAIApplicationColumns([]string{"invalid"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("\"invalid\""))
			Expect(result).To(BeNil())
		})

		It("should return error for multiple invalid columns", func() {
			result, err := ParseAIApplicationColumns([]string{"invalid1", "invalid2"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("\"invalid1\""))
			Expect(err.Error()).To(ContainSubstring("\"invalid2\""))
			Expect(result).To(BeNil())
		})

		It("should return error when mixing valid and invalid columns", func() {
			result, err := ParseAIApplicationColumns([]string{"name", "invalid", "status"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("\"invalid\""))
			Expect(result).To(BeNil())
		})
	})

	Context("when columns contain duplicates", func() {
		It("should include duplicates in result", func() {
			result, err := ParseAIApplicationColumns([]string{"name", "name", "status"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))
			Expect(result[0]).To(Equal(ColumnKeyName))
			Expect(result[1]).To(Equal(ColumnKeyName))
			Expect(result[2]).To(Equal(ColumnKeyStatus))
		})
	})
})

var _ = Describe("GetColumns", func() {
	It("should return default columns when input is empty", func() {
		result := GetColumns(nil)
		Expect(result).To(Equal(defaultColumns))
	})

	It("should return provided columns when non-empty", func() {
		cols := []AIApplicationColumn{ColumnKeyName, ColumnKeyStatus}
		result := GetColumns(cols)
		Expect(result).To(Equal(cols))
	})
})

var _ = Describe("CreateAlignedApplications", func() {
	var (
		data    Data
		columns []AIApplicationColumn
	)

	BeforeEach(func() {
		data = Data{
			AIApplications: []openapi.AIApplication{
				{
					Name:        ptr.To("my-app"),
					Type:        ptr.To("inference"),
					ProjectName: ptr.To("team-alpha"),
				},
			},
			NextToken: ptr.To("next-token"),
		}
		columns = []AIApplicationColumn{ColumnKeyName, ColumnKeyType, ColumnKeyProject}
	})

	It("should create aligned output from raw data", func() {
		result := CreateAlignedApplications(data, columns)

		Expect(result.AIApplications).To(HaveLen(1))
		Expect(result.AIApplications[0].Name).To(Equal("my-app"))
		Expect(result.AIApplications[0].Type).To(Equal("inference"))
		Expect(result.AIApplications[0].Project).To(Equal("team-alpha"))
		Expect(result.NextToken).To(Equal(ptr.To("next-token")))
	})

	It("should only populate columns that are specified", func() {
		columns = []AIApplicationColumn{ColumnKeyName}

		result := CreateAlignedApplications(data, columns)

		Expect(result.AIApplications[0].Name).To(Equal("my-app"))
		Expect(result.AIApplications[0].Type).To(Equal(""))
		Expect(result.AIApplications[0].Project).To(Equal(""))
	})

	It("should handle empty applications list", func() {
		data.AIApplications = []openapi.AIApplication{}

		result := CreateAlignedApplications(data, columns)

		Expect(result.AIApplications).To(HaveLen(0))
		Expect(result.NextToken).To(Equal(ptr.To("next-token")))
	})

	It("should handle nil NextToken", func() {
		data.NextToken = nil

		result := CreateAlignedApplications(data, columns)

		Expect(result.NextToken).To(BeNil())
	})

	It("should extract resource columns for multiple apps", func() {
		gpuVal := 1.5
		alloc := openapi.AIApplicationAllocatedResources{}
		alloc.Gpu.Set(&gpuVal)

		data.AIApplications = []openapi.AIApplication{
			{Name: ptr.To("app-a"), ProjectName: ptr.To("proj-1")},
			{Name: ptr.To("app-b"), ProjectName: ptr.To("proj-2")},
		}
		data.AIApplications[0].AllocatedResources.Set(&alloc)

		columns = []AIApplicationColumn{ColumnKeyName, ColumnKeyGPUComputeAlloc}
		result := CreateAlignedApplications(data, columns)

		Expect(result.AIApplications).To(HaveLen(2))
		Expect(result.AIApplications[0].Name).To(Equal("app-a"))
		Expect(result.AIApplications[0].GPUComputeAlloc).To(Equal("1.50"))
		Expect(result.AIApplications[1].Name).To(Equal("app-b"))
		Expect(result.AIApplications[1].GPUComputeAlloc).To(Equal(""))
	})
})

var _ = Describe("CreateTableData", func() {
	var (
		rows    []AIApplicationAlignedRow
		columns []AIApplicationColumn
	)

	BeforeEach(func() {
		rows = []AIApplicationAlignedRow{
			{
				Name:    "app-beta",
				Type:    "training",
				Status:  "Running",
				Project: "proj-b",
			},
			{
				Name:    "app-alpha",
				Type:    "inference",
				Status:  "Pending",
				Project: "proj-a",
			},
		}
		columns = []AIApplicationColumn{ColumnKeyName, ColumnKeyStatus}
	})

	It("should create table data with format functions applied", func() {
		result := CreateTableData(rows, columns)

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(Equal([]string{"app-beta", "Running"}))
		Expect(result[1]).To(Equal([]string{"app-alpha", "Pending"}))
	})

	It("should apply no-data format for empty fields", func() {
		rows = []AIApplicationAlignedRow{
			{Name: "app-x"},
		}
		columns = []AIApplicationColumn{ColumnKeyName, ColumnKeyType, ColumnKeyStatus, ColumnKeyProject}

		result := CreateTableData(rows, columns)

		Expect(result).To(HaveLen(1))
		Expect(result[0]).To(Equal([]string{"app-x", "-", "-", "-"}))
	})

	It("should handle single column", func() {
		columns = []AIApplicationColumn{ColumnKeyName}

		result := CreateTableData(rows, columns)

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(Equal([]string{"app-beta"}))
		Expect(result[1]).To(Equal([]string{"app-alpha"}))
	})

	It("should handle empty rows list", func() {
		rows = []AIApplicationAlignedRow{}

		result := CreateTableData(rows, columns)

		Expect(result).To(HaveLen(0))
	})

	It("should format resource fields", func() {
		rows = []AIApplicationAlignedRow{
			{
				Name:            "app-x",
				GPUComputeAlloc: "2.00",
				GPUMemoryReq:    "4Gi",
				CPUMemoryAlloc:  "8Gi",
			},
		}
		columns = []AIApplicationColumn{
			ColumnKeyName,
			ColumnKeyGPUComputeAlloc,
			ColumnKeyGPUMemoryReq,
			ColumnKeyCPUMemoryAlloc,
		}

		result := CreateTableData(rows, columns)

		Expect(result).To(HaveLen(1))
		Expect(result[0]).To(Equal([]string{"app-x", "2.00", "4Gi", "8Gi"}))
	})
})
