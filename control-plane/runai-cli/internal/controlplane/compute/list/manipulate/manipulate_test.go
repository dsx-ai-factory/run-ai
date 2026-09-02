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

var _ = Describe("ParseComputeColumns", func() {
	Context("when columns slice is empty", func() {
		It("should return empty slice", func() {
			result, err := ParseComputeColumns([]string{})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(0))
		})
	})

	Context("when columns contain valid values", func() {
		It("should parse single valid column", func() {
			result, err := ParseComputeColumns([]string{"name"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(1))
			Expect(result[0]).To(Equal(ColumnKeyName))
		})

		It("should parse multiple valid columns", func() {
			result, err := ParseComputeColumns([]string{"name", "scope", "gpudevices"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyScope, ColumnKeyGPUDevices))
		})

		It("should handle case insensitive input", func() {
			result, err := ParseComputeColumns([]string{"NAME", "Scope", "GPUDevices"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyScope, ColumnKeyGPUDevices))
		})

		It("should trim whitespace", func() {
			result, err := ParseComputeColumns([]string{"  name  ", " scope ", "gpudevices"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyScope, ColumnKeyGPUDevices))
		})

		It("should parse all valid columns", func() {
			allColumnNames := []string{"name", "scope", "gpudevices", "gpumemory", "cpumemory"}
			result, err := ParseComputeColumns(allColumnNames)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(5))
		})
	})

	Context("when columns contain invalid values", func() {
		It("should return error for single invalid column", func() {
			result, err := ParseComputeColumns([]string{"invalid"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid column name"))
			Expect(result).To(BeNil())
		})

		It("should return error for multiple invalid columns", func() {
			result, err := ParseComputeColumns([]string{"invalid1", "invalid2"})
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("should return error when mixing valid and invalid columns", func() {
			result, err := ParseComputeColumns([]string{"name", "invalid", "scope"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid column name"))
			Expect(result).To(BeNil())
		})
	})

	Context("when columns contain duplicates", func() {
		It("should include duplicates in result", func() {
			result, err := ParseComputeColumns([]string{"name", "name", "scope"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(3))
			Expect(result[0]).To(Equal(ColumnKeyName))
			Expect(result[1]).To(Equal(ColumnKeyName))
			Expect(result[2]).To(Equal(ColumnKeyScope))
		})
	})
})

var _ = Describe("CreateAlignedComputes", func() {
	var (
		computes []openapi.ComputeAsset
		columns  []ComputeColumn
	)

	BeforeEach(func() {
		compute1 := newTestComputeAsset("compute-large", "Project", 8, "80GB", "128GB")
		compute2 := newTestComputeAsset("compute-small", "Cluster", 2, "16GB", "32GB")
		computes = []openapi.ComputeAsset{compute1, compute2}
		columns = []ComputeColumn{ColumnKeyName, ColumnKeyScope, ColumnKeyGPUDevices}
	})

	It("should create aligned computes with specified columns", func() {
		result := CreateAlignedComputes(computes, columns)

		Expect(result.Computes).To(HaveLen(2))
		Expect(result.Computes[0].Name).To(Equal("compute-large"))
		Expect(result.Computes[0].Scope).To(Equal("Project"))
		Expect(result.Computes[0].GPUDevices).To(Equal("8"))

		Expect(result.Computes[1].Name).To(Equal("compute-small"))
		Expect(result.Computes[1].Scope).To(Equal("Cluster"))
		Expect(result.Computes[1].GPUDevices).To(Equal("2"))
	})

	It("should only include selected columns", func() {
		columns = []ComputeColumn{ColumnKeyName}
		result := CreateAlignedComputes(computes, columns)

		Expect(result.Computes).To(HaveLen(2))
		Expect(result.Computes[0].Name).To(Equal("compute-large"))
		Expect(result.Computes[0].Scope).To(BeEmpty())
		Expect(result.Computes[0].GPUDevices).To(BeEmpty())
	})

	It("should handle empty computes list", func() {
		computes = []openapi.ComputeAsset{}
		result := CreateAlignedComputes(computes, columns)

		Expect(result.Computes).To(HaveLen(0))
	})

	It("should handle all columns", func() {
		columns = []ComputeColumn{
			ColumnKeyName, ColumnKeyScope, ColumnKeyGPUDevices,
			ColumnKeyGPUMemory, ColumnKeyCPUMemory,
		}
		result := CreateAlignedComputes(computes, columns)

		Expect(result.Computes).To(HaveLen(2))

		// Verify formatted output via CreateTableData
		tableData := CreateTableData(result.Computes, columns)
		Expect(tableData[0][0]).To(Equal("compute-large")) // Name
		Expect(tableData[0][1]).To(Equal("Project"))       // Scope
		Expect(tableData[0][2]).To(Equal("8"))             // GPU Devices
		Expect(tableData[0][3]).To(Equal("80GB"))          // GPU Memory
		Expect(tableData[0][4]).To(Equal("128GB"))         // CPU Memory
	})
})

var _ = Describe("CreateTableData", func() {
	var (
		alignedComputes []ComputesAlignedRow
		columns         []ComputeColumn
	)

	BeforeEach(func() {
		compute1 := newTestComputeAsset("compute-large", "Project", 8, "80GB", "128GB")
		compute2 := newTestComputeAsset("compute-small", "Cluster", 2, "16GB", "32GB")
		alignedComputes = []ComputesAlignedRow{
			{
				Name:       "compute-large",
				Scope:      "Project",
				GPUDevices: "8",
				GPUMemory:  "80GB",
				CPUMemory:  "128GB",
				Spec:       compute1.Spec,
			},
			{
				Name:       "compute-small",
				Scope:      "Cluster",
				GPUDevices: "2",
				GPUMemory:  "16GB",
				CPUMemory:  "32GB",
				Spec:       compute2.Spec,
			},
		}
		columns = []ComputeColumn{ColumnKeyName, ColumnKeyScope, ColumnKeyGPUDevices}
	})

	It("should create table data in correct format", func() {
		result := CreateTableData(alignedComputes, columns)

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(Equal([]string{"compute-large", "Project", "8"}))
		Expect(result[1]).To(Equal([]string{"compute-small", "Cluster", "2"}))
	})

	It("should handle different column order", func() {
		columns = []ComputeColumn{ColumnKeyScope, ColumnKeyName, ColumnKeyGPUDevices}
		result := CreateTableData(alignedComputes, columns)

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(Equal([]string{"Project", "compute-large", "8"}))
		Expect(result[1]).To(Equal([]string{"Cluster", "compute-small", "2"}))
	})

	It("should handle single column", func() {
		columns = []ComputeColumn{ColumnKeyName}
		result := CreateTableData(alignedComputes, columns)

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(Equal([]string{"compute-large"}))
		Expect(result[1]).To(Equal([]string{"compute-small"}))
	})

	It("should handle empty aligned computes", func() {
		alignedComputes = []ComputesAlignedRow{}
		result := CreateTableData(alignedComputes, columns)

		Expect(result).To(HaveLen(0))
	})

	It("should handle all columns", func() {
		columns = []ComputeColumn{
			ColumnKeyName, ColumnKeyScope, ColumnKeyGPUDevices,
			ColumnKeyGPUMemory, ColumnKeyCPUMemory,
		}
		result := CreateTableData(alignedComputes, columns)

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(HaveLen(5))
		Expect(result[0][0]).To(Equal("compute-large"))
		Expect(result[0][1]).To(Equal("Project"))
		Expect(result[0][2]).To(Equal("8"))
		Expect(result[0][3]).To(Equal("80GB"))
		Expect(result[0][4]).To(Equal("128GB"))
	})
})

var _ = Describe("GetDefaultColumns", func() {
	It("should return all expected default columns", func() {
		result := GetDefaultColumns()

		expectedColumns := []ComputeColumn{
			ColumnKeyName, ColumnKeyScope, ColumnKeyGPUDevices,
			ColumnKeyGPUMemory, ColumnKeyCPUMemory,
		}
		Expect(result).To(Equal(expectedColumns))
	})
})

var _ = Describe("GetColumns", func() {
	It("should return provided columns when not empty", func() {
		customColumns := []ComputeColumn{ColumnKeyName, ColumnKeyScope}
		result := GetColumns(customColumns)

		Expect(result).To(Equal(customColumns))
	})

	It("should return default columns when empty", func() {
		result := GetColumns([]ComputeColumn{})

		expectedColumns := GetDefaultColumns()
		Expect(result).To(Equal(expectedColumns))
	})

	It("should return default columns when nil", func() {
		result := GetColumns(nil)

		expectedColumns := GetDefaultColumns()
		Expect(result).To(Equal(expectedColumns))
	})
})

var _ = Describe("BriefComputeFromAsset", func() {
	It("should convert compute asset to brief compute", func() {
		compute := newTestComputeAsset("test-compute", "Project", 4, "40GB", "64GB")

		result := BriefComputeFromAsset(compute)

		Expect(result.Name).To(Equal("test-compute"))
		Expect(result.Scope).To(Equal("Project"))
		Expect(result.GPUDevices).To(Equal("4"))
		Expect(result.GPUMemory).To(Equal("40GB"))
		Expect(result.CPUMemory).To(Equal("64GB"))
	})

	It("should handle missing GPU devices", func() {
		compute := newTestComputeAsset("test-compute", "Project", 0, "", "64GB")

		result := BriefComputeFromAsset(compute)

		Expect(result.GPUDevices).To(Equal(""))
	})

	It("should handle missing GPU memory", func() {
		compute := newTestComputeAsset("test-compute", "Project", 4, "", "64GB")

		result := BriefComputeFromAsset(compute)

		Expect(result.GPUMemory).To(Equal(""))
	})

	It("should handle missing CPU memory", func() {
		compute := newTestComputeAsset("test-compute", "Project", 4, "40GB", "")

		result := BriefComputeFromAsset(compute)

		Expect(result.CPUMemory).To(Equal(""))
	})

	It("should handle GPU portion request", func() {
		compute := newTestComputeAssetWithPortion("test-compute", "Project", 0.5)

		result := BriefComputeFromAsset(compute)

		Expect(result.GPUMemory).To(Equal("50%"))
	})

	It("should handle 100% GPU portion request", func() {
		compute := newTestComputeAssetWithPortion("test-compute", "Project", 1.0)

		result := BriefComputeFromAsset(compute)

		Expect(result.GPUMemory).To(Equal("100%"))
	})
})

// Helper functions to create test compute assets
func newTestComputeAsset(name, scope string, gpuDevices int32, gpuMemory, cpuMemory string) openapi.ComputeAsset {
	spec := openapi.ComputeAssetSpec{}

	if gpuDevices > 0 {
		spec.SetGpuDevicesRequest(gpuDevices)
	}

	if gpuMemory != "" {
		spec.SetGpuMemoryRequest(gpuMemory)
	}

	if cpuMemory != "" {
		spec.SetCpuMemoryRequest(cpuMemory)
	}

	return openapi.ComputeAsset{
		Meta: openapi.AssetMeta{
			Name:  name,
			Scope: openapi.Scope(scope),
		},
		Spec: spec,
	}
}

func newTestComputeAssetWithPortion(name, scope string, gpuPortion float64) openapi.ComputeAsset {
	spec := openapi.ComputeAssetSpec{}
	spec.SetGpuPortionRequest(gpuPortion)

	return openapi.ComputeAsset{
		Meta: openapi.AssetMeta{
			Name:  name,
			Scope: openapi.Scope(scope),
		},
		Spec: spec,
	}
}
