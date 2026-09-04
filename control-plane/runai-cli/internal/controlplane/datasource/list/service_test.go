/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list_test

import (
	"context"
	"errors"

	cpmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/datasource/list"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Datasource List Service", func() {
	var (
		mockCtrl   *gomock.Controller
		service    *list.Service
		mockClient *cpmocks.MockDatasourceLister
		ctx        context.Context
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockClient = cpmocks.NewMockDatasourceLister(mockCtrl)
		service = list.NewService(mockClient)
		ctx = context.Background()
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Context("List", func() {
		It("should return datasources successfully", func() {
			ds1 := newTestDatasource("ds-1", openapi.ASSETKIND_NFS, "Cluster")
			ds2 := newTestDatasource("ds-2", openapi.ASSETKIND_S3, "Project")
			mockResponse := &openapi.DatasourceListResponse{
				Entries: []openapi.DatasourceListResponseEntry{ds1, ds2},
			}
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mockResponse, nil)

			result, err := service.List(ctx, nil, nil, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})

		It("should return error when API call fails", func() {
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("API error"))

			result, err := service.List(ctx, nil, nil, nil)

			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("datasource API call error"))
			Expect(result).To(BeNil())
		})

		It("should handle empty datasource list", func() {
			mockResponse := &openapi.DatasourceListResponse{
				Entries: []openapi.DatasourceListResponseEntry{},
			}
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mockResponse, nil)

			result, err := service.List(ctx, nil, nil, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})

		It("should apply project filter correctly", func() {
			ds1 := newTestDatasource("ds-1", openapi.ASSETKIND_PVC, "Project")
			mockResponse := &openapi.DatasourceListResponse{
				Entries: []openapi.DatasourceListResponseEntry{ds1},
			}
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mockResponse, nil)

			projectID := int32(123)
			result, err := service.List(ctx, &projectID, nil, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})

		It("should apply cluster filter correctly", func() {
			ds1 := newTestDatasource("ds-1", openapi.ASSETKIND_GIT, "Cluster")
			mockResponse := &openapi.DatasourceListResponse{
				Entries: []openapi.DatasourceListResponseEntry{ds1},
			}
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mockResponse, nil)

			clusterID := "cluster-123"
			result, err := service.List(ctx, nil, &clusterID, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})

		It("should apply type filter and return only matching types", func() {
			// Create multiple datasource types, but API returns only PVC when filtered
			_ = newTestDatasource("nfs-ds", openapi.ASSETKIND_NFS, "Cluster")
			_ = newTestDatasource("s3-ds", openapi.ASSETKIND_S3, "Project")
			pvcDs := newTestDatasource("pvc-ds", openapi.ASSETKIND_PVC, "Project")

			// Mock returns only PVC datasource when pvc type filter is applied
			mockResponse := &openapi.DatasourceListResponse{
				Entries: []openapi.DatasourceListResponseEntry{pvcDs},
			}
			dsType := "pvc"
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), &dsType).Return(mockResponse, nil)

			result, err := service.List(ctx, nil, nil, &dsType)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(len(result.GetList())).To(Equal(1))
			// Verify all returned entries are of the filtered type (no NFS or S3)
			for _, entry := range result.GetList() {
				Expect(entry.Type).To(Equal("Pvc"))
			}
		})

		It("should apply nfs type filter correctly", func() {
			// Create multiple datasource types, but API returns only NFS when filtered
			nfsDs := newTestDatasource("nfs-ds", openapi.ASSETKIND_NFS, "Cluster")
			_ = newTestDatasource("pvc-ds", openapi.ASSETKIND_PVC, "Project")
			_ = newTestDatasource("s3-ds", openapi.ASSETKIND_S3, "Project")

			mockResponse := &openapi.DatasourceListResponse{
				Entries: []openapi.DatasourceListResponseEntry{nfsDs},
			}
			dsType := "nfs"
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), &dsType).Return(mockResponse, nil)

			result, err := service.List(ctx, nil, nil, &dsType)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(len(result.GetList())).To(Equal(1))
			// Note: Type is formatted for display (e.g., "NFS" not "nfs")
			for _, entry := range result.GetList() {
				Expect(entry.Type).To(Equal("Nfs"))
			}
		})

		It("should apply s3 type filter correctly", func() {
			// Create multiple datasource types, but API returns only S3 when filtered
			_ = newTestDatasource("nfs-ds", openapi.ASSETKIND_NFS, "Cluster")
			_ = newTestDatasource("pvc-ds", openapi.ASSETKIND_PVC, "Project")
			s3Ds := newTestDatasource("s3-ds", openapi.ASSETKIND_S3, "Project")

			mockResponse := &openapi.DatasourceListResponse{
				Entries: []openapi.DatasourceListResponseEntry{s3Ds},
			}
			dsType := "s3"
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), &dsType).Return(mockResponse, nil)

			result, err := service.List(ctx, nil, nil, &dsType)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(len(result.GetList())).To(Equal(1))
			// Note: Type is formatted for display (e.g., "S3" not "s3")
			for _, entry := range result.GetList() {
				Expect(entry.Type).To(Equal("S3"))
			}
		})

		It("should return empty list when type filter matches no datasources", func() {
			mockResponse := &openapi.DatasourceListResponse{
				Entries: []openapi.DatasourceListResponseEntry{},
			}
			dsType := "git"
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), &dsType).Return(mockResponse, nil)

			result, err := service.List(ctx, nil, nil, &dsType)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.GetList()).To(BeEmpty())
		})

		It("should return all types when no type filter is provided", func() {
			ds1 := newTestDatasource("nfs-ds", openapi.ASSETKIND_NFS, "Cluster")
			ds2 := newTestDatasource("pvc-ds", openapi.ASSETKIND_PVC, "Project")
			ds3 := newTestDatasource("s3-ds", openapi.ASSETKIND_S3, "Project")
			mockResponse := &openapi.DatasourceListResponse{
				Entries: []openapi.DatasourceListResponseEntry{ds1, ds2, ds3},
			}
			// When dsType is nil, all types should be returned
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), nil).Return(mockResponse, nil)

			result, err := service.List(ctx, nil, nil, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(len(result.GetList())).To(Equal(3))
		})
	})
})

// Helper function to create test datasource entries
func newTestDatasource(name string, kind openapi.AssetKind, scope string) openapi.DatasourceListResponseEntry {
	return openapi.DatasourceListResponseEntry{
		Meta: openapi.AssetMeta{
			Id:    "ds-id-" + name,
			Name:  name,
			Kind:  kind,
			Scope: openapi.Scope(scope),
		},
		Spec: openapi.DatasourceListResponseAssetSpec{},
	}
}
