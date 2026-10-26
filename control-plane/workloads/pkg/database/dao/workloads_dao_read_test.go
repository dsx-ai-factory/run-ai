/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package dao

import (
	"context"
	"errors"
	"net/http"
	"strings"

	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"
)

const (
	wlReadWorkloadSelect = "SELECT * FROM workloads_view WHERE tenant_id = ? AND id = ? ORDER BY workloads_view.id LIMIT ?"

	wlReadAdditionalFieldsSelect = "SELECT additional_fields FROM workloads WHERE tenant_id = ? AND id = ? ORDER BY workloads.id LIMIT ?"

	wlReadCountSelect            = "SELECT count(*) FROM workloads_view WHERE tenant_id = ? AND deleted_at IS NULL"
	wlReadCountFromTableSelect   = "SELECT count(*) FROM workloads WHERE tenant_id = ? AND deleted_at IS NULL"
	wlReadCountFromMatViewSelect = "SELECT count(*) FROM workloads_materialized_wrapper_view WHERE tenant_id = ? AND deleted_at IS NULL"
	wlReadCountFullSelect        = "SELECT count(*) FROM workloads_view WHERE tenant_id = ? AND ((false OR project_id IN ( 'p1'))) AND deleted_at IS NOT NULL AND name = ? AND (search_vector @@ to_tsquery('english', ? || ':*'))"

	wlReadRankingProjectsSelect = "SELECT project_id, COUNT(*) FILTER (WHERE phase = 'Running') as running_count, COUNT(*) FILTER (WHERE phase = 'Pending') as pending_count FROM workloads_view WHERE tenant_id = ? AND cluster_id = ? AND phase IN ('Running', 'Pending') AND deleted_at IS NULL AND project_id IN (?,?) GROUP BY project_id"
	wlReadRankingDeptsSelect    = "SELECT department_id, COUNT(*) FILTER (WHERE phase = 'Running' AND ?::text = ANY(current_nodepools)) as running_count, COUNT(*) FILTER (WHERE phase = 'Pending' AND cardinality(requested_nodepools) = 1 AND ?::text = ANY(requested_nodepools)) as pending_count FROM workloads_view WHERE tenant_id = ? AND cluster_id = ? AND phase IN ('Running', 'Pending') AND deleted_at IS NULL AND department_id IN (?) GROUP BY department_id"

	wlReadListSelect        = "SELECT * FROM workloads_view WHERE tenant_id = ? AND workloads_view.deleted_at IS NULL ORDER BY id DESC"
	wlReadListMatViewSelect = "SELECT * FROM workloads_materialized_wrapper_view WHERE tenant_id = ? AND workloads_materialized_wrapper_view.deleted_at IS NULL ORDER BY id DESC"
	wlReadListFullSelect    = "SELECT * FROM workloads_view WHERE tenant_id = ? AND ((false OR project_id IN ( 'p1'))) AND deleted_at IS NOT NULL AND name = ? AND (search_vector @@ to_tsquery('english', ? || ':*')) ORDER BY name DESC NULLS LAST,id DESC LIMIT ? OFFSET ?"
	wlReadListAscSelect     = "SELECT * FROM workloads_view WHERE tenant_id = ? AND workloads_view.deleted_at IS NULL ORDER BY name ASC NULLS FIRST,id DESC"

	wlReadTelemetryCountSelect       = "SELECT COUNT(id) as data FROM workloads_view WHERE tenant_id = ? AND deleted_at IS NULL"
	wlReadTelemetryCountGroupsSelect = "SELECT COUNT(id) as data, cluster_id as group1_main_value, cluster_name as group1_secondary_value, project_id as group2_main_value FROM workloads_view WHERE tenant_id = ? AND deleted_at IS NULL AND ((false OR project_id IN ( 'p1'))) AND cluster_id = ? AND ? = ANY(current_nodepools) AND department_id = ? AND project_id = ? AND phase IN (?) GROUP BY cluster_id,cluster_name,project_id"
	wlReadTelemetryCountPendingNpSel = "SELECT COUNT(id) as data FROM workloads_view WHERE tenant_id = ? AND deleted_at IS NULL AND array_length(requested_nodepools, 1) = 1 AND ? = ANY(requested_nodepools) AND phase IN (?)"
	wlReadTelemetryGpuSelect         = "SELECT COALESCE(SUM(allocated_gpu), 0) as data FROM workloads_view WHERE tenant_id = ? AND deleted_at IS NULL"
	wlReadTelemetryIdleGpuSelect     = "SELECT COALESCE(SUM(idle_allocated_gpus), 0) as data FROM workloads_view WHERE tenant_id = ? AND deleted_at IS NULL"

	wlReadPhaseSelect            = "SELECT * FROM workloads WHERE tenant_id = ? AND id = ? ORDER BY workloads.id LIMIT ?"
	wlReadPhaseWithClusterSelect = "SELECT * FROM workloads WHERE tenant_id = ? AND id = ? AND cluster_id = ? ORDER BY workloads.id LIMIT ?"

	wlReadChildrenSelect          = "SELECT * FROM workloads_children WHERE workload_id = ?"
	wlReadChildrenWithTypesSelect = "SELECT * FROM workloads_children WHERE workload_id = ? AND child_type IN (?)"

	wlReadClusterIdSelect = "SELECT cluster_id FROM workloads WHERE tenant_id = ? AND id = ? ORDER BY workloads.id LIMIT ?"
	wlReadScopeSelect     = "SELECT cluster_id, project_id, department_id, namespace FROM workloads WHERE tenant_id = ? AND id = ? AND workloads.deleted_at IS NULL ORDER BY workloads.id LIMIT ?"

	wlReadByIdsSelect               = "SELECT * FROM workloads_view WHERE tenant_id = ? AND id IN (?)"
	wlReadByIdsExcludeDeletedSelect = "SELECT * FROM workloads_view WHERE tenant_id = ? AND id IN (?) AND deleted_at IS NULL"
)

