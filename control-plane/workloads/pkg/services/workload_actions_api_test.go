/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	authorization_utils "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/helpers"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/common"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	testing "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/test"
	events "github.com/dsx-ai-factory/run-ai/runai-application-packages/async-events"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb"
	"github.com/google/uuid"
	"github.com/lib/pq"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	testActionProject    = "project-1"
	testActionDepartment = "department-1"
)

var _ = Describe("Workload action history", func() {
	var ctx context.Context
	var workloadID uuid.UUID
	var records []models.WorkloadActionHistory
	service := NewEventsService()

	BeforeEach(func() {
		ctx = testing.SetContext(testing.TestTenant, testing.TestUser, nil)
		workloadID = uuid.New()
		records = testWorkloadActionRecords(workloadID)
		Expect(dao.DbInsertWorkloadActions(ctx, records)).To(Succeed())
	})
	AfterEach(func() { testing.CleanTables() })

	It("paginates deterministically and deduplicates record IDs", func() {
		duplicate := records[0]
		duplicate.Subject = "changed"
		Expect(dao.DbInsertWorkloadActions(ctx, []models.WorkloadActionHistory{duplicate})).To(Succeed())

		first := getWorkloadActionsHistory(service, ctx, workloadID, 0, 2)
		Expect(first.Next).To(Equal(int32(2)))
		Expect(first.Actions).To(HaveLen(2))
		Expect(first.Actions[0].Id).To(Equal(testing.TestUuid3))
		Expect(first.Actions[1].Id).To(Equal(testing.TestUuid2))

		second := getWorkloadActionsHistory(service, ctx, workloadID, first.Next, 2)
		Expect(second.Next).To(BeZero())
		Expect(second.Actions).To(HaveLen(1))
		Expect(second.Actions[0].Subject).To(Equal(testing.TestUser))
		Expect(second.Actions[0].SubjectType).To(Equal("user"))
		Expect(second.Actions[0].WorkloadName).To(Equal("deleted-workspace"))
	})

	It("omits next when the final page exactly matches the limit", func() {
		response := getWorkloadActionsHistory(service, ctx, workloadID, 1, 2)
		Expect(response.Next).To(BeZero())
		Expect(response.Actions).To(HaveLen(2))
	})

	It("isolates history by tenant and workload", func() {
		requests := []struct {
			ctx context.Context
			id  uuid.UUID
		}{
			{testing.SetContext(testing.TestTenant2, testing.TestUser, nil), workloadID},
			{ctx, uuid.New()},
		}
		for _, request := range requests {
			Expect(getWorkloadActionsHistory(service, request.ctx, request.id, 0, 0).Actions).To(BeEmpty())
		}
	})

	It("requires tenant context for DAO reads", func() {
		_, _, apiErr := dao.DbGetWorkloadActions(context.Background(), workloadID, 0, 0)
		Expect(apiErr).NotTo(BeNil())
		Expect(apiErr.Code).To(Equal(int32(http.StatusBadRequest)))

	})

	It("rejects malformed IDs and negative pagination", func() {
		for _, request := range []struct {
			id            string
			offset, limit int32
		}{{"invalid", 0, 0}, {workloadID.String(), -1, 1}, {workloadID.String(), 0, -1}} {
			response, err := service.GetWorkloadActionsHistory(ctx, request.id, request.offset, request.limit)
			Expect(err).NotTo(HaveOccurred())
			Expect(response.Code).To(Equal(http.StatusBadRequest))
		}
	})

	It("defaults to 50 records and caps reads at 500", func() {
		batch := make([]models.WorkloadActionHistory, 501)
		for i := range batch {
			batch[i] = records[0]
			batch[i].ID = uuid.New()
		}
		Expect(dao.DbInsertWorkloadActions(ctx, batch)).To(Succeed())
		Expect(getWorkloadActionsHistory(service, ctx, workloadID, 0, 0).Actions).To(HaveLen(50))
		Expect(storedWorkloadActions(ctx, workloadID, 999)).To(HaveLen(500))
	})

	It("cleans tenant and cluster history without crossing tenant boundaries", func() {
		otherTenant := records[0]
		otherTenant.ID = uuid.New()
		otherTenant.TenantID = 2
		Expect(dao.DbInsertWorkloadActions(ctx, []models.WorkloadActionHistory{otherTenant})).To(Succeed())
		otherTenantCtx := testing.SetContext(testing.TestTenant2, testing.TestUser, nil)

		Expect(common.HandleClusterDeleteRequest(ctx, events.NewClusterDelete(testing.ClusterUuid1))).To(Succeed())
		Expect(storedWorkloadActions(otherTenantCtx, workloadID, 0)).To(HaveLen(1))

		Expect(common.HandleTenantDeleteRequest(ctx, events.NewTenantDelete(testing.TestTenant2))).To(Succeed())
		Expect(storedWorkloadActions(otherTenantCtx, workloadID, 0)).To(BeEmpty())
	})

	It("runs paired migrations repeatedly with all indexes and no workload foreign key", func() {
		up, err := os.ReadFile("../database/migrations/up/03010001_workloads_action_history.up.sql")
		Expect(err).NotTo(HaveOccurred())
		down, err := os.ReadFile("../database/migrations/down/03010001_workloads_action_history.down.sql")
		Expect(err).NotTo(HaveOccurred())

		tx := runaidb.DB().Begin()
		defer func() { _ = tx.Rollback() }()
		for range 2 {
			_, err = tx.Exec(string(down))
			Expect(err).NotTo(HaveOccurred())
		}
		for range 2 {
			_, err = tx.Exec(string(up))
			Expect(err).NotTo(HaveOccurred())
		}

		var count int64
		Expect(tx.Table("pg_indexes").Where("schemaname = ? AND tablename = ?", "workloads", "workloads_action_history").Count(&count)).To(Succeed())
		Expect(count).To(Equal(int64(3))) // Primary key plus two explicit indexes.
		Expect(tx.Raw("SELECT count(*) FROM pg_constraint WHERE conrelid = 'workloads.workloads_action_history'::regclass AND contype = 'f'").Row().Scan(&count)).To(Succeed())
		Expect(count).To(BeZero())

		invalidAction := records[0]
		invalidAction.ID = uuid.New()
		invalidAction.Action = "restart"
		_, err = tx.Create(&invalidAction)
		Expect(err).To(HaveOccurred())
	})

	It("aligns history cleanup with workload row deletion", func() {
		_, err := runaidb.DB().Exec("CREATE EXTENSION IF NOT EXISTS pg_cron")
		Expect(err).NotTo(HaveOccurred())

		stop := make(chan struct{})
		DeferCleanup(func() { close(stop) })
		originalConfig := *runai_config.Get[*config.DBCleanerConfig]()
		cleanerConfig := originalConfig
		cleanerConfig.CleanerIntervalHours = 1
		Expect(runai_config.Set(&cleanerConfig)).To(Succeed())
		DeferCleanup(func() { Expect(runai_config.Set(&originalConfig)).To(Succeed()) })

		database.InitGoRoutines(stop)
		var command string
		Expect(runaidb.DB().Table("cron.job").Where("jobname = ?", "workloads_cleaner").Pluck("command", &command)).To(Succeed())
		actionQuery := findQuery(command, "DELETE FROM workloads_action_history")
		Expect(actionQuery).NotTo(BeEmpty())
		Expect(actionQuery).To(ContainSubstring("NOT EXISTS"))
		Expect(actionQuery).NotTo(ContainSubstring("accepted_at"))

		workload := testActionWorkload(workloadID, records[0].WorkloadActionScope)
		_, err = runaidb.DB().Create(&workload)
		Expect(err).NotTo(HaveOccurred())
		Expect(dao.DbInsertWorkloadActions(ctx, records)).To(Succeed())
		_, err = runaidb.DB().Model(&models.WorkloadActionHistory{}).
			Where("workload_id = ?", workloadID).
			Update("accepted_at", time.Now().AddDate(-1, 0, 0))
		Expect(err).NotTo(HaveOccurred())

		orphan := records[0]
		orphan.ID = uuid.New()
		orphan.WorkloadID = uuid.New()
		Expect(dao.DbInsertWorkloadActions(ctx, []models.WorkloadActionHistory{orphan})).To(Succeed())

		var linkedCount int64
		Expect(runaidb.DB().Table("workloads_action_history AS wah").
			Joins("JOIN workloads AS w ON w.tenant_id = wah.tenant_id AND w.id = wah.workload_id").
			Where("wah.workload_id = ?", workloadID).
			Count(&linkedCount)).To(Succeed())
		Expect(linkedCount).To(Equal(int64(3)))

		_, err = runaidb.DB().Exec(actionQuery)
		Expect(err).NotTo(HaveOccurred())
		Expect(storedWorkloadActions(ctx, orphan.WorkloadID, 0)).To(BeEmpty())
		Expect(getWorkloadActionsHistory(service, ctx, workloadID, 0, 0).Actions).To(HaveLen(3))

		_, err = runaidb.DB().Model(&models.Workload{}).Where("id = ?", workloadID).Update("deleted_at", time.Now())
		Expect(err).NotTo(HaveOccurred())
		_, err = runaidb.DB().Exec(actionQuery)
		Expect(err).NotTo(HaveOccurred())
		expectWorkloadAncestorProject(ctx, workloadID, "current-project")
		Expect(getWorkloadActionsHistory(service, ctx, workloadID, 0, 0).Actions).To(HaveLen(3))

		_, err = runaidb.DB().Unscoped().Where("id = ?", workloadID).Delete(&models.Workload{})
		Expect(err).NotTo(HaveOccurred())
		_, err = runaidb.DB().Exec(actionQuery)
		Expect(err).NotTo(HaveOccurred())
		Expect(storedWorkloadActions(ctx, workloadID, 0)).To(BeEmpty())
	})
})

