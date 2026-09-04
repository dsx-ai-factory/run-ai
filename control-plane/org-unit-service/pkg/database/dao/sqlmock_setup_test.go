package dao

import (
	"testing"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/dbmodels"
	dbtesting "github.com/dsx-ai-factory/run-ai/runai-common-packages/runaidb/testing"
)

const (
	testTenantID     = "tenant-1"
	testDepartmentID = int64(11)
	testProjectID    = int64(22)
	testNodeTypeID   = int64(33)
	testNodePoolID   = "np-1"
	testNodePoolName = "default"
	testNamespace    = "runai-proj"
)

type countResult struct {
	Count int64
}

type clusterIDRow struct {
	ClusterId string `gorm:"column:cluster_id"`
}

type nameRow struct {
	Name string
}

type sumRow struct {
	Sum float64
}

// Slim row types omit GORM associations (slices/structs) that sqlmock cannot serialize.
type departmentRow struct {
	ID            int64
	TenantId      string
	ClusterId     uuid.UUID
	Name          string
	ParentID      *int64
	OrgUnitDataID uuid.UUID
}

type projectRow struct {
	ID            int64
	TenantId      string
	ClusterId     uuid.UUID
	Name          string
	ParentID      *int64
	ParentName    *string
	OrgUnitDataID uuid.UUID
	QuotaStatusID uuid.UUID
}

func departmentRows(ds ...models.Department) []departmentRow {
	rows := make([]departmentRow, 0, len(ds))
	for _, d := range ds {
		rows = append(rows, departmentRow{
			ID: d.ID, TenantId: d.TenantId, ClusterId: d.ClusterId,
			Name: d.Name, ParentID: d.ParentID, OrgUnitDataID: d.OrgUnitDataID,
		})
	}
	return rows
}

func projectRows(ps ...models.Project) []projectRow {
	rows := make([]projectRow, 0, len(ps))
	for _, p := range ps {
		rows = append(rows, projectRow{
			ID: p.ID, TenantId: p.TenantId, ClusterId: p.ClusterId,
			Name: p.Name, ParentID: p.ParentID, ParentName: p.ParentName,
			OrgUnitDataID: p.OrgUnitDataID, QuotaStatusID: p.QuotaStatusID,
		})
	}
	return rows
}

func projectViewRows(ps ...models.ProjectView) []projectRow {
	rows := make([]projectRow, 0, len(ps))
	for _, p := range ps {
		rows = append(rows, projectRow{
			ID: p.ID, TenantId: p.TenantId, ClusterId: p.ClusterId,
			Name: p.Name, ParentID: p.ParentID, ParentName: p.ParentName,
			OrgUnitDataID: p.OrgUnitDataID, QuotaStatusID: p.QuotaStatusID,
		})
	}
	return rows
}

func TestDAOSQL(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Org unit DAO SQL Tests")
}

func setupSQLMock() dbtesting.Mocker {
	return dbtesting.SetSQLMock(GinkgoT())
}

func testClusterID() uuid.UUID {
	return uuid.MustParse("11111111-1111-1111-1111-111111111111")
}

func testOrgUnitDataID() uuid.UUID {
	return uuid.MustParse("22222222-2222-2222-2222-222222222222")
}

func testResourcesID() uuid.UUID {
	return uuid.MustParse("33333333-3333-3333-3333-333333333333")
}

func testQuotaStatusID() uuid.UUID {
	return uuid.MustParse("44444444-4444-4444-4444-444444444444")
}

func testOvertimeDataID() uuid.UUID {
	return uuid.MustParse("55555555-5555-5555-5555-555555555555")
}

func testDepartment() models.Department {
	return models.Department{
		IntBaseModel:  models.IntBaseModel{ID: testDepartmentID},
		TenantId:      testTenantID,
		ClusterId:     testClusterID(),
		Name:          "research",
		OrgUnitDataID: testOrgUnitDataID(),
	}
}

func testDepartmentView() models.DepartmentView {
	return models.DepartmentView{
		IntBaseModel:  models.IntBaseModel{ID: testDepartmentID},
		TenantId:      testTenantID,
		ClusterId:     testClusterID(),
		Name:          "research",
		OrgUnitDataID: testOrgUnitDataID(),
	}
}

func testProject() models.Project {
	parentID := testDepartmentID
	return models.Project{
		IntBaseModel:  models.IntBaseModel{ID: testProjectID},
		TenantId:      testTenantID,
		ClusterId:     testClusterID(),
		Name:          "proj-a",
		ParentID:      &parentID,
		OrgUnitDataID: testOrgUnitDataID(),
		ProjectStatus: models.ProjectStatus{QuotaStatusID: testQuotaStatusID()},
	}
}

func testProjectView() models.ProjectView {
	parentID := testDepartmentID
	parentName := "research"
	return models.ProjectView{
		IntBaseModel:  models.IntBaseModel{ID: testProjectID},
		TenantId:      testTenantID,
		ClusterId:     testClusterID(),
		Name:          "proj-a",
		ParentID:      &parentID,
		ParentName:    &parentName,
		OrgUnitDataID: testOrgUnitDataID(),
	}
}

func testNodeType() models.NodeType {
	return models.NodeType{
		IntBaseModel: models.IntBaseModel{ID: testNodeTypeID},
		Name:         "h100",
		ClusterId:    testClusterID(),
		TenantId:     testTenantID,
	}
}

func testResources() models.Resources {
	return models.Resources{
		BaseModel:     dbmodels.BaseModel{ID: testResourcesID()},
		NodePoolId:    testNodePoolID,
		NodePoolName:  testNodePoolName,
		OrgUnitDataID: testOrgUnitDataID(),
	}
}

func testDepartmentNodePoolResources() models.DepartmentNodePoolResources {
	return models.DepartmentNodePoolResources{
		DeleteBaseModel: dbmodels.DeleteBaseModel{ID: testResourcesID()},
		TenantId:        testTenantID,
		ClusterId:       testClusterID(),
		OrgUnitId:       "11",
		OrgUnitName:     "research",
		NodePoolId:      testNodePoolID,
		NodePoolName:    testNodePoolName,
		Rank:            models.DefaultResourceRank,
	}
}

func testProjectNodePoolResources() models.ProjectNodePoolResources {
	return models.ProjectNodePoolResources{
		DeleteBaseModel: dbmodels.DeleteBaseModel{ID: testResourcesID()},
		TenantId:        testTenantID,
		ClusterId:       testClusterID(),
		OrgUnitId:       "22",
		OrgUnitName:     "proj-a",
		ParentId:        "11",
		NodePoolId:      testNodePoolID,
		NodePoolName:    testNodePoolName,
		Rank:            models.DefaultResourceRank,
	}
}
