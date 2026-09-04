/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"errors"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	applicationsSavepoint  = "SAVEPOINT sp"
	applicationsRollbackTo = "ROLLBACK TO SAVEPOINT sp"

	applicationsSaveUpdate    = "UPDATE applications SET updated_at=?,deleted_at=?,updated_by=?,deleted_by=?,name=?,source=?,type=?,actual_phase=?,tenant_id=?,cluster_id=?,department_id=?,department_name=?,project_id=?,project_name=?,k8s_phase=?,resources=?,message=? WHERE applications.deleted_at IS NULL AND id = ?"
	applicationsFallbackUpdte = "UPDATE applications SET k8s_phase=?,resources=?,updated_at=? WHERE id = ? AND tenant_id = ? AND cluster_id = ?"
	applicationsDelete        = "UPDATE applications SET deleted_at=? WHERE tenant_id = ? AND cluster_id = ? AND id in (?) AND applications.deleted_at IS NULL"
	applicationsDeleteExcl    = "UPDATE applications SET deleted_at=? WHERE tenant_id = ? AND cluster_id = ? AND applications.deleted_at IS NULL"
	applicationsDeleteExclIn  = "UPDATE applications SET deleted_at=? WHERE tenant_id = ? AND cluster_id = ? AND id not in (?) AND applications.deleted_at IS NULL"

	applicationsVerboseSelect    = "SELECT * FROM applications_view WHERE tenant_id = ? AND applications_view.deleted_at IS NULL ORDER BY id DESC"
	applicationsNonVerboseSelect = "SELECT applications.*, acs.spec, acs.desired_phase, acs.update_count FROM applications LEFT JOIN (SELECT application_id, spec, desired_phase, update_count FROM application_crud_sync) AS acs ON acs.application_id = applications.id WHERE tenant_id = ? AND applications.deleted_at IS NULL ORDER BY id DESC"
	applicationsScopedSelect     = "SELECT * FROM applications_view WHERE tenant_id = ? AND (false or tenant_id in (?) or cluster_id in (?) or department_id in (?) or project_id in (?)) AND applications_view.deleted_at IS NULL ORDER BY id DESC"
	applicationsFullSelect       = "SELECT * FROM applications_view WHERE tenant_id = ? AND name = ? AND (search_vector @@ to_tsquery('english', ? || ':*')) AND applications_view.deleted_at IS NULL ORDER BY name DESC NULLS LAST,id DESC LIMIT ? OFFSET ?"
	applicationsCountSelect      = "SELECT count(*) FROM applications WHERE tenant_id = ? AND applications.deleted_at IS NULL"
	applicationsCountScoped      = "SELECT count(*) FROM applications WHERE tenant_id = ? AND (false or project_id in (?)) AND (search_vector @@ to_tsquery('english', ? || ':*')) AND applications.deleted_at IS NULL"
	applicationsByIdSelect       = "SELECT * FROM applications_view WHERE tenant_id = ? AND id = ? AND applications_view.deleted_at IS NULL ORDER BY applications_view.id LIMIT ?"
	applicationsByIdScoped       = "SELECT * FROM applications_view WHERE tenant_id = ? AND id = ? AND (false or cluster_id in (?)) AND applications_view.deleted_at IS NULL ORDER BY applications_view.id LIMIT ?"
	applicationsInsert           = "INSERT INTO applications (id,created_at,updated_at,deleted_at,created_by,updated_by,deleted_by,name,source,type,actual_phase,tenant_id,cluster_id,department_id,department_name,project_id,project_name,k8s_phase,resources,message) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
	applicationsCrudSyncInsert   = "INSERT INTO application_crud_sync (application_id,tenant_id,cluster_id,spec,desired_phase,update_count) VALUES (?,?,?,?,?,?)"
	applicationsMarkDeleting     = "UPDATE applications SET actual_phase=?,updated_at=?,updated_by=? WHERE applications.tenant_id = ? AND id IN (?) AND (actual_phase != ? OR actual_phase IS NULL) AND actual_phase != ? AND applications.deleted_at IS NULL"
	applicationsListUpdates      = "SELECT a.id, a.created_at, a.created_by, a.name, a.source, a.type, a.tenant_id, a.cluster_id, a.department_id, a.project_id, a.project_name, a.actual_phase, s.spec, s.desired_phase, s.update_count FROM application_crud_sync AS s JOIN applications AS a ON s.application_id = a.id AND (a.deleted_at IS NULL OR s.desired_phase = ?) WHERE s.tenant_id = ? AND s.cluster_id = ? AND s.update_count > ? ORDER BY s.update_count LIMIT ?"
	applicationsHighestCount     = "SELECT last_value FROM application_crud_sync_sequence"
	applicationsUpdateMessage    = "UPDATE applications SET message=?,updated_at=? WHERE (applications.id = ? AND applications.tenant_id = ? AND applications.cluster_id = ?) AND applications.deleted_at IS NULL"
	applicationsUpdateDesired    = "UPDATE application_crud_sync SET desired_phase=? WHERE application_crud_sync.application_id = ? AND application_crud_sync.tenant_id = ?"
	applicationsSoftDelete       = "UPDATE applications SET deleted_at=? WHERE (applications.id = ? AND applications.tenant_id = ?) AND applications.deleted_at IS NULL"
	applicationsCleanupCrudSync  = "DELETE FROM application_crud_sync WHERE (application_crud_sync.tenant_id = ? AND application_crud_sync.cluster_id = ? AND application_crud_sync.desired_phase = ?) AND update_count < ?"
)

