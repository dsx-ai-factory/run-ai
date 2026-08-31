/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"encoding/json"
	"fmt"
	"net/http"

	. "github.com/onsi/gomega" //nolint:staticcheck

	"github.com/onsi/gomega/ghttp"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/config"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/generated/swagger-runai-backend-client"
	"go.uber.org/zap"
)

const tokenResponse = "token"

func CreateBasicRunaiBackend() *ghttp.Server {
	server := ghttp.NewServer()
	addHandlerToMockAndValidateLivenessRequest(server, tokenResponse)
	addHandlerToMockClusterInfoSyncRequest(server)
	return server
}

func addHandlerToMockAndValidateLivenessRequest(server *ghttp.Server, _ string) {
	livenessPath := fmt.Sprintf("/v1/k8s/clusters/%s/liveness", ClusterIdForTests)
	server.RouteToHandler("POST", livenessPath, func(writer http.ResponseWriter, request *http.Request) {
		zap.S().Infof("Got Liveness Request in mock")
		ghttp.VerifyContentType("application/json")(writer, request)
		ghttp.RespondWith(http.StatusOK, "ok")(writer, request)
	})
}
func addHandlerToMockClusterInfoSyncRequest(server *ghttp.Server) {
	clusterInfoPath := fmt.Sprintf("/v1/k8s/clusters/%s/cluster-info", ClusterIdForTests)
	server.RouteToHandler("POST", clusterInfoPath, func(writer http.ResponseWriter, request *http.Request) {
		ghttp.VerifyContentType("application/json")(writer, request)

		var clusterInfo swagger.ClusterInfo
		err := json.NewDecoder(request.Body).Decode(&clusterInfo)
		Expect(err).To(BeNil())
		Expect(clusterInfo.Domain).To(Equal(config.Config().ClusterDomain))

		ghttp.RespondWith(http.StatusCreated, "ok")(writer, request)
	})
}
