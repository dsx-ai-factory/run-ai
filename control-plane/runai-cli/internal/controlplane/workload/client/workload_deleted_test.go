/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Workload Client with Deleted Flag", func() {
	var (
		server        *httptest.Server
		client        *BackendWorkloadServiceClient
		capturedQuery url.Values
	)

	BeforeEach(func() {
		// Create a test server that captures the query parameters
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedQuery = r.URL.Query()

			// Return a valid response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{
				"workloads": [
					{
						"id": "test-1",
						"name": "test-workload-1",
						"createdAt": "2023-01-01T00:00:00Z"
					}
				],
				"total": 1
			}`))
		}))

		// Create client pointing to test server
		var err error
		client, err = NewWorkloadServiceClient(server.URL, "test-token")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		server.Close()
	})

	Context("List API with deleted parameter", func() {
		It("should include deleted=false in query when deleted is false", func() {
			ctx := context.Background()
			filters := []string{"status==Running"}

			response, err := client.List(ctx, 10, 0, filters, false)

			Expect(err).ToNot(HaveOccurred())
			Expect(response).ToNot(BeNil())

			// Verify query parameters
			Expect(capturedQuery.Get("limit")).To(Equal("10"))
			Expect(capturedQuery.Get("offset")).To(Equal("0"))
			Expect(capturedQuery.Get("deleted")).To(Equal("false"))
			Expect(capturedQuery["filterBy"]).To(Equal(filters))
		})

		It("should include deleted=true in query when deleted is true", func() {
			ctx := context.Background()
			filters := []string{"project==test-project"}

			response, err := client.List(ctx, 20, 10, filters, true)

			Expect(err).ToNot(HaveOccurred())
			Expect(response).ToNot(BeNil())

			// Verify query parameters
			Expect(capturedQuery.Get("limit")).To(Equal("20"))
			Expect(capturedQuery.Get("offset")).To(Equal("10"))
			Expect(capturedQuery.Get("deleted")).To(Equal("true"))
			Expect(capturedQuery["filterBy"]).To(Equal(filters))
		})

		It("should handle empty filters with deleted flag", func() {
			ctx := context.Background()

			response, err := client.List(ctx, 5, 0, []string{}, true)

			Expect(err).ToNot(HaveOccurred())
			Expect(response).ToNot(BeNil())

			// Verify query parameters
			Expect(capturedQuery.Get("deleted")).To(Equal("true"))
			Expect(capturedQuery["filterBy"]).To(BeEmpty())
		})

		It("should handle nil filters with deleted flag", func() {
			ctx := context.Background()

			response, err := client.List(ctx, 5, 0, nil, true)

			Expect(err).ToNot(HaveOccurred())
			Expect(response).ToNot(BeNil())

			// Verify query parameters
			Expect(capturedQuery.Get("deleted")).To(Equal("true"))
			_, hasFilterBy := capturedQuery["filterBy"]
			Expect(hasFilterBy).To(BeFalse())
		})
	})

	Context("Error handling", func() {
		It("should handle server errors correctly when deleted is set", func() {
			// Create a server that returns an error
			errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "Internal server error"}`))
			}))
			defer errorServer.Close()

			errorClient, err := NewWorkloadServiceClient(errorServer.URL, "test-token")
			Expect(err).ToNot(HaveOccurred())

			ctx := context.Background()
			response, err := errorClient.List(ctx, 10, 0, []string{}, true)

			Expect(err).To(HaveOccurred())
			Expect(response).To(BeNil())
		})

		It("should handle network errors correctly when deleted is set", func() {
			// Create a client with invalid URL
			invalidClient, err := NewWorkloadServiceClient("http://invalid-url-that-does-not-exist", "test-token")
			Expect(err).ToNot(HaveOccurred())

			ctx := context.Background()
			response, err := invalidClient.List(ctx, 10, 0, []string{}, true)

			Expect(err).To(HaveOccurred())
			Expect(response).To(BeNil())
		})
	})

	Context("Response parsing", func() {
		It("should correctly parse response when listing deleted workloads", func() {
			// Create a server that returns deleted workloads
			deletedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify this is a request for deleted workloads
				Expect(r.URL.Query().Get("deleted")).To(Equal("true"))

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"workloads": [
						{
							"id": "deleted-1",
							"name": "deleted-workload-1",
							"createdAt": "2023-01-01T00:00:00Z",
							"deletedAt": "2023-01-02T00:00:00Z"
						},
						{
							"id": "deleted-2",
							"name": "deleted-workload-2",
							"createdAt": "2023-01-01T00:00:00Z",
							"deletedAt": "2023-01-03T00:00:00Z"
						}
					],
					"total": 2
				}`))
			}))
			defer deletedServer.Close()

			deletedClient, err := NewWorkloadServiceClient(deletedServer.URL, "test-token")
			Expect(err).ToNot(HaveOccurred())

			ctx := context.Background()
			response, err := deletedClient.List(ctx, 10, 0, []string{}, true)

			Expect(err).ToNot(HaveOccurred())
			Expect(response).ToNot(BeNil())
			Expect(response.Workloads).To(HaveLen(2))
			Expect(response.Workloads[0].Id).To(Equal("deleted-1"))
			Expect(response.Workloads[1].Id).To(Equal("deleted-2"))
		})
	})
})