type applicationRow struct {
	ID   uuid.UUID
	Name string
}

type applicationLastValueRow struct {
	LastValue int64
}

var (
	applicationsTestId      = uuid.MustParse("33333333-3333-3333-3333-333333333333")
	applicationsTestCluster = uuid.MustParse("44444444-4444-4444-4444-444444444444")
)

func newTestApplication() models.Application {
	app := models.Application{
		Name:      "app-1",
		Source:    models.ApplicationSourceRunai,
		Type:      "inference",
		TenantId:  "tenant-1",
		ClusterId: applicationsTestCluster.String(),
		ApplicationStatus: models.ApplicationStatus{
			K8sPhase:  "Running",
			Resources: models.ApplicationResourceList{{Name: "res", Id: "res-1", Kind: "Deployment"}},
		},
	}
	app.ID = applicationsTestId
	return app
}

var _ = Describe("SyncClusterApplications", func() {
	It("fails when the transaction cannot be started", func() {
		setupDaoMock(GinkgoT())

		synced, failed, err := SyncClusterApplications(testCtx(1), "tenant-1", applicationsTestCluster, nil, nil, false)

		Expect(err).To(HaveOccurred())
		Expect(synced).To(BeNil())
		Expect(failed).To(BeNil())
	})

	It("upserts the given applications successfully", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(applicationsSavepoint).Returns(0)
		m.ExpectExec(applicationsSaveUpdate).Returns(1)
		m.ExpectExec(applicationsSavepoint).Returns(0)
		m.ExpectTransaction().Commit()

		synced, failed, err := SyncClusterApplications(testCtx(1), "tenant-1", applicationsTestCluster, []models.Application{newTestApplication()}, nil, false)

		Expect(err).ToNot(HaveOccurred())
		Expect(synced).To(HaveLen(1))
		Expect(failed).To(BeEmpty())
	})

	It("falls back to a partial update on a unique constraint violation", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(applicationsSavepoint).Returns(0)
		m.ExpectExec(applicationsSaveUpdate).ReturnsError(newPgUniqueError(UniqueApplicationConstraint))
		m.ExpectExec(applicationsRollbackTo).Returns(0)
		m.ExpectExec(applicationsFallbackUpdte).Returns(1)
		m.ExpectExec(applicationsSavepoint).Returns(0)
		m.ExpectTransaction().Commit()

		synced, failed, err := SyncClusterApplications(testCtx(1), "tenant-1", applicationsTestCluster, []models.Application{newTestApplication()}, nil, false)

		Expect(err).ToNot(HaveOccurred())
		Expect(synced).To(Equal([]string{applicationsTestId.String()}))
		Expect(failed).To(BeEmpty())
	})

	It("reports the application as failed when the unique violation fallback update errors", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(applicationsSavepoint).Returns(0)
		m.ExpectExec(applicationsSaveUpdate).ReturnsError(newPgUniqueError(UniqueApplicationConstraint))
		m.ExpectExec(applicationsRollbackTo).Returns(0)
		m.ExpectExec(applicationsFallbackUpdte).ReturnsError(errors.New("boom"))
		m.ExpectExec(applicationsRollbackTo).Returns(0)
		m.ExpectTransaction().Commit()

		synced, failed, err := SyncClusterApplications(testCtx(1), "tenant-1", applicationsTestCluster, []models.Application{newTestApplication()}, nil, false)

		Expect(err).ToNot(HaveOccurred())
		Expect(synced).To(BeEmpty())
		Expect(failed).To(Equal([]string{applicationsTestId.String()}))
	})

	It("marks the application as failed on a non unique constraint error", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(applicationsSavepoint).Returns(0)
		m.ExpectExec(applicationsSaveUpdate).ReturnsError(newPgUniqueError("other_constraint"))
		m.ExpectExec(applicationsRollbackTo).Returns(0)
		m.ExpectTransaction().Commit()

		synced, failed, err := SyncClusterApplications(testCtx(1), "tenant-1", applicationsTestCluster, []models.Application{newTestApplication()}, nil, false)

		Expect(err).ToNot(HaveOccurred())
		Expect(synced).To(BeEmpty())
		Expect(failed).To(HaveLen(1))
	})

	It("deletes the given applications successfully", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(applicationsSavepoint).Returns(0)
		m.ExpectExec(applicationsDelete).Returns(1)
		m.ExpectExec(applicationsSavepoint).Returns(0)
		m.ExpectTransaction().Commit()

		synced, failed, err := SyncClusterApplications(testCtx(1), "tenant-1", applicationsTestCluster, nil, []string{applicationsTestId.String()}, false)

		Expect(err).ToNot(HaveOccurred())
		Expect(synced).To(HaveLen(1))
		Expect(failed).To(BeEmpty())
	})

	It("marks the application as failed when the delete errors", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(applicationsSavepoint).Returns(0)
		m.ExpectExec(applicationsDelete).ReturnsError(errors.New("boom"))
		m.ExpectExec(applicationsRollbackTo).Returns(0)
		m.ExpectTransaction().Commit()

		synced, failed, err := SyncClusterApplications(testCtx(1), "tenant-1", applicationsTestCluster, nil, []string{applicationsTestId.String()}, false)

		Expect(err).ToNot(HaveOccurred())
		Expect(synced).To(BeEmpty())
		Expect(failed).To(HaveLen(1))
	})

	It("syncs deletions when there are no excluded applications", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(applicationsSavepoint).Returns(0)
		m.ExpectExec(applicationsDeleteExcl).Returns(1)
		m.ExpectExec(applicationsSavepoint).Returns(0)
		m.ExpectTransaction().Commit()

		_, _, err := SyncClusterApplications(testCtx(1), "tenant-1", applicationsTestCluster, nil, nil, true)

		Expect(err).ToNot(HaveOccurred())
	})

	It("syncs deletions while excluding the upserted applications", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(applicationsSavepoint).Returns(0)
		m.ExpectExec(applicationsSaveUpdate).Returns(1)
		m.ExpectExec(applicationsSavepoint).Returns(0)
		m.ExpectExec(applicationsDeleteExclIn).Returns(1)
		m.ExpectExec(applicationsSavepoint).Returns(0)
		m.ExpectTransaction().Commit()

		_, _, err := SyncClusterApplications(testCtx(1), "tenant-1", applicationsTestCluster, []models.Application{newTestApplication()}, nil, true)

		Expect(err).ToNot(HaveOccurred())
	})

	It("does not return an error when syncing deletions fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(applicationsSavepoint).Returns(0)
		m.ExpectExec(applicationsDeleteExcl).ReturnsError(errors.New("boom"))
		m.ExpectExec(applicationsRollbackTo).Returns(0)
		m.ExpectTransaction().Commit()

		_, _, err := SyncClusterApplications(testCtx(1), "tenant-1", applicationsTestCluster, nil, nil, true)

		Expect(err).ToNot(HaveOccurred())
	})

	It("fails when the transaction cannot be committed", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(applicationsSavepoint).Returns(0)

		synced, failed, err := SyncClusterApplications(testCtx(1), "tenant-1", applicationsTestCluster, nil, nil, false)

		Expect(err).To(HaveOccurred())
		Expect(synced).To(BeNil())
		Expect(failed).To(BeNil())
	})
})

