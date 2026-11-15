/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"errors"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	dbtesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/testing"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

const (
	elementsDeleteScoped     = "DELETE FROM workload_elements WHERE tenant_id = ? AND workload_id = ? AND cluster_id = ? AND id NOT IN (?)"
	elementsDeleteNoCluster  = "DELETE FROM workload_elements WHERE tenant_id = ? AND workload_id = ? AND id NOT IN (?)"
	elementsDeleteNoElements = "DELETE FROM workload_elements WHERE tenant_id = ? AND workload_id = ? AND cluster_id = ?"
	elementsInsert           = "INSERT INTO workload_elements (id,workload_id,tenant_id,cluster_id,name,level,parent_id,instance_key,replica_key,k8s_ref_uid,k8s_ref_name,gvk_group,gvk_version,gvk_kind,created_at,topology) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT (workload_id,id) DO UPDATE SET topology=excluded.topology WHERE workload_elements.topology IS DISTINCT FROM excluded.topology"
	elementsSelect           = "SELECT * FROM workload_elements WHERE tenant_id = ? AND workload_id = ? ORDER BY level, name"
	elementsSelectPaged      = "SELECT * FROM workload_elements WHERE tenant_id = ? AND workload_id = ? ORDER BY level, name LIMIT ? OFFSET ?"
	elementsCount            = "SELECT count(*) FROM workload_elements WHERE tenant_id = ? AND workload_id = ?"
	elementsSelectOne        = "SELECT * FROM workload_elements WHERE tenant_id = ? AND workload_id = ? AND id = ? ORDER BY workload_elements.id LIMIT ?"
	elementsPodsSelect       = "SELECT pods.id,pods.created_at,pods.updated_at,pods.completed_at,pods.deleted_at,pods.workload_id,pods.revision_id,pods.name,pods.priority_class_name,pods.tenant_id,pods.cluster_id,pods.project_id,pods.node_name,pods.current_nodepool,pods.requested_nodepools,pods.requested_gpu_limit,pods.requested_gpu_request,pods.requested_gpu_memory_limit,pods.requested_gpu_memory_request,pods.requested_cpu_limit,pods.requested_cpu_request,pods.requested_cpu_memory_limit,pods.requested_cpu_memory_request,pods.requested_mig_device,pods.allocated_gpu,pods.allocated_gpu_memory,pods.allocated_cpu,pods.allocated_cpu_memory,pods.allocated_mig_device,pods.phase,pods.phase_updated_at,pods.tolerations,pods.ip,pods.task,pods.node_affinity,pods.network_topology_label_values,pods.resource_claims FROM pods INNER JOIN pod_elements ON pod_elements.pod_id = pods.id WHERE pods.tenant_id = ? AND pods.workload_id = ? AND pod_elements.element_id = ? AND pods.deleted_at IS NULL"
)

var (
	elementsWorkloadID = uuid.MustParse("44444444-4444-4444-4444-444444444444")
	elementsClusterID  = uuid.MustParse("55555555-5555-5555-5555-555555555555")
	elementsPodID      = uuid.MustParse("66666666-6666-6666-6666-666666666666")
)

type elementRow struct {
	ID         string
	WorkloadID uuid.UUID `gorm:"column:workload_id"`
	Name       string
	Level      int
}

type elementPodRow struct {
	ID   uuid.UUID
	Name string
}

func elementsFixture() []models.WorkloadElement {
	return []models.WorkloadElement{{
		ID:         "elem-1",
		WorkloadID: elementsWorkloadID,
		TenantID:   1,
		ClusterID:  elementsClusterID,
		Name:       "worker",
		Level:      0,
	}}
}

func elementsInsertArgs() []any {
	return []any{
		"elem-1", elementsWorkloadID, 1, elementsClusterID, "worker", 0,
		nil, nil, nil, nil, nil, nil, nil, nil, dbtesting.AnyTime(), nil,
	}
}

