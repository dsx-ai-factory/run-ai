/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	wcgenerated "github.com/run-ai/runai/runai-clients/workloads-client/generated"
)

var _ = Describe("ParseTemplateColumns", func() {
	Context("when columns slice is empty", func() {
		It("should return empty slice", func() {
			result, err := ParseTemplateColumns([]string{})
			Expect(err).To(BeNil())
			Expect(result).To(BeEmpty())
		})

		It("should return empty slice for nil input", func() {
			result, err := ParseTemplateColumns(nil)
			Expect(err).To(BeNil())
			Expect(result).To(BeEmpty())
		})
	})

	Context("when columns contain valid values", func() {
		It("should parse single valid column", func() {
			result, err := ParseTemplateColumns([]string{"name"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))
			Expect(result[0]).To(Equal(ColumnKeyName))
		})

		It("should parse multiple valid columns", func() {
			result, err := ParseTemplateColumns([]string{"name", "image", "scopetype"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyImage, ColumnKeyScopeType))
		})

		It("should handle case insensitive input", func() {
			result, err := ParseTemplateColumns([]string{"NAME", "Image", "SCOPETYPE"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyImage, ColumnKeyScopeType))
		})

		It("should trim whitespace", func() {
			result, err := ParseTemplateColumns([]string{"  name  ", " image ", "scopetype"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyImage, ColumnKeyScopeType))
		})

		It("should parse all available columns", func() {
			allColumnNames := []string{
				"name", "scopetype", "image", "gpucomputerequest", "workloadtype", "mandatoryfieldsmissing", "createdat",
			}
			result, err := ParseTemplateColumns(allColumnNames)
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(len(allColumnNames)))
			Expect(result).To(ContainElements(
				ColumnKeyName, ColumnKeyScopeType, ColumnKeyImage, ColumnKeyGpuComputeRequest,
				ColumnKeyWorkloadType, ColumnKeyMandatoryFieldsMissing, ColumnKeyCreatedAt,
			))
		})
	})

	Context("when columns contain invalid values", func() {
		It("should return error for single invalid column", func() {
			result, err := ParseTemplateColumns([]string{"invalid"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unknown column: invalid"))
			Expect(result).To(BeNil())
		})

		It("should return error for multiple invalid columns", func() {
			result, err := ParseTemplateColumns([]string{"invalid1", "invalid2"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unknown column: invalid1"))
			Expect(result).To(BeNil())
		})

		It("should return error when mixing valid and invalid columns", func() {
			result, err := ParseTemplateColumns([]string{"name", "invalid", "image"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unknown column: invalid"))
			Expect(result).To(BeNil())
		})

		It("should return error for empty string column", func() {
			result, err := ParseTemplateColumns([]string{""})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unknown column:"))
			Expect(result).To(BeNil())
		})
	})

	Context("when columns contain duplicates", func() {
		It("should include duplicates in result", func() {
			result, err := ParseTemplateColumns([]string{"name", "name", "image"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))
			Expect(result[0]).To(Equal(ColumnKeyName))
			Expect(result[1]).To(Equal(ColumnKeyName))
			Expect(result[2]).To(Equal(ColumnKeyImage))
		})
	})
})

var _ = Describe("CreateAlignedTemplates", func() {
	var templatesData Data

	BeforeEach(func() {
		// Create test template data
		template1 := wcgenerated.SupersetTemplate{
			Meta: wcgenerated.TemplateMetaFields{
				Name:         "template-1",
				ScopeType:    wcgenerated.SCOPETYPE_PROJECT,
				WorkloadType: wcgenerated.WORKLOADTYPE_TRAINING,
				CreatedAt:    time.Now(),
			},
		}
		spec1 := wcgenerated.SupersetSpec{}
		spec1.SetImage("pytorch:latest")
		compute1 := wcgenerated.WorkspaceSpecSpecAllOfCompute{}
		compute1.SetGpuDevicesRequest(4)
		spec1.SetCompute(compute1)
		template1.Spec = &spec1

		runnable1 := wcgenerated.RunnableInfo{}
		runnable1.SetIsRunnable(true)
		template1.SetRunnable(runnable1)

		template2 := wcgenerated.SupersetTemplate{
			Meta: wcgenerated.TemplateMetaFields{
				Name:         "template-2",
				ScopeType:    wcgenerated.SCOPETYPE_CLUSTER,
				WorkloadType: wcgenerated.WORKLOADTYPE_WORKSPACE,
				CreatedAt:    time.Now(),
			},
		}
		spec2 := wcgenerated.SupersetSpec{}
		spec2.SetImage("tensorflow:2.0")
		compute2 := wcgenerated.WorkspaceSpecSpecAllOfCompute{}
		compute2.SetGpuDevicesRequest(2)
		spec2.SetCompute(compute2)
		template2.Spec = &spec2

		runnable2 := wcgenerated.RunnableInfo{}
		runnable2.SetIsRunnable(false)
		template2.SetRunnable(runnable2)

		templatesData = Data{
			Templates: []wcgenerated.SupersetTemplate{template1, template2},
		}
	})

	It("should create aligned templates structure", func() {
		columns := []TemplateColumn{ColumnKeyName, ColumnKeyImage, ColumnKeyScopeType}

		result := CreateAlignedTemplates(templatesData, columns)

		Expect(result.Templates).To(HaveLen(2))
		Expect(result.NextToken).To(BeNil())

		// Check first template - only requested columns should have values
		Expect(result.Templates[0].Name).To(Equal("template-1"))
		Expect(result.Templates[0].Image).To(Equal("pytorch:latest"))
		Expect(result.Templates[0].ScopeType).To(Equal("project"))
		// Other fields should be empty since not requested
		Expect(result.Templates[0].GpuComputeRequest).To(Equal(""))
		Expect(result.Templates[0].WorkloadType).To(Equal(""))
		Expect(result.Templates[0].MandatoryFieldsMissing).To(BeNil())

		// Check second template
		Expect(result.Templates[1].Name).To(Equal("template-2"))
		Expect(result.Templates[1].Image).To(Equal("tensorflow:2.0"))
		Expect(result.Templates[1].ScopeType).To(Equal("cluster"))
	})

	It("should handle all available columns", func() {
		columns := []TemplateColumn{
			ColumnKeyName, ColumnKeyScopeType, ColumnKeyImage, ColumnKeyGpuComputeRequest,
			ColumnKeyWorkloadType, ColumnKeyMandatoryFieldsMissing,
		}

		result := CreateAlignedTemplates(templatesData, columns)

		Expect(result.Templates).To(HaveLen(2))

		// Check first template - all fields should have values
		Expect(result.Templates[0].Name).To(Equal("template-1"))
		Expect(result.Templates[0].ScopeType).To(Equal("project"))
		Expect(result.Templates[0].Image).To(Equal("pytorch:latest"))
		Expect(result.Templates[0].GpuComputeRequest).To(Equal("4"))
		Expect(result.Templates[0].WorkloadType).To(Equal("Training"))
		Expect(result.Templates[0].MandatoryFieldsMissing).ToNot(BeNil())
		Expect(*result.Templates[0].MandatoryFieldsMissing).To(BeFalse())

		// Check second template
		Expect(result.Templates[1].Name).To(Equal("template-2"))
		Expect(result.Templates[1].ScopeType).To(Equal("cluster"))
		Expect(result.Templates[1].Image).To(Equal("tensorflow:2.0"))
		Expect(result.Templates[1].GpuComputeRequest).To(Equal("2"))
		Expect(result.Templates[1].WorkloadType).To(Equal("Workspace"))
		Expect(result.Templates[1].MandatoryFieldsMissing).ToNot(BeNil())
		Expect(*result.Templates[1].MandatoryFieldsMissing).To(BeTrue())
	})

	It("should handle empty templates data", func() {
		columns := []TemplateColumn{ColumnKeyName, ColumnKeyImage}
		emptyData := Data{Templates: []wcgenerated.SupersetTemplate{}}

		result := CreateAlignedTemplates(emptyData, columns)

		Expect(result.Templates).To(HaveLen(0))
		Expect(result.NextToken).To(BeNil())
	})

	It("should handle single column", func() {
		columns := []TemplateColumn{ColumnKeyName}

		result := CreateAlignedTemplates(templatesData, columns)

		Expect(result.Templates).To(HaveLen(2))
		Expect(result.Templates[0].Name).To(Equal("template-1"))
		Expect(result.Templates[1].Name).To(Equal("template-2"))
		// Other fields should be empty since they weren't requested
		Expect(result.Templates[0].Image).To(Equal(""))
		Expect(result.Templates[0].ScopeType).To(Equal(""))
		Expect(result.Templates[0].GpuComputeRequest).To(Equal(""))
		Expect(result.Templates[0].WorkloadType).To(Equal(""))
		Expect(result.Templates[0].MandatoryFieldsMissing).To(BeNil())
	})

	It("should preserve NextToken when present", func() {
		columns := []TemplateColumn{ColumnKeyName}
		nextToken := "next-token-value"
		templatesData.NextToken = &nextToken

		result := CreateAlignedTemplates(templatesData, columns)

		Expect(result.NextToken).ToNot(BeNil())
		Expect(*result.NextToken).To(Equal("next-token-value"))
	})
})

var _ = Describe("CreateTableData", func() {
	var alignedRows []TemplateAlignedRow

	BeforeEach(func() {
		mandatoryFieldsMissing1 := false
		mandatoryFieldsMissing2 := true
		alignedRows = []TemplateAlignedRow{
			{
				Name:                   "template-1",
				ScopeType:              "project",
				Image:                  "pytorch:latest",
				GpuComputeRequest:      "4",
				WorkloadType:           "Training",
				MandatoryFieldsMissing: &mandatoryFieldsMissing1,
				CreatedAt:              "2024-01-01T00:00:00Z",
			},
			{
				Name:                   "template-2",
				ScopeType:              "cluster",
				Image:                  "tensorflow:2.0",
				GpuComputeRequest:      "2",
				WorkloadType:           "Workspace",
				MandatoryFieldsMissing: &mandatoryFieldsMissing2,
				CreatedAt:              "2024-01-02T00:00:00Z",
			},
		}
	})

	It("should format data correctly for specified columns", func() {
		columns := []TemplateColumn{ColumnKeyName, ColumnKeyImage, ColumnKeyScopeType}

		result := CreateTableData(alignedRows, columns)

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(Equal([]string{"template-1", "pytorch:latest", "project"}))
		Expect(result[1]).To(Equal([]string{"template-2", "tensorflow:2.0", "cluster"}))
	})

	It("should handle empty aligned rows", func() {
		columns := []TemplateColumn{ColumnKeyName, ColumnKeyImage}

		result := CreateTableData([]TemplateAlignedRow{}, columns)

		Expect(result).To(HaveLen(0))
	})

	It("should sort by name when name column is present", func() {
		mandatoryFieldsMissing := false
		unsortedRows := []TemplateAlignedRow{
			{Name: "z-template", Image: "image-z", MandatoryFieldsMissing: &mandatoryFieldsMissing},
			{Name: "a-template", Image: "image-a", MandatoryFieldsMissing: &mandatoryFieldsMissing},
			{Name: "m-template", Image: "image-m", MandatoryFieldsMissing: &mandatoryFieldsMissing},
		}
		columns := []TemplateColumn{ColumnKeyName, ColumnKeyImage}

		result := CreateTableData(unsortedRows, columns)

		Expect(result).To(HaveLen(3))
		// Should be sorted by name alphabetically
		Expect(result[0][0]).To(Equal("a-template"))
		Expect(result[1][0]).To(Equal("m-template"))
		Expect(result[2][0]).To(Equal("z-template"))
	})

	It("should not sort when name column is not present", func() {
		mandatoryFieldsMissing := false
		unsortedRows := []TemplateAlignedRow{
			{Image: "z-image", ScopeType: "project", MandatoryFieldsMissing: &mandatoryFieldsMissing},
			{Image: "a-image", ScopeType: "cluster", MandatoryFieldsMissing: &mandatoryFieldsMissing},
		}
		columns := []TemplateColumn{ColumnKeyImage, ColumnKeyScopeType}

		result := CreateTableData(unsortedRows, columns)

		Expect(result).To(HaveLen(2))
		// Should maintain original order
		Expect(result[0]).To(Equal([]string{"z-image", "project"}))
		Expect(result[1]).To(Equal([]string{"a-image", "cluster"}))
	})

	It("should format MandatoryFieldsMissing correctly", func() {
		columns := []TemplateColumn{ColumnKeyName, ColumnKeyMandatoryFieldsMissing}

		result := CreateTableData(alignedRows, columns)

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(Equal([]string{"template-1", "No"}))
		Expect(result[1]).To(Equal([]string{"template-2", "Yes"}))
	})

	It("should handle nil MandatoryFieldsMissing", func() {
		rowWithNil := []TemplateAlignedRow{
			{
				Name:                   "template-no-runnable",
				Image:                  "image",
				MandatoryFieldsMissing: nil,
			},
		}
		columns := []TemplateColumn{ColumnKeyName, ColumnKeyMandatoryFieldsMissing}

		result := CreateTableData(rowWithNil, columns)

		Expect(result).To(HaveLen(1))
		Expect(result[0]).To(Equal([]string{"template-no-runnable", "-"}))
	})

	It("should handle empty GPU request", func() {
		rowWithNoGpu := []TemplateAlignedRow{
			{
				Name:              "template-no-gpu",
				GpuComputeRequest: "",
			},
		}
		columns := []TemplateColumn{ColumnKeyName, ColumnKeyGpuComputeRequest}

		result := CreateTableData(rowWithNoGpu, columns)

		Expect(result).To(HaveLen(1))
		Expect(result[0]).To(Equal([]string{"template-no-gpu", "0"}))
	})

	It("should handle empty image", func() {
		rowWithNoImage := []TemplateAlignedRow{
			{
				Name:  "template-no-image",
				Image: "",
			},
		}
		columns := []TemplateColumn{ColumnKeyName, ColumnKeyImage}

		result := CreateTableData(rowWithNoImage, columns)

		Expect(result).To(HaveLen(1))
		Expect(result[0]).To(Equal([]string{"template-no-image", "-"}))
	})
})

var _ = Describe("GetDefaultColumns", func() {
	It("should return all expected default columns", func() {
		result := GetDefaultColumns()

		expectedColumns := []TemplateColumn{
			ColumnKeyName,
			ColumnKeyScopeType,
			ColumnKeyImage,
			ColumnKeyGpuComputeRequest,
			ColumnKeyWorkloadType,
			ColumnKeyMandatoryFieldsMissing,
		}

		Expect(result).To(Equal(expectedColumns))
		Expect(result).To(HaveLen(6))
	})
})

var _ = Describe("formatInt32", func() {
	It("should format positive integers correctly", func() {
		result := formatInt32(42)
		Expect(result).To(Equal("42"))
	})

	It("should format zero correctly", func() {
		result := formatInt32(0)
		Expect(result).To(Equal("0"))
	})

	It("should format negative integers correctly", func() {
		result := formatInt32(-10)
		Expect(result).To(Equal("-10"))
	})

	It("should format large numbers correctly", func() {
		result := formatInt32(999999)
		Expect(result).To(Equal("999999"))
	})
})

var _ = Describe("FieldInformation", func() {
	It("should return correct column width", func() {
		field := FieldInformation{
			columnWidth: 25,
		}
		Expect(field.GetColumnWidth()).To(Equal(25))
	})

	It("should return correct column title", func() {
		field := FieldInformation{
			columnTitle: "Test Title",
		}
		Expect(field.GetColumnTitle()).To(Equal("Test Title"))
	})
})
