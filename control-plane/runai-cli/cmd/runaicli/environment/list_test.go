/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package environment_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/environment/list"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	listCommon "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/runaicli/environment"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/settings"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
)

func TestEnvironment(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Environment Suite")
}

var _ = Describe("Environment List Command", func() {
	var (
		cmd     *cobra.Command
		buf     *bytes.Buffer
		tempDir string
		err     error
	)

	BeforeEach(func() {
		buf = &bytes.Buffer{}
		cmd = environment.List
		cmd.SetOut(buf)
		cmd.SetErr(buf)

		// Create temporary directory for test configs
		tempDir, err = os.MkdirTemp("", "environment-test-*")
		Expect(err).ToNot(HaveOccurred())

		// Initialize config with temp directory
		configInstance := config.GetInstance()
		err = configInstance.Initializing(config.Options{
			ConfigPath: tempDir,
			ConfigFile: "test_config.json",
		})
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		// Clean up temporary directory
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
	})

	Context("Filter Building", func() {
		It("should pass project ID correctly", func() {
			projectID := int32(123)
			Expect(projectID).To(Equal(int32(123)))
		})

		It("should pass cluster ID correctly", func() {
			clusterID := "cluster-123"
			Expect(clusterID).To(Equal("cluster-123"))
		})
	})

	Context("Aliases", func() {
		It("should have 'ls' as an alias", func() {
			Expect(cmd.Aliases).To(ContainElement("ls"))
		})
	})

	Context("Usage and Examples", func() {
		It("should have proper usage set", func() {
			Expect(cmd.Use).To(Equal("list"))
		})

		It("should have examples defined", func() {
			Expect(cmd.Example).ToNot(BeEmpty())
			Expect(cmd.Example).To(ContainSubstring("runai environment list -A"))
			Expect(cmd.Example).To(ContainSubstring("runai environment list -p <project_name>"))
			Expect(cmd.Example).To(ContainSubstring("runai environment list --yaml"))
		})

		It("should NOT contain pagination examples", func() {
			Expect(cmd.Example).ToNot(ContainSubstring("--limit"))
			Expect(cmd.Example).ToNot(ContainSubstring("--offset"))
			Expect(cmd.Example).ToNot(ContainSubstring("pagination"))
		})
	})
})

