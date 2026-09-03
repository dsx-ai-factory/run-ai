/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dsx-ai-factory/run-ai/runai-application-packages/queryapi"

	authorization_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	cluster_service_client "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
	tenants_manager_client "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"

	"k8s.io/utils/ptr"

	authorization_api "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	cluster_api "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
)

const (
	DefaultResourceRank      = models.DefaultResourceRank
	DefaultResourceRankValue = 100
)

var RankValuesMap = map[string]int32{
	"Highest":    255,
	"VeryHigh":   240,
	"High":       210,
	"MediumHigh": 180,
	"Medium":     150,
	"MediumLow":  DefaultResourceRankValue,
	"Low":        50,
	"VeryLow":    20,
	"Lowest":     1,
}

type OrgUnitApiService struct {
	AuthorizationClient   authorization_client.IAuthorizationClient
	ClustersClient        cluster_service_client.IClusterClient
	TenantsClient         tenants_manager_client.ITenantsManager
	clusterExternalClient cluster_service_client.IClusterExternalClient
}

func NewOrgUnitApiService(
	authorizationClient authorization_client.IAuthorizationClient,
	clustersClient cluster_service_client.IClusterClient,
	tenantsClient tenants_manager_client.ITenantsManager,
	clusterExternalClient cluster_service_client.IClusterExternalClient,
) api.OrgUnitAPIServicer {
	return &OrgUnitApiService{
		AuthorizationClient:   authorizationClient,
		ClustersClient:        clustersClient,
		TenantsClient:         tenantsClient,
		clusterExternalClient: clusterExternalClient,
	}
}