var _ = Describe("GetApplications", func() {
	It("queries the view when verbose is requested", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(applicationsVerboseSelect).Returns([]applicationRow{{ID: applicationsTestId, Name: "app-1"}})

		applications, err := GetApplications(testCtx(1), "tenant-1", nil, 0, 0, "", "", nil, "", true)

		Expect(err).ToNot(HaveOccurred())
		Expect(*applications).To(HaveLen(1))
	})

	It("joins the crud sync table when verbose is not requested", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(applicationsNonVerboseSelect).Returns([]applicationRow{{ID: applicationsTestId, Name: "app-1"}})

		applications, err := GetApplications(testCtx(1), "tenant-1", nil, 0, 0, "", "", nil, "", false)

		Expect(err).ToNot(HaveOccurred())
		Expect(*applications).To(HaveLen(1))
	})

	It("does not filter by scope for a system scoped caller", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(applicationsVerboseSelect).Returns([]applicationRow{})

		scopes := &openapi.PermittedScopes{System: convutils.Ref(true)}
		_, err := GetApplications(testCtx(1), "tenant-1", scopes, 0, 0, "", "", nil, "", true)

		Expect(err).ToNot(HaveOccurred())
	})

	It("filters by the permitted scopes", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(applicationsScopedSelect).Returns([]applicationRow{})

		scopes := &openapi.PermittedScopes{
			Tenants:     []string{"tenant-1"},
			Clusters:    []string{"cluster-1"},
			Departments: []string{"department-1"},
			Projects:    []string{"project-1"},
		}
		_, err := GetApplications(testCtx(1), "tenant-1", scopes, 0, 0, "", "", nil, "", true)

		Expect(err).ToNot(HaveOccurred())
	})

	It("applies filters, sorting, paging and free text search", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(applicationsFullSelect).Returns([]applicationRow{})

		filters := []queryapi.Filter{{Field: "name", Operator: "=", Values: []string{"app-1"}}}
		_, err := GetApplications(testCtx(1), "tenant-1", nil, 10, 20, "DESC", "name", filters, "app", true)

		Expect(err).ToNot(HaveOccurred())
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(applicationsVerboseSelect).ReturnsError(errors.New("boom"))

		applications, err := GetApplications(testCtx(1), "tenant-1", nil, 0, 0, "", "", nil, "", true)

		Expect(err).To(HaveOccurred())
		Expect(applications).To(BeNil())
	})
})

