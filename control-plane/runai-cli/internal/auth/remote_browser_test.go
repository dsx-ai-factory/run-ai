/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth_test

import (
	"context"
	"errors"
	"net/url"
	"strings"

	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/consts"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RemoteBrowserHandler", func() {
	var (
		handler *auth.RemoteBrowserHandler
		authUrl *url.URL
		ctx     context.Context
	)

	BeforeEach(func() {
		authUrl, _ = url.Parse("https://example.com/auth")
		ctx = context.Background()
	})

	Describe("Handle", func() {
		Context("when user enters a valid code", func() {
			It("should return the code without an error", func() {
				// Create a reader with the simulated user input
				reader := strings.NewReader("123456\n")
				handler = auth.NewRemoteBrowserHandler(reader)

				resultCode, err := handler.Handle(ctx, authUrl)

				Expect(err).NotTo(HaveOccurred())
				Expect(resultCode).To(Equal("123456"))
			})

			It("Windows case - should return the code without an error", func() {
				// Create a reader with the simulated user input
				reader := strings.NewReader("123456\r\n")
				handler = auth.NewRemoteBrowserHandler(reader)

				resultCode, err := handler.Handle(ctx, authUrl)

				Expect(err).NotTo(HaveOccurred())
				Expect(resultCode).To(Equal("123456"))
			})
		})

		Context("when user enters an empty code", func() {
			It("should return an error", func() {
				// Create a reader with an empty code
				reader := strings.NewReader("\n")
				handler = auth.NewRemoteBrowserHandler(reader)

				_, err := handler.Handle(ctx, authUrl)

				Expect(err).To(MatchError(consts.EmptyCodeError))
			})
		})

		Context("when an error occurs while reading user input", func() {
			It("should return an error", func() {
				// Create a reader that returns an error
				reader := &ErrorReader{}
				handler = auth.NewRemoteBrowserHandler(reader)

				_, err := handler.Handle(ctx, authUrl)

				Expect(err).To(HaveOccurred())

			})
		})
		Context("when the user enters a empty code", func() {
			It("should return an error", func() {
				reader := strings.NewReader("\n")
				handler = auth.NewRemoteBrowserHandler(reader)

				_, err := handler.Handle(ctx, authUrl)

				Expect(err).To(MatchError(consts.EmptyCodeError))
			})
		})

	})
})

type ErrorReader struct{}

func (r *ErrorReader) Read(_ []byte) (n int, err error) {
	return 0, errors.New("read error\n")
}
