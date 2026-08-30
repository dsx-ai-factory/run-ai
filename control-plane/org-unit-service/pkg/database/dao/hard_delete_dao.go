package dao

import (
	"context"
	"maps"
	"reflect"
	"slices"

	"github.com/run-ai/runai/runai-common-packages/dao-utils/validators"

	"github.com/google/uuid"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/models"
	daoutils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"gorm.io/gorm"
)

const (
	DeletionErrorLogPrefix = "error on tenant deletion."
	ServiceName            = "org-unit-service"
)

func logError(ctx context.Context, tenantId string, assetName string, err error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s tenantId: <%s>, service: <%s>, asset: <%s>, error: %s", DeletionErrorLogPrefix, tenantId, ServiceName, assetName, err)
}

func GetAllTenantProjects(ctx context.Context, tenantId string) ([]*models.Project, error) {
	var projects []*models.Project
	projectsList, err := GetProjects(ctx, tenantId, []validators.ParameterizedFilter{}, "", "", 0, 0, "", nil, true, models.SlimVerbosity)
	if err != nil {
		return nil, err
	}
	for _, project := range *projectsList {
		projects = append(projects, &project)
	}
	return projects, nil
}

func GetAllTenantDepartments(ctx context.Context, tenantId string) ([]*models.Department, error) {
	var departments []*models.Department
	departmentsList, err := GetDepartments(ctx, tenantId, []validators.ParameterizedFilter{}, "", "", 0, 0, nil, true, models.SlimVerbosity)
	if err != nil {
		return nil, err
	}
	for _, department := range *departmentsList {
		departments = append(departments, &department)
	}

	return departments, nil
}

func GetAllTenantNodeTypes(ctx context.Context, tenantId string) ([]*models.NodeType, error) {
	var nodeTypes []*models.NodeType
	nodeTypesList := &[]models.NodeType{}
	db := daoutils.DB().WithContext(ctx).Table(models.NodeTypeTable).Unscoped()
	result := db.Where(&models.NodeType{
		TenantId: tenantId,
	}).Find(nodeTypesList)
	if result.Error != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed get node type: %v", result.Error)
		return nil, result.Error
	}
	for _, nodeType := range *nodeTypesList {
		nodeTypes = append(nodeTypes, &nodeType)
	}

	return nodeTypes, nil
}

func getProjectsClusterIds(projects []*models.Project) []string {
	clusterIds := map[string]bool{}
	for _, project := range projects {
		clusterIds[project.ClusterId.String()] = true
	}
	return slices.Collect(maps.Keys(clusterIds))
}

func getDepartmentsClusterIds(departments []*models.Department) []string {
	clusterIds := map[string]bool{}
	for _, department := range departments {
		clusterIds[department.ClusterId.String()] = true
	}
	return slices.Collect(maps.Keys(clusterIds))
}

func invalidateCaches(ctx context.Context, tenantId string, clusterIds []string) {
	for _, clusterId := range clusterIds {
		InvalidateAllCache(ctx, tenantId, clusterId)
	}
}

func HardDeleteTenantNodeTypes(ctx context.Context, tenantId string) error {
	nodeTypes, err := GetAllTenantNodeTypes(ctx, tenantId)
	if err != nil {
		logError(ctx, tenantId, "NodeType", err)
		return err
	}

	if len(nodeTypes) == 0 {
		return nil
	}
	tx := daoutils.DB().WithContext(ctx).Begin()
	committed := false
	defer daoutils.FinalizeTransaction(tx, &committed)

	var nodeTypesIds []int64
	for _, nodeType := range nodeTypes {
		nodeTypesIds = append(nodeTypesIds, nodeType.ID)
	}

	err = hardDeleteTenantNodeTypeWorkload[models.OrgUnitTrainNodeType](ctx, tx, models.OrgUnitNodeTypeTrainTable, tenantId, nodeTypesIds)
	if err != nil {
		return err
	}
	err = hardDeleteTenantNodeTypeWorkload[models.OrgUnitWorkspaceNodeType](ctx, tx, models.OrgUnitNodeTypeWorkspaceTable, tenantId, nodeTypesIds)
	if err != nil {
		return err
	}
	result := tx.Table(models.NodeTypeTable).Where(nodeTypesIds).Unscoped().Delete(&[]models.NodeType{})
	if result.Error != nil {
		logError(ctx, tenantId, "NodeType", result.Error)
		return result.Error
	}
	result = tx.Commit()
	if result.Error != nil {
		logError(ctx, tenantId, "NodeType", result.Error)
		return result.Error
	}
	committed = true
	return nil
}

