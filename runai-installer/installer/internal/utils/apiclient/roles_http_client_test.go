/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package apiclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestApiClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ApiClient Suite")
}

var _ = Describe("RolesHTTPClient", func() {
	var (
		server      *httptest.Server
		client      *RolesHTTPClient
		gotRequests []*http.Request
		statusCode  int
		// response body per requested offset
		pages map[string]string
	)

	BeforeEach(func() {
		gotRequests = nil
		statusCode = http.StatusOK
		pages = map[string]string{
			"0": `{"roles":[{"id":1,"name":"System administrator"},{"id":7,"name":"AI practitioner"}]}`,
		}

		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotRequests = append(gotRequests, r)
			w.WriteHeader(statusCode)
			_, _ = w.Write([]byte(pages[r.URL.Query().Get("offset")]))
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

			Expect(gotRequests).To(HaveLen(1))
			Expect(gotRequests[0].Method).To(Equal(http.MethodGet))
			Expect(gotRequests[0].URL.Path).To(Equal("/api/v2/authorization/roles"))
			Expect(gotRequests[0].URL.Query().Get("limit")).To(Equal("500"))
			Expect(gotRequests[0].URL.Query().Get("offset")).To(Equal("0"))
			Expect(gotRequests[0].Header.Get("Authorization")).To(Equal("Bearer test-token"))
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

		It("follows next and returns the roles of every page", func() {
			pages["0"] = `{"roles":[{"id":9,"name":"Custom role"}],"next":500}`
			pages["500"] = `{"roles":[{"id":7,"name":"AI practitioner"}]}`

			roles, code, err := client.GetRoles(context.Background())
			Expect(err).ToNot(HaveOccurred())
			Expect(code).To(Equal(http.StatusOK))
			Expect(roles).To(Equal([]Role{
				{ID: 9, Name: "Custom role"},
				{ID: 7, Name: "AI practitioner"},
			}))
			Expect(gotRequests).To(HaveLen(2))
			Expect(gotRequests[1].URL.Query().Get("offset")).To(Equal("500"))
		})

		It("stops when next does not move forward", func() {
			pages["0"] = `{"roles":[{"id":9,"name":"Custom role"}],"next":0}`

			roles, _, err := client.GetRoles(context.Background())
			Expect(err).ToNot(HaveOccurred())
			Expect(roles).To(HaveLen(1))
			Expect(gotRequests).To(HaveLen(1))
		})

		It("returns an error when the server does not respond with 200", func() {
			statusCode = http.StatusForbidden
			pages["0"] = `{"message":"forbidden"}`

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

		It("finds a role that is only on the second page", func() {
			// Roles are returned newest first, so predefined roles like this one end up on the last page.
			pages["0"] = `{"roles":[{"id":9,"name":"Custom role"}],"next":500}`
			pages["500"] = `{"roles":[{"id":7,"name":"AI practitioner"}]}`

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
