/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
)

var _ = Describe("ParseProjectColumns function", func() {
	Context("when columns slice is empty", func() {
		It("should return nil", func() {
			result, err := ParseProjectColumns([]string{})
			Expect(err).To(BeNil())
			Expect(result).To(BeNil())
		})

		It("should return nil for nil input", func() {
			result, err := ParseProjectColumns(nil)
			Expect(err).To(BeNil())
			Expect(result).To(BeNil())
		})
	})

	Context("when columns contain valid values", func() {
		It("should parse single valid column", func() {
			result, err := ParseProjectColumns([]string{"name"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))
			Expect(result[0]).To(Equal(ColumnKeyName))
		})

		It("should parse multiple valid columns", func() {
			result, err := ParseProjectColumns([]string{"name", "status", "department"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyStatus, ColumnKeyDepartment))
		})

		It("should handle case insensitive input", func() {
			result, err := ParseProjectColumns([]string{"NAME", "Status", "DEPARTMENT"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyStatus, ColumnKeyDepartment))
		})

		It("should trim whitespace", func() {
			result, err := ParseProjectColumns([]string{"  name  ", " status ", "department"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyStatus, ColumnKeyDepartment))
		})

		It("should parse all available columns", func() {
			allColumnNames := []string{
				"name", "status", "department", "gpuquota", "allocatedgpus", "allocationratio",
			}
			result, err := ParseProjectColumns(allColumnNames)
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(len(allColumnNames)))
		})
	})

	Context("when columns contain invalid values", func() {
		It("should return error for single invalid column", func() {
			result, err := ParseProjectColumns([]string{"invalid"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid column names: \"invalid\""))
			Expect(result).To(BeNil())
		})

		It("should return error for multiple invalid columns", func() {
			result, err := ParseProjectColumns([]string{"invalid1", "", "invalid2"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid column names: \"invalid1\", \"\", \"invalid2\""))
			Expect(result).To(BeNil())
		})

		It("should return error when mixing valid and invalid columns", func() {
			result, err := ParseProjectColumns([]string{"name", "invalid", "status"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid column names: \"invalid\""))
			Expect(result).To(BeNil())
		})
	})

	Context("when columns contain duplicates", func() {
		It("should include duplicates in result", func() {
			result, err := ParseProjectColumns([]string{"name", "name", "status"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))
			Expect(result[0]).To(Equal(ColumnKeyName))
			Expect(result[1]).To(Equal(ColumnKeyName))
			Expect(result[2]).To(Equal(ColumnKeyStatus))
		})
	})
})

var _ = Describe("Extract Functions", func() {
	var (
		project    *openapi.Project
		projectRow *ProjectsAlignedRow
	)

	BeforeEach(func() {
		project = &openapi.Project{}
		projectRow = &ProjectsAlignedRow{}
	})

	Context("extractNameData", func() {
		It("should extract project name", func() {
			project.Name = "test-project"

			extractNameData(project, projectRow)

			Expect(projectRow.Name).To(Equal("test-project"))
		})
	})

	Context("extractStatusData", func() {
		It("should extract project status", func() {
			phase := openapi.PROJECTPHASE_CREATING
			project.Status = openapi.ProjectStatus{
				Phase: &phase,
			}

			extractStatusData(project, projectRow)

			Expect(projectRow.Status).To(Equal(string(openapi.PROJECTPHASE_CREATING)))
		})
	})

	Context("extractDepartmentData", func() {
		It("should extract department name when parent is set", func() {
			parent := openapi.DepartmentProjectParent{Name: "Engineering"}
			project.Parent = *openapi.NewNullableDepartmentProjectParent(&parent)

			extractDepartmentData(project, projectRow)

			Expect(projectRow.Department).To(Equal("Engineering"))
		})

		It("should leave department empty when parent is not set", func() {
			project.Parent = openapi.NullableDepartmentProjectParent{}

			extractDepartmentData(project, projectRow)

			Expect(projectRow.Department).To(Equal(""))
		})
	})

	Context("extractGPUQuotaData", func() {
		It("should sum GPU quota from resources", func() {
			project.Resources = []openapi.Resources{
				{Gpu: openapi.ResourcesGpu{Deserved: 2.5}},
				{Gpu: openapi.ResourcesGpu{Deserved: 1.5}},
			}

			extractGPUQuotaData(project, projectRow)

			Expect(projectRow.GPUQuota).To(Equal(4.0))
		})

		It("should return 0 when no resources", func() {
			project.Resources = []openapi.Resources{}

			extractGPUQuotaData(project, projectRow)

			Expect(projectRow.GPUQuota).To(Equal(0.0))
		})
	})

	Context("extractAllocatedGPUsData", func() {
		It("should sum allocated GPUs from quota statuses", func() {
			gpu1 := 2.0
			gpu2 := 1.5
			project.Status = openapi.ProjectStatus{
				NodePoolQuotaStatuses: []openapi.QuotaStatusNodePool{
					{Allocated: &openapi.QuotaStatusResource{Gpu: &gpu1}},
					{Allocated: &openapi.QuotaStatusResource{Gpu: &gpu2}},
				},
			}

			extractAllocatedGPUsData(project, projectRow)

			Expect(projectRow.AllocatedGPUs).To(Equal(3.5))
		})

		It("should handle nil GPU values", func() {
			gpu1 := 2.0
			project.Status = openapi.ProjectStatus{
				NodePoolQuotaStatuses: []openapi.QuotaStatusNodePool{
					{Allocated: &openapi.QuotaStatusResource{Gpu: &gpu1}},
					{Allocated: &openapi.QuotaStatusResource{Gpu: nil}},
				},
			}

			extractAllocatedGPUsData(project, projectRow)

			Expect(projectRow.AllocatedGPUs).To(Equal(2.0))
		})
	})

	Context("extractAllocationRatioData", func() {
		It("should calculate allocation ratio percentage", func() {
			gpu1 := 3.0
			project.Resources = []openapi.Resources{
				{Gpu: openapi.ResourcesGpu{Deserved: 4.0}},
			}
			project.Status = openapi.ProjectStatus{
				NodePoolQuotaStatuses: []openapi.QuotaStatusNodePool{
					{Allocated: &openapi.QuotaStatusResource{Gpu: &gpu1}},
				},
			}

			extractAllocationRatioData(project, projectRow)

			Expect(projectRow.AllocationRatio).To(Equal(75.0)) // 3/4 * 100
		})

		It("should return 0 when quota is 0", func() {
			gpu1 := 2.0
			project.Resources = []openapi.Resources{
				{Gpu: openapi.ResourcesGpu{Deserved: 0.0}},
			}
			project.Status = openapi.ProjectStatus{
				NodePoolQuotaStatuses: []openapi.QuotaStatusNodePool{
					{Allocated: &openapi.QuotaStatusResource{Gpu: &gpu1}},
				},
			}

			extractAllocationRatioData(project, projectRow)

			Expect(projectRow.AllocationRatio).To(Equal(0.0))
		})
	})
})

