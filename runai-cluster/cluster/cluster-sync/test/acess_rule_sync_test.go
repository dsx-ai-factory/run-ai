/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test

//TODO: uncomment
/*
var _ = Describe("AccessRuleHandler", func() {
	var accessRulesSyncPath = fmt.Sprintf("/api/v1/authorization/access-rules/%s/sync", ClusterIdForTests)

	accessRulesStatusPathWithId, err := regexp.Compile(fmt.Sprintf("/api/v1/authorization/access-rules/.+/%s/status", ClusterIdForTests))
	Expect(err).To(BeNil())
	Expect(accessRulesStatusPathWithId.MatchString(fmt.Sprintf("/api/v1/authorization/access-rules/31/%s/status", ClusterIdForTests))).To(BeTrue())

	var ctx context.Context
	var cancel context.CancelFunc
	var server *ghttp.Server
	var backendClient *backendclient.BackendClient
	var c client.WithWatch

	var stopper chan struct{}
	var accessRuleController *controllers.AccessRuleController
	var numberOfsyncedAccessRules chan int
	var statusUpdatedAccessRuleId chan int32

	BeforeEach(func() {
		numberOfsyncedAccessRules = make(chan int, 1)
		DeferCleanup(func() { close(numberOfsyncedAccessRules) })
		statusUpdatedAccessRuleId = make(chan int32, 1)
		DeferCleanup(func() { close(statusUpdatedAccessRuleId) })
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
		var existingUpdatedAccessRuleId int32 = 13434
		var existingAccessRule *v1alpha1.AccessRule

		var addedAccessRuleId int32 = 24343
		var accessRuleForAdd *v1alpha1.AccessRule

		BeforeEach(func() {
			existingAccessRule = BuildAccessRule(fmt.Sprintf("ar-%d-some-type", existingUpdatedAccessRuleId), existingUpdatedAccessRuleId, v1alpha1.AccessRuleReady, "")
			accessRuleForAdd = BuildAccessRule(fmt.Sprintf("ar-%d-some-scope", addedAccessRuleId), addedAccessRuleId, v1alpha1.AccessRuleNotReady, "failed")

			GinkgoWriter.Printf("Created mock server calls...")
			// Sync mock
			server.RouteToHandler("POST", accessRulesSyncPath, func(writer http.ResponseWriter, request *http.Request) {
				var accessRules []swagger.AccessRuleForSync
				ghttp.VerifyContentType("application/json")(writer, request)
				err := json.NewDecoder(request.Body).Decode(&accessRules)
				Expect(err).To(BeNil())
				ghttp.RespondWith(http.StatusOK, "ok")
				numberOfsyncedAccessRules <- len(accessRules)
			})
			// Update status mock
			server.RouteToHandler("PUT", accessRulesStatusPathWithId, func(writer http.ResponseWriter, request *http.Request) {
				ghttp.VerifyContentType("application/json")(writer, request)
				ghttp.RespondWith(http.StatusOK, "ok")(writer, request)
				accessRuleIdStr := strings.TrimPrefix(request.URL.Path, "/api/v1/authorization/access-rules/")
				accessRuleIdStr = strings.TrimSuffix(accessRuleIdStr, fmt.Sprintf("/%s/status", ClusterIdForTests))
				accessRuleId, _ := strconv.ParseInt(accessRuleIdStr, 10, 32)
				statusUpdatedAccessRuleId <- int32(accessRuleId)
			})

			ctx, cancel = context.WithCancel(context.TODO())
		})

		JustBeforeEach(func() {
			GinkgoWriter.Println("Creating access rule controller")
			backendClient = backendclient.NewBackendClient(server.URL(), nil)
			c = fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).WithObjects(existingAccessRule).Build()
			accessRuleController = controllers.NewAccessRuleController(c, backendClient)

			stopper = make(chan struct{})
			go backendClient.StartSendingLivenessPeriodically(stopper)
			go accessRuleController.Start(stopper)
		})

		JustAfterEach(func() {
			cancel()
		})

		It("Test accessrule add + sync", func() {
			Eventually(numberOfsyncedAccessRules).Should(Receive(Equal(1)))
			err := c.Create(ctx, accessRuleForAdd)
			Expect(err).To(BeNil())
			Eventually(statusUpdatedAccessRuleId).WithTimeout(3 * time.Second).Should(Receive(Equal(addedAccessRuleId)))
		})
		It("Test accessrule update + sync", func() {
			Eventually(numberOfsyncedAccessRules).Should(Receive(Equal(1)))
			existingAccessRule.Spec.ScopeName = "some-scope"
			err := c.Update(ctx, existingAccessRule)
			Expect(err).To(BeNil())
			Eventually(statusUpdatedAccessRuleId).WithTimeout(3 * time.Second).Should(Receive(Equal(existingUpdatedAccessRuleId)))
		})
		It("Test accessrule delete + sync", func() {
			Eventually(numberOfsyncedAccessRules).Should(Receive(Equal(1)))
			err := c.Delete(ctx, existingAccessRule)
			Expect(err).To(BeNil())
			Eventually(statusUpdatedAccessRuleId).WithTimeout(3 * time.Second).Should(Receive(Equal(existingUpdatedAccessRuleId)))
		})
	})

})
*/
