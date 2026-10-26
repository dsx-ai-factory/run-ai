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
	"regexp"
	"strings"
	"sync"
	"time"

	flds "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/workload_fields/field_types"
	kaiv2alpha2 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/scheduling/v2alpha2"
	coreV1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"

	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	backendclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/controllers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/test/fakecachedclient"
	dromosv1alpha1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/dromos/api/v1alpha1"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	dinfwv1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedinferenceworkload/v1alpha1"
	diwv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/distributedworkload/v2alpha1"
	infwv2a1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/inferenceworkload/v2alpha1"
	iwv2a1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/interactiveworkload/v2alpha1"
	twv2a1 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
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
	dinfId1               = "88889999-bb89-408b-8560-c63c5374cf90"
	dinfId2               = "00001111-bb89-408b-8560-c63c5374cf90"
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
	dinfName1             = "dinf1"
	dinfName2             = "dinf2"

	waitTimeout  = 4 * time.Second
	waitInterval = 500 * time.Millisecond
)

var _ = Describe("ExternalWorkloadHandler", Ordered, func() {
	var syncPath = "/api/v1/workloads"
	var apiPathPrefix = fmt.Sprintf("%s/", syncPath)

	var podSyncPath = "/api/v1/workloads/pods/sync"

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
	var syncedWorkloadEndpoints chan map[string][]openapi.EndpointSync

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
		syncedWorkloadEndpoints = make(chan map[string][]openapi.EndpointSync, 50)
		DeferCleanup(func() { close(syncedWorkloadEndpoints) })

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

	Context("runai backend is available - sync pods with workloads", Ordered, func() {
		var existingExternalWorkload1, externalWorkloadForAdd1 *v1alpha1.ExternalWorkload
		var externalExistingPod *coreV1.Pod
		var externalsExistingPodGroup *kaiv2alpha2.PodGroup
		var trainExisting, trainToAdd *twv2a1.TrainingWorkload
		var trainExistingPod *coreV1.Pod
		var interactiveExisting, interactiveToAdd *iwv2a1.InteractiveWorkload
		var interactiveExistingPod *coreV1.Pod
		var inferenceExisting, inferenceToAdd *infwv2a1.InferenceWorkload
		var inferenceExistingPod *coreV1.Pod
		var distExisting, distToAdd *diwv2.DistributedWorkload
		var distExistingPod1, distExistingPod2 *coreV1.Pod
		var dinfExisting, dinfToAdd *dinfwv1.DistributedInferenceWorkload
		var dinfExistingPod1, dinfExistingPod2 *coreV1.Pod
		var trainTemplateExisting, trainTemplateToAdd *twv2a1.TrainingWorkload
		workloadsToPods := map[string][]string{}

		BeforeEach(func() {
			existingExternalWorkload1, externalExistingPod, externalsExistingPodGroup = BuildExternalWorkload(workloadName1, namespaceForTest,
				workloadId1, v1alpha1.ExWDegradedPhase)
			existingExternalWorkload1.Status.NetworkStatus = v1alpha1.WorkloadNetworkStatus{
				ExternalEndpoints: []v1alpha1.ExternalEndpoint{
					{
						Host:       "test.example.com",
						Port:       "443",
						Path:       "/api",
						TlsEnabled: ptr.To(true),
						Protocol:   "HTTPS",
						Url:        "https://test.example.com:443/api",
						SourceRef: v1alpha1.ExternalEndpointSourceRef{
							Kind: "Ingress",
							Name: "test-ingress",
						},
					},
					{
						Host:     "1.2.1.1",
						Port:     "30000",
						Path:     "/",
						Protocol: "TCP",
						Url:      "1.2.1.1:30000",
						SourceRef: v1alpha1.ExternalEndpointSourceRef{
							Kind: "Service",
							Name: "test-np",
						},
					},
				},
			}
			externalWorkloadForAdd1, _, _ = BuildExternalWorkload(workloadName2, namespaceForTest,
				workloadId2, v1alpha1.ExWFailedPhase)
			externalWorkloadForAdd1.Annotations = map[string]string{
				commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload: workloadIdAnnotation2,
			}
			trainExisting, trainExistingPod = BuildTrainingWorkloadBackendSourced(trainName1, namespaceForTest, trainId1, trainWorkloadId1)
			trainToAdd, _ = BuildTrainingWorkload(trainName2, namespaceForTest, trainId2)
			interactiveExisting, interactiveExistingPod = BuildInteractiveWorkload(intName1, namespaceForTest, intId1)
			interactiveToAdd, _ = BuildInteractiveWorkload(intName2, namespaceForTest, intId2)
			inferenceExisting, inferenceExistingPod = BuildInferenceWorkload(infName1, namespaceForTest, infId1)
			inferenceToAdd, _ = BuildInferenceWorkload(infName2, namespaceForTest, infId2)
			distExisting, distExistingPod1, distExistingPod2 = BuildDistributedWorkload(distName1, namespaceForTest, distId1)
			distToAdd, _, _ = BuildDistributedWorkload(distName2, namespaceForTest, distId2)
			dinfExisting, dinfExistingPod1, dinfExistingPod2 = BuildDistributedInferenceWorkload(dinfName1, namespaceForTest, dinfId1)
			dinfToAdd, _, _ = BuildDistributedInferenceWorkload(dinfName2, namespaceForTest, dinfId2)

			trainTemplateExisting, _ = BuildTrainingWorkload("unfiltered-train-existing", namespaceForTest,
				"some-uidddd123")
			trainTemplateExisting.Spec.Usage = "Template"
			trainTemplateToAdd, _ = BuildTrainingWorkload("unfiltered-train-to-add", namespaceForTest,
				"some-uidddd456")
			trainTemplateToAdd.Spec.Usage = "Template"

			GinkgoWriter.Printf("Created mock server calls...")
			// Sync pods mock
			server.RouteToHandler("POST", podSyncPath, func(writer http.ResponseWriter, request *http.Request) {
				var workloadsSyncRequest []openapi.WorkloadPodsSyncRequest

				ghttp.VerifyContentType("application/json")(writer, request)
				err := json.NewDecoder(request.Body).Decode(&workloadsSyncRequest)
				Expect(err).To(BeNil())
				ghttp.RespondWith(http.StatusNoContent, "ok")(writer, request)
				GinkgoWriter.Printf("Got a request in externalWorkloads POST Mock")
				numberOfSyncedWorkloadsWithPodsInEachSync <- len(workloadsSyncRequest)

				for _, w := range workloadsSyncRequest {
					pods := []string{}
					for _, p := range w.Pods {
						pods = append(pods, p.Id)
					}
					workloadsToPods[w.WorkloadId] = pods
				}

			})
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
				endpointsMap := map[string][]openapi.EndpointSync{}
				for _, w := range workloadsSyncRequest.Workloads {
					syncedWorkloadUIDs = append(syncedWorkloadUIDs, w.Id)
					if len(w.Endpoints) > 0 {
						endpointsMap[w.Id] = w.Endpoints
					}
				}
				syncedWorkloads <- strings.Join(syncedWorkloadUIDs, ", ")
				syncedWorkloadEndpoints <- endpointsMap

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

		Context("sanity tests", Ordered, func() {
			JustBeforeEach(func() {
				GinkgoWriter.Println("Starting ExternalWorkload informer")
				backendClient = backendclient.NewBackendClient(server.URL(), nil)
				ns := BuildNamespace(namespaceForTest, map[string]string{constants.RunaiQueueLabel: projectForTest})
				ns2 := BuildNamespace(namespaceForTest2, map[string]string{constants.RunaiQueueLabel: projectForTest2})
				project1 := BuildProjectWithDepId(projectForTest, projectIdStr1, namespaceForTest, depId)
				project2 := BuildProjectWithDepId(projectForTest2, projectIdStr2, namespaceForTest2, depId)
				kaiProject1 := BuildKaiProjectWithDepId(projectForTest, projectIdStr1, namespaceForTest, depId)
				kaiProject2 := BuildKaiProjectWithDepId(projectForTest2, projectIdStr2, namespaceForTest2, depId)
				fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).WithObjects(
					ns, ns2, project1, project2, kaiProject1, kaiProject2, existingExternalWorkload1,
					trainExisting, interactiveExisting, inferenceExisting, distExisting, dinfExisting, trainTemplateExisting,
					externalExistingPod, trainExistingPod, interactiveExistingPod, inferenceExistingPod, distExistingPod1, distExistingPod2,
					dinfExistingPod1, dinfExistingPod2, externalsExistingPodGroup).
					WithStatusSubresource(&twv2a1.TrainingWorkload{}, &iwv2a1.InteractiveWorkload{}, &infwv2a1.InferenceWorkload{}, &diwv2.DistributedWorkload{}, &dinfwv1.DistributedInferenceWorkload{}, &v1alpha1.ExternalWorkload{})
				fakeCachedClient = fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
				podHandler := handlers.NewPodHandler(backendClient, fakeCachedClient)
				podRateLimiter := fakeCachedClient.NewRateLimitedController(cachedclient.PodRL, podHandler)
				workloadHandler := handlers.NewWorkloadHandler(backendClient, fakeCachedClient, podHandler, podRateLimiter)
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

			It("Test external workload endpoints are synced", func() {
				// wait for initial sync to complete
				numberOfSyncedEWPerProject := ReadFromChannel(numberOfSyncedExternalWorkloadsInEachSync, 2)
				Eventually(numberOfSyncedEWPerProject).Should(ContainElements(4, 3))

				// collect all endpoint maps from both sync chunks
				allEndpoints := map[string][]openapi.EndpointSync{}
				endpointMaps := ReadFromChannel(syncedWorkloadEndpoints, 2)
				Eventually(endpointMaps).Should(HaveLen(2))
				for _, m := range endpointMaps {
					for k, v := range m {
						allEndpoints[k] = v
					}
				}

				// workloadId1 should have endpoints
				endpoints, ok := allEndpoints[workloadId1]
				Expect(ok).To(BeTrue(), "workload %s should have endpoints in sync request", workloadId1)
				Expect(endpoints).To(HaveLen(2))
				Expect(*endpoints[0].Host).To(Equal("test.example.com"))
				Expect(*endpoints[0].Port).To(Equal(int32(443)))
				Expect(*endpoints[0].Path).To(Equal("/api"))
				Expect(endpoints[0].TlsEnabled.IsSet()).To(BeTrue())
				Expect(*endpoints[0].TlsEnabled.Get()).To(BeTrue())
				Expect(*endpoints[0].Protocol).To(Equal("HTTPS"))
				Expect(*endpoints[0].Url).To(Equal("https://test.example.com:443/api"))
				Expect(endpoints[0].SourceRef).NotTo(BeNil())
				Expect(*endpoints[0].SourceRef.Kind).To(Equal("Ingress"))
				Expect(*endpoints[0].SourceRef.Name).To(Equal("test-ingress"))
				Expect(*endpoints[1].Host).To(Equal("1.2.1.1"))
				Expect(*endpoints[1].Port).To(Equal(int32(30000)))
				Expect(*endpoints[1].Path).To(Equal("/"))
				Expect(endpoints[1].TlsEnabled.IsSet()).To(BeFalse())
				Expect(*endpoints[1].Protocol).To(Equal("TCP"))
				Expect(*endpoints[1].Url).To(Equal("1.2.1.1:30000"))
				Expect(endpoints[1].SourceRef).NotTo(BeNil())
				Expect(*endpoints[1].SourceRef.Kind).To(Equal("Service"))
				Expect(*endpoints[1].SourceRef.Name).To(Equal("test-np"))
			})
			It("Test external workloads add + sync", func() {
				// with pagination - each request's last item will be the next request's first item
				// so 6 workloads with chunk size of 4 will be sent like this:
				// (0, 1, 2, 3),(3, 4, 5)
				numberOfSyncedEWPerProject := ReadFromChannel(numberOfSyncedExternalWorkloadsInEachSync, 2)
				// the syncs should have 4, 3 workloads
				Eventually(numberOfSyncedEWPerProject).Should(ContainElements(4, 3))

				// pods sync
				// chunk size is 2 since 'chunkSize := int(math.Max(2, float64(h.maxNumOfPodsPerRequest/podsPerWorkloadMaxNumber)))'
				// so chunks include: in first workload chunk: (0,1),(1,2),(2,3),(3) in second workload chunk: (3,4),(4,5),(5)
				numberOfSyncedPodsPerProject := ReadFromChannel(numberOfSyncedWorkloadsWithPodsInEachSync, 7)
				Eventually(numberOfSyncedPodsPerProject).Should(ContainElements(2, 2, 2, 1, 2, 2, 1))

				// again, reading 3 times from the channel, as described above
				projectsOfSyncedWorkloads := ReadFromChannel(projectOfSyncedWorkloads, 2)
				Eventually(projectsOfSyncedWorkloads).Should(ContainElements(projectIdStr1))

				err := fakeCachedClient.Create(ctx, externalWorkloadForAdd1)
				Expect(err).To(BeNil())
				Eventually(updatedExternalWorkloadId).WithTimeout(threeSecondTimeout).Should(
					Receive(Equal(workloadIdAnnotation2)))

				err = fakeCachedClient.Delete(ctx, externalWorkloadForAdd1)
				Expect(err).To(BeNil())
				Eventually(deletedExternalWorkloadId).WithTimeout(3 * time.Second).Should(
					Receive(Equal(workloadIdAnnotation2)))

				eventuallyExpectPodGroupAnnotatedWithExternalWorkloadName(externalsExistingPodGroup, workloadName1, fakeCachedClient)
			})
			It("Test external workloads update + sync", func() {
				Eventually(numberOfSyncedExternalWorkloadsInEachSync).Should(Receive(Equal(4)))
				Eventually(numberOfSyncedWorkloadsWithPodsInEachSync).Should(Receive(Equal(2)))

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
				Eventually(numberOfSyncedWorkloadsWithPodsInEachSync).Should(Receive(Equal(2)))

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
				Eventually(numberOfSyncedWorkloadsWithPodsInEachSync).Should(Receive(Equal(2)))

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
				err = fakeCachedClient.Create(ctx, dinfToAdd)
				Expect(err).To(BeNil())

				addedWorkloadIds := ReadFromChannel(updatedExternalWorkloadId, 5)
				Eventually(addedWorkloadIds).Should(ContainElements(trainId2, intId2, infId2, distId2, dinfId2))
			})
			It("Test runai workloads update", func() {
				Eventually(numberOfSyncedExternalWorkloadsInEachSync).Should(Receive(Equal(4)))
				Eventually(numberOfSyncedWorkloadsWithPodsInEachSync).Should(Receive(Equal(2)))

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
				existingDinfwToUpdate := &dinfwv1.DistributedInferenceWorkload{}
				err = fakeCachedClient.Get(ctx, types.NamespacedName{Name: dinfExisting.Name,
					Namespace: dinfExisting.Namespace}, existingDinfwToUpdate)
				Expect(err).To(BeNil())

				existingTwToUpdate.Spec.NodePools = &flds.StringField{Value: ptr.To("np1")}
				existingIwToUpdate.Spec.Gpu = &flds.NumberField{Value: ptr.To("1")}
				existingInfwToUpdate.Spec.Gpu = &flds.NumberField{Value: ptr.To("1")}
				existingDwToUpdate.Spec.Cpu = &flds.NumberField{Value: ptr.To("5")}
				existingDinfwToUpdate.Spec.Replicas = &flds.IntField{Value: ptr.To(3)}

				err = fakeCachedClient.Update(ctx, existingTwToUpdate)
				Expect(err).To(BeNil())
				err = fakeCachedClient.Update(ctx, existingIwToUpdate)
				Expect(err).To(BeNil())
				err = fakeCachedClient.Update(ctx, existingInfwToUpdate)
				Expect(err).To(BeNil())
				err = fakeCachedClient.Update(ctx, existingDwToUpdate)
				Expect(err).To(BeNil())
				err = fakeCachedClient.Update(ctx, existingDinfwToUpdate)
				Expect(err).To(BeNil())

				updatedWorkloadIds := ReadFromChannel(updatedExternalWorkloadId, 5)
				Eventually(updatedWorkloadIds).Should(ContainElements(trainWorkloadId1, intId1, infId1, distId1, dinfId1))
			})
			It("Test runai workloads delete", func() {
				Eventually(numberOfSyncedExternalWorkloadsInEachSync).Should(Receive(Equal(4)))
				Eventually(numberOfSyncedWorkloadsWithPodsInEachSync).Should(Receive(Equal(2)))

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
				existingDinfwToDelete := &dinfwv1.DistributedInferenceWorkload{}
				err = fakeCachedClient.Get(ctx, types.NamespacedName{Name: dinfExisting.Name,
					Namespace: dinfExisting.Namespace}, existingDinfwToDelete)
				Expect(err).To(BeNil())

				err = fakeCachedClient.Delete(ctx, existingTwToDelete)
				Expect(err).To(BeNil())
				err = fakeCachedClient.Delete(ctx, existingIwToDelete)
				Expect(err).To(BeNil())
				err = fakeCachedClient.Delete(ctx, existingInfwToDelete)
				Expect(err).To(BeNil())
				err = fakeCachedClient.Delete(ctx, existingDwToDelete)
				Expect(err).To(BeNil())
				err = fakeCachedClient.Delete(ctx, existingDinfwToDelete)
				Expect(err).To(BeNil())

				deletedWorkloadIds := ReadFromChannel(deletedExternalWorkloadId, 5)
				Eventually(deletedWorkloadIds).Should(ContainElements(trainWorkloadId1, intId1, infId1, distId1, dinfId1))
			})
		})

		Context("deeper tests with more workloads", Ordered, func() {
			var (
				workload0, workload1, workload2, workload3, workload4, workload5                   *v1alpha1.ExternalWorkload
				workloadPod0, workloadPod1, workloadPod2, workloadPod3, workloadPod4, workloadPod5 *coreV1.Pod
				workloadPodGroup0, workloadPodGroup3                                               *kaiv2alpha2.PodGroup
				train0, train1, train2                                                             *twv2a1.TrainingWorkload
				trainPod0, trainPod1, trainPod2                                                    *coreV1.Pod
				int0, int1, int2                                                                   *iwv2a1.InteractiveWorkload
				intPod0, intPod1, intPod2                                                          *coreV1.Pod
				inf0, inf1, inf2                                                                   *infwv2a1.InferenceWorkload
				infPod0, infPod1, infPod2                                                          *coreV1.Pod
				dist0, dist1, dist2                                                                *diwv2.DistributedWorkload
				distPod01, distPod11, distPod21, distPod00, distPod10, distPod20                   *coreV1.Pod
				workloadHandler                                                                    *handlers.WorkloadHandler

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
				workload0, workloadPod0, workloadPodGroup0 = BuildExternalWorkload("ew-10", namespaceForTest, uid0, "")
				workload1, workloadPod1, _ = BuildExternalWorkload("ew-11", namespaceForTest, uid1, "")
				train0, trainPod0 = BuildTrainingWorkloadBackendSourced("train-10", namespaceForTest, uid2, uid2_2)
				int0, intPod0 = BuildInteractiveWorkload("int-10", namespaceForTest, uid3)
				workload2, workloadPod2, _ = BuildExternalWorkload("ew-12", namespaceForTest, uid4, "")
				dist0, distPod00, distPod01 = BuildDistributedWorkload("dist-10", namespaceForTest, uid5)
				workload3, workloadPod3, workloadPodGroup3 = BuildExternalWorkload("ew-13", namespaceForTest, uid6, "")
				dist2, distPod20, distPod21 = BuildDistributedWorkload("dist-12", namespaceForTest, uid7)
				workload4, workloadPod4, _ = BuildExternalWorkload("ew-14", namespaceForTest, uid8, "")
				workload4.Annotations = map[string]string{commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload: uid8_8}
				workloadPod4.Labels[commonconstants.TopOwnerUidLabel] = uid8_8
				workload5, workloadPod5, _ = BuildExternalWorkload("ew-15", namespaceForTest, uid9, "")
				train1, trainPod1 = BuildTrainingWorkload("train-11", namespaceForTest, uid10)
				train2, trainPod2 = BuildTrainingWorkload("train-12", namespaceForTest, uid11)
				int1, intPod1 = BuildInteractiveWorkload("int-11", namespaceForTest, uid12)
				int2, intPod2 = BuildInteractiveWorkload("int-12", namespaceForTest, uid13)
				int2.Annotations = map[string]string{commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload: uid13_13}
				intPod2.Labels[commonconstants.TopOwnerUidLabel] = uid13_13
				intPod2.Labels[commonconstants.TopOwnerWorkloadIdLabel] = uid13_13
				dist1, distPod10, distPod11 = BuildDistributedWorkload("dist-11", namespaceForTest, uid14)
				inf0, infPod0 = BuildInferenceWorkload("inf-10", namespaceForTest, uid_inf0)
				inf1, infPod1 = BuildInferenceWorkload("inf-11", namespaceForTest, uid_inf1)
				inf2, infPod2 = BuildInferenceWorkload("inf-12", namespaceForTest, uid_inf2)
				inf2.Annotations = map[string]string{commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload: uid_inf2_a}
				infPod2.Labels[commonconstants.TopOwnerUidLabel] = uid_inf2_a
				infPod2.Labels[commonconstants.TopOwnerWorkloadIdLabel] = uid_inf2_a

				GinkgoWriter.Println("Creating external workload controller")
				backendClient = backendclient.NewBackendClient(server.URL(), nil)
				ns := BuildNamespace(namespaceForTest, map[string]string{constants.RunaiQueueLabel: projectForTest})
				project1 := BuildProjectWithDepId(projectForTest, projectIdStr1, namespaceForTest, depId)
				kaiProject1 := BuildKaiProjectWithDepId(projectForTest, projectIdStr1, namespaceForTest, depId)
				fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).WithObjects(ns, project1, kaiProject1,
					workload4, workload1, workload3, workload2, workload5, workload0,
					workloadPod4, workloadPod1, workloadPod3, workloadPod2, workloadPod5, workloadPod0,
					train0, train1, train2, int0, int1, int2, inf0, inf1, inf2, dist0, dist1, dist2,
					trainPod0, trainPod1, trainPod2, intPod0, intPod1, intPod2, infPod0, infPod1, infPod2,
					distPod00, distPod10, distPod20, distPod01, distPod11, distPod21,
					workloadPodGroup0, workloadPodGroup3).
					WithStatusSubresource(&twv2a1.TrainingWorkload{}, &iwv2a1.InteractiveWorkload{}, &infwv2a1.InferenceWorkload{}, &diwv2.DistributedWorkload{}, &v1alpha1.ExternalWorkload{})
				fakeCachedClient = fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
				podHandler := handlers.NewPodHandler(backendClient, fakeCachedClient)
				podRateLimiter := fakeCachedClient.NewRateLimitedController(cachedclient.PodRL, podHandler)
				workloadHandler = handlers.NewWorkloadHandler(backendClient, fakeCachedClient, podHandler, podRateLimiter)
				controllers.AddExternalWorkloadHandlerToCache(backendClient, fakeCachedClient, workloadHandler)
				controllers.AddTrainingHandlerToCache(backendClient, fakeCachedClient, workloadHandler)
				controllers.AddWorkspaceHandlerToCache(backendClient, fakeCachedClient, workloadHandler)
				controllers.AddInferenceHandlerToCache(backendClient, fakeCachedClient, workloadHandler)
				controllers.AddDistributedHandlerToCache(backendClient, fakeCachedClient, workloadHandler)
				workloadsToPods = map[string][]string{}

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
				// (0, 1, 2, 3),(3, 4, 5, 6),(6, 7, 8, 9),(9, 10, 11, 12),(12, 13, 14, 15),(15, 16, 17)
				// these are 6 chunks so we read from the channel 6 times
				numberOfSyncedWorkloadsPerProject := ReadFromChannel(numberOfSyncedExternalWorkloadsInEachSync, 6)
				Eventually(numberOfSyncedWorkloadsPerProject).Should(HaveExactElements(4, 4, 4, 4, 4, 3))

				// Pods chunks should be:
				// first workload chunk: (0,1),(1,2),(2,3),(3)
				// second workload chunk: (3,4),(4,5),(5,6),(6)
				// ...
				// fifth workload chunk: (12,13),(13,14),(14,15),(15)
				// sixth workload chunk: (15,16),(16,17),(17)
				numberOfSyncedWorkloadsPodsPerProject := ReadFromChannel(numberOfSyncedWorkloadsWithPodsInEachSync, 23)
				Eventually(numberOfSyncedWorkloadsPodsPerProject).Should(HaveExactElements(2, 2, 2, 1, 2, 2, 2, 1, 2, 2, 2, 1, 2, 2, 2, 1, 2, 2, 2, 1, 2, 2, 1))

				syncedWorkloadsUids := ReadFromChannel(syncedWorkloads, 6)
				Eventually(syncedWorkloadsUids).Should(HaveExactElements(
					strings.Join([]string{uid0, uid1, uid2_2, uid3}, ", "),
					strings.Join([]string{uid3, uid4, uid5, uid6}, ", "),
					strings.Join([]string{uid6, uid7, uid8_8, uid9}, ", "),
					strings.Join([]string{uid9, uid10, uid11, uid12}, ", "),
					strings.Join([]string{uid12, uid13_13, uid14, uid_inf0}, ", "),
					strings.Join([]string{uid_inf0, uid_inf1, uid_inf2_a}, ", "),
				))
				workloadsUpToDate := ReadFromChannel(workloadsUpToDateInSync, 6)
				Eventually(workloadsUpToDate).Should(HaveExactElements("", "", "", "", "", ""))

				expectedWorkloadOfPods := map[string][]string{
					uid0:       {string(workloadPod0.UID)},
					uid1:       {string(workloadPod1.UID)},
					uid2_2:     {string(trainPod0.UID)},
					uid3:       {string(intPod0.UID)},
					uid4:       {string(workloadPod2.UID)},
					uid5:       {string(distPod00.UID), string(distPod01.UID)},
					uid6:       {string(workloadPod3.UID)},
					uid7:       {string(distPod20.UID), string(distPod21.UID)},
					uid8_8:     {string(workloadPod4.UID)},
					uid9:       {string(workloadPod5.UID)},
					uid10:      {string(trainPod1.UID)},
					uid11:      {string(trainPod2.UID)},
					uid12:      {string(intPod1.UID)},
					uid13_13:   {string(intPod2.UID)},
					uid14:      {string(distPod10.UID), string(distPod11.UID)},
					uid_inf0:   {string(infPod0.UID)},
					uid_inf1:   {string(infPod1.UID)},
					uid_inf2_a: {string(infPod2.UID)},
				}
				Expect(workloadsToPods).To(HaveLen(len(expectedWorkloadOfPods)))

				for workloadId, podIds := range workloadsToPods {
					expectedPodIds, ok := expectedWorkloadOfPods[workloadId]
					Expect(ok).To(BeTrue(), "workload %s not found in expected map", workloadId)

					Expect(podIds).To(HaveExactElements(expectedPodIds), "workload %s pods mismatch", workloadId)
				}

				// The next sync marks only external workloads as up-to-date. Legacy Run:ai
				// workloads are read-only in cluster-sync and are therefore sent again.
				// we allow a few seconds for the sync to happen, before performing the test
				time.Sleep(4 * time.Second)
				wg := sync.WaitGroup{}
				wg.Add(1)
				workloadHandler.HandleSyncEvent(&wg)
				wg.Wait()

				// External workloads are reported as up-to-date in their respective pages.
				workloadsUpToDate = ReadFromChannel(workloadsUpToDateInSync, 6)
				Expect(workloadsUpToDate).To(HaveExactElements(
					strings.Join([]string{uid0, uid1}, ", "),
					strings.Join([]string{uid4, uid6}, ", "),
					strings.Join([]string{uid6, uid8_8, uid9}, ", "),
					uid9,
					"",
					"",
				))

				eventuallyExpectPodGroupAnnotatedWithExternalWorkloadName(workloadPodGroup0, "ew-10", fakeCachedClient)
				eventuallyExpectPodGroupAnnotatedWithExternalWorkloadName(workloadPodGroup3, "ew-13", fakeCachedClient)
			})
		})
	})

	Context("queued update of an external workload", func() {
		var externalWorkload *v1alpha1.ExternalWorkload
		var workloadRunner *dromosv1alpha1.WorkloadRunner
		var externalWorkloadHandler *handlers.ExternalWorkloadHandler
		key := types.NamespacedName{Name: workloadName1, Namespace: namespaceForTest}

		BeforeEach(func() {
			externalWorkload, _, _ = BuildExternalWorkload(workloadName1, namespaceForTest, workloadId1, v1alpha1.ExWRunningPhase)
			externalWorkload.Labels = map[string]string{
				commonconstants.WorkloadRunnerNameLabel: workloadName1,
				commonconstants.WorkloadRunnerUidLabel:  "workload-runner-uid",
			}
			workloadRunner = &dromosv1alpha1.WorkloadRunner{
				ObjectMeta: metav1.ObjectMeta{Name: workloadName1, Namespace: namespaceForTest, UID: "workload-runner-uid"},
			}

			server.RouteToHandler("PUT", pathWithId, func(writer http.ResponseWriter, request *http.Request) {
				updatedExternalWorkloadId <- strings.TrimPrefix(request.URL.Path, apiPathPrefix)
				ghttp.RespondWith(http.StatusOK, "ok")(writer, request)
			})
		})

		JustBeforeEach(func() {
			backendClient = backendclient.NewBackendClient(server.URL(), nil)
			ns := BuildNamespace(namespaceForTest, map[string]string{constants.RunaiQueueLabel: projectForTest})
			project1 := BuildProjectWithDepId(projectForTest, projectIdStr1, namespaceForTest, depId)
			kaiProject1 := BuildKaiProjectWithDepId(projectForTest, projectIdStr1, namespaceForTest, depId)
			fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).
				WithObjects(ns, project1, kaiProject1, externalWorkload)
			if workloadRunner != nil {
				fakeClientBuilder = fakeClientBuilder.WithObjects(workloadRunner)
			}
			fakeCachedClient = fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
			podHandler := handlers.NewPodHandler(backendClient, fakeCachedClient)
			podRateLimiter := fakeCachedClient.NewRateLimitedController(cachedclient.PodRL, podHandler)
			workloadHandler := handlers.NewWorkloadHandler(backendClient, fakeCachedClient, podHandler, podRateLimiter)
			externalWorkloadHandler = handlers.NewExternalWorkloadHandler(workloadHandler)
		})

		It("upserts the workload when its WorkloadRunner exists", func() {
			Expect(externalWorkloadHandler.HandleUpdateItem(key)).To(Succeed())

			Expect(updatedExternalWorkloadId).To(Receive(Equal("d5c5bc73-bb89-408b-8560-c63c5374cf90")))
		})

		When("its WorkloadRunner was deleted", func() {
			BeforeEach(func() {
				workloadRunner = nil
			})

			It("does not upsert the workload", func() {
				Expect(externalWorkloadHandler.HandleUpdateItem(key)).To(Succeed())

				Expect(updatedExternalWorkloadId).NotTo(Receive())
			})
		})

		When("its WorkloadRunner was replaced by another one with the same name", func() {
			BeforeEach(func() {
				workloadRunner.UID = "replacement-workload-runner-uid"
			})

			It("does not upsert the workload", func() {
				Expect(externalWorkloadHandler.HandleUpdateItem(key)).To(Succeed())

				Expect(updatedExternalWorkloadId).NotTo(Receive())
			})
		})

		When("it has no WorkloadRunner UID label", func() {
			BeforeEach(func() {
				delete(externalWorkload.Labels, commonconstants.WorkloadRunnerUidLabel)
				workloadRunner.UID = "replacement-workload-runner-uid"
			})

			It("upserts the workload while its WorkloadRunner exists", func() {
				Expect(externalWorkloadHandler.HandleUpdateItem(key)).To(Succeed())

				Expect(updatedExternalWorkloadId).To(Receive(Equal("d5c5bc73-bb89-408b-8560-c63c5374cf90")))
			})
		})

		When("it has no WorkloadRunner", func() {
			BeforeEach(func() {
				externalWorkload.Labels = nil
				workloadRunner = nil
			})

			It("upserts the workload", func() {
				Expect(externalWorkloadHandler.HandleUpdateItem(key)).To(Succeed())

				Expect(updatedExternalWorkloadId).To(Receive(Equal("d5c5bc73-bb89-408b-8560-c63c5374cf90")))
			})
		})
	})
})

