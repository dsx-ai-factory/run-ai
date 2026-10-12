/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
	"encoding/json"
	"net/http"

	trackerclient "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	backendclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/controllers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/test/fakecachedclient"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	resourcev1 "k8s.io/api/resource/v1"
	"k8s.io/apimachinery/pkg/types"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// Exercises the DeviceClass controller/handler after the per-item CRUD
// transition. The sync endpoint still exists as a drift-correction fallback;
// Add / Delete issue direct POST/DELETE and Update feeds the rate-limited
// controller which replays through HandleUpdateItem → PUT.
var _ = Describe("A device_class controller test", func() {
	var server *ghttp.Server
	var fakeCachedClient cachedclient.CachedClient
	var backendClient *backendclient.BackendClient

	// Named receive channels - one per endpoint - so Eventually can block on
	// a specific HTTP call without a shared sentinel making it ambiguous
	// which request actually arrived.
	var syncedDeviceClassNumber chan int
	var syncedDeviceClassName chan string
	var createdDeviceClassName chan string
	var updatedDeviceClassName chan string
	var deletedDeviceClassName chan string

	var stopper chan struct{}
	var ctx context.Context
	var cancel context.CancelFunc

	BeforeEach(func() {
		server = CreateBasicRunaiBackend()
		GinkgoWriter.Printf("Created server for test %s \n", server.URL())
		DeferCleanup(server.Close)
		server.AllowUnhandledRequests = true
	})

	Context("backend is alive, one DeviceClass already lives in the cluster", Serial, func() {
		existingDeviceClass := BuildDeviceClass("gpu.nvidia.com")
		addedDeviceClass := BuildDeviceClass("mig.nvidia.com")

		BeforeEach(func() {
			syncedDeviceClassNumber = make(chan int, 1)
			syncedDeviceClassName = make(chan string, 1)
			createdDeviceClassName = make(chan string, 1)
			updatedDeviceClassName = make(chan string, 1)
			deletedDeviceClassName = make(chan string, 1)

			// Sync endpoint posts the full inventory; verify the count matches
			// what's in the fake cluster and surface the first name so the
			// test can assert on identity.
			server.RouteToHandler("POST", "/api/v1/device-classes/sync", func(writer http.ResponseWriter, request *http.Request) {
				var payload trackerclient.DeviceClasses
				ghttp.VerifyContentType("application/json")(writer, request)
				Expect(json.NewDecoder(request.Body).Decode(&payload)).To(Succeed())
				ghttp.RespondWith(http.StatusOK, "ok")(writer, request)
				syncedDeviceClassNumber <- len(payload.Items)
				if len(payload.Items) > 0 {
					syncedDeviceClassName <- payload.Items[0].DeviceClassName
				}
			})

			// Per-item create (POST) - informer Add fires it directly. The
			// pre-existing DeviceClass also fires its own Add at startup, so
			// the channel is only written to for the one we add during the
			// test; otherwise the test would race on which POST arrived first.
			server.RouteToHandler("POST", "/api/v1/device-classes", func(writer http.ResponseWriter, request *http.Request) {
				var payload trackerclient.DeviceClass
				ghttp.VerifyContentType("application/json")(writer, request)
				Expect(json.NewDecoder(request.Body).Decode(&payload)).To(Succeed())
				ghttp.RespondWith(http.StatusOK, "ok")(writer, request)
				if payload.DeviceClassName == addedDeviceClass.Name {
					createdDeviceClassName <- payload.DeviceClassName
				}
			})

			// Per-item update (PUT) - informer Update feeds the rate limiter
			// which eventually invokes HandleDeviceClassUpdate.
			server.RouteToHandler("PUT", "/api/v1/device-classes", func(writer http.ResponseWriter, request *http.Request) {
				var payload trackerclient.DeviceClass
				ghttp.VerifyContentType("application/json")(writer, request)
				Expect(json.NewDecoder(request.Body).Decode(&payload)).To(Succeed())
				ghttp.RespondWith(http.StatusOK, "ok")(writer, request)
				updatedDeviceClassName <- payload.DeviceClassName
			})

			// Per-item delete - path carries the name as a `name` query param
			// alongside `clusterId`, so the handler asserts on the query
			// string rather than on the (empty) body.
			server.RouteToHandler("DELETE", "/api/v1/device-classes", func(writer http.ResponseWriter, request *http.Request) {
				ghttp.RespondWith(http.StatusOK, "ok")(writer, request)
				deletedDeviceClassName <- request.URL.Query().Get("name")
			})

			fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).WithObjects(existingDeviceClass)
			fakeCachedClient = fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
			backendClient = backendclient.NewBackendClient(server.URL(), nil)
			controllers.AddDeviceClassHandlerToCache(backendClient, fakeCachedClient)

			stopper = make(chan struct{})
			ctx, cancel = context.WithCancel(context.TODO())
			go backendClient.StartSendingLivenessPeriodically(stopper)
			fakeCachedClient.Start(ctx, stopper)
		})

		AfterEach(func() {
			close(stopper)
			cancel()
			server.Reset()

			close(syncedDeviceClassNumber)
			close(syncedDeviceClassName)
			close(createdDeviceClassName)
			close(updatedDeviceClassName)
			close(deletedDeviceClassName)
		})

		It("ships the existing DeviceClass through the full-set sync on startup", func() {
			Eventually(syncedDeviceClassNumber).WithTimeout(threeSecondTimeout).Should(Receive(Equal(1)))
			Eventually(syncedDeviceClassName).WithTimeout(threeSecondTimeout).Should(Receive(Equal(existingDeviceClass.Name)))
		})

		It("POSTs the new DeviceClass per-item (not through sync) on Add", func() {
			Eventually(syncedDeviceClassNumber).WithTimeout(threeSecondTimeout).Should(Receive(Equal(1)))

			Expect(fakeCachedClient.Create(context.TODO(), addedDeviceClass)).To(Succeed())

			Eventually(createdDeviceClassName).WithTimeout(threeSecondTimeout).Should(Receive(Equal(addedDeviceClass.Name)))
		})

		It("DELETEs the DeviceClass per-item on Delete", func() {
			Eventually(syncedDeviceClassNumber).WithTimeout(threeSecondTimeout).Should(Receive(Equal(1)))

			toDelete := &resourcev1.DeviceClass{}
			Expect(fakeCachedClient.Get(context.TODO(),
				types.NamespacedName{Name: existingDeviceClass.Name}, toDelete)).To(Succeed())
			Expect(fakeCachedClient.Delete(context.TODO(), toDelete)).To(Succeed())

			Eventually(deletedDeviceClassName).WithTimeout(threeSecondTimeout).Should(Receive(Equal(existingDeviceClass.Name)))
		})

		It("PUTs the DeviceClass per-item on Update via the rate-limited queue", func() {
			Eventually(syncedDeviceClassNumber).WithTimeout(threeSecondTimeout).Should(Receive(Equal(1)))

			current := &resourcev1.DeviceClass{}
			Expect(fakeCachedClient.Get(context.TODO(),
				types.NamespacedName{Name: existingDeviceClass.Name}, current)).To(Succeed())
			if current.Annotations == nil {
				current.Annotations = map[string]string{}
			}
			current.Annotations["edited"] = "true"
			Expect(fakeCachedClient.Update(context.TODO(), current)).To(Succeed())

			Eventually(updatedDeviceClassName).WithTimeout(threeSecondTimeout).Should(Receive(Equal(existingDeviceClass.Name)))
		})
	})
})
