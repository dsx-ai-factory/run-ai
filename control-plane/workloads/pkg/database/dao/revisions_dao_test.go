/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"errors"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	revisionsChildInsertSql = "INSERT INTO revisions_children (revision_id,child_id,child_type) VALUES (?,?,?) ON CONFLICT DO NOTHING"
	revisionsSavepointSql   = "SAVEPOINT sp"
	revisionsRollbackToSql  = "ROLLBACK TO SAVEPOINT sp"

	revisionsUpdateHead = "UPDATE revisions SET created_at=?,deleted_at=?,name=?,type=?,workload_id=?,tenant_id=?,cluster_id=?,"
	revisionsUpdateTail = "requested_gpu_limit=?,requested_gpu_request=?,requested_gpu_memory_limit=?,requested_gpu_memory_request=?,requested_cpu_limit=?,requested_cpu_request=?,requested_cpu_memory_limit=?,requested_cpu_memory_request=?,requested_mig_device=?,phase=?,phase_message=?,phase_updated_at=?,requested_pods_number=?,requested_pods_min=?,requested_pods_max=?,requested_pods_parallelism=?,requested_pods_completions=?,requested_nodepools=?,images=?,command=?,arguments=?,environment_variables=?,conditions=?,additional_fields=? WHERE revisions.deleted_at IS NULL AND id = ?"

	revisionsUpdateAllSql        = revisionsUpdateHead + "project_id=?,department_id=?," + revisionsUpdateTail
	revisionsUpdateOmitProjectId = revisionsUpdateHead + "department_id=?," + revisionsUpdateTail
	revisionsUpdateOmitDeptId    = revisionsUpdateHead + "project_id=?," + revisionsUpdateTail

	revisionsSyncDeleteSql          = "UPDATE revisions SET deleted_at=? WHERE tenant_id = ? AND project_id = ? AND revisions.deleted_at IS NULL"
	revisionsSyncDeleteExcludingSql = "UPDATE revisions SET deleted_at=? WHERE tenant_id = ? AND project_id = ? AND NOT id IN (?) AND revisions.deleted_at IS NULL"
	revisionsSyncDeleteRangeSql     = "UPDATE revisions SET deleted_at=? WHERE tenant_id = ? AND project_id = ? AND id >= ? AND id <= ? AND NOT id IN (?) AND revisions.deleted_at IS NULL"

	revisionsListSql         = "SELECT * FROM revisions_view WHERE tenant_id = ? AND workload_id = ? AND revisions_view.deleted_at IS NULL ORDER BY id DESC"
	revisionsListScopedSql   = "SELECT * FROM revisions_view WHERE tenant_id = ? AND workload_id = ? AND ((false OR project_id IN ( 'p1'))) AND revisions_view.deleted_at IS NULL ORDER BY id DESC"
	revisionsListDeletedSql  = "SELECT * FROM revisions_view WHERE tenant_id = ? AND workload_id = ? AND ((false OR project_id IN ( 'p1'))) AND deleted_at IS NOT NULL ORDER BY id DESC"
	revisionsCountSql        = "SELECT count(*) FROM revisions_view WHERE tenant_id = ? AND workload_id = ? AND deleted_at IS NULL"
	revisionsCountScopedSql  = "SELECT count(*) FROM revisions_view WHERE tenant_id = ? AND workload_id = ? AND ((false OR project_id IN ( 'p1'))) AND deleted_at IS NULL"
	revisionsCountDeletedSql = "SELECT count(*) FROM revisions_view WHERE tenant_id = ? AND workload_id = ? AND ((false OR project_id IN ( 'p1'))) AND deleted_at IS NOT NULL"

	revisionsChildrenSql      = "SELECT * FROM revisions_children WHERE revision_id = ?"
	revisionsChildrenTypedSql = "SELECT * FROM revisions_children WHERE revision_id = ? AND child_type IN (?)"
	revisionsGetSql           = "SELECT * FROM revisions_view WHERE tenant_id = ? AND id = ? ORDER BY revisions_view.id LIMIT ?"
	revisionsDeleteSql        = "UPDATE revisions SET deleted_at=? WHERE tenant_id = ? AND revisions.id = ? AND revisions.deleted_at IS NULL"
)

type revisionRow struct {
	ID   uuid.UUID
	Name string
}