func hardDeleteTenantNodeTypeWorkload[T any](ctx context.Context, tx *gorm.DB, tableName string, tenantId string, nodeTypesIds []int64) error {
	if len(nodeTypesIds) == 0 {
		return nil
	}
	modelType := reflect.TypeFor[T]()
	assetName := modelType.String()
	result := tx.Table(tableName).Where("node_type_id IN ?", nodeTypesIds).Unscoped().Delete(&[]T{})
	if result.Error != nil {
		logError(ctx, tenantId, assetName, result.Error)
		return result.Error
	}
	return nil
}

func HardDeleteTenantProjects(ctx context.Context, tenantId string, projects []*models.Project) error {
	if len(projects) == 0 {
		return nil
	}

	clusterIds := getProjectsClusterIds(projects)

	tx := daoutils.DB().WithContext(ctx).Begin()
	committed := false
	defer daoutils.FinalizeTransaction(tx, &committed)

	var resourcesIds []uuid.UUID
	var orgUnitsIds []uuid.UUID
	var projectsIds []int64
	for _, project := range projects {
		projectsIds = append(projectsIds, project.ID)
		orgUnitsIds = append(orgUnitsIds, project.OrgUnitData.ID)
		for _, projResources := range project.OrgUnitData.Resources {
			resourcesIds = append(resourcesIds, projResources.ID)
		}
	}

	var result *gorm.DB

	if len(projectsIds) > 0 {
		result = tx.Table(models.ProjectTable).Where(projectsIds).Unscoped().Delete(&[]models.Project{})
		if result.Error != nil {
			logError(ctx, tenantId, "Projects", result.Error)
			return result.Error
		}
	}

	if len(resourcesIds) > 0 {
		result = tx.Table(models.ResourcesTable).Where(resourcesIds).Unscoped().Delete(&[]models.Resources{})
		if result.Error != nil {
			logError(ctx, tenantId, "Resources", result.Error)
			return result.Error
		}
	}

	for _, orgUnitsId := range orgUnitsIds {
		result = tx.Table(models.OrgUnitDataTable).Where("id = ?", orgUnitsId).Unscoped().Delete(&[]models.OrgUnitData{})
		if result.Error != nil {
			logError(ctx, tenantId, "OrgUnitData", result.Error)
			return result.Error
		}
	}

	result = tx.Commit()
	if result.Error != nil {
		logError(ctx, tenantId, "Projects", result.Error)
		return result.Error
	}
	committed = true
	invalidateCaches(ctx, tenantId, clusterIds)

	return nil
}

func HardDeleteTenantDepartments(ctx context.Context, tenantId string, departments []*models.Department) error {
	if len(departments) == 0 {
		return nil
	}
	clusterIds := getDepartmentsClusterIds(departments)
	tx := daoutils.DB().WithContext(ctx).Begin()
	committed := false
	defer daoutils.FinalizeTransaction(tx, &committed)

	var resourcesIds []uuid.UUID
	var orgUnitsIds []uuid.UUID
	var departmentsIds []int64
	for _, department := range departments {
		departmentsIds = append(departmentsIds, department.ID)
		orgUnitsIds = append(orgUnitsIds, department.OrgUnitData.ID)
		for _, deptResources := range department.OrgUnitData.Resources {
			resourcesIds = append(resourcesIds, deptResources.ID)
		}
	}
	result := tx.Table(models.ResourcesTable).Where(resourcesIds).Unscoped().Delete(&models.Resources{})
	if result.Error != nil {
		logError(ctx, tenantId, "Resources", result.Error)
		return result.Error
	}

	result = tx.Where(departmentsIds).Unscoped().Delete(&models.Department{})
	if result.Error != nil {
		logError(ctx, tenantId, "Departments", result.Error)
		return result.Error
	}

	result = tx.Where(orgUnitsIds).Unscoped().Delete(&models.OrgUnitData{})
	if result.Error != nil {
		logError(ctx, tenantId, "OrgUnitData", result.Error)
		return result.Error
	}

	result = tx.Commit()
	if result.Error != nil {
		logError(ctx, tenantId, "Departments", result.Error)
		return result.Error
	}
	committed = true
	invalidateCaches(ctx, tenantId, clusterIds)

	return nil
}

func HardDeleteOvertimeData(ctx context.Context, tenantId string, projects []*models.Project, departments []*models.Department) error {
	var orgUnitsIds []uuid.UUID
	for _, department := range departments {
		orgUnitsIds = append(orgUnitsIds, department.OrgUnitData.ID)
	}
	for _, project := range projects {
		orgUnitsIds = append(orgUnitsIds, project.OrgUnitData.ID)
	}
	result := daoutils.DB().WithContext(ctx).Where("org_unit_data_id in ?", orgUnitsIds).Unscoped().Delete(&models.OvertimeData{})
	if result.Error != nil {
		logError(ctx, tenantId, "OvertimeData", result.Error)
		return result.Error
	}
	return nil
}
