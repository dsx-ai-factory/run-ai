/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"k8s.io/utils/ptr"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/controlplane/cluster/list/manipulate"
)

var _ = Describe("Output", func() {
	var (
		output   *Output
		clusters []openapi.ClustersMinimalData
	)

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

		// Set status information
		state1 := ptr.To(openapi.CLUSTERSTATE_CONNECTED)
		platform1 := openapi.ClusterReportedStatusPlatform{}
		kubeType1 := "vanilla"
		platform1.Type.Set(&kubeType1)
		kubeVersion1 := "v1.28.0"
		platform1.KubeVersion.Set(&kubeVersion1)
		clusters[0].Platform.Set(&platform1)
		clusters[0].State = state1

		state2 := ptr.To(openapi.CLUSTERSTATE_DISCONNECTED)
		platform2 := openapi.ClusterReportedStatusPlatform{}
		kubeType2 := "eks"
		platform2.Type.Set(&kubeType2)
		kubeVersion2 := "v1.29.0"
		platform2.KubeVersion.Set(&kubeVersion2)
		clusters[1].Platform.Set(&platform2)
		clusters[1].State = state2

		var err error
		output, err = NewOutput(clusters, nil)
		Expect(err).ToNot(HaveOccurred())
	})

	Context("SetColumns functionality", func() {
		It("should set custom columns successfully", func() {
			columns := []string{"name", "url", "status"}

			err := output.SetColumns(columns)

			Expect(err).ToNot(HaveOccurred())
			Expect(output.columns).To(HaveLen(3))
			Expect(output.columns).To(ContainElements(
				manipulate.ColumnKeyName,
				manipulate.ColumnKeyURL,
				manipulate.ColumnKeyStatus,
			))
		})

		It("should handle empty columns list", func() {
			err := output.SetColumns([]string{})

			Expect(err).ToNot(HaveOccurred())
		})

		It("should handle nil columns list", func() {
			err := output.SetColumns(nil)

			Expect(err).ToNot(HaveOccurred())
		})

		It("should return error for invalid columns", func() {
			columns := []string{"name", "invalid-column", "url"}

			err := output.SetColumns(columns)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unknown column: invalid-column"))
		})

		It("should handle case insensitive columns", func() {
			columns := []string{"Name", "URL", "Status"}

			err := output.SetColumns(columns)

			Expect(err).ToNot(HaveOccurred())
			Expect(output.columns).To(HaveLen(3))
		})

		It("should handle whitespace in columns", func() {
			columns := []string{"  name  ", " url ", "status"}

			err := output.SetColumns(columns)

			Expect(err).ToNot(HaveOccurred())
			Expect(output.columns).To(HaveLen(3))
		})
	})

	Context("Columns method with custom columns", func() {
		It("should return custom columns when set", func() {
			columns := []string{"name", "status"}
			err := output.SetColumns(columns)
			Expect(err).ToNot(HaveOccurred())

			tableColumns := output.Columns()

			Expect(tableColumns).To(HaveLen(2))
			Expect(tableColumns[0].Title).To(Equal("Name"))
			Expect(tableColumns[1].Title).To(Equal("Status"))
		})

		It("should return default columns when no custom columns set", func() {
			tableColumns := output.Columns()

			defaultColumns := manipulate.GetDefaultColumns()
			Expect(tableColumns).To(HaveLen(len(defaultColumns)))

			expectedTitles := []string{
				"Name", "URL", "Run:AI Cluster Version",
				"Status", "Kubernetes Distribution", "Distribution Version",
			}
			for i, col := range tableColumns {
				Expect(col.Title).To(Equal(expectedTitles[i]))
			}
		})

		It("should return custom columns with correct widths", func() {
			columns := []string{"name", "url"}
			err := output.SetColumns(columns)
			Expect(err).ToNot(HaveOccurred())

			tableColumns := output.Columns()

			Expect(tableColumns).To(HaveLen(2))
			Expect(tableColumns[0].Width).To(Equal(manipulate.DefaultWidthName))
			Expect(tableColumns[1].Width).To(Equal(manipulate.DefaultWidthURL))
		})
	})

	Context("Data method with custom columns", func() {
		It("should return data for custom columns", func() {
			columns := []string{"name", "status"}
			err := output.SetColumns(columns)
			Expect(err).ToNot(HaveOccurred())

			data := output.Data()

			Expect(data).To(HaveLen(2))
			// Check first cluster
			Expect(data[0]).To(HaveLen(2))
			Expect(data[0][0]).To(Equal("cluster-1"))
			Expect(data[0][1]).To(Equal("Connected"))
			// Check second cluster
			Expect(data[1]).To(HaveLen(2))
			Expect(data[1][0]).To(Equal("cluster-2"))
			Expect(data[1][1]).To(Equal("Disconnected"))
		})

		It("should return default data when no custom columns set", func() {
			data := output.Data()

			Expect(data).To(HaveLen(2))
			// Should have all default columns
			defaultColumns := manipulate.GetDefaultColumns()
			Expect(data[0]).To(HaveLen(len(defaultColumns)))
			Expect(data[1]).To(HaveLen(len(defaultColumns)))
		})

		It("should handle single column selection", func() {
			columns := []string{"name"}
			err := output.SetColumns(columns)
			Expect(err).ToNot(HaveOccurred())

			data := output.Data()

			Expect(data).To(HaveLen(2))
			Expect(data[0]).To(HaveLen(1))
			Expect(data[0][0]).To(Equal("cluster-1"))
			Expect(data[1]).To(HaveLen(1))
			Expect(data[1][0]).To(Equal("cluster-2"))
		})

		It("should handle all available columns", func() {
			columns := []string{"name", "url", "version", "status", "kubernetesdistribution", "distributionversion"}
			err := output.SetColumns(columns)
			Expect(err).ToNot(HaveOccurred())

			data := output.Data()

			Expect(data).To(HaveLen(2))
			Expect(data[0]).To(HaveLen(6))
			Expect(data[1]).To(HaveLen(6))

			// Check specific values
			Expect(data[0][0]).To(Equal("cluster-1"))                     // name
			Expect(data[0][1]).To(Equal("https://cluster-1.example.com")) // url
			Expect(data[0][2]).To(Equal("1.0.0"))                         // version
			Expect(data[0][3]).To(Equal("Connected"))                     // status
			Expect(data[0][4]).To(Equal("vanilla"))                       // kube-distribution
			Expect(data[0][5]).To(Equal("v1.28.0"))                       // kube-version
		})
	})

	Context("getActiveColumns method", func() {
		It("should return custom columns when set", func() {
			columns := []string{"name", "status"}
			err := output.SetColumns(columns)
			Expect(err).ToNot(HaveOccurred())

			activeColumns := output.getActiveColumns()

			Expect(activeColumns).To(HaveLen(2))
			Expect(activeColumns).To(ContainElements(
				manipulate.ColumnKeyName,
				manipulate.ColumnKeyStatus,
			))
		})

		It("should return default columns when no custom columns", func() {
			activeColumns := output.getActiveColumns()

			expectedColumns := manipulate.GetDefaultColumns()
			Expect(activeColumns).To(Equal(expectedColumns))
		})

		It("should return default columns when custom columns is empty", func() {
			output.columns = []manipulate.ClusterColumn{}

			activeColumns := output.getActiveColumns()

			expectedColumns := manipulate.GetDefaultColumns()
			Expect(activeColumns).To(Equal(expectedColumns))
		})
	})

	Context("Raw flag functionality", func() {
		It("should set raw flag correctly", func() {
			output.SetRaw(true)
			Expect(output.raw).To(BeTrue())

			output.SetRaw(false)
			Expect(output.raw).To(BeFalse())
		})

		It("should return raw data when raw=true for JSON", func() {
			output.SetRaw(true)

			jsonOutput, err := output.GenerateJson()
			Expect(err).ToNot(HaveOccurred())

			// Should contain original API fields
			Expect(jsonOutput).To(ContainSubstring("\"uuid\""))
			Expect(jsonOutput).To(ContainSubstring("\"tenantId\""))

			// Should not contain aligned format (column titles as keys)
			Expect(jsonOutput).ToNot(ContainSubstring("\"Name\""))
			Expect(jsonOutput).ToNot(ContainSubstring("\"Status\""))
			Expect(jsonOutput).ToNot(ContainSubstring("\"URL\""))
		})

		It("should return raw data when raw=true for YAML", func() {
			output.SetRaw(true)

			yamlOutput, err := output.GenerateYaml()
			Expect(err).ToNot(HaveOccurred())

			// Should contain original API fields
			Expect(yamlOutput).To(ContainSubstring("uuid:"))
			Expect(yamlOutput).To(ContainSubstring("tenantid:")) // YAML lowercases field names
		})

		It("should return non-raw data when raw=false", func() {
			output.SetRaw(false)

			jsonOutput, err := output.GenerateJson()
			Expect(err).ToNot(HaveOccurred())

			Expect(jsonOutput).To(ContainSubstring("\"clusters\""))

			output.SetRaw(true)
			rawOutput, err := output.GenerateJson()
			Expect(err).ToNot(HaveOccurred())

			Expect(jsonOutput).ToNot(Equal(rawOutput))
		})

		It("should return non-raw data when raw=false for YAML", func() {
			output.SetRaw(false)

			yamlOutput, err := output.GenerateYaml()
			Expect(err).ToNot(HaveOccurred())

			// Should contain clusters
			Expect(yamlOutput).To(ContainSubstring("clusters:"))

			// The exact format depends on settings, but should be different from raw
			output.SetRaw(true)
			rawYamlOutput, err := output.GenerateYaml()
			Expect(err).ToNot(HaveOccurred())

			// The outputs should be different
			Expect(yamlOutput).ToNot(Equal(rawYamlOutput))
		})
	})
})