type revisionChildRow struct {
	RevisionID uuid.UUID
	ChildID    uuid.UUID
	ChildType  string
}

func revisionsBoolPtr(b bool) *bool { return &b }

var _ = Describe("DbUpsertRevisionsChildren", func() {
	It("inserts the revision children", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(revisionsChildInsertSql).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := DbUpsertRevisionsChildren(testCtx(1), []models.RevisionChild{{RevisionID: uuid.New(), ChildID: uuid.New(), ChildType: "pod"}}, nil)

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the transaction cannot begin", func() {
		setupDaoMock(GinkgoT())

		apiErr := DbUpsertRevisionsChildren(testCtx(1), []models.RevisionChild{{RevisionID: uuid.New()}}, nil)

		Expect(apiErr).ToNot(BeNil())
	})

	It("skips children of revisions that failed to sync", func() {
		m := setupDaoMock(GinkgoT())
		failed := uuid.New()
		m.ExpectTransaction().Begin()
		m.ExpectExec(revisionsChildInsertSql).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := DbUpsertRevisionsChildren(testCtx(1),
			[]models.RevisionChild{{RevisionID: failed}, {RevisionID: uuid.New()}},
			[]string{failed.String()})

		Expect(apiErr).To(BeNil())
	})

	It("inserts nothing when every revision failed to sync", func() {
		m := setupDaoMock(GinkgoT())
		failed := uuid.New()
		m.ExpectTransaction().Begin()
		m.ExpectTransaction().Commit()

		apiErr := DbUpsertRevisionsChildren(testCtx(1),
			[]models.RevisionChild{{RevisionID: failed}},
			[]string{failed.String()})

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the insert fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(revisionsChildInsertSql).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := DbUpsertRevisionsChildren(testCtx(1), []models.RevisionChild{{RevisionID: uuid.New()}}, nil)

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the commit fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(revisionsChildInsertSql).Returns(1)

		apiErr := DbUpsertRevisionsChildren(testCtx(1), []models.RevisionChild{{RevisionID: uuid.New()}}, nil)

		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbUpsertRevisionsPodsChildren", func() {
	It("filters out the given pod children", func() {
		m := setupDaoMock(GinkgoT())
		filtered := uuid.New()
		m.ExpectTransaction().Begin()
		m.ExpectExec(revisionsChildInsertSql).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := DbUpsertRevisionsPodsChildren(testCtx(1),
			[]models.RevisionChild{{RevisionID: uuid.New(), ChildID: filtered}, {RevisionID: uuid.New(), ChildID: uuid.New()}},
			[]string{filtered.String()})

		Expect(apiErr).To(BeNil())
	})
})

var _ = Describe("DbSyncRevisions", func() {
	It("syncs the revisions", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(revisionsSavepointSql).Returns(0)
		m.ExpectExec(revisionsUpdateAllSql).Returns(1)
		m.ExpectExec(revisionsSavepointSql).Returns(0)
		m.ExpectExec(revisionsSyncDeleteExcludingSql).Returns(0)
		m.ExpectTransaction().Commit()

		synced, failed, apiErr := DbSyncRevisions(testCtx(1), 1, "", "project-1",
			[]models.Revision{{ID: uuid.New(), ClusterID: uuid.New(), ProjectID: "project-1", DepartmentID: "dep-1"}}, nil, nil, nil)

		Expect(apiErr).To(BeNil())
		Expect(synced).To(HaveLen(1))
		Expect(failed).To(BeEmpty())
	})

	It("omits an empty project id from the update", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(revisionsSavepointSql).Returns(0)
		m.ExpectExec(revisionsUpdateOmitProjectId).Returns(1)
		m.ExpectExec(revisionsSavepointSql).Returns(0)
		m.ExpectExec(revisionsSyncDeleteExcludingSql).Returns(0)
		m.ExpectTransaction().Commit()

		synced, _, apiErr := DbSyncRevisions(testCtx(1), 1, "", "project-1",
			[]models.Revision{{ID: uuid.New()}}, nil, nil, nil)

		Expect(apiErr).To(BeNil())
		Expect(synced).To(HaveLen(1))
	})

	It("omits an empty department id from the update", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(revisionsSavepointSql).Returns(0)
		m.ExpectExec(revisionsUpdateOmitDeptId).Returns(1)
		m.ExpectExec(revisionsSavepointSql).Returns(0)
		m.ExpectExec(revisionsSyncDeleteExcludingSql).Returns(0)
		m.ExpectTransaction().Commit()

		synced, _, apiErr := DbSyncRevisions(testCtx(1), 1, "", "project-1",
			[]models.Revision{{ID: uuid.New(), ProjectID: "project-1"}}, nil, nil, nil)

		Expect(apiErr).To(BeNil())
		Expect(synced).To(HaveLen(1))
	})

	It("reports the revision as failed when the save fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(revisionsSavepointSql).Returns(0)
		m.ExpectExec(revisionsUpdateOmitDeptId).ReturnsError(errors.New("boom"))
		m.ExpectExec(revisionsRollbackToSql).Returns(0)
		m.ExpectExec(revisionsSyncDeleteExcludingSql).Returns(0)
		m.ExpectTransaction().Commit()

		synced, failed, apiErr := DbSyncRevisions(testCtx(1), 1, "", "project-1",
			[]models.Revision{{ID: uuid.New(), ProjectID: "project-1"}}, nil, nil, nil)

		Expect(apiErr).To(BeNil())
		Expect(synced).To(BeEmpty())
		Expect(failed).To(HaveLen(1))
	})

	It("returns an error when the transaction cannot begin", func() {
		setupDaoMock(GinkgoT())

		synced, failed, apiErr := DbSyncRevisions(testCtx(1), 1, "", "project-1", nil, nil, nil, nil)

		Expect(apiErr).ToNot(BeNil())
		Expect(synced).To(BeNil())
		Expect(failed).To(BeNil())
	})

	It("returns an error when no project id is given", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(revisionsSavepointSql).Returns(0)
		m.ExpectTransaction().Rollback()

		synced, failed, apiErr := DbSyncRevisions(testCtx(1), 1, "", "", nil, nil, nil, nil)

		Expect(apiErr).ToNot(BeNil())
		Expect(synced).To(BeNil())
		Expect(failed).To(BeNil())
	})

	It("returns an error when the delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(revisionsSavepointSql).Returns(0)
		m.ExpectExec(revisionsSyncDeleteSql).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		_, _, apiErr := DbSyncRevisions(testCtx(1), 1, "", "project-1", nil, nil, nil, nil)

		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the commit fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(revisionsSavepointSql).Returns(0)
		m.ExpectExec(revisionsSyncDeleteSql).Returns(0)

		_, _, apiErr := DbSyncRevisions(testCtx(1), 1, "", "project-1", nil, nil, nil, nil)

		Expect(apiErr).ToNot(BeNil())
	})

	It("deletes within the given id range excluding the given ids", func() {
		m := setupDaoMock(GinkgoT())
		from := uuid.New()
		to := uuid.New()
		m.ExpectTransaction().Begin()
		m.ExpectExec(revisionsSavepointSql).Returns(0)
		m.ExpectExec(revisionsSyncDeleteRangeSql).Returns(0)
		m.ExpectTransaction().Commit()

		_, _, apiErr := DbSyncRevisions(testCtx(1), 1, "", "project-1", nil, &to, &from, []string{uuid.New().String()})

		Expect(apiErr).To(BeNil())
	})
})

