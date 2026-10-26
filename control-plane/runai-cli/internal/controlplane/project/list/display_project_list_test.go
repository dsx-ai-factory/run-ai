/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list_test

import (
	"bytes"
	"encoding/json"
	"text/template"
	"time"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/filter"
	cpmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	apilist "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/project/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// This struct will recognize any change in output.
const (
	project01JsonStr = `{
      "id": "1",
      "name": "project-01",
      "parent": {
        "name": "default"
      },
      "resources": [
        {
          "gpu": {
            "deserved": 8
          }
        }
      ],
      "status": {
        "phase": "Ready",
        "nodePoolQuotaStatuses": [
          {
            "allocated": {
              "gpu": 2
            }
          }
        ]
      }
    }`
	project01YamlStr = `- id: "1"
      name: project-01
      briefdepartment:
        departmentname: default
      resources:
        - gpu:
            deserved: 8
      briefprojectstatus:
        phase: Ready
        quotastatuses:
            - allocated:
                gpu: 2`
)

var _ = Describe("Project Display", Ordered, func() {
	var (
		ctrl     *gomock.Controller
		mockpage *cpmocks.MockProjectLister
		service  *list.ProjectService
		fb       *filter.Builder
		out      *list.Output
	)

	BeforeAll(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockpage = cpmocks.NewMockProjectLister(ctrl)
		service = list.NewProjectService(mockpage)
		fb = filter.NewBuilder()
	})

	Context("Output with next", func() {
		var (
			outputStr   string
			err         error
			pageOptions = apilist.PageOptions{PageSize: 2, NoPagination: true}
		)

		BeforeAll(func() {
			response := &openapi.GetProjects200Response{Projects: []openapi.Project{
				generateProjectFullSpec("1", "project-01", time.Time{}),
				generateProjectFullSpec("2", "project-02", time.Time{}),
				generateProjectFullSpec("3", "project-03", time.Unix(10, 0)),
			}}
			mockpage.EXPECT().List(gomock.Any(), int32(3), int32(0), fb.StringList()).Return(response, nil)
			out, err = service.List(transaction.NewTxn(), &pageOptions, fb)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Table", func() {
			outputStr, err = out.GenerateTable()
			Expect(err).ToNot(HaveOccurred())
			Expect(outputStr).To(ContainSubstring("project-01  Ready       default          8.00        2.00             25.00%"))
			Expect(outputStr).To(ContainSubstring("project-02  Ready       default          8.00        2.00             25.00%"))
			Expect(outputStr).To(ContainSubstring("next token: MTk3MC0wMS0wMVQ"))
		})

		It("Json", func() {
			outputStr, err = out.GenerateJson()
			Expect(err).ToNot(HaveOccurred())
			Expect(outputStr).To(ContainSubstring(project01JsonStr))
			Expect(outputStr).To(ContainSubstring("\"name\": \"project-02\""))
			Expect(outputStr).To(ContainSubstring("\"nextToken\": \"MTk3MC0wMS0wMVQ"))
		})

		It("Yaml", func() {
			outputStr, err = out.GenerateYaml()
			Expect(err).ToNot(HaveOccurred())
			Expect(outputStr).To(ContainSubstring(project01YamlStr))
			Expect(outputStr).To(ContainSubstring("name: project-02"))
			Expect(outputStr).To(ContainSubstring("nextToken: MTk3MC0wMS0wMVQ"))
		})
	})

	Context("Output without next", func() {
		var (
			outputStr   string
			err         error
			pageOptions = apilist.PageOptions{PageSize: 2, NoPagination: true}
		)

		BeforeAll(func() {
			response := &openapi.GetProjects200Response{Projects: []openapi.Project{
				generateProjectFullSpec("1", "project-01", time.Time{}),
				generateProjectFullSpec("2", "project-02", time.Unix(10, 0)),
			}}
			mockpage.EXPECT().List(gomock.Any(), int32(3), int32(0), fb.StringList()).Return(response, nil)
			out, err = service.List(transaction.NewTxn(), &pageOptions, fb)
			Expect(err).ToNot(HaveOccurred())
		})

		It("Table", func() {
			outputStr, err = out.GenerateTable()
			Expect(err).ToNot(HaveOccurred())
			Expect(outputStr).To(ContainSubstring("project-01"))
			Expect(outputStr).To(ContainSubstring("project-02"))
			Expect(outputStr).ToNot(ContainSubstring("next"))
		})

		It("Json", func() {
			outputStr, err = out.GenerateJson()
			Expect(err).ToNot(HaveOccurred())
			Expect(outputStr).To(ContainSubstring("\"name\": \"project-01\""))
			Expect(outputStr).To(ContainSubstring("\"name\": \"project-02\""))
			Expect(outputStr).ToNot(ContainSubstring("\"next\""))
		})

		It("Yaml", func() {
			outputStr, err = out.GenerateYaml()
			Expect(err).ToNot(HaveOccurred())
			Expect(outputStr).To(ContainSubstring("name: project-01"))
			Expect(outputStr).To(ContainSubstring("name: project-02"))
			Expect(outputStr).ToNot(ContainSubstring("next:"))
		})
	})
})