func (o OrgUnitApiService) GetInternalDeletedOrgUnits(ctx context.Context, deletedAfter time.Time) (api.ImplResponse, error) {
	tenantId, tenantErr := ParseTenantId(ctx)
	if tenantErr != nil {
		return HandleError(tenantErr, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	var deletedOrgUnits []api.DeletedOrgUnit

	// clusters:
	clusters, err := o.ClustersClient.GetInternalClusters(ctx, true)
	if err != nil {
		return HandleError(err, http.StatusServiceUnavailable, "Failed to get clusters"), nil
	}
	deletedClusterIds := []string{}
	for _, cluster := range clusters {
		if cluster.DeletedAt.IsSet() && cluster.DeletedAt.Get() != nil {
			deletedClusterIds = append(deletedClusterIds, cluster.Uuid)
			deletedOrgUnit := api.DeletedOrgUnit{
				Id:        cluster.Uuid,
				Name:      cluster.Name,
				Type:      "cluster",
				CreatedAt: cluster.CreatedAt,
				DeletedAt: *cluster.DeletedAt.Get(),
			}
			if cluster.UpdatedAt.Get() != nil {
				deletedOrgUnit.UpdatedAt = *cluster.UpdatedAt.Get()
			}
			deletedOrgUnits = append(deletedOrgUnits, deletedOrgUnit)
		}
	}
	// departments:
	deletedDepartments, getDeptErr := getDeletedDepartments(ctx, deletedAfter, *tenantId, deletedClusterIds)
	if getDeptErr != nil {
		return HandleError(getDeptErr, http.StatusInternalServerError, fmt.Sprintf("failed to fetch deleted departments %s", getDeptErr.Error())), nil
	}
	deletedOrgUnits = append(deletedOrgUnits, MapDepartmentsToDeleted(deletedDepartments)...)
	// projects:
	deletedProjects, getProjErr := getDeletedProjects(ctx, deletedAfter, *tenantId, deletedClusterIds)
	if getProjErr != nil {
		return HandleError(getProjErr, http.StatusInternalServerError, fmt.Sprintf("failed to fetch deleted projects %s", getProjErr.Error())), nil
	}
	deletedOrgUnits = append(deletedOrgUnits, MapProjectsToDeleted(deletedProjects)...)

	return api.Response(http.StatusOK, api.GetInternalDeletedOrgUnits200Response{OrgUnits: deletedOrgUnits}), nil
}

func getDeletedProjects(ctx context.Context, deletedAfter time.Time, tenantId string, deletedClusterIds []string) ([]models.Project, error) {
	projects, err := dao.GetDeletedProjects(ctx, tenantId, deletedAfter)
	if err != nil {
		return projects, nil
	}
	deletedClustersProjects, err := dao.GetProjectsByClusters(ctx, tenantId, deletedClusterIds)
	if err != nil {
		return deletedClustersProjects, nil
	}
	projects = append(projects, deletedClustersProjects...)
	return projects, nil
}

func getDeletedDepartments(ctx context.Context, deletedAfter time.Time, tenantId string, deletedClusterIds []string) ([]models.Department, error) {
	departments, err := dao.GetDeletedDepartments(ctx, tenantId, deletedAfter)
	if err != nil {
		return departments, nil
	}
	deletedClustersDepartments, err := dao.GetDepartmentsByClusters(ctx, tenantId, deletedClusterIds)
	if err != nil {
		return deletedClustersDepartments, nil
	}
	departments = append(departments, deletedClustersDepartments...)
	return departments, nil
}

func (o OrgUnitApiService) GetOrgUnitsNames(ctx context.Context, orgUnitType string, orgUnitsIds string) (api.ImplResponse, error) {
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	orgUnitsIntIds, err := ParseOrgUnitIds(orgUnitsIds)
	if err != nil {
		return HandleError(err, http.StatusBadRequest, fmt.Sprintf("Failed to parse org units ids %s", orgUnitsIds)), nil
	}

	response := []api.OrgUnit{}
	switch orgUnitType {
	case "project":
		for _, orgUnitIntId := range orgUnitsIntIds {
			project, err := dao.GetProjectById(ctx, orgUnitIntId, *tenantId, models.MetaDataVerbosity)
			if err != nil {
				return HandleDbError(ctx, err, "Failed to get project"), nil
			}

			response = append(response, api.OrgUnit{
				Name: project.Name,
				Id:   fmt.Sprintf("%d", orgUnitIntId),
				Type: api.SCOPETYPE_PROJECT,
			})
		}
	case "department":
		for _, orgUnitIntId := range orgUnitsIntIds {
			department, err := dao.GetDepartmentById(ctx, orgUnitIntId, *tenantId, models.MetaDataVerbosity, nil)
			if err != nil {
				return HandleDbError(ctx, err, "Failed to get department"), nil
			}

			response = append(response, api.OrgUnit{
				Name: department.Name,
				Id:   fmt.Sprintf("%d", orgUnitIntId),
				Type: api.SCOPETYPE_DEPARTMENT,
			})
		}
	default:
		return HandleError(err, http.StatusBadRequest, fmt.Sprintf("Got wrong org unit type %s", orgUnitType)), nil
	}

	return api.Response(http.StatusOK, response), nil
}

func (o OrgUnitApiService) GetRanks(ctx context.Context) (api.ImplResponse, error) {
	var ranks []api.Rank
	for rank, value := range RankValuesMap {
		isDefault := rank == DefaultResourceRank
		ranks = append(ranks, api.Rank{
			Name:      rank,
			Value:     value,
			IsDefault: isDefault,
		})
	}

	response := api.GetRanks200Response{Ranks: ranks}
	return api.Response(http.StatusOK, response), nil
}

func (o OrgUnitApiService) GetInternalOrgTree(ctx context.Context) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Getting internal org tree")
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	tenantOrgUnit, err := o.GetTenantOrgUnit(ctx, *tenantId)
	if err != nil {
		return HandleError(err, http.StatusServiceUnavailable, "Failed to get tenant by id"), nil
	}
	orgUnits := []api.OrgUnit{tenantOrgUnit}

	clusterOrgUnits, err := o.GetClusterOrgUnitsInternal(ctx)
	if err != nil {
		return HandleError(err, http.StatusServiceUnavailable, "Failed to query clusters"), nil
	}
	orgUnits = append(orgUnits, clusterOrgUnits...)

	// Internal: do NOT filter by permitted scopes
	departmentOrgUnits, err := o.GetDepartmentOrgUnitsInternal(ctx, *tenantId)
	if err != nil {
		return HandleError(err, http.StatusServiceUnavailable, "Failed to query departments"), nil
	}
	orgUnits = append(orgUnits, departmentOrgUnits...)

	projectOrgUnits, err := o.GetProjectOrgUnitsInternal(ctx, *tenantId, departmentOrgUnits)
	if err != nil {
		return HandleError(err, http.StatusServiceUnavailable, "Failed to query departments"), nil
	}

	orgUnits = append(orgUnits, projectOrgUnits...)

	response := api.GetInternalOrgTree200Response{OrgUnits: orgUnits}
	return api.Response(http.StatusOK, response), nil
}

func (o OrgUnitApiService) GetOrgTree(ctx context.Context) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Getting org tree")
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	// TODO: Cache tenant name
	tenantOrgUnit, err := o.GetTenantOrgUnit(ctx, *tenantId)
	if err != nil {
		return HandleError(err, http.StatusServiceUnavailable, "Failed to get tenant by id"), nil
	}
	orgUnits := []api.OrgUnit{tenantOrgUnit}

	clusterOrgUnits, err := o.GetClusterOrgUnits(ctx)
	if err != nil {
		return HandleError(err, http.StatusServiceUnavailable, "Failed to query clusters"), nil
	}
	orgUnits = append(orgUnits, clusterOrgUnits...)

	departmentOrgUnits, err := o.GetDepartmentOrgUnits(ctx, *tenantId)
	if err != nil {
		return HandleError(err, http.StatusServiceUnavailable, "Failed to query departments"), nil
	}
	orgUnits = append(orgUnits, departmentOrgUnits...)

	projectOrgUnits, err := o.GetProjectOrgUnits(ctx, *tenantId, departmentOrgUnits)
	if err != nil {
		return HandleError(err, http.StatusServiceUnavailable, "Failed to query departments"), nil
	}

	orgUnits = append(orgUnits, projectOrgUnits...)

	// The order of org units is important for the cleanup procedure
	// Need to be Tenants, Clusters, Departments, Projects
	orgUnits = cleanOrphans(orgUnits)

	response := api.GetInternalOrgTree200Response{OrgUnits: orgUnits}
	return api.Response(http.StatusOK, response), nil
}