func testWorkloadActionRecords(workloadID uuid.UUID) []models.WorkloadActionHistory {
	acceptedAt := time.Now().UTC().Truncate(time.Microsecond)
	records := make([]models.WorkloadActionHistory, 0, 3)
	for _, id := range []string{testing.TestUuid1, testing.TestUuid2, testing.TestUuid3} {
		records = append(records, models.WorkloadActionHistory{
			ID:           uuid.MustParse(id),
			WorkloadID:   workloadID,
			WorkloadName: "deleted-workspace",
			WorkloadActionScope: models.WorkloadActionScope{
				TenantID:     testing.TestTenantId,
				ClusterID:    uuid.MustParse(testing.ClusterUuid1),
				ProjectID:    testActionProject,
				DepartmentID: testActionDepartment,
			},
			Subject:     testing.TestUser,
			SubjectType: "user",
			Action:      "delete",
			AcceptedAt:  acceptedAt,
		})
	}
	return records
}

func getWorkloadActionsHistory(service api.EventsAPIServicer, ctx context.Context, workloadID uuid.UUID, offset, limit int32) api.GetWorkloadActionsResponse {
	GinkgoHelper()
	response, err := service.GetWorkloadActionsHistory(ctx, workloadID.String(), offset, limit)
	Expect(err).NotTo(HaveOccurred())
	Expect(response.Code).To(Equal(http.StatusOK))
	return response.Body.(api.GetWorkloadActionsResponse)
}

