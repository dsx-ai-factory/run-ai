/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package list_test

import (
	"context"
	"errors"

	cpmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/credential/list"
	openapi "github.com/dsx-ai-factory/run-ai/runai-clients/assets-service-client/generated"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Credential List Service", func() {
	var (
		mockCtrl   *gomock.Controller
		service    *list.Service
		mockClient *cpmocks.MockCredentialLister
		ctx        context.Context
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockClient = cpmocks.NewMockCredentialLister(mockCtrl)
		service = list.NewService(mockClient)
		ctx = context.Background()
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Context("List", func() {
		It("should return credentials successfully", func() {
			cred1 := newTestCredential("cred-1", openapi.ASSETKIND_GENERIC_SECRET, "Cluster")
			cred2 := newTestCredential("cred-2", openapi.ASSETKIND_DOCKER_REGISTRY, "Project")
			mockResponse := &openapi.CredentialsListResponse{
				Entries: []openapi.CredentialsListResponseEntry{cred1, cred2},
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
			Expect(err.Error()).To(ContainSubstring("credential API call error"))
			Expect(result).To(BeNil())
		})

		It("should handle empty credential list", func() {
			mockResponse := &openapi.CredentialsListResponse{
				Entries: []openapi.CredentialsListResponseEntry{},
			}
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mockResponse, nil)

			result, err := service.List(ctx, nil, nil, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})

		It("should apply project filter correctly", func() {
			cred1 := newTestCredential("cred-1", openapi.ASSETKIND_GENERIC_SECRET, "Project")
			mockResponse := &openapi.CredentialsListResponse{
				Entries: []openapi.CredentialsListResponseEntry{cred1},
			}
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mockResponse, nil)

			projectID := int32(123)
			result, err := service.List(ctx, &projectID, nil, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})

		It("should apply cluster filter correctly", func() {
			cred1 := newTestCredential("cred-1", openapi.ASSETKIND_DOCKER_REGISTRY, "Cluster")
			mockResponse := &openapi.CredentialsListResponse{
				Entries: []openapi.CredentialsListResponseEntry{cred1},
			}
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(mockResponse, nil)

			clusterID := "cluster-123"
			result, err := service.List(ctx, nil, &clusterID, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
		})

		It("should apply genericSecret type filter correctly", func() {
			genericSecretCred := newTestCredential("generic-secret-cred", openapi.ASSETKIND_GENERIC_SECRET, "Project")

			mockResponse := &openapi.CredentialsListResponse{
				Entries: []openapi.CredentialsListResponseEntry{genericSecretCred},
			}
			credType := "genericSecret"
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), &credType).Return(mockResponse, nil)

			result, err := service.List(ctx, nil, nil, &credType)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(len(result.GetList())).To(Equal(1))
			for _, entry := range result.GetList() {
				Expect(entry.Type).To(Equal("Generic secret"))
			}
		})

		It("should apply dockerRegistry type filter correctly", func() {
			dockerRegistryCred := newTestCredential("docker-registry-cred", openapi.ASSETKIND_DOCKER_REGISTRY, "Project")

			mockResponse := &openapi.CredentialsListResponse{
				Entries: []openapi.CredentialsListResponseEntry{dockerRegistryCred},
			}
			credType := "dockerRegistry"
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), &credType).Return(mockResponse, nil)

			result, err := service.List(ctx, nil, nil, &credType)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(len(result.GetList())).To(Equal(1))
			for _, entry := range result.GetList() {
				Expect(entry.Type).To(Equal("Docker registry"))
			}
		})

		It("should apply ngcApiKey type filter correctly", func() {
			ngcApiKeyCred := newTestCredential("ngc-api-key-cred", openapi.ASSETKIND_NGC_API_KEY, "Project")

			mockResponse := &openapi.CredentialsListResponse{
				Entries: []openapi.CredentialsListResponseEntry{ngcApiKeyCred},
			}
			credType := "ngcApiKey"
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), &credType).Return(mockResponse, nil)

			result, err := service.List(ctx, nil, nil, &credType)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(len(result.GetList())).To(Equal(1))
			for _, entry := range result.GetList() {
				Expect(entry.Type).To(Equal("NGC API key"))
			}
		})

		It("should apply password type filter correctly", func() {
			passwordCred := newTestCredential("password-cred", openapi.ASSETKIND_PASSWORD, "Project")

			mockResponse := &openapi.CredentialsListResponse{
				Entries: []openapi.CredentialsListResponseEntry{passwordCred},
			}
			credType := "password"
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), &credType).Return(mockResponse, nil)

			result, err := service.List(ctx, nil, nil, &credType)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(len(result.GetList())).To(Equal(1))
			for _, entry := range result.GetList() {
				Expect(entry.Type).To(Equal("Password"))
			}
		})

		It("should apply accessKey type filter correctly", func() {
			accessKeyCred := newTestCredential("access-key-cred", openapi.ASSETKIND_ACCESS_KEY, "Project")

			mockResponse := &openapi.CredentialsListResponse{
				Entries: []openapi.CredentialsListResponseEntry{accessKeyCred},
			}
			credType := "accessKey"
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), &credType).Return(mockResponse, nil)

			result, err := service.List(ctx, nil, nil, &credType)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(len(result.GetList())).To(Equal(1))
			for _, entry := range result.GetList() {
				Expect(entry.Type).To(Equal("Access key"))
			}
		})

		It("should return empty list when type filter matches no credentials", func() {
			mockResponse := &openapi.CredentialsListResponse{
				Entries: []openapi.CredentialsListResponseEntry{},
			}
			credType := "ngcApiKey"
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), &credType).Return(mockResponse, nil)

			result, err := service.List(ctx, nil, nil, &credType)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(result.GetList()).To(BeEmpty())
		})

		It("should return all types when no type filter is provided", func() {
			cred1 := newTestCredential("generic-secret", openapi.ASSETKIND_GENERIC_SECRET, "Cluster")
			cred2 := newTestCredential("docker-registry", openapi.ASSETKIND_DOCKER_REGISTRY, "Project")
			cred3 := newTestCredential("ngc-api-key", openapi.ASSETKIND_NGC_API_KEY, "Department")
			mockResponse := &openapi.CredentialsListResponse{
				Entries: []openapi.CredentialsListResponseEntry{cred1, cred2, cred3},
			}
			mockClient.EXPECT().List(ctx, gomock.Any(), gomock.Any(), gomock.Any(), nil).Return(mockResponse, nil)

			result, err := service.List(ctx, nil, nil, nil)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).ToNot(BeNil())
			Expect(len(result.GetList())).To(Equal(3))
		})
	})
})

func newTestCredential(name string, kind openapi.AssetKind, scope string) openapi.CredentialsListResponseEntry {
	return openapi.CredentialsListResponseEntry{
		Meta: openapi.AssetMeta{
			Id:    "cred-id-" + name,
			Name:  name,
			Kind:  kind,
			Scope: openapi.Scope(scope),
		},
		Spec: openapi.CredentialsListResponseAssetSpec{},
	}
}
