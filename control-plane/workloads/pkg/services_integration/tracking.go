/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services_integration

import (
	"context"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/datastructs"
	runai_config "github.com/dsx-ai-factory/run-ai/runai-common-packages/config"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/dao-utils/validators"
	"time"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"

	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/dao"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/services_utils"
)

func GetWorkloadDetails(ctx context.Context, workloadId string) (*models.WorkloadView, *api.Error) {
	return dao.GetWorkloadViewFromDb(ctx, workloadId)
}

func GetWorkloadsDetails(ctx context.Context, workloadIds []string, excludeDeleted *bool) ([]*models.WorkloadView, *api.Error) {
	return dao.GetWorkloadsViewByIdsFromDb(ctx, workloadIds, excludeDeleted)
}

func GetPhaseOfWorkload(ctx context.Context, clusterId *string, workloadId string) (api.Phase, *api.Error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return api.PHASE_FAILED, apiErr
	}

	return dao.NewWorkloadsDao(runai_config.Get[*config.ServerConfig]()).DbGetWorkloadPhase(ctx, tenantId, clusterId, workloadId)
}

func UpdateWorkloadFailure(ctx context.Context, workloadId string, message string, updateTime time.Time) *api.Error {
	return services.PatchWorkloadPhaseHelper(ctx, workloadId, api.WorkloadPatchFields{
		Phase:        api.PHASE_FAILED,
		PhaseMessage: message,
	}, updateTime)
}

func UpdateDesiredPhaseOfWorkload(ctx context.Context, workloadId string, desiredPhase api.WorkloadDesiredPhase, desiredPhaseUpdateTime time.Time) *api.Error {
	return services.PatchWorkloadPhaseHelper(ctx, workloadId, api.WorkloadPatchFields{
		Phase: getPhaseFromDesired(desiredPhase),
	}, desiredPhaseUpdateTime)
}

func getPhaseFromDesired(desiredPhase api.WorkloadDesiredPhase) api.Phase {
	switch desiredPhase {
	case api.WORKLOADDESIREDPHASE_DELETED:
		return api.PHASE_DELETING
	case api.WORKLOADDESIREDPHASE_STOPPED:
		return api.PHASE_STOPPING
	case api.WORKLOADDESIREDPHASE_RUNNING:
		return api.PHASE_RESUMING
	default:
		return api.Phase(desiredPhase)
	}
}

func WorkloadCreated(ctx context.Context, creationFields api.WorkloadCreationFields) *api.Error {
	workloadsService := services.NewWorkloadsService().(*services.WorkloadsService)

	_, err := workloadsService.InternalCreateWorkloadHelper(ctx, creationFields)

	return err
}

func UpdateWorkloadConnectionMetadata(ctx context.Context, workloadId string, environments []api.Environment, externalConnections []api.Connection) *api.Error {
	workloadsService := services.NewWorkloadsService().(*services.WorkloadsService)

	return workloadsService.UpdateWorkloadConnectionMetadata(ctx, workloadId, services.WorkloadConnectionMetadata{
		Environments:        environments,
		ExternalConnections: externalConnections,
	})
}

func EnsureWorkloadConnectionMetadata(ctx context.Context, creationFields api.WorkloadCreationFields, environments []api.Environment, externalConnections []api.Connection) *api.Error {
	workloadsService := services.NewWorkloadsService().(*services.WorkloadsService)

	return workloadsService.EnsureWorkloadConnectionMetadata(ctx, creationFields, services.WorkloadConnectionMetadata{
		Environments:        environments,
		ExternalConnections: externalConnections,
	})
}

func IsWorkloadNameInUse(ctx context.Context, clusterId string, projectId string, workloadType string, name string) (bool, *api.Error) {
	tenantId, apiErr := services_utils.ParseTenantId(ctx)
	if apiErr != nil {
		return false, apiErr
	}

	count, err := dao.NewWorkloadsDao(runai_config.Get[*config.ServerConfig]()).DbCountWorkloads(ctx, tenantId, nil, false, []validators.ParameterizedFilter{
		{
			Field:           "cluster_id",
			Operator:        "=",
			FormattedValues: []string{clusterId},
		},
		{
			Field:           "project_id",
			Operator:        "=",
			FormattedValues: []string{projectId},
		},
		{
			Field:           "name",
			Operator:        "=",
			FormattedValues: []string{name},
		},
	}, "", true, true)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// TODO call this function on sync requests as well
func WorkloadDeleted(ctx context.Context, workloadId string) *api.Error {
	return nil
}

// DbDeleteWorkload deletes a workload from the database
func DbDeleteWorkload(ctx context.Context, workloadId string) (api.ImplResponse, error) {
	workloadsService := services.NewWorkloadsService().(*services.WorkloadsService)
	return workloadsService.DeleteWorkload(ctx, workloadId)
}

func GetDBCategoriesNames(ctx context.Context, tenantId uint) (datastructs.Set[string], *api.Error) {
	categories, getErr := dao.DbListCategories(ctx, tenantId)
	if getErr != nil {
		return nil, getErr
	}

	return toNames(categories, func(c api.WorkloadCategory) string { return c.Name }), nil
}

func GetClusterPriorityNamesAndValues(ctx context.Context, tenantIdStr, clusterId string) (datastructs.Set[string], map[string]int32, error) {
	priorities, err := dao.GetClusterPriorities(ctx, tenantIdStr, clusterId, "", "", []validators.ParameterizedFilter{})
	if err != nil {
		return nil, nil, err
	}

	names := toNames(*priorities, func(p models.PriorityClusterView) string { return p.Name })
	values := make(map[string]int32, len(*priorities))
	for _, p := range *priorities {
		values[p.Name] = int32(p.Value)
	}
	return names, values, nil
}

func toNames[T any](items []T, getName func(T) string) datastructs.Set[string] {
	names := datastructs.NewSet[string]()
	for _, item := range items {
		names.Add(getName(item))
	}
	return names
}
