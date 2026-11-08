// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package gpuinventory

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/catalog-service/internal/generated/pkg/api"
	clusterclient "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestGPUInventory(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "GPU inventory suite")
}

var _ = Describe("cluster inventory boundary", func() {
	DescribeTable("distinguishes missing, empty and unavailable clusters",
		func(clusterStatus, inventoryStatus, expectedStatus int) {
			clusterCalls, inventoryCalls := 0, 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer GinkgoRecover()
				w.Header().Set("Content-Type", "application/json")
				status := inventoryStatus
				if r.URL.Path == "/api/v1/clusters/"+clusterID+"/minimal" {
					clusterCalls++
					status = clusterStatus
				} else {
					Expect(r.URL.Path).To(Equal("/api/v1/node-pools/minimal"))
					Expect(r.URL.Query().Get("filterBy")).To(Equal("clusterId==" + clusterID))
					inventoryCalls++
				}
				w.WriteHeader(status)
				body := map[string]interface{}{"nodepools": []interface{}{}}
				if status != http.StatusOK {
					body = map[string]interface{}{"code": status, "message": http.StatusText(status)}
				}
				Expect(json.NewEncoder(w).Encode(body)).To(Succeed())
			}))
			DeferCleanup(server.Close)
			hardware, err := New(clusterclient.NewClusterClient(server.URL), time.Second).Get(context.Background(), clusterID)
			if expectedStatus == http.StatusOK {
				Expect(err).NotTo(HaveOccurred())
				Expect(hardware).To(BeEmpty())
			} else {
				Expect(err).To(HaveOccurred())
				var apiErr *api.Error
				Expect(errors.As(err, &apiErr)).To(BeTrue(), "inventory must return the catalog structured error")
				Expect(apiErr.Code).To(Equal(int32(expectedStatus)))
				Expect(hardware).To(BeNil())
			}
			Expect(clusterCalls).To(BeNumerically("==", 1))
			if clusterStatus != http.StatusOK {
				Expect(inventoryCalls).To(BeZero(), "failed cluster lookup must stop inventory retrieval")
			}
		},
		Entry("missing cluster", http.StatusNotFound, http.StatusOK, http.StatusNotFound),
		Entry("cluster access denied", http.StatusForbidden, http.StatusOK, http.StatusForbidden),
		Entry("existing cluster without GPUs", http.StatusOK, http.StatusOK, http.StatusOK),
		Entry("inventory authentication failed", http.StatusOK, http.StatusUnauthorized, http.StatusUnauthorized),
		Entry("inventory access denied", http.StatusOK, http.StatusForbidden, http.StatusForbidden),
		Entry("cluster disappears during lookup", http.StatusOK, http.StatusNotFound, http.StatusNotFound),
		Entry("cluster dependency fails", http.StatusInternalServerError, http.StatusOK, http.StatusServiceUnavailable),
		Entry("inventory dependency fails", http.StatusOK, http.StatusInternalServerError, http.StatusServiceUnavailable),
	)
})
