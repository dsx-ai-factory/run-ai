/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/test/fakecachedclient"
	"github.com/onsi/gomega/ghttp"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	backendclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	clustercommon "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"

	swagger "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
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

	Context("application deleted while its update is sent", func() {
		var application *v1alpha1.Application
		var fakeCachedClient cachedclient.CachedClient
		var syncedApplications chan string
		var onUpsert func()
		key := types.NamespacedName{Name: "app-2"}

		BeforeEach(func() {
			application = BuildApplication("app-2", "app-uid-1", "source-1", "type-1", "10", "department-ten", "1", "proj-1", "deployed", "123")
			syncedApplications = make(chan string, 5)
			DeferCleanup(func() { close(syncedApplications) })
			onUpsert = func() {}

			server.RouteToHandler("POST", applicationPath, func(writer http.ResponseWriter, request *http.Request) {
				var beApp swagger.AIApplicationsForSync
				Expect(json.NewDecoder(request.Body).Decode(&beApp)).To(Succeed())
				for _, app := range beApp.Applications {
					if app.GetAction() == swagger.AIAPPLICATIONSYNCACTION_UPSERT {
						onUpsert()
					}
					syncedApplications <- fmt.Sprintf("%s %s", app.GetAction(), app.GetId())
				}
				ghttp.RespondWith(http.StatusOK, "ok")(writer, request)
			})
		})

		JustBeforeEach(func() {
			backendClient = backendclient.NewBackendClient(server.URL(), nil)
			fakeCachedClient = fakecachedclient.NewFakeCachedClient(
				fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).WithObjects(application))
			handler = handlers.NewApplicationHandler(backendClient, fakeCachedClient)
		})

		It("does not delete an application that still exists", func() {
			Expect(handler.HandleUpdateItem(key)).To(Succeed())

			Expect(syncedApplications).To(Receive(Equal("Upsert app-uid-1")))
			Expect(syncedApplications).NotTo(Receive())
		})

		It("deletes the application again when it was deleted during the update", func() {
			onUpsert = func() {
				Expect(fakeCachedClient.Delete(context.Background(), application)).To(Succeed())
			}

			Expect(handler.HandleUpdateItem(key)).To(Succeed())

			Expect(syncedApplications).To(Receive(Equal("Upsert app-uid-1")))
			Expect(syncedApplications).To(Receive(Equal("Delete app-uid-1")))
		})

		It("deletes the updated application when another application replaced it under the same name", func() {
			replacement := BuildApplication("app-2", "app-uid-2", "source-1", "type-1", "10", "department-ten", "1", "proj-1", "deployed", "123")
			onUpsert = func() {
				Expect(fakeCachedClient.Delete(context.Background(), application)).To(Succeed())
				Expect(fakeCachedClient.Create(context.Background(), replacement)).To(Succeed())
			}

			Expect(handler.HandleUpdateItem(key)).To(Succeed())

			Expect(syncedApplications).To(Receive(Equal("Upsert app-uid-1")))
			Expect(syncedApplications).To(Receive(Equal("Delete app-uid-1")))
			Expect(syncedApplications).NotTo(Receive())
		})

		Context("managed by Run:ai", func() {
			BeforeEach(func() {
				application.Annotations = map[string]string{clustercommon.ControlPlaneApplicationIdAnnotation: "cp-app-id-1"}
			})

			It("does not delete the application when it was recreated with the same control-plane ID", func() {
				recreated := BuildApplication("app-2", "app-uid-2", "source-1", "type-1", "10", "department-ten", "1", "proj-1", "deployed", "123")
				recreated.Annotations = map[string]string{clustercommon.ControlPlaneApplicationIdAnnotation: "cp-app-id-1"}
				onUpsert = func() {
					Expect(fakeCachedClient.Delete(context.Background(), application)).To(Succeed())
					Expect(fakeCachedClient.Create(context.Background(), recreated)).To(Succeed())
				}

				Expect(handler.HandleUpdateItem(key)).To(Succeed())

				Expect(syncedApplications).To(Receive(Equal("Upsert cp-app-id-1")))
				Expect(syncedApplications).NotTo(Receive())
			})
		})
	})
})
