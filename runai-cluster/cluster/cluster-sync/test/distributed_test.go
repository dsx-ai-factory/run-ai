/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/onsi/gomega/ghttp"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/generated/swagger-runai-backend-client"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	backendclient "github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("DistributedHandlerDeprecated", func() {
	var distributedPath = fmt.Sprintf("/v1/k8s/clusters/%s/distributed", ClusterIdForTests)

	var server *ghttp.Server
	var backendClient *backendclient.BackendClient
	var stopper chan struct{}
	var distributedHandler *handlers.DistributedHandlerDeprecated
	var updatedDistributed chan string
	var updatedDistributedStatus chan string

	BeforeEach(func() {
		updatedDistributed = make(chan string, 1)
		updatedDistributedStatus = make(chan string, 1)

		DeferCleanup(func() {
			close(updatedDistributed)
			close(updatedDistributedStatus)
		})
		DeferCleanup(func() {})
		GinkgoWriter.Println("Creating server for test...")
		server = CreateBasicRunaiBackend()
		DeferCleanup(func() {
			server.Close()
		})
		GinkgoWriter.Printf("Created server for test %s \n", server.URL())
		server.AllowUnhandledRequests = true
		stopper = make(chan struct{}, 1)
		DeferCleanup(func() { close(stopper) })
	})

	Context("runai backend is available", func() {
		testDistributed := BuildDistributed(NamespaceIdForTests, "distributed-1", "user-1", "workload-1", "testStatus")

		BeforeEach(func() {

			GinkgoWriter.Printf("Created mock server calls...")
			// Update status mock
			server.RouteToHandler("PUT", fmt.Sprintf("%s/status", distributedPath), func(writer http.ResponseWriter, request *http.Request) {
				body, err := io.ReadAll(request.Body)
				if err != nil {
					return
				}
				distributedStatusRequest := &swagger.WorkloadStatus{}
				err = json.Unmarshal(body, distributedStatusRequest)
				if err != nil {
					return
				}

				distributedId := distributedStatusRequest.Id
				distributedStatus := distributedStatusRequest.Status
				ghttp.VerifyContentType("application/json")(writer, request)
				ghttp.RespondWith(http.StatusOK, "ok")(writer, request)
				updatedDistributed <- distributedId
				updatedDistributedStatus <- distributedStatus
			})
		})

		JustBeforeEach(func() {
			GinkgoWriter.Println("Creating distributed controller")
			backendClient = backendclient.NewBackendClient(server.URL(), nil)
			runaiClient := fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).Build()
			distributedHandler = handlers.NewDistributedHandlerDeprecated(runaiClient, backendClient)
			stopper = make(chan struct{})
			go backendClient.StartSendingLivenessPeriodically(stopper)
		})

		It("Test add update", func() {
			err := distributedHandler.HandleAdd(testDistributed)
			Expect(err).To(BeNil())
			Eventually(updatedDistributed).WithTimeout(3 * time.Second).Should(Receive(Equal(testDistributed.Annotations[commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload])))
			Eventually(updatedDistributedStatus).WithTimeout(3 * time.Second).Should(Receive(Equal(testDistributed.Status.CreationStatus)))
		})
	})

})
