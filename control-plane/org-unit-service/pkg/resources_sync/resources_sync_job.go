/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package resources_sync

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"k8s.io/utils/ptr"

	"gorm.io/gorm"

	"github.com/google/uuid"
	"github.com/lib/pq"
	clusters_api "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
	clusters_openapi "github.com/run-ai/runai/control-plane/cluster-service/pkg/client/generated"
	org_unit_config "github.com/run-ai/runai/control-plane/org-unit-service/pkg/config"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/models"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/services"
	tenants_api "github.com/run-ai/runai/control-plane/tenants-manager/pkg/client"
	tenants_openapi "github.com/run-ai/runai/control-plane/tenants-manager/pkg/client/generated"
	"github.com/run-ai/runai/runai-common-packages/config"
	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

type ProjectToSync struct {
	ProjectId            string
	ProjectOrgUnitDataId string
}

type DepartmentToSync struct {
	DepartmentId            string
	DepartmentOrgUnitDataId string
}

func InitGoRoutines(ctx context.Context, stopCh <-chan struct{}, tenantsClient tenants_api.ITenantsManager, clusterClient clusters_api.IClusterClient) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("resources sync job: Init go routine for node pools sync job")

	jobConfig := config.Get[*org_unit_config.SyncJob]()
	if jobConfig == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("resources sync job: failed to get config for resources sync")
		return
	}
	intervalHours := jobConfig.SyncJobIntervalHours
	tenantInterval := jobConfig.SyncJobTenantIntervalSeconds
	if intervalHours <= 0 || tenantInterval <= 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("resources sync job: failed to get sync job config")
		return
	}

	ticker := time.NewTicker(time.Duration(intervalHours) * time.Hour)

	go func() {
		syncNodePoolsResources(ctx, tenantsClient, clusterClient, tenantInterval)
		for {
			select {
			case <-ticker.C:
				syncNodePoolsResources(ctx, tenantsClient, clusterClient, tenantInterval)
			case <-stopCh:
				ticker.Stop()
				return
			}
		}
	}()
}

func syncNodePoolsResources(ctx context.Context, tenantsClient tenants_api.ITenantsManager, clusterClient clusters_api.IClusterClient, tenantIntervalSeconds int) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("resources sync job: update node pools resources per tenant")

	tenants, err := getTenants(ctx, tenantsClient)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("resources sync job: failed to get tenants, err %v", err)
		return
	}

	ticker := time.NewTicker(time.Duration(tenantIntervalSeconds) * time.Second)
	defer ticker.Stop()

	for _, tenant := range tenants {
		<-ticker.C
		syncResourcesTenant(ctx, tenant, clusterClient)
		ticker.Reset(time.Duration(tenantIntervalSeconds) * time.Second)
	}
}

func shouldSetUnlimitedGpu(ctx context.Context, tenantId string, clusterId string) (bool, error) {
	return dao.HasUnlimitedGpuDepartment(ctx, clusterId, tenantId)
}

func syncResourcesTenant(ctx context.Context, tenant tenants_openapi.Tenant, clusterClient clusters_api.IClusterClient) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("resources sync job: update node pools resources for tenant %d", tenant.Id)
	db := dao_utils.DB()
	nodePoolsMap, err := getNodePoolsMap(ctx, clusterClient, fmt.Sprint(tenant.Id))
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("resources sync job: failed to get node pools, err %v", err)
		return
	}
	for cluster, nodePools := range nodePoolsMap {
		clusterUuid, err := uuid.Parse(cluster)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("resources sync job: failed to parse cluster %v, err %v", cluster, err)
			continue
		}
		setUnlimitedGpu, err := shouldSetUnlimitedGpu(ctx, fmt.Sprint(tenant.Id), clusterUuid.String())
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("resources sync job: failed to get department flag, err %v", err)
			return
		}
		updateResourcesPerCluster(ctx, fmt.Sprint(tenant.Id), clusterUuid, nodePools, db, setUnlimitedGpu)
	}
}