func storedWorkloadActions(ctx context.Context, workloadID uuid.UUID, limit int32) []models.WorkloadActionHistory {
	GinkgoHelper()
	records, _, apiErr := dao.DbGetWorkloadActions(ctx, workloadID, 0, limit)
	Expect(apiErr).To(BeNil())
	return records
}

func findQuery(command, prefix string) string {
	for _, query := range strings.Split(command, ";") {
		if strings.Contains(query, prefix) {
			return strings.TrimSpace(query)
		}
	}
	return ""
}

func testActionWorkload(id uuid.UUID, scope models.WorkloadActionScope) models.Workload {
	return models.Workload{
		ID: id, TenantID: scope.TenantID, ClusterID: scope.ClusterID,
		Name: "current-name", ProjectID: "current-project", DepartmentID: "current-department",
		Type: "Workspace", Phase: api.PHASE_RUNNING, PhaseUpdatedAt: time.Now(),
		Urls: pq.StringArray{}, Images: pq.StringArray{}, RequestedMigDevice: pq.StringArray{}, RequestedNodepools: pq.StringArray{},
	}
}

func expectWorkloadAncestorProject(ctx context.Context, workloadID uuid.UUID, projectID string) {
	GinkgoHelper()
	response, err := NewWorkloadsService().GetWorkloadAncestors(ctx, workloadID.String())
	Expect(err).NotTo(HaveOccurred())
	Expect(response.Code).To(Equal(http.StatusOK))
	ancestors := response.Body.(authorization_utils.Ancestors)
	Expect(ancestors.ProjectId).NotTo(BeNil())
	Expect(*ancestors.ProjectId).To(Equal(projectID))
}