var _ = Describe("CountApplications", func() {
	It("returns the total count", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(applicationsCountSelect).Returns([]countRow{{Count: 3}})

		count, err := CountApplications(testCtx(1), "tenant-1", nil, nil, "")

		Expect(err).ToNot(HaveOccurred())
		Expect(count).To(Equal(int64(3)))
	})

	It("counts with permitted scopes and free text search", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(applicationsCountScoped).Returns([]countRow{{Count: 1}})

		scopes := &openapi.PermittedScopes{Projects: []string{"project-1"}}
		count, err := CountApplications(testCtx(1), "tenant-1", scopes, nil, "app")

		Expect(err).ToNot(HaveOccurred())
		Expect(count).To(Equal(int64(1)))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(applicationsCountSelect).ReturnsError(errors.New("boom"))

		count, err := CountApplications(testCtx(1), "tenant-1", nil, nil, "")

		Expect(err).To(HaveOccurred())
		Expect(count).To(BeZero())
	})
})

var _ = Describe("GetApplicationById", func() {
	It("returns the application", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(applicationsByIdSelect).Returns([]applicationRow{{ID: applicationsTestId, Name: "app-1"}})

		application, err := GetApplicationById(testCtx(1), "tenant-1", applicationsTestId, nil)

		Expect(err).ToNot(HaveOccurred())
		Expect(application.ID).To(Equal(applicationsTestId))
	})

	It("filters by the permitted scopes", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(applicationsByIdScoped).Returns([]applicationRow{{ID: applicationsTestId, Name: "app-1"}})

		scopes := &openapi.PermittedScopes{Clusters: []string{"cluster-1"}}
		application, err := GetApplicationById(testCtx(1), "tenant-1", applicationsTestId, scopes)

		Expect(err).ToNot(HaveOccurred())
		Expect(application.ID).To(Equal(applicationsTestId))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(applicationsByIdSelect).ReturnsError(errors.New("boom"))

		application, err := GetApplicationById(testCtx(1), "tenant-1", applicationsTestId, nil)

		Expect(err).To(HaveOccurred())
		Expect(application).To(BeNil())
	})
})

