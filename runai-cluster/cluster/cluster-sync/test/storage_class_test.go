/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
	"encoding/json"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	trackerclient "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	backendclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/controllers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/test/fakecachedclient"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/types"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("A storage_class controller test", func() {
	var server *ghttp.Server
	var fakeCachedClient cachedclient.CachedClient
	var backendClient *backendclient.BackendClient

	var syncedStorageClassName chan string
	var deletedStorageClassName chan string
	//var UpdatedStorageClassName chan string
	var createdStorageClassName chan string
	var syncedStorageClassNumber chan int

	var stopper chan struct{}
	var ctx context.Context
	var cancel context.CancelFunc

	BeforeEach(func() {
		// init server
		server = CreateBasicRunaiBackend()
		GinkgoWriter.Printf("Created server for test %s \n", server.URL())
		DeferCleanup(server.Close)
		server.AllowUnhandledRequests = true
	})

	Context("runai backend is alive, one storage_class is in the cluster", Serial, func() {
		var existingStorageClass = BuildStorageClass("existing-storageClasses", false)
		var addedStorageClass = BuildStorageClass("new-storageClasses-1", false)

		BeforeEach(func() {
			// init channels
			syncedStorageClassName = make(chan string, 1)
			deletedStorageClassName = make(chan string, 1)
			//UpdatedStorageClassName = make(chan string, 1)
			createdStorageClassName = make(chan string, 1)
			syncedStorageClassNumber = make(chan int, 1)

			//start server Route Handler of StorageClasses sync
			server.RouteToHandler("POST", "/api/v1/storage-classes/sync", func(writer http.ResponseWriter, request *http.Request) {
				var storageClasses trackerclient.StorageClasses
				ghttp.VerifyContentType("application/json")(writer, request)
				err := json.NewDecoder(request.Body).Decode(&storageClasses)
				Expect(err).To(BeNil())
				ghttp.RespondWith(200, "ok")(writer, request)
				syncedStorageClassNumber <- len(storageClasses.Items)
				syncedStorageClassName <- storageClasses.Items[0].StorageClassName
			})

			//start server Route Handler of storage_class create
			server.RouteToHandler("POST", "/api/v1/storage-classes", func(writer http.ResponseWriter, request *http.Request) {
				var storageClass trackerclient.StorageClass
				ghttp.VerifyContentType("application/json")(writer, request)
				err := json.NewDecoder(request.Body).Decode(&storageClass)
				Expect(err).To(BeNil())
				ghttp.RespondWith(200, "ok")(writer, request)

				if storageClass.StorageClassName == addedStorageClass.Name {
					createdStorageClassName <- storageClass.StorageClassName
					syncedStorageClassNumber <- 1
				}
			})

			//start server Route Handler of storage_class delete
			server.RouteToHandler("DELETE", "/api/v1/storage-classes", func(writer http.ResponseWriter, request *http.Request) {
				ghttp.RespondWith(http.StatusOK, "ok")(writer, request)
				deletedStorageClassName <- "existing-storageClasses"
			})

			//start server Route Handler of storage_class update
			//server.RouteToHandler("PUT", "/api/v1/storage-classes", func(writer http.ResponseWriter, request *http.Request) {
			//	UpdatedStorageClassName <- "existing-storageClasses"
			//	ghttp.RespondWith(http.StatusOK, "ok")(writer, request)
			//})

			fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).WithObjects(
				existingStorageClass)

			// 3. start storageClasses controller and Liveness
			fakeCachedClient = fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
			backendClient = backendclient.NewBackendClient(server.URL(), nil)
			controllers.AddStorageClassHandlerToCache(backendClient, fakeCachedClient)
			stopper = make(chan struct{})
			ctx, cancel = context.WithCancel(context.TODO())
			go backendClient.StartSendingLivenessPeriodically(stopper)
			fakeCachedClient.Start(ctx, stopper)
		})

		AfterEach(func() {
			close(stopper)
			cancel()
			server.Reset()

			// cleanup channels
			close(syncedStorageClassName)
			close(deletedStorageClassName)
			close(createdStorageClassName)
			close(syncedStorageClassNumber)
		})

		It("Tests storage_class sync", func() {
			Eventually(syncedStorageClassNumber).WithTimeout(threeSecondTimeout).Should(Receive(Equal(1)))
			Eventually(syncedStorageClassName).WithTimeout(threeSecondTimeout).Should(Receive(Equal(existingStorageClass.Name)))
		})

		It("Tests storage_class create", func() {
			Eventually(syncedStorageClassNumber).WithTimeout(threeSecondTimeout).Should(Receive(Equal(1)))

			err := fakeCachedClient.Create(context.TODO(), addedStorageClass)
			Expect(err).To(BeNil())

			Eventually(createdStorageClassName).WithTimeout(threeSecondTimeout).Should(Receive(Equal(addedStorageClass.Name)))
		})

		It("Tests storage_class delete", func() {
			Eventually(syncedStorageClassNumber).WithTimeout(threeSecondTimeout).Should(Receive(Equal(1)))
			Eventually(syncedStorageClassName).WithTimeout(threeSecondTimeout).Should(Receive(Equal(existingStorageClass.Name)))

			existingStorageClassToDelete := &storagev1.StorageClass{}
			err := fakeCachedClient.Get(context.TODO(), types.NamespacedName{Name: existingStorageClass.Name}, existingStorageClassToDelete)
			Expect(err).To(BeNil())

			err = fakeCachedClient.Delete(context.TODO(), existingStorageClassToDelete)
			Expect(err).To(BeNil())

			Eventually(deletedStorageClassName).WithTimeout(threeSecondTimeout).Should(Receive(Equal(existingStorageClass.Name)))

		})

		//It("Tests storage_class update", func() {
		//	Eventually(syncedStorageClassNumber).WithTimeout(threeSecondTimeout).Should(Receive(Equal(1)))
		//	Eventually(syncedStorageClassName).WithTimeout(threeSecondTimeout).Should(Receive(Equal(existingStorageClass.Name)))
		//
		//	existingSC := &storagev1.StorageClass{}
		//	err := fakeCachedClient.Get(context.TODO(), types.NamespacedName{Name: existingStorageClass.Name}, existingSC)
		//	Expect(err).To(BeNil())
		//
		//	existingSC.Annotations[IsDefaultClassLabel] = "true"
		//
		//	err = fakeCachedClient.Update(context.TODO(), existingSC)
		//	Expect(err).To(BeNil())
		//
		//	Eventually(createdStorageClassName).Should(Receive(Equal(existingStorageClass.Name)))
		//
		//})
	})
})