func (o OrgUnitApiService) UpdateOrgUnitSettings(ctx context.Context, setting api.OrgUnitSingleSetting) (api.ImplResponse, error) {
	settingName := setting.Name
	settingValue := setting.Value

	if settingName != CpuResourcesFlag || settingValue != "false" {
		return api.Response(http.StatusOK, nil), nil
	}

	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	clusters, err := o.ClustersClient.GetClustersMinimalData(ctx)
	if err != nil {
		return HandleError(err, http.StatusServiceUnavailable, "Failed to get clusters"), nil
	}

	err = dao.DisableCpuQuota(ctx, *tenantId, clusters.Clusters)
	if err != nil {
		return HandleError(err, http.StatusServiceUnavailable, "Failed to update resources"), nil
	}

	return api.Response(http.StatusOK, nil), nil
}

func cleanOrphans(units []api.OrgUnit) []api.OrgUnit {
	format := "%s-%s"
	orgsMap := map[string]bool{}
	orgUnits := []api.OrgUnit{}
	for _, unit := range units {
		if unit.Type != api.SCOPETYPE_PROJECT {
			orgsMap[fmt.Sprintf(format, string(unit.Type), unit.Id)] = true
		}
	}

	for _, unit := range units {
		if unit.Type == api.SCOPETYPE_TENANT {
			orgUnits = append(orgUnits, unit)
			continue
		}
		if found, ok := orgsMap[fmt.Sprintf(format, string(unit.ParentType), unit.ParentId)]; found && ok {
			orgUnits = append(orgUnits, unit)
		} else {
			orgsMap[fmt.Sprintf(format, string(unit.Type), unit.Id)] = false
		}
	}

	return orgUnits
}

