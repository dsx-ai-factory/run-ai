package test

import (
	"context"
	"encoding/json"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	k8sobjectstracker "github.com/run-ai/runai/runai-clients/k8s-objects-tracker-client/generated"
	backendclient "github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/cachedclient"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/constants"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/controllers"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/test/fakecachedclient"
	clusterconstants "github.com/run-ai/runai/runai-cluster/cluster/common/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("A ConfigMap controller test", func() {
	var server *ghttp.Server
	var backendClient *backendclient.BackendClient
	var fakeCachedClient cachedclient.CachedClient

	var syncedConfigMapName chan string
	var deletedConfigMapName chan string
	var createdConfigMapName chan string
	var syncedConfigMapNumber chan int
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

	Context("runai backend is alive, one configMap is in the cluster", Serial, func() {
		var existingConfigMap = BuildConfigMap(NamespaceIdForTests, "existing-config-map", map[string]string{clusterconstants.RunaiResourceLabel: "config-map"})
		var addedConfigMap = BuildConfigMap(NamespaceIdForTests, "new-config-map-1", map[string]string{clusterconstants.RunaiResourceLabel: "config-map"})
		var privateConfigMap = BuildConfigMap(NamespaceIdForTests, "new-config-map-1", nil)
		var expectedIssues []k8sobjectstracker.Issue

		BeforeEach(func() {
			// init channels
			syncedConfigMapName = make(chan string, 1)
			deletedConfigMapName = make(chan string, 1)
			createdConfigMapName = make(chan string, 1)
			syncedConfigMapNumber = make(chan int, 1)
			tested = make(chan bool, 1)

			//start server Route Handler of ConfigMap sync
			server.RouteToHandler("POST", "/api/v1/config-maps/7daba6ee-2ea9-4af3-b735-a364bcdff72f/namespaces/runai-namespace-project", func(writer http.ResponseWriter, request *http.Request) {
				var configMaps k8sobjectstracker.ConfigMaps
				ghttp.VerifyContentType("application/json")(writer, request)
				err := json.NewDecoder(request.Body).Decode(&configMaps)
				Expect(err).To(BeNil())
				ghttp.RespondWith(200, "ok")(writer, request)
				syncedConfigMapNumber <- len(configMaps.Configmaps)
				syncedConfigMapName <- configMaps.Configmaps[0].Name
			})

			//start server Route Handler of ConfigMap create
			server.RouteToHandler("POST", "/api/v1/config-maps/7daba6ee-2ea9-4af3-b735-a364bcdff72f/namespaces/runai-namespace-project/new-config-map-1", func(writer http.ResponseWriter, request *http.Request) {
				ghttp.RespondWith(200, "ok")(writer, request)

				createdConfigMapName <- "new-config-map-1"
				syncedConfigMapNumber <- 1
			})

			//start server Route Handler of ConfigMap delete
			server.RouteToHandler("DELETE", "/api/v1/config-maps/7daba6ee-2ea9-4af3-b735-a364bcdff72f/namespaces/runai-namespace-project/existing-config-map", func(writer http.ResponseWriter, request *http.Request) {
				ghttp.RespondWith(http.StatusOK, "ok")(writer, request)
				deletedConfigMapName <- "existing-config-map"
			})

			server.RouteToHandler("POST", "/api/v1/config-maps/7daba6ee-2ea9-4af3-b735-a364bcdff72f/namespaces/runai-namespace-project/issued-config-map", func(writer http.ResponseWriter, request *http.Request) {
				ghttp.VerifyContentType("application/json")(writer, request)
				ghttp.RespondWith(200, "ok")(writer, request)

				var configMap k8sobjectstracker.ConfigMap
				err := json.NewDecoder(request.Body).Decode(&configMap)
				Expect(err).To(BeNil())

				Expect(configMap.Issues).Should(ContainElements(expectedIssues))
				tested <- true
			})

			runaiManagedNamespace := BuildNamespace(NamespaceIdForTests, map[string]string{constants.RunaiQueueLabel: ProjectIdForTests})
			runaiProject := BuildProject(ProjectIdForTests, ProjectIdForTests, NamespaceIdForTests)
			kaiProject := BuildKaiProject(ProjectIdForTests, ProjectIdForTests, NamespaceIdForTests)

			// start ConfigMap controller and Liveness
			backendClient = backendclient.NewBackendClient(server.URL(), nil)

			fakeClientBuilder := fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).WithObjects(runaiManagedNamespace, runaiProject, kaiProject, existingConfigMap)
			fakeCachedClient = fakecachedclient.NewFakeCachedClient(fakeClientBuilder)
			controllers.AddConfigMapHandlerToCache(backendClient, fakeCachedClient)

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
			close(syncedConfigMapName)
			close(deletedConfigMapName)
			close(createdConfigMapName)
			close(syncedConfigMapNumber)
			close(tested)
		})

		It("Tests ConfigMap sync", func() {
			Eventually(syncedConfigMapNumber).WithTimeout(threeSecondTimeout).Should(Receive(Equal(1)))
			Eventually(syncedConfigMapName).WithTimeout(threeSecondTimeout).Should(Receive(Equal(existingConfigMap.Name)))
		})

		It("Tests ConfigMap create", func() {
			Eventually(syncedConfigMapNumber).WithTimeout(threeSecondTimeout).Should(Receive(Equal(1)))

			err := fakeCachedClient.Create(context.TODO(), addedConfigMap)
			Expect(err).To(BeNil())

			Eventually(createdConfigMapName).WithTimeout(threeSecondTimeout).Should(Receive(Equal(addedConfigMap.Name)))
		})

		It("Tests private ConfigMap", func() {
			Eventually(syncedConfigMapNumber).WithTimeout(threeSecondTimeout).Should(Receive(Equal(1)))

			err := fakeCachedClient.Create(context.TODO(), privateConfigMap)
			Expect(err).To(BeNil())

			Eventually(createdConfigMapName).WithTimeout(threeSecondTimeout).ShouldNot(Receive())
		})

		It("Tests ConfigMap delete", func() {
			Eventually(syncedConfigMapNumber).WithTimeout(threeSecondTimeout).Should(Receive(Equal(1)))
			Eventually(syncedConfigMapName).WithTimeout(threeSecondTimeout).Should(Receive(Equal(existingConfigMap.Name)))

			configMapToDelete := &corev1.ConfigMap{}
			err := fakeCachedClient.Get(context.TODO(), types.NamespacedName{Name: existingConfigMap.Name, Namespace: NamespaceIdForTests}, configMapToDelete)
			Expect(err).To(BeNil())

			err = fakeCachedClient.Delete(context.TODO(), configMapToDelete)
			Expect(err).To(BeNil())

			Eventually(deletedConfigMapName).WithTimeout(threeSecondTimeout).Should(Receive(Equal(existingConfigMap.Name)))

		})
		It("Tests ConfigMap create with issue", func() {
			Eventually(syncedConfigMapNumber).WithTimeout(threeSecondTimeout).Should(Receive(Equal(1)))
			var issuedConfigMap = BuildConfigMap(NamespaceIdForTests, "issued-config-map", map[string]string{clusterconstants.RunaiResourceLabel: "config-map"})
			expectedIssues = SetIssuesOnNamespace(&issuedConfigMap.ObjectMeta, "ns1", "ns2")
			err := fakeCachedClient.Create(context.TODO(), issuedConfigMap)
			Expect(err).To(BeNil())
			Eventually(tested).WithTimeout(threeSecondTimeout).Should(Receive(BeTrue()))
		})

	})
})
