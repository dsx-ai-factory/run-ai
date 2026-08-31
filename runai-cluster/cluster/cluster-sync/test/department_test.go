/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/onsi/gomega/ghttp"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/test/fakecachedclient"
	clustercommon "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	backendclient "github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/handlers"
)

var _ = Describe("DepartmentHandler", func() {
	var departmentId = "1"
	var departmentsPath = "/api/v1/org-unit/departments/%s/status"

	var server *ghttp.Server
	var backendClient *backendclient.BackendClient
	var stopper chan struct{}
	var handler *handlers.DepartmentHandler
	var updatedDep chan string

	BeforeEach(func() {
		updatedDep = make(chan string, 1)
		DeferCleanup(func() { close(updatedDep) })
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
		testDep := BuildKaiDepartment("dep-1", departmentId)

		BeforeEach(func() {

			GinkgoWriter.Printf("Created mock server calls...")
			// Update status mock
			server.RouteToHandler("PUT", fmt.Sprintf(departmentsPath, departmentId), func(writer http.ResponseWriter, request *http.Request) {
				ghttp.VerifyContentType("application/json")(writer, request)
				ghttp.RespondWith(http.StatusOK, "ok")(writer, request)
				updatedDep <- departmentId
			})
		})

		JustBeforeEach(func() {
			GinkgoWriter.Println("Creating department controller")
			backendClient = backendclient.NewBackendClient(server.URL(), nil)
			fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme())
			fakeCachedClient := fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
			handler = handlers.NewDepartmentHandler(backendClient, fakeCachedClient)
			stopper = make(chan struct{})
			go backendClient.StartSendingLivenessPeriodically(stopper)
		})

		It("Test add update", func() {
			err := handler.HandleAddUpdate(testDep)
			Expect(err).To(BeNil())
			Eventually(updatedDep).WithTimeout(3 * time.Second).Should(Receive(Equal(testDep.Labels[clustercommon.SyncedDepartmentIdLabel])))
		})
	})

})