func updateResourcesPerCluster(ctx context.Context, tenantId string, clusterId uuid.UUID, nodePools []clusters_openapi.NodepoolMinimal, db *gorm.DB, setUnlimitedGpu bool) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("resources sync job: update node pools resources for cluster %s", clusterId.String())
	nodePoolsIds := []string{}
	nodePoolsNames := []string{}
	for _, nodePool := range nodePools {
		if nodePool.Id == "" || nodePool.Name == "" {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("resources sync job: failed to update resources on cluster %s, invalid node pool %+v", clusterId.String(), nodePool)
			return
		}
		nodePoolsIds = append(nodePoolsIds, nodePool.Id)
		nodePoolsNames = append(nodePoolsNames, nodePool.Name)
	}

	err := handleMissingResources(ctx, db, tenantId, clusterId, nodePoolsIds, nodePoolsNames, setUnlimitedGpu)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("resources sync job: failed to handle missing resources on cluster %s, err %v", clusterId.String(), err)
	}
	err = handleDeletedResources(ctx, db, tenantId, clusterId, nodePoolsIds, nodePoolsNames)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("resources sync job: failed to handle deleted resources on cluster %s, err %v", clusterId.String(), err)
	}
	err = handleNoResources(ctx, db, tenantId, clusterId, nodePoolsIds, nodePoolsNames, setUnlimitedGpu)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("resources sync job: failed to handle no resources on cluster %s, err %v", clusterId.String(), err)
	}
	err = handleDefaultNodePools(ctx, db, tenantId, clusterId, nodePoolsIds, nodePoolsNames)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("resources sync job: failed to handle default node pools on cluster %s, err %v", clusterId.String(), err)
	}
}

func handleDefaultNodePools(ctx context.Context, db *gorm.DB, tenantId string, clusterId uuid.UUID, nodePoolsIds []string, nodePoolsNames []string) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("resources sync job: handle default node pool list for tenant %s and cluster %s", tenantId, clusterId.String())
	projects, err := getProjectsWithDeletedDefaultNodePools(ctx, db, tenantId, clusterId, nodePoolsIds, nodePoolsNames)
	if err != nil {
		return err
	}
	projectsEffective, err := getProjectsEffectiveWithDeletedDefaultNodePools(ctx, db, tenantId, clusterId, nodePoolsIds, nodePoolsNames)
	if err != nil {
		return err
	}
	departments, err := getDepartmentsWithDeletedDefaultNodePools(ctx, db, tenantId, clusterId, nodePoolsIds, nodePoolsNames)
	if err != nil {
		return err
	}
	nodePoolMap := map[string]bool{}
	for i := range nodePoolsNames {
		nodePoolMap[nodePoolsNames[i]] = true
	}
	handleDepartmentNodePoolList(ctx, departments, nodePoolMap)
	handleProjectNodePoolList(ctx, nodePoolMap, projects)
	handleProjectEffectiveNodePoolList(ctx, nodePoolMap, projectsEffective)
	return nil
}

func updateProjectEffectiveFromProject(ctx context.Context, project *models.Project, validNodePoolList pq.StringArray) {
	project.OrgUnitData.DefaultNodePoolList = ptr.To(validNodePoolList)
	project.OrgUnitData.DefaultNodePoolListEffective = ptr.To(validNodePoolList)
	_, err := dao.UpdateProject(ctx, project, false, false)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("resources sync job: failed to update project %d with invalid default node pool list, err %v", project.ID, err)
	}
}

func updateProjectEffectiveFromParent(ctx context.Context, project *models.Project) {
	project.OrgUnitData.DefaultNodePoolList = nil
	parent, err := dao.GetDepartmentById(ctx, *project.ParentID, project.TenantId, models.SlimVerbosity, nil)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("resources sync job: failed to update project %d effective default node pool list, err %v", project.ID, err)
		return
	}
	parentNodePoolList := parent.OrgUnitData.DefaultNodePoolList
	if parentNodePoolList != nil && len(*parentNodePoolList) > 0 {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("resources sync job: project %d: changing effective node pool list to %v from parent %s", project.ID, *parentNodePoolList, parent.OrgUnitDataID.String())
		project.OrgUnitData.DefaultNodePoolListEffective = ptr.To(ptr.Deref(parentNodePoolList, []string{}))
		project.OrgUnitData.DefaultNodePoolListEffectiveOrigin = ptr.To(parent.OrgUnitDataID)
	} else {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("resources sync job: project %d: changing effective node pool list to {default}", project.ID)
		project.OrgUnitData.DefaultNodePoolListEffective = ptr.To(pq.StringArray{models.DefaultNodePoolName})
	}
	_, err = dao.UpdateProject(ctx, project, false, false)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("resources sync job: failed to update project %d with invalid default node pool list, err %v", project.ID, err)
	}
}

