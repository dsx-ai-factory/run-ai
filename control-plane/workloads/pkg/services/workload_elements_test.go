package services

import (
	contextpkg "context"
	"net/http"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	clusterApi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	metrics_client "github.com/run-ai/runai/control-plane/metrics-service/pkg/client"
	openapi "github.com/run-ai/runai/control-plane/metrics-service/pkg/client/generated"
	clientsclusterservice "github.com/run-ai/runai/control-plane/workloads-and-policies/clients/clusterservice"
	api "github.com/run-ai/runai/control-plane/workloads/internal/generated/pkg/api"
	bc_mocks "github.com/run-ai/runai/control-plane/workloads/mocks/pkg/clients/backend"
	"github.com/run-ai/runai/control-plane/workloads/pkg/clients/clusterservice"
	"github.com/run-ai/runai/control-plane/workloads/pkg/database/models"
	"github.com/run-ai/runai/control-plane/workloads/pkg/runaimetrics"
	testing "github.com/run-ai/runai/control-plane/workloads/pkg/test"
	checker "github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/checker"
	checkerClusterClient "github.com/run-ai/runai/runai-cluster/common/cluster-capabilities/clients/cluster"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"k8s.io/utils/ptr"
)

// helpers

func makeElement(id string, name string, level int32) api.WorkloadElement {
	return api.WorkloadElement{Id: id, Name: name, HierarchyLevel: level}
}

func makeElementWithK8sRef(id string, name string, level int32) api.WorkloadElement {
	return api.WorkloadElement{
		Id: id, Name: name, HierarchyLevel: level,
		KubernetesRef: &api.WorkloadElementKubernetesRef{
			Uid: uuid.NewString(), Name: "obj-" + name,
			Gvk: api.WorkloadElementKubernetesRefGvk{Group: "apps", Version: "v1", Kind: "Deployment"},
		},
	}
}

func elemSlice(elems ...api.WorkloadElement) *[]api.WorkloadElement {
	s := make([]api.WorkloadElement, len(elems))
	copy(s, elems)
	return &s
}

func createTestElements(count int) *[]api.WorkloadElement {
	elements := make([]api.WorkloadElement, 0, count)
	var parentID string
	for i := 0; i < count; i++ {
		level := int32((i % 3) + 1)
		e := api.WorkloadElement{
			Id:             uuid.NewString(),
			Name:           "element-" + uuid.NewString()[:8],
			HierarchyLevel: level,
		}
		if level > 1 && parentID != "" {
			e.ParentId = ptr.To(parentID)
		}
		if i%2 == 0 {
			e.InstanceKey = ptr.To("instance-" + uuid.NewString()[:4])
		}
		if i%3 == 0 {
			e.KubernetesRef = &api.WorkloadElementKubernetesRef{
				Uid:  uuid.NewString(),
				Name: "k8s-obj-" + uuid.NewString()[:4],
				Gvk:  api.WorkloadElementKubernetesRefGvk{Group: "apps", Version: "v1", Kind: "Deployment"},
			}
		}
		parentID = e.Id
		elements = append(elements, e)
	}
	return &elements
}

func fetchElementsFromDB(workloadID string) []models.WorkloadElement {
	var elements []models.WorkloadElement
	res := dao_utils.DB().WithContext(contextpkg.Background()).
		Table(models.WorkloadElementsTableName).
		Where("workload_id = ?", workloadID).
		Order("level, name").
		Find(&elements)
	Expect(res.Error).To(BeNil())
	return elements
}

func fetchElementIDsFromDB(workloadID string) []string {
	elems := fetchElementsFromDB(workloadID)
	ids := make([]string, len(elems))
	for i, e := range elems {
		ids[i] = e.ID
	}
	return ids
}

func fetchWorkloadFromDB(workloadID string) models.Workload {
	var row models.Workload
	res := dao_utils.DB().WithContext(contextpkg.Background()).
		Table("workloads").Where("id = ?", workloadID).First(&row)
	Expect(res.Error).To(BeNil())
	return row
}

func fetchPodFromDB(podID string) models.Pod {
	var row models.Pod
	res := dao_utils.DB().WithContext(contextpkg.Background()).
		Preload("Elements").
		Table("pods").Where("id = ?", podID).First(&row)
	Expect(res.Error).To(BeNil())
	return row
}

func findElementAtLevel(elements []*models.PodElement, level int) *models.PodElement {
	for _, e := range elements {
		if e.Level == level {
			return e
		}
	}
	return nil
}

func upsertWorkloadHelper(ctx contextpkg.Context, svc api.WorkloadsAPIServicer, workloadId string, w api.WorkloadSyncFields) {
	resp, err := svc.UpsertWorkload(ctx, workloadId, w)
	Expect(err).To(BeNil())
	Expect(resp.Code).To(Equal(http.StatusNoContent))
}

func upsertPodHelper(ctx contextpkg.Context, svc *PodsService, workloadId, podId, projectId, clusterId string, pod api.PodSyncFields) {
	// Ensure workload exists first (pod FK)
	resp, err := svc.UpsertPod(ctx, podId, pod)
	Expect(err).To(BeNil())
	Expect(resp.Code).To(Equal(http.StatusNoContent))
}

// ===================== Workload elements tests =====================

var _ = Describe("Structure elements on UpsertWorkload", Ordered, func() {
	ctx := testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
	workloadsService := NewWorkloadsService()

	AfterEach(func() {
		testing.CleanTables()
	})

	It("stores elements in DB when workload is upserted with structureElements", func() {
		workloadId := uuid.NewString()
		elements := createTestElements(3)
		w := SetupBaseWorkload(workloadId, DefaultProjectId, testing.ClusterUuid1)
		w.StructureElements = elements
		totalCount := int32(10)
		w.StructureElementsTotalCount = &totalCount

		upsertWorkloadHelper(ctx, workloadsService, workloadId, w)

		dbElements := fetchElementsFromDB(workloadId)
		Expect(len(dbElements)).To(Equal(3))
		for _, dbEl := range dbElements {
			Expect(dbEl.WorkloadID.String()).To(Equal(workloadId))
			Expect(dbEl.TenantID).To(Equal(uint(testing.TestTenantId)))
			Expect(dbEl.ClusterID.String()).To(Equal(testing.ClusterUuid1))
			Expect(dbEl.CreatedAt).ToNot(Equal(time.Time{}))
		}

		wRow := fetchWorkloadFromDB(workloadId)
		Expect(wRow.StructureElementsTotalCount).ToNot(BeNil())
		Expect(*wRow.StructureElementsTotalCount).To(Equal(10))
	})

	It("stores no elements when structureElements is nil", func() {
		workloadId := uuid.NewString()
		w := SetupBaseWorkload(workloadId, DefaultProjectId, testing.ClusterUuid1)
		w.StructureElements = nil

		upsertWorkloadHelper(ctx, workloadsService, workloadId, w)
		Expect(len(fetchElementsFromDB(workloadId))).To(Equal(0))
	})

	It("preserves k8sRef fields on elements", func() {
		workloadId := uuid.NewString()
		w := SetupBaseWorkload(workloadId, DefaultProjectId, testing.ClusterUuid1)

		k8sRefUID := uuid.NewString()
		w.StructureElements = &[]api.WorkloadElement{
			{
				Id: uuid.NewString(), Name: "with-k8s-ref", HierarchyLevel: 1,
				KubernetesRef: &api.WorkloadElementKubernetesRef{
					Uid: k8sRefUID, Name: "my-deployment",
					Gvk: api.WorkloadElementKubernetesRefGvk{Group: "apps", Version: "v1", Kind: "Deployment"},
				},
			},
			{Id: uuid.NewString(), Name: "logical-element", HierarchyLevel: 1, KubernetesRef: nil},
		}

		upsertWorkloadHelper(ctx, workloadsService, workloadId, w)

		dbElements := fetchElementsFromDB(workloadId)
		Expect(len(dbElements)).To(Equal(2))

		var withRef, withoutRef models.WorkloadElement
		for _, el := range dbElements {
			if el.Name == "with-k8s-ref" {
				withRef = el
			} else {
				withoutRef = el
			}
		}
		Expect(withRef.K8sRefUID).ToNot(BeNil())
		Expect(*withRef.K8sRefUID).To(Equal(k8sRefUID))
		Expect(*withRef.K8sRefName).To(Equal("my-deployment"))
		Expect(*withRef.GVKGroup).To(Equal("apps"))
		Expect(*withRef.GVKVersion).To(Equal("v1"))
		Expect(*withRef.GVKKind).To(Equal("Deployment"))
		Expect(withoutRef.K8sRefUID).To(BeNil())
		Expect(withoutRef.GVKKind).To(BeNil())
	})

	It("handles full element lifecycle: add, grow, shrink, replace, clear", func() {
		workloadId := uuid.NewString()
		w := SetupBaseWorkload(workloadId, DefaultProjectId, testing.ClusterUuid1)

		// Fixed element IDs to track across steps
		idA, idB, idC, idD := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
		idX, idY, idZ := uuid.NewString(), uuid.NewString(), uuid.NewString()
		elemA := makeElement(idA, "elem-A", 1)
		elemB := makeElement(idB, "elem-B", 1)
		elemC := makeElement(idC, "elem-C", 2)
		elemD := makeElementWithK8sRef(idD, "elem-D", 2)
		elemX := makeElement(idX, "elem-X", 1)
		elemY := makeElement(idY, "elem-Y", 2)
		elemZ := makeElement(idZ, "elem-Z", 3)

		// Step 1: Upsert with 3 elements (A, B, C)
		w.StructureElements = elemSlice(elemA, elemB, elemC)
		totalCount := int32(3)
		w.StructureElementsTotalCount = &totalCount
		upsertWorkloadHelper(ctx, workloadsService, workloadId, w)
		Expect(fetchElementIDsFromDB(workloadId)).To(ConsistOf(idA, idB, idC))
		Expect(*fetchWorkloadFromDB(workloadId).StructureElementsTotalCount).To(Equal(3))

		// Step 2: Upsert with 4 elements (A, B, C + D) — one added, three unchanged
		w.StructureElements = elemSlice(elemA, elemB, elemC, elemD)
		totalCount = int32(4)
		w.StructureElementsTotalCount = &totalCount
		upsertWorkloadHelper(ctx, workloadsService, workloadId, w)
		Expect(fetchElementIDsFromDB(workloadId)).To(ConsistOf(idA, idB, idC, idD))
		Expect(*fetchWorkloadFromDB(workloadId).StructureElementsTotalCount).To(Equal(4))

		// Step 3: Upsert with 2 elements (A, B) — C and D deleted
		w.StructureElements = elemSlice(elemA, elemB)
		totalCount = int32(2)
		w.StructureElementsTotalCount = &totalCount
		upsertWorkloadHelper(ctx, workloadsService, workloadId, w)
		Expect(fetchElementIDsFromDB(workloadId)).To(ConsistOf(idA, idB))

		// Step 4: Upsert with 3 NEW elements (X, Y, Z) — A and B deleted, X/Y/Z created
		w.StructureElements = elemSlice(elemX, elemY, elemZ)
		totalCount = int32(500)
		w.StructureElementsTotalCount = &totalCount
		upsertWorkloadHelper(ctx, workloadsService, workloadId, w)
		Expect(fetchElementIDsFromDB(workloadId)).To(ConsistOf(idX, idY, idZ))
		Expect(*fetchWorkloadFromDB(workloadId).StructureElementsTotalCount).To(Equal(500))

		// Step 5: Upsert with 0 elements — all deleted
		empty := []api.WorkloadElement{}
		w.StructureElements = &empty
		totalCount = int32(0)
		w.StructureElementsTotalCount = &totalCount
		upsertWorkloadHelper(ctx, workloadsService, workloadId, w)
		Expect(fetchElementsFromDB(workloadId)).To(BeEmpty())
		Expect(*fetchWorkloadFromDB(workloadId).StructureElementsTotalCount).To(Equal(0))
	})

	It("succeeds when request-level ClusterId is empty (regression: empty string passed to UUID column)", func() {
		// Regression: WorkloadsSyncRequest.ClusterId could be "" while per-workload
		// ClusterId was valid. DbSyncWorkloads handled empty clusterID gracefully
		// (ConstructDeleteQueryForProjectCluster skips empty filters), but
		// dbUpsertWorkloadElements passed "" directly to WHERE cluster_id = ?,
		// causing: ERROR: invalid input syntax for type uuid: ""
		id1 := uuid.NewString()
		w1 := SetupBaseWorkload(id1, DefaultProjectId, testing.ClusterUuid1)
		w1.Name = "regression-workload"
		w1.StructureElements = createTestElements(2)

		// Request with NO ClusterId — just like topology tests and some prod callers
		resp, err := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
			ProjectId: DefaultProjectId,
			Workloads: []api.WorkloadSyncFields{w1},
		})
		Expect(err).To(BeNil())
		Expect(resp.Code).To(SatisfyAny(Equal(http.StatusNoContent), Equal(http.StatusMultiStatus)))

		Expect(len(fetchElementsFromDB(id1))).To(Equal(2))
	})

	It("tracks StructureElementsTotalCount independently of actual elements", func() {
		workloadId := uuid.NewString()
		w := SetupBaseWorkload(workloadId, DefaultProjectId, testing.ClusterUuid1)

		// Send 2 elements but totalCount=500 (simulates truncation)
		w.StructureElements = createTestElements(2)
		totalCount := int32(500)
		w.StructureElementsTotalCount = &totalCount
		upsertWorkloadHelper(ctx, workloadsService, workloadId, w)
		Expect(len(fetchElementsFromDB(workloadId))).To(Equal(2))
		Expect(*fetchWorkloadFromDB(workloadId).StructureElementsTotalCount).To(Equal(500))

		// Update count to 100
		totalCount = int32(100)
		w.StructureElementsTotalCount = &totalCount
		upsertWorkloadHelper(ctx, workloadsService, workloadId, w)
		Expect(*fetchWorkloadFromDB(workloadId).StructureElementsTotalCount).To(Equal(100))

		// Nil totalCount
		w.StructureElementsTotalCount = nil
		upsertWorkloadHelper(ctx, workloadsService, workloadId, w)
		Expect(fetchWorkloadFromDB(workloadId).StructureElementsTotalCount).To(BeNil())
	})
})