type wlReadViewRow struct {
	ID                   string
	SubmittedBy          string
	EnvironmentVariables []byte
}

type wlReadAdditionalFieldsRow struct {
	AdditionalFields []byte
}

type wlReadPhaseRow struct {
	ID    string
	Phase string
}

type wlReadClusterIdRow struct {
	ClusterID string `gorm:"column:cluster_id"`
}

type wlReadScopeRow struct {
	ClusterID    string `gorm:"column:cluster_id"`
	ProjectID    string `gorm:"column:project_id"`
	DepartmentID string `gorm:"column:department_id"`
	Namespace    string `gorm:"column:namespace"`
}

type wlReadChildRow struct {
	WorkloadID string `gorm:"column:workload_id"`
	ChildID    string `gorm:"column:child_id"`
	ChildType  string `gorm:"column:child_type"`
}

type wlReadRankingRow struct {
	ProjectID    string `gorm:"column:project_id"`
	RunningCount int64  `gorm:"column:running_count"`
	PendingCount int64  `gorm:"column:pending_count"`
}

type wlReadRankingDeptRow struct {
	DepartmentID string `gorm:"column:department_id"`
	RunningCount int64  `gorm:"column:running_count"`
	PendingCount int64  `gorm:"column:pending_count"`
}

type wlReadTelemetryRow struct {
	Data                 string `gorm:"column:data"`
	Group1MainValue      string `gorm:"column:group1_main_value"`
	Group1SecondaryValue string `gorm:"column:group1_secondary_value"`
	Group2MainValue      string `gorm:"column:group2_main_value"`
}

type wlReadSimpleTelemetryRow struct {
	Data string `gorm:"column:data"`
}

func wlReadDao() *WorkloadsDao {
	return &WorkloadsDao{view: WorkloadsViewName}
}

func wlReadMatViewDao() *WorkloadsDao {
	return &WorkloadsDao{view: WorkloadsApiViewName}
}

func wlReadScopes(system bool) *openapi.PermittedScopes {
	if system {
		return &openapi.PermittedScopes{System: &system}
	}
	return &openapi.PermittedScopes{System: &system, Projects: []string{"p1"}}
}

func wlReadFilters() []queryapi.Filter {
	return []queryapi.Filter{{Field: "name", Operator: "=", Values: []string{"a"}}}
}

func wlReadCtxWithSubject(tenantId uint, subject string) context.Context {
	return context.WithValue(testCtx(tenantId), runaictx.CfSubjectID, subject)
}

