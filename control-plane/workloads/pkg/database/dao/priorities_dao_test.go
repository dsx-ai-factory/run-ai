/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"errors"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	prioritiesSelect            = "SELECT * FROM priorities WHERE (tenant_id = ? or scope_type='system') AND priorities.deleted_at IS NULL ORDER BY id DESC"
	prioritiesStatusPreload     = "SELECT * FROM priority_statuses WHERE priority_statuses.priority_id = ?"
	prioritiesClusterSelect     = "SELECT * FROM priorities_cluster_view WHERE cluster_id = ? AND (tenant_id = ? or scope_type='system') AND priorities_cluster_view.deleted_at IS NULL ORDER BY id DESC"
	prioritiesSyncSelect        = "SELECT * FROM priorities WHERE (tenant_id = ? or scope_type='system') AND source = ? ORDER BY id DESC"
	prioritiesByNameSelect      = "SELECT * FROM priorities WHERE deleted_at IS NULL AND name = ? AND ((scope_type = 'system' OR tenant_id = ?)) AND priorities.deleted_at IS NULL ORDER BY priorities.id LIMIT ?"
	prioritiesByIdSelect        = "SELECT * FROM priorities WHERE deleted_at IS NULL AND id = ? AND ((scope_type = 'system' OR tenant_id = ?)) AND priorities.deleted_at IS NULL ORDER BY priorities.id LIMIT ?"
	prioritiesFilterSortSelect  = "SELECT * FROM priorities WHERE (tenant_id = ? or scope_type='system') AND name = ? AND priorities.deleted_at IS NULL ORDER BY value DESC NULLS LAST,id DESC"
	prioritiesSaveUpdate        = "UPDATE priorities SET created_at=?,updated_at=?,deleted_at=?,created_by=?,updated_by=?,deleted_by=?,name=?,value=?,source=?,scope_type=?,scope_id=?,tenant_id=? WHERE priorities.deleted_at IS NULL AND id = ?"
	prioritiesStatusUpsert      = "INSERT INTO priority_statuses (id,priority_id,cluster_id,phase) VALUES (?,?,?,?) ON CONFLICT (priority_id,cluster_id) DO UPDATE SET phase=excluded.phase"
	prioritiesStatusAssocUpsert = "INSERT INTO priority_statuses (id,priority_id,cluster_id,phase) VALUES (?,?,?,?) ON CONFLICT (id) DO UPDATE SET priority_id=excluded.priority_id"
	prioritiesRawUpdate         = "UPDATE priorities SET value = ?, source = ? WHERE name = ? AND scope_type = ? AND scope_id = ? AND tenant_id = ? RETURNING id"

	prioritiesClusterStatusDelete         = "DELETE FROM priority_statuses ps USING priorities p WHERE ps.priority_id = p.id AND p.name IN (?) AND p.scope_id = ? AND p.scope_type = ? AND p.tenant_id = ?"
	prioritiesClusterPriorityDelete       = "DELETE FROM priorities p WHERE p.name IN (?) AND p.scope_id = ? AND p.scope_type = ? AND p.tenant_id = ?"
	prioritiesExcludedStatusDelete        = "DELETE FROM priority_statuses ps USING priorities p WHERE ps.priority_id = p.id AND ps.cluster_id = ? AND p.scope_id = ? AND p.scope_type = ? AND p.tenant_id = ?"
	prioritiesExcludedPriorityDelete      = "DELETE FROM priorities p WHERE p.scope_id = ? AND p.scope_type = ? AND p.tenant_id = ?"
	prioritiesExcludedStatusDeleteNotIn   = "DELETE FROM priority_statuses ps USING priorities p WHERE ps.priority_id = p.id AND ps.cluster_id = ? AND p.scope_id = ? AND p.scope_type = ? AND p.tenant_id = ? AND p.name NOT IN (?)"
	prioritiesExcludedPriorityDeleteNotIn = "DELETE FROM priorities p WHERE p.scope_id = ? AND p.scope_type = ? AND p.tenant_id = ? AND p.name NOT IN (?)"
	prioritiesRunaiStatusDelete           = "DELETE FROM priority_statuses ps USING priorities p WHERE ps.priority_id in (?) AND ps.cluster_id = ? AND p.scope_type IN (?,?)"
	prioritiesRunaiExcludedDelete         = "DELETE FROM priority_statuses ps USING priorities p WHERE ps.priority_id = p.id AND ps.cluster_id = ? AND p.scope_type IN (?,?)"
	prioritiesRunaiExcludedDeleteNotIn    = "DELETE FROM priority_statuses ps USING priorities p WHERE ps.priority_id = p.id AND ps.cluster_id = ? AND p.scope_type IN (?,?) AND p.id NOT IN (?)"

	prioritiesSavepoint     = "SAVEPOINT sp"
	prioritiesRollbackTo    = "ROLLBACK TO SAVEPOINT sp"
	prioritiesTenantDelete  = "DELETE FROM priorities WHERE tenant_id = ?"
	prioritiesClusterDelete = "UPDATE priorities SET deleted_at=? WHERE tenant_id = ? AND scope_type = ? AND scope_id = ? AND priorities.deleted_at IS NULL"
)

