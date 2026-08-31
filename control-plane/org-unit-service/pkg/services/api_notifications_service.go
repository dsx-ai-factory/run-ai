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

	"github.com/google/uuid"
	"github.com/lib/pq"
	cluster_service_client "github.com/run-ai/runai/control-plane/cluster-service/pkg/client"
	"github.com/run-ai/runai/control-plane/org-unit-service/generated/pkg/api"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/dao"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/database/models"
	"github.com/run-ai/runai/control-plane/org-unit-service/pkg/pubsub"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"golang.org/x/exp/slices"

	dao_utils "github.com/run-ai/runai/runai-common-packages/dao-utils"
	"k8s.io/utils/ptr"
)

type NotificationsApiService struct {
	clustersClient cluster_service_client.IClusterClient
}

func NewNotificationsApiService(clustersApiService cluster_service_client.IClusterClient) api.NotificationsAPIServicer {
	notificationService := &NotificationsApiService{
		clustersClient: clustersApiService,
	}

	hardDeletionService := NewHardTenantDeletionService()
	pubsub.Register(notificationService, hardDeletionService)

	return notificationService
}

const (
	CreatedOperation  = "CREATED"
	DeletedOperation  = "DELETED"
	NoLimitMinVersion = "2.11"
)

func (s *NotificationsApiService) NodePoolUpdated(ctx context.Context, default_ bool, nodePool api.NodePool) (api.ImplResponse, error) {
	// Acquire exclusive write lock on cluster - blocks all CreateProject calls for this cluster
	GetClusterLockManager().Lock(nodePool.ClusterId)
	defer GetClusterLockManager().Unlock(nodePool.ClusterId)

	nodePoolId := getNodepoolId(nodePool)
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("NotificationsApiService: node pool id: %s, name: %s was: %s, is default: %t", nodePoolId, nodePool.Name, nodePool.Operation, default_)
	if nodePoolId == "" {
		err := errors.New("nodepool update must include id or uid")
		return HandleError(err, http.StatusBadRequest, "nodepool update must include id or uid"), nil
	}
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return HandleError(err, http.StatusBadRequest, "NotificationsApiService: Failed to parse tenant id"), nil
	}

	clusterId, err := uuid.Parse(nodePool.ClusterId)
	if err != nil {
		return HandleError(err, http.StatusBadRequest, fmt.Sprintf("failed to parse cluster id %s", nodePool.ClusterId)), nil
	}

	if nodePool.Operation != CreatedOperation && nodePool.Operation != DeletedOperation {
		return HandleError(errors.New("validateDeleteNodePoolResources"), http.StatusBadRequest, "validateDeleteNodePoolResources"), nil
	}

	if default_ {
		if nodePool.Operation == DeletedOperation {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("NotificationsApiService: got default node pool deleted. id: %s", nodePoolId)
			return HandleError(err, http.StatusBadRequest, fmt.Sprintf("got default node pool deleted. id: %s", nodePoolId)), nil
		}
		if nodePool.Operation == CreatedOperation {
			err := s.createDefaultDepartment(ctx, clusterId, tenantId, nodePool)
			if err != nil {
				logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("NotificationsApiService: failed to add default department for node pool %s: %v", nodePoolId, err.Error())
				return HandleError(err, http.StatusInternalServerError, fmt.Sprintf("failed to add default department for node pool %s", nodePoolId)), nil
			}
		}
	} else {
		unlimitedDeserved := s.setUnlimitedDeserved(ctx, clusterId, tenantId)
		err = handleNonDefaultNodePool(ctx, nodePool, tenantId, clusterId, unlimitedDeserved)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("NotificationsApiService: failed to add resources for node pool %s: %v", nodePoolId, err.Error())
			return HandleError(err, http.StatusInternalServerError, fmt.Sprintf("failed to add resources for node pool %s", nodePoolId)), nil
		}
	}
	dao.InvalidateAllCache(ctx, *tenantId, clusterId.String())

	return api.Response(http.StatusOK, nil), nil
}

