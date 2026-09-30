/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list_test

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/filter"
	cpmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	apilist "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/application/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

var _ = Describe("AIApplication Service Test", func() {
	const (
		appIDOne    = "app-id-1"
		appNameOne  = "app-one"
		appTypeOne  = "inference"
		appProjOne  = "team-alpha"
		appIDTwo    = "app-id-2"
		appNameTwo  = "app-two"
		appProjTwo  = "team-beta"
		clusterUUID = "cluster-uuid-123"
		someError   = "some client error"
	)

	var (
		ctrl       *gomock.Controller
		mockLister *cpmocks.MockAIApplicationLister
		service    *list.AIApplicationService
	)

	var (
		appOne = openapi.AIApplication{
			Id:          ptr.To(appIDOne),
			Name:        ptr.To(appNameOne),
			Type:        ptr.To(appTypeOne),
			ProjectName: ptr.To(appProjOne),
		}
		appTwo = openapi.AIApplication{
			Id:          ptr.To(appIDTwo),
			Name:        ptr.To(appNameTwo),
			ProjectName: ptr.To(appProjTwo),
		}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockLister = cpmocks.NewMockAIApplicationLister(ctrl)
		service = list.NewAIApplicationService(mockLister)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Describe("NewAIApplicationService", func() {
		It("should create a new AIApplicationService", func() {
			Expect(service).NotTo(BeNil())
		})
	})

	Describe("List", func() {
		var (
			fb          *filter.Builder
			twoAppsResp *openapi.ListApplications200Response
			oneAppResp  *openapi.ListApplications200Response
		)

		BeforeEach(func() {
			fb = filter.NewBuilder().WithClusterID(clusterUUID)
			twoAppsResp = &openapi.ListApplications200Response{
				Applications: []openapi.AIApplication{appOne, appTwo},
			}
			oneAppResp = &openapi.ListApplications200Response{
				Applications: []openapi.AIApplication{appOne},
			}
		})

		Context("when the client returns successfully", func() {
			It("should return output without error - no filters", func() {
				mockLister.EXPECT().List(gomock.Any(), int32(51), int32(0), nil).Return(twoAppsResp, nil).AnyTimes()

				output, err := service.List(transaction.NewTxn(), &apilist.PageOptions{PageSize: 50}, nil)

				Expect(err).NotTo(HaveOccurred())
				Expect(output).NotTo(BeNil())

				jsonStr, err := output.GenerateJson()
				Expect(err).To(BeNil())

				var data list.BriefData
				err = json.Unmarshal([]byte(jsonStr), &data)
				Expect(err).To(BeNil())
				Expect(data.NextToken).To(BeNil())
				Expect(data.AIApplications).To(HaveLen(2))
				Expect(data.AIApplications[0].Name).To(Equal(appNameOne))
				Expect(data.AIApplications[0].Type).To(Equal(appTypeOne))
				Expect(data.AIApplications[0].ProjectName).To(Equal(appProjOne))
			})

			It("should return output without error - with filters", func() {
				mockLister.EXPECT().List(gomock.Any(), int32(51), int32(0), fb.StringList()).Return(oneAppResp, nil).AnyTimes()

				output, err := service.List(transaction.NewTxn(), &apilist.PageOptions{PageSize: 50}, fb)

				Expect(err).NotTo(HaveOccurred())
				Expect(output).NotTo(BeNil())

				jsonStr, err := output.GenerateJson()
				Expect(err).To(BeNil())

				var data list.BriefData
				err = json.Unmarshal([]byte(jsonStr), &data)
				Expect(err).To(BeNil())
				Expect(data.AIApplications).To(HaveLen(1))
				Expect(data.AIApplications[0].Name).To(Equal(appNameOne))
			})

			It("should return empty output when no applications", func() {
				emptyResp := &openapi.ListApplications200Response{
					Applications: []openapi.AIApplication{},
				}
				mockLister.EXPECT().List(gomock.Any(), int32(51), int32(0), nil).Return(emptyResp, nil).AnyTimes()

				output, err := service.List(transaction.NewTxn(), &apilist.PageOptions{PageSize: 50}, nil)

				Expect(err).NotTo(HaveOccurred())
				Expect(output).NotTo(BeNil())
				Expect(output.IsEmpty()).To(BeTrue())
			})
		})

		Context("when the client returns an error", func() {
			BeforeEach(func() {
				mockLister.EXPECT().List(gomock.Any(), int32(51), int32(0), nil).Return(nil, errors.New(someError)).AnyTimes()
			})

			It("should return the error", func() {
				output, err := service.List(transaction.NewTxn(), &apilist.PageOptions{PageSize: 50}, nil)

				Expect(err).To(MatchError(someError))
				Expect(output).To(BeNil())
			})
		})

		Context("when txn is nil", func() {
			It("should return an error", func() {
				output, err := service.List(nil, &apilist.PageOptions{PageSize: 50}, nil)

				Expect(err).To(HaveOccurred())
				Expect(output).To(BeNil())
			})
		})

		Context("paginated continuation", func() {
			It("should return a next token and continue from it on the following page", func() {
				firstCreatedAt := time.Unix(10, 0)
				secondCreatedAt := time.Unix(20, 0)

				pageOneApp := openapi.AIApplication{
					Id:        ptr.To(appIDOne),
					Name:      ptr.To(appNameOne),
					CreatedAt: &firstCreatedAt,
				}
				overflowApp := openapi.AIApplication{
					Id:        ptr.To(appIDTwo),
					Name:      ptr.To(appNameTwo),
					CreatedAt: &secondCreatedAt,
				}
				pageTwoApp := openapi.AIApplication{
					Id:        ptr.To("app-id-3"),
					Name:      ptr.To("app-three"),
					CreatedAt: &secondCreatedAt,
				}

				pageOneFb := filter.NewBuilder().WithClusterID(clusterUUID)
				mockLister.EXPECT().List(gomock.Any(), int32(2), int32(0), pageOneFb.StringList()).
					Return(&openapi.ListApplications200Response{
						Applications: []openapi.AIApplication{pageOneApp, overflowApp},
					}, nil)

				firstPageOptions := &apilist.PageOptions{PageSize: 1, NoPagination: true}
				output, err := service.List(transaction.NewTxn(), firstPageOptions, pageOneFb)

				Expect(err).NotTo(HaveOccurred())
				Expect(output).NotTo(BeNil())

				jsonStr, err := output.GenerateJson()
				Expect(err).To(BeNil())

				var firstPageData list.BriefData
				err = json.Unmarshal([]byte(jsonStr), &firstPageData)
				Expect(err).To(BeNil())
				Expect(firstPageData.AIApplications).To(HaveLen(1))
				Expect(firstPageData.AIApplications[0].Name).To(Equal(appNameOne))
				Expect(firstPageData.NextToken).NotTo(BeNil())

				// Continue from the next token: the follow-up call must carry the
				// createdAt cursor, matching a distinct filter than the first page.
				pageTwoFb := filter.NewBuilder().WithClusterID(clusterUUID).WithCreatedAt(secondCreatedAt)
				mockLister.EXPECT().List(gomock.Any(), int32(2), int32(0), pageTwoFb.StringList()).
					Return(&openapi.ListApplications200Response{
						Applications: []openapi.AIApplication{pageTwoApp},
					}, nil)

				secondPageOptions := &apilist.PageOptions{PageSize: 1, NoPagination: true}
				output, err = service.List(transaction.NewTxn(), secondPageOptions, pageTwoFb)

				Expect(err).NotTo(HaveOccurred())
				Expect(output).NotTo(BeNil())

				jsonStr, err = output.GenerateJson()
				Expect(err).To(BeNil())

				var secondPageData list.BriefData
				err = json.Unmarshal([]byte(jsonStr), &secondPageData)
				Expect(err).To(BeNil())
				Expect(secondPageData.AIApplications).To(HaveLen(1))
				Expect(secondPageData.AIApplications[0].Name).To(Equal("app-three"))
			})
		})
	})

	Describe("GetByName", func() {
		var fb *filter.Builder

		BeforeEach(func() {
			fb = filter.NewBuilder().WithClusterID(clusterUUID).WithProject(appProjOne).AppendName(appNameOne)
		})

		It("should return the application matching the exact name", func() {
			mockLister.EXPECT().List(gomock.Any(), int32(2), int32(0), gomock.Eq(fb.StringList())).
				Return(&openapi.ListApplications200Response{
					Applications: []openapi.AIApplication{appOne},
				}, nil)

			app, err := service.GetByName(transaction.NewTxn(), fb, appNameOne)

			Expect(err).NotTo(HaveOccurred())
			Expect(app).NotTo(BeNil())
			Expect(app.GetId()).To(Equal(appIDOne))
		})

		It("should return nil without an error when no application matches", func() {
			mockLister.EXPECT().List(gomock.Any(), int32(2), int32(0), gomock.Eq(fb.StringList())).
				Return(&openapi.ListApplications200Response{
					Applications: []openapi.AIApplication{},
				}, nil)

			app, err := service.GetByName(transaction.NewTxn(), fb, appNameOne)

			Expect(err).NotTo(HaveOccurred())
			Expect(app).To(BeNil())
		})

		It("should not return an application whose name only partially matches", func() {
			mockLister.EXPECT().List(gomock.Any(), int32(2), int32(0), gomock.Eq(fb.StringList())).
				Return(&openapi.ListApplications200Response{
					Applications: []openapi.AIApplication{appTwo},
				}, nil)

			app, err := service.GetByName(transaction.NewTxn(), fb, appNameOne)

			Expect(err).NotTo(HaveOccurred())
			Expect(app).To(BeNil())
		})

		It("should propagate client errors", func() {
			mockLister.EXPECT().List(gomock.Any(), int32(2), int32(0), gomock.Eq(fb.StringList())).
				Return(nil, errors.New(someError))

			app, err := service.GetByName(transaction.NewTxn(), fb, appNameOne)

			Expect(err).To(MatchError(someError))
			Expect(app).To(BeNil())
		})

		It("should return an error when txn is nil", func() {
			app, err := service.GetByName(nil, fb, appNameOne)

			Expect(err).To(HaveOccurred())
			Expect(app).To(BeNil())
		})
	})
})
