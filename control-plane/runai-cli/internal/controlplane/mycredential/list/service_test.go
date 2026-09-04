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
	cpmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/controlplane"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/mycredential/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/controlplane/mycredential/list"
)

var _ = Describe("My Credential List Service", func() {
	var (
		mockCtrl   *gomock.Controller
		mockClient *cpmocks.MockMyCredentialLister
		service    *list.Service
		ctx        context.Context
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockClient = cpmocks.NewMockMyCredentialLister(mockCtrl)
		service = list.NewService(mockClient)
		ctx = context.Background()
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	It("should return output with type column and data in API order", func() {
		entries := []client.MyCredentialEntry{
			{Type: "genericSecret", Name: "beta-cred", Description: "second"},
			{Type: "genericSecret", Name: "alpha-cred", Description: "first"},
		}
		mockClient.EXPECT().ListUserCredentials(gomock.Any(), "genericSecret").Return(entries, nil)

		result, err := service.List(ctx, "genericSecret")

		Expect(err).ToNot(HaveOccurred())
		Expect(result).ToNot(BeNil())
		data := result.Data()
		Expect(data).To(HaveLen(2))
		Expect(data[0]).To(Equal([]string{"genericSecret", "beta-cred", "second"}))
		Expect(data[1]).To(Equal([]string{"genericSecret", "alpha-cred", "first"}))
	})

	It("should call API with empty type filter when listing all", func() {
		entries := []client.MyCredentialEntry{{Type: "dockerRegistry", Name: "cred", Description: "d"}}
		mockClient.EXPECT().ListUserCredentials(gomock.Any(), "").Return(entries, nil)

		result, err := service.List(ctx, "")

		Expect(err).ToNot(HaveOccurred())
		Expect(result).ToNot(BeNil())
		Expect(result.Data()).To(HaveLen(1))
	})

	It("should wrap API errors preserving the original error", func() {
		originalErr := errors.New("connection refused")
		mockClient.EXPECT().ListUserCredentials(gomock.Any(), "").Return(nil, originalErr)

		result, err := service.List(ctx, "")

		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("my-credential API call error"))
		Expect(errors.Is(err, originalErr)).To(BeTrue())
		Expect(result).To(BeNil())
	})

	It("should return empty output for empty list", func() {
		mockClient.EXPECT().ListUserCredentials(gomock.Any(), "").Return([]client.MyCredentialEntry{}, nil)

		result, err := service.List(ctx, "")

		Expect(err).ToNot(HaveOccurred())
		Expect(result).ToNot(BeNil())
		Expect(result.Data()).To(BeEmpty())
	})
})
