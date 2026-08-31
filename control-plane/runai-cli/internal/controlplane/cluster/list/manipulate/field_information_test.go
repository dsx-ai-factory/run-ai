/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	"k8s.io/utils/ptr"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
)

var _ = Describe("Field Extract Functions", func() {
	var (
		cluster    *openapi.ClustersMinimalData
		clusterRow *ClustersAlignedRow
	)

	BeforeEach(func() {
		cluster = &openapi.ClustersMinimalData{
			Uuid:     "test-uuid",
			Name:     "test-cluster",
			TenantId: 1,
		}
		clusterRow = &ClustersAlignedRow{}
	})

	Context("NameField extraction", func() {
		It("should extract cluster name correctly", func() {
			NameField.extractFunc(cluster, clusterRow)

			Expect(clusterRow.Name).To(Equal("test-cluster"))
		})

		It("should handle empty name", func() {
			cluster.Name = ""

			NameField.extractFunc(cluster, clusterRow)

			Expect(clusterRow.Name).To(Equal(""))
		})
	})

	Context("URLField extraction", func() {
		It("should extract domain when set", func() {
			domain := "https://test.example.com"
			cluster.Domain.Set(&domain)

			URLField.extractFunc(cluster, clusterRow)

			Expect(clusterRow.URL).To(Equal("https://test.example.com"))
		})

		It("should handle unset domain", func() {
			// Domain is not set (default state)
			URLField.extractFunc(cluster, clusterRow)

			Expect(clusterRow.URL).To(Equal(""))
		})

		It("should handle nil domain", func() {
			cluster.Domain.Set(nil)

			URLField.extractFunc(cluster, clusterRow)

			Expect(clusterRow.URL).To(Equal(""))
		})
	})

	Context("VersionField extraction", func() {
		It("should extract version when set", func() {
			version := "1.2.3"
			cluster.Version.Set(&version)

			VersionField.extractFunc(cluster, clusterRow)

			Expect(clusterRow.Version).To(Equal("1.2.3"))
		})

		It("should handle unset version", func() {
			VersionField.extractFunc(cluster, clusterRow)

			Expect(clusterRow.Version).To(Equal(""))
		})

		It("should handle nil version", func() {
			cluster.Version.Set(nil)

			VersionField.extractFunc(cluster, clusterRow)

			Expect(clusterRow.Version).To(Equal(""))
		})
	})

	Context("StatusField extraction", func() {
		It("should extract status state when set", func() {
			cluster.State = ptr.To(openapi.CLUSTERSTATE_CONNECTED)

			StatusField.extractFunc(cluster, clusterRow)

			Expect(clusterRow.Status).To(Equal("Connected"))
		})

		It("should handle unset status", func() {
			StatusField.extractFunc(cluster, clusterRow)

			Expect(clusterRow.Status).To(Equal(""))
		})

		It("should handle status with nil state", func() {
			cluster.State = nil

			StatusField.extractFunc(cluster, clusterRow)

			Expect(clusterRow.Status).To(Equal(""))
		})

		It("should handle nil status", func() {
			cluster.State = nil

			StatusField.extractFunc(cluster, clusterRow)

			Expect(clusterRow.Status).To(Equal(""))
		})
	})

	Context("KubeDistributionField extraction", func() {
		It("should extract kubernetes distribution when set", func() {
			platform := openapi.ClusterReportedStatusPlatform{}
			kubeType := "vanilla"
			platform.Type.Set(&kubeType)

			cluster.Platform.Set(&platform)
			KubeDistributionField.extractFunc(cluster, clusterRow)

			Expect(clusterRow.KubeDistribution).To(Equal("vanilla"))
		})

		It("should handle unset status", func() {
			KubeDistributionField.extractFunc(cluster, clusterRow)

			Expect(clusterRow.KubeDistribution).To(Equal(""))
		})

		It("should handle unset platform", func() {
			cluster.State = nil

			KubeDistributionField.extractFunc(cluster, clusterRow)

			Expect(clusterRow.KubeDistribution).To(Equal(""))
		})

		It("should handle unset platform type", func() {
			platform := openapi.ClusterReportedStatusPlatform{}
			cluster.Platform.Set(&platform)

			KubeDistributionField.extractFunc(cluster, clusterRow)

			Expect(clusterRow.KubeDistribution).To(Equal(""))
		})

		It("should handle nil platform type", func() {
			platform := openapi.ClusterReportedStatusPlatform{}
			platform.Type.Set(nil)

			cluster.Platform.Set(&platform)

			KubeDistributionField.extractFunc(cluster, clusterRow)

			Expect(clusterRow.KubeDistribution).To(Equal(""))
		})
	})

	Context("KubeVersionField extraction", func() {
		It("should extract kubernetes version when set", func() {
			platform := openapi.ClusterReportedStatusPlatform{}
			kubeVersion := "v1.28.0"
			platform.KubeVersion.Set(&kubeVersion)

			cluster.Platform.Set(&platform)

			KubeVersionField.extractFunc(cluster, clusterRow)

			Expect(clusterRow.DistributionVersion).To(Equal("v1.28.0"))
		})

		It("should handle unset status", func() {
			KubeVersionField.extractFunc(cluster, clusterRow)

			Expect(clusterRow.DistributionVersion).To(Equal(""))
		})

		It("should handle unset platform", func() {
			cluster.Platform.Set(nil)

			KubeVersionField.extractFunc(cluster, clusterRow)

			Expect(clusterRow.DistributionVersion).To(Equal(""))
		})

		It("should handle unset kube version", func() {
			platform := openapi.ClusterReportedStatusPlatform{}
			cluster.Platform.Set(&platform)

			KubeVersionField.extractFunc(cluster, clusterRow)

			Expect(clusterRow.DistributionVersion).To(Equal(""))
		})

		It("should handle nil kube version", func() {
			platform := openapi.ClusterReportedStatusPlatform{}
			platform.KubeVersion.Set(nil)

			cluster.Platform.Set(&platform)

			KubeVersionField.extractFunc(cluster, clusterRow)

			Expect(clusterRow.DistributionVersion).To(Equal(""))
		})
	})
})

