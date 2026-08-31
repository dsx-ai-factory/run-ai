/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth_test

import (
	"errors"
	"net/url"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"golang.org/x/oauth2"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
	authmocks "github.com/run-ai/runai/control-plane/runai-cli/internal/autogenerate/mocks/auth"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/transaction"
)

var _ = Describe("Login", func() {
	var (
		mockCtrl          *gomock.Controller
		mockAuthenticator *authmocks.MockAuthenticator
		txn               *transaction.Txn
		tokenFilePath     string
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockAuthenticator = authmocks.NewMockAuthenticator(mockCtrl)
		txn = transaction.NewTxn()
		// create temp directory
		fileObj, err := os.CreateTemp("", "token")
		Expect(err).NotTo(HaveOccurred())
		tokenFilePath = fileObj.Name()
		fileObj.Close()
	})

	AfterEach(func() {
		// remove temp file
		err := os.Remove(tokenFilePath)
		Expect(err).NotTo(HaveOccurred())
		mockCtrl.Finish()
	})

	Context("when authentication is successful", func() {
		It("should return a success message and save the refresh token", func() {
			claims := &auth.CustomClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				},
			}
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			validToken, err := token.SignedString([]byte("secret"))
			Expect(err).NotTo(HaveOccurred())

			jsonToken := &oauth2.Token{
				AccessToken:  validToken,
				RefreshToken: validToken,
			}

			mockAuthenticator.EXPECT().Authenticate(txn.GetCtx()).Return(jsonToken, nil)
			err = auth.Login(txn, "auth_timer", mockAuthenticator, tokenFilePath)

			Expect(err).NotTo(HaveOccurred())
			reader := auth.NewFileBased64TokenReadWriter()
			readToken, err := reader.ReadToken(tokenFilePath)
			Expect(err).NotTo(HaveOccurred())
			Expect(readToken.AccessToken).To(Equal(jsonToken.AccessToken))
			Expect(readToken.RefreshToken).To(Equal(jsonToken.RefreshToken))
		})
	})

	Context("when authentication fails", func() {
		It("should return an error", func() {
			expectedErr := errors.New("authentication failed")
			mockAuthenticator.EXPECT().Authenticate(txn.GetCtx()).Return(nil, expectedErr)

			err := auth.Login(txn, "auth_timer", mockAuthenticator, tokenFilePath)

			Expect(err).To(Equal(expectedErr))
		})
	})

	Context("when writing the token fails", func() {
		It("should return an error", func() {
			expectedToken := &oauth2.Token{AccessToken: "access_token"}
			mockAuthenticator.EXPECT().Authenticate(txn.GetCtx()).Return(expectedToken, nil)
			err := auth.Login(txn, "auth_timer", mockAuthenticator, "/nonExistsPath")
			Expect(err).To(HaveOccurred())
		})
	})
	Describe("NewLoginClient", func() {
		var (
			baseUrl *url.URL
		)

		BeforeEach(func() {
			base, err := url.Parse("https://example.com")
			Expect(err).NotTo(HaveOccurred())
			baseUrl = base

		})

		Context("when creating the login client config succeeds", func() {
			It("should return a new API client", func() {

				client, err := auth.NewLoginClient(baseUrl.String())
				Expect(err).NotTo(HaveOccurred())
				Expect(client).NotTo(BeNil())

				Expect(client.GetConfig().Scheme).To(Equal(baseUrl.Scheme))
				Expect(client.GetConfig().Host).To(Equal(baseUrl.Host))
			})
		})

	})
})
