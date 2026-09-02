/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"encoding/json"
	"github.com/google/uuid"
	"k8s.io/utils/ptr"
	"net/http"
	"time"

	"github.com/onsi/gomega/ghttp"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/test/fakecachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	backendclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers"

	swagger "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
)

var _ = Describe("PriorityClassHandler", func() {
	pcId, _ := uuid.NewUUID()
	priorityClassPath := "/api/v1/workload-priorities/runai-status"
	priorityClassClusterPath := "/api/v1/workload-priorities/cluster-status"

	var server *ghttp.Server
	var backendClient *backendclient.BackendClient
	var stopper chan struct{}
	var handler *handlers.PriorityClassHandler
	var updatedPc chan swagger.PrioritiesForStatus

	BeforeEach(func() {
		updatedPc = make(chan swagger.PrioritiesForStatus, 1)
		DeferCleanup(func() { close(updatedPc) })
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

	Context("test handler", func() {
		testPc := BuildPriorityClass("pc-1", pcId.String(), 1, true)
		testPcCluster := BuildPriorityClass("pc-1", pcId.String(), 1, false)

		BeforeEach(func() {

			GinkgoWriter.Printf("Created mock server calls...")
			// Update status mock
			server.RouteToHandler("POST", priorityClassPath, func(writer http.ResponseWriter, request *http.Request) {
				ghttp.VerifyContentType("application/json")(writer, request)

				var bePc swagger.PrioritiesForStatus
				err := json.NewDecoder(request.Body).Decode(&bePc)
				Expect(err).To(BeNil())

				ghttp.RespondWith(http.StatusOK, "ok")(writer, request)
				updatedPc <- bePc
			})
			server.RouteToHandler("POST", priorityClassClusterPath, func(writer http.ResponseWriter, request *http.Request) {
				ghttp.VerifyContentType("application/json")(writer, request)

				var bePc swagger.PrioritiesForStatus
				err := json.NewDecoder(request.Body).Decode(&bePc)
				Expect(err).To(BeNil())

				ghttp.RespondWith(http.StatusOK, "ok")(writer, request)
				updatedPc <- bePc
			})
		})

		JustBeforeEach(func() {
			GinkgoWriter.Println("Creating priority class controller")
			backendClient = backendclient.NewBackendClient(server.URL(), nil)
			fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme())
			fakeCachedClient := fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
			handler = handlers.NewPriorityClassHandler(backendClient, fakeCachedClient)
			stopper = make(chan struct{})
			go backendClient.StartSendingLivenessPeriodically(stopper)
		})

		It("Test add update - runai", func() {
			err := handler.HandleAddUpdate(testPc)
			Expect(err).To(BeNil())
			Eventually(updatedPc).WithTimeout(3 * time.Second).Should(Receive(Equal(swagger.PrioritiesForStatus{
				FullSync: ptr.To(false),
				Priorities: []swagger.PrioritiesForStatusPrioritiesInner{{
					Id:     ptr.To(pcId.String()),
					Name:   ptr.To("pc-1"),
					Value:  ptr.To(int32(1)),
					Source: ptr.To(constants.PriorityClassRunaiSource),
					Status: &swagger.PrioritiesForStatusPrioritiesInnerAllOfStatus{Phase: ptr.To("Ready")},
				}},
			})))
		})

		It("Test add update - cluster", func() {
			err := handler.HandleAddUpdate(testPcCluster)
			Expect(err).To(BeNil())
			Eventually(updatedPc).WithTimeout(3 * time.Second).Should(Receive(Equal(swagger.PrioritiesForStatus{
				FullSync: ptr.To(false),
				Priorities: []swagger.PrioritiesForStatusPrioritiesInner{{
					Name:   ptr.To("pc-1"),
					Value:  ptr.To(int32(1)),
					Source: ptr.To(""),
					Status: &swagger.PrioritiesForStatusPrioritiesInnerAllOfStatus{Phase: ptr.To("Ready")},
				}},
			})))
		})

		It("Test delete", func() {
			err := handler.HandleDelete(testPc)
			Expect(err).To(BeNil())
			Eventually(updatedPc).WithTimeout(3 * time.Second).Should(Receive(Equal(swagger.PrioritiesForStatus{
				FullSync: ptr.To(false),
				Priorities: []swagger.PrioritiesForStatusPrioritiesInner{{
					Id:     ptr.To(pcId.String()),
					Name:   ptr.To("pc-1"),
					Value:  ptr.To(int32(1)),
					Source: ptr.To(constants.PriorityClassRunaiSource),
					Status: &swagger.PrioritiesForStatusPrioritiesInnerAllOfStatus{Phase: ptr.To("Deleted")},
				}},
			})))
		})

		It("Test delete - cluster", func() {
			err := handler.HandleDelete(testPcCluster)
			Expect(err).To(BeNil())
			Eventually(updatedPc).WithTimeout(3 * time.Second).Should(Receive(Equal(swagger.PrioritiesForStatus{
				FullSync: ptr.To(false),
				Priorities: []swagger.PrioritiesForStatusPrioritiesInner{{
					Name:   ptr.To("pc-1"),
					Value:  ptr.To(int32(1)),
					Source: ptr.To(""),
					Status: &swagger.PrioritiesForStatusPrioritiesInnerAllOfStatus{Phase: ptr.To("Deleted")},
				}},
			})))
		})
	})

})
