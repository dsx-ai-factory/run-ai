/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"fmt"

	humanize "github.com/dustin/go-humanize"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/internal/generated/pkg/api"
	"github.com/dsx-ai-factory/run-ai/control-plane/workloads/pkg/database/models"
	"github.com/dsx-ai-factory/run-ai/runai-application-packages/convutils"
	"k8s.io/utils/ptr"
)

// Product defined strings:
const (
	NonPreemtibleOverQuotaUserReason = "Non-preemptible workload is over quota"
	OverLimitUserReason              = "The %s is not allowed to go over quota"

	NonPreemtibleOverQuotaUserMessage = "%s quota exceeded for non-preemptible workloads.\nRequested: %s\n%s quota: %s, with %s already allocated for non-preemptible"
	OverLimitUserMessage              = "%s quota exceeded.\nRequested: %s\n%s quota: %s, with %s already allocated"

	GpuResource    ResourceName = "GPU"
	CpuResource    ResourceName = "CPU"
	MemoryResource ResourceName = "CPU memory"
)

// Product defined info:
var (
	resourceToUnitSuffix = map[ResourceName]string{
		GpuResource:    " devices",
		CpuResource:    " cores",
		MemoryResource: " MB",
	}
)

// internal consts and vars:

type ResourceName string

const (
	BytesToMB float64 = 1000 * 1000
)

var (
	allResources = map[ResourceName]bool{
		GpuResource:    true,
		CpuResource:    true,
		MemoryResource: true,
	}
	resourceToUnitAdjustment = map[ResourceName]float64{
		GpuResource:    1,
		CpuResource:    1,
		MemoryResource: BytesToMB,
	}
)

// api to model

// getSchedulingConditionsFromRequest - converts and returns SchedulingConditions and PhaseReason for workload model
func getSchedulingConditionsFromRequest(workloadSchedulingConditions *[]api.SchedulingCondition) (models.SchedulingConditionArray, *string) {
	if workloadSchedulingConditions == nil {
		return models.SchedulingConditionArray{}, nil
	}

	modelSchedulingConditions := convutils.ConvertSliceTypes(*workloadSchedulingConditions, schedulingConditionApiToModel)
	var phaseReason *string = nil
	if len(modelSchedulingConditions) > 0 {
		phaseReason = ptr.To(string(modelSchedulingConditions[0].Reason))
	}

	return modelSchedulingConditions, phaseReason
}

func schedulingConditionApiToModel(schedulingCondition api.SchedulingCondition) models.SchedulingCondition {
	return models.SchedulingCondition{
		NodePool:     schedulingCondition.NodePool,
		Reason:       schedulingCondition.Reason,
		QueueDetails: schedulingCondition.QueueDetails,
	}
}

// model to api

func getPendingSchedulingMessages(schedulingConditions models.SchedulingConditionArray) *[]api.PendingSchedulingMessage {
	if len(schedulingConditions) == 0 {
		return nil
	}

	return ptr.To(convutils.ConvertSliceTypes(schedulingConditions, schedulingConditionModelToApi))
}

func schedulingConditionModelToApi(schedulingCondition models.SchedulingCondition) api.PendingSchedulingMessage {
	result := api.PendingSchedulingMessage{
		NodePool:    schedulingCondition.NodePool,
		PhaseReason: schedulingCondition.Reason,
		Reason:      phaseReasonToUserReason(schedulingCondition.Reason, schedulingCondition.QueueDetails),
	}

	if schedulingCondition.QueueDetails == nil {
		return result
	}

	result.OrgType = &schedulingCondition.QueueDetails.OrgType
	result.UserMessage = getUnschedulableUserMessage(schedulingCondition.Reason, schedulingCondition.QueueDetails.OrgType, schedulingCondition.QueueDetails)
	return result
}

func phaseReasonToUserReason(phaseReason api.PhaseReason, queueDetails *api.SchedulingConditionQueueDetails) string {
	switch phaseReason {
	case api.PHASEREASON_NON_PREEMPTIBLE_OVER_QUOTA:
		return NonPreemtibleOverQuotaUserReason
	case api.PHASEREASON_OVER_LIMIT:
		orgTypeStr := getOrgTypeStrFromQueueDetails(queueDetails, false)
		return fmt.Sprintf(OverLimitUserReason, orgTypeStr)
	default:
		return ""
	}
}