func handleProjectNodePoolList(ctx context.Context, nodePoolMap map[string]bool, projects []*models.Project) {
	for i := range projects {
		project := *(projects[i])
		validNodePoolList := getValidNodePoolList(project.OrgUnitData.DefaultNodePoolList, nodePoolMap)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("resources sync job: project %d: changing node pool list from %v to %v", project.ID, project.OrgUnitData.DefaultNodePoolList, validNodePoolList)
		if len(validNodePoolList) > 0 {
			// project defined its node pool list
			updateProjectEffectiveFromProject(ctx, &project, validNodePoolList)
		} else {
			// else - if we now need to get the default node pools from the department:
			// if the parent has a default list - update it, else put default in the effective list
			updateProjectEffectiveFromParent(ctx, &project)
		}
	}
}

func handleProjectEffectiveNodePoolList(ctx context.Context, nodePoolMap map[string]bool, projectsEffective []*models.Project) {
	for i := range projectsEffective {
		project := *(projectsEffective[i])
		validNodePoolList := getValidNodePoolList(project.OrgUnitData.DefaultNodePoolListEffective, nodePoolMap)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("resources sync job: project %d: changing effective node pool list from %v to %v", project.ID, project.OrgUnitData.DefaultNodePoolList, validNodePoolList)
		if len(validNodePoolList) == 0 {
			// if the effective is empty - make the effective be the project's list
			project.OrgUnitData.DefaultNodePoolListEffectiveOrigin = nil
			if project.OrgUnitData.DefaultNodePoolList != nil && len(*project.OrgUnitData.DefaultNodePoolList) > 0 {
				validNodePoolList = ptr.Deref(project.OrgUnitData.DefaultNodePoolList, []string{})
			} else {
				validNodePoolList = pq.StringArray{models.DefaultNodePoolName}
			}
		}
		project.OrgUnitData.DefaultNodePoolListEffective = ptr.To(validNodePoolList)
		_, err := dao.UpdateProject(ctx, &project, false, false)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("resources sync job: failed to update project %d with invalid default node pool list, err %v", project.ID, err)
		}
	}
}

func handleDepartmentNodePoolList(ctx context.Context, departments *[]models.Department, nodePoolMap map[string]bool) {
	for i := range *departments {
		department := (*departments)[i]
		validNodePoolList := getValidNodePoolList(department.OrgUnitData.DefaultNodePoolList, nodePoolMap)
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("resources sync job: department %d: changing node pool list from %v to %v", department.ID, department.OrgUnitData.DefaultNodePoolList, validNodePoolList)
		if len(validNodePoolList) == 0 {
			department.OrgUnitData.DefaultNodePoolList = nil
		} else {
			department.OrgUnitData.DefaultNodePoolList = ptr.To(validNodePoolList)
		}
		_, err := dao.UpdateDepartment(ctx, &department)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("resources sync job: failed to update department %d with invalid default node pool list, err %v", department.ID, err)
		}
	}
}

func handleDeletedResources(ctx context.Context, db *gorm.DB, tenantId string, clusterId uuid.UUID, nodePoolsIds []string, nodePoolsNames []string) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("resources sync job: handle deleted resources for tenant %s and cluster %s", tenantId, clusterId.String())
	projects, err := getProjectsWithDeletedResources(ctx, db, tenantId, clusterId, nodePoolsIds)
	if err != nil {
		return err
	}
	departments, err := getDepartmentsWithDeletedResources(ctx, db, tenantId, clusterId, nodePoolsIds)
	if err != nil {
		return err
	}
	nodePoolMap := map[string]string{}
	for i := range nodePoolsIds {
		nodePoolMap[nodePoolsIds[i]] = nodePoolsNames[i]
	}
	resourcesToDelete := []models.Resources{}
	for _, project := range projects {
		projectResources := getResourcesToDelete(ctx, project.OrgUnitData, nodePoolMap, fmt.Sprintf("project %d", project.ID))
		resourcesToDelete = append(resourcesToDelete, projectResources...)
	}
	for _, department := range *departments {
		departmentResources := getResourcesToDelete(ctx, department.OrgUnitData, nodePoolMap, fmt.Sprintf("department %d", department.ID))
		resourcesToDelete = append(resourcesToDelete, departmentResources...)
	}

	if len(resourcesToDelete) == 0 {
		return nil
	}
	return dao.DeleteResources(ctx, resourcesToDelete)
}

