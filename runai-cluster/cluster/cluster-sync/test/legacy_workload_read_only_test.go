/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	backendclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/test/fakecachedclient"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("legacy workload read-only behavior", func() {
	It("does not initialize distributed inference sync status", func() {
		server := CreateBasicRunaiBackend()
		server.AllowUnhandledRequests = true
		server.RouteToHandler("PUT", "/api/v1/workloads/read-only-dinf-uid",
			ghttp.RespondWith(http.StatusOK, nil))
		DeferCleanup(server.Close)

		namespace := BuildNamespace(namespaceForTest, map[string]string{
			constants.RunaiQueueLabel: projectForTest,
		})
		project := BuildKaiProjectWithDepId(projectForTest, projectIdStr1, namespaceForTest, depId)
		fakeCachedClient := fakecachedclient.NewFakeCachedClient(
			fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).WithObjects(namespace, project),
		)
		backendClient := backendclient.NewBackendClient(server.URL(), nil)
		podHandler := handlers.NewPodHandler(backendClient, fakeCachedClient)
		podRateLimiter := fakeCachedClient.NewRateLimitedController(cachedclient.PodRL, podHandler)
		workloadHandler := handlers.NewWorkloadHandler(backendClient, fakeCachedClient, podHandler, podRateLimiter)
		distributedInferenceHandler := handlers.NewDistributedInferenceHandler(workloadHandler)
		workload, _, _ := BuildDistributedInferenceWorkload("read-only-dinf", namespaceForTest, "read-only-dinf-uid")

		Expect(distributedInferenceHandler.HandleAddUpdate(workload)).To(Succeed())

		Expect(workload.Status.SyncStatus).To(BeNil())
	})
})