var _ = Describe("DbGetWorkload", func() {
	It("returns the workload with secret environment variables redacted", func() {
		m := setupDaoMock(GinkgoT())
		id := uuid.New()
		m.ExpectQuery(wlReadWorkloadSelect).Returns([]wlReadViewRow{{
			ID:                   id.String(),
			SubmittedBy:          "alice",
			EnvironmentVariables: []byte(`{"SAFE":"v","SECRET":"USERCREDENTIAL:abc"}`),
		}})

		workload, apiErr := wlReadDao().DbGetWorkload(testCtx(1), 1, id)

		Expect(apiErr).To(BeNil())
		Expect(workload).ToNot(BeNil())
		Expect(workload.ID).To(Equal(id))
		Expect(workload.EnvironmentVariables).To(Equal(models.EnvironmentVariables{"SAFE": "v"}))
	})

	It("returns an error when the workload is not found", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadWorkloadSelect).Returns([]wlReadViewRow{})

		workload, apiErr := wlReadDao().DbGetWorkload(testCtx(1), 1, uuid.New())

		Expect(workload).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadWorkloadSelect).ReturnsError(errors.New("boom"))

		workload, apiErr := wlReadDao().DbGetWorkload(testCtx(1), 1, uuid.New())

		Expect(workload).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbGetWorkloadAdditionalFields", func() {
	It("returns the additional fields", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadAdditionalFieldsSelect).Returns([]wlReadAdditionalFieldsRow{{AdditionalFields: []byte(`{"a":"b"}`)}})

		fields, apiErr := wlReadDao().DbGetWorkloadAdditionalFields(testCtx(1), 1, uuid.New())

		Expect(apiErr).To(BeNil())
		Expect(fields).To(Equal(models.AdditionalFields{"a": "b"}))
	})

	It("returns an error when the workload is not found", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadAdditionalFieldsSelect).Returns([]wlReadAdditionalFieldsRow{})

		fields, apiErr := wlReadDao().DbGetWorkloadAdditionalFields(testCtx(1), 1, uuid.New())

		Expect(fields).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadAdditionalFieldsSelect).ReturnsError(errors.New("boom"))

		fields, apiErr := wlReadDao().DbGetWorkloadAdditionalFields(testCtx(1), 1, uuid.New())

		Expect(fields).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("DbCountWorkloads", func() {
	It("counts workloads with default arguments", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadCountSelect).Returns([]countRow{{Count: 7}})

		count, apiErr := wlReadDao().DbCountWorkloads(testCtx(1), 1, nil, false, nil, "", false, false)

		Expect(apiErr).To(BeNil())
		Expect(count).To(Equal(int64(7)))
	})

	It("counts workloads for system scopes", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadCountSelect).Returns([]countRow{{Count: 1}})

		count, apiErr := wlReadDao().DbCountWorkloads(testCtx(1), 1, wlReadScopes(true), false, nil, "", false, false)

		Expect(apiErr).To(BeNil())
		Expect(count).To(Equal(int64(1)))
	})

	It("counts workloads with scopes, deleted, filters and search", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadCountFullSelect).Returns([]countRow{{Count: 3}})

		count, apiErr := wlReadDao().DbCountWorkloads(testCtx(1), 1, wlReadScopes(false), true, wlReadFilters(), "abc", false, false)

		Expect(apiErr).To(BeNil())
		Expect(count).To(Equal(int64(3)))
	})

	It("counts workloads from the base table", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadCountFromTableSelect).Returns([]countRow{{Count: 2}})

		count, apiErr := wlReadMatViewDao().DbCountWorkloads(testCtx(1), 1, nil, false, nil, "", true, true)

		Expect(apiErr).To(BeNil())
		Expect(count).To(Equal(int64(2)))
	})

	It("uses the plain view when a precise count is requested", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadCountSelect).Returns([]countRow{{Count: 4}})

		count, apiErr := wlReadMatViewDao().DbCountWorkloads(testCtx(1), 1, nil, false, nil, "", true, false)

		Expect(apiErr).To(BeNil())
		Expect(count).To(Equal(int64(4)))
	})

	It("counts workloads from the materialized view", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadCountFromMatViewSelect).Returns([]countRow{{Count: 5}})

		count, apiErr := wlReadMatViewDao().DbCountWorkloads(testCtx(1), 1, nil, false, nil, "", false, false)

		Expect(apiErr).To(BeNil())
		Expect(count).To(Equal(int64(5)))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadCountSelect).ReturnsError(errors.New("boom"))

		count, apiErr := wlReadDao().DbCountWorkloads(testCtx(1), 1, nil, false, nil, "", false, false)

		Expect(apiErr).ToNot(BeNil())
		Expect(count).To(Equal(int64(0)))
	})
})

