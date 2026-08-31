/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openapi "github.com/run-ai/runai/runai-clients/org-unit-service-client/generated"
	"k8s.io/utils/ptr"
)

var _ = Describe("CreateAlignedProjects", func() {
	var (
		projectsRaw Data
		columns     []ProjectColumn
	)

	BeforeEach(func() {
		gpu1 := 2.0
		projectsRaw = Data{
			Projects: []openapi.Project{
				{
					Id:   "proj-1",
					Name: "test-project-1",
					Status: openapi.ProjectStatus{
						Phase: ptr.To(openapi.PROJECTPHASE_CREATING),
						NodePoolQuotaStatuses: []openapi.QuotaStatusNodePool{
							{Allocated: &openapi.QuotaStatusResource{Gpu: &gpu1}},
						},
					},
					Resources: []openapi.Resources{
						{Gpu: openapi.ResourcesGpu{Deserved: 4.0}},
					},
				},
			},
			NextToken: ptr.To("next-token"),
		}
		columns = []ProjectColumn{ColumnKeyName, ColumnKeyStatus}
	})

	It("should create aligned projects from raw data", func() {
		result := CreateAlignedProjects(projectsRaw, columns)

		Expect(result.Projects).To(HaveLen(1))
		Expect(result.Projects[0].Name).To(Equal("test-project-1"))
		Expect(result.Projects[0].Status).To(Equal(string(openapi.PROJECTPHASE_CREATING)))
		Expect(result.NextToken).To(Equal(ptr.To("next-token")))
	})

	It("should handle empty projects list", func() {
		projectsRaw.Projects = []openapi.Project{}

		result := CreateAlignedProjects(projectsRaw, columns)

		Expect(result.Projects).To(HaveLen(0))
		Expect(result.NextToken).To(Equal(ptr.To("next-token")))
	})

	It("should process all specified columns", func() {
		columns = []ProjectColumn{ColumnKeyName, ColumnKeyGPUQuota, ColumnKeyAllocatedGPUs}

		result := CreateAlignedProjects(projectsRaw, columns)

		Expect(result.Projects[0].Name).To(Equal("test-project-1"))
		Expect(result.Projects[0].GPUQuota).To(Equal(4.0))
		Expect(result.Projects[0].AllocatedGPUs).To(Equal(2.0))
	})
})

var _ = Describe("CreateTableData", func() {
	var (
		projectsAligned []ProjectsAlignedRow
		columns         []ProjectColumn
	)

	BeforeEach(func() {
		projectsAligned = []ProjectsAlignedRow{
			{
				Name:            "project-b",
				Status:          "Active",
				GPUQuota:        4.0,
				AllocatedGPUs:   2.5,
				AllocationRatio: 62.5,
			},
			{
				Name:            "project-a",
				Status:          "Pending",
				GPUQuota:        2.0,
				AllocatedGPUs:   1.0,
				AllocationRatio: 50.0,
			},
		}
		columns = []ProjectColumn{ColumnKeyName, ColumnKeyStatus}
	})

	It("should create table data and sort by name", func() {
		result := CreateTableData(projectsAligned, columns)

		Expect(result).To(HaveLen(2))
		// Should be sorted by name (ascending)
		Expect(result[0]).To(Equal([]string{"project-a", "Pending"}))
		Expect(result[1]).To(Equal([]string{"project-b", "Active"}))
	})

	It("should handle different column combinations", func() {
		columns = []ProjectColumn{ColumnKeyName, ColumnKeyGPUQuota, ColumnKeyAllocationRatio}

		result := CreateTableData(projectsAligned, columns)

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(Equal([]string{"project-a", "2.00", "50.00%"}))
		Expect(result[1]).To(Equal([]string{"project-b", "4.00", "62.50%"}))
	})

	It("should handle single column", func() {
		columns = []ProjectColumn{ColumnKeyName}

		result := CreateTableData(projectsAligned, columns)

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(Equal([]string{"project-a"}))
		Expect(result[1]).To(Equal([]string{"project-b"}))
	})

	It("should handle empty projects list", func() {
		projectsAligned = []ProjectsAlignedRow{}

		result := CreateTableData(projectsAligned, columns)

		Expect(result).To(HaveLen(0))
	})
})

var _ = Describe("BriefListFromResponse", func() {
	var projectList []openapi.Project

	BeforeEach(func() {
		gpu1 := 2.0
		parent := openapi.DepartmentProjectParent{Name: "Engineering"}
		phase := openapi.PROJECTPHASE_CREATING

		projectList = []openapi.Project{
			{
				Id:     "proj-1",
				Name:   "test-project",
				Parent: *openapi.NewNullableDepartmentProjectParent(&parent),
				Status: openapi.ProjectStatus{
					Phase: &phase,
					NodePoolQuotaStatuses: []openapi.QuotaStatusNodePool{
						{Allocated: &openapi.QuotaStatusResource{Gpu: &gpu1}},
					},
				},
				Resources: []openapi.Resources{
					{Gpu: openapi.ResourcesGpu{Deserved: 4.0}},
				},
			},
		}
	})

	It("should convert project list to brief list", func() {
		result, err := BriefListFromResponse(projectList)

		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(1))
		Expect(result[0].Id).To(Equal("proj-1"))
		Expect(result[0].Name).To(Equal("test-project"))
		Expect(result[0].DepartmentName).To(Equal("Engineering"))
		Expect(result[0].Phase).To(Equal(string(openapi.PROJECTPHASE_CREATING)))
	})

	It("should handle empty project list", func() {
		result, err := BriefListFromResponse([]openapi.Project{})

		Expect(err).To(BeNil())
		Expect(result).To(HaveLen(0))
	})

	It("should handle project without parent", func() {
		projectList[0].Parent = openapi.NullableDepartmentProjectParent{}

		result, err := BriefListFromResponse(projectList)

		Expect(err).To(BeNil())
		Expect(result[0].DepartmentName).To(Equal(""))
	})
})
