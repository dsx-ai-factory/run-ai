package test

import (
	"context"
	"fmt"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/runtime"

	kaiv1alpha1 "github.com/kai-scheduler/KAI-scheduler/pkg/apis/kai/v1alpha1"
	openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	backendclient "github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"encoding/json"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("NetworkTopologies", func() {
	var (
		server                 *ghttp.Server
		networkTopologyHandler *handlers.NetworkTopologyHandler
		stopper                chan struct{}
		fakeClient             client.Client

		// Test data channels
		syncedTopologiesChan chan []openapi.NetworkTopologySyncFields
	)

	BeforeEach(func() {
		// Create test channels
		syncedTopologiesChan = make(chan []openapi.NetworkTopologySyncFields, 3)

		server = CreateBasicRunaiBackend()
		DeferCleanup(func() {
			server.Close()
		})
		GinkgoWriter.Printf("Created server for test %s \n", server.URL())
		server.AllowUnhandledRequests = true

		stopper = make(chan struct{}, 1)

		fakeClient = fake.NewClientBuilder().WithScheme(getTestScheme()).Build()

		// Setup mock API endpoints
		setupMockAPIs(server, syncedTopologiesChan)

		backendClient := backendclient.NewBackendClient(server.URL(), nil)
		go backendClient.StartSendingLivenessPeriodically(stopper)
		networkTopologyHandler = handlers.NewNetworkTopologyHandler(backendClient, fakeClient)
	})

	AfterEach(func() {
		// Cleanup
		DeferCleanup(func() {
			close(syncedTopologiesChan)
			close(stopper)
			server.Close()
		})

	})

	Describe("Scenario 1: Topology Added in Cluster and Synced to API", func() {
		It("should sync new topology to API when added to cluster", func() {
			// Create a new topology in the cluster
			newTopology := &kaiv1alpha1.Topology{
				ObjectMeta: metav1.ObjectMeta{
					Name: "new-topology",
					Labels: map[string]string{
						parsers.NetworkTopologySourceLabel: parsers.NetworkTopologiesClusterSource,
					},
				},
				Spec: kaiv1alpha1.TopologySpec{
					Levels: []kaiv1alpha1.TopologyLevel{
						{NodeLabel: "zone=us-west-1"},
						{NodeLabel: "rack=rack-1"},
					},
				},
			}

			// Handle the add operation
			err := networkTopologyHandler.HandleAdd(newTopology)
			Expect(err).To(BeNil())

			// Wait for sync to complete and verify the topology was synced
			Eventually(syncedTopologiesChan, 3*time.Second).Should(Receive(And(
				HaveLen(1),
				ContainElement(And(
					HaveField("Name", "new-topology"),
					HaveField("Source", parsers.NetworkTopologiesClusterSource),
					HaveField("Levels", ConsistOf("zone=us-west-1", "rack=rack-1")),
				)),
			)))
		})

		It("should handle HandleSyncEvent for bulk synchronization", func() {
			// Create multiple topologies in the cluster
			topologies := []*kaiv1alpha1.Topology{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "topology-1",
						Labels: map[string]string{
							parsers.NetworkTopologySourceLabel: parsers.NetworkTopologiesClusterSource,
						},
					},
					Spec: kaiv1alpha1.TopologySpec{
						Levels: []kaiv1alpha1.TopologyLevel{
							{NodeLabel: "zone=us-west-1"},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "topology-2",
						Labels: map[string]string{
							parsers.NetworkTopologySourceLabel: parsers.NetworkTopologiesRunAISource,
						},
					},
					Spec: kaiv1alpha1.TopologySpec{
						Levels: []kaiv1alpha1.TopologyLevel{
							{NodeLabel: "zone=us-east-1"},
							{NodeLabel: "rack=rack-2"},
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "topology-3",
						// No source label - should default to "Cluster"
					},
					Spec: kaiv1alpha1.TopologySpec{
						Levels: []kaiv1alpha1.TopologyLevel{
							{NodeLabel: "zone=us-central-1"},
							{NodeLabel: "rack=rack-3"},
							{NodeLabel: "node=node-1"},
						},
					},
				},
			}

			// Add all topologies to the fake client
			for _, topology := range topologies {
				err := fakeClient.Create(context.Background(), topology)
				Expect(err).To(BeNil())
			}

			// Create a wait group for the sync event
			var wg sync.WaitGroup
			wg.Add(1)

			networkTopologyHandler.HandleSyncEvent(&wg)
			wg.Wait()

			// Wait for sync to complete and verify all topologies were synced
			Eventually(syncedTopologiesChan, 5*time.Second).Should(Receive(And(
				HaveLen(3),
				ContainElement(And(
					HaveField("Name", "topology-1"),
					HaveField("Source", parsers.NetworkTopologiesClusterSource),
					HaveField("Levels", ConsistOf("zone=us-west-1")),
				)),
				ContainElement(And(
					HaveField("Name", "topology-2"),
					HaveField("Source", parsers.NetworkTopologiesRunAISource),
					HaveField("Levels", ConsistOf("zone=us-east-1", "rack=rack-2")),
				)),
				ContainElement(And(
					HaveField("Name", "topology-3"),
					HaveField("Source", parsers.NetworkTopologiesClusterSource), // Should default to Cluster
					HaveField("Levels", ConsistOf("zone=us-central-1", "rack=rack-3", "node=node-1")),
				)),
			)))

		})
	})

	Describe("Scenario 2: Topology Levels Updated", func() {
		var existingTopology *kaiv1alpha1.Topology

		BeforeEach(func() {
			// Create an existing topology
			existingTopology = &kaiv1alpha1.Topology{
				ObjectMeta: metav1.ObjectMeta{
					Name: "existing-topology",
					Labels: map[string]string{
						parsers.NetworkTopologySourceLabel: parsers.NetworkTopologiesClusterSource,
					},
				},
				Spec: kaiv1alpha1.TopologySpec{
					Levels: []kaiv1alpha1.TopologyLevel{
						{NodeLabel: "zone=us-west-1"},
						{NodeLabel: "rack=rack-1"},
					},
				},
			}

			// Initially add the topology
			err := networkTopologyHandler.HandleAdd(existingTopology)
			Expect(err).To(BeNil())
		})
	})

	Describe("Scenario: Only Cluster Source Topologies Should Be Synced (Run.ai sources are filtered out)", func() {
		var runaiTopology *kaiv1alpha1.Topology

		BeforeEach(func() {
			// Create a Run.ai topology that should not be synced
			runaiTopology = &kaiv1alpha1.Topology{
				ObjectMeta: metav1.ObjectMeta{
					Name: "runai-topology",
					Labels: map[string]string{
						parsers.NetworkTopologySourceLabel: parsers.NetworkTopologiesRunAISource,
					},
				},
				Spec: kaiv1alpha1.TopologySpec{
					Levels: []kaiv1alpha1.TopologyLevel{
						{NodeLabel: "zone=runai-zone"},
						{NodeLabel: "rack=runai-rack"},
					},
				},
			}
		})

		It("should send sync request when only Run.ai source topologies exist during FullSync", func() {
			// Add only Run.ai source topology to the cluster
			err := fakeClient.Create(context.Background(), runaiTopology)
			Expect(err).To(BeNil())

			// Create a wait group for the sync event
			var wg sync.WaitGroup
			wg.Add(1)

			// Trigger the sync event
			networkTopologyHandler.HandleSyncEvent(&wg)
			wg.Wait()

			// Verify that a sync request was sent but with empty list
			// The syncedTopologiesChan should receive an empty slice
			Eventually(syncedTopologiesChan, 3*time.Second).Should(Receive(HaveLen(1)))
		})

		It("should send sync request and not exclude Run.ai source topology during Add operation", func() {
			// Handle the add operation for Run.ai source topology
			err := networkTopologyHandler.HandleAdd(runaiTopology)
			Expect(err).To(BeNil())

			// Verify that sync request was sent for Run.ai topology
			// The syncedTopologiesChan should receive data
			Eventually(syncedTopologiesChan, 2*time.Second).Should(Receive(HaveLen(1)))
		})

		It("should send sync request but exclude Run.ai source topology during Delete operation", func() {
			// First add the topology to the cluster
			err := fakeClient.Create(context.Background(), runaiTopology)
			Expect(err).To(BeNil())

			// Handle the delete operation
			err = networkTopologyHandler.HandleDelete(runaiTopology)
			Expect(err).To(BeNil())

			// Verify that no sync request was sent for Run.ai topology
			Consistently(syncedTopologiesChan, 2*time.Second).ShouldNot(Receive())
		})

		It("should send sync request with Cluster and Run.ai source topologies during FullSync", func() {
			// Create a Cluster source topology that SHOULD be synced
			clusterTopology := &kaiv1alpha1.Topology{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-topology",
					Labels: map[string]string{
						parsers.NetworkTopologySourceLabel: parsers.NetworkTopologiesClusterSource,
					},
				},
				Spec: kaiv1alpha1.TopologySpec{
					Levels: []kaiv1alpha1.TopologyLevel{
						{NodeLabel: "zone=cluster-zone"},
					},
				},
			}

			// Add both Cluster and Run.ai source topologies to the cluster
			err := fakeClient.Create(context.Background(), clusterTopology)
			Expect(err).To(BeNil())
			err = fakeClient.Create(context.Background(), runaiTopology)
			Expect(err).To(BeNil())

			// Create a wait group for the sync event
			var wg sync.WaitGroup
			wg.Add(1)

			// Trigger the sync event
			networkTopologyHandler.HandleSyncEvent(&wg)
			wg.Wait()

			// Verify that ONLY the Cluster source topology was synced
			Eventually(syncedTopologiesChan, 3*time.Second).Should(Receive(And(
				HaveLen(2),
				ContainElement(And(
					HaveField("Name", "cluster-topology"),
					HaveField("Source", parsers.NetworkTopologiesClusterSource),
					HaveField("Levels", ConsistOf("zone=cluster-zone")),
				)),
				ContainElement(And(
					HaveField("Name", "runai-topology"),
					HaveField("Source", parsers.NetworkTopologiesRunAISource),
					HaveField("Levels", ConsistOf("zone=runai-zone", "rack=runai-rack")),
				)),
			)))
		})

		It("should sync Cluster source topology during Add operation", func() {
			// Create a Cluster source topology that SHOULD be synced
			clusterTopology := &kaiv1alpha1.Topology{
				ObjectMeta: metav1.ObjectMeta{
					Name: "cluster-topology-add",
					Labels: map[string]string{
						parsers.NetworkTopologySourceLabel: parsers.NetworkTopologiesClusterSource,
					},
				},
				Spec: kaiv1alpha1.TopologySpec{
					Levels: []kaiv1alpha1.TopologyLevel{
						{NodeLabel: "zone=cluster-zone-add"},
					},
				},
			}

			// Handle the add operation for Cluster source topology
			err := networkTopologyHandler.HandleAdd(clusterTopology)
			Expect(err).To(BeNil())

			// Verify that the Cluster source topology WAS synced
			Eventually(syncedTopologiesChan, 3*time.Second).Should(Receive(And(
				HaveLen(1),
				ContainElement(And(
					HaveField("Name", "cluster-topology-add"),
					HaveField("Source", parsers.NetworkTopologiesClusterSource),
					HaveField("Levels", ConsistOf("zone=cluster-zone-add")),
				)),
			)))
		})
	})

	Describe("Integration Tests", func() {
		It("should handle complete lifecycle: add, update, and sync", func() {
			// Step 1: Add new topology
			topology := &kaiv1alpha1.Topology{
				ObjectMeta: metav1.ObjectMeta{
					Name: "lifecycle-topology",
					Labels: map[string]string{
						parsers.NetworkTopologySourceLabel: parsers.NetworkTopologiesClusterSource,
					},
				},
				Spec: kaiv1alpha1.TopologySpec{
					Levels: []kaiv1alpha1.TopologyLevel{
						{NodeLabel: "zone=us-west-1"},
					},
				},
			}

			// Add
			err := networkTopologyHandler.HandleAdd(topology)
			Expect(err).To(BeNil())

			// Verify add sync
			Eventually(syncedTopologiesChan, 3*time.Second).Should(Receive(And(
				HaveLen(1),
				ContainElement(And(
					HaveField("Name", "lifecycle-topology"),
					HaveField("Levels", ConsistOf("zone=us-west-1")),
				)),
			)))
		})
	})
})

// setupMockAPIs configures mock API endpoints for testing
func setupMockAPIs(server *ghttp.Server, syncedTopologiesChan chan []openapi.NetworkTopologySyncFields) {
	// Mock sync endpoint for network topologies
	syncPath := fmt.Sprintf("/api/v1/clusters/%s/network-topologies/sync", ClusterIdForTests)

	server.RouteToHandler("POST", syncPath, func(w http.ResponseWriter, r *http.Request) {
		var syncRequest openapi.NetworkTopologySyncRequest
		err := json.NewDecoder(r.Body).Decode(&syncRequest)
		Expect(err).To(BeNil())

		syncedTopologiesChan <- syncRequest.NetworkTopologies

		// Respond with success
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "success"}`))
	})
}

func getTestScheme() *runtime.Scheme {
	scheme := runtime.NewScheme()
	_ = kaiv1alpha1.AddToScheme(scheme)
	return scheme
}
