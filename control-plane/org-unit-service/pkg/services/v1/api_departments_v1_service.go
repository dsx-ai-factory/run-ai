/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package v1

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	auth_client "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client"
	openapi "github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/client/generated"
	cluster_service_client "github.com/dsx-ai-factory/run-ai/control-plane/cluster-service/pkg/client"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/org-unit-service/pkg/services"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
)

const (
	DepartmentAdminRole       = "Department administrator"
	ResearchManagerRole       = "Research manager"
	ResearcherRole            = "L1 researcher"
	MlEngineerRole            = "ML engineer"
	L2ResearcherRole          = "L2 researcher"
	EditorRole                = "Editor"
	SystemAdminRole           = "System administrator"
	ApplicationAdminRole      = "Application administrator"
	DepartmentAuditEntityType = "department"
)

type DepartmentsV1ApiService struct {
	authorizationClient auth_client.IAuthorizationClient
	clustersClient      cluster_service_client.IClusterExternalClient
}

func NewDepartmentV1ApiService(clustersApiService cluster_service_client.IClusterExternalClient) api.DepartmentsV1APIServicer {
	return &DepartmentsV1ApiService{
		authorizationClient: auth_client.NewAuthorizationClient(),
		clustersClient:      clustersApiService,
	}
}

func (s *DepartmentsV1ApiService) GetDepartmentsV1(ctx context.Context, clusterId string, excludePermissions bool, memoryUnitMb bool) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Get V1 departments")
	tenantId, clusterUuid, err := GetTenantAndClusterUuid(ctx, clusterId)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant and cluster id"), nil
	}

	permittedScopes, err := auth_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	departments, err := dao.GetDepartmentsByCluster(ctx, *tenantId, clusterUuid, permittedScopes, models.VerboseVerbosity)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to get department")
		return errResponse, nil
	}

	departmentAdmins := make(map[string]*[]openapi.AccessRule)
	if !excludePermissions {
		departmentIds := []string{}
		for _, department := range *departments {
			departmentIds = append(departmentIds, strconv.FormatInt(department.ID, 10))
		}
		accessRules, err := s.getDepartmentAdmins(ctx, strings.Join(departmentIds, ","), clusterId)
		if err != nil {
			errResponse := services.HandleError(err, http.StatusServiceUnavailable, "Failed to get department admins")
			return errResponse, nil
		}
		for _, accessRule := range *accessRules {
			scopeType := accessRule.ScopeType
			if scopeType == openapi.SCOPETYPE_DEPARTMENT {
				departmentAdminsPerScope, ok := departmentAdmins[accessRule.ScopeId]
				if !ok || departmentAdminsPerScope == nil {
					departmentAdminsPerScope = &[]openapi.AccessRule{}
					departmentAdmins[accessRule.ScopeId] = departmentAdminsPerScope
				}
				*departmentAdminsPerScope = append(*departmentAdminsPerScope, accessRule)
			} else {
				for _, department := range *departments {
					departmentAdminsPerScope, ok := departmentAdmins[strconv.FormatInt(department.ID, 10)]
					if !ok || departmentAdminsPerScope == nil {
						departmentAdminsPerScope = &[]openapi.AccessRule{}
						departmentAdmins[accessRule.ScopeId] = departmentAdminsPerScope
					}
					*departmentAdminsPerScope = append(*departmentAdminsPerScope, accessRule)
				}
			}
		}
	}

	nodePools, err := services.GetNodePoolsMetadataMap(ctx, s.clustersClient, clusterId)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusInternalServerError, "Failed to retrieve node pools")
		return errResponse, nil
	}

	response, err := MapDepartmentsModelToV1Response(ctx, *departments, departmentAdmins, memoryUnitMb, nodePools)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusInternalServerError, "Failed to map department to response")
		return errResponse, nil
	}
	return api.Response(http.StatusOK, response), nil
}

