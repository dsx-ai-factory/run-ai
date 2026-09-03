/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	"k8s.io/utils/ptr"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
)

var _ = Describe("ParseClusterColumns function", func() {
	Context("when columns slice is empty", func() {
		It("should return nil", func() {
			result, err := ParseClusterColumns([]string{})
			Expect(err).To(BeNil())
			Expect(result).To(BeNil())
		})

		It("should return nil for nil input", func() {
			result, err := ParseClusterColumns(nil)
			Expect(err).To(BeNil())
			Expect(result).To(BeNil())
		})
	})

	Context("when columns contain valid values", func() {
		It("should parse single valid column", func() {
			result, err := ParseClusterColumns([]string{"name"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))
			Expect(result[0]).To(Equal(ColumnKeyName))
		})

		It("should parse multiple valid columns", func() {
			result, err := ParseClusterColumns([]string{"name", "url", "status"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyURL, ColumnKeyStatus))
		})

		It("should handle case insensitive input", func() {
			result, err := ParseClusterColumns([]string{"NAME", "Url", "STATUS"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyURL, ColumnKeyStatus))
		})

		It("should trim whitespace", func() {
			result, err := ParseClusterColumns([]string{"  name  ", " url ", "status"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyURL, ColumnKeyStatus))
		})

		It("should parse all available columns", func() {
			allColumnNames := []string{
				"name", "url", "version", "status", "kubernetesdistribution", "distributionversion",
			}
			result, err := ParseClusterColumns(allColumnNames)
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(len(allColumnNames)))
			Expect(result).To(ContainElements(
				ColumnKeyName, ColumnKeyURL, ColumnKeyVersion,
				ColumnKeyStatus, ColumnKeyKubeDistribution, ColumnKeyKubeVersion,
			))
		})
	})

	Context("when columns contain invalid values", func() {
		It("should return error for single invalid column", func() {
			result, err := ParseClusterColumns([]string{"invalid"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unknown column: invalid"))
			Expect(result).To(BeNil())
		})

		It("should return error for multiple invalid columns", func() {
			result, err := ParseClusterColumns([]string{"invalid1", "invalid2"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unknown column: invalid1"))
			Expect(result).To(BeNil())
		})

		It("should return error when mixing valid and invalid columns", func() {
			result, err := ParseClusterColumns([]string{"name", "invalid", "url"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unknown column: invalid"))
			Expect(result).To(BeNil())
		})

		It("should return error for empty string column", func() {
			result, err := ParseClusterColumns([]string{""})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unknown column:"))
			Expect(result).To(BeNil())
		})
	})

	Context("when columns contain duplicates", func() {
		It("should include duplicates in result", func() {
			result, err := ParseClusterColumns([]string{"name", "name", "url"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))
			Expect(result[0]).To(Equal(ColumnKeyName))
			Expect(result[1]).To(Equal(ColumnKeyName))
			Expect(result[2]).To(Equal(ColumnKeyURL))
		})
	})
})

var _ = Describe("ConvertToAlignedRows function", func() {
	var clusters []openapi.ClustersMinimalData

	BeforeEach(func() {
		// Create sample cluster data
		clusters = []openapi.ClustersMinimalData{
			{
				Uuid:     "uuid-1",
				Name:     "cluster-1",
				TenantId: 1,
			},
			{
				Uuid:     "uuid-2",
				Name:     "cluster-2",
				TenantId: 1,
			},
		}

		// Set nullable fields
		domain1 := "https://cluster-1.example.com"
		clusters[0].Domain.Set(&domain1)
		version1 := "1.0.0"
		clusters[0].Version.Set(&version1)

		domain2 := "https://cluster-2.example.com"
		clusters[1].Domain.Set(&domain2)
		version2 := "1.1.0"
		clusters[1].Version.Set(&version2)

		platform1 := openapi.ClusterReportedStatusPlatform{}
		kubeType1 := "vanilla"
		platform1.Type.Set(&kubeType1)
		kubeVersion1 := "v1.28.0"
		platform1.KubeVersion.Set(&kubeVersion1)
		clusters[0].Platform.Set(&platform1)
		clusters[0].State = ptr.To(openapi.CLUSTERSTATE_CONNECTED)

		platform2 := openapi.ClusterReportedStatusPlatform{}
		kubeType2 := "eks"
		platform2.Type.Set(&kubeType2)
		kubeVersion2 := "v1.29.0"
		platform2.KubeVersion.Set(&kubeVersion2)
		clusters[1].Platform.Set(&platform2)
		clusters[1].State = ptr.To(openapi.CLUSTERSTATE_CONNECTED)
	})

})

var _ = Describe("GetFormattedData function", func() {
	var alignedRows []ClustersAlignedRow

	BeforeEach(func() {
		alignedRows = []ClustersAlignedRow{
			{
				Name:                "cluster-1",
				URL:                 "https://cluster-1.example.com",
				Version:             "1.0.0",
				Status:              "Connected",
				KubeDistribution:    "vanilla",
				DistributionVersion: "v1.28.0",
			},
			{
				Name:                "cluster-2",
				URL:                 "https://cluster-2.example.com",
				Version:             "1.1.0",
				Status:              "Disconnected",
				KubeDistribution:    "eks",
				DistributionVersion: "v1.29.0",
			},
		}
	})

	It("should format data correctly for specified columns", func() {
		columns := []ClusterColumn{ColumnKeyName, ColumnKeyStatus, ColumnKeyURL}

		result := GetFormattedData(alignedRows, columns)

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(Equal([]string{"cluster-1", "Connected", "https://cluster-1.example.com"}))
		Expect(result[1]).To(Equal([]string{"cluster-2", "Disconnected", "https://cluster-2.example.com"}))
	})

	It("should handle empty aligned rows", func() {
		columns := []ClusterColumn{ColumnKeyName, ColumnKeyURL}

		result := GetFormattedData([]ClustersAlignedRow{}, columns)

		Expect(result).To(HaveLen(0))
	})

	It("should sort by name when name column is present", func() {
		// Add data in reverse alphabetical order
		unsortedRows := []ClustersAlignedRow{
			{Name: "z-cluster", Status: "Connected"},
			{Name: "a-cluster", Status: "Disconnected"},
		}
		columns := []ClusterColumn{ColumnKeyName, ColumnKeyStatus}

		result := GetFormattedData(unsortedRows, columns)

		Expect(result).To(HaveLen(2))
		// Should be sorted by name
		Expect(result[0][0]).To(Equal("a-cluster"))
		Expect(result[1][0]).To(Equal("z-cluster"))
	})

	It("should not sort when name column is not present", func() {
		unsortedRows := []ClustersAlignedRow{
			{Status: "Connected", URL: "https://z.example.com"},
			{Status: "Disconnected", URL: "https://a.example.com"},
		}
		columns := []ClusterColumn{ColumnKeyStatus, ColumnKeyURL}

		result := GetFormattedData(unsortedRows, columns)

		Expect(result).To(HaveLen(2))
		// Should maintain original order
		Expect(result[0]).To(Equal([]string{"Connected", "https://z.example.com"}))
		Expect(result[1]).To(Equal([]string{"Disconnected", "https://a.example.com"}))
	})
})

var _ = Describe("GetColumnTitles function", func() {
	It("should return correct titles for specified columns", func() {
		columns := []ClusterColumn{ColumnKeyName, ColumnKeyURL, ColumnKeyStatus}

		result := GetColumnTitles(columns)

		Expect(result).To(Equal([]string{"Name", "URL", "Status"}))
	})

	It("should return empty slice for empty columns", func() {
		result := GetColumnTitles([]ClusterColumn{})

		Expect(result).To(HaveLen(0))
	})

	It("should return titles for all columns", func() {
		columns := []ClusterColumn{
			ColumnKeyName, ColumnKeyURL, ColumnKeyVersion,
			ColumnKeyStatus, ColumnKeyKubeDistribution, ColumnKeyKubeVersion,
		}

		result := GetColumnTitles(columns)

		expectedTitles := []string{
			"Name", "URL", "Run:AI Cluster Version",
			"Status", "Kubernetes Distribution", "Distribution Version",
		}
		Expect(result).To(Equal(expectedTitles))
	})
})

var _ = Describe("GetDefaultColumns function", func() {
	It("should return all expected default columns", func() {
		result := GetDefaultColumns()

		expectedColumns := []ClusterColumn{
			ColumnKeyName, ColumnKeyURL, ColumnKeyVersion,
			ColumnKeyStatus, ColumnKeyKubeDistribution, ColumnKeyKubeVersion,
		}

		Expect(result).To(Equal(expectedColumns))
		Expect(result).To(HaveLen(6))
	})
})

var _ = Describe("CreateAlignedClusters function", func() {
	var clustersData Data

	BeforeEach(func() {
		// Create test data using openapi.ClustersMinimalData structs
		cluster1 := openapi.ClustersMinimalData{
			Uuid:     "uuid-1",
			Name:     "cluster-1",
			TenantId: 1,
		}
		domain1 := "https://cluster-1.example.com"
		version1 := "1.0.0"
		cluster1.Domain.Set(&domain1)
		cluster1.Version.Set(&version1)

		state1 := openapi.CLUSTERSTATE_CONNECTED
		platform1 := openapi.ClusterReportedStatusPlatform{}
		type1 := "vanilla"
		platform1.Type.Set(&type1)
		kubeVersion1 := "v1.28.0"
		platform1.KubeVersion.Set(&kubeVersion1)
		cluster1.State = &state1
		cluster1.Platform.Set(&platform1)

		cluster2 := openapi.ClustersMinimalData{
			Uuid:     "uuid-2",
			Name:     "cluster-2",
			TenantId: 2,
		}
		domain2 := "https://cluster-2.example.com"
		version2 := "1.1.0"
		cluster2.Domain.Set(&domain2)
		cluster2.Version.Set(&version2)

		state2 := openapi.CLUSTERSTATE_DISCONNECTED
		platform2 := openapi.ClusterReportedStatusPlatform{}
		type2 := "eks"
		platform2.Type.Set(&type2)
		kubeVersion2 := "v1.29.0"
		platform2.KubeVersion.Set(&kubeVersion2)
		cluster2.Platform.Set(&platform2)
		cluster2.State = &state2

		clustersData = Data{
			Clusters: []openapi.ClustersMinimalData{cluster1, cluster2},
		}
	})

	It("should create aligned clusters structure", func() {
		columns := []ClusterColumn{ColumnKeyName, ColumnKeyStatus, ColumnKeyURL}

		result := CreateAlignedClusters(clustersData, columns)

		Expect(result.Clusters).To(HaveLen(2))
		Expect(result.NextToken).To(BeNil())

		// Check first cluster - only requested columns should have values
		Expect(result.Clusters[0].Name).To(Equal("cluster-1"))
		Expect(result.Clusters[0].Status).To(Equal("Connected"))
		Expect(result.Clusters[0].URL).To(Equal("https://cluster-1.example.com"))
		// Other fields should be empty since not requested
		Expect(result.Clusters[0].Version).To(Equal(""))
		Expect(result.Clusters[0].KubeDistribution).To(Equal(""))
		Expect(result.Clusters[0].DistributionVersion).To(Equal(""))

		// Check second cluster
		Expect(result.Clusters[1].Name).To(Equal("cluster-2"))
		Expect(result.Clusters[1].Status).To(Equal("Disconnected"))
		Expect(result.Clusters[1].URL).To(Equal("https://cluster-2.example.com"))
	})

	It("should handle all available columns", func() {
		columns := []ClusterColumn{
			ColumnKeyName, ColumnKeyURL, ColumnKeyVersion,
			ColumnKeyStatus, ColumnKeyKubeDistribution, ColumnKeyKubeVersion,
		}

		result := CreateAlignedClusters(clustersData, columns)

		Expect(result.Clusters).To(HaveLen(2))

		// Check first cluster - all fields should have values
		Expect(result.Clusters[0].Name).To(Equal("cluster-1"))
		Expect(result.Clusters[0].URL).To(Equal("https://cluster-1.example.com"))
		Expect(result.Clusters[0].Version).To(Equal("1.0.0"))
		Expect(result.Clusters[0].Status).To(Equal("Connected"))
		Expect(result.Clusters[0].KubeDistribution).To(Equal("vanilla"))
		Expect(result.Clusters[0].DistributionVersion).To(Equal("v1.28.0"))

		// Check second cluster
		Expect(result.Clusters[1].Name).To(Equal("cluster-2"))
		Expect(result.Clusters[1].URL).To(Equal("https://cluster-2.example.com"))
		Expect(result.Clusters[1].Version).To(Equal("1.1.0"))
		Expect(result.Clusters[1].Status).To(Equal("Disconnected"))
		Expect(result.Clusters[1].KubeDistribution).To(Equal("eks"))
		Expect(result.Clusters[1].DistributionVersion).To(Equal("v1.29.0"))
	})

	It("should handle empty clusters data", func() {
		columns := []ClusterColumn{ColumnKeyName, ColumnKeyURL}
		emptyData := Data{Clusters: []openapi.ClustersMinimalData{}}

		result := CreateAlignedClusters(emptyData, columns)

		Expect(result.Clusters).To(HaveLen(0))
		Expect(result.NextToken).To(BeNil())
	})

	It("should handle single column", func() {
		columns := []ClusterColumn{ColumnKeyName}

		result := CreateAlignedClusters(clustersData, columns)

		Expect(result.Clusters).To(HaveLen(2))
		Expect(result.Clusters[0].Name).To(Equal("cluster-1"))
		Expect(result.Clusters[1].Name).To(Equal("cluster-2"))
		// Other fields should be empty strings since they weren't requested
		Expect(result.Clusters[0].URL).To(Equal(""))
		Expect(result.Clusters[0].Version).To(Equal(""))
		Expect(result.Clusters[0].Status).To(Equal(""))
		Expect(result.Clusters[0].KubeDistribution).To(Equal(""))
		Expect(result.Clusters[0].DistributionVersion).To(Equal(""))
	})
})

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