var _ = Describe("DbCountWorkloadsForRanking", func() {
	It("returns no counts when there are no org units", func() {
		setupDaoMock(GinkgoT())

		counts, apiErr := wlReadDao().DbCountWorkloadsForRanking(testCtx(1), 1, "c1", nil, "project", nil)

		Expect(apiErr).To(BeNil())
		Expect(counts).To(BeEmpty())
	})

	It("returns running and pending counts per project", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadRankingProjectsSelect).Returns([]wlReadRankingRow{
			{ProjectID: "o1", RunningCount: 2, PendingCount: 1},
			{ProjectID: "o2", RunningCount: 0, PendingCount: 5},
		})

		counts, apiErr := wlReadDao().DbCountWorkloadsForRanking(testCtx(1), 1, "c1", []string{"o1", "o2"}, "project", nil)

		Expect(apiErr).To(BeNil())
		Expect(counts).To(HaveLen(2))
		Expect(counts["o1"].RunningCount).To(Equal(int64(2)))
		Expect(counts["o2"].PendingCount).To(Equal(int64(5)))
	})

	It("returns an error when the project query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadRankingProjectsSelect).ReturnsError(errors.New("boom"))

		counts, apiErr := wlReadDao().DbCountWorkloadsForRanking(testCtx(1), 1, "c1", []string{"o1", "o2"}, "project", nil)

		Expect(counts).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})

	It("returns counts per department filtered by nodepool", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadRankingDeptsSelect).Returns([]wlReadRankingDeptRow{{DepartmentID: "d1", RunningCount: 3, PendingCount: 4}})

		nodepool := "np1"
		counts, apiErr := wlReadDao().DbCountWorkloadsForRanking(testCtx(1), 1, "c1", []string{"d1"}, "department", &nodepool)

		Expect(apiErr).To(BeNil())
		Expect(counts).To(HaveLen(1))
		Expect(counts["d1"].RunningCount).To(Equal(int64(3)))
		Expect(counts["d1"].PendingCount).To(Equal(int64(4)))
	})

	It("returns an error when the department query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadRankingDeptsSelect).ReturnsError(errors.New("boom"))

		nodepool := "np1"
		counts, apiErr := wlReadDao().DbCountWorkloadsForRanking(testCtx(1), 1, "c1", []string{"d1"}, "department", &nodepool)

		Expect(counts).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})

	It("ignores an empty nodepool name", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadRankingProjectsSelect).Returns([]wlReadRankingRow{})

		nodepool := ""
		counts, apiErr := wlReadDao().DbCountWorkloadsForRanking(testCtx(1), 1, "c1", []string{"o1", "o2"}, "project", &nodepool)

		Expect(apiErr).To(BeNil())
		Expect(counts).To(BeEmpty())
	})
})

var _ = Describe("DbGetWorkloads", func() {
	It("returns workloads with secret environment variables redacted", func() {
		m := setupDaoMock(GinkgoT())
		id := uuid.New()
		m.ExpectQuery(wlReadListSelect).Returns([]wlReadViewRow{{
			ID:                   id.String(),
			SubmittedBy:          "alice",
			EnvironmentVariables: []byte(`{"SAFE":"v","SECRET":"USERCREDENTIAL:abc"}`),
		}})

		workloads, apiErr := wlReadDao().DbGetWorkloads(testCtx(1), 1, nil, false, 0, 0, "", "", nil, "")

		Expect(apiErr).To(BeNil())
		Expect(*workloads).To(HaveLen(1))
		Expect((*workloads)[0].EnvironmentVariables).To(Equal(models.EnvironmentVariables{"SAFE": "v"}))
	})

	It("reads from the materialized view", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadListMatViewSelect).Returns([]wlReadViewRow{})

		workloads, apiErr := wlReadMatViewDao().DbGetWorkloads(testCtx(1), 1, wlReadScopes(true), false, 0, 0, "", "", nil, "")

		Expect(apiErr).To(BeNil())
		Expect(*workloads).To(BeEmpty())
	})

	It("applies scopes, deleted, filters, sort, offset, limit and search", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadListFullSelect).Returns([]wlReadViewRow{})

		workloads, apiErr := wlReadDao().DbGetWorkloads(testCtx(1), 1, wlReadScopes(false), true, 5, 10, "DESC", "name", wlReadFilters(), "abc")

		Expect(apiErr).To(BeNil())
		Expect(*workloads).To(BeEmpty())
	})

	It("sorts in ascending order", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadListAscSelect).Returns([]wlReadViewRow{})

		workloads, apiErr := wlReadDao().DbGetWorkloads(testCtx(1), 1, nil, false, 0, 0, "ASC", "name", nil, "")

		Expect(apiErr).To(BeNil())
		Expect(*workloads).To(BeEmpty())
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadListSelect).ReturnsError(errors.New("boom"))

		workloads, apiErr := wlReadDao().DbGetWorkloads(testCtx(1), 1, nil, false, 0, 0, "", "", nil, "")

		Expect(apiErr).ToNot(BeNil())
		Expect(*workloads).To(BeEmpty())
	})
})