type priorityRow struct {
	ID   uuid.UUID
	Name string
}

type priorityIdRow struct {
	ID uuid.UUID
}

type priorityStatusRow struct {
	ID         uuid.UUID
	PriorityID uuid.UUID
	ClusterId  uuid.UUID
	Phase      string
}

var (
	prioritiesTestId      = uuid.MustParse("11111111-1111-1111-1111-111111111111")
	prioritiesTestCluster = uuid.MustParse("22222222-2222-2222-2222-222222222222")
)

func newPgUniqueError(constraint string) error {
	return &pgconn.PgError{Code: UniqueViolationErr, ConstraintName: constraint}
}

func newTestPriority(withStatus bool) models.Priority {
	p := models.Priority{
		Name:      "high",
		Value:     100,
		Source:    "cluster",
		ScopeType: ClusterScopeType,
		ScopeId:   prioritiesTestCluster.String(),
		TenantId:  "tenant-1",
	}
	p.ID = prioritiesTestId
	if withStatus {
		p.Status = []models.PriorityStatus{{PriorityID: prioritiesTestId, ClusterId: prioritiesTestCluster, Phase: "ready"}}
	}
	return p
}

var _ = Describe("GetPriorities", func() {
	It("returns the priorities with their statuses", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(prioritiesSelect).Returns([]priorityRow{{ID: prioritiesTestId, Name: "high"}})
		m.ExpectQuery(prioritiesStatusPreload).Returns([]priorityStatusRow{{ID: prioritiesTestId, PriorityID: prioritiesTestId, Phase: "ready"}})

		priorities, err := GetPriorities(testCtx(1), "tenant-1", "", "", nil)

		Expect(err).ToNot(HaveOccurred())
		Expect(*priorities).To(HaveLen(1))
	})

	It("returns an empty list when there are no priorities", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(prioritiesSelect).Returns([]priorityRow{})

		priorities, err := GetPriorities(testCtx(1), "tenant-1", "", "", nil)

		Expect(err).ToNot(HaveOccurred())
		Expect(*priorities).To(BeEmpty())
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(prioritiesSelect).ReturnsError(errors.New("boom"))

		_, err := GetPriorities(testCtx(1), "tenant-1", "", "", nil)

		Expect(err).To(HaveOccurred())
	})

	It("applies filters and sorting", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(prioritiesFilterSortSelect).Returns([]priorityRow{})

		filters := []queryapi.Filter{{Field: "name", Operator: "=", Values: []string{"high"}}}
		_, err := GetPriorities(testCtx(1), "tenant-1", "DESC", "value", filters)

		Expect(err).ToNot(HaveOccurred())
	})
})

