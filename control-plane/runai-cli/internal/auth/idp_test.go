/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	authmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/auth"
	"golang.org/x/oauth2"
)

var _ = Describe("IDPAuthenticator", func() {
	var (
		mockCtrl           *gomock.Controller
		mockLoginUrlGetter *authmocks.MockLoginUrlGetter
		mockTokenExchanger *authmocks.MockTokenExchanger
		mockCodeHandler    *authmocks.MockAuthorizationCodeHandler
		idpAuthenticator   *IDPAuthenticator
		ctx                context.Context
		authUrl            *url.URL
		authCode           string
		token              *oauth2.Token
		errExpected        error
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockLoginUrlGetter = authmocks.NewMockLoginUrlGetter(mockCtrl)
		mockTokenExchanger = authmocks.NewMockTokenExchanger(mockCtrl)
		mockCodeHandler = authmocks.NewMockAuthorizationCodeHandler(mockCtrl)
		idpAuthenticator = NewIDPAuthenticator(mockLoginUrlGetter, mockTokenExchanger, mockCodeHandler)
		ctx = context.Background()
		authUrl, _ = url.Parse("https://example.com/auth")
		authCode = "abc123"
		token = &oauth2.Token{AccessToken: "access_token", RefreshToken: "refresh_token"}
		errExpected = errors.New("expected error")
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("Authenticate", func() {
		Context("successful authentication", func() {
			It("should return the token with refresh token", func() {
				ctx = context.Background()
				mockLoginUrlGetter.EXPECT().GetAuthUrl(ctx).Return(authUrl, nil)
				mockCodeHandler.EXPECT().Handle(ctx, authUrl).Return(authCode, nil)
				mockTokenExchanger.EXPECT().Exchange(ctx, authCode, gomock.Any()).Return(token, nil)

				result, err := idpAuthenticator.Authenticate(ctx)

				Expect(err).NotTo(HaveOccurred())
				Expect(result.AccessToken).To(Equal(token.AccessToken))
				Expect(result.RefreshToken).To(Equal(token.RefreshToken))
			})
		})

		Context("error getting auth URL", func() {
			It("should return the error", func() {
				mockLoginUrlGetter.EXPECT().GetAuthUrl(ctx).Return(nil, errExpected)

				result, err := idpAuthenticator.Authenticate(ctx)

				Expect(err).To(Equal(errExpected))
				Expect(result).To(BeNil())
			})

			It("should return the error", func() {
				invalidAuthUrl, _ := url.Parse("example.com/auth")
				invalidUrlError := fmt.Errorf("authentication URL is invalid, missing scheme or host: '%s'",
					invalidAuthUrl.String())

				mockLoginUrlGetter.EXPECT().GetAuthUrl(ctx).Return(invalidAuthUrl, nil)

				result, err := idpAuthenticator.Authenticate(ctx)

				Expect(err).To(Equal(invalidUrlError))
				Expect(result).To(BeNil())
			})
		})

		Context("error handling authorization code", func() {
			It("should return the error", func() {
				mockLoginUrlGetter.EXPECT().GetAuthUrl(ctx).Return(authUrl, nil)
				mockCodeHandler.EXPECT().Handle(ctx, authUrl).Return("", errExpected)

				result, err := idpAuthenticator.Authenticate(ctx)

				Expect(err).To(Equal(errExpected))
				Expect(result).To(BeNil())
			})
		})

		Context("error exchanging code for token", func() {
			It("should return the error", func() {
				mockLoginUrlGetter.EXPECT().GetAuthUrl(ctx).Return(authUrl, nil)
				mockCodeHandler.EXPECT().Handle(ctx, authUrl).Return(authCode, nil)
				mockTokenExchanger.EXPECT().Exchange(ctx, authCode, gomock.Any()).Return(nil, errExpected)

				result, err := idpAuthenticator.Authenticate(ctx)

				Expect(err).To(Equal(errExpected))
				Expect(result).To(BeNil())
			})
		})
	})
})
