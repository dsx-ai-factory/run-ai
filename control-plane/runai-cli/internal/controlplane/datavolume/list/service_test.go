/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list_test

import (
	"context"
	"errors"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/datavolume/list"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/datavolumes-service-client/generated"
)

type mockDatavolumeLister struct {
	ctrl     *gomock.Controller
	response *openapi.GetDatavolumes200Response
	err      error
}

func (m *mockDatavolumeLister) List(ctx context.Context, projectID *string) (*openapi.GetDatavolumes200Response, error) {
	return m.response, m.err
}

var _ = Describe("Datavolume List Service", func() {
	var (
		mockCtrl *gomock.Controller
		ctx      context.Context
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		ctx = context.Background()
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Context("List", func() {
		It("should return datavolumes successfully", func() {
			dv1 := newTestDatavolume("dv-1", "pvc1", nil, nil)
			dv2 := newTestDatavolume("dv-2", "pvc2", nil, nil)
			mockResponse := &openapi.GetDatavolumes200Response{
				Datavolumes: []openapi.DatavolumeNoSharedScopes{dv1, dv2},
			}
			mockClient := &mockDatavolumeLister{
				ctrl:     mockCtrl,
				response: mockResponse,
				err:      nil,
			}
			service := list.NewService(mockClient)

			result, err := service.List(ctx, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})

		It("should return error when API call fails", func() {
			mockClient := &mockDatavolumeLister{
				ctrl:     mockCtrl,
				response: nil,
				err:      errors.New("API error"),
			}
			service := list.NewService(mockClient)

			result, err := service.List(ctx, nil)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("datavolume API call error"))
			Expect(result).To(BeNil())
		})

		It("should handle empty datavolume list", func() {
			mockResponse := &openapi.GetDatavolumes200Response{
				Datavolumes: []openapi.DatavolumeNoSharedScopes{},
			}
			mockClient := &mockDatavolumeLister{
				ctrl:     mockCtrl,
				response: mockResponse,
				err:      nil,
			}
			service := list.NewService(mockClient)

			result, err := service.List(ctx, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})

		It("should apply project filter correctly", func() {
			dv1 := newTestDatavolume("dv-1", "pvc1", nil, nil)
			mockResponse := &openapi.GetDatavolumes200Response{
				Datavolumes: []openapi.DatavolumeNoSharedScopes{dv1},
			}
			mockClient := &mockDatavolumeLister{
				ctrl:     mockCtrl,
				response: mockResponse,
				err:      nil,
			}
			service := list.NewService(mockClient)

			projectID := "123"
			result, err := service.List(ctx, &projectID)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})
	})
})

func newTestDatavolume(name string, originPvc string, deptId *string, clusterId *string) openapi.DatavolumeNoSharedScopes {
	status := *openapi.NewDatavolumeStatus(openapi.DATAVOLUMEPHASE_READY)
	dv := openapi.DatavolumeNoSharedScopes{
		Id:            "dv-id-" + name,
		Name:          name,
		OriginPvcName: originPvc,
		ProjectId:     "proj-123",
		ProjectName:   "test-project",
		DepartmentId:  deptId,
		ClusterId:     clusterId,
		Status:        status,
	}
	return dv
}