var _ = Describe("CreateApplication", func() {
	It("fails when the transaction cannot be started", func() {
		setupDaoMock(GinkgoT())

		application := newTestApplication()
		Expect(CreateApplication(testCtx(1), &application, nil, nil)).ToNot(BeNil())
	})

	It("creates the application without a spec", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(applicationsInsert).Returns(1)
		m.ExpectTransaction().Commit()

		application := newTestApplication()
		Expect(CreateApplication(testCtx(1), &application, nil, nil)).To(BeNil())
	})

	It("creates the application together with its crud sync record", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(applicationsInsert).Returns(1)
		m.ExpectExec(applicationsCrudSyncInsert).Returns(1)
		m.ExpectTransaction().Commit()

		application := newTestApplication()
		Expect(CreateApplication(testCtx(1), &application, &api.AiApplicationSpec{}, convutils.Ref(models.ApplicationPhaseCreating))).To(BeNil())
	})

	It("fails when the insert errors", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(applicationsInsert).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		application := newTestApplication()
		Expect(CreateApplication(testCtx(1), &application, nil, nil)).ToNot(BeNil())
	})

	It("fails when the crud sync insert errors", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(applicationsInsert).Returns(1)
		m.ExpectExec(applicationsCrudSyncInsert).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		application := newTestApplication()
		Expect(CreateApplication(testCtx(1), &application, &api.AiApplicationSpec{}, nil)).ToNot(BeNil())
	})

	It("fails when the transaction cannot be committed", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(applicationsInsert).Returns(1)

		application := newTestApplication()
		Expect(CreateApplication(testCtx(1), &application, nil, nil)).ToNot(BeNil())
	})
})

var _ = Describe("MarkApplicationsAsDeleting", func() {
	It("does nothing when no ids are given", func() {
		setupDaoMock(GinkgoT())

		Expect(MarkApplicationsAsDeleting(testCtx(1), "tenant-1", nil, "user")).To(BeNil())
	})

	It("marks the given applications as deleting", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(applicationsMarkDeleting).Returns(1)
		m.ExpectTransaction().Commit()

		Expect(MarkApplicationsAsDeleting(testCtx(1), "tenant-1", []uuid.UUID{applicationsTestId}, "user")).To(BeNil())
	})

	It("returns an error when the update fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(applicationsMarkDeleting).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		Expect(MarkApplicationsAsDeleting(testCtx(1), "tenant-1", []uuid.UUID{applicationsTestId}, "user")).ToNot(BeNil())
	})
})

var _ = Describe("ListApplicationUpdates", func() {
	It("returns the pending update records", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(applicationsListUpdates).Returns([]applicationRow{{ID: applicationsTestId, Name: "app-1"}})

		records, apiErr := ListApplicationUpdates(testCtx(1), "tenant-1", "cluster-1", 0, 10)

		Expect(apiErr).To(BeNil())
		Expect(records).To(HaveLen(1))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(applicationsListUpdates).ReturnsError(errors.New("boom"))

		records, apiErr := ListApplicationUpdates(testCtx(1), "tenant-1", "cluster-1", 0, 10)

		Expect(apiErr).ToNot(BeNil())
		Expect(records).To(BeNil())
	})
})

var _ = Describe("GetApplicationsHighestUpdateCount", func() {
	It("returns the sequence last value", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(applicationsHighestCount).Returns([]applicationLastValueRow{{LastValue: 7}})

		count, apiErr := GetApplicationsHighestUpdateCount(testCtx(1))

		Expect(apiErr).To(BeNil())
		Expect(count).To(Equal(int64(7)))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(applicationsHighestCount).ReturnsError(errors.New("boom"))

		count, apiErr := GetApplicationsHighestUpdateCount(testCtx(1))

		Expect(apiErr).ToNot(BeNil())
		Expect(count).To(BeZero())
	})
})