// ===================== SyncWorkloads elements tests =====================

var _ = Describe("Structure elements on SyncWorkloads", Ordered, func() {
	ctx := testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
	workloadsService := NewWorkloadsService()

	AfterEach(func() {
		testing.CleanTables()
	})

	It("stores elements for multiple workloads in batch sync", func() {
		id1 := uuid.NewString()
		id2 := uuid.NewString()
		w1 := SetupBaseWorkload(id1, DefaultProjectId, testing.ClusterUuid1)
		w1.Name = "workload-1"
		w1.StructureElements = createTestElements(2)
		w2 := SetupBaseWorkload(id2, DefaultProjectId, testing.ClusterUuid1)
		w2.Name = "workload-2"
		w2.StructureElements = createTestElements(3)

		resp, err := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
			ProjectId: DefaultProjectId,
			ClusterId: testing.ClusterUuid1,
			Workloads: []api.WorkloadSyncFields{w1, w2},
		})
		Expect(err).To(BeNil())
		Expect(resp.Code).To(SatisfyAny(Equal(http.StatusNoContent), Equal(http.StatusMultiStatus)))

		Expect(len(fetchElementsFromDB(id1))).To(Equal(2))
		Expect(len(fetchElementsFromDB(id2))).To(Equal(3))
	})

	It("handles mix of workloads with and without elements", func() {
		id1 := uuid.NewString()
		id2 := uuid.NewString()
		w1 := SetupBaseWorkload(id1, DefaultProjectId, testing.ClusterUuid1)
		w1.Name = "with-elements"
		w1.StructureElements = createTestElements(2)
		w2 := SetupBaseWorkload(id2, DefaultProjectId, testing.ClusterUuid1)
		w2.Name = "without-elements"
		w2.StructureElements = nil

		resp, err := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
			ProjectId: DefaultProjectId,
			ClusterId: testing.ClusterUuid1,
			Workloads: []api.WorkloadSyncFields{w1, w2},
		})
		Expect(err).To(BeNil())
		Expect(resp.Code).To(SatisfyAny(Equal(http.StatusNoContent), Equal(http.StatusMultiStatus)))

		Expect(len(fetchElementsFromDB(id1))).To(Equal(2))
		Expect(len(fetchElementsFromDB(id2))).To(Equal(0))
	})

	It("deletes stale elements when workload is re-synced with empty elements", func() {
		id1 := uuid.NewString()
		w1 := SetupBaseWorkload(id1, DefaultProjectId, testing.ClusterUuid1)
		w1.Name = "workload-with-elements"
		w1.StructureElements = createTestElements(3)

		// First sync: 3 elements
		resp, err := workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
			ProjectId: DefaultProjectId,
			ClusterId: testing.ClusterUuid1,
			Workloads: []api.WorkloadSyncFields{w1},
		})
		Expect(err).To(BeNil())
		Expect(resp.Code).To(SatisfyAny(Equal(http.StatusNoContent), Equal(http.StatusMultiStatus)))
		Expect(len(fetchElementsFromDB(id1))).To(Equal(3))

		// Second sync: explicit empty elements — stale rows should be deleted
		w1.StructureElements = &[]api.WorkloadElement{}
		resp, err = workloadsService.SyncWorkloads(ctx, []string{}, api.WorkloadsSyncRequest{
			ProjectId: DefaultProjectId,
			ClusterId: testing.ClusterUuid1,
			Workloads: []api.WorkloadSyncFields{w1},
		})
		Expect(err).To(BeNil())
		Expect(resp.Code).To(SatisfyAny(Equal(http.StatusNoContent), Equal(http.StatusMultiStatus)))
		Expect(fetchElementsFromDB(id1)).To(BeEmpty())
	})
})

// ===================== Pod element IDs tests =====================