func (s *DepartmentsV1ApiService) getDepartmentAdmins(ctx context.Context, departmentIds string, clusterId string) (*[]openapi.AccessRule, error) {
	getAccessRules200Response, response, err := s.authorizationClient.GetScopeAccessRules(ctx).
		DepartmentIds(departmentIds).ClusterId(clusterId).
		Roles(DepartmentAdminRole).Execute()
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		err = fmt.Errorf("failed to get department admins. status is %s", response.Status)
		return nil, err
	}
	accessRules, exists := getAccessRules200Response.GetAccessRulesOk()
	if !exists {
		return &[]openapi.AccessRule{}, nil
	}
	return &accessRules, nil
}

func (s *DepartmentsV1ApiService) SetDepartmentAdminsV1(ctx context.Context, clusterId string, departmentId int32, departmentV1AccessControl api.DepartmentV1AccessControl) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Set V1 department admins")
	departmentV1AccessControl.DepartmentId = departmentId
	succeeded, failed, err := s.AddDepartmentAdmins(ctx, clusterId, departmentId, departmentV1AccessControl)
	if err != nil {
		return services.HandleError(err, http.StatusServiceUnavailable, "Failed to add department admins"), nil
	}
	if len(failed) > 0 {
		return services.HandleError(err, http.StatusServiceUnavailable, fmt.Sprintf("Failed to add department admins %v", failed)), nil
	}
	response := api.DepartmentV1AccessControl{
		DepartmentId:     departmentId,
		DepartmentAdmins: succeeded,
	}
	return api.Response(http.StatusCreated, response), nil
}

func (s *DepartmentsV1ApiService) GetDepartmentMetricsV1(ctx context.Context, clusterId string, departmentId string, start time.Time, end time.Time, numberOfSamples int32, nodepoolName string) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Get V1 department metrics")
	err := ValidateMetricsRequest(ctx, clusterId, start, end, &numberOfSamples, s.clustersClient)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, "metrics request validation failed")
		return errResponse, nil
	}

	tenantId, clusterUuid, err := GetTenantAndClusterUuid(ctx, clusterId)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant and cluster id"), nil
	}

	id, err := strconv.Atoi(departmentId)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse department id"), nil
	}
	department, err := dao.GetDepartmentByIdAndCluster(ctx, int64(id), *tenantId, clusterUuid, models.VerboseVerbosity)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to get department")
		return errResponse, nil
	}
	departments := []string{department.Name}
	projects := []string{".*"}
	nodePools := []string{".*"}
	if nodepoolName != "" {
		nodePools = strings.Split(nodepoolName, ",")
	}
	departmentCurrentMetrics, projectCurrentMetrics, err := GetCurrentMetrics(ctx, clusterId, nodePools, departments, projects, false)
	if err != nil {
		return services.HandleError(err, http.StatusServiceUnavailable, "Failed to get metrics"), nil
	}

	timeRangeMetrics, err := GetTimeRangeMetrics(ctx, clusterId, nodePools, departments, nil, start, end, numberOfSamples, false)
	if err != nil {
		return services.HandleError(err, http.StatusServiceUnavailable, "Failed to get metrics"), nil
	}

	clusterName, err := services.GetClusterName(ctx, clusterId, s.clustersClient)
	if err != nil {
		return services.HandleError(err, http.StatusServiceUnavailable, "Failed to get cluster name"), nil
	}
	response, err := MapDepartmentMetricsToResponse(ctx, clusterId, clusterName, department, departmentCurrentMetrics, projectCurrentMetrics, timeRangeMetrics)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to map department metrics response. err: %v", err)
		return services.HandleError(err, http.StatusServiceUnavailable, "Failed to construct metrics response"), nil
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *DepartmentsV1ApiService) GetDepartmentsMetricsV1(ctx context.Context, clusterId string, start time.Time, end time.Time, numberOfSamples int32, nodepoolName string) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("Get V1 departments metrics")
	err := ValidateMetricsRequest(ctx, clusterId, start, end, &numberOfSamples, s.clustersClient)
	if err != nil {
		errResponse := services.HandleError(err, http.StatusBadRequest, "metrics request validation failed")
		return errResponse, nil
	}
	tenantId, clusterUuid, err := GetTenantAndClusterUuid(ctx, clusterId)
	if err != nil {
		return services.HandleError(err, http.StatusBadRequest, "Failed to parse tenant and cluster id"), nil
	}

	permittedScopes, err := auth_client.ParsePermittedScopes(ctx)
	if err != nil {
		permittedScopes = nil
	}

	departmentModels, err := dao.GetDepartmentsByCluster(ctx, *tenantId, clusterUuid, permittedScopes, models.VerboseVerbosity)
	if err != nil {
		errResponse := services.HandleDbError(ctx, err, "Failed to get department")
		return errResponse, nil
	}
	if departmentModels == nil || len(*departmentModels) == 0 {
		return api.Response(http.StatusOK, api.DepartmentsMetricsV1Response{
			Data: []api.DepartmentsMetricsV1ResponseDataInner{},
		}), nil
	}
	departments := []string{}
	for _, departmentModel := range *departmentModels {
		departments = append(departments, departmentModel.Name)
	}
	nodePools := []string{".*"}
	if nodepoolName != "" {
		nodePools = strings.Split(nodepoolName, ",")
	}
	departmentCurrentMetrics, projectCurrentMetrics, err := GetCurrentMetrics(ctx, clusterId, nodePools, departments, []string{".*"}, false)
	if err != nil {
		return services.HandleError(err, http.StatusServiceUnavailable, "Failed to get metrics"), nil
	}

	fetchUtilization := ShouldFetchUtilization(permittedScopes)
	timeRangeMetrics, err := GetTimeRangeMetrics(ctx, clusterId, nodePools, departments, nil, start, end, numberOfSamples, fetchUtilization)
	if err != nil {
		return services.HandleError(err, http.StatusServiceUnavailable, "Failed to get metrics"), nil
	}

	clusterName, err := services.GetClusterName(ctx, clusterId, s.clustersClient)
	if err != nil {
		return services.HandleError(err, http.StatusServiceUnavailable, "Failed to get cluster name"), nil
	}
	response, err := MapDepartmentsMetricsToResponse(ctx, clusterId, clusterName, *departmentModels, departmentCurrentMetrics, projectCurrentMetrics, timeRangeMetrics, fetchUtilization)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("Failed to map department metrics response. err: %v", err)
		return services.HandleError(err, http.StatusServiceUnavailable, "Failed to construct metrics response"), nil
	}

	return api.Response(http.StatusOK, response), nil
}

