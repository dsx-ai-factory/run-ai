/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"errors"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	ext_api "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

const (
	podsCountBasic      = `SELECT count(*) FROM pods WHERE tenant_id = ? AND pods.deleted_at IS NULL`
	podsCountDeleted    = `SELECT count(*) FROM pods WHERE tenant_id = ? AND pods.deleted_at IS NOT NULL`
	podsCountCompletedT = `SELECT count(*) FROM pods WHERE tenant_id = ? AND pods.deleted_at IS NULL AND pods.completed_at IS NOT NULL`
	podsCountCompletedF = `SELECT count(*) FROM pods WHERE tenant_id = ? AND pods.deleted_at IS NULL AND pods.completed_at IS NULL`
	podsCountScopes     = `SELECT count(*) FROM pods WHERE tenant_id = ? AND ((false OR pods.cluster_id IN ( 'c1') OR pods.project_id IN ( 'p1'))) AND pods.deleted_at IS NULL`
	podsCountDept       = `SELECT count(*) FROM pods LEFT JOIN workloads as w on w.tenant_id = pods.tenant_id AND w.cluster_id = pods.cluster_id AND w.id = pods.workload_id WHERE pods.tenant_id = ? AND ((false OR pods.tenant_id IN ( '2') OR w.department_id IN ( 'd1'))) AND pods.deleted_at IS NULL`
	podsCountElement    = `SELECT COUNT(DISTINCT(pods.id)) FROM pods LEFT JOIN pod_elements ON pod_elements.pod_id = pods.id WHERE tenant_id = ? AND pods.deleted_at IS NULL AND pod_elements.element_id = ?`
	podsCountSearch     = `SELECT count(*) FROM pods WHERE tenant_id = ? AND pods.deleted_at IS NULL AND (search_vector @@ to_tsquery('english', ? || ':*'))`

	podsListBrief    = `SELECT * FROM pods WHERE pods.tenant_id = ? AND pods.deleted_at IS NULL ORDER BY id DESC`
	podsListDeleted  = `SELECT * FROM pods WHERE pods.tenant_id = ? AND deleted_at IS NOT NULL ORDER BY id DESC`
	podsListClusters = `SELECT * FROM pods WHERE pods.tenant_id = ? AND ((false OR pods.cluster_id IN ( 'c1'))) AND pods.deleted_at IS NULL ORDER BY id DESC`
	podsListFull     = `SELECT pods.*, w.name as workload_name, w.project_name as project_name FROM pods LEFT JOIN workloads as w on w.tenant_id = pods.tenant_id AND w.cluster_id = pods.cluster_id AND w.id = pods.workload_id WHERE pods.tenant_id = ? AND pods.deleted_at IS NULL ORDER BY id DESC`
	podsListDept     = `SELECT pods.id,pods.created_at,pods.updated_at,pods.completed_at,pods.deleted_at,pods.workload_id,pods.revision_id,pods.name,pods.priority_class_name,pods.tenant_id,pods.cluster_id,pods.project_id,pods.node_name,pods.current_nodepool,pods.requested_nodepools,pods.requested_gpu_limit,pods.requested_gpu_request,pods.requested_gpu_memory_limit,pods.requested_gpu_memory_request,pods.requested_cpu_limit,pods.requested_cpu_request,pods.requested_cpu_memory_limit,pods.requested_cpu_memory_request,pods.requested_mig_device,pods.allocated_gpu,pods.allocated_gpu_memory,pods.allocated_cpu,pods.allocated_cpu_memory,pods.allocated_mig_device,pods.phase,pods.phase_updated_at,pods.tolerations,pods.ip,pods.task,pods.node_affinity,pods.network_topology_label_values,pods.resource_claims FROM pods LEFT JOIN workloads as w on w.tenant_id = pods.tenant_id AND w.cluster_id = pods.cluster_id AND w.id = pods.workload_id WHERE pods.tenant_id = ? AND ((false OR pods.tenant_id IN ( '2') OR w.department_id IN ( 'd1'))) AND pods.deleted_at IS NULL ORDER BY id DESC`
	podsListPaged    = `SELECT DISTINCT pods.* FROM pods LEFT JOIN pod_elements ON pod_elements.pod_id = pods.id WHERE pods.tenant_id = ? AND pods.completed_at IS NOT NULL AND pod_elements.element_id = ? AND (search_vector @@ to_tsquery('english', ? || ':*')) AND pods.deleted_at IS NULL ORDER BY pods.name DESC NULLS LAST,id DESC LIMIT ? OFFSET ?`

	podsRevisionBasic   = `SELECT * FROM pods WHERE tenant_id = ? AND revision_id = ? AND pods.deleted_at IS NULL`
	podsRevisionDeleted = `SELECT * FROM pods WHERE tenant_id = ? AND revision_id = ? AND deleted_at IS NOT NULL`
	podsRevisionPaged   = `SELECT * FROM pods WHERE tenant_id = ? AND revision_id = ? AND (search_vector @@ to_tsquery('english', ? || ':*')) AND pods.deleted_at IS NULL LIMIT ? OFFSET ?`

	podsGetPod          = `SELECT * FROM pods WHERE (pods.tenant_id = ? AND pods.workload_id = ? AND pods.id = ?) AND pods.deleted_at IS NULL ORDER BY pods.id LIMIT ?`
	podsGetContainers   = `SELECT * FROM containers WHERE containers.pod_id = ?`
	podsGetElements     = `SELECT * FROM pod_elements WHERE pod_elements.pod_id = ?`
	podsGetPodClusterId = `SELECT cluster_id FROM pods WHERE tenant_id = ? AND workload_id = ? AND id = ? AND pods.deleted_at IS NULL ORDER BY pods.id LIMIT ?`

	podsDelete       = `UPDATE pods SET deleted_at=? WHERE tenant_id = ? AND pods.id = ? AND pods.deleted_at IS NULL`
	podsUpdateByNode = `UPDATE pods SET network_topology_label_values=? WHERE tenant_id = ? AND cluster_id = ? AND node_name = ? AND phase = 'Running' AND deleted_at IS NULL`
	podsUpsertUpdate = `UPDATE pods SET created_at=?,updated_at=?,completed_at=?,deleted_at=?,workload_id=?,revision_id=?,name=?,priority_class_name=?,tenant_id=?,cluster_id=?,project_id=?,node_name=?,current_nodepool=?,requested_nodepools=?,requested_gpu_limit=?,requested_gpu_request=?,requested_gpu_memory_limit=?,requested_gpu_memory_request=?,requested_cpu_limit=?,requested_cpu_request=?,requested_cpu_memory_limit=?,requested_cpu_memory_request=?,requested_mig_device=?,allocated_gpu=?,allocated_gpu_memory=?,allocated_cpu=?,allocated_cpu_memory=?,allocated_mig_device=?,phase=?,phase_updated_at=?,tolerations=?,ip=?,task=?,node_affinity=?,network_topology_label_values=?,resource_claims=? WHERE pods.deleted_at IS NULL AND id = ?`

	podsSyncDelete           = `UPDATE pods SET deleted_at=? WHERE tenant_id = ? AND cluster_id = ? AND pods.deleted_at IS NULL`
	podsSyncDeleteExcluding  = `UPDATE pods SET deleted_at=? WHERE tenant_id = ? AND cluster_id = ? AND NOT id IN (?) AND pods.deleted_at IS NULL`
	podsSyncDeleteExcluding2 = `UPDATE pods SET deleted_at=? WHERE tenant_id = ? AND cluster_id = ? AND NOT id IN (?,?) AND pods.deleted_at IS NULL`
	podsSavepoint            = `SAVEPOINT sp`
	podsRollbackTo           = `ROLLBACK TO SAVEPOINT sp`

	podsInsertContainer = `INSERT INTO containers (started_at,cluster_id,pod_id,name,image) VALUES (?,?,?,?,?) ON CONFLICT (cluster_id,pod_id,name) DO UPDATE SET started_at=excluded.started_at,image=excluded.image`
	podsInsertElement   = `INSERT INTO pod_elements (pod_id,element_id,level) VALUES (?,?,?) ON CONFLICT (pod_id,level) DO NOTHING`

	podsWorkloadLca = `WITH running_pods AS ( SELECT network_topology_label_values FROM pods WHERE tenant_id = ? AND workload_id = ? AND deleted_at IS NULL AND phase = 'Running' ), bounds AS ( SELECT MIN(network_topology_label_values) as min_path, MAX(network_topology_label_values) as max_path FROM running_pods ), mismatch AS ( SELECT i FROM bounds, generate_series(1, array_length(min_path, 1)) as i WHERE min_path[i] IS DISTINCT FROM max_path[i] ORDER BY i ASC LIMIT 1 ) SELECT CASE WHEN EXISTS( SELECT 1 FROM running_pods WHERE network_topology_label_values IS NULL OR array_length(network_topology_label_values, 1) IS NULL ) THEN NULL WHEN (SELECT i FROM mismatch) IS NULL THEN (SELECT min_path FROM bounds) ELSE NULLIF((SELECT min_path[1 : i-1] FROM bounds, mismatch), '{}'::text[]) END as lca_topology FROM bounds;`
	podsElementLca  = `WITH running_pods AS ( SELECT network_topology_label_values FROM pods WHERE tenant_id = ? AND workload_id = ? AND deleted_at IS NULL AND phase = 'Running' AND id IN (SELECT pod_id FROM pod_elements WHERE element_id = ?) ), bounds AS ( SELECT MIN(network_topology_label_values) as min_path, MAX(network_topology_label_values) as max_path FROM running_pods ), mismatch AS ( SELECT i FROM bounds, generate_series(1, array_length(min_path, 1)) as i WHERE min_path[i] IS DISTINCT FROM max_path[i] ORDER BY i ASC LIMIT 1 ) SELECT CASE WHEN EXISTS( SELECT 1 FROM running_pods WHERE network_topology_label_values IS NULL OR array_length(network_topology_label_values, 1) IS NULL ) THEN NULL WHEN (SELECT i FROM mismatch) IS NULL THEN (SELECT min_path FROM bounds) ELSE NULLIF((SELECT min_path[1 : i-1] FROM bounds, mismatch), '{}'::text[]) END as lca_topology FROM bounds;`

	podsTestClusterIdStr  = "11111111-1111-1111-1111-111111111111"
	podsTestPodIdStr      = "22222222-2222-2222-2222-222222222222"
	podsTestPodId2Str     = "44444444-4444-4444-4444-444444444444"
	podsTestWorkloadIdStr = "33333333-3333-3333-3333-333333333333"
)