var _ = Describe("Pod element IDs", Ordered, func() {
	ctx := testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
	workloadsService := NewWorkloadsService()
	podsService := NewPodsService(&bc_mocks.TenantSettingsClient{})

	AfterEach(func() {
		testing.CleanTables()
	})

	setupWorkload := func() string {
		workloadId := uuid.NewString()
		upsertWorkloadHelper(ctx, workloadsService, workloadId, SetupBaseWorkload(workloadId, DefaultProjectId, testing.ClusterUuid1))
		return workloadId
	}

	Context("SyncWorkloadsPods", func() {
		It("stores all three element IDs on pod rows", func() {
			workloadId := setupWorkload()
			podId := uuid.NewString()
			pod := SetupBasePod(podId, workloadId, DefaultProjectId, testing.ClusterUuid1)
			pod.Elements = &[]api.PodElement{
				{ElementId: "elem-l1", HierarchyLevel: 1},
				{ElementId: "elem-l2", HierarchyLevel: 2},
				{ElementId: "elem-l3", HierarchyLevel: 3},
			}

			resp, _ := podsService.SyncWorkloadsPods(ctx, api.PodsSyncRequest{
				ProjectId: DefaultProjectId, ClusterId: testing.ClusterUuid1,
				Pods: []api.PodSyncFields{pod},
			})
			Expect(resp.Code).To(Equal(http.StatusNoContent))

			dbPod := fetchPodFromDB(podId)
			Expect(dbPod.Elements).To(HaveLen(3))
			Expect(findElementAtLevel(dbPod.Elements, 1).ElementID).To(Equal("elem-l1"))
			Expect(findElementAtLevel(dbPod.Elements, 2).ElementID).To(Equal("elem-l2"))
			Expect(findElementAtLevel(dbPod.Elements, 3).ElementID).To(Equal("elem-l3"))
		})

		It("stores no elements when not provided", func() {
			workloadId := setupWorkload()
			podId := uuid.NewString()
			pod := SetupBasePod(podId, workloadId, DefaultProjectId, testing.ClusterUuid1)

			resp, _ := podsService.SyncWorkloadsPods(ctx, api.PodsSyncRequest{
				ProjectId: DefaultProjectId, ClusterId: testing.ClusterUuid1,
				Pods: []api.PodSyncFields{pod},
			})
			Expect(resp.Code).To(Equal(http.StatusNoContent))

			dbPod := fetchPodFromDB(podId)
			Expect(dbPod.Elements).To(BeEmpty())
		})

		It("stores partial elements (level 1 and 2 only)", func() {
			workloadId := setupWorkload()
			podId := uuid.NewString()
			pod := SetupBasePod(podId, workloadId, DefaultProjectId, testing.ClusterUuid1)
			pod.Elements = &[]api.PodElement{
				{ElementId: "elem-l1", HierarchyLevel: 1},
				{ElementId: "elem-l2", HierarchyLevel: 2},
			}

			resp, _ := podsService.SyncWorkloadsPods(ctx, api.PodsSyncRequest{
				ProjectId: DefaultProjectId, ClusterId: testing.ClusterUuid1,
				Pods: []api.PodSyncFields{pod},
			})
			Expect(resp.Code).To(Equal(http.StatusNoContent))

			dbPod := fetchPodFromDB(podId)
			Expect(dbPod.Elements).To(HaveLen(2))
			Expect(findElementAtLevel(dbPod.Elements, 1).ElementID).To(Equal("elem-l1"))
			Expect(findElementAtLevel(dbPod.Elements, 2).ElementID).To(Equal("elem-l2"))
			Expect(findElementAtLevel(dbPod.Elements, 3)).To(BeNil())
		})
	})

	Context("UpsertPod", func() {
		It("stores elements via UpsertPod and returns them in ListPods", func() {
			workloadId := setupWorkload()
			podId := uuid.NewString()
			pod := SetupBasePod(podId, workloadId, DefaultProjectId, testing.ClusterUuid1)
			pod.Elements = &[]api.PodElement{
				{ElementId: "upsert-l1", HierarchyLevel: 1},
				{ElementId: "upsert-l2", HierarchyLevel: 2},
				{ElementId: "upsert-l3", HierarchyLevel: 3},
			}

			resp, err := podsService.UpsertPod(ctx, podId, pod)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusNoContent))

			// Verify via DB
			dbPod := fetchPodFromDB(podId)
			Expect(dbPod.Elements).To(HaveLen(3))
			Expect(findElementAtLevel(dbPod.Elements, 1).ElementID).To(Equal("upsert-l1"))
			Expect(findElementAtLevel(dbPod.Elements, 2).ElementID).To(Equal("upsert-l2"))
			Expect(findElementAtLevel(dbPod.Elements, 3).ElementID).To(Equal("upsert-l3"))

			// Verify via ListPods API (read path)
			listResp, err := podsService.ListPods(ctx, false, 0, 0, "", "", []string{"workloadId==" + workloadId}, "full", "", "")
			Expect(err).To(BeNil())
			Expect(listResp.Code).To(Equal(http.StatusOK))
			podsResponse := httpResponseBodyToPodsResponse(listResp)
			Expect(len(podsResponse.Pods)).To(Equal(1))
			Expect(podsResponse.Pods[0].Elements).ToNot(BeNil())
			Expect(*podsResponse.Pods[0].Elements).To(HaveLen(3))
		})

		It("stores partial elements via UpsertPod and returns only those in ListPods", func() {
			workloadId := setupWorkload()
			podId := uuid.NewString()
			pod := SetupBasePod(podId, workloadId, DefaultProjectId, testing.ClusterUuid1)
			pod.Elements = &[]api.PodElement{
				{ElementId: "only-l1", HierarchyLevel: 1},
			}

			resp, err := podsService.UpsertPod(ctx, podId, pod)
			Expect(err).To(BeNil())
			Expect(resp.Code).To(Equal(http.StatusNoContent))

			// Verify via ListPods API
			listResp, err := podsService.ListPods(ctx, false, 0, 0, "", "", []string{"workloadId==" + workloadId}, "full", "", "")
			Expect(err).To(BeNil())
			Expect(listResp.Code).To(Equal(http.StatusOK))
			podsResponse := httpResponseBodyToPodsResponse(listResp)
			Expect(len(podsResponse.Pods)).To(Equal(1))
			Expect(podsResponse.Pods[0].Elements).ToNot(BeNil())
			Expect(*podsResponse.Pods[0].Elements).To(HaveLen(1))
			Expect((*podsResponse.Pods[0].Elements)[0].ElementId).To(Equal("only-l1"))
			Expect((*podsResponse.Pods[0].Elements)[0].HierarchyLevel).To(Equal(int32(1)))
		})
	})

	Context("SyncPods", func() {
		It("stores elements via SyncPods", func() {
			workloadId := setupWorkload()
			podId := uuid.NewString()
			pod := SetupBasePod(podId, workloadId, DefaultProjectId, testing.ClusterUuid1)
			pod.Elements = &[]api.PodElement{
				{ElementId: "sync-l1", HierarchyLevel: 1},
				{ElementId: "sync-l2", HierarchyLevel: 2},
			}

			resp, err := podsService.SyncPods(ctx, []api.WorkloadPodsSyncRequest{
				{
					ProjectId:  DefaultProjectId,
					ClusterId:  testing.ClusterUuid1,
					WorkloadId: workloadId,
					Pods:       []api.PodSyncFields{pod},
				},
			})
			Expect(err).To(BeNil())
			Expect(resp.Code).To(SatisfyAny(Equal(http.StatusNoContent), Equal(http.StatusMultiStatus)))

			dbPod := fetchPodFromDB(podId)
			Expect(dbPod.Elements).To(HaveLen(2))
			Expect(findElementAtLevel(dbPod.Elements, 1).ElementID).To(Equal("sync-l1"))
			Expect(findElementAtLevel(dbPod.Elements, 2).ElementID).To(Equal("sync-l2"))
			Expect(findElementAtLevel(dbPod.Elements, 3)).To(BeNil())
		})
	})
})

// ===================== GetWorkloadElements API tests =====================

