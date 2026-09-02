/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth_test

import (
	"context"
	"encoding/json"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/golang-jwt/jwt/v5"
	"github.com/golang/mock/gomock"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth"
	authmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/auth"
	authv1 "k8s.io/client-go/pkg/apis/clientauthentication/v1"
)

var _ = Describe("Auth token", func() {
	Describe("GetAccessToken", func() {
		var (
			validToken         string
			ctrl               *gomock.Controller
			ctx                = context.Background()
			mockBackendManager *authmocks.MockTokenService
		)

		BeforeEach(func() {
			var err error

			ctrl = gomock.NewController(GinkgoT())
			mockBackendManager = authmocks.NewMockTokenService(ctrl)
			claims := &auth.CustomClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				},
			}
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			validToken, err = token.SignedString([]byte("secret"))
			Expect(err).NotTo(HaveOccurred())
			mockBackendManager.EXPECT().GetAccessToken(ctx).Return(validToken, nil).AnyTimes()
			Expect(err).NotTo(HaveOccurred())
			auth.GetManager().Init(func(config auth.ConfigGetter) (auth.TokenService, error) {
				return mockBackendManager, nil
			}, nil)
		})

		AfterEach(func() {
			ctrl.Finish()
		})

		Context("GetTokenDisplay", func() {
			It("when output is plaintext", func() {
				tokenDisplay, err := auth.GetTokenDisplay(ctx, "plaintext")
				Expect(err).ToNot(HaveOccurred())
				Expect(tokenDisplay).To(Equal(validToken))
			})

			It("when output is kubeconfig", func() {
				tokenDisplay, err := auth.GetTokenDisplay(ctx, "kubeconfig")
				Expect(err).ToNot(HaveOccurred())

				var execCreds authv1.ExecCredential
				Expect(json.Unmarshal([]byte(tokenDisplay), &execCreds)).To(Succeed())

				Expect(execCreds.Kind).To(Equal("ExecCredential"))
				Expect(execCreds.APIVersion).To(Equal("client.authentication.k8s.io/v1"))
				Expect(execCreds.Status.Token).To(Equal(validToken))
			})

			It("when the output format is unknown", func() {
				_, err := auth.GetTokenDisplay(ctx, "bla")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("unknown token display format: bla"))
			})

			It("when GetAccessToken returns an error", func() {
				newCtx := context.TODO()
				mockBackendManager.EXPECT().GetAccessToken(newCtx).Return("", os.ErrNotExist).AnyTimes()
				_, err := auth.GetTokenDisplay(newCtx, "plaintext")
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("file does not exist"))
			})
		})
	})

	Describe("GetClaims", func() {
		It("should parse valid claims", func() {
			claims := &auth.CustomClaims{
				UserID:              "123",
				UserEmail:           "test@example.com",
				Uid:                 "1000",
				Gid:                 "1000",
				SupplementaryGroups: []int{2000, 3000},
			}
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			tokenString, err := token.SignedString([]byte("secret"))
			Expect(err).NotTo(HaveOccurred())

			parsedClaims, err := auth.GetClaims(tokenString)
			Expect(err).NotTo(HaveOccurred())
			Expect(parsedClaims.UserID).To(Equal("123"))
			Expect(parsedClaims.UserEmail).To(Equal("test@example.com"))
			Expect(parsedClaims.Uid).To(Equal("1000"))
			Expect(parsedClaims.Gid).To(Equal("1000"))
			Expect(parsedClaims.SupplementaryGroups).To(Equal([]int{2000, 3000}))
		})

		It("should return error for invalid token", func() {
			_, err := auth.GetClaims("invalid_token")
			Expect(err).To(HaveOccurred())
		})
	})

})