func handleMissingResources(ctx context.Context, db *gorm.DB, tenantId string, clusterId uuid.UUID, nodePoolsIds []string, nodePoolsNames []string, setUnlimitedGpu bool) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("resources sync job: handle missing resources for tenant %s and cluster %s", tenantId, clusterId.String())
	projects, err := getProjectsWithMissingResources(ctx, db, tenantId, clusterId, nodePoolsIds)
	if err != nil {
		return err
	}
	departments, err := getDepartmentsWithMissingResources(ctx, db, tenantId, clusterId, nodePoolsIds)
	if err != nil {
		return err
	}
	resourcesToAdd := []*models.Resources{}
	for _, project := range projects {
		projectResources := getProjectResourcesToAdd(ctx, project.OrgUnitData, nodePoolsIds, nodePoolsNames, project.ID)
		resourcesToAdd = append(resourcesToAdd, projectResources...)
	}
	for _, department := range *departments {
		departmentResources := getDepartmentResourcesToAdd(ctx, department.OrgUnitData, nodePoolsIds, nodePoolsNames, department.ID, setUnlimitedGpu && department.Name == services.DefaultDepartmentName)
		resourcesToAdd = append(resourcesToAdd, departmentResources...)
	}

	if len(resourcesToAdd) == 0 {
		return nil
	}
	return dao.AddResources(ctx, resourcesToAdd)
}

func handleNoResources(ctx context.Context, db *gorm.DB, tenantId string, clusterId uuid.UUID, nodePoolsIds []string, nodePoolsNames []string, setUnlimitedGpu bool) error {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("resources sync job: handle no resources for tenant %s and cluster %s", tenantId, clusterId.String())
	projects, err := getProjectsWithNoResources(ctx, db, tenantId, clusterId)
	if err != nil {
		return err
	}
	departments, err := getDepartmentsWithNoResources(ctx, db, tenantId, clusterId)
	if err != nil {
		return err
	}
	resourcesToAdd := []*models.Resources{}
	for _, project := range projects {
		projectResources := getProjectResourcesToAdd(ctx, project.OrgUnitData, nodePoolsIds, nodePoolsNames, project.ID)
		resourcesToAdd = append(resourcesToAdd, projectResources...)
	}
	for _, department := range *departments {
		departmentResources := getDepartmentResourcesToAdd(ctx, department.OrgUnitData, nodePoolsIds, nodePoolsNames, department.ID, setUnlimitedGpu && department.Name == services.DefaultDepartmentName)
		resourcesToAdd = append(resourcesToAdd, departmentResources...)
	}

	if len(resourcesToAdd) == 0 {
		return nil
	}
	return dao.AddResources(ctx, resourcesToAdd)
}

func getProjectResourcesToAdd(ctx context.Context, orgUnitData models.OrgUnitData, nodePoolsIds []string, nodePoolsNames []string, projectId int64) []*models.Resources {
	resourcesToAdd := []*models.Resources{}
	resourcesMap := createResourcesMap(orgUnitData)
	for i := range nodePoolsIds {
		nodePoolId := nodePoolsIds[i]
		resources, found := resourcesMap[nodePoolId]
		if !found || resources == nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("resources sync job: found missing resources for project %d, org-unit-data %s, node pool id: %s name: %s",
				projectId, orgUnitData.ID.String(), nodePoolId, nodePoolsNames[i])
			resourcesToAdd = append(resourcesToAdd, services.GetProjectResourcesToAdd(orgUnitData, nodePoolId, nodePoolsNames[i]))
		}
	}
	return resourcesToAdd
}