var _ = Describe("GetClusterPriorities", func() {
	It("returns the cluster priorities", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(prioritiesClusterSelect).Returns([]priorityRow{{ID: prioritiesTestId, Name: "high"}})

		priorities, err := GetClusterPriorities(testCtx(1), "tenant-1", "cluster-1", "", "", nil)

		Expect(err).ToNot(HaveOccurred())
		Expect(*priorities).To(HaveLen(1))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(prioritiesClusterSelect).ReturnsError(errors.New("boom"))

		_, err := GetClusterPriorities(testCtx(1), "tenant-1", "cluster-1", "", "", nil)

		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("GetPrioritiesSync", func() {
	It("returns the priorities for the given source", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(prioritiesSyncSelect).Returns([]priorityRow{{ID: prioritiesTestId, Name: "high"}})

		priorities, err := GetPrioritiesSync(testCtx(1), "tenant-1", "runai")

		Expect(err).ToNot(HaveOccurred())
		Expect(*priorities).To(HaveLen(1))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(prioritiesSyncSelect).ReturnsError(errors.New("boom"))

		_, err := GetPrioritiesSync(testCtx(1), "tenant-1", "runai")

		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("GetPriorityByName", func() {
	It("returns the priority", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(prioritiesByNameSelect).Returns([]priorityRow{{ID: prioritiesTestId, Name: "high"}})

		priority, apiErr := GetPriorityByName(testCtx(1), "tenant-1", "high")

		Expect(apiErr).To(BeNil())
		Expect(priority.Name).To(Equal("high"))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(prioritiesByNameSelect).ReturnsError(errors.New("boom"))

		priority, apiErr := GetPriorityByName(testCtx(1), "tenant-1", "high")

		Expect(priority).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("GetPriorityById", func() {
	It("returns the priority", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(prioritiesByIdSelect).Returns([]priorityRow{{ID: prioritiesTestId, Name: "high"}})

		priority, apiErr := GetPriorityById(testCtx(1), "tenant-1", prioritiesTestId.String())

		Expect(apiErr).To(BeNil())
		Expect(priority.ID).To(Equal(prioritiesTestId))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(prioritiesByIdSelect).ReturnsError(errors.New("boom"))

		priority, apiErr := GetPriorityById(testCtx(1), "tenant-1", prioritiesTestId.String())

		Expect(priority).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DeletePrioritiesOfTenantId", func() {
	It("deletes the priorities of the tenant", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(prioritiesTenantDelete).Returns(1)
		m.ExpectTransaction().Commit()

		Expect(DeletePrioritiesOfTenantId(testCtx(1), "tenant-1")).To(Succeed())
	})

	It("returns an error when the delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(prioritiesTenantDelete).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		Expect(DeletePrioritiesOfTenantId(testCtx(1), "tenant-1")).To(HaveOccurred())
	})
})

var _ = Describe("DeletePrioritiesOfClusterId", func() {
	It("deletes the priorities of the cluster", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(prioritiesClusterDelete).Returns(1)
		m.ExpectTransaction().Commit()

		Expect(DeletePrioritiesOfClusterId(testCtx(1), "tenant-1", "cluster-1")).To(Succeed())
	})

	It("returns an error when the delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(prioritiesClusterDelete).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		Expect(DeletePrioritiesOfClusterId(testCtx(1), "tenant-1", "cluster-1")).To(HaveOccurred())
	})
})

var _ = Describe("getPgError", func() {
	It("extracts the postgres error details", func() {
		pgErr := getPgError(newPgUniqueError(UniqueClusterPriorityConstraint))

		Expect(pgErr).ToNot(BeNil())
		Expect(pgErr.Code).To(Equal(UniqueViolationErr))
		Expect(pgErr.ConstraintName).To(Equal(UniqueClusterPriorityConstraint))
	})

	It("returns a nil result for a non postgres error", func() {
		Expect(getPgError(errors.New("boom"))).To(BeNil())
	})
})

var _ = Describe("saveStatus", func() {
	It("returns an error for an invalid status", func() {
		setupDaoMock(GinkgoT())
		_, err := saveStatus(nil, prioritiesTestCluster, models.Priority{})
		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("SyncClusterPrioritiesStatus", func() {
	It("returns an error when the transaction cannot begin", func() {
		setupDaoMock(GinkgoT())

		synced, failed, err := SyncClusterPrioritiesStatus(testCtx(1), "tenant-1", prioritiesTestCluster, nil, nil, false)

		Expect(err).To(HaveOccurred())
		Expect(synced).To(BeNil())
		Expect(failed).To(BeNil())
	})

	It("upserts the priority and its status", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectExec(prioritiesSaveUpdate).Returns(1)
		m.ExpectExec(prioritiesStatusAssocUpsert).Returns(1)
		m.ExpectExec(prioritiesStatusUpsert).Returns(1)
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectTransaction().Commit()

		synced, failed, err := SyncClusterPrioritiesStatus(testCtx(1), "tenant-1", prioritiesTestCluster, []models.Priority{newTestPriority(true)}, nil, false)

		Expect(err).ToNot(HaveOccurred())
		Expect(synced).To(HaveLen(1))
		Expect(failed).To(BeEmpty())
	})

	It("falls back to a raw update on a unique violation and succeeds", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectExec(prioritiesSaveUpdate).ReturnsError(newPgUniqueError(UniqueClusterPriorityConstraint))
		m.ExpectExec(prioritiesRollbackTo).Returns(0)
		m.ExpectQuery(prioritiesRawUpdate).Returns([]priorityIdRow{{ID: prioritiesTestId}})
		m.ExpectExec(prioritiesStatusUpsert).Returns(1)
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectTransaction().Commit()

		synced, failed, err := SyncClusterPrioritiesStatus(testCtx(1), "tenant-1", prioritiesTestCluster, []models.Priority{newTestPriority(true)}, nil, false)

		Expect(err).ToNot(HaveOccurred())
		Expect(synced).To(HaveLen(1))
		Expect(failed).To(BeEmpty())
	})

	It("reports the priority as failed when the unique violation fallback fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectExec(prioritiesSaveUpdate).ReturnsError(newPgUniqueError(UniqueClusterPriorityConstraint))
		m.ExpectExec(prioritiesRollbackTo).Returns(0)
		m.ExpectQuery(prioritiesRawUpdate).ReturnsError(errors.New("boom"))
		m.ExpectExec(prioritiesRollbackTo).Returns(0)
		m.ExpectTransaction().Commit()

		synced, failed, err := SyncClusterPrioritiesStatus(testCtx(1), "tenant-1", prioritiesTestCluster, []models.Priority{newTestPriority(true)}, nil, false)

		Expect(err).ToNot(HaveOccurred())
		Expect(synced).To(BeEmpty())
		Expect(failed).To(Equal([]string{"high"}))
	})

	It("reports the priority as failed when the upsert fails with another constraint", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectExec(prioritiesSaveUpdate).ReturnsError(newPgUniqueError("other_constraint"))
		m.ExpectExec(prioritiesRollbackTo).Returns(0)
		m.ExpectTransaction().Commit()

		synced, failed, err := SyncClusterPrioritiesStatus(testCtx(1), "tenant-1", prioritiesTestCluster, []models.Priority{newTestPriority(true)}, nil, false)

		Expect(err).ToNot(HaveOccurred())
		Expect(synced).To(BeEmpty())
		Expect(failed).To(Equal([]string{"high"}))
	})

	It("reports the priority as failed when the status is invalid", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectExec(prioritiesSaveUpdate).Returns(1)
		m.ExpectExec(prioritiesRollbackTo).Returns(0)
		m.ExpectTransaction().Commit()

		synced, failed, err := SyncClusterPrioritiesStatus(testCtx(1), "tenant-1", prioritiesTestCluster, []models.Priority{newTestPriority(false)}, nil, false)

		Expect(err).ToNot(HaveOccurred())
		Expect(synced).To(BeEmpty())
		Expect(failed).To(HaveLen(1))
	})

	It("deletes the requested priorities", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectExec(prioritiesClusterStatusDelete).Returns(1)
		m.ExpectExec(prioritiesClusterPriorityDelete).Returns(1)
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectTransaction().Commit()

		synced, failed, err := SyncClusterPrioritiesStatus(testCtx(1), "tenant-1", prioritiesTestCluster, nil, []string{"low"}, false)

		Expect(err).ToNot(HaveOccurred())
		Expect(synced).To(Equal([]string{"low"}))
		Expect(failed).To(BeEmpty())
	})

	It("reports the priority as failed when deleting its statuses fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectExec(prioritiesClusterStatusDelete).ReturnsError(errors.New("boom"))
		m.ExpectExec(prioritiesRollbackTo).Returns(0)
		m.ExpectTransaction().Commit()

		synced, failed, err := SyncClusterPrioritiesStatus(testCtx(1), "tenant-1", prioritiesTestCluster, nil, []string{"low"}, false)

		Expect(err).ToNot(HaveOccurred())
		Expect(synced).To(BeEmpty())
		Expect(failed).To(Equal([]string{"low"}))
	})

	It("reports the priority as failed when deleting the priority fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectExec(prioritiesClusterStatusDelete).Returns(1)
		m.ExpectExec(prioritiesClusterPriorityDelete).ReturnsError(errors.New("boom"))
		m.ExpectExec(prioritiesRollbackTo).Returns(0)
		m.ExpectTransaction().Commit()

		synced, failed, err := SyncClusterPrioritiesStatus(testCtx(1), "tenant-1", prioritiesTestCluster, nil, []string{"low"}, false)

		Expect(err).ToNot(HaveOccurred())
		Expect(synced).To(BeEmpty())
		Expect(failed).To(Equal([]string{"low"}))
	})

	It("deletes all cluster priorities when syncing deletions with nothing to exclude", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectExec(prioritiesExcludedStatusDelete).Returns(1)
		m.ExpectExec(prioritiesExcludedPriorityDelete).Returns(1)
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectTransaction().Commit()

		_, _, err := SyncClusterPrioritiesStatus(testCtx(1), "tenant-1", prioritiesTestCluster, nil, nil, true)

		Expect(err).ToNot(HaveOccurred())
	})

	It("does not fail when the status delete errors while syncing deletions with nothing to exclude", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectExec(prioritiesExcludedStatusDelete).ReturnsError(errors.New("boom"))
		m.ExpectExec(prioritiesRollbackTo).Returns(0)
		m.ExpectTransaction().Commit()

		_, _, err := SyncClusterPrioritiesStatus(testCtx(1), "tenant-1", prioritiesTestCluster, nil, nil, true)

		Expect(err).ToNot(HaveOccurred())
	})

	It("deletes everything except the synced priorities when syncing deletions", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectExec(prioritiesSaveUpdate).Returns(1)
		m.ExpectExec(prioritiesStatusAssocUpsert).Returns(1)
		m.ExpectExec(prioritiesStatusUpsert).Returns(1)
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectExec(prioritiesExcludedStatusDeleteNotIn).Returns(1)
		m.ExpectExec(prioritiesExcludedPriorityDeleteNotIn).Returns(1)
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectTransaction().Commit()

		synced, _, err := SyncClusterPrioritiesStatus(testCtx(1), "tenant-1", prioritiesTestCluster, []models.Priority{newTestPriority(true)}, nil, true)

		Expect(err).ToNot(HaveOccurred())
		Expect(synced).To(HaveLen(1))
	})

	It("does not fail when the excluded status delete errors while syncing deletions", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectExec(prioritiesSaveUpdate).Returns(1)
		m.ExpectExec(prioritiesStatusAssocUpsert).Returns(1)
		m.ExpectExec(prioritiesStatusUpsert).Returns(1)
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectExec(prioritiesExcludedStatusDeleteNotIn).ReturnsError(errors.New("boom"))
		m.ExpectExec(prioritiesRollbackTo).Returns(0)
		m.ExpectTransaction().Commit()

		_, _, err := SyncClusterPrioritiesStatus(testCtx(1), "tenant-1", prioritiesTestCluster, []models.Priority{newTestPriority(true)}, nil, true)

		Expect(err).ToNot(HaveOccurred())
	})

	It("returns an error when the commit fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(prioritiesSavepoint).Returns(0)

		synced, failed, err := SyncClusterPrioritiesStatus(testCtx(1), "tenant-1", prioritiesTestCluster, nil, nil, false)

		Expect(err).To(HaveOccurred())
		Expect(synced).To(BeNil())
		Expect(failed).To(BeNil())
	})
})

