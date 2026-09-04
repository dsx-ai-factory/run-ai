/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package service

import (
	"context"
	"encoding/json"
	"errors"

	cpmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/department/model"

	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"

	"testing"

	"k8s.io/utils/ptr"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestJobs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Department Service")
}

var _ = Describe("Department Service Test", func() {
	const (
		ClusterUuid     = "one-uuid"
		TeamOne         = "team-a"
		DepartmentIdOne = "123"
		AllocatedGpuOne = 5.0
		ClusterUuidTwo  = "two-uuid"
		TeamTwo         = "team-b"
		DepartmentIDTwo = "456"
		AllocatedGpuTwo = 12.0
		someError       = "some client error"
	)

	var (
		ctrl    *gomock.Controller
		mocknc  *cpmocks.MockDepartmentClient
		service *DepartmentService
	)

	var (
		departmentOne = openapi.DepartmentForList{
			Name:      TeamOne,
			Id:        DepartmentIdOne,
			ClusterId: ClusterUuid,
			Status: *openapi.NewNullableDepartmentForListAllOfStatus(
				&openapi.DepartmentForListAllOfStatus{
					QuotaStatus: &openapi.QuotaStatus{
						Allocated: &openapi.QuotaStatusResource{
							Gpu: ptr.To(AllocatedGpuOne),
						},
					},
				}),
		}

		departmentTwo = openapi.DepartmentForList{
			Name:      TeamTwo,
			Id:        DepartmentIDTwo,
			ClusterId: ClusterUuidTwo,
			Status: *openapi.NewNullableDepartmentForListAllOfStatus(
				&openapi.DepartmentForListAllOfStatus{
					QuotaStatus: &openapi.QuotaStatus{
						Allocated: &openapi.QuotaStatusResource{
							Gpu: ptr.To(AllocatedGpuTwo),
						},
					},
				},
			),
		}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mocknc = cpmocks.NewMockDepartmentClient(ctrl)
		service = NewDepartmentService(mocknc)
	})

	AfterEach(func() {
		// We add this in order to check that all the registered mocks ware really called
		ctrl.Finish()
	})

	Describe("NewDepartmentService", func() {
		It("should create a new DepartmentService", func() {
			Expect(service).NotTo(BeNil())
		})
	})

	Describe("GetDepartmentList", func() {
		var (
			filters               []model.Filter
			expectedDepartments   *openapi.GetDepartments200Response
			expectedOneDepartment *openapi.GetDepartments200Response
		)

		BeforeEach(func() {
			filters = []model.Filter{
				*model.NewClusterIDFilter(ClusterUuid),
				*model.NewDepartmentFilter(DepartmentIdOne),
			}
			expectedDepartments = &openapi.GetDepartments200Response{
				Next: nil,
				Departments: []openapi.DepartmentForList{
					departmentOne, departmentTwo,
				},
			}
			expectedOneDepartment = &openapi.GetDepartments200Response{
				Next: nil,
				Departments: []openapi.DepartmentForList{
					departmentOne,
				},
			}
		})

		Context("when the client returns successfully", func() {
			It("should return a BriefDepartmentHandler without error - without filters", func() {
				mocknc.EXPECT().GetDepartmentList(gomock.Any(), nil, int32(50), int32(0)).Return(expectedDepartments, nil).AnyTimes()
				handler, err := service.GetDepartmentList(context.Background(), nil, 50, 0)
				Expect(err).NotTo(HaveOccurred())
				Expect(handler).NotTo(BeNil())
				response, err := handler.GenerateJson()
				Expect(err).To(BeNil())
				var data Data
				err = json.Unmarshal([]byte(response), &data)
				Expect(err).To(BeNil())

				briefDepartmentList := data.Departments
				Expect(len(briefDepartmentList)).To(Equal(2))
				Expect(briefDepartmentList[0].Name).To(Equal(TeamOne))
				Expect(briefDepartmentList[0].Id).To(Equal(DepartmentIdOne))
				Expect(briefDepartmentList[0].Status.QuotaStatus.Allocated.Gpu).To(Equal(float32(AllocatedGpuOne)))
			})

			It("should return a BriefDepartmentHandler without error - filters", func() {
				mocknc.EXPECT().GetDepartmentList(gomock.Any(), gomock.Any(), int32(50), int32(0)).Return(expectedOneDepartment, nil).AnyTimes()
				handler, err := service.GetDepartmentList(context.Background(), filters, 50, 0)
				Expect(err).NotTo(HaveOccurred())
				Expect(handler).NotTo(BeNil())
				response, err := handler.GenerateJson()
				Expect(err).To(BeNil())
				var data Data
				err = json.Unmarshal([]byte(response), &data)
				Expect(err).To(BeNil())

				briefDepartmentList := data.Departments
				Expect(len(briefDepartmentList)).To(Equal(1))
				Expect(briefDepartmentList[0].Name).To(Equal(TeamOne))
				Expect(briefDepartmentList[0].Id).To(Equal(DepartmentIdOne))
				Expect(briefDepartmentList[0].Status.QuotaStatus.Allocated.Gpu).To(Equal(float32(AllocatedGpuOne)))
			})

		})

		Context("when the client returns an error", func() {
			BeforeEach(func() {
				mocknc.EXPECT().GetDepartmentList(gomock.Any(), nil, int32(50), int32(0)).Return(nil, errors.New(someError)).AnyTimes()
			})

			It("should return the error", func() {
				handler, err := service.GetDepartmentList(context.Background(), nil, 50, 0)
				Expect(err).To(MatchError(someError))
				Expect(handler).To(BeNil())
			})
		})
	})
})
