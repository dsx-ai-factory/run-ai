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

var _ = Describe("ParseEnvironmentColumns", func() {
	Context("when columns slice is empty", func() {
		It("should return empty slice", func() {
			result, err := ParseEnvironmentColumns([]string{})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(0))
		})
	})

	Context("when columns contain valid values", func() {
		It("should parse single valid column", func() {
			result, err := ParseEnvironmentColumns([]string{"name"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(1))
			Expect(result[0]).To(Equal(ColumnKeyName))
		})

		It("should parse multiple valid columns", func() {
			result, err := ParseEnvironmentColumns([]string{"name", "scope", "image"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyScope, ColumnKeyImage))
		})

		It("should handle case insensitive input", func() {
			result, err := ParseEnvironmentColumns([]string{"NAME", "Scope", "IMAGE"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyScope, ColumnKeyImage))
		})

		It("should trim whitespace", func() {
			result, err := ParseEnvironmentColumns([]string{"  name  ", " scope ", "image"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(3))
			Expect(result).To(ContainElements(ColumnKeyName, ColumnKeyScope, ColumnKeyImage))
		})

		It("should parse all valid columns", func() {
			allColumnNames := []string{"name", "scope", "image", "tools", "workloadtypes"}
			result, err := ParseEnvironmentColumns(allColumnNames)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(5))
		})
	})

	Context("when columns contain invalid values", func() {
		It("should return error for single invalid column", func() {
			result, err := ParseEnvironmentColumns([]string{"invalid"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid column name"))
			Expect(result).To(BeNil())
		})

		It("should return error for multiple invalid columns", func() {
			result, err := ParseEnvironmentColumns([]string{"invalid1", "invalid2"})
			Expect(err).To(HaveOccurred())
			Expect(result).To(BeNil())
		})

		It("should return error when mixing valid and invalid columns", func() {
			result, err := ParseEnvironmentColumns([]string{"name", "invalid", "scope"})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid column name"))
			Expect(result).To(BeNil())
		})
	})

	Context("when columns contain duplicates", func() {
		It("should include duplicates in result", func() {
			result, err := ParseEnvironmentColumns([]string{"name", "name", "scope"})
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(HaveLen(3))
			Expect(result[0]).To(Equal(ColumnKeyName))
			Expect(result[1]).To(Equal(ColumnKeyName))
			Expect(result[2]).To(Equal(ColumnKeyScope))
		})
	})
})

var _ = Describe("CreateAlignedEnvironments", func() {
	var (
		environments []openapi.EnvironmentAsset
		columns      []EnvironmentColumn
	)

	BeforeEach(func() {
		env1 := newTestEnvironmentAsset("env1", "ubuntu:20.04", []string{"JupyterNotebook"}, true, true, false, false, "")
		env2 := newTestEnvironmentAsset("env2", "pytorch:latest", []string{"Tensorboard"}, false, true, false, false, "")
		env2.Meta.Scope = openapi.Scope("Project")
		environments = []openapi.EnvironmentAsset{env1, env2}
		columns = []EnvironmentColumn{ColumnKeyName, ColumnKeyScope, ColumnKeyImage}
	})

	It("should create aligned environments with specified columns", func() {
		result := CreateAlignedEnvironments(environments, columns)

		Expect(result.Environments).To(HaveLen(2))
		Expect(result.Environments[0].Name).To(Equal("env1"))
		Expect(result.Environments[0].Scope).To(Equal("Cluster"))
		Expect(result.Environments[0].Image).To(Equal("ubuntu:20.04"))

		Expect(result.Environments[1].Name).To(Equal("env2"))
		Expect(result.Environments[1].Scope).To(Equal("Project"))
		Expect(result.Environments[1].Image).To(Equal("pytorch:latest"))
	})

	It("should only include selected columns", func() {
		columns = []EnvironmentColumn{ColumnKeyName}
		result := CreateAlignedEnvironments(environments, columns)

		Expect(result.Environments).To(HaveLen(2))
		Expect(result.Environments[0].Name).To(Equal("env1"))
		Expect(result.Environments[0].Scope).To(BeEmpty())
		Expect(result.Environments[0].Image).To(BeEmpty())
	})

	It("should handle empty environments list", func() {
		environments = []openapi.EnvironmentAsset{}
		result := CreateAlignedEnvironments(environments, columns)

		Expect(result.Environments).To(HaveLen(0))
	})

	It("should handle all columns", func() {
		columns = []EnvironmentColumn{
			ColumnKeyName, ColumnKeyScope, ColumnKeyImage,
			ColumnKeyTools, ColumnKeyWorkloadTypes,
		}
		result := CreateAlignedEnvironments(environments, columns)

		Expect(result.Environments).To(HaveLen(2))
		// Check that raw fields are populated
		Expect(result.Environments[0].RawConnections).ToNot(BeEmpty())
		Expect(result.Environments[0].RawWorkloadTypes.IsSet()).To(BeTrue())

		// Verify formatted output via CreateTableData
		tableData := CreateTableData(result.Environments, columns)
		Expect(tableData[0][3]).To(Equal("jupyter-notebook"))     // Tools column
		Expect(tableData[0][4]).To(ContainSubstring("Workspace")) // WorkloadTypes column
		Expect(tableData[0][4]).To(ContainSubstring("Training"))
	})
})

var _ = Describe("CreateTableData", func() {
	var (
		alignedEnvs []EnvironmentsAlignedRow
		columns     []EnvironmentColumn
	)

	BeforeEach(func() {
		// Create test connections for tools
		conn1 := openapi.Connection{}
		conn1.SetInternalToolInfo(openapi.InternalToolInfo{
			ToolType: openapi.InternalToolType("JupyterNotebook"),
		})
		conn2 := openapi.Connection{}
		conn2.SetInternalToolInfo(openapi.InternalToolInfo{
			ToolType: openapi.InternalToolType("Tensorboard"),
		})

		// Create test workload types
		trueVal := true
		workloadTypes1 := openapi.WorkloadSupportedTypes{}
		workloadTypes1.Workspace = *openapi.NewNullableBool(&trueVal)
		workloadTypes2 := openapi.WorkloadSupportedTypes{}
		workloadTypes2.Training = *openapi.NewNullableBool(&trueVal)

		alignedEnvs = []EnvironmentsAlignedRow{
			{
				Name:             "env1",
				Scope:            "cluster",
				Image:            "ubuntu:20.04",
				RawConnections:   []openapi.Connection{conn1},
				RawWorkloadTypes: *openapi.NewNullableWorkloadSupportedTypes(&workloadTypes1),
			},
			{
				Name:             "env2",
				Scope:            "project",
				Image:            "pytorch:latest",
				RawConnections:   []openapi.Connection{conn2},
				RawWorkloadTypes: *openapi.NewNullableWorkloadSupportedTypes(&workloadTypes2),
			},
		}
		columns = []EnvironmentColumn{ColumnKeyName, ColumnKeyScope, ColumnKeyImage}
	})

	It("should create table data in correct format", func() {
		result := CreateTableData(alignedEnvs, columns)

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(Equal([]string{"env1", "cluster", "ubuntu:20.04"}))
		Expect(result[1]).To(Equal([]string{"env2", "project", "pytorch:latest"}))
	})

	It("should handle different column order", func() {
		columns = []EnvironmentColumn{ColumnKeyScope, ColumnKeyName, ColumnKeyImage}
		result := CreateTableData(alignedEnvs, columns)

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(Equal([]string{"cluster", "env1", "ubuntu:20.04"}))
		Expect(result[1]).To(Equal([]string{"project", "env2", "pytorch:latest"}))
	})

	It("should handle single column", func() {
		columns = []EnvironmentColumn{ColumnKeyName}
		result := CreateTableData(alignedEnvs, columns)

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(Equal([]string{"env1"}))
		Expect(result[1]).To(Equal([]string{"env2"}))
	})

	It("should handle empty aligned environments", func() {
		alignedEnvs = []EnvironmentsAlignedRow{}
		result := CreateTableData(alignedEnvs, columns)

		Expect(result).To(HaveLen(0))
	})

	It("should handle all columns", func() {
		columns = []EnvironmentColumn{
			ColumnKeyName, ColumnKeyScope, ColumnKeyImage,
			ColumnKeyTools, ColumnKeyWorkloadTypes,
		}
		result := CreateTableData(alignedEnvs, columns)

		Expect(result).To(HaveLen(2))
		Expect(result[0]).To(HaveLen(5))
		Expect(result[0][0]).To(Equal("env1"))
		Expect(result[0][1]).To(Equal("cluster"))
		Expect(result[0][2]).To(Equal("ubuntu:20.04"))
		Expect(result[0][3]).To(Equal("jupyter-notebook"))
		Expect(result[0][4]).To(Equal("Workspace"))
	})
})

var _ = Describe("GetDefaultColumns", func() {
	It("should return all expected default columns", func() {
		result := GetDefaultColumns()

		expectedColumns := []EnvironmentColumn{
			ColumnKeyName, ColumnKeyScope, ColumnKeyImage,
			ColumnKeyTools, ColumnKeyWorkloadTypes,
		}
		Expect(result).To(Equal(expectedColumns))
	})
})

var _ = Describe("GetColumns", func() {
	It("should return provided columns when not empty", func() {
		customColumns := []EnvironmentColumn{ColumnKeyName, ColumnKeyScope}
		result := GetColumns(customColumns)

		Expect(result).To(Equal(customColumns))
	})

	It("should return default columns when empty", func() {
		result := GetColumns([]EnvironmentColumn{})

		expectedColumns := GetDefaultColumns()
		Expect(result).To(Equal(expectedColumns))
	})

	It("should return default columns when nil", func() {
		result := GetColumns(nil)

		expectedColumns := GetDefaultColumns()
		Expect(result).To(Equal(expectedColumns))
	})
})

var _ = Describe("BriefEnvironmentFromAsset", func() {
	It("should convert environment asset to brief environment", func() {
		env := newTestEnvironmentAsset("test-env", "ubuntu:20.04", []string{}, false, false, false, false, "")

		result := BriefEnvironmentFromAsset(env)

		Expect(result.Name).To(Equal("test-env"))
		Expect(result.Image).To(Equal("ubuntu:20.04"))
	})

	It("should handle missing image", func() {
		env := newTestEnvironmentAsset("test-env", "", []string{}, false, false, false, false, "")

		result := BriefEnvironmentFromAsset(env)

		// When image is empty, extractImage returns noDataFormat ("-")
		Expect(result.Image).To(Or(Equal(noDataFormat), Equal("")))
	})

	It("should extract tools from connections", func() {
		env := newTestEnvironmentAsset("test-env", "ubuntu:20.04", []string{"JupyterNotebook"}, false, false, false, false, "")

		result := BriefEnvironmentFromAsset(env)

		Expect(result.Tools).To(Equal("jupyter-notebook"))
	})

	It("should extract workload types", func() {
		env := newTestEnvironmentAsset("test-env", "ubuntu:20.04", []string{}, true, true, false, false, "")

		result := BriefEnvironmentFromAsset(env)

		Expect(result.WorkloadTypes).To(ContainSubstring("Workspace"))
		Expect(result.WorkloadTypes).To(ContainSubstring("Training"))
	})
})

// Helper function to create test environment assets
func newTestEnvironmentAsset(name, image string, tools []string, isWorkspace, isTraining, isInference, isDistributed bool, distFramework string) openapi.EnvironmentAsset {
	// Create connections for tools
	var connections []openapi.Connection
	for _, tool := range tools {
		toolType := openapi.InternalToolType(tool)
		conn := openapi.Connection{}
		internalTool := openapi.InternalToolInfo{
			ToolType: toolType,
		}
		conn.SetInternalToolInfo(internalTool)
		connections = append(connections, conn)
	}

	// Create workload supported types
	workloadTypes := openapi.WorkloadSupportedTypes{}
	if isWorkspace {
		workloadTypes.Workspace = *openapi.NewNullableBool(&isWorkspace)
	}
	if isTraining {
		workloadTypes.Training = *openapi.NewNullableBool(&isTraining)
	}
	if isInference {
		workloadTypes.Inference = *openapi.NewNullableBool(&isInference)
	}
	if isDistributed {
		workloadTypes.Distributed = *openapi.NewNullableBool(&isDistributed)
		if distFramework != "" {
			workloadTypes.DistFramework = *openapi.NewNullableString(&distFramework)
		}
	}

	spec := openapi.EnvironmentAssetSpec{}
	if image != "" {
		spec.SetImage(image)
	}
	if len(connections) > 0 {
		spec.SetConnections(connections)
	}

	return openapi.EnvironmentAsset{
		Meta: openapi.AssetMeta{
			Name:                   name,
			Scope:                  openapi.Scope("Cluster"),
			WorkloadSupportedTypes: *openapi.NewNullableWorkloadSupportedTypes(&workloadTypes),
		},
		Spec: spec,
	}
}
