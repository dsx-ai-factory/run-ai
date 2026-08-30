package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	swagger "github.com/run-ai/runai/runai-clients/datavolumes-service-client/generated"
	backendclient "github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/handlers"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	fakerunaiclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	dvId1 = uuid.New().String()
	dvId2 = uuid.New().String()
	dvId3 = uuid.New().String()
	dvId4 = uuid.New().String()
	dvId5 = uuid.New().String()
)

var _ = Describe("DatavolumeHandler", func() {
	var syncPath = "/api/v1/datavolumes/sync"
	pathWithId, err := regexp.Compile("/api/v1/datavolumes/([0-9a-zA-Z-].*)/status")
	Expect(err).To(BeNil())
	matches := pathWithId.FindStringSubmatch(fmt.Sprintf("/api/v1/datavolumes/%s/status", dvId1))
	Expect(len(matches)).To(Equal(2))
	Expect(pathWithId.MatchString(fmt.Sprintf("/api/v1/datavolumes/%s/status", dvId1))).To(BeTrue())

	var server *ghttp.Server
	var backendClient *backendclient.BackendClient
	var stopper chan struct{}
	var datavolumeHandler *handlers.DatavolumeHandler
	var syncedDatavolumes chan int
	var updatedDatavolume chan string
	var deletedDatavolume chan string
	BeforeEach(func() {
		syncedDatavolumes = make(chan int, 1)
		DeferCleanup(func() { close(syncedDatavolumes) })
		updatedDatavolume = make(chan string, 1)
		DeferCleanup(func() { close(updatedDatavolume) })
		deletedDatavolume = make(chan string, 1)
		DeferCleanup(func() { close(deletedDatavolume) })
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
		existingDatavolume := BuildDatavolume("datavolume-1", dvId1, v1alpha1.DataVolumeReady, "some message", "pvc-1", "pv-1")
		existingDatavolume2 := BuildDatavolume("datavolume-2", dvId2, v1alpha1.DataVolumeReady, "some message", "pvc-2", "pv-2")
		existingDatavolume3 := BuildDatavolume("datavolume-3", dvId3, v1alpha1.DataVolumeReady, "some message", "pvc-3", "pv-3")
		existingDatavolume4 := BuildDatavolume("datavolume-4", dvId4, v1alpha1.DataVolumeReady, "some message", "pvc-4", "pv-4")
		datavolumeForUpdate := BuildDatavolume("datavolume-1", dvId1, v1alpha1.DataVolumeNotReady, "some message", "pvc-1", "pv-1")
		datavolumeForAdd := BuildDatavolume("datavolume-100", dvId5, v1alpha1.DataVolumeReady, "some message", "pvc-100", "pv-100")

		BeforeEach(func() {

			GinkgoWriter.Printf("Created mock server calls...")
			// Sync mock
			server.RouteToHandler("POST", syncPath, func(writer http.ResponseWriter, request *http.Request) {
				var syncRequest swagger.DatavolumesSyncRequest
				ghttp.VerifyContentType("application/json")(writer, request)
				err := json.NewDecoder(request.Body).Decode(&syncRequest)
				Expect(err).To(BeNil())
				ghttp.RespondWith(200, "ok")
				syncedDatavolumes <- len(syncRequest.Datavolumes)
			})
			// Update status mock
			server.RouteToHandler("PUT", pathWithId, func(writer http.ResponseWriter, request *http.Request) {
				var datavolumeStatusInternal swagger.DatavolumeStatusInternal
				ghttp.VerifyContentType("application/json")(writer, request)
				err := json.NewDecoder(request.Body).Decode(&datavolumeStatusInternal)
				Expect(err).To(BeNil())
				ghttp.RespondWith(http.StatusOK, "ok")(writer, request)

				matches := pathWithId.FindStringSubmatch(request.URL.Path)
				Expect(len(matches)).To(Equal(2))
				datavolumeId := matches[1]

				isDeleted := datavolumeStatusInternal.IsDeleted.Get()
				if isDeleted != nil && *isDeleted {
					deletedDatavolume <- datavolumeId
				} else {
					updatedDatavolume <- datavolumeId
				}
			})
		})

		JustBeforeEach(func() {
			GinkgoWriter.Println("Creating datavolume controller")
			backendClient = backendclient.NewBackendClient(server.URL(), nil)
			runaiClient := fakerunaiclient.NewClientBuilder().WithScheme(GetTestScheme()).WithObjects(
				existingDatavolume, existingDatavolume2, existingDatavolume3, existingDatavolume4).
				WithStatusSubresource(&v1alpha1.DataVolume{}).Build()
			datavolumeHandler = handlers.NewDatavolumeHandler(backendClient, runaiClient)
			stopper = make(chan struct{})
			go backendClient.StartSendingLivenessPeriodically(stopper)
		})

		It("Test add", func() {
			err := datavolumeHandler.HandleAddUpdate(datavolumeForAdd)
			Expect(err).To(BeNil())
			Eventually(updatedDatavolume).WithTimeout(3 * time.Second).Should(Receive(Equal(datavolumeForAdd.Spec.DatavolumeId)))
		})
		It("Test update", func() {
			err := datavolumeHandler.HandleAddUpdate(datavolumeForUpdate)
			Expect(err).To(BeNil())
			Eventually(updatedDatavolume).WithTimeout(3 * time.Second).Should(Receive(Equal(datavolumeForUpdate.Spec.DatavolumeId)))
		})
		It("Test delete", func() {
			err := datavolumeHandler.HandleDelete(existingDatavolume)
			Expect(err).To(BeNil())
			Eventually(deletedDatavolume).WithTimeout(3 * time.Second).Should(Receive(Equal(existingDatavolume.Spec.DatavolumeId)))
		})
		It("Test sync", func() {
			wg := sync.WaitGroup{}
			wg.Add(1)
			datavolumeHandler.HandleSyncEvent(&wg)
			wg.Wait()

			Eventually(syncedDatavolumes).WithTimeout(3 * time.Second).Should(Receive(Equal(4)))
		})
	})

})