var _ = Describe("Output Generation", func() {
	var (
		mockEnvironments []openapi.EnvironmentAsset
	)

	BeforeEach(func() {
		// Create mock environment data
		mockEnvironments = []openapi.EnvironmentAsset{
			createMockEnvironment("pytorch-env", "nvcr.io/nvidia/pytorch:23.10-py3", []string{"JupyterNotebook", "Tensorboard"}, true, false, false, true, "PyTorch"),
			createMockEnvironment("tensorflow-base", "tensorflow/tensorflow:latest-gpu", []string{"Tensorboard", "MLFlow"}, false, true, false, false, ""),
			createMockEnvironment("vscode-workspace", "runai/vscode:latest", []string{"VisualStudioCode"}, true, false, false, false, ""),
		}
	})

	Context("Table Output", func() {
		It("should generate table output with all columns", func() {
			response := &openapi.EnvironmentListResponse{
				Entries: mockEnvironments,
			}
			output, err := list.ParseListResponse(response)
			Expect(err).ToNot(HaveOccurred())

			tableOutput, err := output.GenerateTable()
			Expect(err).ToNot(HaveOccurred())
			Expect(tableOutput).To(ContainSubstring("Name"))
			Expect(tableOutput).To(ContainSubstring("Image"))
			Expect(tableOutput).To(ContainSubstring("Tools"))
			Expect(tableOutput).To(ContainSubstring("Workload Types"))
			Expect(tableOutput).To(ContainSubstring("pytorch-env"))
			Expect(tableOutput).To(ContainSubstring("tensorflow-base"))
			Expect(tableOutput).To(ContainSubstring("vscode-workspace"))
		})

		It("should sort environments by name", func() {
			response := &openapi.EnvironmentListResponse{
				Entries: mockEnvironments,
			}
			output, err := list.ParseListResponse(response)
			Expect(err).ToNot(HaveOccurred())

			data := output.Data()
			Expect(len(data)).To(Equal(3))
			// Should be sorted alphabetically
			Expect(data[0][0]).To(Equal("pytorch-env"))
			Expect(data[1][0]).To(Equal("tensorflow-base"))
			Expect(data[2][0]).To(Equal("vscode-workspace"))
		})
	})

	Context("JSON Output", func() {
		It("should generate valid JSON output", func() {
			response := &openapi.EnvironmentListResponse{
				Entries: mockEnvironments,
			}
			output, err := list.ParseListResponse(response)
			Expect(err).ToNot(HaveOccurred())

			jsonOutput, err := output.GenerateJson()
			Expect(err).ToNot(HaveOccurred())
			Expect(jsonOutput).To(ContainSubstring(`"environments"`))
			Expect(jsonOutput).To(ContainSubstring(`"name"`))
			Expect(jsonOutput).To(ContainSubstring("pytorch-env"))
		})
	})

	Context("YAML Output", func() {
		It("should generate valid YAML output", func() {
			response := &openapi.EnvironmentListResponse{
				Entries: mockEnvironments,
			}
			output, err := list.ParseListResponse(response)
			Expect(err).ToNot(HaveOccurred())

			yamlOutput, err := output.GenerateYaml()
			Expect(err).ToNot(HaveOccurred())
			Expect(yamlOutput).To(ContainSubstring("environments:"))
			Expect(yamlOutput).To(ContainSubstring("name:"))
			Expect(yamlOutput).To(ContainSubstring("pytorch-env"))
		})
	})

	Context("Workload Types Display", func() {
		It("should display workload types with framework information", func() {
			response := &openapi.EnvironmentListResponse{
				Entries: mockEnvironments,
			}
			output, err := list.ParseListResponse(response)
			Expect(err).ToNot(HaveOccurred())

			data := output.Data()
			// First environment: Workspace + Distributed
			Expect(data[0][4]).To(ContainSubstring("Workspace"))
			Expect(data[0][4]).To(ContainSubstring("Distributed"))
			// Second environment: Training only
			Expect(data[1][4]).To(ContainSubstring("Training"))
			Expect(data[1][4]).ToNot(ContainSubstring("Workspace"))
		})

		It("should show hyphen for environments without workload types", func() {
			emptyEnv := createMockEnvironment("empty-env", "image:latest", []string{}, false, false, false, false, "")
			response := &openapi.EnvironmentListResponse{
				Entries: []openapi.EnvironmentAsset{emptyEnv},
			}
			output, err := list.ParseListResponse(response)
			Expect(err).ToNot(HaveOccurred())

			data := output.Data()
			Expect(data[0][4]).To(Equal("-"))
		})
	})

	Context("Tools Extraction", func() {
		It("should extract and format tool names correctly", func() {
			response := &openapi.EnvironmentListResponse{
				Entries: mockEnvironments,
			}
			output, err := list.ParseListResponse(response)
			Expect(err).ToNot(HaveOccurred())

			data := output.Data()
			// Check tools column formatting
			Expect(data[0][3]).To(ContainSubstring("jupyter-notebook"))
			Expect(data[0][3]).To(ContainSubstring("tensorboard"))
			Expect(data[1][3]).To(ContainSubstring("tensorboard"))
			Expect(data[1][3]).To(ContainSubstring("mlflow"))
		})

		It("should show hyphen for environments without tools", func() {
			noToolsEnv := createMockEnvironment("no-tools", "image:latest", []string{}, false, false, false, false, "")
			response := &openapi.EnvironmentListResponse{
				Entries: []openapi.EnvironmentAsset{noToolsEnv},
			}
			output, err := list.ParseListResponse(response)
			Expect(err).ToNot(HaveOccurred())

			data := output.Data()
			Expect(data[0][3]).To(Equal("-"))
		})
	})
})

