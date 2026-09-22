/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list

import (
	"context"
	"encoding/json"
	"time"

	cptemplatemocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane/template"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/cli/structs"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/template/list/manipulate"
	wcgenerated "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

var _ = Describe("Service", func() {
	var (
		ctrl       *gomock.Controller
		mockLister *cptemplatemocks.MockTemplateLister
		service    *Service
		ctx        context.Context
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockLister = cptemplatemocks.NewMockTemplateLister(ctrl)
		service = NewService(mockLister, "test-cluster-id")
		ctx = context.Background()
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("GetTemplateList", func() {
		Context("with All flag", func() {
			It("should not add project filter", func() {
				response := &wcgenerated.TemplatesListResponse{}
				response.SetEntries([]wcgenerated.SupersetTemplate{})

				mockLister.EXPECT().
					List(gomock.Any(), gomock.Len(0), false, gomock.Any(), gomock.Any()).
					Return(response, nil)

				params := GetListParams{
					ProjectId: "project-123",
					All:       true,
				}

				_, err := service.GetTemplateList(ctx, params)

				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("without All flag", func() {
			It("should add project filter and include ancestors", func() {
				response := &wcgenerated.TemplatesListResponse{}
				response.SetEntries([]wcgenerated.SupersetTemplate{})

				mockLister.EXPECT().
					List(gomock.Any(), gomock.Eq([]string{"projectId==project-123"}), true, gomock.Any(), gomock.Any()).
					Return(response, nil)

				params := GetListParams{
					ProjectId: "project-123",
					All:       false,
				}

				_, err := service.GetTemplateList(ctx, params)

				Expect(err).ToNot(HaveOccurred())
			})

			It("should not add filter when ProjectId is empty", func() {
				response := &wcgenerated.TemplatesListResponse{}
				response.SetEntries([]wcgenerated.SupersetTemplate{})

				mockLister.EXPECT().
					List(gomock.Any(), gomock.Len(0), false, gomock.Any(), gomock.Any()).
					Return(response, nil)

				params := GetListParams{
					ProjectId: "",
					All:       false,
				}

				_, err := service.GetTemplateList(ctx, params)

				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with pagination params", func() {
			It("should pass limit and offset to lister", func() {
				response := &wcgenerated.TemplatesListResponse{}
				response.SetEntries([]wcgenerated.SupersetTemplate{})

				limit := int32(50)
				offset := int32(100)

				mockLister.EXPECT().
					List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Eq(&limit), gomock.Eq(&offset)).
					Return(response, nil)

				params := GetListParams{
					All:    true,
					Limit:  &limit,
					Offset: &offset,
				}

				_, err := service.GetTemplateList(ctx, params)

				Expect(err).ToNot(HaveOccurred())
			})

			It("should handle nil limit and offset", func() {
				response := &wcgenerated.TemplatesListResponse{}
				response.SetEntries([]wcgenerated.SupersetTemplate{})

				mockLister.EXPECT().
					List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Nil(), gomock.Nil()).
					Return(response, nil)

				params := GetListParams{
					All:    true,
					Limit:  nil,
					Offset: nil,
				}

				_, err := service.GetTemplateList(ctx, params)

				Expect(err).ToNot(HaveOccurred())
			})
		})

		Context("with next offset in response", func() {
			It("should convert nextOffset to NextToken", func() {
				nextOffset := float32(150)
				pagination := wcgenerated.PaginationInfo{}
				pagination.SetNextOffset(nextOffset)

				response := &wcgenerated.TemplatesListResponse{}
				response.SetEntries([]wcgenerated.SupersetTemplate{})
				response.SetPagination(pagination)

				mockLister.EXPECT().
					List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(response, nil)

				params := GetListParams{All: true}

				output, err := service.GetTemplateList(ctx, params)

				Expect(err).ToNot(HaveOccurred())
				Expect(output).ToNot(BeNil())

				tableOutput, err := output.GenerateTable()
				Expect(err).ToNot(HaveOccurred())
				Expect(tableOutput).To(ContainSubstring("next token:"))
			})

			It("should not create NextToken when nextOffset is 0", func() {
				nextOffset := float32(0)
				pagination := wcgenerated.PaginationInfo{}
				pagination.SetNextOffset(nextOffset)

				response := &wcgenerated.TemplatesListResponse{}
				response.SetEntries([]wcgenerated.SupersetTemplate{})
				response.SetPagination(pagination)

				mockLister.EXPECT().
					List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(response, nil)

				params := GetListParams{All: true}

				output, err := service.GetTemplateList(ctx, params)

				Expect(err).ToNot(HaveOccurred())

				// Verify no next token (check via table generation)
				tableOutput, err := output.GenerateTable()
				Expect(err).ToNot(HaveOccurred())
				Expect(tableOutput).ToNot(ContainSubstring("next token:"))
			})

			It("should not create NextToken when nextOffset is not set", func() {
				pagination := wcgenerated.PaginationInfo{}

				response := &wcgenerated.TemplatesListResponse{}
				response.SetEntries([]wcgenerated.SupersetTemplate{})
				response.SetPagination(pagination)

				mockLister.EXPECT().
					List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(response, nil)

				params := GetListParams{All: true}

				output, err := service.GetTemplateList(ctx, params)

				Expect(err).ToNot(HaveOccurred())

				// Verify no next token (check via table generation)
				tableOutput, err := output.GenerateTable()
				Expect(err).ToNot(HaveOccurred())
				Expect(tableOutput).ToNot(ContainSubstring("next token:"))
			})
		})

		Context("with templates in response", func() {
			It("should return templates", func() {
				template1 := wcgenerated.SupersetTemplate{
					Meta: wcgenerated.TemplateMetaFields{
						Name:         "template-1",
						ScopeType:    wcgenerated.SCOPETYPE_PROJECT,
						WorkloadType: wcgenerated.WORKLOADTYPE_TRAINING,
						CreatedAt:    time.Now(),
					},
				}
				template2 := wcgenerated.SupersetTemplate{
					Meta: wcgenerated.TemplateMetaFields{
						Name:         "template-2",
						ScopeType:    wcgenerated.SCOPETYPE_CLUSTER,
						WorkloadType: wcgenerated.WORKLOADTYPE_WORKSPACE,
						CreatedAt:    time.Now(),
					},
				}

				response := &wcgenerated.TemplatesListResponse{}
				response.SetEntries([]wcgenerated.SupersetTemplate{template1, template2})

				mockLister.EXPECT().
					List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(response, nil)

				params := GetListParams{All: true}

				output, err := service.GetTemplateList(ctx, params)

				Expect(err).ToNot(HaveOccurred())
				Expect(output).ToNot(BeNil())

				templates := output.GetList()
				Expect(templates).To(HaveLen(2))
				Expect(templates[0].Name).To(Equal("template-1"))
				Expect(templates[1].Name).To(Equal("template-2"))
			})

			It("should handle empty templates list", func() {
				response := &wcgenerated.TemplatesListResponse{}
				response.SetEntries([]wcgenerated.SupersetTemplate{})

				mockLister.EXPECT().
					List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(response, nil)

				params := GetListParams{All: true}

				output, err := service.GetTemplateList(ctx, params)

				Expect(err).ToNot(HaveOccurred())
				Expect(output).ToNot(BeNil())

				templates := output.GetList()
				Expect(templates).To(HaveLen(0))
			})
		})

		Context("error handling", func() {
			It("should return error when lister fails", func() {
				mockLister.EXPECT().
					List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(nil, context.DeadlineExceeded)

				params := GetListParams{All: true}

				output, err := service.GetTemplateList(ctx, params)

				Expect(err).To(MatchError(context.DeadlineExceeded))
				Expect(output).To(BeNil())
			})

			It("should ignore negative nextOffset (treat as no next page)", func() {
				nextOffset := float32(-50)
				pagination := wcgenerated.PaginationInfo{}
				pagination.SetNextOffset(nextOffset)

				response := &wcgenerated.TemplatesListResponse{}
				response.SetEntries([]wcgenerated.SupersetTemplate{})
				response.SetPagination(pagination)

				mockLister.EXPECT().
					List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(response, nil)

				params := GetListParams{All: true}

				output, err := service.GetTemplateList(ctx, params)

				Expect(err).ToNot(HaveOccurred())
				Expect(output).ToNot(BeNil())

				tableOutput, err := output.GenerateTable()
				Expect(err).ToNot(HaveOccurred())
				Expect(tableOutput).ToNot(ContainSubstring("next token:"))
			})
		})

		Context("Output Formats", func() {
			var (
				template1 wcgenerated.SupersetTemplate
				template2 wcgenerated.SupersetTemplate
			)

			BeforeEach(func() {
				template1 = wcgenerated.SupersetTemplate{
					Meta: wcgenerated.TemplateMetaFields{
						Name:         "training-template",
						ScopeType:    wcgenerated.SCOPETYPE_PROJECT,
						WorkloadType: wcgenerated.WORKLOADTYPE_TRAINING,
						CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					},
				}
				template1.Spec = wcgenerated.NewSupersetSpec()
				template1.Spec.SetImage("nginx:latest")
				compute := wcgenerated.NewWorkspaceSpecSpecAllOfCompute()
				compute.SetGpuDevicesRequest(2)
				template1.Spec.SetCompute(*compute)
				template1.Runnable = &wcgenerated.RunnableInfo{}
				template1.Runnable.SetIsRunnable(true)

				template2 = wcgenerated.SupersetTemplate{
					Meta: wcgenerated.TemplateMetaFields{
						Name:         "workspace-template",
						ScopeType:    wcgenerated.SCOPETYPE_CLUSTER,
						WorkloadType: wcgenerated.WORKLOADTYPE_WORKSPACE,
						CreatedAt:    time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
					},
				}
				template2.Runnable = &wcgenerated.RunnableInfo{}
				template2.Runnable.SetIsRunnable(false)
			})

			Context("JSON format", func() {
				It("should generate JSON with template data", func() {
					response := &wcgenerated.TemplatesListResponse{}
					response.SetEntries([]wcgenerated.SupersetTemplate{template1, template2})

					mockLister.EXPECT().
						List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(response, nil)

					params := GetListParams{All: true}
					output, err := service.GetTemplateList(ctx, params)

					Expect(err).ToNot(HaveOccurred())

					jsonOutput, err := output.GenerateJson()
					Expect(err).ToNot(HaveOccurred())

					var data manipulate.BriefData
					err = json.Unmarshal([]byte(jsonOutput), &data)
					Expect(err).ToNot(HaveOccurred())

					Expect(data.Templates).To(HaveLen(2))
					Expect(data.Templates[0].Name).To(Equal("training-template"))
					Expect(data.Templates[1].Name).To(Equal("workspace-template"))
				})
			})

			Context("YAML format", func() {
				It("should generate YAML with template data", func() {
					response := &wcgenerated.TemplatesListResponse{}
					response.SetEntries([]wcgenerated.SupersetTemplate{template1, template2})

					mockLister.EXPECT().
						List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
						Return(response, nil)

					params := GetListParams{All: true}
					output, err := service.GetTemplateList(ctx, params)

					Expect(err).ToNot(HaveOccurred())

					yamlOutput, err := output.GenerateYaml()
					Expect(err).ToNot(HaveOccurred())

					var data manipulate.BriefData
					err = yaml.Unmarshal([]byte(yamlOutput), &data)
					Expect(err).ToNot(HaveOccurred())

					Expect(data.Templates).To(HaveLen(2))
				})
			})
		})
	})

	Describe("NewService", func() {
		It("should create a new Service with lister and cluster ID", func() {
			lister := cptemplatemocks.NewMockTemplateLister(ctrl)
			clusterId := "test-cluster-123"

			service := NewService(lister, clusterId)

			Expect(service).ToNot(BeNil())
			Expect(service.lister).To(Equal(lister))
			Expect(service.clusterId).To(Equal(clusterId))
		})
	})
})

var _ = Describe("Pagination Functions", func() {
	Describe("offsetToNextToken", func() {
		It("should convert positive offset to NextToken", func() {
			offset := int32(100)

			token, err := offsetToNextToken(offset)

			Expect(err).ToNot(HaveOccurred())
			Expect(token).ToNot(BeNil())
			Expect(token.IsSet()).To(BeTrue())

			timestamp := token.GetTimestamp()
			Expect(timestamp.Unix()).To(Equal(int64(100)))
		})

		It("should handle offset of 0", func() {
			offset := int32(0)

			token, err := offsetToNextToken(offset)

			Expect(err).ToNot(HaveOccurred())
			Expect(token).ToNot(BeNil())
			Expect(token.IsSet()).To(BeTrue())

			timestamp := token.GetTimestamp()
			Expect(timestamp.Unix()).To(Equal(int64(0)))
		})

		It("should handle large offsets", func() {
			offset := int32(999999)

			token, err := offsetToNextToken(offset)

			Expect(err).ToNot(HaveOccurred())
			Expect(token).ToNot(BeNil())

			timestamp := token.GetTimestamp()
			Expect(timestamp.Unix()).To(Equal(int64(999999)))
		})

		It("should return error for negative offset", func() {
			offset := int32(-10)

			token, err := offsetToNextToken(offset)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("offset cannot be negative"))
			Expect(token).To(BeNil())
		})

		It("should create UTC timestamp", func() {
			offset := int32(50)

			token, err := offsetToNextToken(offset)

			Expect(err).ToNot(HaveOccurred())
			timestamp := token.GetTimestamp()
			Expect(timestamp.Location()).To(Equal(time.UTC))
		})
	})

	Describe("NextTokenToOffset", func() {
		It("should convert NextToken back to offset", func() {
			originalOffset := int32(100)
			token, _ := offsetToNextToken(originalOffset)

			offset, err := NextTokenToOffset(token)

			Expect(err).ToNot(HaveOccurred())
			Expect(offset).ToNot(BeNil())
			Expect(*offset).To(Equal(originalOffset))
		})

		It("should handle nil token", func() {
			offset, err := NextTokenToOffset(nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(offset).To(BeNil())
		})

		It("should handle unset token", func() {
			token := &structs.NextToken{}

			offset, err := NextTokenToOffset(token)

			Expect(err).ToNot(HaveOccurred())
			Expect(offset).To(BeNil())
		})

		It("should handle offset of 0", func() {
			token, _ := offsetToNextToken(0)

			offset, err := NextTokenToOffset(token)

			Expect(err).ToNot(HaveOccurred())
			Expect(offset).ToNot(BeNil())
			Expect(*offset).To(Equal(int32(0)))
		})

		It("should handle large offsets", func() {
			originalOffset := int32(999999)
			token, _ := offsetToNextToken(originalOffset)

			offset, err := NextTokenToOffset(token)

			Expect(err).ToNot(HaveOccurred())
			Expect(offset).ToNot(BeNil())
			Expect(*offset).To(Equal(originalOffset))
		})

		It("should return error for negative timestamp (invalid token)", func() {
			invalidTime := time.Unix(-100, 0)
			token, _ := structs.NextTokenFromTimestamp(invalidTime)

			offset, err := NextTokenToOffset(token)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("offset cannot be negative"))
			Expect(offset).To(BeNil())
		})
	})

	Describe("Round-trip conversion", func() {
		It("should preserve offset through round-trip conversion", func() {
			testOffsets := []int32{0, 1, 50, 100, 1000, 50000, 999999}

			for _, originalOffset := range testOffsets {
				token, err := offsetToNextToken(originalOffset)
				Expect(err).ToNot(HaveOccurred())

				resultOffset, err := NextTokenToOffset(token)
				Expect(err).ToNot(HaveOccurred())
				Expect(resultOffset).ToNot(BeNil())
				Expect(*resultOffset).To(Equal(originalOffset), "Failed for offset: %d", originalOffset)
			}
		})
	})
})