func (s *NotificationsApiService) setUnlimitedDeserved(ctx context.Context, clusterId uuid.UUID, tenantId *string) bool {
	unlimitedGpuDepartmentExists, err := dao.HasUnlimitedGpuDepartment(ctx, clusterId.String(), *tenantId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("NotificationsApiService: failed to get departments. setting department deserved to 0. error: %v", err)
		return false
	}
	return unlimitedGpuDepartmentExists
}

func (s *NotificationsApiService) ClusterDeleted(ctx context.Context, clusterUuid string) (api.ImplResponse, error) {
	logging.WithCtx(runaictx.LoggerHook(ctx)).Infof("NotificationsApiService: NotificationsApiService: cluster id: %s, deleted", clusterUuid)
	tenantId, err := ParseTenantId(ctx)
	if err != nil {
		return HandleError(err, http.StatusBadRequest, "Failed to parse tenant id"), nil
	}

	clusterId, err := uuid.Parse(clusterUuid)
	if err != nil {
		return HandleError(err, http.StatusBadRequest, fmt.Sprintf("failed to parse cluster id %s", clusterUuid)), nil
	}

	var errorMsg string
	err = dao.DeleteClusterProjects(ctx, *tenantId, clusterId)
	if err != nil {
		errorMsg = fmt.Sprintf("NotificationsApiService: failed to delete projects for cluster %s: %v", clusterUuid, err.Error())
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", errorMsg)
	}

	err = dao.DeleteClusterDepartments(ctx, *tenantId, clusterId)
	if err != nil {
		errorMsg = fmt.Sprintf("NotificationsApiService: failed to delete departments for cluster %s: %v", clusterUuid, err.Error())
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", errorMsg)
	}

	err = dao.DeleteClusterNodeTypes(ctx, *tenantId, clusterId)
	if err != nil {
		errorMsg = fmt.Sprintf("NotificationsApiService: failed to delete nodetypes for cluster %s: %v", clusterUuid, err.Error())
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("%s", errorMsg)
	}

	dao.InvalidateAllCache(ctx, *tenantId, clusterId.String())

	if err != nil {
		return HandleError(err, http.StatusInternalServerError, errorMsg), nil
	}

	return api.Response(http.StatusOK, nil), nil
}

func (s *NotificationsApiService) createDefaultDepartment(ctx context.Context, clusterId uuid.UUID, tenantId *string, nodePool api.NodePool) error {
	defaultResources := createDefaultNodePoolResources(nodePool.Name, getNodepoolId(nodePool))
	createdTime := time.Now()
	department := models.Department{
		IntBaseModel: models.IntBaseModel{
			CreatedAt: createdTime,
			UpdatedAt: createdTime,
		},
		TenantId:    *tenantId,
		ClusterId:   clusterId,
		Name:        DefaultDepartmentName,
		OrgUnitData: models.OrgUnitData{Resources: []models.Resources{*defaultResources}},
	}
	_, err := dao.CreateDepartment(ctx, &department)
	return err
}

func getNodepoolId(nodePool api.NodePool) string {
	if nodePool.Id != "" {
		return nodePool.Id
	}
	return strconv.FormatFloat(float64(nodePool.Uid), 'f', -1, 32)
}

func handleNonDefaultNodePool(ctx context.Context, nodePool api.NodePool, tenantId *string, clusterId uuid.UUID, unlimitedDeserved bool) error {
	departments, err := dao.GetDepartmentsByCluster(ctx, *tenantId, clusterId, nil, models.SlimVerbosity)
	if err != nil {
		return err
	}
	projects, err := dao.GetProjectsByCluster(ctx, *tenantId, clusterId, nil, models.SlimVerbosity)
	if err != nil {
		return err
	}
	if nodePool.Operation == DeletedOperation {
		err = removeDeletedNodePoolFromDefault(ctx, nodePool, departments, projects)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("NotificationsApiService: Failed to delete nodepool %s from default lists. err: %s", nodePool.Name, err.Error())
		}
		return errors.Join(err, removeDeletedNodePoolResources(ctx, nodePool, departments, projects))
	}
	if nodePool.Operation == CreatedOperation {
		return addCreatedNodePoolResources(ctx, nodePool, departments, projects, unlimitedDeserved)
	}
	return nil
}