func getDepartmentResourcesToAdd(ctx context.Context, orgUnitData models.OrgUnitData, nodePoolsIds []string, nodePoolsNames []string, departmentId int64, setUnlimitedDeserved bool) []*models.Resources {
	resourcesToAdd := []*models.Resources{}
	resourcesMap := createResourcesMap(orgUnitData)
	for i := range nodePoolsIds {
		nodePoolId := nodePoolsIds[i]
		resources, found := resourcesMap[nodePoolId]
		if !found || resources == nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("resources sync job: found missing resources for department %d, org-unit-data %s, node pool id: %s name: %s",
				departmentId, orgUnitData.ID.String(), nodePoolId, nodePoolsNames[i])
			resourcesToAdd = append(resourcesToAdd, services.GetDepartmentResourcesToAdd(orgUnitData, setUnlimitedDeserved, nodePoolId, nodePoolsNames[i]))
		}
	}
	return resourcesToAdd
}

func getResourcesToDelete(ctx context.Context, orgUnitData models.OrgUnitData, nodePoolMap map[string]string, orgUnitType string) []models.Resources {
	resourcesToDelete := []models.Resources{}
	resources := orgUnitData.Resources
	for i := range resources {
		existingNodePool, found := nodePoolMap[resources[i].NodePoolId]
		if !found || existingNodePool != resources[i].NodePoolName {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("resources sync job: found resources to delete for %s, org-unit-data %s, node pool id: %s name: %s",
				orgUnitType, orgUnitData.ID.String(), resources[i].NodePoolId, resources[i].NodePoolName)
			resourcesToDelete = append(resourcesToDelete, resources[i])
		}
	}
	return resourcesToDelete
}

func getValidNodePoolList(defaultNodePoolList *pq.StringArray, nodePoolMap map[string]bool) pq.StringArray {
	validNodePoolList := []string{}
	if defaultNodePoolList == nil {
		return validNodePoolList
	}
	for _, nodePool := range *defaultNodePoolList {
		np, exists := nodePoolMap[nodePool]
		if np && exists {
			validNodePoolList = append(validNodePoolList, nodePool)
		}
	}
	return validNodePoolList
}

func getProjectsByQuery(ctx context.Context, db *gorm.DB, query string, tenantId string, clusterId uuid.UUID) ([]*models.Project, error) {
	projectsToSync := []ProjectToSync{}
	result := db.Raw(query).Scan(&projectsToSync)
	if result.Error != nil {
		return nil, result.Error
	}
	projectIds := []string{}
	for i := range projectsToSync {
		projectIds = append(projectIds, projectsToSync[i].ProjectId)
	}
	if len(projectIds) == 0 {
		return []*models.Project{}, nil
	}
	return dao.GetProjectsByIds(ctx, tenantId, &clusterId, projectIds, nil)
}

func getDepartmentsByQuery(ctx context.Context, db *gorm.DB, query string, tenantId string, clusterId uuid.UUID) (*[]models.Department, error) {
	departmentsToSync := []DepartmentToSync{}
	result := db.Raw(query).Scan(&departmentsToSync)
	if result.Error != nil {
		return nil, result.Error
	}
	departmentIds := []string{}
	for i := range departmentsToSync {
		departmentIds = append(departmentIds, departmentsToSync[i].DepartmentId)
	}
	if len(departmentIds) == 0 {
		return &[]models.Department{}, nil
	}
	return dao.GetDepartmentByClusterAndIds(ctx, tenantId, clusterId, departmentIds, nil)
}

func getProjectsWithMissingResources(ctx context.Context, db *gorm.DB, tenantId string, clusterId uuid.UUID, nodePoolsIds []string) ([]*models.Project, error) {
	query := fmt.Sprintf(MissingProjectNodePoolsOrgUnits, "['"+strings.Join(nodePoolsIds, "','")+"']", clusterId.String(), tenantId, len(nodePoolsIds))
	projects, err := getProjectsByQuery(ctx, db, query, tenantId, clusterId)
	if err == nil && len(projects) > 0 {
		projectIds := []int64{}
		for _, project := range projects {
			projectIds = append(projectIds, project.ID)
		}
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("resources sync job: found %d projects with missing node pool resources: %v", len(projects), projectIds)
	}
	return projects, err
}