var _ = Describe("DbGetWorkloadCount", func() {
	It("returns the workload count", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadTelemetryCountSelect).Returns([]wlReadSimpleTelemetryRow{{Data: "9"}})

		values, apiErr := wlReadDao().DbGetWorkloadCount(testCtx(1), 1, "", "", "", "", nil, nil, nil, nil)

		Expect(apiErr).To(BeNil())
		Expect(*values).To(HaveLen(1))
		Expect((*values)[0].Data).To(Equal("9"))
	})

	It("returns the workload count for system scopes", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadTelemetryCountSelect).Returns([]wlReadSimpleTelemetryRow{})

		values, apiErr := wlReadDao().DbGetWorkloadCount(testCtx(1), 1, "", "", "", "", wlReadScopes(true), nil, nil, nil)

		Expect(apiErr).To(BeNil())
		Expect(*values).To(BeEmpty())
	})

	It("applies all filters and grouping", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadTelemetryCountGroupsSelect).Returns([]wlReadTelemetryRow{{
			Data:                 "4",
			Group1MainValue:      "c1",
			Group1SecondaryValue: "cluster-one",
			Group2MainValue:      "p1",
		}})

		values, apiErr := wlReadDao().DbGetWorkloadCount(testCtx(1), 1, "c1", "np1", "d1", "p1", wlReadScopes(false),
			[][]string{{"cluster_id", "cluster_name"}, {"project_id"}}, []api.Phase{api.PHASE_RUNNING}, nil)

		Expect(apiErr).To(BeNil())
		Expect(*values).To(HaveLen(1))
		Expect((*values)[0].Group1SecondaryValue).To(Equal("cluster-one"))
		Expect((*values)[0].Group2MainValue).To(Equal("p1"))
	})

	It("filters pending workloads by requested nodepool", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadTelemetryCountPendingNpSel).Returns([]wlReadSimpleTelemetryRow{{Data: "1"}})

		values, apiErr := wlReadDao().DbGetWorkloadCount(testCtx(1), 1, "", "np1", "", "", nil, nil, []api.Phase{api.PHASE_PENDING}, nil)

		Expect(apiErr).To(BeNil())
		Expect(*values).To(HaveLen(1))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadTelemetryCountSelect).ReturnsError(errors.New("boom"))

		values, apiErr := wlReadDao().DbGetWorkloadCount(testCtx(1), 1, "", "", "", "", nil, nil, nil, nil)

		Expect(apiErr).ToNot(BeNil())
		Expect(*values).To(BeEmpty())
	})
})

var _ = Describe("DbGetGpuAllocation", func() {
	It("returns the allocated gpu sum", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadTelemetryGpuSelect).Returns([]wlReadSimpleTelemetryRow{{Data: "2.5"}})

		values, apiErr := wlReadDao().DbGetGpuAllocation(testCtx(1), 1, "", "", "", "", nil, nil, nil, nil)

		Expect(apiErr).To(BeNil())
		Expect(*values).To(HaveLen(1))
		Expect((*values)[0].Data).To(Equal("2.5"))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadTelemetryGpuSelect).ReturnsError(errors.New("boom"))

		values, apiErr := wlReadDao().DbGetGpuAllocation(testCtx(1), 1, "", "", "", "", nil, nil, nil, nil)

		Expect(apiErr).ToNot(BeNil())
		Expect(*values).To(BeEmpty())
	})
})

