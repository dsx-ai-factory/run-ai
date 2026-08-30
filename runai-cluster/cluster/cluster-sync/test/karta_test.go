package test

import (
	"context"
	"encoding/json"
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/handlers/parsers"
	"github.com/run-ai/runai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"net/http"
	"sync"
	"time"

	kartav1alpha1 "github.com/run-ai/karta/pkg/api/runai/v1alpha1"
	openapi "github.com/run-ai/runai/runai-clients/workloads-client/generated"
	backendclient "github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/backend-client"
	"github.com/run-ai/runai/runai-cluster/cluster/cluster-sync/pkg/handlers"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const (
	riVersion = "v1"
	kartaId   = "karta-123-uid"
)

var _ = Describe("KartaHandler", func() {
	var server *ghttp.Server
	var backendClient *backendclient.BackendClient
	var stopper chan struct{}

	BeforeEach(func() {
		server = CreateBasicRunaiBackend()
		DeferCleanup(func() { server.Close() })
		server.AllowUnhandledRequests = true
		stopper = make(chan struct{}, 1)
		DeferCleanup(func() { close(stopper) })
	})

	Context("backend is available", func() {
		var (
			fakeClient   client.WithWatch
			kartaHandler *handlers.KartaHandler

			receivedSync chan openapi.WorkloadTypesForSync
			testRi       = &kartav1alpha1.Karta{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "karta-1",
					Labels: map[string]string{v1alpha1.WorkloadTypeIdLabel: kartaId},
				},
				Spec: kartav1alpha1.KartaSpec{
					StructureDefinition: kartav1alpha1.StructureDefinition{
						RootComponent: kartav1alpha1.ComponentDefinition{
							Kind: &kartav1alpha1.GroupVersionKind{
								Version: riVersion,
							},
						},
					},
				},
				Status: kartav1alpha1.KartaStatus{
					Conditions: []metav1.Condition{
						{Type: "A", Status: metav1.ConditionFalse, Message: "m", Reason: "r", LastTransitionTime: metav1.NewTime(time.Now())},
						{Type: string(v1alpha1.SyncEnabled), Status: metav1.ConditionTrue, Message: "", Reason: "", LastTransitionTime: metav1.NewTime(time.Now())},
					},
				},
			}
		)

		BeforeEach(func() {
			receivedSync = make(chan openapi.WorkloadTypesForSync, 1)
			DeferCleanup(func() { close(receivedSync) })

			// Mock sync endpoint
			syncPath := "/api/v1/workload-types/sync"
			server.RouteToHandler("POST", syncPath, func(writer http.ResponseWriter, request *http.Request) {
				ghttp.VerifyContentType("application/json")(writer, request)

				// Verify clusterId query param exists
				Expect(request.URL.Query().Get("clusterId")).To(Equal(ClusterIdForTests))

				var body openapi.WorkloadTypesForSync
				decoder := json.NewDecoder(request.Body)
				err := decoder.Decode(&body)
				Expect(err).To(BeNil())
				receivedSync <- body
				ghttp.RespondWith(http.StatusOK, "ok")(writer, request)
			})

			backendClient = backendclient.NewBackendClient(server.URL(), nil)
			go backendClient.StartSendingLivenessPeriodically(stopper)

			fakeClient = fake.NewClientBuilder().WithScheme(GetTestScheme()).WithObjects().Build()
			kartaHandler = handlers.NewKartaHandler(backendClient, fakeClient)
		})

		It("sends expected payload on add/update", func() {
			err := kartaHandler.HandleAddUpdate(testRi)
			Expect(err).To(BeNil())

			Eventually(receivedSync).WithTimeout(3 * time.Second).Should(Receive(SatisfyAll(
				WithTransform(func(r openapi.WorkloadTypesForSync) bool { return r.FullSync != nil && *r.FullSync == false }, BeTrue()),
				WithTransform(func(r openapi.WorkloadTypesForSync) int { return len(r.WorkloadTypes) }, Equal(1)),
				WithTransform(func(r openapi.WorkloadTypesForSync) string { return r.WorkloadTypes[0].GetId() }, Equal(kartaId)),
				WithTransform(func(r openapi.WorkloadTypesForSync) string { return r.WorkloadTypes[0].GetPhase() }, Equal("Ready")),
				WithTransform(func(r openapi.WorkloadTypesForSync) string { return ptr.Deref(r.WorkloadTypes[0].Version, "") }, Equal(riVersion)),
				WithTransform(func(r openapi.WorkloadTypesForSync) int { return len(r.WorkloadTypes[0].Conditions) }, Equal(2)),
				// First condition: A, False, message/reason present
				WithTransform(func(r openapi.WorkloadTypesForSync) string { return r.WorkloadTypes[0].Conditions[0].GetType() }, Equal("A")),
				WithTransform(func(r openapi.WorkloadTypesForSync) string { return r.WorkloadTypes[0].Conditions[0].GetStatus() }, Equal(string(metav1.ConditionFalse))),
				WithTransform(func(r openapi.WorkloadTypesForSync) string { return r.WorkloadTypes[0].Conditions[0].GetReason() }, Equal("r")),
				WithTransform(func(r openapi.WorkloadTypesForSync) string { return r.WorkloadTypes[0].Conditions[0].GetMessage() }, Equal("m")),
				// Second condition: SyncEnabled, True, message/reason present
				WithTransform(func(r openapi.WorkloadTypesForSync) string { return r.WorkloadTypes[0].Conditions[1].GetType() }, BeEquivalentTo(v1alpha1.SyncEnabled)),
				WithTransform(func(r openapi.WorkloadTypesForSync) string { return r.WorkloadTypes[0].Conditions[1].GetStatus() }, Equal(string(metav1.ConditionTrue))),
				WithTransform(func(r openapi.WorkloadTypesForSync) string { return r.WorkloadTypes[0].Conditions[1].GetReason() }, Equal("")),
				WithTransform(func(r openapi.WorkloadTypesForSync) string { return r.WorkloadTypes[0].Conditions[1].GetMessage() }, Equal("")),
			)))
		})

		It("sends expected payload on add/update deleting", func() {
			testRi.DeletionTimestamp = ptr.To(metav1.NewTime(time.Now()))
			err := kartaHandler.HandleAddUpdate(testRi)
			Expect(err).To(BeNil())

			Eventually(receivedSync).WithTimeout(3 * time.Second).Should(Receive(SatisfyAll(
				WithTransform(func(r openapi.WorkloadTypesForSync) bool { return r.FullSync != nil && *r.FullSync == false }, BeTrue()),
				WithTransform(func(r openapi.WorkloadTypesForSync) int { return len(r.WorkloadTypes) }, Equal(1)),
				WithTransform(func(r openapi.WorkloadTypesForSync) string { return r.WorkloadTypes[0].GetId() }, Equal(kartaId)),
				WithTransform(func(r openapi.WorkloadTypesForSync) string { return r.WorkloadTypes[0].GetPhase() }, Equal("Deleting")),
				WithTransform(func(r openapi.WorkloadTypesForSync) string { return ptr.Deref(r.WorkloadTypes[0].Version, "") }, Equal(riVersion)),
				WithTransform(func(r openapi.WorkloadTypesForSync) int { return len(r.WorkloadTypes[0].Conditions) }, Equal(2)),
			)))
		})

		It("sends expected payload on delete", func() {
			err := kartaHandler.HandleDelete(testRi)
			Expect(err).To(BeNil())

			Eventually(receivedSync).WithTimeout(3 * time.Second).Should(Receive(SatisfyAll(
				WithTransform(func(r openapi.WorkloadTypesForSync) bool { return r.FullSync != nil && *r.FullSync == false }, BeTrue()),
				WithTransform(func(r openapi.WorkloadTypesForSync) int { return len(r.WorkloadTypes) }, Equal(1)),
				WithTransform(func(r openapi.WorkloadTypesForSync) string { return r.WorkloadTypes[0].GetId() }, Equal(kartaId)),
				WithTransform(func(r openapi.WorkloadTypesForSync) string { return r.WorkloadTypes[0].GetPhase() }, Equal("Deleted")),
				WithTransform(func(r openapi.WorkloadTypesForSync) string { return ptr.Deref(r.WorkloadTypes[0].Version, "") }, Equal(riVersion)),
				WithTransform(func(r openapi.WorkloadTypesForSync) int { return len(r.WorkloadTypes[0].Conditions) }, Equal(0)),
			)))
		})

		It("returns error when resource interface id label is missing", func() {
			karta := &kartav1alpha1.Karta{
				ObjectMeta: metav1.ObjectMeta{Name: "karta-missing-id"},
				Spec: kartav1alpha1.KartaSpec{
					StructureDefinition: kartav1alpha1.StructureDefinition{
						RootComponent: kartav1alpha1.ComponentDefinition{
							Kind: &kartav1alpha1.GroupVersionKind{
								Version: "v1alpha1",
							},
						},
					},
				},
			}
			err := kartaHandler.HandleAddUpdate(karta)
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(ContainSubstring("failed to get backend ID"))
		})

		It("returns error when Kind is nil", func() {
			karta := &kartav1alpha1.Karta{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "karta-no-kind",
					Labels: map[string]string{v1alpha1.WorkloadTypeIdLabel: "karta"},
				},
				Spec: kartav1alpha1.KartaSpec{
					StructureDefinition: kartav1alpha1.StructureDefinition{
						RootComponent: kartav1alpha1.ComponentDefinition{Kind: nil},
					},
				},
			}
			_, err := parsers.ParseBackendKarta(karta, false)
			Expect(err).ToNot(BeNil())
		})

		It("Parse returns Deleting phase when DeletionTimestamp is set", func() {
			karta := &kartav1alpha1.Karta{
				ObjectMeta: metav1.ObjectMeta{
					Name:              "karta-1",
					Labels:            map[string]string{v1alpha1.WorkloadTypeIdLabel: kartaId},
					DeletionTimestamp: &metav1.Time{Time: time.Now()},
				},
				Spec: kartav1alpha1.KartaSpec{
					StructureDefinition: kartav1alpha1.StructureDefinition{
						RootComponent: kartav1alpha1.ComponentDefinition{
							Kind: &kartav1alpha1.GroupVersionKind{
								Version: "v1",
							},
						},
					},
				},
			}
			parsed, err := parsers.ParseBackendKarta(karta, false)
			Expect(err).To(BeNil())
			Expect(ptr.Deref(parsed.Phase, "")).To(Equal(parsers.KartaDeletingPhase))
		})

		It("Parse returns Deleting phase when isDeleted is true", func() {
			karta := &kartav1alpha1.Karta{
				ObjectMeta: metav1.ObjectMeta{
					Name:   "karta-1",
					Labels: map[string]string{v1alpha1.WorkloadTypeIdLabel: kartaId},
				},
				Spec: kartav1alpha1.KartaSpec{
					StructureDefinition: kartav1alpha1.StructureDefinition{
						RootComponent: kartav1alpha1.ComponentDefinition{
							Kind: &kartav1alpha1.GroupVersionKind{
								Version: "v1",
							},
						},
					},
				},
			}
			parsed, err := parsers.ParseBackendKarta(karta, true)
			Expect(err).To(BeNil())
			Expect(ptr.Deref(parsed.Phase, "")).To(Equal(parsers.KartaDeletedPhase))
		})

		It("test sync", func() {
			// use fake client to create 4 karta in a forloop
			workloadTypeIds := []string{}
			for i := 1; i <= 4; i++ {
				workloadTypeId := fmt.Sprintf("karta-%d-id", i)
				karta := buildKartaObj(fmt.Sprintf("karta-%d", i), workloadTypeId, "v1")
				err := fakeClient.Create(context.Background(), karta)
				Expect(err).To(BeNil())
				workloadTypeIds = append(workloadTypeIds, workloadTypeId)
			}

			wg := sync.WaitGroup{}
			wg.Add(1)
			kartaHandler.HandleSyncEvent(&wg)
			wg.Wait()

			Eventually(receivedSync).WithTimeout(3 * time.Second).Should(Receive(SatisfyAll(
				WithTransform(func(r openapi.WorkloadTypesForSync) bool { return r.FullSync != nil && *r.FullSync == true }, BeTrue()),
				WithTransform(func(r openapi.WorkloadTypesForSync) int { return len(r.WorkloadTypes) }, Equal(4)),
				WithTransform(func(r openapi.WorkloadTypesForSync) []string {
					result := make([]string, 0, len(r.WorkloadTypes))
					for _, wt := range r.WorkloadTypes {
						result = append(result, wt.GetId())
					}
					return result
				}, ConsistOf(workloadTypeIds)),
			)))
		})
	})
})

func buildKartaObj(name, id, version string) *kartav1alpha1.Karta {
	return &kartav1alpha1.Karta{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: map[string]string{v1alpha1.WorkloadTypeIdLabel: id},
		},
		Spec: kartav1alpha1.KartaSpec{
			StructureDefinition: kartav1alpha1.StructureDefinition{
				RootComponent: kartav1alpha1.ComponentDefinition{
					Kind: &kartav1alpha1.GroupVersionKind{
						Version: version,
					},
				},
			},
		},
	}
}