var _ = Describe("Field Format Functions", func() {
	var clusterRow *ClustersAlignedRow

	BeforeEach(func() {
		clusterRow = &ClustersAlignedRow{}
	})

	Context("NameField formatting", func() {
		It("should format name correctly when present", func() {
			clusterRow.Name = "test-cluster"

			result := NameField.formatFunc(clusterRow)

			Expect(result).To(Equal("test-cluster"))
		})

		It("should handle empty name", func() {
			clusterRow.Name = ""

			result := NameField.formatFunc(clusterRow)

			Expect(result).To(Equal(""))
		})

		It("should handle missing name field", func() {
			result := NameField.formatFunc(clusterRow)

			Expect(result).To(Equal(""))
		})

		It("should handle non-string name value", func() {
			clusterRow.Name = "123"

			result := NameField.formatFunc(clusterRow)

			Expect(result).To(Equal("123"))
		})
	})

	Context("URLField formatting", func() {
		It("should format URL correctly when present", func() {
			clusterRow.URL = "https://test.example.com"

			result := URLField.formatFunc(clusterRow)

			Expect(result).To(Equal("https://test.example.com"))
		})

		It("should handle empty URL", func() {
			clusterRow.URL = ""

			result := URLField.formatFunc(clusterRow)

			Expect(result).To(Equal(""))
		})

		It("should handle missing URL field", func() {
			result := URLField.formatFunc(clusterRow)

			Expect(result).To(Equal(""))
		})
	})

	Context("VersionField formatting", func() {
		It("should format version correctly when present", func() {
			clusterRow.Version = "1.2.3"

			result := VersionField.formatFunc(clusterRow)

			Expect(result).To(Equal("1.2.3"))
		})

		It("should handle empty version", func() {
			clusterRow.Version = ""

			result := VersionField.formatFunc(clusterRow)

			Expect(result).To(Equal(""))
		})

		It("should handle missing version field", func() {
			result := VersionField.formatFunc(clusterRow)

			Expect(result).To(Equal(""))
		})
	})

	Context("StatusField formatting", func() {
		It("should format status correctly when present", func() {
			clusterRow.Status = "Connected"

			result := StatusField.formatFunc(clusterRow)

			Expect(result).To(Equal("Connected"))
		})

		It("should handle empty status", func() {
			clusterRow.Status = ""

			result := StatusField.formatFunc(clusterRow)

			Expect(result).To(Equal(""))
		})

		It("should handle missing status field", func() {
			result := StatusField.formatFunc(clusterRow)

			Expect(result).To(Equal(""))
		})
	})

	Context("KubeDistributionField formatting", func() {
		It("should format kubernetes distribution correctly when present", func() {
			clusterRow.KubeDistribution = "vanilla"

			result := KubeDistributionField.formatFunc(clusterRow)

			Expect(result).To(Equal("vanilla"))
		})

		It("should handle empty kubernetes distribution", func() {
			clusterRow.KubeDistribution = ""

			result := KubeDistributionField.formatFunc(clusterRow)

			Expect(result).To(Equal(""))
		})

		It("should handle missing kubernetes distribution field", func() {
			result := KubeDistributionField.formatFunc(clusterRow)

			Expect(result).To(Equal(""))
		})
	})

	Context("KubeVersionField formatting", func() {
		It("should format kubernetes version correctly when present", func() {
			clusterRow.DistributionVersion = "v1.28.0"

			result := KubeVersionField.formatFunc(clusterRow)

			Expect(result).To(Equal("v1.28.0"))
		})

		It("should handle empty kubernetes version", func() {
			clusterRow.DistributionVersion = ""

			result := KubeVersionField.formatFunc(clusterRow)

			Expect(result).To(Equal(""))
		})

		It("should handle missing kubernetes version field", func() {
			result := KubeVersionField.formatFunc(clusterRow)

			Expect(result).To(Equal(""))
		})
	})
})