var _ = Describe("DbGetIdleAllocatedGpus", func() {
	It("returns the idle allocated gpu sum", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadTelemetryIdleGpuSelect).Returns([]wlReadSimpleTelemetryRow{{Data: "0"}})

		values, apiErr := wlReadDao().DbGetIdleAllocatedGpus(testCtx(1), 1, "", "", "", "", nil, nil, nil, nil)

		Expect(apiErr).To(BeNil())
		Expect(*values).To(HaveLen(1))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadTelemetryIdleGpuSelect).ReturnsError(errors.New("boom"))

		values, apiErr := wlReadDao().DbGetIdleAllocatedGpus(testCtx(1), 1, "", "", "", "", nil, nil, nil, nil)

		Expect(apiErr).ToNot(BeNil())
		Expect(*values).To(BeEmpty())
	})
})

var _ = Describe("DbGetWorkloadPhase", func() {
	It("returns the workload phase", func() {
		m := setupDaoMock(GinkgoT())
		id := uuid.New()
		m.ExpectQuery(wlReadPhaseSelect).Returns([]wlReadPhaseRow{{ID: id.String(), Phase: string(api.PHASE_RUNNING)}})

		phase, apiErr := wlReadDao().DbGetWorkloadPhase(testCtx(1), 1, nil, id.String())

		Expect(apiErr).To(BeNil())
		Expect(phase).To(Equal(api.PHASE_RUNNING))
	})

	It("returns the workload phase scoped to a cluster", func() {
		m := setupDaoMock(GinkgoT())
		id := uuid.New()
		clusterId := "c1"
		m.ExpectQuery(wlReadPhaseWithClusterSelect).Returns([]wlReadPhaseRow{{ID: id.String(), Phase: string(api.PHASE_PENDING)}})

		phase, apiErr := wlReadDao().DbGetWorkloadPhase(testCtx(1), 1, &clusterId, id.String())

		Expect(apiErr).To(BeNil())
		Expect(phase).To(Equal(api.PHASE_PENDING))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadPhaseSelect).ReturnsError(errors.New("boom"))

		phase, apiErr := wlReadDao().DbGetWorkloadPhase(testCtx(1), 1, nil, uuid.New().String())

		Expect(apiErr).ToNot(BeNil())
		Expect(string(phase)).To(BeEmpty())
	})
})

var _ = Describe("DbGetChildren", func() {
	It("returns the children of a workload", func() {
		m := setupDaoMock(GinkgoT())
		workloadId := uuid.New()
		childId := uuid.New()
		m.ExpectQuery(wlReadChildrenSelect).Returns([]wlReadChildRow{{
			WorkloadID: workloadId.String(), ChildID: childId.String(), ChildType: "Pod",
		}})

		children, apiErr := wlReadDao().DbGetChildren(testCtx(1), workloadId, nil)

		Expect(apiErr).To(BeNil())
		Expect(*children).To(HaveLen(1))
		Expect((*children)[0].ChildType).To(Equal("Pod"))
	})

	It("filters children by child type", func() {
		m := setupDaoMock(GinkgoT())
		workloadId := uuid.New()
		childTypes := []string{"Pod"}
		m.ExpectQuery(wlReadChildrenWithTypesSelect).Returns([]wlReadChildRow{})

		children, apiErr := wlReadDao().DbGetChildren(testCtx(1), workloadId, &childTypes)

		Expect(apiErr).To(BeNil())
		Expect(*children).To(BeEmpty())
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadChildrenSelect).ReturnsError(errors.New("boom"))

		children, apiErr := wlReadDao().DbGetChildren(testCtx(1), uuid.New(), nil)

		Expect(apiErr).ToNot(BeNil())
		Expect(*children).To(BeEmpty())
	})
})

var _ = Describe("DbGetWorkloadClusterId", func() {
	It("returns the cluster id of the workload", func() {
		m := setupDaoMock(GinkgoT())
		clusterId := uuid.New()
		m.ExpectQuery(wlReadClusterIdSelect).Returns([]wlReadClusterIdRow{{ClusterID: clusterId.String()}})

		got, apiErr := wlReadDao().DbGetWorkloadClusterId(testCtx(1), 1, uuid.New().String())

		Expect(apiErr).To(BeNil())
		Expect(got).To(Equal(clusterId))
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadClusterIdSelect).ReturnsError(errors.New("boom"))

		got, apiErr := wlReadDao().DbGetWorkloadClusterId(testCtx(1), 1, uuid.New().String())

		Expect(apiErr).ToNot(BeNil())
		Expect(got).To(Equal(uuid.UUID{}))
	})
})

