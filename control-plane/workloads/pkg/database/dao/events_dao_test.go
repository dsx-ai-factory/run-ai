/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"errors"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	eventsInsertSql           = "INSERT INTO events (tenant_id,id,cluster_id,type,message,reason,created_at,source,involved_object_uid,involved_object_kind,involved_object_name,involved_object_namespace) VALUES (?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT DO NOTHING"
	eventsWorkloadSelectAsc   = "SELECT * FROM workloads_events_view WHERE tenant_id = ? AND workload_id = ? ORDER BY created_at asc"
	eventsWorkloadSelectDesc  = "SELECT * FROM workloads_events_view WHERE tenant_id = ? AND workload_id = ? ORDER BY created_at DESC"
	eventsWorkloadSelectPaged = "SELECT * FROM workloads_events_view WHERE tenant_id = ? AND workload_id = ? ORDER BY created_at asc LIMIT ? OFFSET ?"
	eventsRevisionSelectAsc   = "SELECT * FROM revisions_events_view WHERE tenant_id = ? AND revision_id = ? ORDER BY created_at asc"

	eventsWorkloadHistorySql      = "SELECT coalesce(e.workload_id, p.workload_id) as workload_id,coalesce(e.created_at, p.created_at) as created_at,coalesce(e.cluster_id, p.cluster_id) as cluster_id,coalesce(e.tenant_id, p.tenant_id) as tenant_id,e.event_id,e.type,e.message,e.reason,e.source,e.involved_object_uid,e.involved_object_kind,e.involved_object_name,e.involved_object_namespace,p.phase,p.phase_message FROM (SELECT workload_id, created_at, cluster_id, tenant_id, event_id, type, message, reason, source, involved_object_uid, involved_object_kind, involved_object_name, involved_object_namespace FROM workloads_events_view WHERE tenant_id = ? AND workload_id = ?) as e FULL OUTER JOIN (SELECT workload_id, created_at, cluster_id, tenant_id, phase, phase_message FROM workloads_phase_history WHERE tenant_id = ? AND workload_id = ?) as p on 1=0 ORDER BY created_at desc"
	eventsWorkloadHistoryPagedSql = eventsWorkloadHistorySql + " LIMIT ? OFFSET ?"
	eventsRevisionHistorySql      = "SELECT coalesce(e.revision_id, p.revision_id) as revision_id,coalesce(e.created_at, p.created_at) as created_at,coalesce(e.cluster_id, p.cluster_id) as cluster_id,coalesce(e.tenant_id, p.tenant_id) as tenant_id,e.event_id,e.type,e.message,e.reason,e.source,e.involved_object_uid,e.involved_object_kind,e.involved_object_name,e.involved_object_namespace,p.phase,p.phase_message FROM (SELECT revision_id, created_at, cluster_id, tenant_id, event_id, type, message, reason, source, involved_object_uid, involved_object_kind, involved_object_name, involved_object_namespace FROM revisions_events_view WHERE tenant_id = ? AND revision_id = ?) as e FULL OUTER JOIN (SELECT revision_id, created_at, cluster_id, tenant_id, phase, phase_message FROM revisions_phase_history WHERE tenant_id = ? AND revision_id = ?) as p on 1=0 ORDER BY created_at desc"

	eventsPhaseHistorySql    = "SELECT * FROM workloads_phases_view WHERE tenant_id = ? AND created_at > ? ORDER BY created_at asc"
	eventsTimeoutSql         = "SELECT * FROM timeout_events_of_workload_view WHERE tenant_id = ? AND reason = ? AND created_at > ?"
	eventsDeleteByClusterSql = "DELETE FROM events WHERE tenant_id = ? AND cluster_id = ?"
	eventsDeleteByTenantSql  = "DELETE FROM events WHERE tenant_id = ?"
)

type eventRow struct {
	EventID uuid.UUID
	Type    string
}

