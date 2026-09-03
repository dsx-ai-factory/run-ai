/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package token_test

import (
	"context"
	"fmt"
	"net/http"

	api "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth/token"

	authmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/auth"

	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/api/config"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ControlPlaneTokenRevoker", func() {
	var (
		revoker       *token.ControlPlaneTokenRevoker
		mockCtrl      *gomock.Controller
		mockLogoutAPI *authmocks.MockLogoutAPI
		mockApiClient *api.APIClient
		tokenStr      string
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockLogoutAPI = authmocks.NewMockLogoutAPI(mockCtrl)
		tokenStr = "test-token"

		// Create a mock APIClient with our mock LogoutAPI
		mockApiClient = api.NewAPIClient(api.NewConfiguration())
		mockApiClient.LogoutAPI = mockLogoutAPI
		revoker = token.NewControlPlaneTokenRevoker(mockApiClient)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("Revoke", func() {
		It("should successfully revoke a tokenStr", func() {
			// Create a mock response with OK status
			mockResponse := &http.Response{
				StatusCode: http.StatusOK,
			}

			// Set up expectations
			mockLogoutAPI.EXPECT().
				Logout(gomock.Any()).
				Return(api.ApiLogoutRequest{ApiService: mockLogoutAPI}).
				Times(1)

			mockLogoutAPI.EXPECT().
				LogoutExecute(gomock.Any()).
				Return(mockResponse, nil).
				Times(1)

			err := revoker.Revoke(context.Background(), tokenStr)
			Expect(err).NotTo(HaveOccurred())
			Expect(mockApiClient.GetConfig().DefaultHeader[config.AuthorizationHeader]).To(Equal(fmt.Sprintf(config.BearerTokenStr, tokenStr)))
		})

		It("should return error when logout API call fails", func() {
			// Set up expectations for API error
			mockLogoutAPI.EXPECT().
				Logout(gomock.Any()).
				Return(api.ApiLogoutRequest{ApiService: mockLogoutAPI}).
				Times(1)

			mockLogoutAPI.EXPECT().
				LogoutExecute(gomock.Any()).
				Return(nil, &api.GenericOpenAPIError{}).
				Times(1)

			err := revoker.Revoke(context.Background(), tokenStr)
			Expect(err).To(HaveOccurred())
		})

		It("should return error when response is nil", func() {
			// Set up expectations for nil response
			mockLogoutAPI.EXPECT().
				Logout(gomock.Any()).
				Return(api.ApiLogoutRequest{ApiService: mockLogoutAPI}).
				Times(1)

			mockLogoutAPI.EXPECT().
				LogoutExecute(gomock.Any()).
				Return(nil, nil).
				Times(1)

			err := revoker.Revoke(context.Background(), tokenStr)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(Equal("response is nil"))
		})

		It("should return error when response status is not OK", func() {
			// Create a mock response with non-OK status
			mockResponse := &http.Response{
				StatusCode: http.StatusUnauthorized,
				Status:     "401 Unauthorized",
			}

			// Set up expectations
			mockLogoutAPI.EXPECT().
				Logout(gomock.Any()).
				Return(api.ApiLogoutRequest{ApiService: mockLogoutAPI}).
				Times(1)

			mockLogoutAPI.EXPECT().
				LogoutExecute(gomock.Any()).
				Return(mockResponse, nil).
				Times(1)

			err := revoker.Revoke(context.Background(), tokenStr)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to revoke token"))
		})

		It("should return error when response status is 404 Not Found", func() {
			mockResponse := &http.Response{
				StatusCode: http.StatusNotFound,
				Status:     "404 Not Found",
			}

			mockLogoutAPI.EXPECT().
				Logout(gomock.Any()).
				Return(api.ApiLogoutRequest{ApiService: mockLogoutAPI}).
				Times(1)

			mockLogoutAPI.EXPECT().
				LogoutExecute(gomock.Any()).
				Return(mockResponse, nil).
				Times(1)

			err := revoker.Revoke(context.Background(), tokenStr)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("your session has expired"))
		})
	})
})