func (o OrgUnitApiService) GetTenantOrgUnit(ctx context.Context, tenantIdStr string) (api.OrgUnit, error) {
	tenantId, parseErr := strconv.Atoi(tenantIdStr)
	if parseErr != nil {
		return api.OrgUnit{}, parseErr
	}

	tenant, _, err := o.TenantsClient.GetTenantByIdInternal(ctx, int32(tenantId))
	if err != nil {
		return api.OrgUnit{}, err
	}

	tenantOrgUnit := api.OrgUnit{
		Id:   tenantIdStr,
		Type: api.SCOPETYPE_TENANT,
		Name: tenant.Name,
	}

	return tenantOrgUnit, nil
}

func (o OrgUnitApiService) GetClusterOrgUnits(ctx context.Context) ([]api.OrgUnit, error) {
	clusters, err := o.ClustersClient.GetClustersMinimalData(ctx)
	if err != nil {
		return nil, err
	}

	clusterOrgUnits := constructOrgUnit(
		clusters.Clusters,
		api.SCOPETYPE_CLUSTER,
		api.SCOPETYPE_TENANT,
		func(cluster cluster_api.ClustersMinimalData) string {
			return cluster.Name
		},
		func(cluster cluster_api.ClustersMinimalData) string {
			return cluster.Uuid
		},
		func(cluster cluster_api.ClustersMinimalData) string {
			return fmt.Sprintf("%d", cluster.TenantId)
		},
	)

	return clusterOrgUnits, nil
}

func (o OrgUnitApiService) GetDepartmentOrgUnits(ctx context.Context, tenantId string) ([]api.OrgUnit, error) {
	permittedScopes, _, err := o.AuthorizationClient.GetPermittedScopes(ctx, authorization_api.RESOURCETYPE_DEPARTMENT, ptr.To(authorization_api.ACTION_READ))
	if err != nil {
		return nil, err
	}

	departments, err := dao.GetDepartments(ctx, tenantId, []queryapi.Filter{}, "", "", 0, 0, &authorization_api.PermittedScopes{
		Tenants:     permittedScopes.Read.Tenants,
		Clusters:    permittedScopes.Read.Clusters,
		Departments: permittedScopes.Read.Departments,
		Projects:    permittedScopes.Read.Projects,
	}, false, models.MetaDataVerbosity)
	if err != nil {
		return nil, err
	}

	departmentOrgUnits := constructOrgUnit(
		*departments,
		api.SCOPETYPE_DEPARTMENT,
		api.SCOPETYPE_CLUSTER,
		func(department models.Department) string {
			return department.Name
		},
		func(department models.Department) string {
			return fmt.Sprintf("%d", department.ID)
		},
		func(department models.Department) string {
			return department.ClusterId.String()
		},
	)

	return departmentOrgUnits, nil
}

func (o OrgUnitApiService) GetProjectOrgUnits(ctx context.Context, tenantId string, departmentUnits []api.OrgUnit) ([]api.OrgUnit, error) {
	permittedScopes, _, err := o.AuthorizationClient.GetPermittedScopes(ctx, authorization_api.RESOURCETYPE_PROJECT, ptr.To(authorization_api.ACTION_READ))
	if err != nil {
		return nil, err
	}

	filters, _ := queryapi.NewParser(models.Project{}.SortAndFilterFieldsToColumn()).Filters([]string{"phase!='Deleting'"})

	projects, err := dao.GetProjects(ctx, tenantId, filters, "", "", 0, 0, "", &authorization_api.PermittedScopes{
		Tenants:     permittedScopes.Read.Tenants,
		Clusters:    permittedScopes.Read.Clusters,
		Departments: permittedScopes.Read.Departments,
		Projects:    permittedScopes.Read.Projects,
	}, false, models.MetaDataVerbosity)
	if err != nil {
		return nil, err
	}

	var projectParentType api.ScopeType
	var getProjectParentIdFunc func(models.Project) string

	projectParentType = api.SCOPETYPE_DEPARTMENT
	getProjectParentIdFunc = func(project models.Project) string {
		return fmt.Sprintf("%d", convutils.Deref(project.ParentID, 0))
	}
	missingDepartments, err := findMissingDepartments(ctx, tenantId, *projects, departmentUnits)
	if err != nil {
		return nil, err
	}

	projectOrgUnits := constructOrgUnit(
		*projects,
		api.SCOPETYPE_PROJECT,
		projectParentType,
		func(project models.Project) string {
			return project.Name
		},
		func(project models.Project) string {
			return fmt.Sprintf("%d", project.ID)
		},
		getProjectParentIdFunc,
	)

	return append(projectOrgUnits, missingDepartments...), nil
}