var _ = Describe("DbSyncEvents", func() {
	It("does nothing when there are no events", func() {
		setupDaoMock(GinkgoT())

		apiErr := DbSyncEvents(testCtx(1), []models.Event{})

		Expect(apiErr).To(BeNil())
	})

	It("inserts the events", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(eventsInsertSql).Returns(1)
		m.ExpectTransaction().Commit()

		apiErr := DbSyncEvents(testCtx(1), []models.Event{{ID: uuid.New(), TenantID: 1}})

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the insert fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(eventsInsertSql).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		apiErr := DbSyncEvents(testCtx(1), []models.Event{{ID: uuid.New(), TenantID: 1}})

		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbGetWorkloadEvents", func() {
	It("returns the events sorted ascending", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(eventsWorkloadSelectAsc).Returns([]eventRow{{Type: "Normal"}})

		events, apiErr := DbGetWorkloadEvents(testCtx(1), 1, "wl-1", 0, 0, "asc")

		Expect(apiErr).To(BeNil())
		Expect(*events).To(HaveLen(1))
	})

	It("returns the events sorted descending", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(eventsWorkloadSelectDesc).Returns([]eventRow{})

		events, apiErr := DbGetWorkloadEvents(testCtx(1), 1, "wl-1", 0, 0, "DESC")

		Expect(apiErr).To(BeNil())
		Expect(*events).To(BeEmpty())
	})

	It("applies the offset and limit", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(eventsWorkloadSelectPaged).Returns([]eventRow{})

		_, apiErr := DbGetWorkloadEvents(testCtx(1), 1, "wl-1", 5, 10, "asc")

		Expect(apiErr).To(BeNil())
	})

	It("returns an error for an invalid sort order", func() {
		setupDaoMock(GinkgoT())

		events, apiErr := DbGetWorkloadEvents(testCtx(1), 1, "wl-1", 0, 0, "sideways")

		Expect(apiErr).ToNot(BeNil())
		Expect(*events).To(BeEmpty())
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(eventsWorkloadSelectAsc).ReturnsError(errors.New("boom"))

		events, apiErr := DbGetWorkloadEvents(testCtx(1), 1, "wl-1", 0, 0, "asc")

		Expect(apiErr).ToNot(BeNil())
		Expect(*events).To(BeEmpty())
	})
})

var _ = Describe("DbGetRevisionEvents", func() {
	It("returns the revision events", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(eventsRevisionSelectAsc).Returns([]eventRow{{Type: "Warning"}})

		events, apiErr := DbGetRevisionEvents(testCtx(1), 1, "rev-1", 0, 0, "asc")

		Expect(apiErr).To(BeNil())
		Expect(*events).To(HaveLen(1))
	})
})

var _ = Describe("dbGetWorkloadOrRevisionEvents", func() {
	It("returns an error for an unknown entity type", func() {
		setupDaoMock(GinkgoT())

		events, apiErr := dbGetWorkloadOrRevisionEvents[models.WorkloadEventView](
			testCtx(1), 1, "bogus", "wl-1", EventsViewName, 0, 0, "asc")

		Expect(apiErr).ToNot(BeNil())
		Expect(*events).To(BeEmpty())
	})
})

var _ = Describe("DbGetWorkloadHistoryRecords", func() {
	It("returns the history records", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(eventsWorkloadHistorySql).Returns([]eventRow{{Type: "Normal"}})

		records, apiErr := DbGetWorkloadHistoryRecords(testCtx(1), 1, "wl-1", 0, 0)

		Expect(apiErr).To(BeNil())
		Expect(*records).To(HaveLen(1))
	})

	It("applies the offset and limit", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(eventsWorkloadHistoryPagedSql).Returns([]eventRow{})

		_, apiErr := DbGetWorkloadHistoryRecords(testCtx(1), 1, "wl-1", 5, 10)

		Expect(apiErr).To(BeNil())
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(eventsWorkloadHistorySql).ReturnsError(errors.New("boom"))

		records, apiErr := DbGetWorkloadHistoryRecords(testCtx(1), 1, "wl-1", 0, 0)

		Expect(apiErr).ToNot(BeNil())
		Expect(*records).To(BeEmpty())
	})
})