type podRow struct {
	ID string
}

type podClusterIdRow struct {
	ClusterID string
}

type podLcaRow struct {
	LcaTopology string
}

func podsBoolPtr(v bool) *bool {
	return &v
}

func podsElementFilter() []queryapi.Filter {
	return []queryapi.Filter{
		{Field: "pod_elements.element_id", Operator: "=", Values: []string{"el-1"}},
	}
}

func podsNameFilter() []queryapi.Filter {
	return []queryapi.Filter{
		{Field: "pods.name", Operator: "=", Values: []string{"p1"}},
	}
}

func podsTestPod() models.Pod {
	return models.Pod{
		ID:         uuid.MustParse(podsTestPodIdStr),
		WorkloadID: uuid.MustParse(podsTestWorkloadIdStr),
		ClusterID:  uuid.MustParse(podsTestClusterIdStr),
		TenantID:   1,
		Name:       "pod-1",
		Phase:      "Running",
	}
}

func podsTestContainer() models.Container {
	return models.Container{
		ClusterID: uuid.MustParse(podsTestClusterIdStr),
		PodID:     uuid.MustParse(podsTestPodIdStr),
		Name:      "c1",
		Image:     "img",
	}
}

func podsTestElement() models.PodElement {
	return models.PodElement{
		PodID:     uuid.MustParse(podsTestPodIdStr),
		ElementID: "el-1",
		Level:     0,
	}
}

