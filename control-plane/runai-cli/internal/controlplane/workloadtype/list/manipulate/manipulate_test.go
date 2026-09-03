/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ParseWorkloadTypeColumns", func() {
	Context("when columns slice is empty", func() {
		It("should return empty slice", func() {
			result, err := ParseWorkloadTypeColumns([]string{})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(0))
		})
	})

	Context("when columns contain valid values", func() {
		It("should parse single valid column", func() {
			result, err := ParseWorkloadTypeColumns([]string{"name"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(1))
			Expect(result[0]).To(Equal(ColumnKeyName))
		})

		It("should parse multiple valid columns", func() {
			result, err := ParseWorkloadTypeColumns([]string{"name", "category", "priority"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyCategory, ColumnKeyPriority))
		})

		It("should handle case insensitive input", func() {
			result, err := ParseWorkloadTypeColumns([]string{"NAME", "Category", "PRIORITY"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyCategory, ColumnKeyPriority))
		})

		It("should trim whitespace", func() {
			result, err := ParseWorkloadTypeColumns([]string{"  name  ", " category ", "priority"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyCategory, ColumnKeyPriority))
		})

		It("should parse all valid columns", func() {
			allColumnNames := []string{"name", "category", "priority", "preemptibility", "group", "source"}
			result, err := ParseWorkloadTypeColumns(allColumnNames)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(6))
		})
	})

	Context("when columns contain invalid values", func() {
		It("should return error for single invalid column", func() {
			result, err := ParseWorkloadTypeColumns([]string{"invalid"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid column name"))
			Expect(result).To(BeNil())
		})

		It("should return error for multiple invalid columns", func() {
			result, err := ParseWorkloadTypeColumns([]string{"invalid1", "invalid2"})
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("should return error when mixing valid and invalid columns", func() {
			result, err := ParseWorkloadTypeColumns([]string{"name", "invalid", "priority"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid column name"))
			Expect(result).To(BeNil())
		})
	})

	Context("when columns contain duplicates", func() {
		It("should include duplicates in result", func() {
			result, err := ParseWorkloadTypeColumns([]string{"name", "name", "category"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(3))
			Expect(result[0]).To(Equal(ColumnKeyName))
			Expect(result[1]).To(Equal(ColumnKeyName))
			Expect(result[2]).To(Equal(ColumnKeyCategory))
		})
	})
})

var _ = Describe("CreateAlignedWorkloadTypes", func() {
	var (
		workloadTypesRaw Data
		columns          []WorkloadTypeColumn
	)

	BeforeEach(func() {
		wt1 := newTestWorkloadTypeConfig("interactive", "Interactive", "Medium", openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE)
		wt2 := newTestWorkloadTypeConfig("train", "Train", "High", openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE)
		workloadTypesRaw = Data{
			WorkloadTypes: []openapi.WorkloadTypeConfig{wt1, wt2},
		}
		columns = []WorkloadTypeColumn{ColumnKeyName, ColumnKeyCategory, ColumnKeyPriority}
	})

	It("should create aligned workload types with specified columns", func() {
		result := CreateAlignedWorkloadTypes(workloadTypesRaw, columns)

		Expect(result.WorkloadTypes).To(HaveLen(2))
		Expect(result.WorkloadTypes[0].Name).To(Equal("interactive"))
		Expect(result.WorkloadTypes[0].CategoryName).To(Equal("Interactive"))
		Expect(result.WorkloadTypes[0].PriorityName).To(Equal("Medium"))

		Expect(result.WorkloadTypes[1].Name).To(Equal("train"))
		Expect(result.WorkloadTypes[1].CategoryName).To(Equal("Train"))
		Expect(result.WorkloadTypes[1].PriorityName).To(Equal("High"))
	})

	It("should only include selected columns", func() {
		columns = []WorkloadTypeColumn{ColumnKeyName}
		result := CreateAlignedWorkloadTypes(workloadTypesRaw, columns)

		Expect(result.WorkloadTypes).To(HaveLen(2))
		Expect(result.WorkloadTypes[0].Name).To(Equal("interactive"))
		Expect(result.WorkloadTypes[0].CategoryName).To(BeEmpty())
		Expect(result.WorkloadTypes[0].PriorityName).To(BeEmpty())
	})

	It("should handle empty workload types list", func() {
		workloadTypesRaw.WorkloadTypes = []openapi.WorkloadTypeConfig{}
		result := CreateAlignedWorkloadTypes(workloadTypesRaw, columns)

		Expect(result.WorkloadTypes).To(HaveLen(0))
	})
})

var _ = Describe("CreateTableData", func() {
	var (
		alignedWorkloadTypes []WorkloadTypeAlignedRow
		columns              []WorkloadTypeColumn
	)

	BeforeEach(func() {
		alignedWorkloadTypes = []WorkloadTypeAlignedRow{
			{
				Name:           "interactive",
				CategoryName:   "Interactive",
				PriorityName:   "Medium",
				Preemptibility: "Preemptible",
				Group:          "default",
			},
			{
				Name:           "train",
				CategoryName:   "Train",
				PriorityName:   "High",
				Preemptibility: "NonPreemptible",
				Group:          "default",
			},
		}
		columns = []WorkloadTypeColumn{ColumnKeyName, ColumnKeyCategory, ColumnKeyPriority}
	})

	It("should create table data in correct format", func() {
		result := CreateTableData(alignedWorkloadTypes, columns)

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(Equal([]string{"interactive", "Interactive", "Medium"}))
		Expect(result[1]).To(Equal([]string{"train", "Train", "High"}))
	})

	It("should handle different column order", func() {
		columns = []WorkloadTypeColumn{ColumnKeyPriority, ColumnKeyName, ColumnKeyCategory}
		result := CreateTableData(alignedWorkloadTypes, columns)

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(Equal([]string{"High", "train", "Train"}))
		Expect(result[1]).To(Equal([]string{"Medium", "interactive", "Interactive"}))
	})

	It("should handle single column", func() {
		columns = []WorkloadTypeColumn{ColumnKeyName}
		result := CreateTableData(alignedWorkloadTypes, columns)

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(Equal([]string{"interactive"}))
		Expect(result[1]).To(Equal([]string{"train"}))
	})

	It("should handle empty aligned workload types", func() {
		alignedWorkloadTypes = []WorkloadTypeAlignedRow{}
		result := CreateTableData(alignedWorkloadTypes, columns)

		Expect(result).To(HaveLen(0))
	})

	It("should sort by name ascending", func() {
		alignedWorkloadTypes = []WorkloadTypeAlignedRow{
			{Name: "zebra", CategoryName: "Cat1", PriorityName: "High"},
			{Name: "apple", CategoryName: "Cat2", PriorityName: "Low"},
			{Name: "banana", CategoryName: "Cat3", PriorityName: "Medium"},
		}
		columns = []WorkloadTypeColumn{ColumnKeyName, ColumnKeyCategory}
		result := CreateTableData(alignedWorkloadTypes, columns)

		Expect(result).To(HaveLen(3))
		Expect(result[0][0]).To(Equal("apple"))
		Expect(result[1][0]).To(Equal("banana"))
		Expect(result[2][0]).To(Equal("zebra"))
	})
})

var _ = Describe("GetDefaultColumns", func() {
	It("should return all expected default columns", func() {
		result := GetDefaultColumns()

		expectedColumns := []WorkloadTypeColumn{
			ColumnKeyName, ColumnKeyCategory, ColumnKeyPriority,
			ColumnKeyPreemptibility, ColumnKeyGroup,
		}
		Expect(result).To(Equal(expectedColumns))
	})
})

var _ = Describe("GetColumns", func() {
	It("should return provided columns when not empty", func() {
		customColumns := []WorkloadTypeColumn{ColumnKeyName, ColumnKeyCategory}
		result := GetColumns(customColumns)

		Expect(result).To(Equal(customColumns))
	})

	It("should return default columns when empty", func() {
		result := GetColumns([]WorkloadTypeColumn{})

		expectedColumns := GetDefaultColumns()
		Expect(result).To(Equal(expectedColumns))
	})

	It("should return default columns when nil", func() {
		result := GetColumns(nil)

		expectedColumns := GetDefaultColumns()
		Expect(result).To(Equal(expectedColumns))
	})
})

var _ = Describe("BriefListFromResponse", func() {
	It("should convert workload type configs to brief list", func() {
		wt1 := newTestWorkloadTypeConfig("interactive", "Interactive", "Medium", openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE)
		wt2 := newTestWorkloadTypeConfig("train", "Train", "High", openapi.WORKLOADTYPEPREEMPTIBILITY_NON_PREEMPTIBLE)

		result, err := BriefListFromResponse([]openapi.WorkloadTypeConfig{wt1, wt2})

		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(HaveLen(2))
		Expect(result[0].Name).To(Equal("interactive"))
		Expect(result[0].CategoryName).To(Equal("Interactive"))
		Expect(result[0].PriorityName).To(Equal("Medium"))
		Expect(result[0].Preemptibility).To(Equal("preemptible"))

		Expect(result[1].Name).To(Equal("train"))
		Expect(result[1].CategoryName).To(Equal("Train"))
		Expect(result[1].PriorityName).To(Equal("High"))
		Expect(result[1].Preemptibility).To(Equal("non-preemptible"))
	})

	It("should handle empty list", func() {
		result, err := BriefListFromResponse([]openapi.WorkloadTypeConfig{})

		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(HaveLen(0))
	})

	It("should handle workload types with source", func() {
		source := "external"
		wt := newTestWorkloadTypeConfig("external-wt", "Interactive", "Low", openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE)
		wt.Source = &source

		result, err := BriefListFromResponse([]openapi.WorkloadTypeConfig{wt})

		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(HaveLen(1))
		Expect(result[0].Source).To(Equal("external"))
	})

	It("should handle workload types without optional fields", func() {
		wt := openapi.WorkloadTypeConfig{
			Name:           "minimal",
			Preemptibility: openapi.WORKLOADTYPEPREEMPTIBILITY_PREEMPTIBLE,
			Group:          "default",
			CategoryId:     "cat-id",
			PriorityId:     "pri-id",
		}

		result, err := BriefListFromResponse([]openapi.WorkloadTypeConfig{wt})

		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(HaveLen(1))
		Expect(result[0].Name).To(Equal("minimal"))
		Expect(result[0].CategoryName).To(BeEmpty())
		Expect(result[0].PriorityName).To(BeEmpty())
		Expect(result[0].Source).To(BeEmpty())
	})
})

// Helper function to create test workload type configs
func newTestWorkloadTypeConfig(name string, category string, priority string, preemptibility openapi.WorkloadTypePreemptibility) openapi.WorkloadTypeConfig {
	return openapi.WorkloadTypeConfig{
		Name:           name,
		CategoryName:   &category,
		PriorityName:   &priority,
		Preemptibility: preemptibility,
		Group:          "default",
		CategoryId:     "cat-id",
		PriorityId:     "pri-id",
	}
}