var _ = Describe("DbGetRevisionHistoryRecords", func() {
	It("returns the revision history records", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(eventsRevisionHistorySql).Returns([]eventRow{{Type: "Normal"}})

		records, apiErr := DbGetRevisionHistoryRecords(testCtx(1), 1, "rev-1", 0, 0)

		Expect(apiErr).To(BeNil())
		Expect(*records).To(HaveLen(1))
	})
})

var _ = Describe("dbGetWorkloadOrRevisionHistoryRecords", func() {
	It("returns an error for an unknown entity type", func() {
		setupDaoMock(GinkgoT())

		records, apiErr := dbGetWorkloadOrRevisionHistoryRecords[models.WorkloadHistoryRecord](
			testCtx(1), 1, "bogus", "wl-1", EventsViewName, PhaseHistoryTableName, 0, 0)

		Expect(apiErr).ToNot(BeNil())
		Expect(*records).To(BeEmpty())
	})
})

var _ = Describe("DbGetTenantWorkloadPhaseHistoryRecords", func() {
	It("returns the phase history records", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(eventsPhaseHistorySql).Returns([]eventRow{{Type: "Running"}})

		records, apiErr := DbGetTenantWorkloadPhaseHistoryRecords(testCtx(1), "1", time.Unix(0, 0).UTC())

		Expect(apiErr).To(BeNil())
		Expect(*records).To(HaveLen(1))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(eventsPhaseHistorySql).ReturnsError(errors.New("boom"))

		records, apiErr := DbGetTenantWorkloadPhaseHistoryRecords(testCtx(1), "1", time.Unix(0, 0).UTC())

		Expect(apiErr).ToNot(BeNil())
		Expect(*records).To(BeEmpty())
	})
})

var _ = Describe("DbGetTimeoutEventsWithReason", func() {
	It("returns the timeout events", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(eventsTimeoutSql).Returns([]eventRow{{Type: "Warning"}})

		events, apiErr := DbGetTimeoutEventsWithReason(testCtx(1), "1", "Timeout", time.Unix(0, 0).UTC())

		Expect(apiErr).To(BeNil())
		Expect(*events).To(HaveLen(1))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(eventsTimeoutSql).ReturnsError(errors.New("boom"))

		events, apiErr := DbGetTimeoutEventsWithReason(testCtx(1), "1", "Timeout", time.Unix(0, 0).UTC())

		Expect(apiErr).ToNot(BeNil())
		Expect(*events).To(BeEmpty())
	})
})

var _ = Describe("DbDeleteWorkloadEventsOfClusterId", func() {
	It("deletes the events of the cluster", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(eventsDeleteByClusterSql).Returns(3)
		m.ExpectTransaction().Commit()

		err := DbDeleteWorkloadEventsOfClusterId(testCtx(1), "1", "cluster-1")

		Expect(err).ToNot(HaveOccurred())
	})

	It("returns an error when the delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(eventsDeleteByClusterSql).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		err := DbDeleteWorkloadEventsOfClusterId(testCtx(1), "1", "cluster-1")

		Expect(err).To(HaveOccurred())
	})
})

var _ = Describe("DbDeleteWorkloadEventsOfTenantId", func() {
	It("deletes the events of the tenant", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(eventsDeleteByTenantSql).Returns(3)
		m.ExpectTransaction().Commit()

		err := DbDeleteWorkloadEventsOfTenantId(testCtx(1), "1")

		Expect(err).ToNot(HaveOccurred())
	})

	It("returns an error when the delete fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectTransaction().Begin()
		m.ExpectExec(eventsDeleteByTenantSql).ReturnsError(errors.New("boom"))
		m.ExpectTransaction().Rollback()

		err := DbDeleteWorkloadEventsOfTenantId(testCtx(1), "1")

		Expect(err).To(HaveOccurred())
	})
})
