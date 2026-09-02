/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package token

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	api "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth"
	authmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/auth"
)

var _ = Describe("RefresherClient", func() {
	var (
		ctrl      *gomock.Controller
		mockApi   *authmocks.MockTokenApi
		refresher *RefresherClient
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(GinkgoT())
		mockApi = authmocks.NewMockTokenApi(ctrl)
		refresher = NewTokenRefresher(mockApi)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("RefreshToken", func() {
		It("returns a token with correct access, refresh, and expiry on success", func() {
			refreshToken := "refresh_token_value"
			expiresAt := int64(1893456000) // 2030-01-01T00:00:00Z

			claims := &auth.CustomClaims{}
			claims.ExpiresAt = jwt.NewNumericDate(time.Unix(expiresAt, 0))

			accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("secret"))
			Expect(err).NotTo(HaveOccurred())

			apiResp := &api.TokenResponse{}
			apiResp.SetAccessToken(accessToken)
			apiResp.SetRefreshToken(refreshToken)

			mockApi.EXPECT().GrantToken(gomock.Any()).Return(api.ApiGrantTokenRequest{ApiService: mockApi})
			mockApi.EXPECT().GrantTokenExecute(gomock.Any()).Return(apiResp, nil, nil)

			tok, err := refresher.RefreshToken(context.Background(), refreshToken)
			Expect(err).NotTo(HaveOccurred())
			Expect(tok.AccessToken).To(Equal(accessToken))
			Expect(tok.RefreshToken).To(Equal(refreshToken))
			Expect(tok.Expiry.Unix()).To(Equal(expiresAt))
		})

		It("returns an error if the API call fails", func() {
			refreshToken := "refresh_token_value"
			mockApi.EXPECT().GrantToken(gomock.Any()).Return(api.ApiGrantTokenRequest{ApiService: mockApi})
			mockApi.EXPECT().GrantTokenExecute(gomock.Any()).Return(nil, nil, errors.New("api error"))

			tok, err := refresher.RefreshToken(context.Background(), refreshToken)
			Expect(err).To(HaveOccurred())
			Expect(tok).To(BeNil())
		})

		It("returns an error if claims parsing fails", func() {
			refreshToken := "refresh_token_value"
			accessToken := "invalid_token"

			apiResp := &api.TokenResponse{}
			apiResp.SetAccessToken(accessToken)
			apiResp.SetRefreshToken(refreshToken)

			mockApi.EXPECT().GrantToken(gomock.Any()).Return(api.ApiGrantTokenRequest{ApiService: mockApi})
			mockApi.EXPECT().GrantTokenExecute(gomock.Any()).Return(apiResp, nil, nil)

			tok, err := refresher.RefreshToken(context.Background(), refreshToken)
			Expect(err).To(HaveOccurred())
			Expect(tok).To(BeNil())
		})
	})
})