var _ = Describe("GetWorkloadElements", Ordered, func() {
	ctx := testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
	workloadsService := NewWorkloadsService()

	BeforeAll(func() {
		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(testing.ClusterUuid1, testing.BuildDisplayedCluster(testing.ClusterUuid1, testing.TestDefaultClusterVersion, nil)))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")
	})

	AfterEach(func() {
		testing.CleanTables()
	})

	It("returns elements ordered by level with all fields populated", func() {
		workloadId := uuid.NewString()
		elem1Id := uuid.NewString()
		elem2Id := uuid.NewString()
		elem3Id := uuid.NewString()
		elements := elemSlice(
			makeElement(elem2Id, "level2-elem", 2),
			makeElement(elem1Id, "level1-elem", 1),
			makeElementWithK8sRef(elem3Id, "level3-elem", 3),
		)
		w := SetupBaseWorkload(workloadId, DefaultProjectId, testing.ClusterUuid1)
		w.StructureElements = elements
		totalCount := int32(3)
		w.StructureElementsTotalCount = &totalCount
		upsertWorkloadHelper(ctx, workloadsService, workloadId, w)

		resp, err := workloadsService.GetWorkloadElements(ctx, workloadId, 0, 0)
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusOK))

		body := resp.Body.(api.WorkloadElementsResponse)
		Expect(body.Elements).To(HaveLen(3))

		// Verify ordering by level and all fields on each element
		e1 := body.Elements[0]
		Expect(e1.Id).To(Equal(elem1Id))
		Expect(e1.Name).To(Equal("level1-elem"))
		Expect(e1.HierarchyLevel).To(Equal(int32(1)))
		Expect(e1.ParentId).To(BeNil())
		Expect(e1.InstanceKey).To(BeNil())
		Expect(e1.ReplicaKey).To(BeNil())
		Expect(e1.CreatedAt).ToNot(BeZero())
		Expect(e1.KubernetesRef).To(BeNil())

		e2 := body.Elements[1]
		Expect(e2.Id).To(Equal(elem2Id))
		Expect(e2.Name).To(Equal("level2-elem"))
		Expect(e2.HierarchyLevel).To(Equal(int32(2)))
		Expect(e2.CreatedAt).ToNot(BeZero())

		e3 := body.Elements[2]
		Expect(e3.Id).To(Equal(elem3Id))
		Expect(e3.Name).To(Equal("level3-elem"))
		Expect(e3.HierarchyLevel).To(Equal(int32(3)))
		Expect(e3.CreatedAt).ToNot(BeZero())
		Expect(e3.KubernetesRef).ToNot(BeNil())
		Expect(e3.KubernetesRef.Uid).ToNot(BeEmpty())
		Expect(e3.KubernetesRef.Name).To(Equal("obj-level3-elem"))
		Expect(e3.KubernetesRef.Gvk.Group).To(Equal("apps"))
		Expect(e3.KubernetesRef.Gvk.Version).To(Equal("v1"))
		Expect(e3.KubernetesRef.Gvk.Kind).To(Equal("Deployment"))

		// Verify count endpoint
		countResp, err := workloadsService.CountWorkloadElements(ctx, workloadId)
		Expect(err).To(BeNil())
		Expect(countResp.Code).To(Equal(http.StatusOK))
		countBody := countResp.Body.(api.CountWorkloadElements200Response)
		Expect(countBody.TotalElementsCount).To(Equal(int32(3)))
		Expect(countBody.StoredElementsCount).To(Equal(int32(3)))
	})

	It("returns k8sRef fields correctly", func() {
		workloadId := uuid.NewString()
		elemWithRef := makeElementWithK8sRef(uuid.NewString(), "with-ref", 1)
		elemWithoutRef := makeElement(uuid.NewString(), "without-ref", 1)
		w := SetupBaseWorkload(workloadId, DefaultProjectId, testing.ClusterUuid1)
		w.StructureElements = elemSlice(elemWithRef, elemWithoutRef)
		totalCount := int32(2)
		w.StructureElementsTotalCount = &totalCount
		upsertWorkloadHelper(ctx, workloadsService, workloadId, w)

		resp, err := workloadsService.GetWorkloadElements(ctx, workloadId, 0, 0)
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusOK))

		body := resp.Body.(api.WorkloadElementsResponse)
		Expect(body.Elements).To(HaveLen(2))

		var withRef, withoutRef api.WorkloadElement
		for _, e := range body.Elements {
			if e.Name == "with-ref" {
				withRef = e
			} else {
				withoutRef = e
			}
		}
		Expect(withRef.KubernetesRef).ToNot(BeNil())
		Expect(withRef.KubernetesRef.Uid).ToNot(BeEmpty())
		Expect(withRef.KubernetesRef.Name).To(Equal("obj-with-ref"))
		Expect(withRef.KubernetesRef.Gvk.Group).To(Equal("apps"))
		Expect(withRef.KubernetesRef.Gvk.Version).To(Equal("v1"))
		Expect(withRef.KubernetesRef.Gvk.Kind).To(Equal("Deployment"))
		Expect(withRef.CreatedAt).ToNot(BeZero())
		Expect(withoutRef.KubernetesRef).To(BeNil())
		Expect(withoutRef.CreatedAt).ToNot(BeZero())
	})

	It("total reflects cluster count even when elements were truncated", func() {
		workloadId := uuid.NewString()
		w := SetupBaseWorkload(workloadId, DefaultProjectId, testing.ClusterUuid1)
		w.StructureElements = createTestElements(2)
		totalCount := int32(500)
		w.StructureElementsTotalCount = &totalCount
		upsertWorkloadHelper(ctx, workloadsService, workloadId, w)

		resp, err := workloadsService.GetWorkloadElements(ctx, workloadId, 0, 0)
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusOK))

		body := resp.Body.(api.WorkloadElementsResponse)
		Expect(body.Elements).To(HaveLen(2))

		// Verify count endpoint reflects cluster total vs stored
		countResp, err := workloadsService.CountWorkloadElements(ctx, workloadId)
		Expect(err).To(BeNil())
		countBody := countResp.Body.(api.CountWorkloadElements200Response)
		Expect(countBody.TotalElementsCount).To(Equal(int32(500)))
		Expect(countBody.StoredElementsCount).To(Equal(int32(2)))
	})

	It("returns empty list and total=0 for workload with no elements", func() {
		workloadId := uuid.NewString()
		w := SetupBaseWorkload(workloadId, DefaultProjectId, testing.ClusterUuid1)
		upsertWorkloadHelper(ctx, workloadsService, workloadId, w)

		resp, err := workloadsService.GetWorkloadElements(ctx, workloadId, 0, 0)
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusOK))

		body := resp.Body.(api.WorkloadElementsResponse)
		Expect(body.Elements).To(BeEmpty())

		// Verify count endpoint returns zeros
		countResp, err := workloadsService.CountWorkloadElements(ctx, workloadId)
		Expect(err).To(BeNil())
		countBody := countResp.Body.(api.CountWorkloadElements200Response)
		Expect(countBody.TotalElementsCount).To(Equal(int32(0)))
		Expect(countBody.StoredElementsCount).To(Equal(int32(0)))
	})

	It("returns optional fields (parentId, instanceKey, replicaKey) correctly", func() {
		workloadId := uuid.NewString()
		parentElemId := uuid.NewString()
		childElemId := uuid.NewString()
		parent := api.WorkloadElement{Id: parentElemId, Name: "parent", HierarchyLevel: 1, InstanceKey: ptr.To("Frontend")}
		child := api.WorkloadElement{Id: childElemId, Name: "child", HierarchyLevel: 2, ParentId: &parentElemId, ReplicaKey: ptr.To("replica-0")}
		w := SetupBaseWorkload(workloadId, DefaultProjectId, testing.ClusterUuid1)
		w.StructureElements = elemSlice(parent, child)
		totalCount := int32(2)
		w.StructureElementsTotalCount = &totalCount
		upsertWorkloadHelper(ctx, workloadsService, workloadId, w)

		resp, err := workloadsService.GetWorkloadElements(ctx, workloadId, 0, 0)
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusOK))

		body := resp.Body.(api.WorkloadElementsResponse)
		Expect(body.Elements).To(HaveLen(2))

		var parentResult, childResult api.WorkloadElement
		for _, e := range body.Elements {
			if e.Name == "parent" {
				parentResult = e
			} else {
				childResult = e
			}
		}
		Expect(parentResult.InstanceKey).ToNot(BeNil())
		Expect(*parentResult.InstanceKey).To(Equal("Frontend"))
		Expect(parentResult.ParentId).To(BeNil())
		Expect(parentResult.ReplicaKey).To(BeNil())

		Expect(childResult.ParentId).ToNot(BeNil())
		Expect(*childResult.ParentId).To(Equal(parentElemId))
		Expect(childResult.ReplicaKey).ToNot(BeNil())
		Expect(*childResult.ReplicaKey).To(Equal("replica-0"))
	})

	It("does not leak elements between workloads", func() {
		workloadId1 := uuid.NewString()
		workloadId2 := uuid.NewString()
		w1 := SetupBaseWorkload(workloadId1, DefaultProjectId, testing.ClusterUuid1)
		w1.Name = "workload-1"
		w1.StructureElements = elemSlice(makeElement(uuid.NewString(), "elem-w1", 1))
		tc1 := int32(1)
		w1.StructureElementsTotalCount = &tc1

		w2 := SetupBaseWorkload(workloadId2, DefaultProjectId, testing.ClusterUuid1)
		w2.Name = "workload-2"
		w2.StructureElements = elemSlice(
			makeElement(uuid.NewString(), "elem-w2-a", 1),
			makeElement(uuid.NewString(), "elem-w2-b", 2),
		)
		tc2 := int32(2)
		w2.StructureElementsTotalCount = &tc2

		upsertWorkloadHelper(ctx, workloadsService, workloadId1, w1)
		upsertWorkloadHelper(ctx, workloadsService, workloadId2, w2)

		resp1, err := workloadsService.GetWorkloadElements(ctx, workloadId1, 0, 0)
		Expect(err).To(BeNil())
		body1 := resp1.Body.(api.WorkloadElementsResponse)
		Expect(body1.Elements).To(HaveLen(1))
		Expect(body1.Elements[0].Name).To(Equal("elem-w1"))

		resp2, err := workloadsService.GetWorkloadElements(ctx, workloadId2, 0, 0)
		Expect(err).To(BeNil())
		body2 := resp2.Body.(api.WorkloadElementsResponse)
		Expect(body2.Elements).To(HaveLen(2))
	})

	It("reflects element changes after re-sync", func() {
		workloadId := uuid.NewString()
		elemA := makeElement(uuid.NewString(), "elem-A", 1)
		elemB := makeElement(uuid.NewString(), "elem-B", 2)
		w := SetupBaseWorkload(workloadId, DefaultProjectId, testing.ClusterUuid1)
		w.StructureElements = elemSlice(elemA, elemB)
		tc := int32(2)
		w.StructureElementsTotalCount = &tc
		upsertWorkloadHelper(ctx, workloadsService, workloadId, w)

		resp, err := workloadsService.GetWorkloadElements(ctx, workloadId, 0, 0)
		Expect(err).To(BeNil())
		body := resp.Body.(api.WorkloadElementsResponse)
		Expect(body.Elements).To(HaveLen(2))

		// Re-sync with only elemA — elemB should be removed
		w.StructureElements = elemSlice(elemA)
		tc = int32(1)
		w.StructureElementsTotalCount = &tc
		upsertWorkloadHelper(ctx, workloadsService, workloadId, w)

		resp, err = workloadsService.GetWorkloadElements(ctx, workloadId, 0, 0)
		Expect(err).To(BeNil())
		body = resp.Body.(api.WorkloadElementsResponse)
		Expect(body.Elements).To(HaveLen(1))
		Expect(body.Elements[0].Name).To(Equal("elem-A"))

		// Verify count reflects re-sync
		countResp, err := workloadsService.CountWorkloadElements(ctx, workloadId)
		Expect(err).To(BeNil())
		countBody := countResp.Body.(api.CountWorkloadElements200Response)
		Expect(countBody.TotalElementsCount).To(Equal(int32(1)))
		Expect(countBody.StoredElementsCount).To(Equal(int32(1)))
	})

	It("count: totalElementsCount defaults to storedElementsCount when StructureElementsTotalCount is nil", func() {
		workloadId := uuid.NewString()
		w := SetupBaseWorkload(workloadId, DefaultProjectId, testing.ClusterUuid1)
		w.StructureElements = createTestElements(3)
		w.StructureElementsTotalCount = nil
		upsertWorkloadHelper(ctx, workloadsService, workloadId, w)

		resp, err := workloadsService.GetWorkloadElements(ctx, workloadId, 0, 0)
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusOK))

		body := resp.Body.(api.WorkloadElementsResponse)
		Expect(body.Elements).To(HaveLen(3))

		countResp, err := workloadsService.CountWorkloadElements(ctx, workloadId)
		Expect(err).To(BeNil())
		Expect(countResp.Code).To(Equal(http.StatusOK))

		countBody := countResp.Body.(api.CountWorkloadElements200Response)
		Expect(countBody.TotalElementsCount).To(Equal(int32(3)))
		Expect(countBody.StoredElementsCount).To(Equal(int32(3)))
	})

	It("paginates with offset and limit", func() {
		workloadId := uuid.NewString()
		w := SetupBaseWorkload(workloadId, DefaultProjectId, testing.ClusterUuid1)
		w.StructureElements = createTestElements(5)
		totalCount := int32(5)
		w.StructureElementsTotalCount = &totalCount
		upsertWorkloadHelper(ctx, workloadsService, workloadId, w)

		// First page: limit=2
		resp, err := workloadsService.GetWorkloadElements(ctx, workloadId, 0, 2)
		Expect(err).To(BeNil())
		body := resp.Body.(api.WorkloadElementsResponse)
		Expect(body.Elements).To(HaveLen(2))
		Expect(body.Next).To(Equal(int32(2)))

		// Second page: offset=2, limit=2
		resp, err = workloadsService.GetWorkloadElements(ctx, workloadId, 2, 2)
		Expect(err).To(BeNil())
		body = resp.Body.(api.WorkloadElementsResponse)
		Expect(body.Elements).To(HaveLen(2))
		Expect(body.Next).To(Equal(int32(4)))

		// Third page: offset=4, limit=2 — only 1 left
		resp, err = workloadsService.GetWorkloadElements(ctx, workloadId, 4, 2)
		Expect(err).To(BeNil())
		body = resp.Body.(api.WorkloadElementsResponse)
		Expect(body.Elements).To(HaveLen(1))
		Expect(body.Next).To(Equal(int32(0))) // no more pages
	})

	It("returns 400 for invalid workload ID", func() {
		resp, err := workloadsService.GetWorkloadElements(ctx, "not-a-uuid", 0, 0)
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusBadRequest))

		countResp, err := workloadsService.CountWorkloadElements(ctx, "not-a-uuid")
		Expect(err).To(BeNil())
		Expect(countResp.Code).To(Equal(http.StatusBadRequest))
	})

	It("returns 404 for nonexistent workload", func() {
		missingId := uuid.NewString()
		resp, err := workloadsService.GetWorkloadElements(ctx, missingId, 0, 0)
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusNotFound))

		countResp, err := workloadsService.CountWorkloadElements(ctx, missingId)
		Expect(err).To(BeNil())
		Expect(countResp.Code).To(Equal(http.StatusNotFound))
	})

	It("returns 400 for workload in old cluster version", func() {
		// Set up a cluster with old version
		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(testing.ClusterUuid1, testing.BuildDisplayedCluster(testing.ClusterUuid1, testing.TestDefaultClusterVersion, nil)).
			WithCluster(testing.ClusterUuid2, testing.BuildDisplayedCluster(testing.ClusterUuid2, "2.22.0", nil)))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")

		workloadId := uuid.NewString()
		w := SetupBaseWorkload(workloadId, DefaultProjectId, testing.ClusterUuid2)
		w.StructureElements = createTestElements(2)
		totalCount := int32(2)
		w.StructureElementsTotalCount = &totalCount
		upsertWorkloadHelper(ctx, workloadsService, workloadId, w)

		resp, err := workloadsService.GetWorkloadElements(ctx, workloadId, 0, 0)
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusBadRequest))

		// Restore mock for subsequent tests
		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(testing.ClusterUuid1, testing.BuildDisplayedCluster(testing.ClusterUuid1, testing.TestDefaultClusterVersion, nil)))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")
	})

})

