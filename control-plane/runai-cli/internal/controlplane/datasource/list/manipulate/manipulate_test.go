/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
)

var _ = Describe("ParseDatasourceColumns", func() {
	Context("when columns slice is empty", func() {
		It("should return empty slice", func() {
			result, err := ParseDatasourceColumns([]string{})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(0))
		})
	})

	Context("when columns contain valid values", func() {
		It("should parse single valid column", func() {
			result, err := ParseDatasourceColumns([]string{"name"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(1))
			Expect(result[0]).To(Equal(ColumnKeyName))
		})

		It("should parse multiple valid columns", func() {
			result, err := ParseDatasourceColumns([]string{"name", "type", "scope"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyType, ColumnKeyScope))
		})

		It("should handle case insensitive input", func() {
			result, err := ParseDatasourceColumns([]string{"NAME", "Type", "SCOPE"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyType, ColumnKeyScope))
		})

		It("should trim whitespace", func() {
			result, err := ParseDatasourceColumns([]string{"  name  ", " type ", "scope"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyType, ColumnKeyScope))
		})

		It("should parse all valid columns", func() {
			allColumnNames := []string{"name", "type", "scope"}
			result, err := ParseDatasourceColumns(allColumnNames)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(3))
		})
	})

	Context("when columns contain invalid values", func() {
		It("should return error for single invalid column", func() {
			result, err := ParseDatasourceColumns([]string{"invalid"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid column name"))
			Expect(result).To(BeNil())
		})

		It("should return error for multiple invalid columns", func() {
			result, err := ParseDatasourceColumns([]string{"invalid1", "invalid2"})
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("should return error when mixing valid and invalid columns", func() {
			result, err := ParseDatasourceColumns([]string{"name", "invalid", "scope"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid column name"))
			Expect(result).To(BeNil())
		})
	})

	Context("when columns contain duplicates", func() {
		It("should include duplicates in result", func() {
			result, err := ParseDatasourceColumns([]string{"name", "name", "type"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(3))
			Expect(result[0]).To(Equal(ColumnKeyName))
			Expect(result[1]).To(Equal(ColumnKeyName))
			Expect(result[2]).To(Equal(ColumnKeyType))
		})
	})
})

var _ = Describe("CreateAlignedDatasources", func() {
	var (
		datasources []openapi.DatasourceListResponseEntry
		columns     []DatasourceColumn
	)

	BeforeEach(func() {
		ds1 := newTestDatasourceEntry("ds1", openapi.ASSETKIND_NFS, "Cluster")
		ds2 := newTestDatasourceEntry("ds2", openapi.ASSETKIND_S3, "Project")
		datasources = []openapi.DatasourceListResponseEntry{ds1, ds2}
		columns = []DatasourceColumn{ColumnKeyName, ColumnKeyType, ColumnKeyScope}
	})

	It("should create aligned datasources with specified columns", func() {
		result := CreateAlignedDatasources(datasources, columns)

		Expect(result.Datasources).To(HaveLen(2))
		Expect(result.Datasources[0].Name).To(Equal("ds1"))
		Expect(result.Datasources[0].Type).To(Equal("Nfs"))
		Expect(result.Datasources[0].Scope).To(Equal("Cluster"))

		Expect(result.Datasources[1].Name).To(Equal("ds2"))
		Expect(result.Datasources[1].Type).To(Equal("S3"))
		Expect(result.Datasources[1].Scope).To(Equal("Project"))
	})

	It("should only include selected columns", func() {
		columns = []DatasourceColumn{ColumnKeyName}
		result := CreateAlignedDatasources(datasources, columns)

		Expect(result.Datasources).To(HaveLen(2))
		Expect(result.Datasources[0].Name).To(Equal("ds1"))
		Expect(result.Datasources[0].Type).To(BeEmpty())
		Expect(result.Datasources[0].Scope).To(BeEmpty())
	})

	It("should handle empty datasources list", func() {
		datasources = []openapi.DatasourceListResponseEntry{}
		result := CreateAlignedDatasources(datasources, columns)

		Expect(result.Datasources).To(HaveLen(0))
	})
})

var _ = Describe("CreateTableData", func() {
	var (
		alignedDatasources []DatasourcesAlignedRow
		columns            []DatasourceColumn
	)

	BeforeEach(func() {
		alignedDatasources = []DatasourcesAlignedRow{
			{
				Name:  "ds1",
				Type:  "Nfs",
				Scope: "Cluster",
			},
			{
				Name:  "ds2",
				Type:  "S3",
				Scope: "Project",
			},
		}
		columns = []DatasourceColumn{ColumnKeyName, ColumnKeyType, ColumnKeyScope}
	})

	It("should create table data in correct format", func() {
		result := CreateTableData(alignedDatasources, columns)

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(Equal([]string{"ds1", "Nfs", "Cluster"}))
		Expect(result[1]).To(Equal([]string{"ds2", "S3", "Project"}))
	})

	It("should handle different column order", func() {
		columns = []DatasourceColumn{ColumnKeyScope, ColumnKeyName, ColumnKeyType}
		result := CreateTableData(alignedDatasources, columns)

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(Equal([]string{"Cluster", "ds1", "Nfs"}))
		Expect(result[1]).To(Equal([]string{"Project", "ds2", "S3"}))
	})

	It("should handle single column", func() {
		columns = []DatasourceColumn{ColumnKeyName}
		result := CreateTableData(alignedDatasources, columns)

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(Equal([]string{"ds1"}))
		Expect(result[1]).To(Equal([]string{"ds2"}))
	})

	It("should handle empty aligned datasources", func() {
		alignedDatasources = []DatasourcesAlignedRow{}
		result := CreateTableData(alignedDatasources, columns)

		Expect(result).To(HaveLen(0))
	})
})

var _ = Describe("GetDefaultColumns", func() {
	It("should return all expected default columns", func() {
		result := GetDefaultColumns()

		expectedColumns := []DatasourceColumn{
			ColumnKeyName, ColumnKeyType, ColumnKeyScope,
		}
		Expect(result).To(Equal(expectedColumns))
	})
})

var _ = Describe("GetColumns", func() {
	It("should return provided columns when not empty", func() {
		customColumns := []DatasourceColumn{ColumnKeyName, ColumnKeyType}
		result := GetColumns(customColumns)

		Expect(result).To(Equal(customColumns))
	})

	It("should return default columns when empty", func() {
		result := GetColumns([]DatasourceColumn{})

		expectedColumns := GetDefaultColumns()
		Expect(result).To(Equal(expectedColumns))
	})

	It("should return default columns when nil", func() {
		result := GetColumns(nil)

		expectedColumns := GetDefaultColumns()
		Expect(result).To(Equal(expectedColumns))
	})
})

var _ = Describe("BriefDatasourceFromEntry", func() {
	It("should convert datasource entry to brief datasource", func() {
		ds := newTestDatasourceEntry("test-ds", openapi.ASSETKIND_PVC, "Project")

		result := BriefDatasourceFromEntry(ds)

		Expect(result.Name).To(Equal("test-ds"))
		Expect(result.Type).To(Equal("Pvc"))
		Expect(result.Scope).To(Equal("Project"))
	})

	It("should handle different datasource types", func() {
		testCases := []struct {
			kind         openapi.AssetKind
			expectedType string
		}{
			{openapi.ASSETKIND_NFS, "Nfs"},
			{openapi.ASSETKIND_S3, "S3"},
			{openapi.ASSETKIND_GIT, "Git"},
			{openapi.ASSETKIND_PVC, "Pvc"},
			{openapi.ASSETKIND_HOST_PATH, "HostPath"},
			{openapi.ASSETKIND_CONFIG_MAP, "ConfigMap"},
			{openapi.ASSETKIND_SECRET_VOLUME, "Secret"},
			{openapi.ASSETKIND_DATA_VOLUME, "DataVolume"},
		}

		for _, tc := range testCases {
			ds := newTestDatasourceEntry("test-ds", tc.kind, "Cluster")
			result := BriefDatasourceFromEntry(ds)
			Expect(result.Type).To(Equal(tc.expectedType), "for kind %s", tc.kind)
		}
	})
})

// Helper function to create test datasource entries
func newTestDatasourceEntry(name string, kind openapi.AssetKind, scope string) openapi.DatasourceListResponseEntry {
	return openapi.DatasourceListResponseEntry{
		Meta: openapi.AssetMeta{
			Id:    "ds-id-" + name,
			Name:  name,
			Kind:  kind,
			Scope: openapi.Scope(scope),
		},
		Spec: openapi.DatasourceListResponseAssetSpec{},
	}
}