var _ = Describe("DbCountPods", func() {
	It("counts pods with no filters", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsCountBasic).Returns([]countRow{{Count: 3}})

		count, apiErr := DbCountPods(testCtx(1), 1, false, nil, "", "", nil)

		Expect(apiErr).To(BeNil())
		Expect(count).To(Equal(int64(3)))
	})

	It("counts deleted pods", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsCountDeleted).Returns([]countRow{{Count: 1}})

		count, apiErr := DbCountPods(testCtx(1), 1, true, nil, "", "", nil)

		Expect(apiErr).To(BeNil())
		Expect(count).To(Equal(int64(1)))
	})

	It("counts completed pods when completed is true", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsCountCompletedT).Returns([]countRow{{Count: 2}})

		count, apiErr := DbCountPods(testCtx(1), 1, false, nil, "", "TRUE", nil)

		Expect(apiErr).To(BeNil())
		Expect(count).To(Equal(int64(2)))
	})

	It("counts non completed pods when completed is false", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsCountCompletedF).Returns([]countRow{{Count: 2}})

		count, apiErr := DbCountPods(testCtx(1), 1, false, nil, "", "false", nil)

		Expect(apiErr).To(BeNil())
		Expect(count).To(Equal(int64(2)))
	})

	It("counts all pods for system permitted scopes", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsCountBasic).Returns([]countRow{{Count: 5}})

		scopes := &openapi.PermittedScopes{System: podsBoolPtr(true)}
		count, apiErr := DbCountPods(testCtx(1), 1, false, nil, "", "", scopes)

		Expect(apiErr).To(BeNil())
		Expect(count).To(Equal(int64(5)))
	})

	It("counts all pods for tenant permitted scopes", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsCountBasic).Returns([]countRow{{Count: 5}})

		scopes := &openapi.PermittedScopes{System: podsBoolPtr(false), Tenants: []string{"1"}}
		count, apiErr := DbCountPods(testCtx(1), 1, false, nil, "", "", scopes)

		Expect(apiErr).To(BeNil())
		Expect(count).To(Equal(int64(5)))
	})

	It("filters by cluster and project scopes", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsCountScopes).Returns([]countRow{{Count: 4}})

		scopes := &openapi.PermittedScopes{Tenants: []string{"2"}, Clusters: []string{"c1"}, Projects: []string{"p1"}}
		count, apiErr := DbCountPods(testCtx(1), 1, false, nil, "", "", scopes)

		Expect(apiErr).To(BeNil())
		Expect(count).To(Equal(int64(4)))
	})

	It("filters by department scopes", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsCountDept).Returns([]countRow{{Count: 6}})

		scopes := &openapi.PermittedScopes{Tenants: []string{"2"}, Departments: []string{"d1"}}
		count, apiErr := DbCountPods(testCtx(1), 1, false, nil, "", "", scopes)

		Expect(apiErr).To(BeNil())
		Expect(count).To(Equal(int64(6)))
	})

	It("joins pod elements when an element filter is given", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsCountElement).Returns([]countRow{{Count: 7}})

		count, apiErr := DbCountPods(testCtx(1), 1, false, podsElementFilter(), "", "", nil)

		Expect(apiErr).To(BeNil())
		Expect(count).To(Equal(int64(7)))
	})

	It("applies a free text search", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsCountSearch).Returns([]countRow{{Count: 8}})

		count, apiErr := DbCountPods(testCtx(1), 1, false, nil, "abc", "", nil)

		Expect(apiErr).To(BeNil())
		Expect(count).To(Equal(int64(8)))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsCountBasic).ReturnsError(errors.New("boom"))

		count, apiErr := DbCountPods(testCtx(1), 1, false, nil, "", "", nil)

		Expect(apiErr).ToNot(BeNil())
		Expect(count).To(Equal(int64(0)))
	})
})

