/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package manipulate

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
)

var _ = Describe("CreateAlignedNodes", func() {
	var (
		testData Data
		columns  []NodeColumn
	)

	BeforeEach(func() {
		testData = Data{
			Nodes: []openapi.Node{
				{
					Name:     "node1",
					Status:   "Ready",
					NodePool: "pool1",
				},
				{
					Name:     "node2",
					Status:   "NotReady",
					NodePool: "pool2",
				},
			},
		}
		columns = []NodeColumn{ColumnKeyName, ColumnKeyStatus, ColumnKeyNodepool}
	})

	It("should create aligned nodes with specified columns", func() {
		result := CreateAlignedNodes(testData, columns)

		Expect(result.Nodes).To(HaveLen(2))
		Expect(result.Nodes[0].Name).To(Equal("node1"))
		Expect(result.Nodes[0].Status).To(Equal("Ready"))
		Expect(result.Nodes[0].Nodepool).To(Equal("pool1"))

		Expect(result.Nodes[1].Name).To(Equal("node2"))
		Expect(result.Nodes[1].Status).To(Equal("NotReady"))
		Expect(result.Nodes[1].Nodepool).To(Equal("pool2"))
	})

	It("should preserve NextToken", func() {
		nextToken := "test-token"
		testData.NextToken = &nextToken

		result := CreateAlignedNodes(testData, columns)

		Expect(result.NextToken).NotTo(BeNil())
		Expect(*result.NextToken).To(Equal("test-token"))
	})

	It("should handle empty nodes list", func() {
		testData.Nodes = []openapi.Node{}

		result := CreateAlignedNodes(testData, columns)

		Expect(result.Nodes).To(HaveLen(0))
	})
})

var _ = Describe("CreateTableData", func() {
	var (
		alignedNodes []NodesAlignedRow
		columns      []NodeColumn
	)

	BeforeEach(func() {
		alignedNodes = []NodesAlignedRow{
			{
				Name:     "node2",
				Status:   "NotReady",
				Nodepool: "pool2",
			},
			{
				Name:     "node1",
				Status:   "Ready",
				Nodepool: "pool1",
			},
		}
		columns = []NodeColumn{ColumnKeyName, ColumnKeyStatus, ColumnKeyNodepool}
	})

	It("should create table data in correct format", func() {
		result := CreateTableData(alignedNodes, columns)

		Expect(result).To(HaveLen(2))
		// Should be sorted by name (first column)
		Expect(result[0]).To(Equal([]string{"node1", "Ready", "pool1"}))
		Expect(result[1]).To(Equal([]string{"node2", "NotReady", "pool2"}))
	})

	It("should handle different column order", func() {
		columns = []NodeColumn{ColumnKeyStatus, ColumnKeyName, ColumnKeyNodepool}

		result := CreateTableData(alignedNodes, columns)

		Expect(result).To(HaveLen(2))
		// Data should follow column order, still sorted by name (which is now at index 1)
		// So node1 should come first, node2 second
		Expect(result[0]).To(Equal([]string{"Ready", "node1", "pool1"}))
		Expect(result[1]).To(Equal([]string{"NotReady", "node2", "pool2"}))
	})

	It("should handle empty aligned nodes", func() {
		alignedNodes = []NodesAlignedRow{}

		result := CreateTableData(alignedNodes, columns)

		Expect(result).To(HaveLen(0))
	})
})

var _ = Describe("GetDefaultColumns function", func() {
	It("should return all expected default columns", func() {
		result := GetDefaultColumns()

		expectedColumns := []NodeColumn{
			ColumnKeyName, ColumnKeyStatus, ColumnKeyGpuType,
			ColumnKeyGpuAllocation, ColumnKeyFreeGpu, ColumnKeyNodepool,
		}
		Expect(result).To(Equal(expectedColumns))
	})
})

var _ = Describe("GetColumns function", func() {
	It("should return provided columns when not empty", func() {
		customColumns := []NodeColumn{ColumnKeyName, ColumnKeyStatus}
		result := GetColumns(customColumns)

		Expect(result).To(Equal(customColumns))
	})

	It("should return default columns when empty", func() {
		result := GetColumns([]NodeColumn{})

		expectedColumns := GetDefaultColumns()
		Expect(result).To(Equal(expectedColumns))
	})

	It("should return default columns when nil", func() {
		result := GetColumns(nil)

		expectedColumns := GetDefaultColumns()
		Expect(result).To(Equal(expectedColumns))
	})
})
