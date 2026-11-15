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

var _ = Describe("Workload Client slim list", func() {
	var (
		server        *httptest.Server
		client        *BackendWorkloadServiceClient
		slimNotFound  bool
		capturedPaths []string
		capturedQuery url.Values
	)

	BeforeEach(func() {
		slimNotFound = false
		capturedPaths = nil
		server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedPaths = append(capturedPaths, r.URL.Path)
			capturedQuery = r.URL.Query()
			w.Header().Set("Content-Type", "application/json")

			switch r.URL.Path {
			case "/api/v1/workloads/slim":
				if slimNotFound {
					w.WriteHeader(http.StatusNotFound)
					return
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"workloads": [
						{
							"id": "slim-1",
							"name": "slim-workload-1",
							"type": "Training",
							"projectName": "test-project",
							"clusterId": "cluster-1"
						}
					]
				}`))
			case "/api/v1/workloads":
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"workloads": [
						{
							"id": "full-1",
							"name": "full-workload-1",
							"type": "Distributed",
							"projectId": "1",
							"projectName": "test-project",
							"clusterId": "cluster-1",
							"externalWorkloadName": "ew-distributed-full-1",
							"distributedFramework": "PyTorch",
							"phase": "Running",
							"source": "CLI"
						}
					]
				}`))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}))

		var err error
		client, err = NewWorkloadServiceClient(server.URL, "test-token")
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		server.Close()
	})

	It("should list from the slim endpoint and parse the slim workloads", func() {
		response, err := client.ListSlim(context.Background(), 10, 0, []string{"name==slim-workload-1"})

		Expect(err).ToNot(HaveOccurred())
		Expect(capturedPaths).To(Equal([]string{"/api/v1/workloads/slim"}))
		Expect(capturedQuery.Get("filterBy")).To(Equal("name==slim-workload-1"))
		Expect(response.Workloads).To(HaveLen(1))
		Expect(response.Workloads[0].Id).To(Equal("slim-1"))
		Expect(response.Workloads[0].Name).To(Equal("slim-workload-1"))
		Expect(response.Workloads[0].GetType()).To(Equal("Training"))
		Expect(response.Workloads[0].ProjectName).To(Equal("test-project"))
		Expect(response.Workloads[0].ClusterId).To(Equal("cluster-1"))
		Expect(response.Workloads[0].DistributedFramework).To(BeNil())
	})

	It("should fall back to the full list when the control plane has no slim endpoint", func() {
		slimNotFound = true

		response, err := client.ListSlim(context.Background(), 10, 0, []string{"name==full-workload-1"})

		Expect(err).ToNot(HaveOccurred())
		Expect(capturedPaths).To(Equal([]string{"/api/v1/workloads/slim", "/api/v1/workloads"}))
		Expect(capturedQuery.Get("filterBy")).To(Equal("name==full-workload-1"))
		Expect(capturedQuery.Get("deleted")).To(Equal("false"))
		Expect(response.Workloads).To(HaveLen(1))
		Expect(response.Workloads[0].Id).To(Equal("full-1"))
		Expect(response.Workloads[0].Name).To(Equal("full-workload-1"))
		Expect(response.Workloads[0].GetType()).To(Equal("Distributed"))
		Expect(response.Workloads[0].ProjectName).To(Equal("test-project"))
		Expect(response.Workloads[0].ClusterId).To(Equal("cluster-1"))
		Expect(response.Workloads[0].GetDistributedFramework()).To(Equal("PyTorch"))
	})
})