func generateProjectFullSpec(id, name string, createdAt time.Time) openapi.Project {
	// removed created by for page mock changes
	projSample := `{
      "clusterId": "b754457d-7374-4da4-a43f-ac13fb899838",
      "createdAt": "2025-04-09T12:51:39.718826Z",
      "createdBy": "test@example.com",
      "description": "",
      "effective": {
        "defaultNodePools": [
          "default"
        ],
        "nodeTypes": {}
      },
      "enforceRunaiScheduler": true,
      "id": "{{ .id }}",
      "name": "{{ .name }}",
      "nodeTypes": {},
      "overtimeData": {
        "range24hData": {
          "averageGpuAllocation": 1.914,
          "averageGpuUtilization": 99.483,
          "updatedAt": "2025-04-09T16:55:00.420334Z"
        },
        "range7dData": {
          "averageGpuAllocation": 1.931,
          "averageGpuUtilization": 99.587,
          "updatedAt": "2025-04-09T17:55:10.352268Z"
        }
      },
      "parent": {
        "id": "4500000",
        "name": "default"
      },
      "parentId": "4500000",
      "requestedNamespace": "",
      "resources": [
        {
          "gpu": {
            "deserved": 8,
            "limit": -1,
            "overQuotaWeight": 8
          },
          "nodePool": {
            "id": "100000",
            "name": "default"
          },
          "priority": "MediumLow"
        }
      ],
      "status": {
        "additionalStatusData": { },
        "namespace": "runai-team-one",
        "nodePoolQuotaStatuses": [
          {
            "allocated": {
              "cpu": 200,
              "gpu": 2,
              "memory": 200
            },
            "allocatedNonPreemptible": {
              "cpu": 200,
              "gpu": 2,
              "memory": 200
            },
            "nodePoolId": "100000",
            "nodePoolName": "default",
            "requested": {
              "cpu": 200,
              "gpu": 2,
              "memory": 200
            }
          }
        ],
        "phase": "Ready",
        "phaseMessage": "",
        "quotaStatus": {
          "allocated": {
            "cpu": 200,
            "gpu": 2,
            "memory": 200
          },
          "allocatedNonPreemptible": {
            "cpu": 200,
            "gpu": 2,
            "memory": 200
          },
          "requested": {
            "cpu": 200,
            "gpu": 2,
            "memory": 200
          }
        }
      },
      "totalResources": {
        "gpuQuota": 8
      },
      "updatedAt": "2025-04-09T12:51:39.718826Z"
    }`

	stringBuf := bytes.Buffer{}
	temp := template.Must(template.New("project").Parse(projSample))
	err := temp.Execute(&stringBuf, map[string]interface{}{
		"name": name,
		"id":   id,
	})
	if err != nil {
		panic(err)
	}

	project := openapi.Project{}
	err = json.Unmarshal(stringBuf.Bytes(), &project)
	if err != nil {
		panic(err)
	}
	if !createdAt.IsZero() {
		project.CreatedAt = createdAt
	}
	return project
}