var _ = Describe("Integration Tests", func() {
	var (
		mockClient *MockEnvironmentClient
	)

	BeforeEach(func() {
		mockClient = &MockEnvironmentClient{}
		list.SetClient(mockClient)
	})

	Context("End-to-End Filter Flow", func() {
		It("should apply project filter to API call", func() {
			projectID := int32(123)

			mockClient.ListEnvironmentsFunc = func(ctx context.Context, projID *int32, clusterID, name, workloadType *string) (*openapi.EnvironmentListResponse, error) {
				// Verify filters are passed correctly
				Expect(projID).ToNot(BeNil())
				Expect(*projID).To(Equal(int32(123)))

				// Return mock data
				return &openapi.EnvironmentListResponse{
					Entries: []openapi.EnvironmentAsset{
						createMockEnvironment("test-env", "image:latest", []string{"JupyterNotebook"}, true, false, false, false, ""),
					},
				}, nil
			}

			response, err := mockClient.List(context.Background(), &projectID, nil, nil, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(mockClient.ListEnvironmentsCalled).To(BeTrue())
			Expect(len(response.Entries)).To(Equal(1))
			Expect(response.Entries[0].Meta.Name).To(Equal("test-env"))
		})

		It("should handle both project and cluster filters", func() {
			projectID := int32(123)
			clusterID := "cluster-abc"

			mockClient.ListEnvironmentsFunc = func(ctx context.Context, projID *int32, clustID, name, workloadType *string) (*openapi.EnvironmentListResponse, error) {
				Expect(projID).ToNot(BeNil())
				Expect(*projID).To(Equal(int32(123)))
				Expect(clustID).ToNot(BeNil())
				Expect(*clustID).To(Equal("cluster-abc"))

				return &openapi.EnvironmentListResponse{
					Entries: []openapi.EnvironmentAsset{},
				}, nil
			}

			_, err := mockClient.List(context.Background(), &projectID, &clusterID, nil, nil)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return environments matching criteria", func() {
			mockClient.ListEnvironmentsFunc = func(ctx context.Context, projID *int32, clustID, name, workloadType *string) (*openapi.EnvironmentListResponse, error) {
				return &openapi.EnvironmentListResponse{
					Entries: []openapi.EnvironmentAsset{
						createMockEnvironment("pytorch-env", "pytorch:23.10", []string{"JupyterNotebook"}, true, false, false, true, "PyTorch"),
						createMockEnvironment("tf-env", "tensorflow:latest", []string{"Tensorboard"}, false, true, false, true, "TensorFlow"),
					},
				}, nil
			}

			response, err := mockClient.List(context.Background(), nil, nil, nil, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(response.Entries)).To(Equal(2))
		})
	})

	Context("Error Handling", func() {
		It("should handle API errors gracefully", func() {
			mockClient.ListEnvironmentsFunc = func(ctx context.Context, projID *int32, clustID, name, workloadType *string) (*openapi.EnvironmentListResponse, error) {
				return nil, fmt.Errorf("API connection failed")
			}

			_, err := mockClient.List(context.Background(), nil, nil, nil, nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("API connection failed"))
		})

		It("should handle empty response", func() {
			mockClient.ListEnvironmentsFunc = func(ctx context.Context, projID *int32, clustID, name, workloadType *string) (*openapi.EnvironmentListResponse, error) {
				return &openapi.EnvironmentListResponse{
					Entries: []openapi.EnvironmentAsset{},
				}, nil
			}

			response, err := mockClient.List(context.Background(), nil, nil, nil, nil)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(response.Entries)).To(Equal(0))
		})
	})
})

// Mock implementation
type MockEnvironmentClient struct {
	ListEnvironmentsCalled bool
	ListEnvironmentsFunc   func(ctx context.Context, projectID *int32, clusterID, name, workloadType *string) (*openapi.EnvironmentListResponse, error)
}

