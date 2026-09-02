/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package token

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/auth"
	"golang.org/x/oauth2"

	"github.com/golang/mock/gomock"
	authmocks "github.com/dsx-ai-factory/run-ai/control-plane/runai-cli/internal/autogenerate/mocks/auth"
)

const (
	refreshTokenStr = "refresh"
	accessToken     = "access"
)

var _ = Describe("TokenService", func() {
	var (
		rm            *Service
		mockCtrl      *gomock.Controller
		mockChecker   *authmocks.MockChecker
		mockRefresher *authmocks.MockRefresher
		mockRW        *authmocks.MockReadWriter
		token         *oauth2.Token
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockChecker = authmocks.NewMockChecker(mockCtrl)
		mockRefresher = authmocks.NewMockRefresher(mockCtrl)
		mockRW = authmocks.NewMockReadWriter(mockCtrl)
		token = &oauth2.Token{
			AccessToken:  accessToken,
			RefreshToken: refreshTokenStr,
			Expiry:       time.Now().Add(1 * time.Hour),
		}
		rm = NewTokenService("/tmp/authfile", mockChecker, mockRefresher, mockRW)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("SetToken", func() {
		It("sets the token and writes it", func() {
			mockRW.EXPECT().WriteToken(gomock.Any(), token).Return(nil)
			err := rm.SetToken(token)
			Expect(err).NotTo(HaveOccurred())
			Expect(rm.token).To(Equal(token))
		})
		It("returns error if token is nil", func() {
			err := rm.SetToken(nil)
			Expect(err).To(MatchError(ContainSubstring("token cannot be nil")))
		})
		It("returns error if WriteToken fails", func() {
			mockRW.EXPECT().WriteToken(gomock.Any(), token).Return(errors.New("writefail"))
			err := rm.SetToken(token)
			Expect(err).To(MatchError(ContainSubstring("failed to write token")))
		})
	})

	Describe("GetAccessToken", func() {
		Context("when token is already set and valid", func() {
			BeforeEach(func() {
				rm.token = token
			})
			It("returns the access token", func() {
				access, err := rm.GetAccessToken(context.Background())
				Expect(err).NotTo(HaveOccurred())
				Expect(access).To(Equal(accessToken))
			})
		})
		Context("when token is not set and must be read", func() {
			BeforeEach(func() {
				rm.token = nil
			})
			It("reads the token and returns the access token", func() {
				mockRW.EXPECT().ReadToken(gomock.Any()).Return(token, nil)
				access, err := rm.GetAccessToken(context.Background())
				Expect(err).NotTo(HaveOccurred())
				Expect(access).To(Equal(accessToken))
			})
			It("returns error if reading token fails", func() {
				mockRW.EXPECT().ReadToken(gomock.Any()).Return(nil, errors.New("failread"))
				_, err := rm.GetAccessToken(context.Background())
				Expect(err).To(MatchError(ContainSubstring("failread")))
			})
		})
		Context("when token is expired", func() {
			BeforeEach(func() {
				rm.token = &oauth2.Token{AccessToken: "a", RefreshToken: "r", Expiry: time.Now().Add(-1 * time.Hour)}
			})
			It("returns error for expired token", func() {
				mockChecker.EXPECT().Validate("r").Return(false, nil)
				_, err := rm.GetAccessToken(context.Background())
				Expect(err).To(MatchError(ContainSubstring("expired")))
			})
		})
		Context("when token is not expired but refresh fails", func() {
			BeforeEach(func() {
				rm.token = &oauth2.Token{AccessToken: "a", RefreshToken: "r", Expiry: time.Now().Add(-1 * time.Hour)}
			})
			It("returns error for refresh fail", func() {
				mockChecker.EXPECT().Validate("r").Return(true, nil)
				mockRefresher.EXPECT().RefreshToken(gomock.Any(), "r").Return(nil, errors.New("refreshfail"))
				_, err := rm.GetAccessToken(context.Background())
				Expect(err).To(MatchError(ContainSubstring("refreshfail")))
			})
		})
		Context("when token is not expired and refresh succeeds", func() {
			BeforeEach(func() {
				rm.token = &oauth2.Token{AccessToken: "a", RefreshToken: "r", Expiry: time.Now().Add(-1 * time.Hour)}
			})
			It("refreshes and returns the new access token", func() {
				mockChecker.EXPECT().Validate("r").Return(true, nil)
				newToken := &oauth2.Token{AccessToken: "new", RefreshToken: "r", Expiry: time.Now().Add(1 * time.Hour)}
				mockRefresher.EXPECT().RefreshToken(gomock.Any(), "r").Return(newToken, nil)
				mockRW.EXPECT().WriteToken(gomock.Any(), newToken).Return(nil)
				access, err := rm.GetAccessToken(context.Background())
				Expect(err).NotTo(HaveOccurred())
				Expect(access).To(Equal("new"))
				Expect(rm.token.AccessToken).To(Equal("new"))
			})
		})
		Context("when token is not expired and refresh succeeds but WriteToken fails", func() {
			BeforeEach(func() {
				rm.token = &oauth2.Token{AccessToken: "a", RefreshToken: "r", Expiry: time.Now().Add(-1 * time.Hour)}
			})
			It("returns error if WriteToken fails after refresh", func() {
				mockChecker.EXPECT().Validate("r").Return(true, nil)
				newToken := &oauth2.Token{AccessToken: "new", RefreshToken: "r", Expiry: time.Now().Add(1 * time.Hour)}
				mockRefresher.EXPECT().RefreshToken(gomock.Any(), "r").Return(newToken, nil)
				mockRW.EXPECT().WriteToken(gomock.Any(), newToken).Return(errors.New("writefail"))
				_, err := rm.GetAccessToken(context.Background())
				Expect(err).To(MatchError(ContainSubstring("failed to write token")))
			})
		})
		Context("when Validate returns an error", func() {
			BeforeEach(func() {
				rm.token = &oauth2.Token{AccessToken: "a", RefreshToken: "r", Expiry: time.Now().Add(-1 * time.Hour)}
			})
			It("returns the error from Validate", func() {
				mockChecker.EXPECT().Validate("r").Return(false, errors.New("expirederror"))
				_, err := rm.GetAccessToken(context.Background())
				Expect(err).To(MatchError(ContainSubstring("expirederror")))
			})
		})
	})

	Describe("NewSimpleTokenServiceFromConfig", func() {
		var (
			mockGetter *authmocks.MockConfigGetter
		)
		BeforeEach(func() {
			mockCtrl = gomock.NewController(GinkgoT())
			mockGetter = authmocks.NewMockConfigGetter(mockCtrl)
		})
		AfterEach(func() {
			mockCtrl.Finish()
		})
		It("returns error if authFile or baseUrl is empty", func() {
			mockGetter.EXPECT().GetAuthenticationPath().Return("")
			mockGetter.EXPECT().GetConfigValue(gomock.Any()).Return("")
			mgr, err := NewSimpleTokenServiceFromConfig(mockGetter)
			Expect(err).To(HaveOccurred())
			Expect(mgr).To(BeNil())
		})
		It("returns manager if both values are set and login client succeeds", func() {
			mockGetter.EXPECT().GetAuthenticationPath().Return("/tmp/authfile")
			mockGetter.EXPECT().GetConfigValue(gomock.Any()).Return("http://baseurl")
			mgr, err := NewSimpleTokenServiceFromConfig(mockGetter)
			Expect(err).NotTo(HaveOccurred())
			Expect(mgr).NotTo(BeNil())
		})
	})

	Describe("NewSimpleTokenService", func() {
		It("returns manager if all dependencies succeed", func() {
			// Use a valid URL, but we don't care about actual network calls
			mgr, err := NewSimpleTokenService("/tmp/authfile", "http://localhost")
			Expect(err).NotTo(HaveOccurred())
			Expect(mgr).NotTo(BeNil())
		})
	})
	Describe("TokenService with real token", func() {
		var (
			ctrl         *gomock.Controller
			mockSaver    *authmocks.MockReadWriter
			manager      *Service
			validToken   *oauth2.Token
			tempFile     *os.File
			tokenstr     string
			refreshToken string
		)

		BeforeEach(func() {
			ctrl = gomock.NewController(GinkgoT())
			mockRefresher = authmocks.NewMockRefresher(ctrl)
			mockSaver = authmocks.NewMockReadWriter(ctrl)

			claims := &auth.CustomClaims{
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				},
			}
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			var err error
			tokenstr, err = token.SignedString([]byte("secret"))
			Expect(err).NotTo(HaveOccurred())
			refreshToken = tokenstr
			validToken = &oauth2.Token{
				AccessToken:  tokenstr,
				RefreshToken: refreshToken,
			}
			tempFile, _ = os.CreateTemp("", "tokenrefresher_test")

		})

		AfterEach(func() {
			ctrl.Finish()
			_ = os.Remove(tempFile.Name())
		})

		It("returns access token if token is valid", func() {
			manager = NewTokenService(tempFile.Name(), NewTokenValidator(0), mockRefresher, mockSaver)
			validToken := &oauth2.Token{AccessToken: "access_token", RefreshToken: refreshToken, Expiry: time.Now().Add(time.Hour)}
			mockSaver.EXPECT().ReadToken(tempFile.Name()).Return(validToken, nil)
			token, err := manager.GetAccessToken(nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(token).To(Equal("access_token"))
		})

		It("returns error if token is invalid and cannot be refreshed", func() {
			manager = NewTokenService(tempFile.Name(), NewTokenValidator(0), mockRefresher, mockSaver)
			mockSaver.EXPECT().ReadToken(tempFile.Name()).Return(nil, os.ErrNotExist)
			_, err := manager.GetAccessToken(nil)
			Expect(err).To(HaveOccurred())
		})

		It("refreshes token if expired and saves it", func() {
			expiredToken := &oauth2.Token{AccessToken: "expired", RefreshToken: refreshToken, Expiry: time.Now().Add(-time.Hour)}
			manager = NewTokenService(tempFile.Name(), NewTokenValidator(0), mockRefresher, mockSaver)
			mockSaver.EXPECT().ReadToken(tempFile.Name()).Return(expiredToken, nil)
			mockRefresher.EXPECT().RefreshToken(gomock.Any(), refreshToken).Return(validToken, nil)
			mockSaver.EXPECT().WriteToken(tempFile.Name(), validToken).Return(nil)
			token, err := manager.GetAccessToken(nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(token).To(Equal(tokenstr))
		})

		It("returns error if refresh fails", func() {
			expiredToken := &oauth2.Token{AccessToken: "expired", RefreshToken: refreshToken, Expiry: time.Now().Add(-time.Hour)}
			manager = NewTokenService(tempFile.Name(), NewTokenValidator(0), mockRefresher, mockSaver)
			mockSaver.EXPECT().ReadToken(tempFile.Name()).Return(expiredToken, nil)
			mockRefresher.EXPECT().RefreshToken(gomock.Any(), refreshToken).Return(nil, os.ErrPermission)
			_, err := manager.GetAccessToken(nil)
			Expect(err).To(HaveOccurred())
		})

		It("returns error if saving refreshed token fails", func() {
			expiredToken := &oauth2.Token{AccessToken: "expired", RefreshToken: refreshToken, Expiry: time.Now().Add(-time.Hour)}
			manager = NewTokenService(tempFile.Name(), NewTokenValidator(0), mockRefresher, mockSaver)
			mockSaver.EXPECT().ReadToken(tempFile.Name()).Return(expiredToken, nil)
			mockRefresher.EXPECT().RefreshToken(gomock.Any(), refreshToken).Return(validToken, nil)
			mockSaver.EXPECT().WriteToken(tempFile.Name(), validToken).Return(os.ErrPermission)
			token, err := manager.GetAccessToken(nil)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("permission denied"))
			Expect(token).To(Equal(""))
		})
	})
})