func getDepartmentsWithMissingResources(ctx context.Context, db *gorm.DB, tenantId string, clusterId uuid.UUID, nodePoolsIds []string) (*[]models.Department, error) {
	query := fmt.Sprintf(MissingDepartmentNodePoolsOrgUnits, "['"+strings.Join(nodePoolsIds, "','")+"']", clusterId.String(), tenantId, len(nodePoolsIds))
	departments, err := getDepartmentsByQuery(ctx, db, query, tenantId, clusterId)
	if err == nil && len(*departments) > 0 {
		departmentIds := []int64{}
		for _, department := range *departments {
			departmentIds = append(departmentIds, department.ID)
		}
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("resources sync job: found %d departments with missing node pool resources: %v", len(*departments), departmentIds)
	}
	return departments, err
}

func getProjectsWithNoResources(ctx context.Context, db *gorm.DB, tenantId string, clusterId uuid.UUID) ([]*models.Project, error) {
	query := fmt.Sprintf(EmptyProjects, models.ProjectViewName, clusterId.String(), tenantId)
	projects, err := getProjectsByQuery(ctx, db, query, tenantId, clusterId)
	if err == nil && len(projects) > 0 {
		projectIds := []int64{}
		for _, project := range projects {
			projectIds = append(projectIds, project.ID)
		}
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("resources sync job: found %d projects with no node pool resources: %v", len(projects), projectIds)
	}
	return projects, err
}

func getDepartmentsWithNoResources(ctx context.Context, db *gorm.DB, tenantId string, clusterId uuid.UUID) (*[]models.Department, error) {
	query := fmt.Sprintf(EmptyDepartments, models.DepartmentStatusViewName, clusterId.String(), tenantId)
	departments, err := getDepartmentsByQuery(ctx, db, query, tenantId, clusterId)
	if err == nil && len(*departments) > 0 {
		departmentIds := []int64{}
		for _, department := range *departments {
			departmentIds = append(departmentIds, department.ID)
		}
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("resources sync job: found %d departments with no node pool resources: %v", len(*departments), departmentIds)
	}
	return departments, err
}

func getProjectsWithDeletedResources(ctx context.Context, db *gorm.DB, tenantId string, clusterId uuid.UUID, nodePoolsIds []string) ([]*models.Project, error) {
	query := fmt.Sprintf(DeletedProjectNodePoolsOrgUnits, "['"+strings.Join(nodePoolsIds, "','")+"']", clusterId.String(), tenantId)
	projects, err := getProjectsByQuery(ctx, db, query, tenantId, clusterId)
	if err == nil && len(projects) > 0 {
		projectIds := []int64{}
		for _, project := range projects {
			projectIds = append(projectIds, project.ID)
		}
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("resources sync job: found %d projects with deleted node pool resources: %v", len(projects), projectIds)
	}
	return projects, err
}

func getDepartmentsWithDeletedResources(ctx context.Context, db *gorm.DB, tenantId string, clusterId uuid.UUID, nodePoolsIds []string) (*[]models.Department, error) {
	query := fmt.Sprintf(DeletedDepartmentNodePoolsOrgUnits, "['"+strings.Join(nodePoolsIds, "','")+"']", clusterId.String(), tenantId)
	departments, err := getDepartmentsByQuery(ctx, db, query, tenantId, clusterId)
	if err == nil && len(*departments) > 0 {
		departmentIds := []int64{}
		for _, department := range *departments {
			departmentIds = append(departmentIds, department.ID)
		}
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("resources sync job: found %d departments with deleted node pool resources: %v", len(*departments), departmentIds)
	}
	return departments, err
}