var _ = Describe("DbUpsertWorkloadsElements", func() {
	It("deletes stale elements and inserts the given ones", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec("SAVEPOINT sp").Returns(0)
		m.ExpectExec(elementsDeleteScoped, 1, elementsWorkloadID, elementsClusterID.String(), "elem-1").Returns(1)
		m.ExpectExec(elementsInsert, elementsInsertArgs()...).Returns(1)
		m.ExpectExec("SAVEPOINT sp").Returns(0)
		m.ExpectTransaction().Commit()

		DbUpsertWorkloadsElements(testCtx(1), 1, elementsClusterID.String(),
			map[uuid.UUID][]models.WorkloadElement{elementsWorkloadID: elementsFixture()}, nil)
	})

	It("skips workloads that are listed as failed", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec("SAVEPOINT sp").Returns(0)
		m.ExpectTransaction().Commit()

		DbUpsertWorkloadsElements(testCtx(1), 1, elementsClusterID.String(),
			map[uuid.UUID][]models.WorkloadElement{elementsWorkloadID: elementsFixture()},
			[]string{elementsWorkloadID.String()})
	})

	It("rolls back to the savepoint when the upsert fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec("SAVEPOINT sp").Returns(0)
		m.ExpectExec(elementsDeleteScoped, 1, elementsWorkloadID, elementsClusterID.String(), "elem-1").ReturnsError(errors.New("boom"))
		m.ExpectExec("ROLLBACK TO SAVEPOINT sp").Returns(0)
		m.ExpectTransaction().Commit()

		DbUpsertWorkloadsElements(testCtx(1), 1, elementsClusterID.String(),
			map[uuid.UUID][]models.WorkloadElement{elementsWorkloadID: elementsFixture()}, nil)
	})

	It("tolerates a failing commit", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec("SAVEPOINT sp").Returns(0)

		DbUpsertWorkloadsElements(testCtx(1), 1, elementsClusterID.String(),
			map[uuid.UUID][]models.WorkloadElement{}, nil)
	})

	It("tolerates a failing transaction begin", func() {
		setupDaoMock(GinkgoT())

		DbUpsertWorkloadsElements(testCtx(1), 1, elementsClusterID.String(),
			map[uuid.UUID][]models.WorkloadElement{elementsWorkloadID: elementsFixture()}, nil)
	})
})

var _ = Describe("DbUpsertWorkloadElements", func() {
	It("deletes stale elements and inserts the given ones", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(elementsDeleteScoped, 1, elementsWorkloadID, elementsClusterID.String(), "elem-1").Returns(1)
		m.ExpectExec(elementsInsert, elementsInsertArgs()...).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := DbUpsertWorkloadElements(testCtx(1), 1, elementsClusterID.String(), elementsWorkloadID, elementsFixture())

		Expect(apiErr).To(BeNil())
	})

	It("returns an internal server error when the delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(elementsDeleteScoped, 1, elementsWorkloadID, elementsClusterID.String(), "elem-1").ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := DbUpsertWorkloadElements(testCtx(1), 1, elementsClusterID.String(), elementsWorkloadID, elementsFixture())

		Expect(apiErr).ToNot(BeNil())
		Expect(apiErr.Code).To(BeEquivalentTo(http.StatusInternalServerError))
	})

	It("returns an internal server error when the commit fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(elementsDeleteNoElements, 1, elementsWorkloadID, elementsClusterID.String()).Returns(0)

		apiErr := DbUpsertWorkloadElements(testCtx(1), 1, elementsClusterID.String(), elementsWorkloadID, nil)

		Expect(apiErr).ToNot(BeNil())
		Expect(apiErr.Code).To(BeEquivalentTo(http.StatusInternalServerError))
	})

	It("omits the cluster filter when no cluster id is given", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(elementsDeleteNoCluster, 1, elementsWorkloadID, "elem-1").Returns(1)
		m.ExpectExec(elementsInsert, elementsInsertArgs()...).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := DbUpsertWorkloadElements(testCtx(1), 1, "", elementsWorkloadID, elementsFixture())

		Expect(apiErr).To(BeNil())
	})

	It("returns an internal server error when the insert fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(elementsDeleteScoped, 1, elementsWorkloadID, elementsClusterID.String(), "elem-1").Returns(1)
		m.ExpectExec(elementsInsert, elementsInsertArgs()...).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := DbUpsertWorkloadElements(testCtx(1), 1, elementsClusterID.String(), elementsWorkloadID, elementsFixture())

		Expect(apiErr).ToNot(BeNil())
		Expect(apiErr.Code).To(BeEquivalentTo(http.StatusInternalServerError))
	})

	It("returns an internal server error when the transaction cannot begin", func() {
		setupDaoMock(GinkgoT())

		apiErr := DbUpsertWorkloadElements(testCtx(1), 1, elementsClusterID.String(), elementsWorkloadID, elementsFixture())

		Expect(apiErr).ToNot(BeNil())
		Expect(apiErr.Code).To(BeEquivalentTo(http.StatusInternalServerError))
	})
})