var _ = Describe("DbGetWorkloadScope", func() {
	It("returns the scope columns of the workload", func() {
		m := setupDaoMock(GinkgoT())
		clusterId := uuid.New()
		m.ExpectQuery(wlReadScopeSelect).Returns([]wlReadScopeRow{{
			ClusterID: clusterId.String(), ProjectID: "p1", DepartmentID: "d1", Namespace: "runai-p1",
		}})

		got, apiErr := wlReadDao().DbGetWorkloadScope(testCtx(1), 1, uuid.New())

		Expect(apiErr).To(BeNil())
		Expect(got.ClusterID).To(Equal(clusterId))
		Expect(got.ProjectID).To(Equal("p1"))
		Expect(got.DepartmentID).To(Equal("d1"))
		Expect(got.Namespace).To(Equal("runai-p1"))
	})

	It("returns not found when the workload does not exist", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadScopeSelect).ReturnsError(gorm.ErrRecordNotFound)

		got, apiErr := wlReadDao().DbGetWorkloadScope(testCtx(1), 1, uuid.New())

		Expect(got).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
		Expect(apiErr.Code).To(BeEquivalentTo(http.StatusNotFound))
	})

	It("returns an internal server error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadScopeSelect).ReturnsError(errors.New("boom"))

		got, apiErr := wlReadDao().DbGetWorkloadScope(testCtx(1), 1, uuid.New())

		Expect(got).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
		Expect(apiErr.Code).To(BeEquivalentTo(http.StatusInternalServerError))
	})
})

var _ = Describe("GetWorkloadsViewByIdsFromDb", func() {
	It("returns an error when the tenant id is missing from the context", func() {
		setupDaoMock(GinkgoT())

		workloads, apiErr := GetWorkloadsViewByIdsFromDb(context.Background(), []string{uuid.New().String()}, nil)

		Expect(workloads).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})

	It("returns nothing when no ids are given", func() {
		setupDaoMock(GinkgoT())

		workloads, apiErr := GetWorkloadsViewByIdsFromDb(testCtx(1), nil, nil)

		Expect(apiErr).To(BeNil())
		Expect(workloads).To(BeEmpty())
	})

	It("returns an error when an id is not a valid uuid", func() {
		setupDaoMock(GinkgoT())

		workloads, apiErr := GetWorkloadsViewByIdsFromDb(testCtx(1), []string{"not-a-uuid"}, nil)

		Expect(workloads).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when too many ids are given", func() {
		setupDaoMock(GinkgoT())

		ids := make([]string, 101)
		for i := range ids {
			ids[i] = uuid.New().String()
		}

		workloads, apiErr := GetWorkloadsViewByIdsFromDb(testCtx(1), ids, nil)

		Expect(workloads).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})

	It("returns the workloads with secret environment variables redacted", func() {
		m := setupDaoMock(GinkgoT())
		id := uuid.New()
		m.ExpectQuery(wlReadByIdsSelect).Returns([]wlReadViewRow{{
			ID:                   id.String(),
			SubmittedBy:          "alice",
			EnvironmentVariables: []byte(`{"SAFE":"v","SECRET":"USERCREDENTIAL:abc"}`),
		}})

		workloads, apiErr := GetWorkloadsViewByIdsFromDb(testCtx(1), []string{id.String()}, nil)

		Expect(apiErr).To(BeNil())
		Expect(workloads).To(HaveLen(1))
		Expect(workloads[0].EnvironmentVariables).To(Equal(models.EnvironmentVariables{"SAFE": "v"}))
	})

	It("excludes deleted workloads when asked to", func() {
		m := setupDaoMock(GinkgoT())
		id := uuid.New()
		excludeDeleted := true
		m.ExpectQuery(wlReadByIdsExcludeDeletedSelect).Returns([]wlReadViewRow{})

		workloads, apiErr := GetWorkloadsViewByIdsFromDb(testCtx(1), []string{id.String()}, &excludeDeleted)

		Expect(apiErr).To(BeNil())
		Expect(workloads).To(BeEmpty())
	})

	It("returns an error when the query fails", func() {
		m := setupDaoMock(GinkgoT())
		m.ExpectQuery(wlReadByIdsSelect).ReturnsError(errors.New("boom"))

		workloads, apiErr := GetWorkloadsViewByIdsFromDb(testCtx(1), []string{uuid.New().String()}, nil)

		Expect(workloads).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})
})

