/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"

	"github.com/onsi/gomega/ghttp"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/test/fakecachedclient"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	backendclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers"

	swagger "github.com/dsx-ai-factory/run-ai/runai-clients/workloads-client/generated"
)

var _ = Describe("ApplicationHandler", func() {
	applicationId, _ := uuid.NewUUID()
	applicationPath := "/api/v1/ai-applications/sync"

	var server *ghttp.Server
	var backendClient *backendclient.BackendClient
	var stopper chan struct{}
	var handler *handlers.ApplicationHandler
	var updatedApplication chan swagger.AIApplicationsForSync

	BeforeEach(func() {
		updatedApplication = make(chan swagger.AIApplicationsForSync, 1)
		DeferCleanup(func() { close(updatedApplication) })
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
		testApp := BuildApplication("app-1", types.UID(applicationId.String()), "source-1", "type-1", "10", "department-ten", "1", "proj-1", "deployed", "123")

		BeforeEach(func() {

			GinkgoWriter.Printf("Created mock server calls...")
			// Update status mock
			server.RouteToHandler("POST", applicationPath, func(writer http.ResponseWriter, request *http.Request) {
				ghttp.VerifyContentType("application/json")(writer, request)

				var beApp swagger.AIApplicationsForSync
				err := json.NewDecoder(request.Body).Decode(&beApp)
				Expect(err).To(BeNil())

				ghttp.RespondWith(http.StatusOK, "ok")(writer, request)
				updatedApplication <- beApp
			})
		})

		JustBeforeEach(func() {
			GinkgoWriter.Println("Creating applications controller")
			backendClient = backendclient.NewBackendClient(server.URL(), nil)
			fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme())
			fakeCachedClient := fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
			handler = handlers.NewApplicationHandler(backendClient, fakeCachedClient)
			stopper = make(chan struct{})
			go backendClient.StartSendingLivenessPeriodically(stopper)
		})

		It("Test add update", func() {
			err := handler.HandleAddUpdate(testApp)
			Expect(err).To(BeNil())
			Eventually(updatedApplication).WithTimeout(3 * time.Second).Should(Receive(Equal(swagger.AIApplicationsForSync{
				FullSync: ptr.To(false),
				Applications: []swagger.AIApplicationForSync{{
					Id:             ptr.To(applicationId.String()),
					Name:           ptr.To("app-1"),
					Source:         ptr.To("source-1"),
					Type:           ptr.To("type-1"),
					DepartmentId:   ptr.To("10"),
					DepartmentName: ptr.To("department-ten"),
					ProjectId:      ptr.To("1"),
					ProjectName:    ptr.To("proj-1"),
					K8sPhase:       ptr.To("deployed"),
					Resources: []swagger.AIApplicationSyncResource{{
						Id:          ptr.To("123123"),
						Name:        ptr.To("pvc-1-cp"),
						ClusterId:   ptr.To("123"),
						ClusterName: ptr.To("pvc-1"),
						Kind:        ptr.To("pvc"),
					}},
					Action: ptr.To(swagger.AIAPPLICATIONSYNCACTION_UPSERT),
				}},
			})))
		})

		It("Test delete", func() {
			err := handler.HandleDelete(testApp)
			Expect(err).To(BeNil())
			Eventually(updatedApplication).WithTimeout(3 * time.Second).Should(Receive(Equal(swagger.AIApplicationsForSync{
				FullSync: ptr.To(false),
				Applications: []swagger.AIApplicationForSync{{
					Id:             ptr.To(applicationId.String()),
					Name:           ptr.To("app-1"),
					Source:         ptr.To("source-1"),
					Type:           ptr.To("type-1"),
					DepartmentId:   ptr.To("10"),
					DepartmentName: ptr.To("department-ten"),
					ProjectId:      ptr.To("1"),
					ProjectName:    ptr.To("proj-1"),
					K8sPhase:       ptr.To("deployed"),
					Resources: []swagger.AIApplicationSyncResource{{
						Id:          ptr.To("123123"),
						Name:        ptr.To("pvc-1-cp"),
						ClusterId:   ptr.To("123"),
						ClusterName: ptr.To("pvc-1"),
						Kind:        ptr.To("pvc"),
					}},
					Action: ptr.To(swagger.AIAPPLICATIONSYNCACTION_DELETE),
				}},
			})))
		})
	})

})