var _ = Describe("GetWorkloadElement", Ordered, func() {
	ctx := testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
	workloadsService := NewWorkloadsService()
	podsService := NewPodsService(&bc_mocks.TenantSettingsClient{})

	// Shared IDs populated in BeforeAll
	var workloadId string
	// Elements: parentElem (level 1, k8sRef, instanceKey), childElem (level 2, parentId, replicaKey),
	//           elemWithPods (level 1, 2 active pods + 1 completed pod), isolatedElem (level 1, 1 pod), noPodElem (level 1, no pods)
	var parentElemId, childElemId, elemWithPodsId, isolatedElemId, noPodElemId string

	// Topology fixture: pods 1+2 share host-1 so elemWithPods placement should be deepest level.
	// Pod 3 is completed and on a different rack — only running-pod placement should count.
	// The isolated pod sits in a different zone — it must not contaminate elemWithPods placement.
	const topologyName = "test-net-topology"
	topologyLevels := []string{"zone", "rack", "host"}

	BeforeAll(func() {
		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(testing.ClusterUuid1, testing.BuildDisplayedCluster(testing.ClusterUuid1, testing.TestDefaultClusterVersion, nil)))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")

		clusterservice.SetMock(clientsclusterservice.NewSimpleMock().
			WithTopology(topologyName, &clusterApi.NetworkTopology{Name: topologyName, Levels: topologyLevels}).
			WithNodeNetworkTopologyLabelValues(testing.ClusterUuid1, "node-active-1", []string{"zone-a", "rack-1", "host-1"}).
			WithNodeNetworkTopologyLabelValues(testing.ClusterUuid1, "node-active-2", []string{"zone-a", "rack-1", "host-1"}).
			WithNodeNetworkTopologyLabelValues(testing.ClusterUuid1, "node-completed", []string{"zone-a", "rack-2", "host-3"}).
			WithNodeNetworkTopologyLabelValues(testing.ClusterUuid1, "node-isolated", []string{"zone-b", "rack-9", "host-9"}))

		workloadId = uuid.NewString()
		parentElemId = uuid.NewString()
		childElemId = uuid.NewString()
		elemWithPodsId = uuid.NewString()
		isolatedElemId = uuid.NewString()
		noPodElemId = uuid.NewString()

		w := SetupBaseWorkload(workloadId, DefaultProjectId, testing.ClusterUuid1)
		w.Topology = &api.WorkloadSyncFieldsTopology{Name: topologyName}
		w.StructureElements = elemSlice(
			makeElementWithK8sRef(parentElemId, "parent-elem", 1),
			api.WorkloadElement{Id: childElemId, Name: "child-elem", HierarchyLevel: 2, ParentId: &parentElemId, ReplicaKey: ptr.To("replica-0")},
			makeElement(elemWithPodsId, "elem-with-pods", 1),
			makeElement(isolatedElemId, "isolated-elem", 1),
			makeElement(noPodElemId, "no-pod-elem", 1),
		)
		// Override instanceKey on parent after makeElementWithK8sRef
		elems := *w.StructureElements
		elems[0].InstanceKey = ptr.To("Frontend")
		tc := int32(5)
		w.StructureElementsTotalCount = &tc
		upsertWorkloadHelper(ctx, workloadsService, workloadId, w)

		// Per-pod resources:
		//   Requested: GPU portion limit=2 request=1, CPU limit=4 request=2, CPU memory limit=1Gi request=512Mi
		//   Allocated: GPU=2, GPU memory=4Gi, CPU=3.5, CPU memory=800Mi
		fullResources := &api.PodRequestResources{
			GpuRequestType: (*api.GpuRequestType)(ptr.To(string(api.GPUREQUESTTYPE_PORTION))),
			Gpu:            &api.RequestResourceCores{Limit: ptr.To(float64(2)), Request: ptr.To(float64(1))},
			Cpu:            &api.RequestResourceCores{Limit: ptr.To(float64(4)), Request: ptr.To(float64(2))},
			CpuMemory:      &api.RequestResourceQuantity{Limit: ptr.To("1Gi"), Request: ptr.To("512Mi")},
		}
		fullAllocated := &api.AllocatedResources{
			Gpu: ptr.To(float64(2)), GpuMemory: ptr.To("4Gi"),
			Cpu: ptr.To(float64(3.5)), CpuMemory: ptr.To("800Mi"),
		}

		pod1 := SetupBasePod(uuid.NewString(), workloadId, DefaultProjectId, testing.ClusterUuid1)
		pod1.Name = "active-pod-1"
		pod1.NodeName = ptr.To("node-active-1")
		pod1.RequestedResources = fullResources
		pod1.AllocatedResources = fullAllocated
		pod1.Elements = &[]api.PodElement{{ElementId: elemWithPodsId, HierarchyLevel: 1}}

		pod2 := SetupBasePod(uuid.NewString(), workloadId, DefaultProjectId, testing.ClusterUuid1)
		pod2.Name = "active-pod-2"
		pod2.NodeName = ptr.To("node-active-2")
		pod2.RequestedResources = fullResources
		pod2.AllocatedResources = fullAllocated
		pod2.Elements = &[]api.PodElement{{ElementId: elemWithPodsId, HierarchyLevel: 1}}

		pod3 := SetupBasePod(uuid.NewString(), workloadId, DefaultProjectId, testing.ClusterUuid1)
		pod3.Name = "completed-pod"
		pod3.NodeName = ptr.To("node-completed")
		pod3.K8sPhase = "Succeeded" // DB phase column is sourced from K8sPhase; non-Running so the LCA filter excludes it
		pod3.RequestedResources = fullResources
		pod3.AllocatedResources = fullAllocated
		pod3.Elements = &[]api.PodElement{{ElementId: elemWithPodsId, HierarchyLevel: 1}}
		completedAt := time.Now().Add(-1 * time.Hour)
		pod3.CompletedAt = &completedAt

		podIso := SetupBasePod(uuid.NewString(), workloadId, DefaultProjectId, testing.ClusterUuid1)
		podIso.Name = "isolated-pod"
		podIso.NodeName = ptr.To("node-isolated")
		podIso.RequestedResources = fullResources
		podIso.AllocatedResources = fullAllocated
		podIso.Elements = &[]api.PodElement{{ElementId: isolatedElemId, HierarchyLevel: 1}}

		// Sync all pods in a single call so SyncWorkloadsPods doesn't soft-delete earlier pods
		resp, err := podsService.SyncWorkloadsPods(ctx, api.PodsSyncRequest{
			ProjectId: DefaultProjectId, ClusterId: testing.ClusterUuid1,
			Pods: []api.PodSyncFields{pod1, pod2, pod3, podIso},
		})
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusNoContent))
	})

	AfterAll(func() {
		testing.CleanTables()
	})

	It("returns element with correct basic fields and parentId", func() {
		resp, err := workloadsService.GetWorkloadElement(ctx, workloadId, parentElemId)
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusOK))

		body := resp.Body.(api.WorkloadElementResponse)
		Expect(body.Id).To(Equal(parentElemId))
		Expect(body.Name).To(Equal("parent-elem"))
		Expect(body.HierarchyLevel).To(Equal(int32(1)))
		Expect(body.ParentId).To(BeNil())
		Expect(body.CreatedAt).ToNot(BeZero())
		// parentElem has no pods of its own → no actualPlacement, so topology is omitted entirely.
		Expect(body.Topology).To(BeNil())

		// Child has parentId set
		resp, err = workloadsService.GetWorkloadElement(ctx, workloadId, childElemId)
		Expect(err).To(BeNil())
		child := resp.Body.(api.WorkloadElementResponse)
		Expect(child.Id).To(Equal(childElemId))
		Expect(child.HierarchyLevel).To(Equal(int32(2)))
		Expect(child.ParentId).ToNot(BeNil())
		Expect(*child.ParentId).To(Equal(parentElemId))
		Expect(child.CreatedAt).ToNot(BeZero())
		Expect(child.Topology).To(BeNil())
	})

	It("returns k8sRef fields correctly", func() {
		// parentElem has k8sRef (created via makeElementWithK8sRef)
		resp, err := workloadsService.GetWorkloadElement(ctx, workloadId, parentElemId)
		Expect(err).To(BeNil())
		body := resp.Body.(api.WorkloadElementResponse)
		Expect(body.KubernetesRef).ToNot(BeNil())
		Expect(body.KubernetesRef.Uid).ToNot(BeEmpty())
		Expect(body.KubernetesRef.Name).To(Equal("obj-parent-elem"))
		Expect(body.KubernetesRef.Gvk.Group).To(Equal("apps"))
		Expect(body.KubernetesRef.Gvk.Version).To(Equal("v1"))
		Expect(body.KubernetesRef.Gvk.Kind).To(Equal("Deployment"))

		// noPodElem has no k8sRef
		resp, err = workloadsService.GetWorkloadElement(ctx, workloadId, noPodElemId)
		Expect(err).To(BeNil())
		body = resp.Body.(api.WorkloadElementResponse)
		Expect(body.KubernetesRef).To(BeNil())
	})

	It("returns optional fields (instanceKey, replicaKey) correctly", func() {
		resp, err := workloadsService.GetWorkloadElement(ctx, workloadId, parentElemId)
		Expect(err).To(BeNil())
		parentResult := resp.Body.(api.WorkloadElementResponse)
		Expect(parentResult.InstanceKey).ToNot(BeNil())
		Expect(*parentResult.InstanceKey).To(Equal("Frontend"))
		Expect(parentResult.ReplicaKey).To(BeNil())

		resp, err = workloadsService.GetWorkloadElement(ctx, workloadId, childElemId)
		Expect(err).To(BeNil())
		childResult := resp.Body.(api.WorkloadElementResponse)
		Expect(childResult.ReplicaKey).ToNot(BeNil())
		Expect(*childResult.ReplicaKey).To(Equal("replica-0"))
		Expect(childResult.InstanceKey).To(BeNil())
	})

	It("aggregates requested and allocated resources from element pods", func() {
		// elemWithPods has 2 active pods + 1 completed pod (completed excluded)
		// Per active pod: GPU limit=2/request=1, CPU limit=4/request=2, CPU mem limit=1Gi/request=512Mi
		//                 Allocated: GPU=2, GPU mem=4Gi, CPU=3.5, CPU mem=800Mi
		resp, err := workloadsService.GetWorkloadElement(ctx, workloadId, elemWithPodsId)
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusOK))

		body := resp.Body.(api.WorkloadElementResponse)

		// Requested resources (2 active pods summed)
		req := body.PodsRequestedResources
		Expect(req).ToNot(BeNil())
		Expect(*req.GpuRequestType).To(Equal(api.GPUREQUESTTYPE_PORTION))
		Expect(req.Gpu).ToNot(BeNil())
		Expect(*req.Gpu.Limit).To(Equal(float64(4)))   // 2+2
		Expect(*req.Gpu.Request).To(Equal(float64(2))) // 1+1
		Expect(req.GpuMemory).To(BeNil())              // portion type, no GPU memory
		Expect(req.Cpu).ToNot(BeNil())
		Expect(*req.Cpu.Limit).To(Equal(float64(8)))   // 4+4
		Expect(*req.Cpu.Request).To(Equal(float64(4))) // 2+2
		Expect(req.CpuMemory).ToNot(BeNil())
		Expect(*req.CpuMemory.Limit).To(Equal("2Gi"))   // 1Gi+1Gi
		Expect(*req.CpuMemory.Request).To(Equal("1Gi")) // 512Mi+512Mi
		Expect(req.ExtendedResources).To(BeNil())

		// Allocated resources (2 active pods summed)
		alloc := body.AllocatedResources
		Expect(alloc).ToNot(BeNil())
		Expect(*alloc.Gpu).To(Equal(float64(4)))     // 2+2
		Expect(*alloc.GpuMemory).To(Equal("8Gi"))    // 4Gi+4Gi
		Expect(*alloc.Cpu).To(Equal(float64(7)))     // 3.5+3.5
		Expect(*alloc.CpuMemory).To(Equal("1600Mi")) // 800Mi+800Mi

		// Both running pods share zone-a/rack-1/host-1 → deepest level wins.
		// The completed pod sits on a different rack and must not contribute.
		Expect(body.Topology).ToNot(BeNil())
		Expect(body.Topology.Name).To(Equal(topologyName))
		Expect(body.Topology.ActualPlacement).ToNot(BeNil())
		Expect(body.Topology.ActualPlacement.Level).To(Equal("host"))
		Expect(body.Topology.ActualPlacement.Value).To(Equal("host-1"))
	})

	It("excludes completed pods from both requested and allocated resources, and from actual placement", func() {
		// elemWithPods has 3 pods total: 2 active + 1 completed
		// If all 3 counted: GPU limit=6, CPU limit=12, allocated GPU=6, etc.
		// With completed excluded: only 2 active pods contribute
		resp, err := workloadsService.GetWorkloadElement(ctx, workloadId, elemWithPodsId)
		Expect(err).To(BeNil())

		body := resp.Body.(api.WorkloadElementResponse)
		Expect(*body.PodsRequestedResources.Gpu.Limit).To(Equal(float64(4)))   // 2+2, not 6
		Expect(*body.PodsRequestedResources.Gpu.Request).To(Equal(float64(2))) // 1+1, not 3
		Expect(*body.PodsRequestedResources.Cpu.Limit).To(Equal(float64(8)))   // 4+4, not 12
		Expect(*body.AllocatedResources.Gpu).To(Equal(float64(4)))             // 2+2, not 6
		Expect(*body.AllocatedResources.Cpu).To(Equal(float64(7)))             // 3.5+3.5, not 10.5

		// Completed pod is on rack-2/host-3; if it counted toward placement, the LCA would
		// shrink to {rack: ?} or {zone: zone-a} instead of staying at {host: host-1}.
		Expect(body.Topology).ToNot(BeNil())
		Expect(body.Topology.Name).To(Equal(topologyName))
		Expect(body.Topology.ActualPlacement).ToNot(BeNil())
		Expect(body.Topology.ActualPlacement.Level).To(Equal("host"))
		Expect(body.Topology.ActualPlacement.Value).To(Equal("host-1"))
	})

	It("returns nil resources when element has no pods", func() {
		resp, err := workloadsService.GetWorkloadElement(ctx, workloadId, noPodElemId)
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusOK))

		body := resp.Body.(api.WorkloadElementResponse)
		req := body.PodsRequestedResources
		Expect(req).ToNot(BeNil())
		Expect(req.Gpu).To(BeNil())
		Expect(req.GpuMemory).To(BeNil())
		Expect(req.Cpu).To(BeNil())
		Expect(req.CpuMemory).To(BeNil())
		Expect(req.ExtendedResources).To(BeNil())
		alloc := body.AllocatedResources
		Expect(alloc).ToNot(BeNil())
		Expect(alloc.Gpu).To(BeNil())
		Expect(alloc.GpuMemory).To(BeNil())
		Expect(alloc.Cpu).To(BeNil())
		Expect(alloc.CpuMemory).To(BeNil())
		// noPodElem has no pods → no actualPlacement, so topology is omitted entirely.
		Expect(body.Topology).To(BeNil())
	})

	It("only aggregates pods belonging to the requested element", func() {
		// isolatedElem has 1 pod, elemWithPods has 3 — resources must not leak
		resp, err := workloadsService.GetWorkloadElement(ctx, workloadId, isolatedElemId)
		Expect(err).To(BeNil())
		body := resp.Body.(api.WorkloadElementResponse)
		Expect(*body.PodsRequestedResources.Gpu.Limit).To(Equal(float64(2))) // only its 1 pod
		Expect(*body.PodsRequestedResources.Cpu.Limit).To(Equal(float64(4))) // only its 1 pod
		Expect(*body.AllocatedResources.Gpu).To(Equal(float64(2)))           // only its 1 pod
		Expect(*body.AllocatedResources.Cpu).To(Equal(float64(3.5)))         // only its 1 pod
		Expect(*body.AllocatedResources.GpuMemory).To(Equal("4Gi"))          // only its 1 pod

		// Isolated pod is on zone-b/rack-9/host-9 — its placement must reflect only that pod
		// and must not be polluted by elemWithPods's pods on zone-a.
		Expect(body.Topology).ToNot(BeNil())
		Expect(body.Topology.Name).To(Equal(topologyName))
		Expect(body.Topology.ActualPlacement).ToNot(BeNil())
		Expect(body.Topology.ActualPlacement.Level).To(Equal("host"))
		Expect(body.Topology.ActualPlacement.Value).To(Equal("host-9"))
	})

	It("returns 400 for invalid workload ID", func() {
		resp, err := workloadsService.GetWorkloadElement(ctx, "not-a-uuid", uuid.NewString())
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusBadRequest))
	})

	It("returns 400 for invalid element ID", func() {
		resp, err := workloadsService.GetWorkloadElement(ctx, workloadId, "not-a-uuid")
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusBadRequest))
	})

	It("returns 404 for nonexistent workload", func() {
		resp, err := workloadsService.GetWorkloadElement(ctx, uuid.NewString(), uuid.NewString())
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusNotFound))
	})

	It("returns 404 for nonexistent element within valid workload", func() {
		resp, err := workloadsService.GetWorkloadElement(ctx, workloadId, uuid.NewString())
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusNotFound))
	})

	It("returns 404 when element belongs to a different workload", func() {
		otherWorkloadId := uuid.NewString()
		foreignElemId := uuid.NewString()
		w := SetupBaseWorkload(otherWorkloadId, DefaultProjectId, testing.ClusterUuid1)
		w.StructureElements = elemSlice(makeElement(foreignElemId, "foreign-elem", 1))
		tc := int32(1)
		w.StructureElementsTotalCount = &tc
		upsertWorkloadHelper(ctx, workloadsService, otherWorkloadId, w)

		resp, err := workloadsService.GetWorkloadElement(ctx, workloadId, foreignElemId)
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusNotFound))
	})

	It("returns nil Topology when the workload has no topology configured", func() {
		// Separate workload, same cluster, but no topology on the workload at all.
		noTopoWorkloadId := uuid.NewString()
		noTopoElemId := uuid.NewString()
		w := SetupBaseWorkload(noTopoWorkloadId, DefaultProjectId, testing.ClusterUuid1)
		w.StructureElements = elemSlice(makeElement(noTopoElemId, "elem-no-topo", 1))
		tc := int32(1)
		w.StructureElementsTotalCount = &tc
		upsertWorkloadHelper(ctx, workloadsService, noTopoWorkloadId, w)

		// Pod on a fully-labeled node — placement must still be nil because the workload
		// itself has no topology configured.
		pod := SetupBasePod(uuid.NewString(), noTopoWorkloadId, DefaultProjectId, testing.ClusterUuid1)
		pod.NodeName = ptr.To("node-active-1")
		pod.Elements = &[]api.PodElement{{ElementId: noTopoElemId, HierarchyLevel: 1}}
		syncResp, syncErr := podsService.SyncWorkloadsPods(ctx, api.PodsSyncRequest{
			ProjectId: DefaultProjectId, ClusterId: testing.ClusterUuid1,
			Pods: []api.PodSyncFields{pod},
		})
		Expect(syncErr).To(BeNil())
		Expect(syncResp.Code).To(Equal(http.StatusNoContent))

		resp, err := workloadsService.GetWorkloadElement(ctx, noTopoWorkloadId, noTopoElemId)
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusOK))
		body := resp.Body.(api.WorkloadElementResponse)
		Expect(body.Topology).To(BeNil())
	})

	It("returns 400 for workload in old cluster version", func() {
		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(testing.ClusterUuid1, testing.BuildDisplayedCluster(testing.ClusterUuid1, testing.TestDefaultClusterVersion, nil)).
			WithCluster(testing.ClusterUuid2, testing.BuildDisplayedCluster(testing.ClusterUuid2, "2.22.0", nil)))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")

		oldWorkloadId := uuid.NewString()
		oldElemId := uuid.NewString()
		w := SetupBaseWorkload(oldWorkloadId, DefaultProjectId, testing.ClusterUuid2)
		w.StructureElements = elemSlice(makeElement(oldElemId, "elem", 1))
		tc := int32(1)
		w.StructureElementsTotalCount = &tc
		upsertWorkloadHelper(ctx, workloadsService, oldWorkloadId, w)

		resp, err := workloadsService.GetWorkloadElement(ctx, oldWorkloadId, oldElemId)
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusBadRequest))

		// Restore mock
		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(testing.ClusterUuid1, testing.BuildDisplayedCluster(testing.ClusterUuid1, testing.TestDefaultClusterVersion, nil)))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")
	})
})