var _ = Describe("SyncRunaiPrioritiesStatus", func() {
	It("returns an error when the transaction cannot begin", func() {
		setupDaoMock(GinkgoT())

		synced, failed, err := SyncRunaiPrioritiesStatus(testCtx(1), prioritiesTestCluster, nil, nil, false)

		Expect(err).To(HaveOccurred())
		Expect(synced).To(BeNil())
		Expect(failed).To(BeNil())
	})

	It("upserts the priority statuses", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectExec(prioritiesStatusUpsert).Returns(1)
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectTransaction().Commit()

		synced, failed, err := SyncRunaiPrioritiesStatus(testCtx(1), prioritiesTestCluster, []models.Priority{newTestPriority(true)}, nil, false)

		Expect(err).ToNot(HaveOccurred())
		Expect(synced).To(HaveLen(1))
		Expect(failed).To(BeEmpty())
	})

	It("reports the priority as failed when the status is invalid", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectExec(prioritiesRollbackTo).Returns(0)
		m.ExpectTransaction().Commit()

		synced, failed, err := SyncRunaiPrioritiesStatus(testCtx(1), prioritiesTestCluster, []models.Priority{newTestPriority(false)}, nil, false)

		Expect(err).ToNot(HaveOccurred())
		Expect(synced).To(BeEmpty())
		Expect(failed).To(HaveLen(1))
	})

	It("deletes the requested priority statuses", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectExec(prioritiesRunaiStatusDelete).Returns(1)
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectTransaction().Commit()

		synced, failed, err := SyncRunaiPrioritiesStatus(testCtx(1), prioritiesTestCluster, nil, []string{prioritiesTestId.String()}, false)

		Expect(err).ToNot(HaveOccurred())
		Expect(synced).To(HaveLen(1))
		Expect(failed).To(BeEmpty())
	})

	It("reports the priority as failed when the delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectExec(prioritiesRunaiStatusDelete).ReturnsError(errors.New("boom"))
		m.ExpectExec(prioritiesRollbackTo).Returns(0)
		m.ExpectTransaction().Commit()

		synced, failed, err := SyncRunaiPrioritiesStatus(testCtx(1), prioritiesTestCluster, nil, []string{prioritiesTestId.String()}, false)

		Expect(err).ToNot(HaveOccurred())
		Expect(synced).To(BeEmpty())
		Expect(failed).To(HaveLen(1))
	})

	It("deletes all statuses when syncing deletions with nothing to exclude", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectExec(prioritiesRunaiExcludedDelete).Returns(1)
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectTransaction().Commit()

		_, _, err := SyncRunaiPrioritiesStatus(testCtx(1), prioritiesTestCluster, nil, nil, true)

		Expect(err).ToNot(HaveOccurred())
	})

	It("deletes everything except the synced statuses when syncing deletions", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectExec(prioritiesStatusUpsert).Returns(1)
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectExec(prioritiesRunaiExcludedDeleteNotIn).Returns(1)
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectTransaction().Commit()

		_, _, err := SyncRunaiPrioritiesStatus(testCtx(1), prioritiesTestCluster, []models.Priority{newTestPriority(true)}, nil, true)

		Expect(err).ToNot(HaveOccurred())
	})

	It("does not fail when the delete errors while syncing deletions", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(prioritiesSavepoint).Returns(0)
		m.ExpectExec(prioritiesRunaiExcludedDelete).ReturnsError(errors.New("boom"))
		m.ExpectExec(prioritiesRollbackTo).Returns(0)
		m.ExpectTransaction().Commit()

		_, _, err := SyncRunaiPrioritiesStatus(testCtx(1), prioritiesTestCluster, nil, nil, true)

		Expect(err).ToNot(HaveOccurred())
	})

	It("returns an error when the commit fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(prioritiesSavepoint).Returns(0)

		synced, failed, err := SyncRunaiPrioritiesStatus(testCtx(1), prioritiesTestCluster, nil, nil, false)

		Expect(err).To(HaveOccurred())
		Expect(synced).To(BeNil())
		Expect(failed).To(BeNil())
	})
})
