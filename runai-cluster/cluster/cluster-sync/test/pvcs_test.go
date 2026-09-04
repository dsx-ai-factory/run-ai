/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/util"
	constants2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	k8sobjectstracker "github.com/dsx-ai-factory/run-ai/runai-clients/k8s-objects-tracker-client/generated"
	backendclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/controllers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/test/fakecachedclient"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("A PVC controller test", func() {
	var server *ghttp.Server
	var backendClient *backendclient.BackendClient
	var fakeCachedClient cachedclient.CachedClient

	var syncedPVCName chan string
	var deletedPVCName chan string
	var createdPVCName chan string
	var syncedPVCNumber chan int
	var tested chan bool

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

	Context("runai backend is alive, one pvc is in the cluster", Serial, func() {
		var existingPVC = BuildPVC(NamespaceIdForTests, "existing-pvc")
		var addedPVC = BuildPVC(NamespaceIdForTests, "new-pvc-1")
		var expectedIssues []k8sobjectstracker.Issue

		BeforeEach(func() {
			// init channels
			syncedPVCName = make(chan string, 1)
			deletedPVCName = make(chan string, 1)
			createdPVCName = make(chan string, 1)
			syncedPVCNumber = make(chan int, 1)
			tested = make(chan bool, 1)
			expectedIssues = nil

			//start server Route Handler of PVC sync
			server.RouteToHandler("POST", "/api/v2/pvcs/7daba6ee-2ea9-4af3-b735-a364bcdff72f/namespace/runai-namespace-project", func(writer http.ResponseWriter, request *http.Request) {
				var pvcs k8sobjectstracker.PVCs
				ghttp.VerifyContentType("application/json")(writer, request)
				err := json.NewDecoder(request.Body).Decode(&pvcs)
				Expect(err).To(BeNil())
				ghttp.RespondWith(200, "ok")(writer, request)
				syncedPVCNumber <- len(pvcs.Pvcs)
				syncedPVCName <- pvcs.Pvcs[0].Name
			})

			//start server Route Handler of PVC create
			server.RouteToHandler("POST", "/api/v2/pvcs/7daba6ee-2ea9-4af3-b735-a364bcdff72f/namespace/runai-namespace-project/new-pvc-1", func(writer http.ResponseWriter, request *http.Request) {
				ghttp.VerifyContentType("application/json")(writer, request)
				ghttp.RespondWith(200, "ok")(writer, request)

				createdPVCName <- "new-pvc-1"
				syncedPVCNumber <- 1
			})

			server.RouteToHandler("POST", "/api/v2/pvcs/7daba6ee-2ea9-4af3-b735-a364bcdff72f/namespace/runai-namespace-project/issued-pvc", func(writer http.ResponseWriter, request *http.Request) {
				ghttp.VerifyContentType("application/json")(writer, request)
				ghttp.RespondWith(200, "ok")(writer, request)

				var pvc k8sobjectstracker.PVC
				err := json.NewDecoder(request.Body).Decode(&pvc)
				Expect(err).To(BeNil())
				Expect(pvc.Issues).Should(ContainElements(expectedIssues))
				tested <- true
			})

			server.RouteToHandler("POST", "/api/v2/pvcs/7daba6ee-2ea9-4af3-b735-a364bcdff72f/namespace/runai-namespace-project/data-sharing-pvc", func(writer http.ResponseWriter, request *http.Request) {
				ghttp.VerifyContentType("application/json")(writer, request)
				ghttp.RespondWith(200, "ok")(writer, request)

				var pvc k8sobjectstracker.PVC
				err := json.NewDecoder(request.Body).Decode(&pvc)
				Expect(err).To(BeNil())
				Expect(*pvc.DataSharing.Get()).Should(BeEquivalentTo(true))
				tested <- true
			})

			//start server Route Handler of PVC delete
			server.RouteToHandler("DELETE", "/api/v2/pvcs/7daba6ee-2ea9-4af3-b735-a364bcdff72f/namespace/runai-namespace-project/existing-pvc", func(writer http.ResponseWriter, request *http.Request) {
				ghttp.RespondWith(http.StatusOK, "ok")(writer, request)
				deletedPVCName <- "existing-pvc"
			})

			//start server Route Handler for SyncAssetsResourcesIds
			server.RouteToHandler("POST", "/api/v1/asset/sync/asset-resources-ids", func(writer http.ResponseWriter, request *http.Request) {
				ghttp.VerifyContentType("application/json")(writer, request)
				ghttp.RespondWith(200, "ok")(writer, request)
				syncedPVCNumber <- 1
			})

			runaiManagedNamespace := BuildNamespace(NamespaceIdForTests, map[string]string{constants.RunaiQueueLabel: ProjectIdForTests})
			runaiProject := BuildProject(ProjectIdForTests, ProjectIdForTests, NamespaceIdForTests)
			kaiProject := BuildKaiProject(ProjectIdForTests, ProjectIdForTests, NamespaceIdForTests)

			// start pvc controller and Liveness
			backendClient = backendclient.NewBackendClient(server.URL(), nil)

			fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).WithObjects(runaiManagedNamespace, runaiProject, kaiProject, existingPVC)
			fakeCachedClient = fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
			controllers.AddPVCHandlerToCache(backendClient, fakeCachedClient)

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
			close(syncedPVCName)
			close(deletedPVCName)
			close(createdPVCName)
			close(syncedPVCNumber)
			close(tested)
		})

		It("Tests PVC sync", func() {
			Eventually(syncedPVCNumber).WithTimeout(threeSecondTimeout).Should(Receive(Equal(1)))
			Eventually(syncedPVCName).WithTimeout(threeSecondTimeout).Should(Receive(Equal(existingPVC.Name)))
		})

		It("Tests PVC create", func() {
			Eventually(syncedPVCNumber).WithTimeout(threeSecondTimeout).Should(Receive(Equal(1)))

			err := fakeCachedClient.Create(context.TODO(), addedPVC)
			Expect(err).To(BeNil())

			Eventually(createdPVCName).WithTimeout(threeSecondTimeout).Should(Receive(Equal(addedPVC.Name)))
		})

		It("Tests PVC delete", func() {
			Eventually(syncedPVCNumber).WithTimeout(threeSecondTimeout).Should(Receive(Equal(1)))
			Eventually(syncedPVCName).WithTimeout(threeSecondTimeout).Should(Receive(Equal(existingPVC.Name)))

			pvcToDelete := &corev1.PersistentVolumeClaim{}
			err := fakeCachedClient.Get(context.TODO(), types.NamespacedName{Name: existingPVC.Name, Namespace: NamespaceIdForTests}, pvcToDelete)
			Expect(err).To(BeNil())

			err = fakeCachedClient.Delete(context.TODO(), pvcToDelete)
			Expect(err).To(BeNil())

			Eventually(deletedPVCName).WithTimeout(threeSecondTimeout).Should(Receive(Equal(existingPVC.Name)))

		})

		It("Tests PVC create with issue", func() {
			Eventually(syncedPVCNumber).WithTimeout(threeSecondTimeout).Should(Receive(Equal(1)))
			issuedPVC := BuildPVC(NamespaceIdForTests, "issued-pvc")
			issuedPVC.Status.Phase = corev1.ClaimLost
			// Add clone issue on the pvc (mimic project controller clone issue)
			expectedIssues = SetIssuesOnNamespace(&issuedPVC.ObjectMeta, "ns1", "ns2")
			// Set lost status on the PVC
			expectedIssues = append(expectedIssues, k8sobjectstracker.Issue{
				Namespace: issuedPVC.Namespace,
				Issue:     string(k8sobjectstracker.CLUSTEROBJECTISSUETYPE_LOST_PHASE),
			})

			err := fakeCachedClient.Create(context.TODO(), issuedPVC)
			Expect(err).To(BeNil())
			Eventually(tested).WithTimeout(threeSecondTimeout).Should(Receive(BeTrue()))
		})

		It("Test PVC with dataSharingLabel", func() {
			Eventually(syncedPVCNumber).WithTimeout(threeSecondTimeout).Should(Receive(Equal(1)))
			sharedPVC := BuildPVC(NamespaceIdForTests, "data-sharing-pvc")
			sharedPVC.Labels[constants2.RunaiDataSharingLabel] = util.TrueAsString

			err := fakeCachedClient.Create(context.TODO(), sharedPVC)
			Expect(err).To(BeNil())
			Eventually(tested).WithTimeout(threeSecondTimeout).Should(Receive(BeTrue()))
		})

		It("Test PVC with asset ID label", func() {
			Eventually(syncedPVCNumber).WithTimeout(threeSecondTimeout).Should(Receive(Equal(1)))
			assetPVC := BuildPVC(NamespaceIdForTests, "asset-pvc")
			assetPVC.Labels[constants2.RunaiAssetIdLabel] = "test-asset-id"
			err := fakeCachedClient.Create(context.TODO(), assetPVC)
			Expect(err).To(BeNil())
			// The SyncAssetsResourcesIds endpoint should be called when a PVC with asset ID is created
		})

	})
})