func removeDeletedNodePoolFromDefault(ctx context.Context, nodePool api.NodePool, departments *[]models.Department, projects []*models.Project) error {
	emptyDepartmentNodepoolLists := map[uuid.UUID]bool{}
	orgUnitDataEntities := []models.OrgUnitData{}
	for _, department := range *departments {
		if department.OrgUnitData.DefaultNodePoolList != nil &&
			slices.Contains(*department.OrgUnitData.DefaultNodePoolList, nodePool.Name) {
			department.OrgUnitData.DefaultNodePoolList = removeNodepoolFromList(department.OrgUnitData.DefaultNodePoolList, nodePool.Name)
			if department.OrgUnitData.DefaultNodePoolList == nil {
				emptyDepartmentNodepoolLists[department.OrgUnitData.ID] = true
			}

			orgUnitDataEntities = append(orgUnitDataEntities, department.OrgUnitData)
		}
	}
	for _, project := range projects {
		if project.OrgUnitData.DefaultNodePoolList != nil &&
			(slices.Contains(*project.OrgUnitData.DefaultNodePoolList, nodePool.Name) ||
				slices.Contains(*project.OrgUnitData.DefaultNodePoolListEffective, nodePool.Name)) {
			project.OrgUnitData.DefaultNodePoolList = removeNodepoolFromList(project.OrgUnitData.DefaultNodePoolList, nodePool.Name)
			project.OrgUnitData.DefaultNodePoolListEffective = removeNodepoolFromList(project.OrgUnitData.DefaultNodePoolListEffective, nodePool.Name)
			if project.OrgUnitData.DefaultNodePoolListEffectiveOrigin != nil {
				if found, ok := emptyDepartmentNodepoolLists[*project.OrgUnitData.DefaultNodePoolListEffectiveOrigin]; found && ok {
					project.OrgUnitData.DefaultNodePoolListEffectiveOrigin = nil
				}
			}

			orgUnitDataEntities = append(orgUnitDataEntities, project.OrgUnitData)
		}
	}

	return dao.SaveOrgUnitDataDefaultNodepoolLists(ctx, orgUnitDataEntities)
}

func removeNodepoolFromList(list *pq.StringArray, name string) *pq.StringArray {
	list = ptr.To(slices.DeleteFunc(*list, func(s string) bool {
		return s == name
	}))
	if len(*list) == 0 {
		return nil
	}

	return list
}

func removeDeletedNodePoolResources(ctx context.Context, nodePool api.NodePool, departments *[]models.Department, projects []*models.Project) error {
	nodePoolId := getNodepoolId(nodePool)
	resourcesToDelete := []models.Resources{}
	for _, department := range *departments {
		toDelete, err := getResourcesToDelete(department.OrgUnitData, nodePoolId)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("NotificationsApiService: Node pool %s deleted, resource for department %d does not exists", nodePoolId, department.ID)
		} else {
			resourcesToDelete = append(resourcesToDelete, toDelete)
		}
	}
	for _, project := range projects {
		toDelete, err := getResourcesToDelete(project.OrgUnitData, nodePoolId)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("NotificationsApiService: Node pool %s deleted, resource for project %d does not exists", nodePool.Id, project.ID)
		} else {
			resourcesToDelete = append(resourcesToDelete, toDelete)
		}
	}
	return dao.DeleteResources(ctx, resourcesToDelete)
}

