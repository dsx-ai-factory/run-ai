/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"bytes"
	"os"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/environment/list/manipulate"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/settings"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/spf13/cobra"

	listCommon "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/list"
)

var _ = Describe("Output", func() {
	var (
		output       *Output
		environments []openapi.EnvironmentAsset
	)

	BeforeEach(func() {
		// Create sample environment data using helper
		env1 := newTestEnvironmentAsset("env-1", "ubuntu:20.04", []string{"JupyterNotebook"}, true, true, false, false, "")
		env2 := newTestEnvironmentAsset("env-2", "pytorch:latest", []string{"Tensorboard"}, false, true, false, false, "")
		// Override env2 scope to Project for testing different scopes
		env2.Meta.Scope = openapi.Scope("Project")
		environments = []openapi.EnvironmentAsset{env1, env2}

		response := &openapi.EnvironmentListResponse{Entries: environments}
		var err error
		output, err = ParseListResponse(response)
		Expect(err).ToNot(HaveOccurred())
	})

	Context("ParseListResponse", func() {
		It("should parse environment list response successfully", func() {
			Expect(output).ToNot(BeNil())
			Expect(output.data.Environments).To(HaveLen(2))
		})

		It("should convert assets to brief environments on-demand", func() {
			briefEnvs := output.GetList()
			Expect(briefEnvs).To(HaveLen(2))

			Expect(briefEnvs[0].Name).To(Equal("env-1"))
			Expect(briefEnvs[0].Scope).To(Equal("Cluster"))
			Expect(briefEnvs[0].Image).To(Equal("ubuntu:20.04"))
			Expect(briefEnvs[0].Tools).To(Equal("jupyter-notebook"))

			Expect(briefEnvs[1].Name).To(Equal("env-2"))
			Expect(briefEnvs[1].Scope).To(Equal("Project"))
			Expect(briefEnvs[1].Image).To(Equal("pytorch:latest"))
		})

		It("should handle empty environment list", func() {
			response := &openapi.EnvironmentListResponse{Entries: []openapi.EnvironmentAsset{}}
			output, err := ParseListResponse(response)

			Expect(err).ToNot(HaveOccurred())
			Expect(output.data.Environments).To(HaveLen(0))
			Expect(output.GetList()).To(HaveLen(0))
		})
	})

	Context("SetColumns functionality", func() {
		It("should set custom columns successfully", func() {
			columns := []string{"name", "scope", "image"}

			err := output.SetColumns(columns)

			Expect(err).ToNot(HaveOccurred())
			Expect(output.columns).To(HaveLen(3))
			Expect(output.columns).To(ContainElements(
				manipulate.ColumnKeyName,
				manipulate.ColumnKeyScope,
				manipulate.ColumnKeyImage,
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
			columns := []string{"name", "invalid-column", "scope"}

			err := output.SetColumns(columns)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid column name"))
		})

		It("should handle case insensitive columns", func() {
			columns := []string{"Name", "Scope", "Image"}

			err := output.SetColumns(columns)

			Expect(err).ToNot(HaveOccurred())
			Expect(output.columns).To(HaveLen(3))
		})

		It("should handle whitespace in columns", func() {
			columns := []string{"  name  ", " scope ", "image"}

			err := output.SetColumns(columns)

			Expect(err).ToNot(HaveOccurred())
			Expect(output.columns).To(HaveLen(3))
		})
	})

	Context("Columns method with custom columns", func() {
		It("should return custom columns when set", func() {
			columns := []string{"name", "scope"}
			err := output.SetColumns(columns)
			Expect(err).ToNot(HaveOccurred())

			tableColumns := output.Columns()

			Expect(tableColumns).To(HaveLen(2))
			Expect(tableColumns[0].Title).To(Equal("Name"))
			Expect(tableColumns[1].Title).To(Equal("Scope"))
		})

		It("should return default columns when no custom columns set", func() {
			tableColumns := output.Columns()

			defaultColumns := manipulate.GetDefaultColumns()
			Expect(tableColumns).To(HaveLen(len(defaultColumns)))
		})

		It("should return custom columns with correct widths", func() {
			columns := []string{"name", "scope"}
			err := output.SetColumns(columns)
			Expect(err).ToNot(HaveOccurred())

			tableColumns := output.Columns()

			Expect(tableColumns).To(HaveLen(2))
			Expect(tableColumns[0].Width).To(Equal(manipulate.DefaultWidth))
			Expect(tableColumns[1].Width).To(Equal(manipulate.DefaultWidth))
		})
	})

	Context("Data method with custom columns", func() {
		It("should return data for custom columns", func() {
			columns := []string{"name", "scope"}
			err := output.SetColumns(columns)
			Expect(err).ToNot(HaveOccurred())

			data := output.Data()

			Expect(data).To(HaveLen(2))
			// Check first environment (sorted by name)
			Expect(data[0]).To(HaveLen(2))
			Expect(data[0][0]).To(Equal("env-1"))
			Expect(data[0][1]).To(Equal("Cluster"))
			// Check second environment
			Expect(data[1]).To(HaveLen(2))
			Expect(data[1][0]).To(Equal("env-2"))
			Expect(data[1][1]).To(Equal("Project"))
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
			Expect(data[0][0]).To(Equal("env-1"))
			Expect(data[1]).To(HaveLen(1))
			Expect(data[1][0]).To(Equal("env-2"))
		})

		It("should handle all available columns", func() {
			columns := []string{"name", "scope", "image", "tools", "workloadtypes"}
			err := output.SetColumns(columns)
			Expect(err).ToNot(HaveOccurred())

			data := output.Data()

			Expect(data).To(HaveLen(2))
			Expect(data[0]).To(HaveLen(5))
			Expect(data[1]).To(HaveLen(5))

			// Check specific values for first environment
			Expect(data[0][0]).To(Equal("env-1"))                // name
			Expect(data[0][1]).To(Equal("Cluster"))              // scope
			Expect(data[0][2]).To(Equal("ubuntu:20.04"))         // image
			Expect(data[0][3]).To(Equal("jupyter-notebook"))     // tools
			Expect(data[0][4]).To(ContainSubstring("Workspace")) // workloadtypes
		})
	})

	Context("getActiveColumns method", func() {
		It("should return custom columns when set", func() {
			columns := []string{"name", "scope"}
			err := output.SetColumns(columns)
			Expect(err).ToNot(HaveOccurred())

			activeColumns := output.getActiveColumns()

			Expect(activeColumns).To(HaveLen(2))
			Expect(activeColumns).To(ContainElements(
				manipulate.ColumnKeyName,
				manipulate.ColumnKeyScope,
			))
		})

		It("should return default columns when no custom columns", func() {
			activeColumns := output.getActiveColumns()

			expectedColumns := manipulate.GetDefaultColumns()
			Expect(activeColumns).To(Equal(expectedColumns))
		})

		It("should return default columns when custom columns is empty", func() {
			output.columns = []manipulate.EnvironmentColumn{}

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
			Expect(jsonOutput).To(ContainSubstring("\"meta\""))
			Expect(jsonOutput).To(ContainSubstring("\"spec\""))
			Expect(jsonOutput).To(ContainSubstring("\"environments\""))
		})

		It("should return raw data when raw=true for YAML", func() {
			output.SetRaw(true)

			yamlOutput, err := output.GenerateYaml()
			Expect(err).ToNot(HaveOccurred())

			// Should contain original API fields
			Expect(yamlOutput).To(ContainSubstring("meta:"))
			Expect(yamlOutput).To(ContainSubstring("spec:"))
		})

		It("should return non-raw data when raw=false", func() {
			output.SetRaw(false)

			jsonOutput, err := output.GenerateJson()
			Expect(err).ToNot(HaveOccurred())

			Expect(jsonOutput).To(ContainSubstring("\"environments\""))

			output.SetRaw(true)
			rawOutput, err := output.GenerateJson()
			Expect(err).ToNot(HaveOccurred())

			Expect(jsonOutput).ToNot(Equal(rawOutput))
		})
	})

	Context("Subject method", func() {
		It("should return environment subject", func() {
			subject := output.Subject()
			Expect(subject).To(Equal("environment"))
		})
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

var _ = Describe("HandleColumnsFlags", func() {
	var (
		output  listCommon.Output
		cmd     *cobra.Command
		buf     *bytes.Buffer
		tempDir string
	)

	BeforeEach(func() {
		output = listCommon.Output{}
		cmd = &cobra.Command{}
		buf = new(bytes.Buffer)
		cmd.SetOut(buf)

		// Create temporary directory for settings
		var err error
		tempDir, err = os.MkdirTemp("", "environment-list-test-settings-*")
		Expect(err).ToNot(HaveOccurred())

		// Initialize settings with temp directory
		settingsInstance := settings.GetInstance()
		err = settingsInstance.Initializing(tempDir)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		// Clean up temporary directory
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
	})

	Context("Valid columns with set-default-columns flag", func() {
		It("should save valid columns to settings", func() {
			setOutputValues(&output, []string{"name", "scope", "image"}, true)

			err := HandleColumnsFlags(output, cmd, "Default environment columns set to: %s\n")

			Expect(err).ToNot(HaveOccurred())
			savedColumns := settings.GetInstance().GetEnvironmentColumns()
			Expect(savedColumns).To(Equal([]string{"name", "scope", "image"}))
			Expect(buf.String()).To(ContainSubstring("Default environment columns set to"))
		})
	})

	Context("Invalid columns", func() {
		It("should return error for invalid columns", func() {
			setOutputValues(&output, []string{"invalid-column"}, true)

			err := HandleColumnsFlags(output, cmd, "Default environment columns set to: %s\n")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid column name"))
		})
	})

	Context("Valid columns without set-default-columns flag", func() {
		It("should validate columns but not save to settings", func() {
			initialColumns := settings.GetInstance().GetEnvironmentColumns()
			initialLen := len(initialColumns)

			setOutputValues(&output, []string{"name", "tools"}, false)

			err := HandleColumnsFlags(output, cmd, "Default environment columns set to: %s\n")

			Expect(err).ToNot(HaveOccurred())
			savedColumns := settings.GetInstance().GetEnvironmentColumns()
			Expect(len(savedColumns)).To(Equal(initialLen))
			Expect(buf.String()).To(BeEmpty())
		})
	})
})

// Helper function to set output values
func setOutputValues(output *listCommon.Output, columns []string, setDefaultColumns bool) {
	cmd := &cobra.Command{}
	listCommon.FlagOutputType(cmd, output)
	listCommon.FlagOutputTypeWithColumns(cmd, output)

	if len(columns) > 0 {
		_ = cmd.Flags().Set(listCommon.ColumnsStr, strings.Join(columns, ","))
	}
	if setDefaultColumns {
		_ = cmd.Flags().Set(listCommon.SetDefaultColumnsStr, "true")
	}
}