func (s *DepartmentsV1ApiService) AddDepartmentAdmins(ctx context.Context, clusterId string, departmentId int32, request api.DepartmentV1AccessControl) ([]string, []string, error) {
	succeeded := []string{}
	failed := []string{}

	role, err := s.getDepartmentRole(ctx)
	if err != nil {
		return succeeded, failed, err
	}
	for _, user := range request.DepartmentAdmins {
		accessRuleCreateRequest := s.authorizationClient.CreateAccessRule(ctx).AccessRuleCreationFields(openapi.AccessRuleCreationFields{
			SubjectId:   user,
			SubjectType: openapi.SUBJECTTYPE_USER,
			RoleId:      role.Id,
			ScopeId:     strconv.Itoa(int(departmentId)),
			ScopeType:   openapi.SCOPETYPE_DEPARTMENT,
			ClusterId:   &clusterId,
		})
		logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("creating department admin for department %d, user %s", departmentId, user)
		_, _, err := s.authorizationClient.CreateAccessRuleExecute(accessRuleCreateRequest)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to set department admin %s. Error: %v", user, err)
			failed = append(failed, user)
		} else {
			succeeded = append(succeeded, user)
		}
	}
	return succeeded, failed, nil
}

func (s *DepartmentsV1ApiService) getDepartmentRole(ctx context.Context) (*openapi.RoleV1, error) {
	getRoles := s.authorizationClient.GetRolesV1(ctx)
	roleExecute, response, err := s.authorizationClient.GetRolesV1Execute(getRoles)
	if err != nil || (response == nil || response.StatusCode != http.StatusOK) {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to set permissions. Error: %v", err)
		return nil, err
	}
	var departmentAdminRole openapi.RoleV1
	for _, role := range roleExecute {
		if role.Name == DepartmentAdminRole {
			departmentAdminRole = role
		}
	}
	return &departmentAdminRole, nil
}