func findMissingDepartments(ctx context.Context, tenantId string, projects []models.Project, departmentUnits []api.OrgUnit) ([]api.OrgUnit, error) {
	missingDepartments := []api.OrgUnit{}

	departmentsToFetch := []int64{}
	for _, project := range projects {
		if project.ParentID == nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Got a project without parentId %s, cannot retrieve parent", project.Name)
			return nil, fmt.Errorf("got a project without parentId %s, cannot retrieve parent", project.Name)
		}
		found := false
		for _, department := range departmentUnits {
			if department.Id == fmt.Sprintf("%d", *project.ParentID) {
				found = true
			}
		}
		if !found {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Got a project without parent %s", project.Name)
			departmentsToFetch = append(departmentsToFetch, *project.ParentID)
		}
	}
	if len(departmentsToFetch) > 0 {
		departments, err := dao.GetDepartmentsByIds(ctx, tenantId, departmentsToFetch, nil)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Cannot retrieve departments for org unit: %s", err.Error())
			return nil, err
		}
		for _, department := range *departments {
			missingDepartments = append(missingDepartments, api.OrgUnit{
				Id:         fmt.Sprintf("%d", department.ID),
				Type:       api.SCOPETYPE_DEPARTMENT,
				Name:       department.Name,
				ParentId:   department.ClusterId.String(),
				ParentType: api.SCOPETYPE_CLUSTER,
			})
		}
	}

	return missingDepartments, nil
}

func constructOrgUnit[T any](entities []T, orgType api.ScopeType, parentOrgType api.ScopeType, getName func(T) string, getId func(T) string, getParentId func(T) string) []api.OrgUnit {
	orgUnits := []api.OrgUnit{}

	for _, entity := range entities {
		orgUnits = append(orgUnits, api.OrgUnit{
			Id:         getId(entity),
			Type:       orgType,
			Name:       getName(entity),
			ParentId:   getParentId(entity),
			ParentType: parentOrgType,
		})
	}

	return orgUnits
}

func (o OrgUnitApiService) GetClusterOrgUnitsInternal(ctx context.Context) ([]api.OrgUnit, error) {
	clusters, err := o.ClustersClient.GetInternalClusters(ctx, false)
	if err != nil {
		return nil, err
	}

	clusterOrgUnits := constructOrgUnit(
		clusters,
		api.SCOPETYPE_CLUSTER,
		api.SCOPETYPE_TENANT,
		func(cluster cluster_api.DisplayedCluster) string {
			return cluster.Name
		},
		func(cluster cluster_api.DisplayedCluster) string {
			return cluster.Uuid
		},
		func(cluster cluster_api.DisplayedCluster) string {
			return fmt.Sprintf("%d", cluster.TenantId)
		},
	)

	return clusterOrgUnits, nil
}

func (o OrgUnitApiService) GetDepartmentOrgUnitsInternal(ctx context.Context, tenantId string) ([]api.OrgUnit, error) {
	departments, err := dao.GetDepartments(ctx, tenantId, []queryapi.Filter{}, "", "", 0, 0, nil, false, models.MetaDataVerbosity)
	if err != nil {
		return nil, err
	}

	departmentOrgUnits := constructOrgUnit(
		*departments,
		api.SCOPETYPE_DEPARTMENT,
		api.SCOPETYPE_CLUSTER,
		func(department models.Department) string {
			return department.Name
		},
		func(department models.Department) string {
			return fmt.Sprintf("%d", department.ID)
		},
		func(department models.Department) string {
			return department.ClusterId.String()
		},
	)

	return departmentOrgUnits, nil
}

