/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	api "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/client/generated"
	authmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/auth"
)

var _ = Describe("ControlPlaneTokenExchanger", func() {
	var (
		mockCtrl       *gomock.Controller
		mockTokenApi   *authmocks.MockTokenApi
		tokenExchanger *ControlPlaneTokenExchanger
		redirectUri    string
		authURL        *url.URL
		code           string
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockTokenApi = authmocks.NewMockTokenApi(mockCtrl)
		redirectUri = "https://example.com/callback"
		authURL, _ = url.Parse("https://example.com/auth")
		authURL.Query().Set(redirectQueryKey, redirectUri)
		code = "abc123"

		tokenExchanger = NewControlPlaneTokenExchanger(mockTokenApi, "")
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("Exchange", func() {

		Context("successful exchange", func() {
			var (
				validToken string
			)
			BeforeEach(func() {
				var err error
				expiresAt := time.Now().Add(-1 * time.Hour)
				claims := &CustomClaims{}
				claims.ExpiresAt = jwt.NewNumericDate(expiresAt)
				token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte("secret"))
				Expect(err).NotTo(HaveOccurred())
				validToken = token
			})
			It("should return the access token and refresh token", func() {
				expectedToken := &api.TokenResponse{
					AccessToken:  api.PtrString(validToken),
					RefreshToken: api.PtrString("refresh_token"),
				}

				mockTokenApi.EXPECT().
					GrantToken(gomock.Any()).
					Return(api.ApiGrantTokenRequest{
						ApiService: mockTokenApi,
					})

				mockTokenApi.EXPECT().
					GrantTokenExecute(gomock.Any()).
					Return(expectedToken, nil, nil)

				token, err := tokenExchanger.Exchange(context.Background(), code, authURL)

				Expect(err).NotTo(HaveOccurred())
				Expect(token.AccessToken).To(Equal(expectedToken.GetAccessToken()))
				Expect(token.RefreshToken).To(Equal(expectedToken.GetRefreshToken()))
			})
		})

		Context("error during exchange", func() {
			It("should return an error", func() {
				mockTokenApi.EXPECT().
					GrantToken(gomock.Any()).
					Return(api.ApiGrantTokenRequest{
						ApiService: mockTokenApi,
					})

				mockTokenApi.EXPECT().
					GrantTokenExecute(gomock.Any()).
					Return(nil, nil, errors.New("exchange error"))

				token, err := tokenExchanger.Exchange(context.Background(), code, authURL)

				Expect(err).To(HaveOccurred())
				Expect(token).To(BeNil())
				Expect(err.Error()).To(ContainSubstring("failed to exchange code for token"))
			})
		})
	})
})
