/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"encoding/json"
	"net/http/httptest"

	jwt "github.com/golang-jwt/jwt/v5"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-api/internal/api"
)

var _ = DescribeTable("Audience Claim Parsing",
	func(data []byte, expectedAud jwt.ClaimStrings) {
		var claims RunAiClaims
		Expect(json.Unmarshal(data, &claims)).To(Succeed())
		aud, err := claims.GetAudience()
		Expect(err).ToNot(HaveOccurred())
		Expect(aud).To(Equal(expectedAud))
	},
	Entry("aud as a single string", []byte(`{"aud":"runai"}`), jwt.ClaimStrings{"runai"}),
	Entry("aud as an array of strings", []byte(`{"aud":["runai","runai-dummy"]}`), jwt.ClaimStrings{"runai", "runai-dummy"}),
	Entry("missing aud claim", []byte(`{"email":"user@run.ai"}`), jwt.ClaimStrings(nil)),
)

var _ = Describe("Auth Header Extraction", func() {
	Describe("ExtractAuthTokenFromRequest", func() {
		Context("with standard Authorization header", func() {
			It("should extract token from valid Bearer token", func() {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("Authorization", "Bearer valid-token-123")

				token, err := ExtractAuthTokenFromRequest(req)

				Expect(err).ToNot(HaveOccurred())
				Expect(token).To(Equal("valid-token-123"))
			})

			It("should return error for malformed Authorization header without Bearer prefix", func() {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")

				_, err := ExtractAuthTokenFromRequest(req)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid authorization header"))
			})

			It("should return error for empty Authorization header", func() {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("Authorization", "")

				_, err := ExtractAuthTokenFromRequest(req)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid authorization header"))
			})

			It("should return error for Bearer without token", func() {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("Authorization", "Bearer")

				_, err := ExtractAuthTokenFromRequest(req)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid authorization header"))
			})
		})

		Context("with WebSocket authorization header", func() {
			It("should extract token from valid WebSocket protocol header", func() {
				req := httptest.NewRequest("GET", "/ws", nil)
				req.Header.Set("Sec-Websocket-Protocol", "auth.bearer.valid-ws-token-123, Bearer")

				token, err := ExtractAuthTokenFromRequest(req)

				Expect(err).ToNot(HaveOccurred())
				Expect(token).To(Equal("valid-ws-token-123"))
			})

			// this is valid, although JS add the suffix.
			It("should extract token from WebSocket header without Bearer suffix", func() {
				req := httptest.NewRequest("GET", "/ws", nil)
				req.Header.Set("Sec-Websocket-Protocol", "auth.bearer.token-without-suffix")

				token, err := ExtractAuthTokenFromRequest(req)

				Expect(err).ToNot(HaveOccurred())
				Expect(token).To(Equal("token-without-suffix"))
			})

			// this will fail in the token validation, but not the extraction
			It("should extract empty token from WebSocket header with only prefix", func() {
				req := httptest.NewRequest("GET", "/ws", nil)
				req.Header.Set("Sec-Websocket-Protocol", "auth.bearer.")

				_, err := ExtractAuthTokenFromRequest(req)

				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(errTokenIsEmpty))
			})
		})

		Context("with missing headers", func() {
			It("should return error when no authorization headers are present", func() {
				req := httptest.NewRequest("GET", "/test", nil)

				_, err := ExtractAuthTokenFromRequest(req)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("missing authorization header"))
			})
		})

		Context("with both headers present", func() {
			It("should prioritize standard Authorization header over WebSocket header", func() {
				req := httptest.NewRequest("GET", "/test", nil)
				req.Header.Set("Authorization", "Bearer standard-token")
				req.Header.Set("Sec-Websocket-Protocol", "auth.bearer.ws-token, Bearer")

				token, err := ExtractAuthTokenFromRequest(req)

				Expect(err).ToNot(HaveOccurred())
				Expect(token).To(Equal("standard-token"))
			})
		})
	})

	Describe("getAuthorizationHeader", func() {
		It("should extract token from valid Authorization header", func() {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", "Bearer test-token")

			token, err := getAuthorizationHeader(req)

			Expect(err).ToNot(HaveOccurred())
			Expect(*token).To(Equal("test-token"))
		})

		It("should extract token from valid Authorization header", func() {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", "Bearer ")

			_, err := getAuthorizationHeader(req)

			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(errTokenIsEmpty))
		})

		It("should return NotFoundError when Authorization header is missing", func() {
			req := httptest.NewRequest("GET", "/test", nil)

			token, err := getAuthorizationHeader(req)

			Expect(token).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(api.NotFoundError{}))
		})

		It("should return error for invalid Authorization header format", func() {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", "InvalidFormat token")

			token, err := getAuthorizationHeader(req)

			Expect(token).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid authorization header"))
		})

		It("should handle multiple Authorization headers by using the first one", func() {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Add("Authorization", "Bearer first-token")
			req.Header.Add("Authorization", "Bearer second-token")

			token, err := getAuthorizationHeader(req)

			Expect(err).ToNot(HaveOccurred())
			Expect(*token).To(Equal("first-token"))
		})
	})

	Describe("geWebSocketAuthorizationHeader", func() {
		It("should extract token from valid WebSocket protocol header", func() {
			req := httptest.NewRequest("GET", "/ws", nil)
			req.Header.Set("Sec-Websocket-Protocol", "auth.bearer.ws-test-token, Bearer")

			token, err := geWebSocketAuthorizationHeader(req)

			Expect(err).ToNot(HaveOccurred())
			Expect(*token).To(Equal("ws-test-token"))
		})

		It("should return NotFoundError when WebSocket protocol header is missing", func() {
			req := httptest.NewRequest("GET", "/ws", nil)

			token, err := geWebSocketAuthorizationHeader(req)

			Expect(token).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(api.NotFoundError{}))
		})

		It("should handle WebSocket header without Bearer suffix", func() {
			req := httptest.NewRequest("GET", "/ws", nil)
			req.Header.Set("Sec-Websocket-Protocol", "auth.bearer.token-no-suffix")

			token, err := geWebSocketAuthorizationHeader(req)

			Expect(err).ToNot(HaveOccurred())
			Expect(*token).To(Equal("token-no-suffix"))
		})

		It("should handle WebSocket header with only prefix", func() {
			req := httptest.NewRequest("GET", "/ws", nil)
			req.Header.Set("Sec-Websocket-Protocol", "auth.bearer.")

			_, err := geWebSocketAuthorizationHeader(req)

			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(errTokenIsEmpty))
		})

		It("should handle multiple WebSocket protocol headers by using the first one", func() {
			req := httptest.NewRequest("GET", "/ws", nil)
			req.Header.Add("Sec-Websocket-Protocol", "auth.bearer.first-ws-token, Bearer")
			req.Header.Add("Sec-Websocket-Protocol", "auth.bearer.second-ws-token, Bearer")

			token, err := geWebSocketAuthorizationHeader(req)

			Expect(err).ToNot(HaveOccurred())
			Expect(*token).To(Equal("first-ws-token"))
		})
	})

	Describe("getToken", func() {
		It("should return standard auth token when both headers are present", func() {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", "Bearer standard-token")
			req.Header.Set("Sec-Websocket-Protocol", "auth.bearer.ws-token, Bearer")

			token, err := getToken(req)

			Expect(err).ToNot(HaveOccurred())
			Expect(*token).To(Equal("standard-token"))
		})

		It("should fallback to WebSocket token when standard auth header is missing", func() {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Sec-Websocket-Protocol", "auth.bearer.fallback-token, Bearer")

			token, err := getToken(req)

			Expect(err).ToNot(HaveOccurred())
			Expect(*token).To(Equal("fallback-token"))
		})

		It("should return missing auth header error when no headers are present", func() {
			req := httptest.NewRequest("GET", "/test", nil)

			token, err := getToken(req)

			Expect(token).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(errMissingAuthHeader))
		})

		It("should return error when standard auth header is malformed and no WebSocket header", func() {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", "InvalidFormat token")

			token, err := getToken(req)

			Expect(token).To(BeNil())
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid authorization header"))
		})
	})
})
