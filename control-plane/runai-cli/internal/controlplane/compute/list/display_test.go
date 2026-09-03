/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"bytes"
	"os"
	"strings"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/compute/list/manipulate"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/settings"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	listCommon "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/list"
)

var _ = Describe("Output", func() {
	var (
		output   *Output
		computes []openapi.ComputeAsset
	)

	BeforeEach(func() {
		// Create sample compute data using helper
		compute1 := newTestComputeAsset("compute-large", "Project", 8, "80GB", "128GB")
		compute2 := newTestComputeAsset("compute-small", "Cluster", 2, "16GB", "32GB")
		computes = []openapi.ComputeAsset{compute1, compute2}

		response := &openapi.ComputeListResponse{Entries: computes}
		var err error
		output, err = ParseListResponse(response)
		Expect(err).ToNot(HaveOccurred())
	})

	Context("ParseListResponse", func() {
		It("should parse compute list response successfully", func() {
			Expect(output).ToNot(BeNil())
			Expect(output.GetList()).To(HaveLen(2))
			Expect(output.rawData).To(HaveLen(2))
		})

		It("should convert assets to brief computes", func() {
			briefComputes := output.GetList()
			Expect(briefComputes[0].Name).To(Equal("compute-large"))
			Expect(briefComputes[0].Scope).To(Equal("Project"))
			Expect(briefComputes[0].GPUDevices).To(Equal("8"))
			Expect(briefComputes[0].GPUMemory).To(Equal("80GB"))
			Expect(briefComputes[0].CPUMemory).To(Equal("128GB"))

			Expect(briefComputes[1].Name).To(Equal("compute-small"))
			Expect(briefComputes[1].Scope).To(Equal("Cluster"))
		})

		It("should handle empty compute list", func() {
			response := &openapi.ComputeListResponse{Entries: []openapi.ComputeAsset{}}
			output, err := ParseListResponse(response)

			Expect(err).ToNot(HaveOccurred())
			Expect(output.GetList()).To(HaveLen(0))
			Expect(output.rawData).To(HaveLen(0))
		})
	})

	Context("SetColumns functionality", func() {
		It("should set custom columns successfully", func() {
			columns := []string{"name", "scope", "gpudevices"}

			err := output.SetColumns(columns)

			Expect(err).ToNot(HaveOccurred())
			Expect(output.columns).To(HaveLen(3))
			Expect(output.columns).To(ContainElements(
				manipulate.ColumnKeyName,
				manipulate.ColumnKeyScope,
				manipulate.ColumnKeyGPUDevices,
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
			columns := []string{"Name", "Scope", "GPUDevices"}

			err := output.SetColumns(columns)

			Expect(err).ToNot(HaveOccurred())
			Expect(output.columns).To(HaveLen(3))
		})

		It("should handle whitespace in columns", func() {
			columns := []string{"  name  ", " scope ", "gpudevices"}

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
			Expect(tableColumns[0].Width).To(Equal(20))
			Expect(tableColumns[1].Width).To(Equal(10))
		})
	})

	Context("Data method with custom columns", func() {
		It("should return data for custom columns", func() {
			columns := []string{"name", "scope"}
			err := output.SetColumns(columns)
			Expect(err).ToNot(HaveOccurred())

			data := output.Data()

			Expect(data).To(HaveLen(2))
			// Check first compute (sorted by name)
			Expect(data[0]).To(HaveLen(2))
			Expect(data[0][0]).To(Equal("compute-large"))
			Expect(data[0][1]).To(Equal("Project"))
			// Check second compute
			Expect(data[1]).To(HaveLen(2))
			Expect(data[1][0]).To(Equal("compute-small"))
			Expect(data[1][1]).To(Equal("Cluster"))
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
			Expect(data[0][0]).To(Equal("compute-large"))
			Expect(data[1]).To(HaveLen(1))
			Expect(data[1][0]).To(Equal("compute-small"))
		})

		It("should handle all available columns", func() {
			columns := []string{"name", "scope", "gpudevices", "gpumemory", "cpumemory"}
			err := output.SetColumns(columns)
			Expect(err).ToNot(HaveOccurred())

			data := output.Data()

			Expect(data).To(HaveLen(2))
			Expect(data[0]).To(HaveLen(5))
			Expect(data[1]).To(HaveLen(5))

			// Check specific values for first compute
			Expect(data[0][0]).To(Equal("compute-large")) // name
			Expect(data[0][1]).To(Equal("Project"))       // scope
			Expect(data[0][2]).To(Equal("8"))             // gpudevicesperpod
			Expect(data[0][3]).To(Equal("80GB"))          // gpumemoryperpod
			Expect(data[0][4]).To(Equal("128GB"))         // cpumemory
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
			output.columns = []manipulate.ComputeColumn{}

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
			Expect(jsonOutput).To(ContainSubstring("\"computes\""))
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

			Expect(jsonOutput).To(ContainSubstring("\"computes\""))

			output.SetRaw(true)
			rawOutput, err := output.GenerateJson()
			Expect(err).ToNot(HaveOccurred())

			Expect(jsonOutput).ToNot(Equal(rawOutput))
		})
	})

	Context("Subject method", func() {
		It("should return compute subject", func() {
			subject := output.Subject()
			Expect(subject).To(Equal("compute"))
		})
	})
})

// Helper function to create test compute assets
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
		tempDir, err = os.MkdirTemp("", "compute-list-test-settings-*")
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
			setOutputValues(&output, []string{"name", "scope", "gpudevices"}, true)

			err := HandleColumnsFlags(output, cmd, "Default compute columns set to: %s\n")

			Expect(err).ToNot(HaveOccurred())
			savedColumns := settings.GetInstance().GetComputeColumns()
			Expect(savedColumns).To(Equal([]string{"name", "scope", "gpudevices"}))
			Expect(buf.String()).To(ContainSubstring("Default compute columns set to"))
		})
	})

	Context("Invalid columns", func() {
		It("should return error for invalid columns", func() {
			setOutputValues(&output, []string{"invalid-column"}, true)

			err := HandleColumnsFlags(output, cmd, "Default compute columns set to: %s\n")

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid column name"))
		})
	})

	Context("Valid columns without set-default-columns flag", func() {
		It("should validate columns but not save to settings", func() {
			initialColumns := settings.GetInstance().GetComputeColumns()
			initialLen := len(initialColumns)

			setOutputValues(&output, []string{"name", "scope"}, false)

			err := HandleColumnsFlags(output, cmd, "Default compute columns set to: %s\n")

			Expect(err).ToNot(HaveOccurred())
			savedColumns := settings.GetInstance().GetComputeColumns()
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