var _ = Describe("Field Information Properties", func() {
	Context("Column widths", func() {
		It("should have correct default widths", func() {
			Expect(NameField.GetColumnWidth()).To(Equal(DefaultWidthName))
			Expect(URLField.GetColumnWidth()).To(Equal(DefaultWidthURL))
			Expect(VersionField.GetColumnWidth()).To(Equal(DefaultWidthVersion))
			Expect(StatusField.GetColumnWidth()).To(Equal(DefaultWidthStatus))
			Expect(KubeDistributionField.GetColumnWidth()).To(Equal(DefaultWidthKubeDistribution))
			Expect(KubeVersionField.GetColumnWidth()).To(Equal(DefaultWidthKubeVersion))
		})
	})

	Context("Column titles", func() {
		It("should have correct titles", func() {
			Expect(NameField.GetColumnTitle()).To(Equal("Name"))
			Expect(URLField.GetColumnTitle()).To(Equal("URL"))
			Expect(VersionField.GetColumnTitle()).To(Equal("Run:AI Cluster Version"))
			Expect(StatusField.GetColumnTitle()).To(Equal("Status"))
			Expect(KubeDistributionField.GetColumnTitle()).To(Equal("Kubernetes Distribution"))
			Expect(KubeVersionField.GetColumnTitle()).To(Equal("Distribution Version"))
		})
	})
})

var _ = Describe("ColumnMap", func() {
	It("should contain all expected columns", func() {
		expectedColumns := []ClusterColumn{
			ColumnKeyName, ColumnKeyURL, ColumnKeyVersion,
			ColumnKeyStatus, ColumnKeyKubeDistribution, ColumnKeyKubeVersion,
		}

		Expect(ColumnMap).To(HaveLen(len(expectedColumns)))

		for _, column := range expectedColumns {
			Expect(ColumnMap).To(HaveKey(column))
		}
	})

	It("should have valid field information for each column", func() {
		for column, fieldInfo := range ColumnMap {
			Expect(fieldInfo.GetColumnTitle()).ToNot(BeEmpty(), "Column %s should have a title", column)
			Expect(fieldInfo.GetColumnWidth()).To(BeNumerically(">", 0), "Column %s should have a positive width", column)
			Expect(fieldInfo.extractFunc).ToNot(BeNil(), "Column %s should have an extract function", column)
			Expect(fieldInfo.formatFunc).ToNot(BeNil(), "Column %s should have a format function", column)
		}
	})
})

// Helper function to create string pointers - using the one from manipulate_cluster_list_test.go