// ===================== Actual placement when no topology constraints requested =====================

var _ = Describe("Actual placement without topology constraints", Ordered, func() {
	ctx := testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
	workloadsService := NewWorkloadsService()
	podsService := NewPodsService(&bc_mocks.TenantSettingsClient{})

	const topologyName = "no-constraint-topo"
	const nodepoolWithTopo = DefaultNodepool // "default" — matches SetupBasePod's CurrentNodePool
	const nodepoolNoTopo = "no-topo-pool"
	topologyLevels := []string{"zone", "rack", "host"}

	var (
		discoveredWorkloadId  string
		discoveredElemId      string
		nonTopologyWorkloadId string
	)

	BeforeAll(func() {
		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(testing.ClusterUuid1, testing.BuildDisplayedCluster(testing.ClusterUuid1, testing.TestDefaultClusterVersion, nil)))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")

		// One nodepool carries a NetworkTopologyName (so the API can discover topology even
		// when no constraint was requested) and one doesn't (so we cover the "no shared
		// topology among current nodepools" branch).
		clusterservice.SetMock(clientsclusterservice.NewSimpleMock().
			WithTopology(topologyName, &clusterApi.NetworkTopology{Name: topologyName, Levels: topologyLevels}).
			WithNodePoolsWithTopology(testing.ClusterUuid1, topologyName, nodepoolWithTopo).
			WithNodePools(testing.ClusterUuid1, nodepoolNoTopo).
			WithNodeNetworkTopologyLabelValues(testing.ClusterUuid1, "node-d-1", []string{"zone-a", "rack-7", "host-77"}).
			WithNodeNetworkTopologyLabelValues(testing.ClusterUuid1, "node-d-2", []string{"zone-a", "rack-7", "host-77"}))

		// Workload #1: no Topology constraint, pods running on topology-aware nodepool
		discoveredWorkloadId = uuid.NewString()
		discoveredElemId = uuid.NewString()
		w := SetupBaseWorkload(discoveredWorkloadId, DefaultProjectId, testing.ClusterUuid1)
		w.Topology = nil
		w.StructureElements = elemSlice(makeElement(discoveredElemId, "elem", 1))
		tc := int32(1)
		w.StructureElementsTotalCount = &tc
		upsertWorkloadHelper(ctx, workloadsService, discoveredWorkloadId, w)

		pod1 := SetupBasePod(uuid.NewString(), discoveredWorkloadId, DefaultProjectId, testing.ClusterUuid1)
		pod1.Name = "discovered-pod-1"
		pod1.NodeName = ptr.To("node-d-1")
		pod1.CurrentNodePool = ptr.To(nodepoolWithTopo)
		pod1.Elements = &[]api.PodElement{{ElementId: discoveredElemId, HierarchyLevel: 1}}

		pod2 := SetupBasePod(uuid.NewString(), discoveredWorkloadId, DefaultProjectId, testing.ClusterUuid1)
		pod2.Name = "discovered-pod-2"
		pod2.NodeName = ptr.To("node-d-2")
		pod2.CurrentNodePool = ptr.To(nodepoolWithTopo)
		pod2.Elements = &[]api.PodElement{{ElementId: discoveredElemId, HierarchyLevel: 1}}

		// Workload #2: no Topology constraint, pod runs on a nodepool with no NetworkTopologyName
		nonTopologyWorkloadId = uuid.NewString()
		w2 := SetupBaseWorkload(nonTopologyWorkloadId, DefaultProjectId, testing.ClusterUuid1)
		w2.Topology = nil
		upsertWorkloadHelper(ctx, workloadsService, nonTopologyWorkloadId, w2)

		pod3 := SetupBasePod(uuid.NewString(), nonTopologyWorkloadId, DefaultProjectId, testing.ClusterUuid1)
		pod3.Name = "non-topo-pod"
		pod3.NodeName = ptr.To("node-d-1")
		pod3.CurrentNodePool = ptr.To(nodepoolNoTopo)

		resp, err := podsService.SyncWorkloadsPods(ctx, api.PodsSyncRequest{
			ProjectId: DefaultProjectId, ClusterId: testing.ClusterUuid1,
			Pods: []api.PodSyncFields{pod1, pod2, pod3},
		})
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusNoContent))
	})

	AfterAll(func() {
		testing.CleanTables()
	})

	It("populates topology.name + actualPlacement on GET workload when nodepools carry a topology", func() {
		resp, err := workloadsService.GetWorkload(ctx, discoveredWorkloadId)
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusOK))

		wd := resp.Body.(api.WorkloadDetailed)
		Expect(wd.Topology).ToNot(BeNil())
		Expect(wd.Topology.Name).To(Equal(topologyName))
		// preferred/required are not synthesized — only requested constraints populate them
		Expect(wd.Topology.PreferredPlacement).To(BeEmpty())
		Expect(wd.Topology.RequiredPlacement).To(BeEmpty())
		Expect(wd.Topology.ActualPlacement).ToNot(BeNil())
		Expect(wd.Topology.ActualPlacement.Level).To(Equal("host"))
		Expect(wd.Topology.ActualPlacement.Value).To(Equal("host-77"))
	})

	It("populates element.topology.{name,actualPlacement} on GET element when nodepools carry a topology", func() {
		resp, err := workloadsService.GetWorkloadElement(ctx, discoveredWorkloadId, discoveredElemId)
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusOK))

		body := resp.Body.(api.WorkloadElementResponse)
		Expect(body.Topology).ToNot(BeNil())
		Expect(body.Topology.Name).To(Equal(topologyName))
		Expect(body.Topology.ActualPlacement).ToNot(BeNil())
		Expect(body.Topology.ActualPlacement.Level).To(Equal("host"))
		Expect(body.Topology.ActualPlacement.Value).To(Equal("host-77"))
	})

	It("leaves topology nil when no constraint AND current nodepool has no topology", func() {
		resp, err := workloadsService.GetWorkload(ctx, nonTopologyWorkloadId)
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusOK))

		wd := resp.Body.(api.WorkloadDetailed)
		Expect(wd.Topology).To(BeNil())
	})
})

