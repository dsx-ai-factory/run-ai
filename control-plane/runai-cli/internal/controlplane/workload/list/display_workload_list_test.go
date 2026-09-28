/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"bytes"
	"encoding/json"
	"strings"
	"text/template"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/settings"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/workload/list/manipulate"

	apilist "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/list"

	"os"

	commonlist "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/cmd/common/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/filter"
	cpworkloadmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane/workload"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/pkg/config"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

const (
	workload01JsonStr = `{
      "id": "1",
      "name": "workload-01",
      "type": "Workspace",
      "phase": "Running",
      "projectName": "team-one",
      "preemptible": false,
      "runningPods": 1,
      "requestedPods": {
        "number": 1
      },
      "allocatedResources": {
        "gpu": 1
      }
    }`
	workload01YamlStr = `- id: "1"
      name: workload-01
      type: Workspace
      phase: Running
      project: team-one
      preemptible: false
      runningpods: 1
      requestedpods:
        number: {}
        min: {}
        max: {}
        parallelism: {}
        completions: {}
      allocatedresources:
        gpu: 1
      distributedframework: null`
)

var _ = Describe("Workload Display", Ordered, func() {
	var (
		ctrl     *gomock.Controller
		mockpage *cpworkloadmocks.MockPaginator
		service  *Pager
		fb       *filter.Builder
		out      *Output
	)

	BeforeAll(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockpage = cpworkloadmocks.NewMockPaginator(ctrl)
		service = NewPager(mockpage)
		fb = filter.NewBuilder()
	})

	Context("Output with next", func() {
		var (
			outputStr string
			err       error
		)

		BeforeAll(func() {
			response := &openapi.GetWorkloads200Response{Workloads: []openapi.Workload{
				generateWorkloadFullSpec("1", "workload-01", time.Time{}),
				generateWorkloadFullSpec("2", "workload-02", time.Time{}),
				generateWorkloadFullSpec("3", "workload-03", time.Unix(10, 0)),
			}}
			mockpage.EXPECT().List(gomock.Any(), int32(3), int32(0), fb.StringList(), false).Return(response, nil)
			workloadList, next, err := service.singlePage(transaction.NewTxn(), &apilist.PageOptions{PageSize: 2}, fb, false)
			Expect(err).ToNot(HaveOccurred())
			out, err = NewOutput(workloadList, next, false)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Table", func() {
			outputStr, err = out.GenerateTable()
			Expect(err).ToNot(HaveOccurred())
			Expect(outputStr).To(ContainSubstring("workload-01  Workspace   Runai       Running     team-one    No               1/1                 1.00"))
			Expect(outputStr).To(ContainSubstring("workload-02  Workspace   Runai       Running     team-one    No               1/1                 1.00"))
			Expect(outputStr).To(ContainSubstring("next token: MTk3MC0wMS0wMVQ"))
		})

		It("Json", func() {
			outputStr, err = out.GenerateJson()
			Expect(err).ToNot(HaveOccurred())
			Expect(outputStr).To(ContainSubstring(workload01JsonStr))
			Expect(outputStr).To(ContainSubstring("\"name\": \"workload-02\""))
		})

		It("Yaml", func() {
			outputStr, err = out.GenerateYaml()
			Expect(err).ToNot(HaveOccurred())
			Expect(outputStr).To(ContainSubstring(workload01YamlStr))
			Expect(outputStr).To(ContainSubstring("name: workload-02"))
		})
	})

	Context("Output without next", func() {
		var (
			outputStr string
			err       error
		)

		BeforeAll(func() {
			response := &openapi.GetWorkloads200Response{Workloads: []openapi.Workload{
				generateWorkloadFullSpec("1", "workload-01", time.Time{}),
				generateWorkloadFullSpec("2", "workload-02", time.Unix(10, 0)),
			}}
			mockpage.EXPECT().List(gomock.Any(), int32(3), int32(0), fb.StringList(), false).Return(response, nil)
			workloadList, next, err := service.singlePage(transaction.NewTxn(), &apilist.PageOptions{PageSize: 2}, fb, false)
			Expect(err).ToNot(HaveOccurred())
			out, err = NewOutput(workloadList, next, false)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Table", func() {
			outputStr, err = out.GenerateTable()
			Expect(err).ToNot(HaveOccurred())
			Expect(outputStr).To(ContainSubstring("workload-01"))
			Expect(outputStr).To(ContainSubstring("workload-02"))
			Expect(outputStr).ToNot(ContainSubstring("next"))
		})

		It("Json", func() {
			outputStr, err = out.GenerateJson()
			Expect(err).ToNot(HaveOccurred())
			Expect(outputStr).To(ContainSubstring("\"name\": \"workload-01\""))
			Expect(outputStr).To(ContainSubstring("\"name\": \"workload-02\""))
			Expect(outputStr).ToNot(ContainSubstring("\"next\""))
		})

		It("Yaml", func() {
			outputStr, err = out.GenerateYaml()
			Expect(err).ToNot(HaveOccurred())
			Expect(outputStr).To(ContainSubstring("name: workload-01"))
			Expect(outputStr).To(ContainSubstring("name: workload-02"))
			Expect(outputStr).ToNot(ContainSubstring("next:"))
		})
	})

	Context("Output with deleted flag", func() {
		var (
			outputStr  string
			err        error
			deletedOut *Output
		)

		BeforeAll(func() {
			response := &openapi.GetWorkloads200Response{Workloads: []openapi.Workload{
				generateWorkloadFullSpec("1", "deleted-workload-01", time.Time{}),
				generateWorkloadFullSpec("2", "deleted-workload-02", time.Time{}),
			}}
			// Set phases to Deleting for realism
			deletingPhase, _ := openapi.NewPhaseFromValue("Deleting")
			response.Workloads[0].Phase = *deletingPhase
			response.Workloads[1].Phase = *deletingPhase

			mockpage.EXPECT().List(gomock.Any(), int32(3), int32(0), fb.StringList(), true).Return(response, nil)
			workloadList, next, err := service.singlePage(transaction.NewTxn(), &apilist.PageOptions{PageSize: 2}, fb, true)
			Expect(err).ToNot(HaveOccurred())
			deletedOut, err = NewOutput(workloadList, next, true)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Table should not have Status column", func() {
			columns := deletedOut.Columns()
			// Verify Status column is not present
			columnTitles := []string{}
			for _, col := range columns {
				columnTitles = append(columnTitles, col.Title)
			}
			Expect(columnTitles).To(Equal([]string{manipulate.NameField.GetColumnTitle(), manipulate.TypeField.GetColumnTitle(), manipulate.FrameworkField.GetColumnTitle(), manipulate.ProjectField.GetColumnTitle(), manipulate.PreemptibleField.GetColumnTitle(), manipulate.PodsField.GetColumnTitle(), manipulate.GPUAllocationField.GetColumnTitle()}))
			Expect(columnTitles).ToNot(ContainElement(manipulate.StatusField.GetColumnTitle()))
		})

		It("Table data should not contain status values", func() {
			outputStr, err = deletedOut.GenerateTable()
			Expect(err).ToNot(HaveOccurred())
			// Check that workload names are present
			Expect(outputStr).To(ContainSubstring("deleted-workload-01"))
			Expect(outputStr).To(ContainSubstring("deleted-workload-02"))
			// Check that the Status column header is not present
			Expect(outputStr).ToNot(ContainSubstring(manipulate.StatusField.GetColumnTitle()))
			// Check that status values like "Deleting" are not present in the output
			Expect(outputStr).ToNot(ContainSubstring("Deleting"))
			// The word "Running" appears in "Running/Req. Pods" column header, so we need to be more specific
			// Check that common status values don't appear in the data rows (after the header separator)
			lines := strings.Split(outputStr, "\n")
			dataStarted := false
			for _, line := range lines {
				if strings.Contains(line, "───") {
					dataStarted = true
					continue
				}
				if dataStarted && line != "" {
					// Check that status values are not in data rows
					Expect(line).ToNot(ContainSubstring("Deleting"))
					Expect(line).ToNot(ContainSubstring("Pending"))
					Expect(line).ToNot(ContainSubstring("Initializing"))
					Expect(line).ToNot(ContainSubstring("Failed"))
					Expect(line).ToNot(ContainSubstring("Completed"))
					Expect(line).ToNot(ContainSubstring("Stopped"))
				}
			}
			// Verify the row format matches expected pattern without status
			Expect(outputStr).To(MatchRegexp(`deleted-workload-01\s+Workspace\s+Runai\s+team-one\s+No\s+1/1\s+1\.00`))
		})

		It("JSON should not contain phase field", func() {
			outputStr, err = deletedOut.GenerateJson()
			Expect(err).ToNot(HaveOccurred())
			// Phase field should be completely omitted due to omitempty tag
			Expect(outputStr).ToNot(ContainSubstring("\"phase\""))
			// But other fields should be present
			Expect(outputStr).To(ContainSubstring("\"name\": \"deleted-workload-01\""))
			Expect(outputStr).To(ContainSubstring("\"name\": \"deleted-workload-02\""))
			Expect(outputStr).To(ContainSubstring("\"type\": \"Workspace\""))
			Expect(outputStr).To(ContainSubstring("\"projectName\": \"team-one\""))
		})

		It("YAML should contain phase: null", func() {
			outputStr, err = deletedOut.GenerateYaml()
			Expect(err).ToNot(HaveOccurred())
			// Phase field should be present but null in YAML
			Expect(outputStr).To(ContainSubstring("phase: null"))
			// Other fields should be present
			Expect(outputStr).To(ContainSubstring("name: deleted-workload-01"))
			Expect(outputStr).To(ContainSubstring("name: deleted-workload-02"))
			Expect(outputStr).To(ContainSubstring("type: Workspace"))
			Expect(outputStr).To(ContainSubstring("project: team-one"))
		})
	})

	Context("Raw flag and format alignment features", func() {
		var (
			workloadList []openapi.Workload
		)

		BeforeEach(func() {
			// Create test workload data
			workloadList = []openapi.Workload{
				generateWorkloadFullSpec("test-1", "raw-test-workload-1", time.Unix(10, 0)),
				generateWorkloadFullSpec("test-2", "raw-test-workload-2", time.Unix(20, 0)),
			}
		})

		Context("Raw flag behavior", func() {
			It("should return original API response when raw=true for JSON", func() {
				out, err := NewOutput(workloadList, nil, false)
				Expect(err).ToNot(HaveOccurred())
				out.SetRaw(true)

				jsonOutput, err := out.GenerateJson()
				Expect(err).ToNot(HaveOccurred())

				// Should contain original API fields
				Expect(jsonOutput).To(ContainSubstring("\"actionsSupport\""))
				Expect(jsonOutput).To(ContainSubstring("\"tenantId\""))
				Expect(jsonOutput).To(ContainSubstring("\"clusterId\""))
				Expect(jsonOutput).To(ContainSubstring("\"departmentId\""))
				Expect(jsonOutput).To(ContainSubstring("\"id\""))          // id is part of original API response
				Expect(jsonOutput).To(ContainSubstring("\"projectName\"")) // projectName is part of original API response

				// Should not contain column title format (aligned output)
				Expect(jsonOutput).ToNot(ContainSubstring("\"Workload\""))
				Expect(jsonOutput).ToNot(ContainSubstring("\"Type\""))
				Expect(jsonOutput).ToNot(ContainSubstring("\"Framework\""))
			})

			It("should return original API response when raw=true for YAML", func() {
				out, err := NewOutput(workloadList, nil, false)
				Expect(err).ToNot(HaveOccurred())
				out.SetRaw(true)

				yamlOutput, err := out.GenerateYaml()
				Expect(err).ToNot(HaveOccurred())

				// Should contain original API fields (YAML uses lowercase field names)
				Expect(yamlOutput).To(ContainSubstring("actionssupport:"))
				Expect(yamlOutput).To(ContainSubstring("tenantid:"))
				Expect(yamlOutput).To(ContainSubstring("clusterid:"))
				Expect(yamlOutput).To(ContainSubstring("departmentid:"))
				Expect(yamlOutput).To(ContainSubstring("id:"))          // id is part of original API response
				Expect(yamlOutput).To(ContainSubstring("projectname:")) // projectName is part of original API response
			})

		})

		Context("Brief format output (default behavior)", func() {
			It("should use brief format when raw=false for JSON", func() {
				out, err := NewOutput(workloadList, nil, false)
				Expect(err).ToNot(HaveOccurred())

				jsonOutput, err := out.GenerateJson()
				Expect(err).ToNot(HaveOccurred())

				// The output format depends on the current config setting
				// But should always contain workload data in some form
				Expect(jsonOutput).To(ContainSubstring("raw-test-workload-1"))
				Expect(jsonOutput).To(ContainSubstring("raw-test-workload-2"))
				Expect(jsonOutput).To(ContainSubstring("\"workloads\""))

				// Should not contain original API fields when not raw
				Expect(jsonOutput).ToNot(ContainSubstring("\"actionsSupport\""))
				Expect(jsonOutput).ToNot(ContainSubstring("\"tenantId\""))
			})

			It("should use brief format when raw=false for YAML", func() {
				out, err := NewOutput(workloadList, nil, false)
				Expect(err).ToNot(HaveOccurred())

				yamlOutput, err := out.GenerateYaml()
				Expect(err).ToNot(HaveOccurred())

				// Should contain workload data
				Expect(yamlOutput).To(ContainSubstring("raw-test-workload-1"))
				Expect(yamlOutput).To(ContainSubstring("raw-test-workload-2"))
				Expect(yamlOutput).To(ContainSubstring("workloads:"))

				// Should not contain original API fields when not raw (YAML uses lowercase)
				Expect(yamlOutput).ToNot(ContainSubstring("actionssupport:"))
				Expect(yamlOutput).ToNot(ContainSubstring("tenantid:"))
			})
		})

		Context("Raw flag always takes precedence", func() {
			It("should use raw format regardless of config when raw=true", func() {
				out, err := NewOutput(workloadList, nil, false) // raw=true
				Expect(err).ToNot(HaveOccurred())
				out.SetRaw(true)

				jsonOutput, err := out.GenerateJson()
				Expect(err).ToNot(HaveOccurred())

				// Should always use original API format when raw=true
				Expect(jsonOutput).To(ContainSubstring("\"actionsSupport\""))
				Expect(jsonOutput).To(ContainSubstring("\"tenantId\""))
				Expect(jsonOutput).To(ContainSubstring("\"id\"")) // id is part of original API response

				// Should not contain aligned format column titles
				Expect(jsonOutput).ToNot(ContainSubstring("\"Workload\""))
			})
		})

		Context("Deleted workloads behavior", func() {
			It("should handle deleted workloads correctly", func() {
				out, err := NewOutput(workloadList, nil, true) // deleted=true
				Expect(err).ToNot(HaveOccurred())

				jsonOutput, err := out.GenerateJson()
				Expect(err).ToNot(HaveOccurred())

				// Should contain workload data
				Expect(jsonOutput).To(ContainSubstring("raw-test-workload-1"))
				Expect(jsonOutput).To(ContainSubstring("raw-test-workload-2"))
				Expect(jsonOutput).To(ContainSubstring("\"workloads\""))
			})
		})

		Context("Edge cases", func() {
			It("should handle empty workload list", func() {
				out, err := NewOutput([]openapi.Workload{}, nil, false)
				Expect(err).ToNot(HaveOccurred())

				jsonOutput, err := out.GenerateJson()
				Expect(err).ToNot(HaveOccurred())
				Expect(jsonOutput).To(ContainSubstring("\"workloads\": []"))

				yamlOutput, err := out.GenerateYaml()
				Expect(err).ToNot(HaveOccurred())
				Expect(yamlOutput).To(ContainSubstring("workloads: []"))
			})

			It("should handle workloads with nil values gracefully", func() {
				// Create workload with minimal data
				minimalWorkload := openapi.Workload{
					Id:          "minimal-1",
					Name:        "minimal-workload",
					Type:        "Workspace",
					Phase:       openapi.PHASE_RUNNING,
					ProjectName: "test-project",
					RunningPods: 0,
					TenantId:    1001,
				}
				// Preemptible and AllocatedResources are nil

				out, err := NewOutput([]openapi.Workload{minimalWorkload}, nil, false)
				Expect(err).ToNot(HaveOccurred())

				jsonOutput, err := out.GenerateJson()
				Expect(err).ToNot(HaveOccurred())

				// Should contain the workload data
				Expect(jsonOutput).To(ContainSubstring("minimal-workload"))
				Expect(jsonOutput).To(ContainSubstring("\"workloads\""))
			})

			It("should handle raw output with empty workload list", func() {
				out, err := NewOutput([]openapi.Workload{}, nil, false)
				Expect(err).ToNot(HaveOccurred())
				out.SetRaw(true)

				jsonOutput, err := out.GenerateJson()
				Expect(err).ToNot(HaveOccurred())
				Expect(jsonOutput).To(ContainSubstring("\"workloads\": []"))

				yamlOutput, err := out.GenerateYaml()
				Expect(err).ToNot(HaveOccurred())
				Expect(yamlOutput).To(ContainSubstring("workloads: []"))
			})

		})
	})

	Context("Brief in output tests", func() {
		var (
			workloadList []openapi.Workload
		)

		BeforeEach(func() {
			workloadList = []openapi.Workload{
				generateWorkloadFullSpec("test-1", "memory-test-1", time.Unix(10, 0)),
				generateWorkloadFullSpec("test-2", "memory-test-2", time.Unix(20, 0)),
			}
		})

		It("should store only original workloads in Data struct", func() {
			out, err := NewOutput(workloadList, nil, false)
			Expect(err).ToNot(HaveOccurred())

			// Should store original workloads
			Expect(out.data.Workloads).To(HaveLen(2))
			Expect(out.data.Workloads[0].Name).To(Equal("memory-test-1"))
			Expect(out.data.Workloads[1].Name).To(Equal("memory-test-2"))

			// Should store full original API response fields
			Expect(out.data.Workloads[0].Id).To(Equal("test-1"))
			Expect(out.data.Workloads[0].TenantId).To(Equal(int32(1001)))
		})

		It("getBriefWorkloads should convert on-demand without storing", func() {
			out, err := NewOutput(workloadList, nil, false)
			Expect(err).ToNot(HaveOccurred())

			// First call - should convert
			brief1 := out.getBriefWorkloads()
			Expect(brief1).To(HaveLen(2))
			Expect(brief1[0].Name).To(Equal("memory-test-1"))
			Expect(brief1[0].Id).To(Equal("test-1"))

			// Second call - should convert again (no caching)
			brief2 := out.getBriefWorkloads()
			Expect(brief2).To(HaveLen(2))
			Expect(brief2[0].Name).To(Equal("memory-test-1"))
		})

		It("should handle empty workload list", func() {
			out, err := NewOutput([]openapi.Workload{}, nil, false)
			Expect(err).ToNot(HaveOccurred())

			brief := out.getBriefWorkloads()
			Expect(brief).To(HaveLen(0))
			Expect(brief).ToNot(BeNil())
		})

		It("should handle nil workload list", func() {
			out, err := NewOutput(nil, nil, false)
			Expect(err).ToNot(HaveOccurred())

			brief := out.getBriefWorkloads()
			Expect(brief).To(HaveLen(0))
		})

		It("should not store duplicate data - memory optimization verification", func() {
			workloadList := []openapi.Workload{
				generateWorkloadFullSpec("memory-1", "memory-check", time.Unix(10, 0)),
			}

			out, err := NewOutput(workloadList, nil, false)
			Expect(err).ToNot(HaveOccurred())

			// Should store original workloads
			Expect(out.data.Workloads).To(HaveLen(1))
			Expect(out.data.Workloads[0].Name).To(Equal("memory-check"))

			// Should not pre-convert to brief format (no duplicate storage)
			// We verify this by checking that getBriefWorkloads does fresh conversion
			brief1 := out.getBriefWorkloads()
			brief2 := out.getBriefWorkloads()

			// Both should be equal but separate conversions
			Expect(brief1).To(Equal(brief2))
			Expect(brief1[0].Name).To(Equal("memory-check"))
		})
	})

	Context("SetRaw functionality tests", func() {
		var (
			workloadList []openapi.Workload
			out          *Output
		)

		BeforeEach(func() {
			workloadList = []openapi.Workload{
				generateWorkloadFullSpec("test-1", "raw-test-1", time.Unix(10, 0)),
			}
			var err error
			out, err = NewOutput(workloadList, nil, false)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should default to raw=false", func() {
			Expect(out.raw).To(BeFalse())
		})

		It("should set raw flag correctly", func() {
			out.SetRaw(true)
			Expect(out.raw).To(BeTrue())

			out.SetRaw(false)
			Expect(out.raw).To(BeFalse())
		})

		It("should use raw data when raw=true in JSON", func() {
			out.SetRaw(true)
			jsonOutput, err := out.GenerateJson()
			Expect(err).ToNot(HaveOccurred())

			// Should contain original API fields
			Expect(jsonOutput).To(ContainSubstring("\"tenantId\""))
			Expect(jsonOutput).To(ContainSubstring("\"clusterId\""))
			Expect(jsonOutput).To(ContainSubstring("\"actionsSupport\""))
		})

		It("should use brief data when raw=false in JSON", func() {
			out.SetRaw(false)
			jsonOutput, err := out.GenerateJson()
			Expect(err).ToNot(HaveOccurred())

			// Should contain brief fields, not original API fields
			Expect(jsonOutput).To(ContainSubstring("\"id\""))
			Expect(jsonOutput).To(ContainSubstring("\"name\""))
			Expect(jsonOutput).ToNot(ContainSubstring("\"tenantId\""))
			Expect(jsonOutput).ToNot(ContainSubstring("\"actionsSupport\""))
		})
	})

	Context("Deleted workloads handling", func() {
		var (
			workloadList []openapi.Workload
		)

		BeforeEach(func() {
			workloadList = []openapi.Workload{
				generateWorkloadFullSpec("deleted-1", "deleted-workload", time.Unix(10, 0)),
			}
		})

		It("should pass deleted flag to BriefListFromResponse", func() {
			out, err := NewOutput(workloadList, nil, true) // deleted = true
			Expect(err).ToNot(HaveOccurred())

			brief := out.getBriefWorkloads()
			Expect(brief).To(HaveLen(1))
			// For deleted workloads, phase should be nil in brief format
			Expect(brief[0].Phase).To(BeNil())
		})

		It("should include phase for non-deleted workloads", func() {
			out, err := NewOutput(workloadList, nil, false) // deleted = false
			Expect(err).ToNot(HaveOccurred())

			brief := out.getBriefWorkloads()
			Expect(brief).To(HaveLen(1))
			// For non-deleted workloads, phase should be included
			Expect(brief[0].Phase).ToNot(BeNil())
		})
	})

	Context("Generation method consistency", func() {
		var (
			workloadList []openapi.Workload
			out          *Output
		)

		BeforeEach(func() {
			workloadList = []openapi.Workload{
				generateWorkloadFullSpec("consistency-1", "consistency-test", time.Unix(10, 0)),
			}
			var err error
			out, err = NewOutput(workloadList, nil, false)
			Expect(err).ToNot(HaveOccurred())
		})

		It("should have consistent behavior between JSON and YAML for raw output", func() {
			out.SetRaw(true)

			jsonOutput, err := out.GenerateJson()
			Expect(err).ToNot(HaveOccurred())

			yamlOutput, err := out.GenerateYaml()
			Expect(err).ToNot(HaveOccurred())

			// Both should contain original API fields
			Expect(jsonOutput).To(ContainSubstring("tenantId"))
			Expect(yamlOutput).To(ContainSubstring("tenantid"))
		})

		It("should have consistent behavior between JSON and YAML for brief output", func() {
			out.SetRaw(false)

			jsonOutput, err := out.GenerateJson()
			Expect(err).ToNot(HaveOccurred())

			yamlOutput, err := out.GenerateYaml()
			Expect(err).ToNot(HaveOccurred())

			// Both should contain brief fields
			Expect(jsonOutput).To(ContainSubstring("\"name\""))
			Expect(yamlOutput).To(ContainSubstring("name:"))
		})
	})

	Context("Interface compatibility", func() {
		It("should implement GenerateOutput interface correctly", func() {
			out, err := NewOutput([]openapi.Workload{}, nil, false)
			Expect(err).ToNot(HaveOccurred())

			// Should implement all required methods
			var _ commonlist.GenerateOutput = out

			// SetRaw should be callable without errors
			out.SetRaw(true)
			out.SetRaw(false)
		})
	})

	Context("GetList method tests", func() {
		var (
			workloadList []openapi.Workload
		)

		BeforeEach(func() {
			workloadList = []openapi.Workload{
				generateWorkloadFullSpec("getlist-1", "getlist-test-1", time.Unix(10, 0)),
				generateWorkloadFullSpec("getlist-2", "getlist-test-2", time.Unix(20, 0)),
			}
		})

		It("should return brief workloads via GetList", func() {
			out, err := NewOutput(workloadList, nil, false)
			Expect(err).ToNot(HaveOccurred())

			briefList := out.GetList()
			Expect(briefList).To(HaveLen(2))
			Expect(briefList[0].Name).To(Equal("getlist-test-1"))
			Expect(briefList[1].Name).To(Equal("getlist-test-2"))
		})

		It("should handle empty list in GetList", func() {
			out, err := NewOutput([]openapi.Workload{}, nil, false)
			Expect(err).ToNot(HaveOccurred())

			briefList := out.GetList()
			Expect(briefList).To(HaveLen(0))
			Expect(briefList).ToNot(BeNil())
		})
	})

	Context("Custom columns functionality", func() {
		var (
			outputStr  string
			err        error
			columnsOut *Output
			tmpDir     string
		)

		BeforeAll(func() {
			// Create temporary directory for config
			tmpDir, err = os.MkdirTemp("", "workload_test_config")
			Expect(err).ToNot(HaveOccurred())

			// Initialize config with test options
			err = config.GetInstance().Initializing(config.Options{
				Debug:      true,
				ConfigPath: tmpDir,
				ConfigFile: "test_config.json",
			})
			err = settings.GetInstance().Initializing(tmpDir)
			Expect(err).ToNot(HaveOccurred())

			response := &openapi.GetWorkloads200Response{Workloads: []openapi.Workload{
				generateWorkloadFullSpec("1", "workload-01", time.Time{}),
				generateWorkloadFullSpec("2", "workload-02", time.Time{}),
			}}
			mockpage.EXPECT().List(gomock.Any(), int32(3), int32(0), fb.StringList(), false).Return(response, nil)
			workloadList, next, err := service.singlePage(transaction.NewTxn(), &apilist.PageOptions{PageSize: 2}, fb, false)
			Expect(err).ToNot(HaveOccurred())

			// Create output and set specific columns for testing
			columnsOut, err = NewOutput(workloadList, next, false)
			Expect(err).ToNot(HaveOccurred())
			columnsOut.SetColumns([]string{"name", "type", "status"}) // Only these columns
		})

		AfterAll(func() {
			if tmpDir != "" {
				_ = os.RemoveAll(tmpDir)
			}
		})

		Context("when specific columns are set", func() {
			It("GenerateTable should only show selected columns", func() {
				outputStr, err = columnsOut.GenerateTable()
				Expect(err).ToNot(HaveOccurred())

				// Should contain headers for selected columns
				Expect(outputStr).To(ContainSubstring("Workload")) // name column header
				Expect(outputStr).To(ContainSubstring("Type"))     // type column header
				Expect(outputStr).To(ContainSubstring("Status"))   // status column header

				// Should contain the actual data
				Expect(outputStr).To(ContainSubstring("workload-01"))
				Expect(outputStr).To(ContainSubstring("workload-02"))
				Expect(outputStr).To(ContainSubstring("Workspace"))
				Expect(outputStr).To(ContainSubstring("Running"))
			})

			It("GenerateJson should populate only selected columns in aligned format", func() {
				// Save original setting
				originalFormatAlign := settings.GetInstance().GetListFormatAlign()

				// Set format alignment to true to enable aligned output
				err := settings.GetInstance().SetListFormatAlign(true)
				Expect(err).ToNot(HaveOccurred())

				// Ensure we restore the original setting after the test
				defer func() {
					err := settings.GetInstance().SetListFormatAlign(originalFormatAlign)
					Expect(err).ToNot(HaveOccurred())
				}()

				outputStr, err = columnsOut.GenerateJson()
				Expect(err).ToNot(HaveOccurred())

				// Parse JSON to verify structure
				var result map[string]interface{}
				err = json.Unmarshal([]byte(outputStr), &result)
				Expect(err).ToNot(HaveOccurred())

				// Should have workloads array
				workloads, ok := result["workloads"].([]interface{})
				Expect(ok).To(BeTrue())
				Expect(workloads).To(HaveLen(2))

				// Check first workload
				firstWorkload, ok := workloads[0].(map[string]interface{})
				Expect(ok).To(BeTrue())

				// Selected columns should have actual values
				Expect(firstWorkload["name"]).To(Equal("workload-01")) // name was selected and populated
				Expect(firstWorkload["type"]).To(Equal("Workspace"))   // type was selected and populated
				Expect(firstWorkload["status"]).To(Equal("Running"))   // status was selected and populated

				// Unselected columns should be omitted (all fields now have omitempty)
				_, hasProjectName := firstWorkload["projectName"]
				Expect(hasProjectName).To(BeFalse()) // project not selected, omitted

				_, hasFramework := firstWorkload["framework"]
				Expect(hasFramework).To(BeFalse()) // framework not selected, omitted

				_, hasPreemptible := firstWorkload["preemptible"]
				Expect(hasPreemptible).To(BeFalse()) // preemptible not selected, omitted
			})

			It("GenerateYaml should populate only selected columns in aligned format", func() {
				// Save original setting
				originalFormatAlign := settings.GetInstance().GetListFormatAlign()

				// Set format alignment to true to enable aligned output
				err := settings.GetInstance().SetListFormatAlign(true)
				Expect(err).ToNot(HaveOccurred())

				// Ensure we restore the original setting after the test
				defer func() {
					err := settings.GetInstance().SetListFormatAlign(originalFormatAlign)
					Expect(err).ToNot(HaveOccurred())
				}()

				outputStr, err = columnsOut.GenerateYaml()
				Expect(err).ToNot(HaveOccurred())

				// Parse JSON to verify structure
				var result map[string]interface{}
				err = yaml.Unmarshal([]byte(outputStr), &result)
				Expect(err).ToNot(HaveOccurred())

				// Should have workloads array
				workloads, ok := result["workloads"].([]interface{})
				Expect(ok).To(BeTrue())
				Expect(workloads).To(HaveLen(2))

				// Check first workload
				firstWorkload, ok := workloads[0].(map[string]interface{})
				Expect(ok).To(BeTrue())

				// Selected columns should have actual values
				Expect(firstWorkload["name"]).To(Equal("workload-01")) // name was selected and populated
				Expect(firstWorkload["type"]).To(Equal("Workspace"))   // type was selected and populated
				Expect(firstWorkload["status"]).To(Equal("Running"))   // status was selected and populated

				// Unselected columns should be omitted (all fields now have omitempty)
				_, hasProjectName := firstWorkload["projectName"]
				Expect(hasProjectName).To(BeFalse()) // project not selected, omitted

				_, hasFramework := firstWorkload["framework"]
				Expect(hasFramework).To(BeFalse()) // framework not selected, omitted

				_, hasPreemptible := firstWorkload["preemptible"]
				Expect(hasPreemptible).To(BeFalse()) // preemptible not selected, omitted
			})
		})

		Context("when only name column is set", func() {
			BeforeEach(func() {
				columnsOut.SetColumns([]string{"name"})
			})

			It("GenerateTable should only show name column", func() {
				outputStr, err = columnsOut.GenerateTable()
				Expect(err).ToNot(HaveOccurred())

				// Should contain name data and header
				Expect(outputStr).To(ContainSubstring("Workload"))    // name header
				Expect(outputStr).To(ContainSubstring("workload-01")) // name data
				Expect(outputStr).To(ContainSubstring("workload-02")) // name data

				// Should NOT contain other column headers
				Expect(outputStr).ToNot(ContainSubstring("Type"))
				Expect(outputStr).ToNot(ContainSubstring("Status"))
				Expect(outputStr).ToNot(ContainSubstring("Project"))
			})
		})

		Context("when resource-related columns are set", func() {
			BeforeEach(func() {
				columnsOut.SetColumns([]string{"name", "allocatedgpu", "urls"})
			})

			It("GenerateTable should display resource columns", func() {
				outputStr, err = columnsOut.GenerateTable()
				Expect(err).ToNot(HaveOccurred())

				// Should contain selected columns
				Expect(outputStr).To(ContainSubstring("Workload"))   // name header
				Expect(outputStr).To(ContainSubstring("GPU Alloc.")) // gpu header
				Expect(outputStr).To(ContainSubstring("URLs"))       // urls header

				// Should contain resource data
				Expect(outputStr).To(ContainSubstring("1.00")) // GPU allocation value
			})
		})

		Context("when no columns are set", func() {
			BeforeEach(func() {
				columnsOut.SetColumns([]string{})
			})

			It("should fall back to default columns", func() {
				outputStr, err = columnsOut.GenerateTable()
				Expect(err).ToNot(HaveOccurred())

				// Should contain default column headers
				Expect(outputStr).To(ContainSubstring("Workload"))
				Expect(outputStr).To(ContainSubstring("Type"))
				Expect(outputStr).To(ContainSubstring("Framework"))
				Expect(outputStr).To(ContainSubstring("Project"))
			})
		})
	})
})

