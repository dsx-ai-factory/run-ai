/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package pvc_test

import (
	"context"
	"fmt"

	cpmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/pvc"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

const (
	projectId = "proj"
)

var _ = Describe("Pvc List", func() {
	var (
		mockClient *cpmocks.MockPvcLister
		service    *pvc.Service
	)

	BeforeEach(func() {
		mockCtrl := gomock.NewController(GinkgoT())
		mockClient = cpmocks.NewMockPvcLister(mockCtrl)
		service = pvc.NewListerService(mockClient)
	})

	Context("ListPvcs", func() {
		It("return error ", func() {
			mockClient.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("general error"))
			_, err := service.ListPvcsBrief(context.TODO(), projectId)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("list API call error: general error"))
		})

		It("empty response", func() {
			mockClient.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, nil)
			_, err := service.ListPvcsBrief(context.TODO(), projectId)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("list attribute is missing"))
		})

		It("empty list", func() {
			mockClient.EXPECT().List(gomock.Any(), gomock.Any()).Return(&openapi.PVCListResponse{}, nil)
			listerService := pvc.NewListerService(mockClient)
			out, err := listerService.ListPvcsBrief(context.TODO(), projectId)
			Expect(err).ToNot(HaveOccurred())
			Expect(out.GetList()).To(BeNil())
		})

		It("all permutation", func() {
			assets, err := openapi.NewOriginFromValue("assets")
			existing, err := openapi.NewOriginFromValue("existing")
			mockClient.EXPECT().List(gomock.Any(), gomock.Any()).Return(&openapi.PVCListResponse{Pvcs: []openapi.PVCListInfo{
				{
					Name: "origin-nil",
				},
				{
					Name:   "origin-assets",
					Origin: *openapi.NewNullableOrigin(assets),
				},
				{
					Name:   "origin-existing",
					Origin: *openapi.NewNullableOrigin(existing),
				},
			}}, nil)
			out, err := service.ListPvcsBrief(context.TODO(), projectId)
			Expect(err).ToNot(HaveOccurred())
			Expect(out.GetList()).To(HaveLen(3))
			Expect(out.GetList()).To(ContainElement(pvc.Brief{
				Name:    "origin-nil",
				Managed: nil,
			}))
			Expect(out.GetList()).To(ContainElement(pvc.Brief{
				Name:    "origin-assets",
				Managed: ptr.To(true),
			}))
			Expect(out.GetList()).To(ContainElement(pvc.Brief{
				Name:    "origin-existing",
				Managed: ptr.To(false),
			}))
		})
	})
})