var _ = Describe("DbGetWorkloadElements", func() {
	It("returns the workload elements", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(elementsSelect, 1, elementsWorkloadID).
			Returns([]elementRow{{ID: "elem-1", WorkloadID: elementsWorkloadID, Name: "worker", Level: 0}})

		elements, apiErr := DbGetWorkloadElements(testCtx(1), 1, elementsWorkloadID, 0, 0)

		Expect(apiErr).To(BeNil())
		Expect(elements).To(HaveLen(1))
		Expect(elements[0].ID).To(Equal("elem-1"))
	})

	It("applies the offset and limit when given", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(elementsSelectPaged, 1, elementsWorkloadID, 5, 10).Returns([]elementRow{})

		elements, apiErr := DbGetWorkloadElements(testCtx(1), 1, elementsWorkloadID, 10, 5)

		Expect(apiErr).To(BeNil())
		Expect(elements).To(BeEmpty())
	})

	It("returns an internal server error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(elementsSelect, 1, elementsWorkloadID).ReturnsError(errors.New("boom"))

		elements, apiErr := DbGetWorkloadElements(testCtx(1), 1, elementsWorkloadID, 0, 0)

		Expect(elements).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
		Expect(apiErr.Code).To(BeEquivalentTo(http.StatusInternalServerError))
	})
})

var _ = Describe("DbCountWorkloadElements", func() {
	It("returns the element count", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(elementsCount, 1, elementsWorkloadID).Returns([]countRow{{Count: 3}})

		count, apiErr := DbCountWorkloadElements(testCtx(1), 1, elementsWorkloadID)

		Expect(apiErr).To(BeNil())
		Expect(count).To(BeEquivalentTo(3))
	})

	It("returns an internal server error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(elementsCount, 1, elementsWorkloadID).ReturnsError(errors.New("boom"))

		count, apiErr := DbCountWorkloadElements(testCtx(1), 1, elementsWorkloadID)

		Expect(count).To(BeEquivalentTo(0))
		Expect(apiErr).ToNot(BeNil())
		Expect(apiErr.Code).To(BeEquivalentTo(http.StatusInternalServerError))
	})
})

var _ = Describe("DbGetWorkloadElement", func() {
	It("returns the single workload element", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(elementsSelectOne, 1, elementsWorkloadID, "elem-1", 1).
			Returns([]elementRow{{ID: "elem-1", WorkloadID: elementsWorkloadID, Name: "worker"}})

		element, apiErr := DbGetWorkloadElement(testCtx(1), 1, elementsWorkloadID, "elem-1")

		Expect(apiErr).To(BeNil())
		Expect(element).ToNot(BeNil())
		Expect(element.ID).To(Equal("elem-1"))
	})

	It("returns not found when the record does not exist", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(elementsSelectOne, 1, elementsWorkloadID, "elem-1", 1).ReturnsError(gorm.ErrRecordNotFound)

		element, apiErr := DbGetWorkloadElement(testCtx(1), 1, elementsWorkloadID, "elem-1")

		Expect(element).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
		Expect(apiErr.Code).To(BeEquivalentTo(http.StatusNotFound))
	})

	It("returns an internal server error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(elementsSelectOne, 1, elementsWorkloadID, "elem-1", 1).ReturnsError(errors.New("boom"))

		element, apiErr := DbGetWorkloadElement(testCtx(1), 1, elementsWorkloadID, "elem-1")

		Expect(element).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
		Expect(apiErr.Code).To(BeEquivalentTo(http.StatusInternalServerError))
	})
})

var _ = Describe("DbGetElementPods", func() {
	It("returns the pods of the element", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(elementsPodsSelect, 1, elementsWorkloadID, "elem-1").
			Returns([]elementPodRow{{ID: elementsPodID, Name: "pod-name"}})

		pods, apiErr := DbGetElementPods(testCtx(1), 1, elementsWorkloadID, "elem-1")

		Expect(apiErr).To(BeNil())
		Expect(pods).To(HaveLen(1))
	})

	It("returns an internal server error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(elementsPodsSelect, 1, elementsWorkloadID, "elem-1").ReturnsError(errors.New("boom"))

		pods, apiErr := DbGetElementPods(testCtx(1), 1, elementsWorkloadID, "elem-1")

		Expect(pods).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
		Expect(apiErr.Code).To(BeEquivalentTo(http.StatusInternalServerError))
	})
})