func (o OrgUnitApiService) GetProjectOrgUnitsInternal(ctx context.Context, tenantId string, departmentUnits []api.OrgUnit) ([]api.OrgUnit, error) {
	filters, _ := queryapi.NewParser(models.Project{}.SortAndFilterFieldsToColumn()).Filters([]string{"phase!='Deleting'"})
	projects, err := dao.GetProjects(ctx, tenantId, filters, "", "", 0, 0, "", nil, false, models.MetaDataVerbosity)
	if err != nil {
		return nil, err
	}

	var projectParentType api.ScopeType
	var getProjectParentIdFunc func(models.Project) string

	projectParentType = api.SCOPETYPE_DEPARTMENT
	getProjectParentIdFunc = func(project models.Project) string {
		return fmt.Sprintf("%d", convutils.Deref(project.ParentID, 0))
	}
	missingDepartments, err := findMissingDepartments(ctx, tenantId, *projects, departmentUnits)
	if err != nil {
		return nil, err
	}

	projectOrgUnits := constructOrgUnit(
		*projects,
		api.SCOPETYPE_PROJECT,
		projectParentType,
		func(project models.Project) string {
			return project.Name
		},
		func(project models.Project) string {
			return fmt.Sprintf("%d", project.ID)
		},
		getProjectParentIdFunc,
	)

	return append(projectOrgUnits, missingDepartments...), nil
}

func (o OrgUnitApiService) OrgsNodePoolSummary(ctx context.Context, nodepoolId string, nodepoolSummaryType api.OrgUnitsNodepoolSummaryType) (api.ImplResponse, error) {
	if !nodepoolSummaryType.IsValid() {
		return HandleError(errors.New("invalid nodepool summary type"), http.StatusBadRequest, "Invalid nodepool summary type"), nil
	}

	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	summary, err := dao.GetOrgUnitsNodepoolSummary(ctx, *tenantId, nodepoolId, nodepoolSummaryType)
	if err != nil {
		return HandleDbError(ctx, err, "Failed to get org units nodepool summary"), nil
	}

	response := api.OrgsNodePoolSummary200Response{OrgUnitsNodepoolSummary: summary}
	return api.Response(http.StatusOK, response), nil
}

func (o OrgUnitApiService) NodePoolResetResources(ctx context.Context, request api.OrgUnitsNodepoolResourcesResetRequest) (api.ImplResponse, error) {
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	err = ValidateResetResourcesRequest(ctx, *tenantId, request.ClusterId, "", []string{request.NodePoolId}, o.clusterExternalClient)
	if err != nil {
		errResponse := HandleError(err, http.StatusBadRequest, err.Error())
		return errResponse, nil
	}

	dbErr := dao.NodepoolResetResources(ctx, *tenantId, request.ClusterId, request.NodePoolId, request.Action)
	if dbErr != nil {
		errResponse := HandleDbError(ctx, dbErr, "Failed to reset node pool resources")
		return errResponse, nil
	}

	return api.Response(http.StatusNoContent, nil), nil
}

func MapProjectsToDeleted(projects []models.Project) []api.DeletedOrgUnit {
	deletedOrgUnits := []api.DeletedOrgUnit{}
	for _, project := range projects {
		deletedOrgUnits = append(deletedOrgUnits, api.DeletedOrgUnit{
			Id:        fmt.Sprint(project.ID),
			Type:      "project",
			Name:      project.Name,
			CreatedAt: project.CreatedAt,
			UpdatedAt: project.UpdatedAt,
			DeletedAt: project.DeletedAt.Time,
		})
	}
	return deletedOrgUnits
}

func MapDepartmentsToDeleted(departments []models.Department) []api.DeletedOrgUnit {
	deletedOrgUnits := []api.DeletedOrgUnit{}
	for _, department := range departments {
		deletedOrgUnits = append(deletedOrgUnits, api.DeletedOrgUnit{
			Id:        fmt.Sprint(department.ID),
			Type:      "department",
			Name:      department.Name,
			CreatedAt: department.CreatedAt,
			UpdatedAt: department.UpdatedAt,
			DeletedAt: department.DeletedAt.Time,
		})
	}
	return deletedOrgUnits
}
