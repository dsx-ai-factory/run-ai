/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package test_deprecated

import (
	"context"
	"encoding/json"
	"fmt"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/test"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/test/fakecachedclient"
	twv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"k8s.io/utils/ptr"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	backendclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/controllers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	podId1                      = "123"
	knativeWorkloadIdAnnotation = "some-knative-iddddd"

	namespaceForTest  = "namespace"
	namespaceForTest2 = "runai-project2"
	projectForTest    = "project"
	projectForTest2   = "project2"

	threeSecondTimeout = 3 * time.Second //nolint:staticcheck
)

// This test suite is a suite that tests a flow from change of pods state in k8s to request sent to backend

var _ = Describe("A Pod controller test", func() {
	var syncPath = "/api/v1/workloads/pods"

	pathWithIds, err := regexp.Compile("/api/v1/workloads/pods/([0-9a-zA-Z-].*)")
	Expect(err).To(BeNil())
	matches := pathWithIds.FindStringSubmatch(fmt.Sprintf("/api/v1/workloads/pods/%s", podId1))
	Expect(len(matches)).To(Equal(2))
	Expect(pathWithIds.MatchString(fmt.Sprintf("/api/v1/workloads/pods/%s", podId1))).To(BeTrue())

	var workloadsSyncPath = "/api/v1/workloads"
	var workloadsApiPathPrefix = fmt.Sprintf("%s/", workloadsSyncPath)
	workloadPathWithId, err := regexp.Compile(fmt.Sprintf("%s/.+", workloadsSyncPath))
	Expect(err).To(BeNil())
	Expect(workloadPathWithId.MatchString(fmt.Sprintf("%s/%s", syncPath, workloadId1))).To(BeTrue())

	var ctx context.Context
	var cancel context.CancelFunc
	var server *ghttp.Server
	var backendClient *backendclient.BackendClient
	var fakeCachedClient cachedclient.CachedClient

	var stopper chan struct{}
	var numberOfSyncedPodsInEachSync chan int
	var projectOfSyncedPods chan string
	var projectOfFailedSyncedPods chan string
	var updatedPodId chan string
	var failedUpdatePodId chan string
	var updatedPodsWorkloadId chan string
	var deletedPodId chan string
	var syncedPods chan string
	var updatedWorkloadId chan string

	var podIdsToReturn412 = map[string]bool{}
	var projectsToReturn412 = map[string]bool{}

	runaiManagedNamespace := test.BuildNamespace(namespaceForTest, map[string]string{constants.RunaiQueueLabel: projectForTest})
	runaiManagedNamespace2 := test.BuildNamespace(namespaceForTest2, map[string]string{constants.RunaiQueueLabel: projectForTest2})
	project1 := test.BuildProjectWithDepId(projectForTest, projectIdStr1, namespaceForTest, depId)
	project2 := test.BuildProjectWithDepId(projectForTest2, projectIdStr2, namespaceForTest2, depId)
	kaiProject1 := test.BuildKaiProjectWithDepId(projectForTest, projectIdStr1, namespaceForTest, depId)
	kaiProject2 := test.BuildKaiProjectWithDepId(projectForTest2, projectIdStr2, namespaceForTest2, depId)
	interactiveExisting, _ := test.BuildInteractiveWorkload(intName1, namespaceForTest, intId1)
	trainExisting, _ := test.BuildTrainingWorkloadBackendSourced(trainName1, namespaceForTest2, trainId1, trainWorkloadId1)
	existingService := test.BuildService("srv-3", namespaceForTest2, "srv-3-uid")
	existingService.Annotations = map[string]string{commonconstants.ControlPlaneWorkloadIdAnnotationForWorkload: knativeWorkloadIdAnnotation}

	BeforeEach(func() {
		numberOfSyncedPodsInEachSync = make(chan int, 12)
		DeferCleanup(func() { close(numberOfSyncedPodsInEachSync) })
		projectOfSyncedPods = make(chan string, 12)
		DeferCleanup(func() { close(projectOfSyncedPods) })
		projectOfFailedSyncedPods = make(chan string, 12)
		DeferCleanup(func() { close(projectOfFailedSyncedPods) })
		deletedPodId = make(chan string, 12)
		DeferCleanup(func() { close(deletedPodId) })
		updatedPodId = make(chan string, 12)
		DeferCleanup(func() { close(updatedPodId) })
		failedUpdatePodId = make(chan string, 12)
		DeferCleanup(func() { close(failedUpdatePodId) })
		updatedPodsWorkloadId = make(chan string, 12)
		DeferCleanup(func() { close(updatedPodsWorkloadId) })
		syncedPods = make(chan string, 12)
		DeferCleanup(func() { close(syncedPods) })
		updatedWorkloadId = make(chan string, 4)
		DeferCleanup(func() { close(updatedWorkloadId) })

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

	Context("runai backend is alive", func() {
		var existingPod *corev1.Pod
		var podToAdd0 *corev1.Pod
		var podToAdd1 *corev1.Pod
		var podToAdd2 *corev1.Pod
		var podToAdd3 *corev1.Pod
		var podHandler *handlers.PodHandler

		BeforeEach(func() {
			existingPod = test.BuildPod(namespaceForTest, "p", "imageName", "node1", corev1.PodRunning,
				test.BuildResourceList("1000m", "1G"), []metav1.OwnerReference{{APIVersion: "run.ai/v2alpha1",
					Kind: "InteractiveWorkload", Name: intName1, UID: intId1, Controller: nil, BlockOwnerDeletion: nil,
				}}, make(map[string]string))
			podToAdd0 = test.BuildPod(namespaceForTest2, "pod0", "imageName", "node1", corev1.PodPending,
				test.BuildResourceList("1000m", "1G"), []metav1.OwnerReference{{APIVersion: "run.ai/v2alpha1",
					Kind: "TrainingWorkload", Name: trainName1, UID: trainId1, Controller: nil, BlockOwnerDeletion: nil,
				}}, map[string]string{v1alpha1.NodePoolNameLabel: "some-node-pool"})
			podToAdd1 = test.BuildPod(namespaceForTest2, "pod1", "imageName", "node1", corev1.PodRunning,
				test.BuildResourceList("1000m", "1G"), []metav1.OwnerReference{}, map[string]string{v1alpha1.NodePoolNameLabel: "some-node-pool"})
			podToAdd2 = test.BuildPod(namespaceForTest2, "pod2", "imageName", "node1", corev1.PodRunning,
				test.BuildResourceList("1000m", "1G"), []metav1.OwnerReference{{APIVersion: "serving.knative.dev/v1",
					Kind: "Service", Name: "srv-3", UID: "srv-3-uid", Controller: nil, BlockOwnerDeletion: nil,
				}}, map[string]string{})
			podToAdd3 = test.BuildPod(namespaceForTest, "pod3", "imageName", "node1", corev1.PodRunning,
				test.BuildResourceList("1000m", "1G"), []metav1.OwnerReference{{APIVersion: "run.ai/v2alpha1",
					Kind: "InteractiveWorkload", Name: intName1, UID: intId1, Controller: nil, BlockOwnerDeletion: nil,
				}}, make(map[string]string))

			GinkgoWriter.Printf("Created mock server calls...")
			server.RouteToHandler("POST", syncPath, func(writer http.ResponseWriter, request *http.Request) {
				ghttp.VerifyContentType("application/json")(writer, request)
				GinkgoWriter.Printf("Got a request in pods POST Mock")

				podSyncRequest := &openapi.PodsSyncRequest{}
				err := json.NewDecoder(request.Body).Decode(podSyncRequest)
				Expect(err).To(BeNil())

				if _, found := projectsToReturn412[podSyncRequest.ProjectId]; found {
					ghttp.RespondWith(http.StatusPreconditionFailed, "")(writer, request)
					projectOfFailedSyncedPods <- podSyncRequest.ProjectId
					return
				}
				ghttp.RespondWith(http.StatusNoContent, "ok")(writer, request)

				numberOfSyncedPodsInEachSync <- len(podSyncRequest.Pods)
				projectOfSyncedPods <- podSyncRequest.ProjectId

				syncedPodUIDs := []string{}
				for _, w := range podSyncRequest.Pods {
					syncedPodUIDs = append(syncedPodUIDs, w.Id)
				}
				syncedPods <- strings.Join(syncedPodUIDs, ", ")
			})
			server.RouteToHandler("PUT", pathWithIds, func(writer http.ResponseWriter, request *http.Request) {
				ghttp.VerifyContentType("application/json")(writer, request)
				GinkgoWriter.Printf("Got a request in pods PUT Mock")

				// extract ids from request.URL.Path
				matches := pathWithIds.FindStringSubmatch(request.URL.Path)
				Expect(len(matches)).To(Equal(2))
				podId := matches[1]

				if _, found := podIdsToReturn412[podId]; found {
					ghttp.RespondWith(http.StatusPreconditionFailed, "")(writer, request)
					failedUpdatePodId <- podId
					return
				}
				ghttp.RespondWith(http.StatusOK, "ok")(writer, request)

				pod := &openapi.PodSyncFields{}
				err := json.NewDecoder(request.Body).Decode(pod)
				Expect(err).To(BeNil())

				updatedPodId <- podId
				updatedPodsWorkloadId <- pod.WorkloadId
			})
			server.RouteToHandler("DELETE", pathWithIds, func(writer http.ResponseWriter, request *http.Request) {
				ghttp.RespondWith(http.StatusOK, "ok")(writer, request)
				GinkgoWriter.Printf("Got a request in pods DELETE Mock")

				// extract ids from request.URL.Path
				matches := pathWithIds.FindStringSubmatch(request.URL.Path)
				Expect(len(matches)).To(Equal(2))
				podId := matches[1]
				deletedPodId <- podId
			})

			ctx, cancel = context.WithCancel(context.TODO())
		})

		Context("sanity tests", func() {
			JustBeforeEach(func() {
				backendClient = backendclient.NewBackendClient(server.URL(), nil)
				fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(test.GetTestScheme()).WithObjects(
					runaiManagedNamespace, runaiManagedNamespace2, project1, project2, kaiProject1, kaiProject2,
					trainExisting, interactiveExisting, existingService, existingPod)
				fakeCachedClient = fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
				nodeHandler := handlers.NewNodeHandler(backendClient, fakeCachedClient)
				podHandler = handlers.NewPodHandler(backendClient, fakeCachedClient)
				controllers.AddPodHandlerToCache(backendClient, fakeCachedClient, podHandler, nodeHandler)

				stopper = make(chan struct{})
				go backendClient.StartSendingLivenessPeriodically(stopper)
				fakeCachedClient.Start(ctx, stopper)
			})

			JustAfterEach(func() {
				cancel()
			})

			It("Tests pod add+sync", func() {
				numberOfSyncedPodsPerNamespace := test.ReadFromChannel(numberOfSyncedPodsInEachSync, 2)
				Eventually(numberOfSyncedPodsPerNamespace).Should(ContainElements(0, 1))
				projectsOfSyncedPods := test.ReadFromChannel(projectOfSyncedPods, 2)
				Eventually(projectsOfSyncedPods).Should(ContainElements(projectIdStr1, projectIdStr2))

				err := fakeCachedClient.Create(context.TODO(), podToAdd0)
				Expect(err).To(BeNil())
				Eventually(updatedPodId).WithTimeout(threeSecondTimeout).Should(Receive(Equal(string(podToAdd0.UID))))
				Eventually(updatedPodsWorkloadId).WithTimeout(threeSecondTimeout).Should(
					Receive(Equal(trainWorkloadId1)))

				err = fakeCachedClient.Create(context.TODO(), podToAdd1)
				Expect(err).To(BeNil())
				Eventually(updatedPodId).WithTimeout(threeSecondTimeout).Should(Receive(Equal(string(podToAdd1.UID))))
				Eventually(updatedPodsWorkloadId).WithTimeout(threeSecondTimeout).Should(
					Receive(Equal(string(podToAdd1.UID))))

				err = fakeCachedClient.Create(context.TODO(), podToAdd2)
				Expect(err).To(BeNil())
				Eventually(updatedPodId).WithTimeout(threeSecondTimeout).Should(Receive(Equal(string(podToAdd2.UID))))
				Eventually(updatedPodsWorkloadId).WithTimeout(threeSecondTimeout).Should(
					Receive(Equal(knativeWorkloadIdAnnotation)))

				err = fakeCachedClient.Create(context.TODO(), podToAdd3)
				Expect(err).To(BeNil())
				Eventually(updatedPodId).WithTimeout(threeSecondTimeout).Should(Receive(Equal(string(podToAdd3.UID))))
				Eventually(updatedPodsWorkloadId).WithTimeout(threeSecondTimeout).Should(
					Receive(Equal(intId1)))
			})
			It("Test pods update + sync", func() {
				Eventually(numberOfSyncedPodsInEachSync).Should(Receive(Equal(1)))

				podToUpdate := &corev1.Pod{}
				err := fakeCachedClient.Get(ctx, client.ObjectKey{Namespace: existingPod.Namespace, Name: existingPod.Name},
					podToUpdate)
				Expect(err).To(BeNil())

				podToUpdate.Spec.Priority = ptr.To(int32(1))
				err = fakeCachedClient.Update(ctx, podToUpdate)
				Expect(err).To(BeNil())
				Eventually(updatedPodId).WithTimeout(threeSecondTimeout).Should(Receive(Equal(string(existingPod.UID))))
				Eventually(updatedPodsWorkloadId).WithTimeout(threeSecondTimeout).Should(
					Receive(Equal(string(existingPod.OwnerReferences[0].UID))))
			})
			It("Test pods delete + sync", func() {
				Eventually(numberOfSyncedPodsInEachSync).Should(Receive(Equal(1)))
				podToDelete := &corev1.Pod{}
				err := fakeCachedClient.Get(ctx, client.ObjectKey{Namespace: existingPod.Namespace, Name: existingPod.Name},
					podToDelete)
				Expect(err).To(BeNil())
				err = fakeCachedClient.Delete(ctx, podToDelete)
				Expect(err).To(BeNil())
				Eventually(deletedPodId).WithTimeout(threeSecondTimeout).Should(Receive(Equal(string(existingPod.UID))))
			})
		})

		Context("deeper tests", func() {
			var (
				pod0, pod1, pod2, pod3, pod4, pod5, pod6, pod7         *corev1.Pod
				podWithWrongSchedulerName0, podWithWrongSchedulerName1 *corev1.Pod

				uid0   = "10234"
				uid1   = "11234"
				uid2   = "12236"
				uid3   = "13236"
				uid3_0 = "13237"
				uid4   = "14236"
				uid4_0 = "14237"
				uid5   = "15236"
				uid6   = "24567"
				uid7   = "26567"
			)

			JustBeforeEach(func() {
				pod0 = test.BuildPodGenericWithSchedulerName(namespaceForTest, "p0", uid0, constants.RunaiSchedulerName)
				pod1 = test.BuildPodGenericWithSchedulerName(namespaceForTest, "p1", uid1, constants.RunaiSchedulerName)
				pod2 = test.BuildPodGenericWithSchedulerName(namespaceForTest, "p2", uid2, constants.RunaiSchedulerName)
				pod3 = test.BuildPodGenericWithSchedulerName(namespaceForTest, "p3", uid3, constants.RunaiSchedulerName)
				pod4 = test.BuildPodGenericWithSchedulerName(namespaceForTest, "p4", uid4, constants.RunaiSchedulerName)
				pod5 = test.BuildPodGenericWithSchedulerName(namespaceForTest, "p5", uid5, constants.RunaiSchedulerName)
				pod6 = test.BuildPodGenericWithSchedulerName(namespaceForTest, "p6", uid6, constants.RunaiSchedulerName)
				pod7 = test.BuildPodGenericWithSchedulerName(namespaceForTest, "p7", uid7, constants.RunaiSchedulerName)

				podWithWrongSchedulerName0 = test.BuildPodGenericWithSchedulerName(namespaceForTest, "p3-0", uid3_0, "wrong-scheduler")
				podWithWrongSchedulerName1 = test.BuildPodGenericWithSchedulerName(namespaceForTest, "p4-0", uid4_0, "wrong-scheduler")

				backendClient = backendclient.NewBackendClient(server.URL(), nil)
				fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(test.GetTestScheme()).WithObjects(
					runaiManagedNamespace, project1, kaiProject1,
					pod0, pod1, pod2, pod3, pod4, pod5, pod6, pod7,
					podWithWrongSchedulerName0, podWithWrongSchedulerName1)
				fakeCachedClient = fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
				nodeHandler := handlers.NewNodeHandler(backendClient, fakeCachedClient)
				podHandler = handlers.NewPodHandler(backendClient, fakeCachedClient)
				controllers.AddPodHandlerToCache(backendClient, fakeCachedClient, podHandler, nodeHandler)

				stopper = make(chan struct{})
				go backendClient.StartSendingLivenessPeriodically(stopper)
				fakeCachedClient.Start(ctx, stopper)
			})

			JustAfterEach(func() {
				cancel()
			})

			It("Test sync in namespace with pagination", func() {
				numberOfSyncedPodsPerProject := test.ReadFromChannel(numberOfSyncedPodsInEachSync, 4)
				// the pods with wrong scheduler names should be ignored.
				// with pagination - each request's last item will be the next request's first item
				// so 8 pods with chunk size of 3 will be sent like this:
				// (0, 1, 2),(2, 3, 4),(4, 5, 6),(6, 7)
				Eventually(numberOfSyncedPodsPerProject).Should(HaveExactElements(3, 3, 3, 2))

				syncedPodsUids := test.ReadFromChannel(syncedPods, 4)
				Eventually(syncedPodsUids).Should(HaveExactElements(
					strings.Join([]string{uid0, uid1, uid2}, ", "),
					strings.Join([]string{uid2, uid3, uid4}, ", "),
					strings.Join([]string{uid4, uid5, uid6}, ", "),
					strings.Join([]string{uid6, uid7}, ", "),
				))
			})
		})

		Context("pod re-synced after its workload is synced tests", func() {
			var wg sync.WaitGroup
			var pod0, pod1 *corev1.Pod
			var workload0 *twv2.TrainingWorkload

			BeforeEach(func() {
				pod0 = test.BuildPod(namespaceForTest, "pod0", "imageName", "node1", corev1.PodPending,
					test.BuildResourceList("1000m", "1G"), []metav1.OwnerReference{},
					map[string]string{commonconstants.TopOwnerUidLabel: trainId1})
				pod1 = test.BuildPod(namespaceForTest, "pod1", "imageName", "", corev1.PodPending,
					test.BuildResourceList("1000m", "1G"), []metav1.OwnerReference{},
					map[string]string{commonconstants.TopOwnerUidLabel: trainId1})

				workload0, _ = test.BuildTrainingWorkload(trainName1, namespaceForTest, trainId1)
			})

			JustBeforeEach(func() {
				backendClient = backendclient.NewBackendClient(server.URL(), nil)
				fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(test.GetTestScheme()).WithObjects(
					runaiManagedNamespace, project1, kaiProject1)
				fakeCachedClient = fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
				nodeHandler := handlers.NewNodeHandler(backendClient, fakeCachedClient)
				podHandler = handlers.NewPodHandler(backendClient, fakeCachedClient)
				podRateLimiter := controllers.AddPodHandlerToCache(backendClient, fakeCachedClient, podHandler, nodeHandler)
				workloadHandler := handlers.NewWorkloadHandler(backendClient, fakeCachedClient, podHandler, podRateLimiter)
				controllers.AddExternalWorkloadHandlerToCache(backendClient, fakeCachedClient, workloadHandler)
				controllers.AddWorkspaceHandlerToCache(backendClient, fakeCachedClient, workloadHandler)
				controllers.AddDistributedHandlerToCache(backendClient, fakeCachedClient, workloadHandler)
				controllers.AddTrainingHandlerToCache(backendClient, fakeCachedClient, workloadHandler)

				stopper = make(chan struct{})
				go backendClient.StartSendingLivenessPeriodically(stopper)
				fakeCachedClient.Start(ctx, stopper)

				wg = sync.WaitGroup{}
			})

			JustAfterEach(func() {
				cancel()
			})

			BeforeEach(func() {
				// Sync mock
				server.RouteToHandler("POST", workloadsSyncPath, func(writer http.ResponseWriter, request *http.Request) {
					ghttp.VerifyContentType("application/json")(writer, request)
					ghttp.RespondWith(http.StatusNoContent, "ok")(writer, request)
					GinkgoWriter.Printf("Got a request in externalWorkloads POST Mock")
				})
				// Update status mock
				server.RouteToHandler("PUT", workloadPathWithId, func(writer http.ResponseWriter, request *http.Request) {
					ghttp.VerifyContentType("application/json")(writer, request)
					ghttp.RespondWith(http.StatusOK, "ok")(writer, request)
					GinkgoWriter.Printf("Got a request in externalWorkloads PUT Mock")

					workloadId := strings.TrimPrefix(request.URL.Path, workloadsApiPathPrefix)
					updatedWorkloadId <- workloadId
				})
			})

			It("Test the HandleAdditionOfWorkloadPodsAfterItWasAdded mechanism (enqueuePodsHandlingForLaterIfNeeded)", func() {
				// create a pod that will be synced after its workload is synced
				podIdsToReturn412[string(pod0.UID)] = true
				err := fakeCachedClient.Create(context.Background(), pod0)
				Expect(err).To(BeNil())

				Eventually(failedUpdatePodId).WithTimeout(threeSecondTimeout).Should(Receive(Equal(string(pod0.UID))))

				// create a workload that will be updated and should trigger pod update
				delete(podIdsToReturn412, string(pod0.UID))
				err = fakeCachedClient.Create(context.Background(), workload0)
				Expect(err).To(BeNil())
				Eventually(updatedWorkloadId).WithTimeout(threeSecondTimeout).Should(Receive(Equal(string(workload0.UID))))

				// verify that the pod was synced now
				Eventually(updatedPodId).WithTimeout(threeSecondTimeout).Should(Receive(Equal(string(pod0.UID))))
			})

			It("Test the ResyncPodsAfterWorkloadAdded mechanism (sync failed..)", func() {
				// create 2 pods that will be synced after their workload is synced
				projectsToReturn412[projectIdStr1] = true
				err := fakeCachedClient.Create(context.Background(), pod0)
				Expect(err).To(BeNil())
				err = fakeCachedClient.Create(context.Background(), pod1)
				Expect(err).To(BeNil())

				wg.Add(1)
				go podHandler.HandleSyncEvent(&wg)
				wg.Wait()

				Eventually(projectOfFailedSyncedPods).WithTimeout(threeSecondTimeout).Should(Receive(Equal(projectIdStr1)))

				// create a workload that will be updated and should trigger pods sync
				delete(projectsToReturn412, projectIdStr1)
				err = fakeCachedClient.Create(context.Background(), workload0)
				Expect(err).To(BeNil())
				Eventually(updatedWorkloadId).WithTimeout(threeSecondTimeout).Should(Receive(Equal(string(workload0.UID))))

				Eventually(projectOfSyncedPods).WithTimeout(threeSecondTimeout).Should(Receive(Equal(projectIdStr1)))
			})
		})
	})
})