// ===================== GetWorkloadElementMetrics API tests =====================

var _ = Describe("GetWorkloadElementMetrics", Ordered, func() {
	ctx := testing.SetContext(testing.TestTenant, testing.TestUser, []string{})
	workloadsService := NewWorkloadsService()

	const baseTimestampUnix = 1708958503653
	const numOfSamples = 10

	BeforeAll(func() {
		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(testing.ClusterUuid1, testing.BuildDisplayedCluster(testing.ClusterUuid1, testing.TestDefaultClusterVersion, nil)))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")

		metrics_client.SetMockMetricsServer(map[string]openapi.GetMetrics200Response{
			string(runaimetrics.ElementGpuUtilization):          getDefaultMetricMockResult(80, baseTimestampUnix, numOfSamples),
			string(runaimetrics.ElementGpuAllocation):           getConstMetricMockResult(4, baseTimestampUnix, numOfSamples),
			string(runaimetrics.ElementGpuMemoryUsedBytes):      getDefaultMetricMockResult(900, baseTimestampUnix, numOfSamples),
			string(runaimetrics.ElementGpuMemoryRequestedBytes): getConstMetricMockResult(1000, baseTimestampUnix, numOfSamples),
			string(runaimetrics.ElementCpuUsedCores):            getDefaultMetricMockResult(50, baseTimestampUnix, numOfSamples),
			string(runaimetrics.ElementCpuRequestedCores):       getConstMetricMockResult(1, baseTimestampUnix, numOfSamples),
			string(runaimetrics.ElementCpuLimitCores):           getConstMetricMockResult(2, baseTimestampUnix, numOfSamples),
			string(runaimetrics.ElementMemoryUsedBytes):         getDefaultMetricMockResult(150, baseTimestampUnix, numOfSamples),
			string(runaimetrics.ElementMemoryRequestedBytes):    getConstMetricMockResult(100, baseTimestampUnix, numOfSamples),
			string(runaimetrics.ElementMemoryLimitBytes):        getConstMetricMockResult(500, baseTimestampUnix, numOfSamples),
		})
	})

	AfterEach(func() {
		testing.CleanTables()
	})

	It("returns metrics for a valid element", func() {
		workloadId := uuid.NewString()
		elementId := uuid.NewString()
		elements := elemSlice(makeElement(elementId, "worker", 1))
		w := SetupBaseWorkload(workloadId, DefaultProjectId, testing.ClusterUuid1)
		w.StructureElements = elements
		totalCount := int32(1)
		w.StructureElementsTotalCount = &totalCount
		upsertWorkloadHelper(ctx, workloadsService, workloadId, w)

		resp, err := workloadsService.GetWorkloadElementMetrics(ctx, workloadId, elementId,
			[]api.ElementMetricType{
				api.ELEMENTMETRICTYPE_GPU_UTILIZATION,
				api.ELEMENTMETRICTYPE_CPU_USAGE_CORES,
				api.ELEMENTMETRICTYPE_CPU_MEMORY_USAGE_BYTES,
			},
			time.Now(), time.Now().Add(-5*time.Minute), 20)
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusOK))

		metrics, ok := resp.Body.(api.MetricsCompositeResponse)
		Expect(ok).To(BeTrue())
		Expect(metrics.Measurements).To(HaveLen(3))
		Expect(metrics.Measurements[0].Type).To(BeEquivalentTo(api.ELEMENTMETRICTYPE_GPU_UTILIZATION))
		Expect(metrics.Measurements[1].Type).To(BeEquivalentTo(api.ELEMENTMETRICTYPE_CPU_USAGE_CORES))
		Expect(metrics.Measurements[2].Type).To(BeEquivalentTo(api.ELEMENTMETRICTYPE_CPU_MEMORY_USAGE_BYTES))
		Expect(*metrics.Measurements[0].Values).To(HaveLen(numOfSamples))
	})

	It("returns all 10 metric types", func() {
		workloadId := uuid.NewString()
		elementId := uuid.NewString()
		elements := elemSlice(makeElement(elementId, "worker", 1))
		w := SetupBaseWorkload(workloadId, DefaultProjectId, testing.ClusterUuid1)
		w.StructureElements = elements
		totalCount := int32(1)
		w.StructureElementsTotalCount = &totalCount
		upsertWorkloadHelper(ctx, workloadsService, workloadId, w)

		allMetrics := []api.ElementMetricType{
			api.ELEMENTMETRICTYPE_GPU_UTILIZATION,
			api.ELEMENTMETRICTYPE_GPU_ALLOCATION,
			api.ELEMENTMETRICTYPE_GPU_MEMORY_USAGE_BYTES,
			api.ELEMENTMETRICTYPE_GPU_MEMORY_REQUEST_BYTES,
			api.ELEMENTMETRICTYPE_CPU_USAGE_CORES,
			api.ELEMENTMETRICTYPE_CPU_REQUEST_CORES,
			api.ELEMENTMETRICTYPE_CPU_LIMIT_CORES,
			api.ELEMENTMETRICTYPE_CPU_MEMORY_USAGE_BYTES,
			api.ELEMENTMETRICTYPE_CPU_MEMORY_REQUEST_BYTES,
			api.ELEMENTMETRICTYPE_CPU_MEMORY_LIMIT_BYTES,
		}

		resp, err := workloadsService.GetWorkloadElementMetrics(ctx, workloadId, elementId,
			allMetrics, time.Now(), time.Now().Add(-5*time.Minute), 20)
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusOK))

		metrics, ok := resp.Body.(api.MetricsCompositeResponse)
		Expect(ok).To(BeTrue())
		Expect(metrics.Measurements).To(HaveLen(10))
		for i, mt := range allMetrics {
			Expect(metrics.Measurements[i].Type).To(BeEquivalentTo(mt))
			Expect(*metrics.Measurements[i].Values).To(HaveLen(numOfSamples))
		}
	})

	It("returns 400 for invalid workload UUID", func() {
		resp, err := workloadsService.GetWorkloadElementMetrics(ctx, "not-a-uuid", uuid.NewString(),
			[]api.ElementMetricType{api.ELEMENTMETRICTYPE_GPU_UTILIZATION},
			time.Now(), time.Now().Add(-5*time.Minute), 10)
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusBadRequest))
		Expect(resp.Body.(api.Error).Message).To(ContainSubstring("Invalid workload id"))
	})

	It("returns 400 for invalid element UUID", func() {
		workloadId := uuid.NewString()
		w := SetupBaseWorkload(workloadId, DefaultProjectId, testing.ClusterUuid1)
		upsertWorkloadHelper(ctx, workloadsService, workloadId, w)

		resp, err := workloadsService.GetWorkloadElementMetrics(ctx, workloadId, "not-a-uuid",
			[]api.ElementMetricType{api.ELEMENTMETRICTYPE_GPU_UTILIZATION},
			time.Now(), time.Now().Add(-5*time.Minute), 10)
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusBadRequest))
		Expect(resp.Body.(api.Error).Message).To(ContainSubstring("Invalid element id"))
	})

	It("returns 404 for nonexistent workload", func() {
		resp, err := workloadsService.GetWorkloadElementMetrics(ctx, uuid.NewString(), uuid.NewString(),
			[]api.ElementMetricType{api.ELEMENTMETRICTYPE_GPU_UTILIZATION},
			time.Now(), time.Now().Add(-5*time.Minute), 10)
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusNotFound))
	})

	It("returns 400 for workload in old cluster version", func() {
		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(testing.ClusterUuid1, testing.BuildDisplayedCluster(testing.ClusterUuid1, testing.TestDefaultClusterVersion, nil)).
			WithCluster(testing.ClusterUuid2, testing.BuildDisplayedCluster(testing.ClusterUuid2, "2.22.0", nil)))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")

		workloadId := uuid.NewString()
		w := SetupBaseWorkload(workloadId, DefaultProjectId, testing.ClusterUuid2)
		upsertWorkloadHelper(ctx, workloadsService, workloadId, w)

		resp, err := workloadsService.GetWorkloadElementMetrics(ctx, workloadId, uuid.NewString(),
			[]api.ElementMetricType{api.ELEMENTMETRICTYPE_GPU_UTILIZATION},
			time.Now(), time.Now().Add(-5*time.Minute), 10)
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusBadRequest))
		Expect(resp.Body.(api.Error).Message).To(ContainSubstring("not supported"))

		// Restore mock
		checkerClusterClient.SetMock(checkerClusterClient.NewSimpleMockClient().
			WithCluster(testing.ClusterUuid1, testing.BuildDisplayedCluster(testing.ClusterUuid1, testing.TestDefaultClusterVersion, nil)))
		checker.InitClusterCapabilitiesChecker("", testing.GetDefaultClustersCacheConfig(), "")
	})

	It("returns 404 for nonexistent element", func() {
		workloadId := uuid.NewString()
		w := SetupBaseWorkload(workloadId, DefaultProjectId, testing.ClusterUuid1)
		upsertWorkloadHelper(ctx, workloadsService, workloadId, w)

		resp, err := workloadsService.GetWorkloadElementMetrics(ctx, workloadId, uuid.NewString(),
			[]api.ElementMetricType{api.ELEMENTMETRICTYPE_GPU_UTILIZATION},
			time.Now(), time.Now().Add(-5*time.Minute), 10)
		Expect(err).To(BeNil())
		Expect(resp.Code).To(Equal(http.StatusNotFound))
	})
})
