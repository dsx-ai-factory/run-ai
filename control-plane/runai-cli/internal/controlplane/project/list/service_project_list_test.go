/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list_test

import (
	"encoding/json"
	"errors"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/filter"
	apilist "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/api/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/project/list"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/transaction"

	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"

	"k8s.io/utils/ptr"

	cpmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Project Service Test", func() {
	const (
		ClusterUuid     = "one-uuid"
		TeamOne         = "team-a"
		ProjectIdOne    = "123"
		PhaseOne        = openapi.PROJECTPHASE_READY
		DepartmentOne   = "area 51"
		GpuQuotaOne     = 10.0
		AllocatedGpuOne = 5.0
		ClusterUuidTwo  = "two-uuid"
		TeamTwo         = "team-b"
		ProjectIDTwo    = "456"
		PhaseTwo        = openapi.PROJECTPHASE_NOT_READY
		GpuQuotaTwo     = 30.0
		AllocatedGpuTwo = 12.0
		someError       = "some client error"
	)

	var (
		ctrl    *gomock.Controller
		mocknc  *cpmocks.MockProjectLister
		service *list.ProjectService
	)

	var (
		parent = openapi.DepartmentProjectParent{
			Name: DepartmentOne,
		}
		department = openapi.NewNullableDepartmentProjectParent(&parent)
	)

	var (
		projectOne = openapi.Project{
			Name:      TeamOne,
			Id:        ProjectIdOne,
			ClusterId: ClusterUuid,
			Resources: []openapi.Resources{
				{
					Gpu: openapi.ResourcesGpu{
						Deserved: GpuQuotaOne,
					},
				},
			},
			Parent: *department,
			Status: openapi.ProjectStatus{
				Phase: ptr.To(PhaseOne),
				NodePoolQuotaStatuses: []openapi.QuotaStatusNodePool{
					{
						Allocated: &openapi.QuotaStatusResource{
							Gpu: ptr.To(AllocatedGpuOne),
						},
					},
				},
			},
		}

		projectTwo = openapi.Project{
			Name:      TeamTwo,
			Id:        ProjectIDTwo,
			ClusterId: ClusterUuidTwo,
			Resources: []openapi.Resources{
				{
					Gpu: openapi.ResourcesGpu{
						Deserved: GpuQuotaTwo,
					},
				},
			},
			Parent: *department,
			Status: openapi.ProjectStatus{
				Phase: ptr.To(PhaseTwo),
				NodePoolQuotaStatuses: []openapi.QuotaStatusNodePool{
					{
						Allocated: &openapi.QuotaStatusResource{
							Gpu: ptr.To(AllocatedGpuTwo),
						},
					},
				},
			},
		}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mocknc = cpmocks.NewMockProjectLister(ctrl)
		service = list.NewProjectService(mocknc)
	})

	AfterEach(func() {
		// We add this in order to check that all the registered mocks ware really called
		ctrl.Finish()
	})

	Describe("NewProjectService", func() {
		It("should create a new ProjectService", func() {
			Expect(service).NotTo(BeNil())
		})
	})

	Describe("GetProjectList", func() {
		var (
			fb                 *filter.Builder
			expectedProjects   *openapi.GetProjects200Response
			expectedOneProject *openapi.GetProjects200Response
		)

		BeforeEach(func() {
			fb = filter.NewBuilder().WithClusterID(ClusterUuid).WithProject(ProjectIdOne)
			expectedProjects = &openapi.GetProjects200Response{
				Next: nil,
				Projects: []openapi.Project{
					projectOne, projectTwo,
				},
			}
			expectedOneProject = &openapi.GetProjects200Response{
				Next: nil,
				Projects: []openapi.Project{
					projectOne,
				},
			}
		})

		Context("when the client returns successfully", func() {
			It("should return a Output without error - without filters", func() {
				mocknc.EXPECT().List(gomock.Any(), int32(51), int32(0), nil).Return(expectedProjects, nil).AnyTimes()
				output, err := service.List(transaction.NewTxn(), &apilist.PageOptions{PageSize: 50}, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(output).NotTo(BeNil())
				response, err := output.GenerateJson()
				Expect(err).To(BeNil())
				var data list.BriefData
				err = json.Unmarshal([]byte(response), &data)
				Expect(err).To(BeNil())
				Expect(data.NextToken).To(BeNil())

				briefProjectList := data.Projects
				Expect(len(briefProjectList)).To(Equal(2))
				Expect(briefProjectList[0].Name).To(Equal(TeamOne))
				Expect(briefProjectList[0].Id).To(Equal(ProjectIdOne))
				Expect(briefProjectList[0].Phase).To(Equal(string(PhaseOne)))
				Expect(briefProjectList[0].DepartmentName).To(Equal(DepartmentOne))
				Expect(briefProjectList[0].Resources[0].Gpu.Deserved).To(Equal(GpuQuotaOne))
				Expect(briefProjectList[0].QuotaStatuses[0].Allocated.Gpu).To(Equal(AllocatedGpuOne))
			})

			It("should return a Output without error - filters", func() {
				mocknc.EXPECT().List(gomock.Any(), int32(51), int32(0), fb.StringList()).Return(expectedOneProject, nil).AnyTimes()
				output, err := service.List(transaction.NewTxn(), &apilist.PageOptions{PageSize: 50}, fb)
				Expect(err).NotTo(HaveOccurred())
				Expect(output).NotTo(BeNil())
				response, err := output.GenerateJson()
				Expect(err).To(BeNil())
				var data list.BriefData
				err = json.Unmarshal([]byte(response), &data)
				Expect(err).To(BeNil())
				Expect(data.NextToken).To(BeNil())

				briefProjectList := data.Projects
				Expect(len(briefProjectList)).To(Equal(1))
				Expect(briefProjectList[0].Name).To(Equal(TeamOne))
				Expect(briefProjectList[0].Id).To(Equal(ProjectIdOne))
				Expect(briefProjectList[0].Phase).To(Equal(string(PhaseOne)))
				Expect(briefProjectList[0].DepartmentName).To(Equal(DepartmentOne))
				Expect(briefProjectList[0].Resources[0].Gpu.Deserved).To(Equal(GpuQuotaOne))
				Expect(briefProjectList[0].QuotaStatuses[0].Allocated.Gpu).To(Equal(AllocatedGpuOne))
			})

		})

		Context("when the client returns an error", func() {
			BeforeEach(func() {
				mocknc.EXPECT().List(gomock.Any(), int32(51), int32(0), nil).Return(nil, errors.New(someError)).AnyTimes()
			})

			It("should return the error", func() {
				handler, err := service.List(transaction.NewTxn(), &apilist.PageOptions{PageSize: 50}, nil)
				Expect(err).To(MatchError(someError))
				Expect(handler).To(BeNil())
			})
		})
	})
})