var _ = Describe("UpdateApplicationSyncMessage", func() {
	It("returns an error for an invalid application id", func() {
		setupDaoMock(GinkgoT())

		Expect(UpdateApplicationSyncMessage(testCtx(1), "tenant-1", "cluster-1", "not-a-uuid", "message")).ToNot(BeNil())
	})

	It("updates the sync message", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(applicationsUpdateMessage).Returns(1)
		m.ExpectTransaction().Commit()

		Expect(UpdateApplicationSyncMessage(testCtx(1), "tenant-1", "cluster-1", applicationsTestId.String(), "message")).To(BeNil())
	})

	It("returns an error when the update fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(applicationsUpdateMessage).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		Expect(UpdateApplicationSyncMessage(testCtx(1), "tenant-1", "cluster-1", applicationsTestId.String(), "message")).ToNot(BeNil())
	})
})

var _ = Describe("UpdateCrudSyncDesiredPhase", func() {
	It("updates the desired phase", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(applicationsUpdateDesired).Returns(1)
		m.ExpectTransaction().Commit()

		Expect(UpdateCrudSyncDesiredPhase(testCtx(1), "tenant-1", applicationsTestId, models.ApplicationPhaseDeleted)).To(BeNil())
	})

	It("returns an error when the update fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(applicationsUpdateDesired).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		Expect(UpdateCrudSyncDesiredPhase(testCtx(1), "tenant-1", applicationsTestId, models.ApplicationPhaseDeleted)).ToNot(BeNil())
	})
})

var _ = Describe("DeleteApplicationImmediate", func() {
	It("fails when the transaction cannot be started", func() {
		setupDaoMock(GinkgoT())

		Expect(DeleteApplicationImmediate(testCtx(1), "tenant-1", applicationsTestId)).ToNot(BeNil())
	})

	It("soft deletes the application and updates the desired phase", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(applicationsSoftDelete).Returns(1)
		m.ExpectExec(applicationsUpdateDesired).Returns(1)
		m.ExpectTransaction().Commit()

		Expect(DeleteApplicationImmediate(testCtx(1), "tenant-1", applicationsTestId)).To(BeNil())
	})

	It("fails when the soft delete errors", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(applicationsSoftDelete).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		Expect(DeleteApplicationImmediate(testCtx(1), "tenant-1", applicationsTestId)).ToNot(BeNil())
	})

	It("fails when the desired phase update errors", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(applicationsSoftDelete).Returns(1)
		m.ExpectExec(applicationsUpdateDesired).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		Expect(DeleteApplicationImmediate(testCtx(1), "tenant-1", applicationsTestId)).ToNot(BeNil())
	})

	It("fails when the transaction cannot be committed", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(applicationsSoftDelete).Returns(1)
		m.ExpectExec(applicationsUpdateDesired).Returns(1)

		Expect(DeleteApplicationImmediate(testCtx(1), "tenant-1", applicationsTestId)).ToNot(BeNil())
	})
})

var _ = Describe("CleanupApplicationCrudSync", func() {
	It("deletes the obsolete crud sync records", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(applicationsCleanupCrudSync).Returns(1)
		m.ExpectTransaction().Commit()

		Expect(CleanupApplicationCrudSync(testCtx(1), "tenant-1", "cluster-1", 5)).To(BeNil())
	})

	It("returns an error when the delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(applicationsCleanupCrudSync).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		Expect(CleanupApplicationCrudSync(testCtx(1), "tenant-1", "cluster-1", 5)).ToNot(BeNil())
	})
})

var _ = Describe("getPermissionQuery", func() {
	It("returns a false predicate when no scopes are given", func() {
		query, values := getPermissionQuery(&openapi.PermittedScopes{})

		Expect(query).To(Equal("false"))
		Expect(values).To(BeEmpty())
	})

	It("returns a predicate covering every scope kind", func() {
		query, values := getPermissionQuery(&openapi.PermittedScopes{
			Tenants:     []string{"tenant-1"},
			Clusters:    []string{"cluster-1"},
			Departments: []string{"department-1"},
			Projects:    []string{"project-1"},
		})

		Expect(query).To(Equal("false or tenant_id in ? or cluster_id in ? or department_id in ? or project_id in ?"))
		Expect(values).To(HaveLen(4))
	})
})