var _ = Describe("DbListPods", func() {
	It("lists pods with brief verbosity", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsListBrief).Returns([]podRow{})

		result, apiErr := DbListPods(testCtx(1), 1, false, 0, 0, "", "", nil, ext_api.VerbosityBrief, "", "", nil)

		Expect(apiErr).To(BeNil())
		pods, ok := result.(*[]models.Pod)
		Expect(ok).To(BeTrue())
		Expect(*pods).To(BeEmpty())
	})

	It("lists pods with full verbosity", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsListFull).Returns([]podRow{})

		result, apiErr := DbListPods(testCtx(1), 1, false, 0, 0, "", "", nil, ext_api.VerbosityFull, "", "", nil)

		Expect(apiErr).To(BeNil())
		pods, ok := result.(*[]models.PodFull)
		Expect(ok).To(BeTrue())
		Expect(*pods).To(BeEmpty())
	})

	It("lists deleted pods", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsListDeleted).Returns([]podRow{})

		result, apiErr := DbListPods(testCtx(1), 1, true, 0, 0, "", "", nil, ext_api.VerbosityBrief, "", "", nil)

		Expect(apiErr).To(BeNil())
		Expect(result).ToNot(BeNil())
	})

	It("filters by department scopes", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsListDept).Returns([]podRow{})

		scopes := &openapi.PermittedScopes{Tenants: []string{"2"}, Departments: []string{"d1"}}
		result, apiErr := DbListPods(testCtx(1), 1, false, 0, 0, "", "", nil, ext_api.VerbosityBrief, "", "", scopes)

		Expect(apiErr).To(BeNil())
		Expect(result).ToNot(BeNil())
	})

	It("filters by cluster scopes", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsListClusters).Returns([]podRow{})

		scopes := &openapi.PermittedScopes{Tenants: []string{"2"}, Clusters: []string{"c1"}}
		result, apiErr := DbListPods(testCtx(1), 1, false, 0, 0, "", "", nil, ext_api.VerbosityBrief, "", "", scopes)

		Expect(apiErr).To(BeNil())
		Expect(result).ToNot(BeNil())
	})

	It("applies element filter, sorting, paging and search together", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsListPaged).Returns([]podRow{})

		result, apiErr := DbListPods(testCtx(1), 1, false, 10, 20, "desc", "pods.name", podsElementFilter(), ext_api.VerbosityBrief, "true", "abc", nil)

		Expect(apiErr).To(BeNil())
		Expect(result).ToNot(BeNil())
	})

	It("returns an error when the brief query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsListBrief).ReturnsError(errors.New("boom"))

		result, apiErr := DbListPods(testCtx(1), 1, false, 0, 0, "", "", nil, ext_api.VerbosityBrief, "", "", nil)

		Expect(result).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the full query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsListFull).ReturnsError(errors.New("boom"))

		result, apiErr := DbListPods(testCtx(1), 1, false, 0, 0, "", "", nil, ext_api.VerbosityFull, "", "", nil)

		Expect(result).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbGetRevisionPods", func() {
	It("returns the pods of a revision", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsRevisionBasic).Returns([]podRow{})

		pods, apiErr := DbGetRevisionPods(testCtx(1), 1, "rev-1", false, 0, 0, "")

		Expect(apiErr).To(BeNil())
		Expect(*pods).To(BeEmpty())
	})

	It("returns the deleted pods of a revision", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsRevisionDeleted).Returns([]podRow{})

		pods, apiErr := DbGetRevisionPods(testCtx(1), 1, "rev-1", true, 0, 0, "")

		Expect(apiErr).To(BeNil())
		Expect(*pods).To(BeEmpty())
	})

	It("applies paging and search", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsRevisionPaged).Returns([]podRow{})

		pods, apiErr := DbGetRevisionPods(testCtx(1), 1, "rev-1", false, 5, 10, "abc")

		Expect(apiErr).To(BeNil())
		Expect(*pods).To(BeEmpty())
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsRevisionBasic).ReturnsError(errors.New("boom"))

		pods, apiErr := DbGetRevisionPods(testCtx(1), 1, "rev-1", false, 0, 0, "")

		Expect(apiErr).ToNot(BeNil())
		Expect(*pods).To(BeEmpty())
	})
})