func generateWorkloadFullSpec(id, name string, createdAt time.Time) openapi.Workload {
	// removed created by for page mock changes
	workloadSample := `{
  "actionsSupport": {
    "delete": true,
    "suspend": true
  },
  "allocatedResources": {
    "cpu": 0.1,
    "cpuMemory": "100M",
    "gpu": 1,
    "gpuMemory": "12582M"
  },
  "arguments": "--NotebookApp.base_url=/${RUNAI_PROJECT}/${RUNAI_JOB_NAME} --NotebookApp.token=''",
  "clusterId": "b754457d-7374-4da4-a43f-ac13fb899838",
  "command": "start-notebook.sh",
  "conditions": [
  ],
  "currentNodePools": [
    "default"
  ],
  "deletedAt": null,
  "departmentId": "4500000",
  "departmentName": "default",
  "environmentVariables": {
    "JUPYTER_PORT": "8888",
    "RUNAI_JOB_NAME": "${RUNAI_JOB_NAME}",
    "RUNAI_PROJECT": "${RUNAI_PROJECT}"
  },
  "environments": [
    {
      "connections": [
        {
          "connectionType": "ExternalUrl",
          "containerPort": 8888,
          "name": "Jupyter",
          "toolType": "jupyter-notebook",
          "url": "https://fox-demo.example.com/team-one/jup-01"
        }
      ],
      "id": "ab6b7bac-922c-4dea-964e-288e28ae1804",
      "name": "jupyter-lab"
    }
  ],
  "fullySaasWorkload": true,
  "id": "{{ .id }}",
  "idleAllocatedGpus": 0,
  "idleGpus": 0,
  "images": [
    "jupyter/scipy-notebook"
  ],
  "k8sPhase": "",
  "k8sPhaseUpdatedAt": "2025-04-09T12:52:23.735709Z",
  "name": "{{ .name }}",
  "namespace": "runai-team-one",
  "phase": "Running",
  "phaseUpdatedAt": "2025-04-09T12:53:22.000758Z",
  "podsRequestedResources": {
    "cpu": {
      "request": 0.1
    },
    "cpuMemory": {
      "request": "100M"
    },
    "gpu": {
      "limit": 1,
      "request": 1
    },
    "gpuRequestType": "portion"
  },
  "preemptible": false,
  "priority": 100,
  "priorityClassName": "build",
  "projectId": "4500003",
  "projectName": "team-one",
  "requestedNodePools": [
    "default"
  ],
  "requestedPods": {
    "number": 1
  },
  "runningPods": 1,
  "source": "Control-plane",
  "submittedBy": "test@example.com",
  "tenantId": 1001,
  "totalPendingTimeSeconds": 5,
  "type": "Workspace",
  "updatedAt": "2025-04-10T06:28:43.632927Z",
  "urls": [
    "https://fox-demo.example.com/team-one/jup-01"
  ],
  "workloadRequestedResources": {
    "cpu": {
      "request": 0.1
    },
    "cpuMemory": {
      "request": "100M"
    },
    "gpu": {
      "request": 1
    },
    "gpuRequestType": "portion"
  }
}`

	stringBuf := bytes.Buffer{}
	temp := template.Must(template.New("workload").Parse(workloadSample))
	err := temp.Execute(&stringBuf, map[string]interface{}{
		"name": name,
		"id":   id,
	})
	if err != nil {
		panic(err)
	}

	workload := openapi.Workload{}
	err = json.Unmarshal(stringBuf.Bytes(), &workload)
	if err != nil {
		panic(err)
	}
	if !createdAt.IsZero() {
		workload.CreatedAt = createdAt
	}
	return workload
}