func validateNodePoolCreation(orgUnitData models.OrgUnitData, nodePoolId string) error {
	for _, resources := range orgUnitData.Resources {
		if resources.NodePoolId == nodePoolId {
			return errors.New("nodepool resource already exists")
		}
	}
	return nil
}

func addCreatedNodePoolResources(ctx context.Context, nodePool api.NodePool, departments *[]models.Department, projects []*models.Project, unlimitedDeserved bool) error {
	nodePoolId := getNodepoolId(nodePool)
	resourcesToAdd := []*models.Resources{}
	for _, department := range *departments {
		err := validateNodePoolCreation(department.OrgUnitData, nodePoolId)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("NotificationsApiService: Department %d already has nodepool %s resources", department.ID, nodePoolId)
		} else {
			resourcesToAdd = append(resourcesToAdd, GetDepartmentResourcesToAdd(department.OrgUnitData,
				unlimitedDeserved && department.Name == DefaultDepartmentName, nodePoolId, nodePool.Name))
		}
	}
	for _, project := range projects {
		err := validateNodePoolCreation(project.OrgUnitData, nodePoolId)
		if err != nil {
			logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("NotificationsApiService: Project %d already has nodepool %s resources", project.ID, nodePoolId)
		} else {
			resourcesToAdd = append(resourcesToAdd, GetProjectResourcesToAdd(project.OrgUnitData, nodePoolId, nodePool.Name))
		}
	}
	return dao.AddResources(ctx, resourcesToAdd)
}

func getResourcesToDelete(orgUnitData models.OrgUnitData, nodePoolId string) (models.Resources, error) {
	for _, resources := range orgUnitData.Resources {
		if resources.NodePoolId == nodePoolId {
			return resources, nil
		}
	}
	return models.Resources{}, errors.New("resource for nodepool not found")
}

func GetProjectResourcesToAdd(orgUnitData models.OrgUnitData, nodePoolId string, nodePoolName string) *models.Resources {
	return getResourcesToAdd(orgUnitData, nodePoolId, nodePoolName, 0, 2)
}

func GetDepartmentResourcesToAdd(orgUnitData models.OrgUnitData, setUnlimitedDeserved bool, nodePoolId string, nodePoolName string) *models.Resources {
	if setUnlimitedDeserved {
		return getResourcesToAdd(orgUnitData, nodePoolId, nodePoolName, -1, -1)
	} else {
		return getResourcesToAdd(orgUnitData, nodePoolId, nodePoolName, 0, 0)
	}
}

func getResourcesToAdd(orgUnitData models.OrgUnitData, nodePoolId string, nodePoolName string, deserved float64, overQuotaWeight float64) *models.Resources {
	return &models.Resources{
		BaseModel: dao_utils.BaseModel{
			CreatedAt: time.Now(),
		},
		NodePoolId:            nodePoolId,
		NodePoolName:          nodePoolName,
		GpuDeserved:           ptr.To(deserved),
		GpuLimit:              ptr.To(float64(-1)),
		GpuOverQuotaWeight:    ptr.To(overQuotaWeight),
		CpuDeserved:           nil,
		CpuLimit:              nil,
		CpuOverQuotaWeight:    nil,
		MemoryDeserved:        nil,
		MemoryLimit:           nil,
		MemoryOverQuotaWeight: nil,
		OrgUnitDataID:         orgUnitData.ID,
	}
}

func createDefaultNodePoolResources(nodePoolName string, nodePoolId string) *models.Resources {
	return &models.Resources{
		BaseModel: dao_utils.BaseModel{
			CreatedAt: time.Now(),
		},
		NodePoolId:         nodePoolId,
		NodePoolName:       nodePoolName,
		GpuDeserved:        ptr.To(float64(-1)),
		GpuLimit:           ptr.To(float64(-1)),
		GpuOverQuotaWeight: ptr.To(float64(-1)),
	}
}