func eventuallyExpectPodGroupAnnotatedWithExternalWorkloadName(pg *kaiv2alpha2.PodGroup, expectedExternalWorkloadName string,
	fakeCachedClient cachedclient.CachedClient) {
	EventuallyWithOffset(1, func(g Gomega) {
		apiPg := &kaiv2alpha2.PodGroup{}
		err := fakeCachedClient.Get(context.Background(), types.NamespacedName{Name: pg.Name, Namespace: pg.Namespace}, apiPg)
		g.Expect(err).To(BeNil())

		externalWorkloadName, found := apiPg.Annotations[commonconstants.ExternalWorkloadName]
		g.Expect(found).To(BeTrue())
		g.Expect(externalWorkloadName).To(Equal(expectedExternalWorkloadName), "ExternalWorkloadName annotation on podgroup")

		// validate original labels were not deleted
		g.Expect(apiPg.Labels["some-key1"]).To(Equal("some-value1"), "Original label some-key1")
		g.Expect(apiPg.Labels["some-key2"]).To(Equal("some-value2"), "Original label some-key2")
		g.Expect(apiPg.Annotations["some-key11"]).To(Equal("some-value11"), "Original annotation some-key11")
		g.Expect(apiPg.Annotations["some-key22"]).To(Equal("some-value22"), "Original annotation some-key22")
	}, time.Second*5, time.Millisecond*500).Should(Succeed())
}
