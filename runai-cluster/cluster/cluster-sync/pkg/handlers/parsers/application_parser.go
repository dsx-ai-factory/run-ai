/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package parsers

import (
	"fmt"
	"math"

	swagger "github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/client/generated"
	global_consts "github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/common/constants"
	"github.com/dsx-ai-factory/run-ai/runai-cluster/cluster/sdk/apis/run/v1alpha1"
	"k8s.io/utils/ptr"
)

func ParseBackendApplication(application *v1alpha1.Application, isDelete bool) (swagger.AIApplicationForSync, error) {
	applicationId := GetApplicationId(application)

	departmentId, ok := application.Labels[global_consts.DepartmentIdLabel]
	if !ok || departmentId == "" {
		return swagger.AIApplicationForSync{}, fmt.Errorf("failed to get departmentId for application %s", application.Name)
	}
	// departmentName is best effort
	departmentName := application.Labels[global_consts.DepartmentNameLabel]
	projectId, ok := application.Labels[global_consts.ProjectIdLabel]
	if !ok || projectId == "" {
		return swagger.AIApplicationForSync{}, fmt.Errorf("failed to get projectId for application %s", application.Name)
	}
	projectName, ok := application.Labels[global_consts.ProjectNameLabel]
	if !ok || projectName == "" {
		return swagger.AIApplicationForSync{}, fmt.Errorf("failed to get projectName for application %s", application.Name)
	}
	action := swagger.AIAPPLICATIONSYNCACTION_UPSERT
	if isDelete {
		action = swagger.AIAPPLICATIONSYNCACTION_DELETE
	}
	resources := []swagger.AIApplicationSyncResource{}
	for _, resource := range application.Status.Resources {
		// we send the id as the controlPlaneId and clusterId as id since it's ControlPlane model, same for name
		backendResource := swagger.AIApplicationSyncResource{
			Id:          ptr.To(resource.ControlPlaneId),
			ClusterId:   ptr.To(resource.UID),
			Name:        ptr.To(resource.ControlPlaneName),
			ClusterName: ptr.To(resource.Name),
			Kind:        ptr.To(resource.Kind),
		}
		resources = append(resources, backendResource)
	}

	actionInfoFound := false
	actionInfo := swagger.AIApplicationSyncActionInfo{}

	if application.Status.ActionStatus != nil {
		actionInfoFound = true

		actionInfo.Type = nonEmptyPtr(application.Status.ActionStatus.Type)
		actionInfo.Status = nonEmptyPtr(application.Status.ActionStatus.Phase)
		actionInfo.Message = nonEmptyPtr(application.Status.ActionStatus.Message)
		actionInfo.Error = nonEmptyPtr(application.Status.ActionStatus.Error)

		if !application.Status.ActionStatus.StartedAt.IsZero() {
			actionInfo.Started = nonEmptyPtr(application.Status.ActionStatus.StartedAt.String())
		}

		if !application.Status.ActionStatus.CompletedAt.IsZero() {
			actionInfo.Completed = nonEmptyPtr(application.Status.ActionStatus.CompletedAt.String())
		}
	}

	if application.Status.HelmStatus != nil {
		actionInfoFound = true

		actionInfo.HelmPhase = nonEmptyPtr(application.Status.HelmStatus.Phase)
		actionInfo.HelmError = nonEmptyPtr(application.Status.HelmStatus.Error)
		actionInfo.ChartUrl = nonEmptyPtr(application.Status.HelmStatus.ChartUrl)
		if application.Status.HelmStatus.ReleaseVersion != nil {
			if v := *application.Status.HelmStatus.ReleaseVersion; v >= math.MinInt32 && v <= math.MaxInt32 {
				actionInfo.ReleaseVersion = ptr.To(int32(v))
			}
		}
	}

	backendApplication := swagger.AIApplicationForSync{
		Id:             &applicationId,
		Name:           &application.Name,
		Source:         &application.Spec.Source,
		Type:           &application.Spec.Type,
		K8sPhase:       &application.Status.K8sPhase,
		DepartmentId:   &departmentId,
		DepartmentName: nonEmptyPtr(departmentName),
		ProjectId:      &projectId,
		ProjectName:    &projectName,
		Resources:      resources,
		Action:         &action,
	}

	if actionInfoFound {
		backendApplication.ActionInfo = *swagger.NewNullableAIApplicationSyncActionInfo(&actionInfo)
	}

	return backendApplication, nil
}

func GetApplicationId(application *v1alpha1.Application) string {
	// if the application is created in the cluster, use the cluster UID as application identifier
	// if it is run-ai managed application, need to use the CP application ID
	if cpApplicationId, ok := application.Annotations[global_consts.ControlPlaneApplicationIdAnnotation]; ok {
		return cpApplicationId
	}
	return string(application.UID)
}

func nonEmptyPtr(str string) *string {
	if str != "" {
		return &str
	}
	return nil
}
