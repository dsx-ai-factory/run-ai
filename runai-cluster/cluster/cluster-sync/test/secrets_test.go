/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	k8sobjectstracker "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client/generated"
	backendclient "github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/controllers"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/test/fakecachedclient"
	commonconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("A Secret controller test", func() {
	var server *ghttp.Server
	var fakeCachedClient cachedclient.CachedClient
	var backendClient *backendclient.BackendClient

	var syncedSecretName chan string
	var deletedSecretName chan string
	var createdSecretName chan string
	var syncedSecretNumber chan int
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

	Context("runai backend is alive, one secret is in the cluster", Serial, func() {
		var existingSecret = BuildSecret(NamespaceIdForTests, "existing-secret", commonconstants.RunaiResourceLabel, "password")
		var addedSecret = BuildSecret(NamespaceIdForTests, "new-secret-1", commonconstants.RunaiResourceLabel, "password")
		var expectedIssues []k8sobjectstracker.Issue

		BeforeEach(func() {
			// init channels
			syncedSecretName = make(chan string, 1)
			deletedSecretName = make(chan string, 1)
			createdSecretName = make(chan string, 1)
			syncedSecretNumber = make(chan int, 1)
			expectedIssues = nil
			tested = make(chan bool, 1)

			//start server Route Handler of Secret sync
			server.RouteToHandler("POST", "/api/v1/secrets/7daba6ee-2ea9-4af3-b735-a364bcdff72f/namespace/runai-namespace-project", func(writer http.ResponseWriter, request *http.Request) {
				var secrets k8sobjectstracker.Secrets
				ghttp.VerifyContentType("application/json")(writer, request)
				err := json.NewDecoder(request.Body).Decode(&secrets)
				Expect(err).To(BeNil())
				ghttp.RespondWith(200, "ok")(writer, request)
				syncedSecretNumber <- len(secrets.Secrets)
				syncedSecretName <- secrets.Secrets[0].Name
			})

			//start server Route Handler of Secret create
			server.RouteToHandler("POST", "/api/v1/secrets/7daba6ee-2ea9-4af3-b735-a364bcdff72f/namespace/runai-namespace-project/new-secret-1", func(writer http.ResponseWriter, request *http.Request) {
				ghttp.VerifyContentType("application/json")(writer, request)
				ghttp.RespondWith(200, "ok")(writer, request)

				createdSecretName <- "new-secret-1"
				syncedSecretNumber <- 1
			})

			server.RouteToHandler("POST", "/api/v1/secrets/7daba6ee-2ea9-4af3-b735-a364bcdff72f/namespace/runai-namespace-project/issued-secret", func(writer http.ResponseWriter, request *http.Request) {
				ghttp.VerifyContentType("application/json")(writer, request)
				ghttp.RespondWith(200, "ok")(writer, request)

				var secret k8sobjectstracker.Secret
				err := json.NewDecoder(request.Body).Decode(&secret)
				Expect(err).To(BeNil())

				Expect(secret.Issues).Should(ContainElements(expectedIssues))
				tested <- true
			})

			//start server Route Handler of Secret delete
			server.RouteToHandler("DELETE", "/api/v1/secrets/7daba6ee-2ea9-4af3-b735-a364bcdff72f/namespace/runai-namespace-project/existing-secret", func(writer http.ResponseWriter, request *http.Request) {
				ghttp.RespondWith(http.StatusOK, "ok")(writer, request)
				deletedSecretName <- "existing-secret"
			})

			runaiManagedNamespace := BuildNamespace(NamespaceIdForTests, map[string]string{constants.RunaiQueueLabel: ProjectIdForTests})
			runaiProject := BuildProject(ProjectIdForTests, ProjectIdForTests, NamespaceIdForTests)
			kaiProject := BuildKaiProject(ProjectIdForTests, ProjectIdForTests, NamespaceIdForTests)

			fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).WithObjects(runaiManagedNamespace, runaiProject, kaiProject, existingSecret)
			fakeCachedClient = fakecachedclient.NewFakeCachedClient(fakeClientBuilder)

			// 3. start secret controller and Liveness
			backendClient = backendclient.NewBackendClient(server.URL(), nil)
			controllers.AddSecretHandlerToCache(backendClient, fakeCachedClient)
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
			close(syncedSecretName)
			close(deletedSecretName)
			close(createdSecretName)
			close(syncedSecretNumber)
			close(tested)
		})

		It("Tests Secret sync", func() {
			Eventually(syncedSecretNumber).WithTimeout(threeSecondTimeout).Should(Receive(Equal(1)))
			Eventually(syncedSecretName).WithTimeout(threeSecondTimeout).Should(Receive(Equal(existingSecret.Name)))
		})

		It("Tests Secret create", func() {
			Eventually(syncedSecretNumber).WithTimeout(threeSecondTimeout).Should(Receive(Equal(1)))

			err := fakeCachedClient.Create(context.TODO(), addedSecret)
			Expect(err).To(BeNil())

			Eventually(createdSecretName).WithTimeout(threeSecondTimeout).Should(Receive(Equal(addedSecret.Name)))
		})

		It("Tests Secret delete", func() {
			Eventually(syncedSecretNumber).WithTimeout(threeSecondTimeout).Should(Receive(Equal(1)))
			Eventually(syncedSecretName).WithTimeout(threeSecondTimeout).Should(Receive(Equal(existingSecret.Name)))

			secretToDelete := &corev1.Secret{}
			err := fakeCachedClient.Get(context.TODO(), types.NamespacedName{Name: existingSecret.Name, Namespace: NamespaceIdForTests}, secretToDelete)
			Expect(err).To(BeNil())

			err = fakeCachedClient.Delete(context.TODO(), secretToDelete)
			Expect(err).To(BeNil())

			Eventually(deletedSecretName).WithTimeout(threeSecondTimeout).Should(Receive(Equal(existingSecret.Name)))

		})

		It("Tests Secret create with issue", func() {
			Eventually(syncedSecretNumber).WithTimeout(threeSecondTimeout).Should(Receive(Equal(1)))
			var issuedSecret = BuildSecret(NamespaceIdForTests, "issued-secret", commonconstants.RunaiResourceLabel, "password")
			expectedIssues = SetIssuesOnNamespace(&issuedSecret.ObjectMeta, "ns1", "ns2")
			err := fakeCachedClient.Create(context.TODO(), issuedSecret)
			Expect(err).To(BeNil())
			Eventually(tested).WithTimeout(time.Second * 10).Should(Receive(BeTrue()))
		})
	})
})