var _ = Describe("DbGetPod", func() {
	It("returns the pod with its containers and elements", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsGetPod).Returns([]podRow{{ID: podsTestPodIdStr}})
		m.ExpectQuery(podsGetContainers).Returns([]podRow{})
		m.ExpectQuery(podsGetElements).Returns([]podRow{})

		pod, apiErr := DbGetPod(testCtx(1), 1, uuid.MustParse(podsTestWorkloadIdStr), uuid.MustParse(podsTestPodIdStr))

		Expect(apiErr).To(BeNil())
		Expect(pod.ID).To(Equal(uuid.MustParse(podsTestPodIdStr)))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsGetPod).ReturnsError(errors.New("boom"))

		pod, apiErr := DbGetPod(testCtx(1), 1, uuid.MustParse(podsTestWorkloadIdStr), uuid.MustParse(podsTestPodIdStr))

		Expect(pod).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbDeletePod", func() {
	It("soft deletes the pod", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(podsDelete).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := DbDeletePod(testCtx(1), 1, uuid.MustParse(podsTestPodIdStr))

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(podsDelete).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := DbDeletePod(testCtx(1), 1, uuid.MustParse(podsTestPodIdStr))

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the pod is not found", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(podsDelete).ReturnsError(gorm.ErrRecordNotFound)
		m.ExpectTransaction().Rollback()

		apiErr := DbDeletePod(testCtx(1), 1, uuid.MustParse(podsTestPodIdStr))

		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbUpdatePodsByNodeName", func() {
	It("updates pods with the given topology labels", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(podsUpdateByNode).Returns(2)
		m.ExpectTransaction().Commit()

		labels := []string{"a", "b"}
		affected, apiErr := DbUpdatePodsByNodeName(testCtx(1), 1, uuid.MustParse(podsTestClusterIdStr), "node-1",
			api.UpdateNodePodsRequest{NetworkTopologyLabelValues: &labels})

		Expect(apiErr).To(BeNil())
		Expect(affected).To(Equal(int64(2)))
	})

	It("updates pods when no labels are given", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(podsUpdateByNode).Returns(1)
		m.ExpectTransaction().Commit()

		affected, apiErr := DbUpdatePodsByNodeName(testCtx(1), 1, uuid.MustParse(podsTestClusterIdStr), "node-1",
			api.UpdateNodePodsRequest{})

		Expect(apiErr).To(BeNil())
		Expect(affected).To(Equal(int64(1)))
	})

	It("returns an error when the update fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(podsUpdateByNode).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		affected, apiErr := DbUpdatePodsByNodeName(testCtx(1), 1, uuid.MustParse(podsTestClusterIdStr), "node-1",
			api.UpdateNodePodsRequest{})

		Expect(apiErr).ToNot(BeNil())
		Expect(affected).To(Equal(int64(0)))
	})
})

