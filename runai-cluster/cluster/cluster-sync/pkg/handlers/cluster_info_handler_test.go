// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	swagger "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("updateClusterInfo mount-check compatibility", func() {
	for _, tc := range []struct {
		name      string
		status    int
		message   string
		capable   bool
		wantCalls int
		wantError bool
	}{
		{name: "current control plane", status: http.StatusNoContent, capable: true, wantCalls: 1},
		{name: "older control plane", status: http.StatusBadRequest, message: `json: unknown field "s3FileSystemMountCheckImage"`, capable: true, wantCalls: 2},
		{name: "unrelated unknown field", status: http.StatusBadRequest, message: `json: unknown field "other"`, capable: true, wantCalls: 1, wantError: true},
		{name: "server failure", status: http.StatusInternalServerError, capable: true, wantCalls: 1, wantError: true},
		{name: "capability already absent", status: http.StatusBadRequest, message: `json: unknown field "s3FileSystemMountCheckImage"`, wantCalls: 1, wantError: true},
	} {
		It(tc.name, func() {
			reports := make(chan swagger.ClusterReportedConfig, 2)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer GinkgoRecover()
				var body swagger.ClusterInfoSyncRequest
				Expect(json.NewDecoder(r.Body).Decode(&body)).To(Succeed())
				reportedStatus := body.GetStatus()
				reports <- reportedStatus.GetConfig()
				if len(reports) > 1 {
					w.WriteHeader(http.StatusNoContent)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if tc.status != http.StatusNoContent {
					Expect(json.NewEncoder(w).Encode(swagger.Error{Code: int32(tc.status), Message: tc.message})).To(Succeed())
				}
			}))
			DeferCleanup(server.Close)
			clientConfig := swagger.NewConfiguration()
			clientConfig.Servers = swagger.ServerConfigurations{{URL: server.URL}}
			clientConfig.HTTPClient = server.Client()
			config := swagger.ClusterReportedConfig{}
			config.SetS3FileSystemImage("example.com/goofys:current")
			if tc.capable {
				config.SetS3FileSystemMountCheckImage(config.GetS3FileSystemImage())
			}
			status := swagger.ClusterReportedStatus{}
			status.SetConfig(config)
			info := swagger.ClusterInfoSyncRequest{Domain: "cluster.example.com", Version: "current"}
			info.SetStatus(status)
			res, err := updateClusterInfo(context.Background(), swagger.NewAPIClient(clientConfig), "cluster-id", info)
			if res != nil {
				Expect(res.Body.Close()).To(Succeed())
			}
			Expect(err != nil).To(Equal(tc.wantError))
			Expect(reports).To(HaveLen(tc.wantCalls))
			for i := 0; i < tc.wantCalls; i++ {
				report := <-reports
				Expect(report.GetS3FileSystemImage()).To(Equal("example.com/goofys:current"))
				Expect(report.HasS3FileSystemMountCheckImage()).To(Equal(tc.capable && i == 0))
			}
		})
	}
})