var _ = Describe("DbGetWorkloadRevisions", func() {
	It("returns the revisions of the workload", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(revisionsListSql).Returns([]revisionRow{{ID: uuid.New(), Name: "rev-1"}})

		revisions, apiErr := DbGetWorkloadRevisions(testCtx(1), 1, "wl-1", nil, false)

		Expect(apiErr).To(BeNil())
		Expect(*revisions).To(HaveLen(1))
	})

	It("does not filter by scope for a system scope", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(revisionsListSql).Returns([]revisionRow{})

		_, apiErr := DbGetWorkloadRevisions(testCtx(1), 1, "wl-1", &openapi.PermittedScopes{System: revisionsBoolPtr(true)}, false)

		Expect(apiErr).To(BeNil())
	})

	It("returns the deleted revisions within the permitted scope", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(revisionsListDeletedSql).Returns([]revisionRow{})

		_, apiErr := DbGetWorkloadRevisions(testCtx(1), 1, "wl-1", &openapi.PermittedScopes{Projects: []string{"p1"}}, true)

		Expect(apiErr).To(BeNil())
	})

	It("returns the non deleted revisions within the permitted scope", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(revisionsListScopedSql).Returns([]revisionRow{})

		_, apiErr := DbGetWorkloadRevisions(testCtx(1), 1, "wl-1", &openapi.PermittedScopes{System: revisionsBoolPtr(false), Projects: []string{"p1"}}, false)

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(revisionsListSql).ReturnsError(errors.New("boom"))

		revisions, apiErr := DbGetWorkloadRevisions(testCtx(1), 1, "wl-1", nil, false)

		Expect(apiErr).ToNot(BeNil())
		Expect(*revisions).To(BeEmpty())
	})
})