var _ = Describe("GetWorkloadViewFromDb", func() {
	It("returns an error when the tenant id is missing from the context", func() {
		setupDaoMock(GinkgoT())

		workload, apiErr := GetWorkloadViewFromDb(context.Background(), uuid.New().String())

		Expect(workload).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})

	It("returns an error when the id is not a valid uuid", func() {
		setupDaoMock(GinkgoT())

		workload, apiErr := GetWorkloadViewFromDb(testCtx(1), "not-a-uuid")

		Expect(workload).To(BeNil())
		Expect(apiErr).ToNot(BeNil())
	})

	It("returns the workload", func() {
		m := setupDaoMock(GinkgoT())
		id := uuid.New()
		m.ExpectQuery(wlReadWorkloadSelect).Returns([]wlReadViewRow{{ID: id.String(), SubmittedBy: "alice", EnvironmentVariables: []byte(`{}`)}})

		workload, apiErr := GetWorkloadViewFromDb(testCtx(1), id.String())

		Expect(apiErr).To(BeNil())
		Expect(workload).ToNot(BeNil())
		Expect(workload.ID).To(Equal(id))
	})
})

var _ = Describe("getPermittedScopesFilter", func() {
	It("builds a filter for every scope kind when not a system scope", func() {
		system := false
		filter := getPermittedScopesFilter(openapi.PermittedScopes{
			System:      &system,
			Tenants:     []string{"1"},
			Clusters:    []string{"c1"},
			Departments: []string{"d1"},
			Projects:    []string{"p1"},
		})

		Expect(strings.Contains(filter, "tenant_id IN")).To(BeTrue())
		Expect(strings.Contains(filter, "cluster_id IN")).To(BeTrue())
		Expect(strings.Contains(filter, "department_id IN")).To(BeTrue())
		Expect(strings.Contains(filter, "project_id IN")).To(BeTrue())
	})

	It("returns true for a system scope", func() {
		system := true
		Expect(getPermittedScopesFilter(openapi.PermittedScopes{System: &system})).To(Equal("true"))
	})
})

var _ = Describe("redactWorkloadEnvironmentVariables", func() {
	It("does not panic on a nil workload", func() {
		Expect(func() { redactWorkloadEnvironmentVariables(testCtx(1), nil) }).ToNot(Panic())
	})

	It("leaves nil environment variables untouched", func() {
		workload := &models.WorkloadView{}

		redactWorkloadEnvironmentVariables(testCtx(1), workload)

		Expect(workload.EnvironmentVariables).To(BeNil())
	})

	It("keeps secrets for the submitter", func() {
		workload := &models.WorkloadView{
			SubmittedBy:          "alice",
			EnvironmentVariables: models.EnvironmentVariables{"SECRET": "USERCREDENTIAL:abc"},
		}

		redactWorkloadEnvironmentVariables(wlReadCtxWithSubject(1, "alice"), workload)

		Expect(workload.EnvironmentVariables).To(Equal(models.EnvironmentVariables{"SECRET": "USERCREDENTIAL:abc"}))
	})

	It("removes secrets for another user", func() {
		workload := &models.WorkloadView{
			SubmittedBy:          "alice",
			EnvironmentVariables: models.EnvironmentVariables{"SECRET": "USERCREDENTIAL:abc", "SAFE": "v"},
		}

		redactWorkloadEnvironmentVariables(wlReadCtxWithSubject(1, "bob"), workload)

		Expect(workload.EnvironmentVariables).To(Equal(models.EnvironmentVariables{"SAFE": "v"}))
	})
})

var _ = Describe("redactWorkloadsEnvironmentVariables", func() {
	It("does not panic on a nil slice", func() {
		Expect(func() { redactWorkloadsEnvironmentVariables(testCtx(1), nil) }).ToNot(Panic())
	})
})

var _ = Describe("redactWorkloadsEnvironmentVariablesPointers", func() {
	It("does not panic on a nil slice", func() {
		Expect(func() { redactWorkloadsEnvironmentVariablesPointers(testCtx(1), nil) }).ToNot(Panic())
	})
})
