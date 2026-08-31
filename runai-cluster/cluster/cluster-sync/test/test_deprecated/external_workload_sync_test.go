/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test_deprecated

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/test"
	flds "github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	commonconstants "github.com/run-ai/runai/runai-cluster/common/constants"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	backendclient "github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/controllers"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/test/fakecachedclient"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	diwv2 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	infwv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	iwv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv2a1 "github.com/run-ai/runai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
)

const (
	projectIdStr1 = "1"
	projectIdStr2 = "2"
	depId         = "1"

	workloadId1           = "d5c5bc73-bb89-408b-8560-c63c5374cf90"
	workloadId2           = "abcd1234-bb89-408b-8560-c63c5374cf90"
	workloadIdAnnotation2 = "abcd5678-bb89-408b-8560-c63c5374cf90"
	trainWorkloadId1      = "44446666-bb89-408b-8560-c63c5374cf90"
	trainId1              = "44445555-bb89-408b-8560-c63c5374cf90"
	trainId2              = "11112222-bb89-408b-8560-c63c5374cf90"
	intId1                = "33334444-bb89-408b-8560-c63c5374cf90"
	intId2                = "55556666-bb89-408b-8560-c63c5374cf90"
	infId1                = "22223333-bb89-408b-8560-c63c5374cf90"
	infId2                = "33336666-bb89-408b-8560-c63c5374cf90"
	distId1               = "77778888-bb89-408b-8560-c63c5374cf90"
	distId2               = "99990000-bb89-408b-8560-c63c5374cf90"
	workloadName1         = "ew-deployment-depl1"
	workloadName2         = "ew-pod-pod2"
	trainName1            = "train1"
	trainName2            = "train2"
	intName1              = "int1"
	intName2              = "int2"
	infName1              = "inf1"
	infName2              = "inf2"
	distName1             = "dist1"
	distName2             = "dist2"

	waitTimeout  = 4 * time.Second
	waitInterval = 500 * time.Millisecond
)