var _ = Describe("DbCountWorkloadRevisions", func() {
	It("returns the number of revisions", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(revisionsCountSql).Returns([]countRow{{Count: 3}})

		count, apiErr := DbCountWorkloadRevisions(testCtx(1), 1, "wl-1", nil, false)

		Expect(apiErr).To(BeNil())
		Expect(count).To(Equal(int64(3)))
	})

	It("does not filter by scope for a system scope", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(revisionsCountSql).Returns([]countRow{{Count: 0}})

		_, apiErr := DbCountWorkloadRevisions(testCtx(1), 1, "wl-1", &openapi.PermittedScopes{System: revisionsBoolPtr(true)}, false)

		Expect(apiErr).To(BeNil())
	})

	It("counts the deleted revisions within the permitted scope", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(revisionsCountDeletedSql).Returns([]countRow{{Count: 1}})

		_, apiErr := DbCountWorkloadRevisions(testCtx(1), 1, "wl-1", &openapi.PermittedScopes{Projects: []string{"p1"}}, true)

		Expect(apiErr).To(BeNil())
	})

	It("counts the non deleted revisions within the permitted scope", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(revisionsCountScopedSql).Returns([]countRow{{Count: 1}})

		_, apiErr := DbCountWorkloadRevisions(testCtx(1), 1, "wl-1", &openapi.PermittedScopes{System: revisionsBoolPtr(false), Projects: []string{"p1"}}, false)

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(revisionsCountSql).ReturnsError(errors.New("boom"))

		count, apiErr := DbCountWorkloadRevisions(testCtx(1), 1, "wl-1", nil, false)

		Expect(apiErr).ToNot(BeNil())
		Expect(count).To(Equal(int64(0)))
	})
})

var _ = Describe("DbGetRevisionChildren", func() {
	It("returns all children when no child types are given", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(revisionsChildrenSql).Returns([]revisionChildRow{{ChildType: "pod"}})

		children, apiErr := DbGetRevisionChildren(testCtx(1), "rev-1", nil)

		Expect(apiErr).To(BeNil())
		Expect(*children).To(HaveLen(1))
	})

	It("filters the children by the given child types", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(revisionsChildrenTypedSql).Returns([]revisionChildRow{})

		_, apiErr := DbGetRevisionChildren(testCtx(1), "rev-1", &[]string{"pod"})

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(revisionsChildrenSql).ReturnsError(errors.New("boom"))

		children, apiErr := DbGetRevisionChildren(testCtx(1), "rev-1", nil)

		Expect(apiErr).ToNot(BeNil())
		Expect(*children).To(BeEmpty())
	})
})

var _ = Describe("DbGetRevision", func() {
	It("returns the revision", func() {
		m := setupDaoMock(GinkgoT())
		id := uuid.New()
		m.ExpectQuery(revisionsGetSql).Returns([]revisionRow{{ID: id, Name: "rev-1"}})

		revision, apiErr := DbGetRevision(testCtx(1), 1, id.String())

		Expect(apiErr).To(BeNil())
		Expect(revision.Name).To(Equal("rev-1"))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(revisionsGetSql).ReturnsError(errors.New("boom"))

		revision, apiErr := DbGetRevision(testCtx(1), 1, uuid.New().String())

		Expect(apiErr).ToNot(BeNil())
		Expect(revision).ToNot(BeNil())
	})
})

var _ = Describe("DbDeleteRevision", func() {
	It("deletes the revision", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(revisionsDeleteSql).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := DbDeleteRevision(testCtx(1), 1, uuid.New())

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(revisionsDeleteSql).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := DbDeleteRevision(testCtx(1), 1, uuid.New())

		Expect(apiErr).ToNot(BeNil())
	})
})