var _ = Describe("DbUpsertPod", func() {
	It("upserts the pod", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(podsUpsertUpdate).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := DbUpsertPod(testCtx(1), podsTestPod())

		Expect(apiErr).To(BeNil())
	})

	It("returns an error on a postgres foreign key violation", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(podsUpsertUpdate).ReturnsError(&pgconn.PgError{Code: "23503"})
		m.ExpectTransaction().Rollback()

		apiErr := DbUpsertPod(testCtx(1), podsTestPod())

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error on a gorm foreign key violation", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(podsUpsertUpdate).ReturnsError(gorm.ErrForeignKeyViolated)
		m.ExpectTransaction().Rollback()

		apiErr := DbUpsertPod(testCtx(1), podsTestPod())

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the upsert fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(podsUpsertUpdate).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := DbUpsertPod(testCtx(1), podsTestPod())

		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbSyncPods", func() {
	It("syncs all pods and deletes the rest", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(podsSavepoint).Returns(0)
		m.ExpectExec(podsUpsertUpdate).Returns(1)
		m.ExpectExec(podsSavepoint).Returns(0)
		m.ExpectExec(podsUpsertUpdate).Returns(1)
		m.ExpectExec(podsSavepoint).Returns(0)
		m.ExpectExec(podsSyncDeleteExcluding2).Returns(1)
		m.ExpectTransaction().Commit()

		second := podsTestPod()
		second.ID = uuid.MustParse(podsTestPodId2Str)
		synced, failed, apiErr := DbSyncPods(testCtx(1), 1, "", podsTestClusterIdStr,
			[]models.Pod{second, podsTestPod()}, nil, nil, nil, "")

		Expect(apiErr).To(BeNil())
		Expect(synced).To(Equal([]string{podsTestPodIdStr, podsTestPodId2Str}))
		Expect(failed).To(BeEmpty())
	})

	It("reports a pod as failed when its save fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(podsSavepoint).Returns(0)
		m.ExpectExec(podsUpsertUpdate).ReturnsError(errors.New("boom"))
		m.ExpectExec(podsRollbackTo).Returns(0)
		m.ExpectExec(podsSyncDeleteExcluding).Returns(1)
		m.ExpectTransaction().Commit()

		synced, failed, apiErr := DbSyncPods(testCtx(1), 1, "", podsTestClusterIdStr,
			[]models.Pod{podsTestPod()}, nil, nil, nil, "")

		Expect(apiErr).To(BeNil())
		Expect(synced).To(BeEmpty())
		Expect(failed).To(HaveLen(1))
	})

	It("reports a pod as failed on a foreign key violation", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(podsSavepoint).Returns(0)
		m.ExpectExec(podsUpsertUpdate).ReturnsError(&pgconn.PgError{Code: "23503"})
		m.ExpectExec(podsRollbackTo).Returns(0)
		m.ExpectExec(podsSyncDeleteExcluding).Returns(1)
		m.ExpectTransaction().Commit()

		synced, failed, apiErr := DbSyncPods(testCtx(1), 1, "", podsTestClusterIdStr,
			[]models.Pod{podsTestPod()}, nil, nil, nil, "")

		Expect(apiErr).To(BeNil())
		Expect(synced).To(BeEmpty())
		Expect(failed).To(HaveLen(1))
	})

	It("returns an error when no scope is given", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(podsSavepoint).Returns(0)
		m.ExpectTransaction().Rollback()

		synced, failed, apiErr := DbSyncPods(testCtx(1), 1, "", "", nil, nil, nil, nil, "")

		Expect(synced).To(BeNil())
		Expect(failed).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(podsSavepoint).Returns(0)
		m.ExpectExec(podsSyncDelete).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		synced, failed, apiErr := DbSyncPods(testCtx(1), 1, "", podsTestClusterIdStr, nil, nil, nil, nil, "")

		Expect(synced).To(BeNil())
		Expect(failed).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the commit fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(podsSavepoint).Returns(0)
		m.ExpectExec(podsSyncDelete).Returns(1)

		synced, failed, apiErr := DbSyncPods(testCtx(1), 1, "", podsTestClusterIdStr, nil, nil, nil, nil, "")

		Expect(synced).To(BeNil())
		Expect(failed).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the transaction cannot be started", func() {
		setupDaoMock(GinkgoT())

		synced, failed, apiErr := DbSyncPods(testCtx(1), 1, "", podsTestClusterIdStr, nil, nil, nil, nil, "")

		Expect(synced).To(BeNil())
		Expect(failed).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbUpsertContainers", func() {
	It("upserts the containers", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(podsInsertContainer).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := DbUpsertContainers(testCtx(1), []models.Container{podsTestContainer()}, nil)

		Expect(apiErr).To(BeNil())
	})

	It("skips containers of failed pods", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectTransaction().Commit()

		apiErr := DbUpsertContainers(testCtx(1), []models.Container{podsTestContainer()}, []string{podsTestPodIdStr})

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the insert fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(podsInsertContainer).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := DbUpsertContainers(testCtx(1), []models.Container{podsTestContainer()}, nil)

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the commit fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()

		apiErr := DbUpsertContainers(testCtx(1), nil, nil)

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the transaction cannot be started", func() {
		setupDaoMock(GinkgoT())

		apiErr := DbUpsertContainers(testCtx(1), nil, nil)

		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbUpsertPodElements", func() {
	It("does nothing when there are no elements", func() {
		setupDaoMock(GinkgoT())

		apiErr := DbUpsertPodElements(testCtx(1), nil, nil)

		Expect(apiErr).To(BeNil())
	})

	It("upserts the pod elements", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(podsInsertElement).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := DbUpsertPodElements(testCtx(1), []models.PodElement{podsTestElement()}, nil)

		Expect(apiErr).To(BeNil())
	})

	It("skips elements of failed pods", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectTransaction().Commit()

		apiErr := DbUpsertPodElements(testCtx(1), []models.PodElement{podsTestElement()}, []string{podsTestPodIdStr})

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the insert fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(podsInsertElement).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := DbUpsertPodElements(testCtx(1), []models.PodElement{podsTestElement()}, nil)

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the commit fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(podsInsertElement).Returns(1)

		apiErr := DbUpsertPodElements(testCtx(1), []models.PodElement{podsTestElement()}, nil)

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the transaction cannot be started", func() {
		setupDaoMock(GinkgoT())

		apiErr := DbUpsertPodElements(testCtx(1), []models.PodElement{podsTestElement()}, nil)

		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbGetWorkloadTopologyLca", func() {
	It("returns the lowest common ancestor topology", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsWorkloadLca).Returns([]podLcaRow{{LcaTopology: "{a,b}"}})

		lca, apiErr := DbGetWorkloadTopologyLca(testCtx(1), 1, uuid.MustParse(podsTestWorkloadIdStr))

		Expect(apiErr).To(BeNil())
		Expect(lca).To(Equal(pq.StringArray{"a", "b"}))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsWorkloadLca).ReturnsError(errors.New("boom"))

		lca, apiErr := DbGetWorkloadTopologyLca(testCtx(1), 1, uuid.MustParse(podsTestWorkloadIdStr))

		Expect(lca).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbGetElementTopologyLca", func() {
	It("returns the lowest common ancestor topology", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsElementLca).Returns([]podLcaRow{{LcaTopology: "{a}"}})

		lca, apiErr := DbGetElementTopologyLca(testCtx(1), 1, uuid.MustParse(podsTestWorkloadIdStr), "el-1")

		Expect(apiErr).To(BeNil())
		Expect(lca).To(Equal(pq.StringArray{"a"}))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsElementLca).ReturnsError(errors.New("boom"))

		lca, apiErr := DbGetElementTopologyLca(testCtx(1), 1, uuid.MustParse(podsTestWorkloadIdStr), "el-1")

		Expect(lca).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbGetPodClusterId", func() {
	It("returns the cluster id of the pod", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsGetPodClusterId).Returns([]podClusterIdRow{{ClusterID: podsTestClusterIdStr}})

		clusterId, apiErr := DbGetPodClusterId(testCtx(1), 1, podsTestWorkloadIdStr, podsTestPodIdStr)

		Expect(apiErr).To(BeNil())
		Expect(clusterId).To(Equal(uuid.MustParse(podsTestClusterIdStr)))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(podsGetPodClusterId).ReturnsError(errors.New("boom"))

		clusterId, apiErr := DbGetPodClusterId(testCtx(1), 1, podsTestWorkloadIdStr, podsTestPodIdStr)

		Expect(apiErr).ToNot(BeNil())
		Expect(clusterId).To(Equal(uuid.UUID{}))
	})
})

var _ = Describe("hasElementIdFilter", func() {
	It("detects an element id filter and rejects other filters", func() {
		Expect(hasElementIdFilter(podsElementFilter())).To(BeTrue())
		Expect(hasElementIdFilter(podsNameFilter())).To(BeFalse())
		Expect(hasElementIdFilter(nil)).To(BeFalse())
	})
})
