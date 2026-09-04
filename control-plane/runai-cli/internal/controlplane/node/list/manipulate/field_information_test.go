/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ParseNodeColumns function", func() {
	Context("when columns slice is empty", func() {
		It("should return nil", func() {
			result, err := ParseNodeColumns([]string{})
			Expect(err).To(BeNil())
			Expect(result).To(BeNil())
		})

		It("should return nil for nil input", func() {
			result, err := ParseNodeColumns(nil)
			Expect(err).To(BeNil())
			Expect(result).To(BeNil())
		})
	})

	Context("when columns contain valid values", func() {
		It("should parse single valid column", func() {
			result, err := ParseNodeColumns([]string{"name"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(1))
			Expect(result[0]).To(Equal(ColumnKeyName))
		})

		It("should parse multiple valid columns", func() {
			result, err := ParseNodeColumns([]string{"name", "status", "gputype"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyStatus, ColumnKeyGpuType))
		})

		It("should handle case insensitive input", func() {
			result, err := ParseNodeColumns([]string{"NAME", "Status", "GPUTYPE"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyStatus, ColumnKeyGpuType))
		})

		It("should trim whitespace", func() {
			result, err := ParseNodeColumns([]string{"  name  ", " status ", "gputype"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyStatus, ColumnKeyGpuType))
		})

		It("should parse all valid columns", func() {
			allColumnNames := []string{
				"name", "status", "gputype", "gpuallocation", "freegpu", "nodepool",
			}
			result, err := ParseNodeColumns(allColumnNames)
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(len(allColumnNames)))
		})
	})

	Context("when columns contain invalid values", func() {
		It("should return error for single invalid column", func() {
			result, err := ParseNodeColumns([]string{"invalid"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid column names: \"invalid\""))
			Expect(result).To(BeNil())
		})

		It("should return error for multiple invalid columns", func() {
			result, err := ParseNodeColumns([]string{"invalid1", "", "invalid2"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid column names: \"invalid1\", \"\", \"invalid2\""))
			Expect(result).To(BeNil())
		})

		It("should return error when mixing valid and invalid columns", func() {
			result, err := ParseNodeColumns([]string{"name", "invalid", "status"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid column names: \"invalid\""))
			Expect(result).To(BeNil())
		})
	})

	Context("when columns contain duplicates", func() {
		It("should include duplicates in result", func() {
			result, err := ParseNodeColumns([]string{"name", "name", "status"})
			Expect(err).To(BeNil())
			Expect(result).To(HaveLen(3))
			Expect(result[0]).To(Equal(ColumnKeyName))
			Expect(result[1]).To(Equal(ColumnKeyName))
			Expect(result[2]).To(Equal(ColumnKeyStatus))
		})
	})
})

var _ = Describe("Field extraction functions", func() {
	var (
		testNode openapi.Node
		nodeRow  NodesAlignedRow
	)

	BeforeEach(func() {
		testNode = openapi.Node{
			Name:     "test-node",
			Status:   "Ready",
			NodePool: "default-pool",
		}
		nodeRow = NodesAlignedRow{}
	})

	Context("extractNameData", func() {
		It("should extract node name", func() {
			extractNameData(&testNode, &nodeRow)
			Expect(nodeRow.Name).To(Equal("test-node"))
		})
	})

	Context("extractStatusData", func() {
		It("should extract node status", func() {
			extractStatusData(&testNode, &nodeRow)
			Expect(nodeRow.Status).To(Equal("Ready"))
		})
	})

	Context("extractNodepoolData", func() {
		It("should extract nodepool", func() {
			extractNodepoolData(&testNode, &nodeRow)
			Expect(nodeRow.Nodepool).To(Equal("default-pool"))
		})
	})

	Context("extractGpuTypeData", func() {
		It("should extract GPU type when available", func() {
			gpuInfo := openapi.NewGpuInfo("V100", 8)
			testNode.GpuInfo = *openapi.NewNullableGpuInfo(gpuInfo)

			extractGpuTypeData(&testNode, &nodeRow)
			Expect(nodeRow.GpuType).To(Equal("V100"))
		})

		It("should handle missing GPU info", func() {
			extractGpuTypeData(&testNode, &nodeRow)
			Expect(nodeRow.GpuType).To(Equal(""))
		})
	})
})

var _ = Describe("Format functions", func() {
	Context("formatNameValue", func() {
		It("should return the name", func() {
			nodeRow := NodesAlignedRow{Name: "test-node"}
			result := formatNameValue(&nodeRow)
			Expect(result).To(Equal("test-node"))
		})
	})

	Context("formatStatusValue", func() {
		It("should return the status", func() {
			nodeRow := NodesAlignedRow{Status: "Ready"}
			result := formatStatusValue(&nodeRow)
			Expect(result).To(Equal("Ready"))
		})
	})

	Context("formatGpuTypeValue", func() {
		It("should return GPU type when available", func() {
			nodeRow := NodesAlignedRow{GpuType: "V100"}
			result := formatGpuTypeValue(&nodeRow)
			Expect(result).To(Equal("V100"))
		})

		It("should return noDataFormat when empty", func() {
			nodeRow := NodesAlignedRow{GpuType: ""}
			result := formatGpuTypeValue(&nodeRow)
			Expect(result).To(Equal(noDataFormat))
		})
	})

	Context("formatNodepoolValue", func() {
		It("should return nodepool when available", func() {
			nodeRow := NodesAlignedRow{Nodepool: "default-pool"}
			result := formatNodepoolValue(&nodeRow)
			Expect(result).To(Equal("default-pool"))
		})

		It("should return noDataFormat when empty", func() {
			nodeRow := NodesAlignedRow{Nodepool: ""}
			result := formatNodepoolValue(&nodeRow)
			Expect(result).To(Equal(noDataFormat))
		})
	})
})
