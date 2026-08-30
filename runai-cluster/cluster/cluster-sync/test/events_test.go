package test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/onsi/gomega/ghttp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	assetsApi "github.com/run-ai/runai/runai-clients/assets-service-client/generated"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	backendclient "github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/controllers"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	anotherNamespaceForTests = NamespaceIdForTests + "00"
)

var _ = Describe("EventHandler", func() {
	var eventsPath = "/api/v1/workloads/events"
	var assetsEventsPath = "/api/v1/asset/sync/events"

	var server *ghttp.Server
	var k8sClient client.Client
	var backendClient *backendclient.BackendClient
	var stopper chan struct{}
	var eventController *controllers.EventController
	var numOfSyncedEvents chan int
	var syncedEventsIds chan string
	BeforeEach(func() {
		numOfSyncedEvents = make(chan int, 30)
		DeferCleanup(func() { close(numOfSyncedEvents) })
		syncedEventsIds = make(chan string, 30)
		DeferCleanup(func() { close(syncedEventsIds) })

		GinkgoWriter.Println("Creating server for test...")
		server = CreateBasicRunaiBackend()
		DeferCleanup(func() {
			server.Close()
		})
		GinkgoWriter.Printf("Created server for test %s \n", server.URL())
		server.AllowUnhandledRequests = true
		stopper = make(chan struct{}, 1)
	})

	Context("runai backend is available", Ordered, func() {
		// we want the amount of events to be bigger than the events chunk size (which is currently 5 in tests)
		numOfExistingEventsInEachNamespace := 8
		testNamespaces := []string{NamespaceIdForTests, anotherNamespaceForTests}

		existingEvents := []*corev1.Event{}
		for _, namespace := range testNamespaces {
			for i := 0; i < numOfExistingEventsInEachNamespace; i++ {
				eventName := fmt.Sprintf("event-%d", i)
				existingEvents = append(existingEvents, BuildEvent(namespace, eventName))
			}
		}

		eventsToAddInTest1 := []*corev1.Event{BuildEvent(NamespaceIdForTests, "eventToAdd")}
		eventsToAddInTest2 := []*corev1.Event{BuildEvent(NamespaceIdForTests, "eventToAdd2"),
			BuildEvent(anotherNamespaceForTests, "eventToAdd2.0"),
			BuildEvent(NamespaceIdForTests, "eventToAdd2.1"),
		}

		BeforeAll(func() {
			runaiManagedNamespace := BuildNamespace(NamespaceIdForTests, map[string]string{constants.RunaiQueueLabel: ProjectIdForTests})
			runaiManagedNamespace2 := BuildNamespace(anotherNamespaceForTests, map[string]string{constants.RunaiQueueLabel: ProjectIdForTests + "00"})

			k8sClient = fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).WithObjects(runaiManagedNamespace, runaiManagedNamespace2).Build()
			for _, event := range existingEvents {
				createEvent(event, k8sClient)
			}
		})

		BeforeEach(func() {
			GinkgoWriter.Printf("Created mock server calls...")
			// Sync mock
			server.RouteToHandler("POST", eventsPath, func(writer http.ResponseWriter, request *http.Request) {
				events := &openapi.EventsSyncRequest{}
				ghttp.VerifyContentType("application/json")(writer, request)
				err := json.NewDecoder(request.Body).Decode(events)
				Expect(err).To(BeNil())
				ghttp.RespondWith(200, "ok")(writer, request)

				numOfSyncedEvents <- len(events.Events)

				syncedEvents := []string{}
				for _, e := range events.Events {
					syncedEvents = append(syncedEvents, e.Id)
				}
				syncedEventsIds <- strings.Join(syncedEvents, ", ")
			})

			server.RouteToHandler("POST", assetsEventsPath, func(writer http.ResponseWriter, request *http.Request) {
				events := &assetsApi.EventsSyncRequest{}
				ghttp.VerifyContentType("application/json")(writer, request)
				err := json.NewDecoder(request.Body).Decode(events)
				Expect(err).To(BeNil())
				ghttp.RespondWith(200, "ok")(writer, request)
			})
		})

		JustBeforeEach(func() {
			GinkgoWriter.Println("Creating event controller")
			backendClient = backendclient.NewBackendClient(server.URL(), nil)
			eventController = controllers.NewEventController(k8sClient, backendClient)
			stopper = make(chan struct{})
			go backendClient.StartSendingLivenessPeriodically(stopper)

			go eventController.StartSyncingPeriodically(stopper)
		})

		AfterEach(func() {
			close(stopper)
			server.Reset()
		})

		Context("Test events handler", func() {
			BeforeAll(func() {
				runaiManagedNamespace := BuildNamespace(NamespaceIdForTests, map[string]string{constants.RunaiQueueLabel: ProjectIdForTests})
				runaiManagedNamespace2 := BuildNamespace(anotherNamespaceForTests, map[string]string{constants.RunaiQueueLabel: ProjectIdForTests + "00"})

				k8sClient = fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).WithObjects(runaiManagedNamespace, runaiManagedNamespace2).Build()
				for _, event := range existingEvents {
					createEvent(event, k8sClient)
				}
			})

			It("Test sync events in all namespaces + events added + some events are not sent in the request", func() {
				numberOfSyncedPodsPerProject := ReadFromChannel(numOfSyncedEvents, 4)
				Eventually(numberOfSyncedPodsPerProject).Should(HaveExactElements(5, 5, 5, 1))

				// add some events and validate only they are sent in the HTTP request
				time.Sleep(syncEventsIntervalSeconds*time.Second + time.Second)
				for _, event := range eventsToAddInTest1 {
					createEvent(event, k8sClient)
				}

				Eventually(syncedEventsIds).WithTimeout(syncEventsIntervalSeconds*time.Second + time.Second).Should(Receive(Equal(eventsToAddInTest1[0].Name)))

				time.Sleep(syncEventsIntervalSeconds*time.Second + time.Second)
				expectedEventIds := []string{}
				for _, event := range eventsToAddInTest2 {
					createEvent(event, k8sClient)
					expectedEventIds = append(expectedEventIds, event.Name)
				}

				syncedEvents := ReadFromChannel(syncedEventsIds, 1)[0]
				syncedEventsIds := strings.Split(syncedEvents, ", ")
				Expect(syncedEventsIds).To(ConsistOf(expectedEventIds))

				Eventually(numOfSyncedEvents).WithTimeout(syncEventsIntervalSeconds*time.Second + time.Second).Should(Receive(Equal(3)))
			})
		})
	})

})

func createEvent(event *corev1.Event, k8sClient client.Client) {
	e := event.DeepCopy()
	e.LastTimestamp = metav1.Time{Time: time.Now()}
	err := k8sClient.Create(context.TODO(), e)
	Expect(err).To(BeNil())
}
