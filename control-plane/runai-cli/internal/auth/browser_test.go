/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth_test

import (
	"context"
	"errors"
	"net/url"

	"github.com/pkg/browser"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
	authmocks "github.com/run-ai/runai/control-plane/runai-cli/internal/autogenerate/mocks/auth"
)

var _ = Describe("BrowserAuthExchangeCodeHandler", func() {
	var (
		mockCtrl                     *gomock.Controller
		mockExchangeCodeServerGetter *authmocks.MockExchangeCodeServerGetter
		handler                      *auth.BrowserAuthExchangeCodeHandler
		ctx                          context.Context
		authUrl                      *url.URL
		skiped                       bool
	)

	BeforeEach(func() {
		err := browser.OpenURL("https://example.com")
		if err != nil {
			skiped = true
			Skip("Browser not available, skipping tests")
		}
		mockCtrl = gomock.NewController(GinkgoT())
		mockExchangeCodeServerGetter = authmocks.NewMockExchangeCodeServerGetter(mockCtrl)
		handler = auth.NewBrowserAuthExchangeCodeHandler(mockExchangeCodeServerGetter, false)
		ctx = context.Background()
		authUrl, _ = url.Parse("https://example.com/auth?state=test-state")
	})

	AfterEach(func() {
		if !skiped {
			mockCtrl.Finish()
		}
	})

	Describe("Handle", func() {
		Context("when the exchange code server starts successfully", func() {
			It("should return the exchange code", func() {
				expectedCode := "abc123"
				mockExchangeCodeServerGetter.EXPECT().SetExpectedState("test-state")
				mockExchangeCodeServerGetter.EXPECT().Start(gomock.Any()).Return(nil)
				mockExchangeCodeServerGetter.EXPECT().GetExchangeCode(gomock.Any()).Return(expectedCode, nil)

				code, err := handler.Handle(ctx, authUrl)

				Expect(err).NotTo(HaveOccurred())
				Expect(code).To(Equal(expectedCode))
			})
		})

		Context("when starting the exchange code server fails", func() {
			It("should return an error", func() {
				expectedErr := errors.New("failed to start server")
				mockExchangeCodeServerGetter.EXPECT().SetExpectedState("test-state")
				mockExchangeCodeServerGetter.EXPECT().Start(gomock.Any()).Return(expectedErr)

				code, err := handler.Handle(ctx, authUrl)

				Expect(err).To(Equal(expectedErr))
				Expect(code).To(BeEmpty())
			})
		})

		Context("when getting the exchange code fails", func() {
			It("should return an error", func() {
				expectedErr := errors.New("failed to get exchange code")
				mockExchangeCodeServerGetter.EXPECT().SetExpectedState("test-state")
				mockExchangeCodeServerGetter.EXPECT().Start(gomock.Any()).Return(nil)
				mockExchangeCodeServerGetter.EXPECT().GetExchangeCode(gomock.Any()).Return("", expectedErr)

				code, err := handler.Handle(ctx, authUrl)

				Expect(err).To(Equal(expectedErr))
				Expect(code).To(BeEmpty())
			})
		})
	})
})