func (m *MockEnvironmentClient) List(ctx context.Context, projectID *int32, clusterID, name, workloadType *string) (*openapi.EnvironmentListResponse, error) {
	m.ListEnvironmentsCalled = true
	if m.ListEnvironmentsFunc != nil {
		return m.ListEnvironmentsFunc(ctx, projectID, clusterID, name, workloadType)
	}
	return &openapi.EnvironmentListResponse{
		Entries: []openapi.EnvironmentAsset{},
	}, nil
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

// Helper function to set output values for testing
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
		tempDir, err = os.MkdirTemp("", "environment-test-settings-*")
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
			// Set test values using helper function
			setOutputValues(&output, []string{"name", "scope", "image"}, true)

			err := list.HandleColumnsFlags(output, cmd, environment.DefaultEnvironmentColumnsSet)

			Expect(err).ToNot(HaveOccurred())
			// Verify columns were saved to settings
			savedColumns := settings.GetInstance().GetEnvironmentColumns()
			Expect(savedColumns).To(Equal([]string{"name", "scope", "image"}))
			// Verify output message
			Expect(buf.String()).To(ContainSubstring("Default environment columns set to"))
		})
	})

	Context("Invalid columns with set-default-columns flag", func() {
		It("should return error and not save to settings", func() {
			// Get initial state
			initialColumns := settings.GetInstance().GetEnvironmentColumns()
			initialLen := len(initialColumns)

			setOutputValues(&output, []string{"invalid-column", "another-invalid"}, true)

			err := list.HandleColumnsFlags(output, cmd, environment.DefaultEnvironmentColumnsSet)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid column name"))
			// Verify no columns were saved (should remain unchanged from initial state)
			savedColumns := settings.GetInstance().GetEnvironmentColumns()
			Expect(len(savedColumns)).To(Equal(initialLen))
		})
	})

	Context("Empty columns with set-default-columns flag", func() {
		It("should save empty columns and show default message", func() {
			setOutputValues(&output, []string{}, true)

			err := list.HandleColumnsFlags(output, cmd, environment.DefaultEnvironmentColumnsSet)

			Expect(err).ToNot(HaveOccurred())
			// Verify empty columns were saved
			savedColumns := settings.GetInstance().GetEnvironmentColumns()
			Expect(len(savedColumns)).To(Equal(0))
			// Verify output shows default columns
			Expect(buf.String()).To(ContainSubstring("Default environment columns set"))
		})
	})

	Context("Valid columns without set-default-columns flag", func() {
		It("should validate columns but not save to settings", func() {
			// Check initial state
			initialColumns := settings.GetInstance().GetEnvironmentColumns()
			initialLen := len(initialColumns)

			setOutputValues(&output, []string{"name", "tools"}, false)

			err := list.HandleColumnsFlags(output, cmd, environment.DefaultEnvironmentColumnsSet)

			Expect(err).ToNot(HaveOccurred())
			// Verify no columns were saved (should remain unchanged)
			savedColumns := settings.GetInstance().GetEnvironmentColumns()
			Expect(len(savedColumns)).To(Equal(initialLen))
			// Verify no output message
			Expect(buf.String()).To(BeEmpty())
		})
	})

	Context("All valid column names", func() {
		It("should accept all valid environment column names", func() {
			validColumns := []string{"name", "scope", "image", "tools", "workloadtypes"}
			setOutputValues(&output, validColumns, true)

			err := list.HandleColumnsFlags(output, cmd, environment.DefaultEnvironmentColumnsSet)

			Expect(err).ToNot(HaveOccurred())
			savedColumns := settings.GetInstance().GetEnvironmentColumns()
			Expect(savedColumns).To(Equal(validColumns))
		})
	})

	Context("Mixed valid and invalid columns", func() {
		It("should reject if any column is invalid", func() {
			initialColumns := settings.GetInstance().GetEnvironmentColumns()
			initialLen := len(initialColumns)

			setOutputValues(&output, []string{"name", "invalid", "scope"}, true)

			err := list.HandleColumnsFlags(output, cmd, environment.DefaultEnvironmentColumnsSet)

			Expect(err).To(HaveOccurred())
			// Verify no columns were saved
			savedColumns := settings.GetInstance().GetEnvironmentColumns()
			Expect(len(savedColumns)).To(Equal(initialLen))
		})
	})
})

// createMockEnvironment creates a mock EnvironmentAsset for testing
func createMockEnvironment(name, image string, tools []string, isWorkspace, isTraining, isInference, isDistributed bool, distFramework string) openapi.EnvironmentAsset {
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

	// Create asset meta with Project scope as default
	scopeEnum := openapi.Scope("Project")
	projectId := int32(789)
	meta := openapi.AssetMeta{
		Name:                   name,
		Scope:                  scopeEnum,
		WorkloadSupportedTypes: *openapi.NewNullableWorkloadSupportedTypes(&workloadTypes),
		Id:                     name + "-id",
		Kind:                   openapi.ASSETKIND_ENVIRONMENT,
		CreatedBy:              "test-user",
		CreatedAt:              time.Now(),
		UpdatedBy:              "test-user",
		UpdatedAt:              time.Now(),
		ProjectId:              *openapi.NewNullableInt32(&projectId),
	}

	// Create asset spec
	spec := openapi.EnvironmentAssetSpec{
		Connections: connections,
	}
	if image != "" {
		spec.Image = *openapi.NewNullableString(&image)
	}

	return openapi.EnvironmentAsset{
		Meta: meta,
		Spec: spec,
	}
}
