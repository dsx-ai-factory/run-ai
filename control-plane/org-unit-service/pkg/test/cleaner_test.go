package testing

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
)

var _ = Describe("DB Cleaner", Ordered, func() {
	tenantId := "cleaner-test-tenant"
	clusterId := uuid.New()
	daysToLive := 30
	old := time.Now().AddDate(0, 0, -60)
	recent := time.Now().AddDate(0, 0, -5)

	var (
		createdOrgUnitDataIds []uuid.UUID
		createdNodeTypeIds    []int64

		oldOrgUnitDataId      uuid.UUID
		recentOrgUnitDataId   uuid.UUID
		liveOrgUnitDataId     uuid.UUID
		orphanOrgUnitDataId   uuid.UUID
		referencedOrgUnitData uuid.UUID

		oldResourceId    uuid.UUID
		recentResourceId uuid.UUID

		oldProjectId    int64
		recentProjectId int64
		liveProjectId   int64

		oldDeptId    int64
		recentDeptId int64
		liveDeptId   int64

		oldOvertimeId    uuid.UUID
		recentOvertimeId uuid.UUID
		liveOvertimeId   uuid.UUID

		nodeTypeId int64
	)

	insertOrgUnitData := func(deletedAt *time.Time) uuid.UUID {
		id := uuid.New()
		var deletedExpr interface{} = nil
		if deletedAt != nil {
			deletedExpr = *deletedAt
		}
		Expect(dao_utils.DB().Exec(
			"INSERT INTO org_unit_data (id, created_at, deleted_at) VALUES (?, NOW(), ?)",
			id, deletedExpr,
		).Error).To(BeNil())
		createdOrgUnitDataIds = append(createdOrgUnitDataIds, id)
		return id
	}

	insertResource := func(orgUnitDataId uuid.UUID, deletedAt *time.Time, nodePoolName string) uuid.UUID {
		id := uuid.New()
		var deletedExpr interface{} = nil
		if deletedAt != nil {
			deletedExpr = *deletedAt
		}
		Expect(dao_utils.DB().Exec(
			"INSERT INTO resources (id, created_at, deleted_at, node_pool_id, node_pool_name, org_unit_data_id) VALUES (?, NOW(), ?, ?, ?, ?)",
			id, deletedExpr, nodePoolName+"-id", nodePoolName, orgUnitDataId,
		).Error).To(BeNil())
		return id
	}

	insertProject := func(orgUnitDataId uuid.UUID, deletedAt *time.Time, name string) int64 {
		var deletedExpr interface{} = nil
		if deletedAt != nil {
			deletedExpr = *deletedAt
		}
		var id int64
		Expect(dao_utils.DB().Raw(
			"INSERT INTO project (created_at, deleted_at, tenant_id, cluster_id, name, org_unit_data_id) "+
				"VALUES (NOW(), ?, ?, ?, ?, ?) RETURNING id",
			deletedExpr, tenantId, clusterId, name, orgUnitDataId,
		).Scan(&id).Error).To(BeNil())
		return id
	}

	insertDepartment := func(orgUnitDataId uuid.UUID, deletedAt *time.Time, name string) int64 {
		var deletedExpr interface{} = nil
		if deletedAt != nil {
			deletedExpr = *deletedAt
		}
		var id int64
		Expect(dao_utils.DB().Raw(
			"INSERT INTO department (created_at, deleted_at, tenant_id, cluster_id, name, org_unit_data_id) "+
				"VALUES (NOW(), ?, ?, ?, ?, ?) RETURNING id",
			deletedExpr, tenantId, clusterId, name, orgUnitDataId,
		).Scan(&id).Error).To(BeNil())
		return id
	}

	insertOvertimeData := func(orgUnitDataId uuid.UUID) uuid.UUID {
		id := uuid.New()
		Expect(dao_utils.DB().Exec(
			"INSERT INTO overtime_data (id, org_unit_data_id) VALUES (?, ?)",
			id, orgUnitDataId,
		).Error).To(BeNil())
		return id
	}

	insertNodeType := func(name string) int64 {
		var id int64
		Expect(dao_utils.DB().Raw(
			"INSERT INTO node_type (created_at, tenant_id, cluster_id, name) "+
				"VALUES (NOW(), ?, ?, ?) RETURNING id",
			tenantId, clusterId, name,
		).Scan(&id).Error).To(BeNil())
		createdNodeTypeIds = append(createdNodeTypeIds, id)
		return id
	}

	insertJoinRow := func(table string, orgUnitDataId uuid.UUID, ntId int64) {
		Expect(dao_utils.DB().Exec(
			fmt.Sprintf("INSERT INTO %s (org_unit_data_id, node_type_id) VALUES (?, ?)", table),
			orgUnitDataId, ntId,
		).Error).To(BeNil())
	}

	countWhere := func(table, where string, args ...interface{}) int64 {
		var count int64
		query := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", table, where)
		Expect(dao_utils.DB().Raw(query, args...).Scan(&count).Error).To(BeNil())
		return count
	}

	BeforeAll(func() {
		oldOrgUnitDataId = insertOrgUnitData(&old)
		recentOrgUnitDataId = insertOrgUnitData(&recent)
		liveOrgUnitDataId = insertOrgUnitData(nil)
		orphanOrgUnitDataId = insertOrgUnitData(&old)
		referencedOrgUnitData = insertOrgUnitData(&old)

		oldResourceId = insertResource(oldOrgUnitDataId, &old, "np-old")
		recentResourceId = insertResource(recentOrgUnitDataId, &recent, "np-recent")
		_ = insertResource(liveOrgUnitDataId, nil, "np-live")

		oldProjectId = insertProject(oldOrgUnitDataId, &old, "old-project")
		recentProjectId = insertProject(recentOrgUnitDataId, &recent, "recent-project")
		liveProjectId = insertProject(liveOrgUnitDataId, nil, "live-project")
		// Live project pinned to a soft-deleted org_unit_data — should protect that OUD from cleanup.
		_ = insertProject(referencedOrgUnitData, nil, "live-project-pointing-at-deleted-oud")

		oldDeptId = insertDepartment(insertOrgUnitData(&old), &old, "old-dept")
		recentDeptId = insertDepartment(insertOrgUnitData(&recent), &recent, "recent-dept")
		liveDeptId = insertDepartment(insertOrgUnitData(nil), nil, "live-dept")

		// overtime_data + node-type joins on each org_unit_data state.
		oldOvertimeId = insertOvertimeData(oldOrgUnitDataId)
		recentOvertimeId = insertOvertimeData(recentOrgUnitDataId)
		liveOvertimeId = insertOvertimeData(liveOrgUnitDataId)

		nodeTypeId = insertNodeType("cleaner-test-node-type")
		insertJoinRow("org_unit_train_node_type", oldOrgUnitDataId, nodeTypeId)
		insertJoinRow("org_unit_train_node_type", recentOrgUnitDataId, nodeTypeId)
		insertJoinRow("org_unit_train_node_type", liveOrgUnitDataId, nodeTypeId)
		insertJoinRow("org_unit_workspace_node_type", oldOrgUnitDataId, nodeTypeId)
		insertJoinRow("org_unit_workspace_node_type", recentOrgUnitDataId, nodeTypeId)
		insertJoinRow("org_unit_workspace_node_type", liveOrgUnitDataId, nodeTypeId)
	})

	AfterAll(func() {
		Expect(dao_utils.DB().Exec("DELETE FROM project    WHERE tenant_id = ?", tenantId).Error).To(BeNil())
		Expect(dao_utils.DB().Exec("DELETE FROM department WHERE tenant_id = ?", tenantId).Error).To(BeNil())
		if len(createdOrgUnitDataIds) > 0 {
			Expect(dao_utils.DB().Exec(
				"DELETE FROM org_unit_train_node_type     WHERE org_unit_data_id IN ?", createdOrgUnitDataIds,
			).Error).To(BeNil())
			Expect(dao_utils.DB().Exec(
				"DELETE FROM org_unit_workspace_node_type WHERE org_unit_data_id IN ?", createdOrgUnitDataIds,
			).Error).To(BeNil())
			Expect(dao_utils.DB().Exec(
				"DELETE FROM overtime_data WHERE org_unit_data_id IN ?", createdOrgUnitDataIds,
			).Error).To(BeNil())
			Expect(dao_utils.DB().Exec(
				"DELETE FROM resources     WHERE org_unit_data_id IN ?", createdOrgUnitDataIds,
			).Error).To(BeNil())
			Expect(dao_utils.DB().Exec(
				"DELETE FROM org_unit_data WHERE id IN ?", createdOrgUnitDataIds,
			).Error).To(BeNil())
		}
		if len(createdNodeTypeIds) > 0 {
			Expect(dao_utils.DB().Exec(
				"DELETE FROM node_type WHERE id IN ?", createdNodeTypeIds,
			).Error).To(BeNil())
		}
	})

	It("hard-deletes only old soft-deleted rows and parent-driven dependents", func() {
		for _, q := range database.BuildCleanerQueries(daysToLive) {
			Expect(dao_utils.DB().Exec(q).Error).To(BeNil())
		}

		// Old soft-deleted: gone
		Expect(countWhere("project", "id = ?", oldProjectId)).To(Equal(int64(0)))
		Expect(countWhere("resources", "id = ?", oldResourceId)).To(Equal(int64(0)))
		Expect(countWhere("department", "id = ?", oldDeptId)).To(Equal(int64(0)))

		// Recent soft-deleted: still there
		Expect(countWhere("project", "id = ?", recentProjectId)).To(Equal(int64(1)))
		Expect(countWhere("resources", "id = ?", recentResourceId)).To(Equal(int64(1)))
		Expect(countWhere("department", "id = ?", recentDeptId)).To(Equal(int64(1)))

		// Live: still there
		Expect(countWhere("project", "id = ?", liveProjectId)).To(Equal(int64(1)))
		Expect(countWhere("department", "id = ?", liveDeptId)).To(Equal(int64(1)))

		// overtime_data tied to old org_unit_data: gone. Recent/live: preserved.
		Expect(countWhere("overtime_data", "id = ?", oldOvertimeId)).To(Equal(int64(0)))
		Expect(countWhere("overtime_data", "id = ?", recentOvertimeId)).To(Equal(int64(1)))
		Expect(countWhere("overtime_data", "id = ?", liveOvertimeId)).To(Equal(int64(1)))

		// Join tables tied to old org_unit_data: gone. Recent/live: preserved.
		Expect(countWhere("org_unit_train_node_type", "org_unit_data_id = ?", oldOrgUnitDataId)).To(Equal(int64(0)))
		Expect(countWhere("org_unit_train_node_type", "org_unit_data_id = ?", recentOrgUnitDataId)).To(Equal(int64(1)))
		Expect(countWhere("org_unit_train_node_type", "org_unit_data_id = ?", liveOrgUnitDataId)).To(Equal(int64(1)))
		Expect(countWhere("org_unit_workspace_node_type", "org_unit_data_id = ?", oldOrgUnitDataId)).To(Equal(int64(0)))
		Expect(countWhere("org_unit_workspace_node_type", "org_unit_data_id = ?", recentOrgUnitDataId)).To(Equal(int64(1)))
		Expect(countWhere("org_unit_workspace_node_type", "org_unit_data_id = ?", liveOrgUnitDataId)).To(Equal(int64(1)))

		// Orphan org_unit_data (soft-deleted, no live refs): gone
		Expect(countWhere("org_unit_data", "id = ?", orphanOrgUnitDataId)).To(Equal(int64(0)))

		// Old org_unit_data with all children cleaned: gone
		Expect(countWhere("org_unit_data", "id = ?", oldOrgUnitDataId)).To(Equal(int64(0)))

		// Soft-deleted org_unit_data still pinned by a live project: preserved
		Expect(countWhere("org_unit_data", "id = ?", referencedOrgUnitData)).To(Equal(int64(1)))
	})
})