func getUnschedulableUserMessage(ur api.PhaseReason, orgType api.OrgType, queueDetails *api.SchedulingConditionQueueDetails) *string {
	switch ur {
	case api.PHASEREASON_NON_PREEMPTIBLE_OVER_QUOTA:
		return ptr.To(getNonPreemtibleOverQuotaUserMessage(orgType, queueDetails))
	case api.PHASEREASON_OVER_LIMIT:
		return ptr.To(getOverLimitUserMessage(orgType, queueDetails))
	default:
		return nil
	}
}

func getNonPreemtibleOverQuotaUserMessage(orgType api.OrgType, qd *api.SchedulingConditionQueueDetails) string {
	exceedingResourceName := getExceedingResourceName(qd.ExceedingResources)

	requested := getResourceQuantityStrWithUnits(exceedingResourceName, qd.PodGroupRequestedNonPreemptibleResources)
	orgTypeStr := getOrgTypeStr(orgType, true)
	deserved := getResourceQuantityStrWithUnits(exceedingResourceName, qd.QueueDeservedResources)
	allocatedNonPreemptible := getResourceQuantityStrWithUnits(exceedingResourceName, qd.QueueAllocatedNonPreemptibleResources)

	// "%s quota exceeded for non-preemptible workloads.\nRequested: %s\n%s quota: %s, with %s already allocated for non-preemptible"
	return fmt.Sprintf(NonPreemtibleOverQuotaUserMessage,
		string(exceedingResourceName),
		requested,
		orgTypeStr, deserved, allocatedNonPreemptible)
}

func getOverLimitUserMessage(orgType api.OrgType, qd *api.SchedulingConditionQueueDetails) string {
	exceedingResourceName := getExceedingResourceName(qd.ExceedingResources)

	requested := getResourceQuantityStrWithUnits(exceedingResourceName, qd.PodGroupRequestedResources)
	orgTypeStr := getOrgTypeStr(orgType, true)
	limit := getResourceQuantityStrWithUnits(exceedingResourceName, qd.QueueResourceLimits)
	allocated := getResourceQuantityStrWithUnits(exceedingResourceName, qd.QueueAllocatedResources)

	// "%s quota exceeded.\nRequested: %s\n%s quota: %s, with %s already allocated"
	return fmt.Sprintf(OverLimitUserMessage,
		string(exceedingResourceName),
		requested,
		orgTypeStr, limit, allocated)
}

func getExceedingResourceName(exceedingResources api.QueueResources) ResourceName {
	if exceedingResources.Gpu > 0 {
		return GpuResource
	}
	if exceedingResources.Cpu > 0 {
		return CpuResource
	}
	if exceedingResources.Memory > 0 {
		return MemoryResource
	}
	// default
	return GpuResource
}

func getResourceQuantity(resourceName ResourceName, queueResources api.QueueResources) float64 {
	switch resourceName {
	case GpuResource:
		return queueResources.Gpu
	case CpuResource:
		return queueResources.Cpu
	case MemoryResource:
		return queueResources.Memory
	default:
		return 0
	}
}

func getQuantityStrWithUnits(resourceName ResourceName, quantity float64) string {
	if quantity == -1 {
		return "Unlimited"
	}

	if _, ok := allResources[resourceName]; !ok {
		return ""
	}

	unitAdjustment := resourceToUnitAdjustment[resourceName]
	suffix := resourceToUnitSuffix[resourceName]

	quantityStr := humanize.FtoaWithDigits(quantity/unitAdjustment, 3)
	return fmt.Sprintf("%s%s", quantityStr, suffix)
}

func getResourceQuantityStrWithUnits(resourceName ResourceName, queueResources api.QueueResources) string {
	quantity := getResourceQuantity(resourceName, queueResources)
	return getQuantityStrWithUnits(resourceName, quantity)
}

func getOrgTypeStrFromQueueDetails(queueDetails *api.SchedulingConditionQueueDetails, capital bool) string {
	orgType := api.ORGTYPE_PROJECT
	if queueDetails != nil {
		orgType = queueDetails.OrgType
	}
	return getOrgTypeStr(orgType, capital)
}

func getOrgTypeStr(orgType api.OrgType, capital bool) string {
	switch orgType {
	case api.ORGTYPE_PROJECT:
		if capital {
			return "Project"
		}
		return "project"
	case api.ORGTYPE_DEPARTMENT:
		if capital {
			return "Department"
		}
		return "department"
	default:
		return ""
	}
}
