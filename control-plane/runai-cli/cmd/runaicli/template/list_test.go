/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package template_test

import (
	"context"
	"encoding/json"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	cptemplatemocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane/template"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/template/list"
	wcgenerated "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
)

var _ = Describe("Template List", func() {
	var (
		ctrl       *gomock.Controller
		mockLister *cptemplatemocks.MockTemplateLister
		service    *list.Service
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockLister = cptemplatemocks.NewMockTemplateLister(ctrl)
		service = list.NewService(mockLister, "cluster-1")
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("Basic Listing", func() {
		It("should list templates and return correct data", func() {
			mockTemplates := []wcgenerated.SupersetTemplate{
				{
					Meta: wcgenerated.TemplateMetaFields{
						Id:           "template-1",
						Name:         "pytorch-training",
						ScopeType:    wcgenerated.SCOPETYPE_PROJECT,
						WorkloadType: wcgenerated.WORKLOADTYPE_TRAINING,
					},
				},
				{
					Meta: wcgenerated.TemplateMetaFields{
						Id:           "template-2",
						Name:         "tensorflow-dist",
						ScopeType:    wcgenerated.SCOPETYPE_DEPARTMENT,
						WorkloadType: wcgenerated.WORKLOADTYPE_TRAINING,
					},
				},
			}

			mockResponse := &wcgenerated.TemplatesListResponse{}
			mockResponse.SetEntries(mockTemplates)

			mockLister.EXPECT().
				List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(mockResponse, nil)

			params := list.GetListParams{
				ProjectId: "project-1",
				All:       false,
			}
			output, err := service.GetTemplateList(context.Background(), params)

			Expect(err).ToNot(HaveOccurred())
			Expect(output).ToNot(BeNil())

			templates := output.GetList()
			Expect(templates).To(HaveLen(2))
			Expect(templates[0].Name).To(Equal("pytorch-training"))
			Expect(templates[0].ScopeType).To(Equal("project"))
			Expect(templates[1].Name).To(Equal("tensorflow-dist"))
			Expect(templates[1].ScopeType).To(Equal("department"))
		})

		It("should return error when API fails", func() {
			mockLister.EXPECT().
				List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(nil, context.DeadlineExceeded)

			params := list.GetListParams{
				ProjectId: "project-1",
				All:       false,
			}
			output, err := service.GetTemplateList(context.Background(), params)

			Expect(err).To(HaveOccurred())
			Expect(output).To(BeNil())
		})

		It("should return empty list when no templates exist", func() {
			var mockTemplates []wcgenerated.SupersetTemplate
			mockResponse := &wcgenerated.TemplatesListResponse{Entries: mockTemplates, Pagination: nil}

			mockLister.EXPECT().
				List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(mockResponse, nil)

			params := list.GetListParams{
				ProjectId: "project-1",
				All:       false,
			}
			output, err := service.GetTemplateList(context.Background(), params)

			Expect(err).ToNot(HaveOccurred())
			Expect(output.GetList()).To(BeEmpty())
		})
	})

	Describe("Output Formats", func() {
		It("should generate table with template data", func() {
			mockTemplates := []wcgenerated.SupersetTemplate{
				{
					Meta: wcgenerated.TemplateMetaFields{
						Id:           "template-1",
						Name:         "pytorch-training",
						ScopeType:    wcgenerated.SCOPETYPE_PROJECT,
						WorkloadType: wcgenerated.WORKLOADTYPE_TRAINING,
					},
					Spec: &wcgenerated.SupersetSpec{
						Image: *wcgenerated.NewNullableString(wcgenerated.PtrString("pytorch/pytorch:2.0.1")),
					},
				},
			}
			mockResponse := &wcgenerated.TemplatesListResponse{Entries: mockTemplates, Pagination: nil}

			mockLister.EXPECT().
				List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(mockResponse, nil)

			params := list.GetListParams{ProjectId: "project-1", All: false}
			output, err := service.GetTemplateList(context.Background(), params)
			Expect(err).ToNot(HaveOccurred())

			tableOutput, err := output.GenerateTable()
			Expect(err).ToNot(HaveOccurred())
			Expect(tableOutput).To(ContainSubstring("pytorch-training"))
			Expect(tableOutput).To(ContainSubstring("pytorch/pytorch:2.0.1"))
			Expect(tableOutput).To(ContainSubstring("project"))
		})

		It("should generate JSON with correct structure", func() {
			mockTemplates := []wcgenerated.SupersetTemplate{
				{
					Meta: wcgenerated.TemplateMetaFields{
						Id:           "template-1",
						Name:         "pytorch-training",
						ScopeType:    wcgenerated.SCOPETYPE_PROJECT,
						WorkloadType: wcgenerated.WORKLOADTYPE_TRAINING,
					},
				},
			}
			mockResponse := &wcgenerated.TemplatesListResponse{Entries: mockTemplates, Pagination: nil}

			mockLister.EXPECT().
				List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(mockResponse, nil)

			params := list.GetListParams{ProjectId: "project-1", All: false}
			output, err := service.GetTemplateList(context.Background(), params)
			Expect(err).ToNot(HaveOccurred())

			jsonOutput, err := output.GenerateJson()
			Expect(err).ToNot(HaveOccurred())

			var result map[string]interface{}
			err = json.Unmarshal([]byte(jsonOutput), &result)
			Expect(err).ToNot(HaveOccurred())

			templates := result["templates"].([]interface{})
			Expect(templates).To(HaveLen(1))

			firstTemplate := templates[0].(map[string]interface{})
			Expect(firstTemplate["name"]).To(Equal("pytorch-training"))
			Expect(firstTemplate["scopeType"]).To(Equal("project"))
			Expect(firstTemplate["workloadType"]).To(Equal("Training"))
		})
	})

	Describe("Pagination", func() {
		It("should paginate through multiple pages using --next-token flag", func() {
			var firstPageOutput, secondPageOutput *list.Output
			var nextOffset float32

			By("Setting up mock responses for 2-page pagination")
			nextOffset = float32(1)

			mockLister.EXPECT().
				List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Nil()).
				Return(&wcgenerated.TemplatesListResponse{
					Entries: []wcgenerated.SupersetTemplate{
						{
							Meta: wcgenerated.TemplateMetaFields{
								Id:           "template-1",
								Name:         "first-template",
								ScopeType:    wcgenerated.SCOPETYPE_PROJECT,
								WorkloadType: wcgenerated.WORKLOADTYPE_TRAINING,
							},
						},
					},
					Pagination: &wcgenerated.PaginationInfo{NextOffset: &nextOffset},
				}, nil)

			offsetInt32 := int32(1)
			mockLister.EXPECT().
				List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(&offsetInt32)).
				Return(&wcgenerated.TemplatesListResponse{
					Entries: []wcgenerated.SupersetTemplate{
						{
							Meta: wcgenerated.TemplateMetaFields{
								Id:           "template-2",
								Name:         "second-template",
								ScopeType:    wcgenerated.SCOPETYPE_PROJECT,
								WorkloadType: wcgenerated.WORKLOADTYPE_TRAINING,
							},
						},
					},
					Pagination: nil,
				}, nil)

			By("Fetching the first page without offset")
			params := list.GetListParams{ProjectId: "project-1", All: false}
			firstPageOutput, err := service.GetTemplateList(context.Background(), params)
			Expect(err).ToNot(HaveOccurred())

			By("Verifying first page returns correct template")
			firstPageList := firstPageOutput.GetList()
			Expect(firstPageList).To(HaveLen(1))
			Expect(firstPageList[0].Name).To(Equal("first-template"))

			By("Verifying first page includes nextToken")
			firstPageJSON, err := firstPageOutput.GenerateJson()
			Expect(err).ToNot(HaveOccurred())
			Expect(firstPageJSON).To(ContainSubstring("nextToken"))

			By("Fetching the second page using offset from first response")
			paramsWithOffset := list.GetListParams{
				ProjectId: "project-1",
				All:       false,
				Offset:    &offsetInt32,
			}
			secondPageOutput, err = service.GetTemplateList(context.Background(), paramsWithOffset)
			Expect(err).ToNot(HaveOccurred())

			By("Verifying second page returns different template")
			secondPageList := secondPageOutput.GetList()
			Expect(secondPageList).To(HaveLen(1))
			Expect(secondPageList[0].Name).To(Equal("second-template"))

			By("Verifying second page has no nextToken (end of results)")
			secondPageJSON, err := secondPageOutput.GenerateJson()
			Expect(err).ToNot(HaveOccurred())
			var result map[string]interface{}
			json.Unmarshal([]byte(secondPageJSON), &result)
			_, hasNextToken := result["nextToken"]
			Expect(hasNextToken).To(BeFalse(), "Last page should not have nextToken")
		})

		It("should limit results using --page-size flag", func() {
			mockTemplates := []wcgenerated.SupersetTemplate{
				{
					Meta: wcgenerated.TemplateMetaFields{
						Id:           "template-1",
						Name:         "limited-template",
						ScopeType:    wcgenerated.SCOPETYPE_PROJECT,
						WorkloadType: wcgenerated.WORKLOADTYPE_TRAINING,
					},
				},
			}
			mockResponse := &wcgenerated.TemplatesListResponse{Entries: mockTemplates, Pagination: nil}

			limit := int32(10)
			mockLister.EXPECT().
				List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(&limit), gomock.Any()).
				Return(mockResponse, nil)

			params := list.GetListParams{
				ProjectId: "project-1",
				All:       false,
				Limit:     &limit,
			}
			output, err := service.GetTemplateList(context.Background(), params)

			Expect(err).ToNot(HaveOccurred())
			Expect(output.GetList()).To(HaveLen(1))
		})
	})

	Describe("Custom Columns", func() {
		It("should support custom column selection", func() {
			mockTemplates := []wcgenerated.SupersetTemplate{
				{
					Meta: wcgenerated.TemplateMetaFields{
						Id:           "template-1",
						Name:         "test-template",
						ScopeType:    wcgenerated.SCOPETYPE_PROJECT,
						WorkloadType: wcgenerated.WORKLOADTYPE_TRAINING,
					},
					Spec: &wcgenerated.SupersetSpec{
						Image: *wcgenerated.NewNullableString(wcgenerated.PtrString("nginx:latest")),
					},
				},
			}
			mockResponse := &wcgenerated.TemplatesListResponse{Entries: mockTemplates, Pagination: nil}

			mockLister.EXPECT().
				List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(mockResponse, nil)

			params := list.GetListParams{ProjectId: "project-1", All: false}
			output, err := service.GetTemplateList(context.Background(), params)
			Expect(err).ToNot(HaveOccurred())

			err = output.SetColumns([]string{"name", "image"})
			Expect(err).ToNot(HaveOccurred())

			tableOutput, err := output.GenerateTable()
			Expect(err).ToNot(HaveOccurred())
			Expect(tableOutput).To(ContainSubstring("test-template"))
			Expect(tableOutput).To(ContainSubstring("nginx:latest"))
			Expect(tableOutput).To(ContainSubstring("Template"))
			Expect(tableOutput).To(ContainSubstring("Image"))
		})
	})
})
