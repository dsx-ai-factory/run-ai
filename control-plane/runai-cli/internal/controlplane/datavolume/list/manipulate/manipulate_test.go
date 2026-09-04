/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/datavolumes-service-client/generated"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ParseDatavolumeColumns", func() {
	Context("when columns slice is empty", func() {
		It("should return empty slice", func() {
			result, err := ParseDatavolumeColumns([]string{})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(0))
		})
	})

	Context("when columns contain valid values", func() {
		It("should parse single valid column", func() {
			result, err := ParseDatavolumeColumns([]string{"name"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(1))
			Expect(result[0]).To(Equal(ColumnKeyName))
		})

		It("should parse multiple valid columns", func() {
			result, err := ParseDatavolumeColumns([]string{"name", "scope", "originpvc"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyScope, ColumnKeyOriginPvc))
		})

		It("should parse the status column", func() {
			result, err := ParseDatavolumeColumns([]string{"status"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(1))
			Expect(result[0]).To(Equal(ColumnKeyStatus))
		})

		It("should handle case insensitive input", func() {
			result, err := ParseDatavolumeColumns([]string{"NAME", "SCOPE", "OriginPvc"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyScope, ColumnKeyOriginPvc))
		})

		It("should trim whitespace", func() {
			result, err := ParseDatavolumeColumns([]string{"  name  ", "scope", " originpvc "})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyScope, ColumnKeyOriginPvc))
		})

		It("should parse all valid columns", func() {
			allColumnNames := []string{"name", "scope", "originpvc"}
			result, err := ParseDatavolumeColumns(allColumnNames)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(3))
		})
	})

	Context("when columns contain invalid values", func() {
		It("should return error for single invalid column", func() {
			result, err := ParseDatavolumeColumns([]string{"invalid"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid column name"))
			Expect(result).To(BeNil())
		})

		It("should return error for multiple invalid columns", func() {
			result, err := ParseDatavolumeColumns([]string{"invalid1", "invalid2"})
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("should return error when mixing valid and invalid columns", func() {
			result, err := ParseDatavolumeColumns([]string{"name", "invalid", "scope"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid column name"))
			Expect(result).To(BeNil())
		})
	})

	Context("when columns contain duplicates", func() {
		It("should include duplicates in result", func() {
			result, err := ParseDatavolumeColumns([]string{"name", "name", "scope"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(3))
			Expect(result[0]).To(Equal(ColumnKeyName))
			Expect(result[1]).To(Equal(ColumnKeyName))
			Expect(result[2]).To(Equal(ColumnKeyScope))
		})
	})
})

var _ = Describe("CreateAlignedDatavolumes", func() {
	var (
		datavolumes []openapi.DatavolumeNoSharedScopes
		columns     []DatavolumeColumn
	)

	BeforeEach(func() {
		dv1 := newTestDatavolumeEntry("dv1", "pvc1", nil, nil)
		deptId := "dept-123"
		dv2 := newTestDatavolumeEntry("dv2", "pvc2", &deptId, nil)
		datavolumes = []openapi.DatavolumeNoSharedScopes{dv1, dv2}
		columns = []DatavolumeColumn{ColumnKeyName, ColumnKeyScope, ColumnKeyOriginPvc}
	})

	It("should create aligned datavolumes with specified columns", func() {
		result := CreateAlignedDatavolumes(datavolumes, columns)

		Expect(result.Datavolumes).To(HaveLen(2))
		Expect(result.Datavolumes[0].Name).To(Equal("dv1"))
		Expect(result.Datavolumes[0].Scope).To(Equal("Project"))
		Expect(result.Datavolumes[0].OriginPvc).To(Equal("pvc1"))

		Expect(result.Datavolumes[1].Name).To(Equal("dv2"))
		Expect(result.Datavolumes[1].Scope).To(Equal("Department"))
		Expect(result.Datavolumes[1].OriginPvc).To(Equal("pvc2"))
	})

	It("should only include selected columns", func() {
		columns = []DatavolumeColumn{ColumnKeyName}
		result := CreateAlignedDatavolumes(datavolumes, columns)

		Expect(result.Datavolumes).To(HaveLen(2))
		Expect(result.Datavolumes[0].Name).To(Equal("dv1"))
		Expect(result.Datavolumes[0].Scope).To(BeEmpty())
		Expect(result.Datavolumes[0].OriginPvc).To(BeEmpty())
	})

	It("should handle empty datavolumes list", func() {
		datavolumes = []openapi.DatavolumeNoSharedScopes{}
		result := CreateAlignedDatavolumes(datavolumes, columns)

		Expect(result.Datavolumes).To(HaveLen(0))
	})
})

var _ = Describe("CreateTableData", func() {
	var (
		alignedDatavolumes []DatavolumesAlignedRow
		columns            []DatavolumeColumn
	)

	BeforeEach(func() {
		alignedDatavolumes = []DatavolumesAlignedRow{
			{
				Name:      "dv1",
				Scope:     "Project",
				OriginPvc: "pvc1",
			},
			{
				Name:      "dv2",
				Scope:     "Department",
				OriginPvc: "pvc2",
			},
		}
		columns = []DatavolumeColumn{ColumnKeyName, ColumnKeyScope, ColumnKeyOriginPvc}
	})

	It("should create table data in correct format", func() {
		result := CreateTableData(alignedDatavolumes, columns)

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(Equal([]string{"dv1", "Project", "pvc1"}))
		Expect(result[1]).To(Equal([]string{"dv2", "Department", "pvc2"}))
	})

	It("should handle different column order", func() {
		columns = []DatavolumeColumn{ColumnKeyScope, ColumnKeyName, ColumnKeyOriginPvc}
		result := CreateTableData(alignedDatavolumes, columns)

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(Equal([]string{"Project", "dv1", "pvc1"}))
		Expect(result[1]).To(Equal([]string{"Department", "dv2", "pvc2"}))
	})

	It("should handle single column", func() {
		columns = []DatavolumeColumn{ColumnKeyName}
		result := CreateTableData(alignedDatavolumes, columns)

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(Equal([]string{"dv1"}))
		Expect(result[1]).To(Equal([]string{"dv2"}))
	})

	It("should handle empty aligned datavolumes", func() {
		alignedDatavolumes = []DatavolumesAlignedRow{}
		result := CreateTableData(alignedDatavolumes, columns)

		Expect(result).To(HaveLen(0))
	})

	It("should include status column when requested", func() {
		alignedDatavolumes = []DatavolumesAlignedRow{
			{Name: "dv1", Status: "Ready"},
			{Name: "dv2", Status: "NotReady"},
		}
		columns = []DatavolumeColumn{ColumnKeyName, ColumnKeyStatus}

		result := CreateTableData(alignedDatavolumes, columns)

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(Equal([]string{"dv1", "No issues found"}))
		Expect(result[1]).To(Equal([]string{"dv2", "Issues found"}))
	})
})

var _ = Describe("GetDefaultColumns", func() {
	It("should return all expected default columns", func() {
		result := GetDefaultColumns()

		expectedColumns := []DatavolumeColumn{
			ColumnKeyName, ColumnKeyStatus, ColumnKeyScope, ColumnKeyOriginPvc,
		}
		Expect(result).To(Equal(expectedColumns))
	})
})

var _ = Describe("GetColumns", func() {
	It("should return provided columns when not empty", func() {
		customColumns := []DatavolumeColumn{ColumnKeyName, ColumnKeyScope}
		result := GetColumns(customColumns)

		Expect(result).To(Equal(customColumns))
	})

	It("should return default columns when empty", func() {
		result := GetColumns([]DatavolumeColumn{})

		expectedColumns := GetDefaultColumns()
		Expect(result).To(Equal(expectedColumns))
	})

	It("should return default columns when nil", func() {
		result := GetColumns(nil)

		expectedColumns := GetDefaultColumns()
		Expect(result).To(Equal(expectedColumns))
	})
})

var _ = Describe("BriefDatavolumeFromEntry", func() {
	It("should convert datavolume entry to brief datavolume", func() {
		dv := newTestDatavolumeEntry("test-dv", "test-pvc", nil, nil)

		result := BriefDatavolumeFromEntry(dv)

		Expect(result.Name).To(Equal("test-dv"))
		Expect(result.Scope).To(Equal("Project"))
		Expect(result.OriginPvc).To(Equal("test-pvc"))
	})

	It("should expose status information", func() {
		dv := newTestDatavolumeEntry("test-dv", "test-pvc", nil, nil)
		notReadyStatus := openapi.NewDatavolumeStatus(openapi.DATAVOLUMEPHASE_NOT_READY)
		dv.Status = *notReadyStatus

		result := BriefDatavolumeFromEntry(dv)
		Expect(result.Status).To(Equal("Issues found"))
	})

	It("should handle different scopes", func() {
		deptId := "dept-123"
		clusterId := "cluster-123"

		testCases := []struct {
			deptId        *string
			clusterId     *string
			expectedScope string
		}{
			{nil, nil, "Project"},
			{&deptId, nil, "Department"},
			{nil, &clusterId, "Cluster"},
		}

		for _, tc := range testCases {
			dv := newTestDatavolumeEntry("test-dv", "test-pvc", tc.deptId, tc.clusterId)
			result := BriefDatavolumeFromEntry(dv)
			Expect(result.Scope).To(Equal(tc.expectedScope), "for deptId=%v, clusterId=%v", tc.deptId, tc.clusterId)
		}
	})
})

func newTestDatavolumeEntry(name string, originPvc string, deptId *string, clusterId *string) openapi.DatavolumeNoSharedScopes {
	status := *openapi.NewDatavolumeStatus(openapi.DATAVOLUMEPHASE_READY)
	dv := openapi.DatavolumeNoSharedScopes{
		Id:            "dv-id-" + name,
		Name:          name,
		OriginPvcName: originPvc,
		ProjectId:     "proj-123",
		ProjectName:   "test-project",
		DepartmentId:  deptId,
		ClusterId:     clusterId,
		Status:        status,
	}
	return dv
}
