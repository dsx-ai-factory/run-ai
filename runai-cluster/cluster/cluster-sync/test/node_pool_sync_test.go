/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"encoding/json"
	"fmt"
	"time"

	openapi2 "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"

	"github.com/kai-scheduler/kai-resource-management-api/kai/v1alpha1"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/handlers"

	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	backendclient "github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// Sibling of node_pool_legacy_sync_test.go for the kai.resources handler. The handler is
// type-selected rather than flag-guarded, so these specs exercise the kai path regardless
// of the compile-time UseKaiNodePools value.
var _ = Describe("NodePoolHandler", func() {
	var nodePoolsSyncPath = fmt.Sprintf("/api/v1/clusters/%s/nodepools/status", ClusterIdForTests)

	var server *ghttp.Server
	var backendClient *backendclient.BackendClient
	var stopper chan struct{}
	var nodePoolHandler *handlers.NodePoolHandler
	var syncedNodePoolsCountChan chan int
	var updatedNodePool chan string
	var phaseMessage chan string
	var gpuNetworkAccelerationDetected chan bool
	var conditionDataChan chan map[string]string
	BeforeEach(func() {
		syncedNodePoolsCountChan = make(chan int, 1)
		DeferCleanup(func() { close(syncedNodePoolsCountChan) })
		updatedNodePool = make(chan string, 1)
		DeferCleanup(func() { close(updatedNodePool) })
		phaseMessage = make(chan string, 1)
		DeferCleanup(func() { close(phaseMessage) })
		gpuNetworkAccelerationDetected = make(chan bool, 1)
		DeferCleanup(func() { close(gpuNetworkAccelerationDetected) })
		conditionDataChan = make(chan map[string]string, 1)
		DeferCleanup(func() { close(conditionDataChan) })
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
		existingNodePool := BuildNodePool("node-pool-1", "key1", "value1", v1alpha1.NodePoolReady, "update key")
		nodePoolForUpdate := BuildNodePool("node-pool-1", "key1", "value1", v1alpha1.NodePoolReady, "")
		nodePoolForAdd := BuildNodePool("node-pool-2", "key2", "value2", v1alpha1.NodePoolEmpty, "empty")

		// Node pool with NodeTopologyMismatch condition
		nodePoolWithTopologyMismatch := BuildNodePoolWithCondition(
			"node-pool-3", "key3", "value3", v1alpha1.NodePoolReady, "topology mismatch",
			constants.NodeTopologyMismatchCondition, "TopologyMismatch",
			"Node topology does not match expected configuration")

		BeforeEach(func() {

			GinkgoWriter.Printf("Created mock server calls...")
			// Sync mock
			server.RouteToHandler("POST", nodePoolsSyncPath, func(writer http.ResponseWriter, request *http.Request) {
				var syncStatusRequest openapi2.SyncNodepoolsStatusRequest
				ghttp.VerifyContentType("application/json")(writer, request)
				err := json.NewDecoder(request.Body).Decode(&syncStatusRequest)
				Expect(err).To(BeNil())
				ghttp.RespondWith(200, "ok")
				syncedNodePoolsCountChan <- len(syncStatusRequest.Nodepools)
				nodePoolName := syncStatusRequest.Nodepools[0].Name
				updatedNodePool <- nodePoolName
				gpuNetAccDetected := false
				if syncStatusRequest.Nodepools[0].Status != nil && syncStatusRequest.Nodepools[0].Status.GpuNetworkAccelerationDetected != nil {
					gpuNetAccDetected = *syncStatusRequest.Nodepools[0].Status.GpuNetworkAccelerationDetected
				}
				currentPhaseMessage := syncStatusRequest.Nodepools[0].PhaseMessage
				if currentPhaseMessage != nil {
					phaseMessage <- *currentPhaseMessage
				}

				// Capture condition data if conditions exist
				if syncStatusRequest.Nodepools[0].Status != nil && syncStatusRequest.Nodepools[0].Status.Conditions != nil && len(syncStatusRequest.Nodepools[0].Status.Conditions) > 0 {
					condition := syncStatusRequest.Nodepools[0].Status.Conditions[0]
					conditionData := map[string]string{
						"Type":    condition.Type,
						"Reason":  condition.Reason,
						"Message": "",
						"Status":  "True",
					}
					if condition.Message != nil {
						conditionData["Message"] = *condition.Message
					}
					if condition.Status != nil {
						conditionData["Status"] = *condition.Status
					}
					conditionDataChan <- conditionData
				}

				gpuNetworkAccelerationDetected <- gpuNetAccDetected
			})
		})

		JustBeforeEach(func() {
			GinkgoWriter.Println("Creating node pool controller")

			backendClient = backendclient.NewBackendClient(server.URL(), nil)
			runaiClient := fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).Build()
			nodePoolHandler = handlers.NewNodePoolHandler(backendClient, runaiClient)

			stopper = make(chan struct{})
			go backendClient.StartSendingLivenessPeriodically(stopper)
		})

		It("Test add", func() {
			err := nodePoolHandler.HandleAdd(nodePoolForAdd)
			Expect(err).To(BeNil())
			Eventually(updatedNodePool).WithTimeout(3 * time.Second).Should(Receive(Equal(nodePoolForAdd.Name)))
			Eventually(syncedNodePoolsCountChan).WithTimeout(3 * time.Second).Should(Receive(Equal(1)))
		})

		It("Test add with NodeTopologyMismatch condition", func() {
			err := nodePoolHandler.HandleAdd(nodePoolWithTopologyMismatch)
			Expect(err).To(BeNil())
			Eventually(updatedNodePool).WithTimeout(3 * time.Second).Should(Receive(Equal(nodePoolWithTopologyMismatch.Name)))
			Eventually(syncedNodePoolsCountChan).WithTimeout(3 * time.Second).Should(Receive(Equal(1)))

			// Validate condition fields are properly synced
			Eventually(conditionDataChan).WithTimeout(3 * time.Second).Should(Receive(Equal(map[string]string{
				"Type":    constants.NodeTopologyMismatchCondition,
				"Reason":  "TopologyMismatch",
				"Message": "Node topology does not match expected configuration",
				"Status":  "True",
			})))
		})

		It("Test update", func() {
			err := nodePoolHandler.HandleUpdate(nodePoolForUpdate)
			Expect(err).To(BeNil())
			Eventually(phaseMessage).WithTimeout(3 * time.Second).Should(Receive(Equal(nodePoolForUpdate.Status.Message)))
			// On the kai CRD the detection result is an annotation, not a status field.
			Eventually(gpuNetworkAccelerationDetected).WithTimeout(3 * time.Second).
				Should(Receive(BeTrue()))
		})

		It("Test delete", func() {
			err := nodePoolHandler.HandleDelete(existingNodePool)
			Expect(err).To(BeNil())
			Eventually(updatedNodePool).WithTimeout(3 * time.Second).Should(Receive(Equal(existingNodePool.Name)))
		})
	})

})
