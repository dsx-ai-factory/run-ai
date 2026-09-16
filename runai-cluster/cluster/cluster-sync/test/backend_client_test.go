/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"fmt"
	"net/http"
	"sync"

	backendclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/config"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("A backend client", func() {
	var server *ghttp.Server
	BeforeEach(func() {
		GinkgoWriter.Println("Creating server for test...")
		server = ghttp.NewServer()
		GinkgoWriter.Println("Created server for test. With URL: %s", server.URL())
		server.AllowUnhandledRequests = true
	})
	AfterEach(func() {
		server.Close()
	})
	Context("backend Client Initializing", func() {
		var backendClient *backendclient.BackendClient
		var stopper chan struct{}
		JustBeforeEach(func() {
			GinkgoWriter.Printf("Started another test")

			stopper = make(chan struct{})
			backendClient = backendclient.NewBackendClient(server.URL(), nil)
			go backendClient.StartSendingLivenessPeriodically(stopper)
		})
		JustAfterEach(func() {
			close(stopper)
			backendClient = nil
			GinkgoWriter.Printf("=============================================Ended test=============================================\n")
		})
		Context("Backend is UP", func() {
			var livenessRequests <-chan *http.Request
			BeforeEach(func() {
				livenessRequests = addHandlerToMockAndValidateLivenessRequest(server, tokenResponse)
				addHandlerToMockClusterInfoSyncRequest(server)
			})
			AfterEach(func() {
				server.Reset()
			})
			It("Sends liveness", func() {
				Eventually(livenessRequests).Should(Receive())
			})
			It("checks that backend is marked as available in backend client", func() {
				//Time-outs after 1 seconds
				Eventually(backendClient.IsRunaiBackendAvailable).Should(BeTrue())
			})
		})

		Context("Backend Does not respond to liveness", func() {
			var retryCounter *int
			var retryCounterLock sync.RWMutex
			BeforeEach(func() {
				retryCounterLock = sync.RWMutex{}
				retryCounter = new(int)
				livenessPath := fmt.Sprintf("/v1/k8s/clusters/%s/liveness", ClusterIdForTests)
				server.RouteToHandler("POST", livenessPath, func(writer http.ResponseWriter, request *http.Request) {
					GinkgoWriter.Printf("Got Liveness Request in mock %d\n", retryCounter)
					retryCounterLock.Lock()
					*retryCounter++
					retryCounterLock.Unlock()

					ghttp.RespondWith(http.StatusServiceUnavailable, "service unavailable")(writer, request)

				})
			})
			AfterEach(func() {
				server.Reset()
			})
			It("checks that service retries liveness requests", func() {
				const numOfRetries = 3
				intervals := fmt.Sprintf("%ds", numOfRetries+1)
				GinkgoWriter.Printf("%d\n", config.Config().LivenessIntervalSeconds)
				Eventually(func() int {
					retryCounterLock.RLock()
					defer retryCounterLock.RUnlock()
					return *retryCounter
				}, intervals).Should(BeNumerically(">=", numOfRetries))
			})
			It("checks that backend is marked as down", func() {
				Eventually(backendClient.IsRunaiBackendAvailable).Should(BeFalse())
			})
		})
	})
})