var _ = Describe("ExternalWorkloadHandler", Ordered, func() {
	var syncPath = "/api/v1/workloads"
	var apiPathPrefix = fmt.Sprintf("%s/", syncPath)

	pathWithId, err := regexp.Compile(fmt.Sprintf("%s/.+", syncPath))
	Expect(err).To(BeNil())
	Expect(pathWithId.MatchString(fmt.Sprintf("%s/%s", syncPath, workloadId1))).To(BeTrue())

	var ctx context.Context
	var cancel context.CancelFunc
	var server *ghttp.Server
	var backendClient *backendclient.BackendClient
	var fakeCachedClient cachedclient.CachedClient

	var stopper chan struct{}
	var numberOfSyncedWorkloadsWithPodsInEachSync chan int
	var numberOfSyncedExternalWorkloadsInEachSync chan int
	var projectOfSyncedWorkloads chan string
	var numberOfUpToDateExternalWorkloadsInSync chan int
	var updatedExternalWorkloadId chan string
	var deletedExternalWorkloadId chan string
	var syncedWorkloads chan string
	var syncedWorkloadsPods chan map[string][]string
	var workloadsUpToDateInSync chan string

	BeforeEach(func() {
		numberOfSyncedExternalWorkloadsInEachSync = make(chan int, 50)
		DeferCleanup(func() { close(numberOfSyncedExternalWorkloadsInEachSync) })
		numberOfSyncedWorkloadsWithPodsInEachSync = make(chan int, 50)
		DeferCleanup(func() { close(numberOfSyncedWorkloadsWithPodsInEachSync) })
		projectOfSyncedWorkloads = make(chan string, 50)
		DeferCleanup(func() { close(projectOfSyncedWorkloads) })
		numberOfUpToDateExternalWorkloadsInSync = make(chan int, 50)
		DeferCleanup(func() { close(numberOfUpToDateExternalWorkloadsInSync) })
		updatedExternalWorkloadId = make(chan string, 50)
		DeferCleanup(func() { close(updatedExternalWorkloadId) })
		deletedExternalWorkloadId = make(chan string, 50)
		DeferCleanup(func() { close(deletedExternalWorkloadId) })
		syncedWorkloads = make(chan string, 50)
		DeferCleanup(func() { close(syncedWorkloads) })
		syncedWorkloadsPods = make(chan map[string][]string, 50)
		DeferCleanup(func() { close(syncedWorkloadsPods) })
		workloadsUpToDateInSync = make(chan string, 50)
		DeferCleanup(func() { close(workloadsUpToDateInSync) })

		GinkgoWriter.Println("Creating server for test...")
		server = test.CreateBasicRunaiBackend()
		DeferCleanup(func() {
			server.Close()
		})
		GinkgoWriter.Printf("Created server for test %s \n", server.URL())
		server.AllowUnhandledRequests = true
		stopper = make(chan struct{}, 1)
		DeferCleanup(func() { close(stopper) })
	})

	Context("runai backend is available", Ordered, func() {
		var existingExternalWorkload1, externalWorkloadForAdd1 *v1alpha1.ExternalWorkload
		var trainExisting, trainToAdd *twv2a1.TrainingWorkload
		var interactiveExisting, interactiveToAdd *iwv2a1.InteractiveWorkload
		var inferenceExisting, inferenceToAdd *infwv2a1.InferenceWorkload
		var distExisting, distToAdd *diwv2.DistributedWorkload
		var trainTemplateExisting, trainTemplateToAdd *twv2a1.TrainingWorkload
		var workloadHandler *handlers.WorkloadHandler
		var podHandler *handlers.PodHandler

		BeforeEach(func() {
			existingExternalWorkload1, _, _ = test.BuildExternalWorkload(workloadName1, namespaceForTest,
				workloadId1, v1alpha1.ExWDegradedPhase)
			externalWorkloadForAdd1, _, _ = test.BuildExternalWorkload(workloadName2, namespaceForTest,
				workloadId2, v1alpha1.ExWFailedPhase)
			externalWorkloadForAdd1.Annotations = map[string]string{
				commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload: workloadIdAnnotation2,
			}
			trainExisting, _ = test.BuildTrainingWorkloadBackendSourced(trainName1, namespaceForTest, trainId1, trainWorkloadId1)
			trainToAdd, _ = test.BuildTrainingWorkload(trainName2, namespaceForTest, trainId2)
			interactiveExisting, _ = test.BuildInteractiveWorkload(intName1, namespaceForTest, intId1)
			interactiveToAdd, _ = test.BuildInteractiveWorkload(intName2, namespaceForTest, intId2)
			inferenceExisting, _ = test.BuildInferenceWorkload(infName1, namespaceForTest, infId1)
			inferenceToAdd, _ = test.BuildInferenceWorkload(infName2, namespaceForTest, infId2)
			distExisting, _, _ = test.BuildDistributedWorkload(distName1, namespaceForTest, distId1)
			distToAdd, _, _ = test.BuildDistributedWorkload(distName2, namespaceForTest, distId2)

			trainTemplateExisting, _ = test.BuildTrainingWorkload("unfiltered-train-existing", namespaceForTest,
				"some-uidddd123")
			trainTemplateExisting.Spec.Usage = "Template"
			trainTemplateToAdd, _ = test.BuildTrainingWorkload("unfiltered-train-to-add", namespaceForTest,
				"some-uidddd456")
			trainTemplateToAdd.Spec.Usage = "Template"

			GinkgoWriter.Printf("Created mock server calls...")
			// Sync mock
			server.RouteToHandler("POST", syncPath, func(writer http.ResponseWriter, request *http.Request) {
				workloadsSyncRequest := &openapi.WorkloadsSyncRequest{}

				ghttp.VerifyContentType("application/json")(writer, request)
				err := json.NewDecoder(request.Body).Decode(workloadsSyncRequest)
				Expect(err).To(BeNil())
				ghttp.RespondWith(http.StatusNoContent, "ok")(writer, request)
				GinkgoWriter.Printf("Got a request in externalWorkloads POST Mock")
				numberOfSyncedExternalWorkloadsInEachSync <- len(workloadsSyncRequest.Workloads)
				numberOfUpToDateExternalWorkloadsInSync <- len(workloadsSyncRequest.IdsToIgnore)
				projectOfSyncedWorkloads <- workloadsSyncRequest.ProjectId

				syncedWorkloadUIDs := []string{}
				for _, w := range workloadsSyncRequest.Workloads {
					syncedWorkloadUIDs = append(syncedWorkloadUIDs, w.Id)
				}
				syncedWorkloads <- strings.Join(syncedWorkloadUIDs, ", ")

				workloadsUpToDateInSync <- strings.Join(workloadsSyncRequest.IdsToIgnore, ", ")
			})
			// Update status mock
			server.RouteToHandler("PUT", pathWithId, func(writer http.ResponseWriter, request *http.Request) {
				ghttp.VerifyContentType("application/json")(writer, request)
				ghttp.RespondWith(http.StatusOK, "ok")(writer, request)
				GinkgoWriter.Printf("Got a request in externalWorkloads PUT Mock")
				workloadId := strings.TrimPrefix(request.URL.Path, apiPathPrefix)
				updatedExternalWorkloadId <- workloadId
			})
			server.RouteToHandler("DELETE", pathWithId, func(writer http.ResponseWriter, request *http.Request) {
				ghttp.RespondWith(http.StatusOK, "ok")(writer, request)
				GinkgoWriter.Printf("Got a request in externalWorkloads DELETE Mock")
				workloadId := strings.TrimPrefix(request.URL.Path, apiPathPrefix)
				deletedExternalWorkloadId <- workloadId
			})

			ctx, cancel = context.WithCancel(context.TODO())
		})

		Context("sanity tests - do not sync pods with workloads", Ordered, func() {
			JustBeforeEach(func() {
				GinkgoWriter.Println("Starting ExternalWorkload informer")
				backendClient = backendclient.NewBackendClient(server.URL(), nil)
				ns := test.BuildNamespace(namespaceForTest, map[string]string{constants.RunaiQueueLabel: projectForTest})
				ns2 := test.BuildNamespace(namespaceForTest2, map[string]string{constants.RunaiQueueLabel: projectForTest2})
				project1 := test.BuildProjectWithDepId(projectForTest, projectIdStr1, namespaceForTest, depId)
				project2 := test.BuildProjectWithDepId(projectForTest2, projectIdStr2, namespaceForTest2, depId)
				kaiProject1 := test.BuildKaiProjectWithDepId(projectForTest, projectIdStr1, namespaceForTest, depId)
				kaiProject2 := test.BuildKaiProjectWithDepId(projectForTest2, projectIdStr2, namespaceForTest2, depId)
				fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(test.GetTestScheme()).WithObjects(
					ns, ns2, project1, project2, kaiProject1, kaiProject2, existingExternalWorkload1,
					trainExisting, interactiveExisting, inferenceExisting, distExisting, trainTemplateExisting).
					WithStatusSubresource(&twv2a1.TrainingWorkload{}, &iwv2a1.InteractiveWorkload{}, &infwv2a1.InferenceWorkload{}, &diwv2.DistributedWorkload{}, &v1alpha1.ExternalWorkload{})
				fakeCachedClient = fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
				podHandler = handlers.NewPodHandler(backendClient, fakeCachedClient)
				podRateLimiter := fakeCachedClient.NewRateLimitedController(cachedclient.PodRL, podHandler)
				workloadHandler = handlers.NewWorkloadHandler(backendClient, fakeCachedClient, podHandler, podRateLimiter)
				externalRLC := controllers.AddExternalWorkloadHandlerToCache(backendClient, fakeCachedClient, workloadHandler)
				trainingRLC := controllers.AddTrainingHandlerToCache(backendClient, fakeCachedClient, workloadHandler)
				workspaceRLC := controllers.AddWorkspaceHandlerToCache(backendClient, fakeCachedClient, workloadHandler)
				inferenceRLC := controllers.AddInferenceHandlerToCache(backendClient, fakeCachedClient, workloadHandler)
				distributedRLC := controllers.AddDistributedHandlerToCache(backendClient, fakeCachedClient, workloadHandler)
				distributedInfRLC := controllers.AddDistributedInferenceHandlerToCache(backendClient, fakeCachedClient, workloadHandler)
				controllers.AddPodGroupHandlerToCache(fakeCachedClient,
					trainingRLC, workspaceRLC, distributedRLC, inferenceRLC, externalRLC, distributedInfRLC)

				stopper = make(chan struct{})
				go backendClient.StartSendingLivenessPeriodically(stopper)
				fakeCachedClient.Start(ctx, stopper)
			})

			JustAfterEach(func() {
				cancel()
			})

			It("Test external workloads add + sync", func() {
				// with pagination - each request's last item will be the next request's first item
				// so 5 workloads with chunk size of 4 will be sent like this:
				// (0, 1, 2, 3),(3, 4)
				// we read from the channel 3 times (2+1) to verify the additional sync is empty
				numberOfSyncedEWPerProject := test.ReadFromChannel(numberOfSyncedExternalWorkloadsInEachSync, 3)
				// the syncs should have 4, 2 and 0 workloads
				Eventually(numberOfSyncedEWPerProject).Should(ContainElements(4, 2, 0))

				// again, reading 3 times from the channel, as described above
				projectsOfSyncedWorkloads := test.ReadFromChannel(projectOfSyncedWorkloads, 3)
				Eventually(projectsOfSyncedWorkloads).Should(ContainElements(projectIdStr1, projectIdStr2))

				err := fakeCachedClient.Create(ctx, externalWorkloadForAdd1)
				Expect(err).To(BeNil())
				Eventually(updatedExternalWorkloadId).WithTimeout(threeSecondTimeout).Should(
					Receive(Equal(workloadIdAnnotation2)))

				err = fakeCachedClient.Delete(ctx, externalWorkloadForAdd1)
				Expect(err).To(BeNil())
				Eventually(deletedExternalWorkloadId).WithTimeout(3 * time.Second).Should(
					Receive(Equal(workloadIdAnnotation2)))
			})
			It("Test external workloads update + sync", func() {
				Eventually(numberOfSyncedExternalWorkloadsInEachSync).Should(Receive(Equal(4)))

				existingEwToUpdate := &v1alpha1.ExternalWorkload{}
				err := fakeCachedClient.Get(ctx, types.NamespacedName{Name: existingExternalWorkload1.Name,
					Namespace: existingExternalWorkload1.Namespace}, existingEwToUpdate)
				Expect(err).To(BeNil())
				Expect(existingEwToUpdate.Status.SyncStatus.LastSyncStatusChangeTime).To(Equal(existingEwToUpdate.Status.ObservedStatus.LastStatusChangeTime))

				err = fakeCachedClient.Update(ctx, existingEwToUpdate)
				Expect(err).To(BeNil())
				// fake client doesn't update generation - object would be considered up to date
				Eventually(updatedExternalWorkloadId).WithTimeout(3 * time.Second).ShouldNot(Receive())

				existingEwToUpdate.Generation += 1
				err = fakeCachedClient.Update(ctx, existingEwToUpdate)
				Expect(err).To(BeNil())
				Eventually(updatedExternalWorkloadId).WithTimeout(3 * time.Second).Should(
					Receive(Equal(string(existingExternalWorkload1.Spec.OriginalWorkloadUID))))
			})
			It("Test external workloads delete + sync", func() {
				Eventually(numberOfSyncedExternalWorkloadsInEachSync).Should(Receive(Equal(4)))

				existingEwToDelete := &v1alpha1.ExternalWorkload{}
				err := fakeCachedClient.Get(ctx, types.NamespacedName{Name: existingExternalWorkload1.Name,
					Namespace: existingExternalWorkload1.Namespace}, existingEwToDelete)
				Expect(err).To(BeNil())

				err = fakeCachedClient.Delete(ctx, existingEwToDelete)
				Expect(err).To(BeNil())
				Eventually(deletedExternalWorkloadId).WithTimeout(3 * time.Second).Should(
					Receive(Equal(string(existingExternalWorkload1.Spec.OriginalWorkloadUID))))
			})
			It("Test runai workloads add", func() {
				Eventually(numberOfSyncedExternalWorkloadsInEachSync).Should(Receive(Equal(4)))

				// template shouldn't be synced to control-plane
				err = fakeCachedClient.Create(ctx, trainTemplateToAdd)
				Expect(err).To(BeNil())

				err := fakeCachedClient.Create(ctx, trainToAdd)
				Expect(err).To(BeNil())
				err = fakeCachedClient.Create(ctx, interactiveToAdd)
				Expect(err).To(BeNil())
				err = fakeCachedClient.Create(ctx, inferenceToAdd)
				Expect(err).To(BeNil())
				err = fakeCachedClient.Create(ctx, distToAdd)
				Expect(err).To(BeNil())

				addedWorkloadIds := test.ReadFromChannel(updatedExternalWorkloadId, 4)
				Eventually(addedWorkloadIds).Should(ContainElements(trainId2, intId2, infId2, distId2))
			})
			It("Test runai workloads update", func() {
				Eventually(numberOfSyncedExternalWorkloadsInEachSync).Should(Receive(Equal(4)))

				existingTwToUpdate := &twv2a1.TrainingWorkload{}
				Eventually(func(g Gomega) {
					err := fakeCachedClient.Get(ctx, types.NamespacedName{Name: trainExisting.Name,
						Namespace: trainExisting.Namespace}, existingTwToUpdate)
					g.Expect(err).To(BeNil())
					g.Expect(existingTwToUpdate.Status.SyncStatus).To(BeNil())
				}, waitTimeout, waitInterval).Should(Succeed())
				existingIwToUpdate := &iwv2a1.InteractiveWorkload{}
				Eventually(func(g Gomega) {
					err = fakeCachedClient.Get(ctx, types.NamespacedName{Name: interactiveExisting.Name,
						Namespace: interactiveExisting.Namespace}, existingIwToUpdate)
					g.Expect(err).To(BeNil())
					g.Expect(existingIwToUpdate.Status.SyncStatus).To(BeNil())
				}, waitTimeout, waitInterval).Should(Succeed())
				existingInfwToUpdate := &infwv2a1.InferenceWorkload{}
				Eventually(func(g Gomega) {
					err = fakeCachedClient.Get(ctx, types.NamespacedName{Name: inferenceExisting.Name,
						Namespace: inferenceExisting.Namespace}, existingInfwToUpdate)
					g.Expect(err).To(BeNil())
					g.Expect(existingInfwToUpdate.Status.SyncStatus).To(BeNil())
				}, waitTimeout, waitInterval).Should(Succeed())
				existingDwToUpdate := &diwv2.DistributedWorkload{}
				err = fakeCachedClient.Get(ctx, types.NamespacedName{Name: distExisting.Name,
					Namespace: distExisting.Namespace}, existingDwToUpdate)
				Expect(err).To(BeNil())

				existingTwToUpdate.Spec.NodePools = &flds.StringField{Value: ptr.To("np1")}
				existingIwToUpdate.Spec.Gpu = &flds.NumberField{Value: ptr.To("1")}
				existingInfwToUpdate.Spec.Gpu = &flds.NumberField{Value: ptr.To("1")}
				existingDwToUpdate.Spec.Cpu = &flds.NumberField{Value: ptr.To("5")}

				err = fakeCachedClient.Update(ctx, existingTwToUpdate)
				Expect(err).To(BeNil())
				err = fakeCachedClient.Update(ctx, existingIwToUpdate)
				Expect(err).To(BeNil())
				err = fakeCachedClient.Update(ctx, existingInfwToUpdate)
				Expect(err).To(BeNil())
				err = fakeCachedClient.Update(ctx, existingDwToUpdate)
				Expect(err).To(BeNil())

				updatedWorkloadIds := test.ReadFromChannel(updatedExternalWorkloadId, 4)
				Eventually(updatedWorkloadIds).Should(ContainElements(trainWorkloadId1, intId1, infId1, distId1))
			})
			It("Test runai workloads delete", func() {
				Eventually(numberOfSyncedExternalWorkloadsInEachSync).Should(Receive(Equal(4)))

				existingTwToDelete := &twv2a1.TrainingWorkload{}
				err := fakeCachedClient.Get(ctx, types.NamespacedName{Name: trainExisting.Name,
					Namespace: trainExisting.Namespace}, existingTwToDelete)
				Expect(err).To(BeNil())
				existingIwToDelete := &iwv2a1.InteractiveWorkload{}
				err = fakeCachedClient.Get(ctx, types.NamespacedName{Name: interactiveExisting.Name,
					Namespace: interactiveExisting.Namespace}, existingIwToDelete)
				Expect(err).To(BeNil())
				existingInfwToDelete := &infwv2a1.InferenceWorkload{}
				err = fakeCachedClient.Get(ctx, types.NamespacedName{Name: inferenceExisting.Name,
					Namespace: inferenceExisting.Namespace}, existingInfwToDelete)
				Expect(err).To(BeNil())
				existingDwToDelete := &diwv2.DistributedWorkload{}
				err = fakeCachedClient.Get(ctx, types.NamespacedName{Name: distExisting.Name,
					Namespace: distExisting.Namespace}, existingDwToDelete)
				Expect(err).To(BeNil())

				err = fakeCachedClient.Delete(ctx, existingTwToDelete)
				Expect(err).To(BeNil())
				err = fakeCachedClient.Delete(ctx, existingIwToDelete)
				Expect(err).To(BeNil())
				err = fakeCachedClient.Delete(ctx, existingInfwToDelete)
				Expect(err).To(BeNil())
				err = fakeCachedClient.Delete(ctx, existingDwToDelete)
				Expect(err).To(BeNil())

				deletedWorkloadIds := test.ReadFromChannel(deletedExternalWorkloadId, 4)
				Eventually(deletedWorkloadIds).Should(ContainElements(trainWorkloadId1, intId1, infId1, distId1))
			})
		})

		Context("deeper tests with more workloads - do not sync pods with workloads", Ordered, func() {
			var (
				workload0, workload1, workload2, workload3, workload4, workload5 *v1alpha1.ExternalWorkload
				train0, train1, train2                                           *twv2a1.TrainingWorkload
				int0, int1, int2                                                 *iwv2a1.InteractiveWorkload
				inf0, inf1, inf2                                                 *infwv2a1.InferenceWorkload
				dist0, dist1, dist2                                              *diwv2.DistributedWorkload

				uid0       = "10234"
				uid1       = "11234"
				uid2       = "12236"
				uid2_2     = "12237"
				uid3       = "13236"
				uid4       = "14236"
				uid5       = "15236"
				uid6       = "24567"
				uid7       = "bbc34"
				uid8       = "bcc34"
				uid8_8     = "bcc35"
				uid9       = "bcf34"
				uid10      = "cf534"
				uid11      = "df534"
				uid12      = "ef534"
				uid13      = "ff534"
				uid13_13   = "ff535"
				uid14      = "gf534"
				uid_inf0   = "if000"
				uid_inf1   = "if001"
				uid_inf2   = "if002"
				uid_inf2_a = "if002a"
			)
			JustBeforeEach(func() {
				workload0, _, _ = test.BuildExternalWorkload("ew-10", namespaceForTest, uid0, "")
				workload1, _, _ = test.BuildExternalWorkload("ew-11", namespaceForTest, uid1, "")
				train0, _ = test.BuildTrainingWorkloadBackendSourced("train-10", namespaceForTest, uid2, uid2_2)
				int0, _ = test.BuildInteractiveWorkload("int-10", namespaceForTest, uid3)
				workload2, _, _ = test.BuildExternalWorkload("ew-12", namespaceForTest, uid4, "")
				dist0, _, _ = test.BuildDistributedWorkload("dist-10", namespaceForTest, uid5)
				workload3, _, _ = test.BuildExternalWorkload("ew-13", namespaceForTest, uid6, "")
				dist2, _, _ = test.BuildDistributedWorkload("dist-12", namespaceForTest, uid7)
				workload4, _, _ = test.BuildExternalWorkload("ew-14", namespaceForTest, uid8, "")
				workload4.Annotations = map[string]string{commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload: uid8_8}
				workload5, _, _ = test.BuildExternalWorkload("ew-15", namespaceForTest, uid9, "")
				train1, _ = test.BuildTrainingWorkload("train-11", namespaceForTest, uid10)
				train2, _ = test.BuildTrainingWorkload("train-12", namespaceForTest, uid11)
				int1, _ = test.BuildInteractiveWorkload("int-11", namespaceForTest, uid12)
				int2, _ = test.BuildInteractiveWorkload("int-12", namespaceForTest, uid13)
				int2.Annotations = map[string]string{commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload: uid13_13}
				dist1, _, _ = test.BuildDistributedWorkload("dist-11", namespaceForTest, uid14)
				inf0, _ = test.BuildInferenceWorkload("inf-10", namespaceForTest, uid_inf0)
				inf1, _ = test.BuildInferenceWorkload("inf-11", namespaceForTest, uid_inf1)
				inf2, _ = test.BuildInferenceWorkload("inf-12", namespaceForTest, uid_inf2)
				inf2.Annotations = map[string]string{commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload: uid_inf2_a}

				GinkgoWriter.Println("Creating external workload controller")
				backendClient = backendclient.NewBackendClient(server.URL(), nil)
				ns := test.BuildNamespace(namespaceForTest, map[string]string{constants.RunaiQueueLabel: projectForTest})
				project1 := test.BuildProjectWithDepId(projectForTest, projectIdStr1, namespaceForTest, depId)
				kaiProject1 := test.BuildKaiProjectWithDepId(projectForTest, projectIdStr1, namespaceForTest, depId)
				fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(test.GetTestScheme()).WithObjects(ns, project1, kaiProject1,
					workload4, workload1, workload3, workload2, workload5, workload0,
					train0, train1, train2, int0, int1, int2, inf0, inf1, inf2, dist0, dist1, dist2).
					WithStatusSubresource(&twv2a1.TrainingWorkload{}, &iwv2a1.InteractiveWorkload{}, &infwv2a1.InferenceWorkload{}, &diwv2.DistributedWorkload{}, &v1alpha1.ExternalWorkload{})
				fakeCachedClient = fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
				podHandler = handlers.NewPodHandler(backendClient, fakeCachedClient)
				podRateLimiter := fakeCachedClient.NewRateLimitedController(cachedclient.PodRL, podHandler)
				workloadHandler = handlers.NewWorkloadHandler(backendClient, fakeCachedClient, podHandler, podRateLimiter)
				controllers.AddExternalWorkloadHandlerToCache(backendClient, fakeCachedClient, workloadHandler)
				controllers.AddTrainingHandlerToCache(backendClient, fakeCachedClient, workloadHandler)
				controllers.AddWorkspaceHandlerToCache(backendClient, fakeCachedClient, workloadHandler)
				controllers.AddInferenceHandlerToCache(backendClient, fakeCachedClient, workloadHandler)
				controllers.AddDistributedHandlerToCache(backendClient, fakeCachedClient, workloadHandler)

				stopper = make(chan struct{})
				go backendClient.StartSendingLivenessPeriodically(stopper)
				fakeCachedClient.Start(ctx, stopper)
			})

			JustAfterEach(func() {
				cancel()
			})

			It("Test sync in namespace with pagination", func() {
				// with pagination - each request's last item will be the next request's first item
				// so 18 workloads with chunk size of 4 will be sent like this:
				// (0, 1, 2, 3),(3, 4, 5, 6),(6, 7, 8, 9),(9, 10, 11, 12),(12, 13, 14, 15),(15,16,17)
				// these are 6 chunks so we read from the channel 6 times
				numberOfSyncedWorkloadsPerProject := test.ReadFromChannel(numberOfSyncedExternalWorkloadsInEachSync, 6)
				Eventually(numberOfSyncedWorkloadsPerProject).Should(HaveExactElements(4, 4, 4, 4, 4, 3))

				syncedWorkloadsUids := test.ReadFromChannel(syncedWorkloads, 6)
				Eventually(syncedWorkloadsUids).Should(HaveExactElements(
					strings.Join([]string{uid0, uid1, uid2_2, uid3}, ", "),
					strings.Join([]string{uid3, uid4, uid5, uid6}, ", "),
					strings.Join([]string{uid6, uid7, uid8_8, uid9}, ", "),
					strings.Join([]string{uid9, uid10, uid11, uid12}, ", "),
					strings.Join([]string{uid12, uid13_13, uid14, uid_inf0}, ", "),
					strings.Join([]string{uid_inf0, uid_inf1, uid_inf2_a}, ", "),
				))
				workloadsUpToDate := test.ReadFromChannel(workloadsUpToDateInSync, 6)
				Eventually(workloadsUpToDate).Should(HaveExactElements("", "", "", "", "", ""))

				// The next sync marks only external workloads as up-to-date. Legacy Run:ai
				// workloads are read-only in cluster-sync and are therefore sent again.
				// we allow a few seconds for the sync to happen, before performing the test
				time.Sleep(4 * time.Second)
				wg := sync.WaitGroup{}
				wg.Add(1)
				workloadHandler.HandleSyncEvent(&wg)
				wg.Wait()

				// External workloads are reported as up-to-date in the next request.
				workloadsUpToDate = test.ReadFromChannel(workloadsUpToDateInSync, 1)
				workloadIds := strings.Split(workloadsUpToDate[0], ", ")
				Expect(workloadIds).To(ConsistOf(uid0, uid1, uid4, uid6, uid8_8, uid9))
			})
		})
	})
})