var _ = Describe("Format Functions", func() {
	var projectRow *ProjectsAlignedRow

	BeforeEach(func() {
		projectRow = &ProjectsAlignedRow{}
	})

	Context("formatNameValue", func() {
		It("should return project name", func() {
			projectRow.Name = "test-project"

			result := formatNameValue(projectRow)

			Expect(result).To(Equal("test-project"))
		})
	})

	Context("formatStatusValue", func() {
		It("should return status string", func() {
			projectRow.Status = "Active"

			result := formatStatusValue(projectRow)

			Expect(result).To(Equal("Active"))
		})
	})

	Context("formatDepartmentValue", func() {
		It("should return department name when set", func() {
			projectRow.Department = "Engineering"

			result := formatDepartmentValue(projectRow)

			Expect(result).To(Equal("Engineering"))
		})

		It("should return '-' when department is empty", func() {
			projectRow.Department = ""

			result := formatDepartmentValue(projectRow)

			Expect(result).To(Equal("-"))
		})
	})

	Context("formatGPUQuotaValue", func() {
		It("should format normal GPU quota", func() {
			projectRow.GPUQuota = 4.75

			result := formatGPUQuotaValue(projectRow)

			Expect(result).To(Equal("4.75"))
		})

		It("should return '-' for -1 quota", func() {
			projectRow.GPUQuota = -1

			result := formatGPUQuotaValue(projectRow)

			Expect(result).To(Equal("-"))
		})

		It("should format zero quota", func() {
			projectRow.GPUQuota = 0

			result := formatGPUQuotaValue(projectRow)

			Expect(result).To(Equal("0.00"))
		})
	})

	Context("formatAllocatedGPUsValue", func() {
		It("should format allocated GPUs", func() {
			projectRow.AllocatedGPUs = 2.5

			result := formatAllocatedGPUsValue(projectRow)

			Expect(result).To(Equal("2.50"))
		})

		It("should format zero allocated GPUs", func() {
			projectRow.AllocatedGPUs = 0

			result := formatAllocatedGPUsValue(projectRow)

			Expect(result).To(Equal("0.00"))
		})
	})

	Context("formatAllocationRatioValue", func() {
		It("should format allocation ratio as percentage", func() {
			projectRow.AllocationRatio = 75.5

			result := formatAllocationRatioValue(projectRow)

			Expect(result).To(Equal("75.50%"))
		})

		It("should format zero ratio", func() {
			projectRow.AllocationRatio = 0

			result := formatAllocationRatioValue(projectRow)

			Expect(result).To(Equal("0.00%"))
		})

		It("should format 100% ratio", func() {
			projectRow.AllocationRatio = 100

			result := formatAllocationRatioValue(projectRow)

			Expect(result).To(Equal("100.00%"))
		})
	})
})
