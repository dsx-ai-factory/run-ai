/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package apiclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RolesHTTPClient", func() {
	var (
		server       *httptest.Server
		client       *RolesHTTPClient
		gotRequest   *http.Request
		statusCode   int
		rolesResBody string
	)

	BeforeEach(func() {
		gotRequest = nil
		statusCode = http.StatusOK
		rolesResBody = `{"roles":[{"id":1,"name":"System administrator"},{"id":7,"name":"AI practitioner"}],"next":2}`

		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotRequest = r
			w.WriteHeader(statusCode)
			_, _ = w.Write([]byte(rolesResBody))
		}))
		DeferCleanup(server.Close)

		client = &RolesHTTPClient{
			httpClient:  server.Client(),
			serverURL:   server.URL,
			accessToken: "test-token",
			tokenExpiry: time.Now().Add(time.Hour),
		}
	})

	Describe("GetRoles", func() {
		It("requests the v2 roles API with the maximum page size", func() {
			_, _, err := client.GetRoles(context.Background())
			Expect(err).ToNot(HaveOccurred())

			Expect(gotRequest).ToNot(BeNil())
			Expect(gotRequest.Method).To(Equal(http.MethodGet))
			Expect(gotRequest.URL.Path).To(Equal("/api/v2/authorization/roles"))
			Expect(gotRequest.URL.Query().Get("limit")).To(Equal("500"))
			Expect(gotRequest.Header.Get("Authorization")).To(Equal("Bearer test-token"))
		})

		It("decodes the roles from the paginated response", func() {
			roles, code, err := client.GetRoles(context.Background())
			Expect(err).ToNot(HaveOccurred())
			Expect(code).To(Equal(http.StatusOK))
			Expect(roles).To(Equal([]Role{
				{ID: 1, Name: "System administrator"},
				{ID: 7, Name: "AI practitioner"},
			}))
		})

		It("returns an error when the server does not respond with 200", func() {
			statusCode = http.StatusForbidden
			rolesResBody = `{"message":"forbidden"}`

			roles, code, err := client.GetRoles(context.Background())
			Expect(err).To(HaveOccurred())
			Expect(code).To(Equal(http.StatusForbidden))
			Expect(roles).To(BeNil())
		})
	})

	Describe("GetRoleIDByName", func() {
		It("finds the role id in the v2 response", func() {
			id, err := client.GetRoleIDByName(context.Background(), "AI practitioner")
			Expect(err).ToNot(HaveOccurred())
			Expect(id).To(Equal(int32(7)))
		})

		It("returns an error when the role is missing", func() {
			_, err := client.GetRoleIDByName(context.Background(), "Missing role")
			Expect(err).To(HaveOccurred())
		})
	})
})