func getProjectsWithDeletedDefaultNodePools(ctx context.Context, db *gorm.DB, tenantId string, clusterId uuid.UUID, nodePoolsIds []string, nodePoolsNames []string) ([]*models.Project, error) {
	query := fmt.Sprintf(DeletedProjectDefaultNodePoolsOrgUnits, "['"+strings.Join(nodePoolsIds, "','")+"']", "['"+strings.Join(nodePoolsNames, "','")+"']", clusterId.String(), tenantId)
	projects, err := getProjectsByQuery(ctx, db, query, tenantId, clusterId)
	if err == nil && len(projects) > 0 {
		projectIds := []int64{}
		for _, project := range projects {
			projectIds = append(projectIds, project.ID)
		}
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("resources sync job: found %d projects with deleted node pool in node pool list: %v", len(projects), projectIds)
	}
	return projects, err
}

func getProjectsEffectiveWithDeletedDefaultNodePools(ctx context.Context, db *gorm.DB, tenantId string, clusterId uuid.UUID, nodePoolsIds []string, nodePoolsNames []string) ([]*models.Project, error) {
	query := fmt.Sprintf(DeletedProjectDefaultNodePoolsEffectiveOrgUnits, "['"+strings.Join(nodePoolsIds, "','")+"']", "['"+strings.Join(nodePoolsNames, "','")+"']", clusterId.String(), tenantId)
	projects, err := getProjectsByQuery(ctx, db, query, tenantId, clusterId)
	if err == nil && len(projects) > 0 {
		projectIds := []int64{}
		for _, project := range projects {
			projectIds = append(projectIds, project.ID)
		}
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("resources sync job: found %d projects with deleted node pool in effective  node pool list: %v", len(projects), projectIds)
	}
	return projects, err
}

func getDepartmentsWithDeletedDefaultNodePools(ctx context.Context, db *gorm.DB, tenantId string, clusterId uuid.UUID, nodePoolsIds []string, nodePoolsNames []string) (*[]models.Department, error) {
	query := fmt.Sprintf(DeletedDepartmentDefaultNodePoolsOrgUnits, "['"+strings.Join(nodePoolsIds, "','")+"']", "['"+strings.Join(nodePoolsNames, "','")+"']", clusterId.String(), tenantId)
	departments, err := getDepartmentsByQuery(ctx, db, query, tenantId, clusterId)
	if err == nil && len(*departments) > 0 {
		departmentIds := []int64{}
		for _, department := range *departments {
			departmentIds = append(departmentIds, department.ID)
		}
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("resources sync job: found %d departments with deleted node pool in node pool list: %v", len(*departments), departmentIds)
	}
	return departments, err
}

func createResourcesMap(orgUnitData models.OrgUnitData) map[string]*models.Resources {
	resourcesMap := map[string]*models.Resources{}
	for i := range orgUnitData.Resources {
		resource := orgUnitData.Resources[i]
		resourcesMap[resource.NodePoolId] = &resource
	}
	return resourcesMap
}

func getTenants(ctx context.Context, tenantsClient tenants_api.ITenantsManager) ([]tenants_openapi.Tenant, error) {
	tenants, response, err := tenantsClient.ListAllTenantsInternal(ctx)
	if err != nil {
		return []tenants_openapi.Tenant{}, err
	}
	if response.StatusCode != http.StatusOK {
		return []tenants_openapi.Tenant{}, fmt.Errorf("failed to get tenants, response code: %v", response.StatusCode)
	}
	return tenants, nil
}

func getNodePoolsMap(ctx context.Context, clusterClient clusters_api.IClusterClient, tenantId string) (map[string][]clusters_openapi.NodepoolMinimal, error) {
	clusterNodePools, err := clusterClient.GetNodePoolsMinimalForTenants(ctx, []string{tenantId})
	nodePoolsMap := map[string][]clusters_openapi.NodepoolMinimal{}
	if err != nil {
		return nodePoolsMap, err
	}
	if clusterNodePools == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("resources sync job: did not find any clusters with node pools for tenant %s", tenantId)
		return nodePoolsMap, nil
	}
	for i := range clusterNodePools.Nodepools {
		nodePool := clusterNodePools.Nodepools[i]
		clusterId := nodePool.ClusterId
		nodePools, found := nodePoolsMap[clusterId]
		if !found {
			nodePoolsMap[clusterId] = []clusters_openapi.NodepoolMinimal{}
			nodePools = nodePoolsMap[clusterId]
		}
		nodePoolsMap[clusterId] = append(nodePools, nodePool)
	}
	return nodePoolsMap, nil
}
