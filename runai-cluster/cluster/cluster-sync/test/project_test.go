/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	orgUnitApi "github.com/dsx-ai-factory/run-ai/runai-clients/org-unit-service-client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"

	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/test/fakecachedclient"
	"github.com/onsi/gomega/ghttp"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	backendclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers"
)

var _ = Describe("ProjectHandler", func() {
	var projectsPath = "/api/v1/org-unit/projects/status"

	var server *ghttp.Server
	var backendClient *backendclient.BackendClient
	var stopper chan struct{}
	var projectHandler *handlers.ProjectHandler
	var updatedProject chan string
	BeforeEach(func() {
		updatedProject = make(chan string, 1)
		DeferCleanup(func() { close(updatedProject) })
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
		testProject := BuildKaiProject("project-1", "1", "")

		BeforeEach(func() {

			GinkgoWriter.Printf("Created mock server calls...")
			// Update status mock
			server.RouteToHandler("PUT", projectsPath, func(writer http.ResponseWriter, request *http.Request) {
				body, err := io.ReadAll(request.Body)
				if err != nil {
					return
				}
				projectStatusRequest := &orgUnitApi.ProjectStatusUpdateRequest{}
				err = json.Unmarshal(body, projectStatusRequest)
				if err != nil {
					return
				}

				projectId := projectStatusRequest.Id
				ghttp.VerifyContentType("application/json")(writer, request)
				ghttp.RespondWith(http.StatusOK, "ok")(writer, request)
				updatedProject <- projectId
			})
		})

		JustBeforeEach(func() {
			GinkgoWriter.Println("Creating project controller")
			backendClient = backendclient.NewBackendClient(server.URL(), nil)
			fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme())
			fakeCachedClient := fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
			projectHandler = handlers.NewProjectHandler(backendClient, fakeCachedClient)
			stopper = make(chan struct{})
			go backendClient.StartSendingLivenessPeriodically(stopper)
		})

		It("Test add update", func() {
			err := projectHandler.HandleAddUpdate(testProject)
			Expect(err).To(BeNil())
			Eventually(updatedProject).WithTimeout(3 * time.Second).Should(Receive(Equal(testProject.Labels[parsers.ProjectIdLabel])))
		})
	})

})
