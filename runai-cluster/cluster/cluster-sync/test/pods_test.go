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
	"time"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/test/fakecachedclient"
	twv2 "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/workload-controller/pkg/api/trainingworkload/v2alpha1"
	commonconstants "github.com/dsx-ai-factory/run-ai/runai-cluster/common/constants"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"k8s.io/utils/ptr"

	backendclient "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/cluster-sync/pkg/controllers"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const threeSecondTimeout = 3 * time.Second

const (
	podId1                      = "123"
	knativeWorkloadIdAnnotation = "some-knative-iddddd"
)

// This test suite is a suite that tests a flow from change of pods state in k8s to request sent to backend

var _ = Describe("A Pod controller test", func() {
	var syncPath = "/api/v1/workloads/pods"

	pathWithIds, err := regexp.Compile("/api/v1/workloads/pods/([0-9a-zA-Z-].*)")
	Expect(err).To(BeNil())
	matches := pathWithIds.FindStringSubmatch(fmt.Sprintf("/api/v1/workloads/pods/%s", podId1))
	Expect(len(matches)).To(Equal(2))
	Expect(pathWithIds.MatchString(fmt.Sprintf("/api/v1/workloads/pods/%s", podId1))).To(BeTrue())

	var podHandler *handlers.PodHandler

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

	runaiManagedNamespace := BuildNamespace(namespaceForTest, map[string]string{constants.RunaiQueueLabel: projectForTest})
	runaiManagedNamespace2 := BuildNamespace(namespaceForTest2, map[string]string{constants.RunaiQueueLabel: projectForTest2})
	project1 := BuildProjectWithDepId(projectForTest, projectIdStr1, namespaceForTest, depId)
	project2 := BuildProjectWithDepId(projectForTest2, projectIdStr2, namespaceForTest2, depId)
	kaiProject1 := BuildKaiProjectWithDepId(projectForTest, projectIdStr1, namespaceForTest, depId)
	kaiProject2 := BuildKaiProjectWithDepId(projectForTest2, projectIdStr2, namespaceForTest2, depId)
	interactiveExisting, _ := BuildInteractiveWorkload(intName1, namespaceForTest, intId1)
	trainExisting, _ := BuildTrainingWorkloadBackendSourced(trainName1, namespaceForTest2, trainId1, trainWorkloadId1)
	existingService := BuildService("srv-3", namespaceForTest2, "srv-3-uid")
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
		server = CreateBasicRunaiBackend()
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

		BeforeEach(func() {
			existingPod = BuildPod(namespaceForTest, "p", "imageName", "node1", corev1.PodRunning,
				BuildResourceList("1000m", "1G"), []metav1.OwnerReference{{APIVersion: "run.ai/v2alpha1",
					Kind: "InteractiveWorkload", Name: intName1, UID: intId1, Controller: nil, BlockOwnerDeletion: nil,
				}}, make(map[string]string))
			podToAdd0 = BuildPod(namespaceForTest2, "pod0", "imageName", "node1", corev1.PodPending,
				BuildResourceList("1000m", "1G"), []metav1.OwnerReference{{APIVersion: "run.ai/v2alpha1",
					Kind: "TrainingWorkload", Name: trainName1, UID: trainId1, Controller: nil, BlockOwnerDeletion: nil,
				}}, map[string]string{v1alpha1.NodePoolNameLabel: "some-node-pool"})
			podToAdd1 = BuildPod(namespaceForTest2, "pod1", "imageName", "node1", corev1.PodRunning,
				BuildResourceList("1000m", "1G"), []metav1.OwnerReference{}, map[string]string{v1alpha1.NodePoolNameLabel: "some-node-pool"})
			podToAdd2 = BuildPod(namespaceForTest2, "pod2", "imageName", "node1", corev1.PodRunning,
				BuildResourceList("1000m", "1G"), []metav1.OwnerReference{{APIVersion: "serving.knative.dev/v1",
					Kind: "Service", Name: "srv-3", UID: "srv-3-uid", Controller: nil, BlockOwnerDeletion: nil,
				}}, map[string]string{})
			podToAdd3 = BuildPod(namespaceForTest, "pod3", "imageName", "node1", corev1.PodRunning,
				BuildResourceList("1000m", "1G"), []metav1.OwnerReference{{APIVersion: "run.ai/v2alpha1",
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
				fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).WithObjects(
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

			It("Tests pod add", func() {
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
			It("Test pods update", func() {
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
				podToDelete := &corev1.Pod{}
				err := fakeCachedClient.Get(ctx, client.ObjectKey{Namespace: existingPod.Namespace, Name: existingPod.Name},
					podToDelete)
				Expect(err).To(BeNil())
				err = fakeCachedClient.Delete(ctx, podToDelete)
				Expect(err).To(BeNil())
				Eventually(deletedPodId).WithTimeout(threeSecondTimeout).Should(Receive(Equal(string(existingPod.UID))))
			})
		})

		Context("pod re-synced after its workload is synced tests", func() {
			var pod0 *corev1.Pod
			var workload0 *twv2.TrainingWorkload

			BeforeEach(func() {
				pod0 = BuildPod(namespaceForTest, "pod0", "imageName", "node1", corev1.PodPending,
					BuildResourceList("1000m", "1G"), []metav1.OwnerReference{},
					map[string]string{commonconstants.TopOwnerUidLabel: trainId1, commonconstants.TopOwnerWorkloadIdLabel: trainId1})

				workload0, _ = BuildTrainingWorkload(trainName1, namespaceForTest, trainId1)
			})

			JustBeforeEach(func() {
				backendClient = backendclient.NewBackendClient(server.URL(), nil)
				fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).WithObjects(
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
		})
	})
})

//nolint:unused
func _expectInternalServerError(err error, errorStringRegexp string) {
	Expect(err).To(HaveOccurred())

	errString := err.Error()
	Expect(errString).To(ContainSubstring("500"))
	Expect(errString).To(ContainSubstring("Internal Server Error"))
	Expect(errString).To(MatchRegexp(errorStringRegexp))
}

//nolint:unused
func _expectInternalServerErrorPods(err error) {
	_expectInternalServerError(err, "URL: .*/api/v1/workloads/pods.*")
}
