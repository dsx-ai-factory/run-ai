/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type {
  RankingMetricType,
  RankingOrderType,
  ResourcesNullableGpu,
  ResourcesPerNodePoolNullable,
  BatchPatchNodePoolResourcesRequest,
  NodePoolResourcesPatchRequest,
} from "@/swagger-models/org-unit-service-client";
import {
  type Department,
  type DepartmentCreationRequest,
  DepartmentFilterSortFields,
  type DepartmentUpdateRequest,
  type OvertimeData,
  type Project,
  type ProjectCreationRequest,
  ProjectFilterSortFields,
  ProjectPhase,
  type ProjectStatus,
  type ProjectUpdateRequest,
  type Resources,
  type ResourcesCpu,
  type ResourcesGpu,
  type ResourcesMemory,
  type ResourcesNodePool,
  type SchedulingRules,
  type TotalResources,
  Units,
} from "@/swagger-models/org-unit-service-client";
import {
  EOverQuotaPriority,
  OVER_QUOTA_DISABLED_VALUE,
  OVER_QUOTA_ENABLED_VALUE,
  RESOURCE_MAX_ALLOWED_INFINITE_VALUE,
  RESOURCE_MAX_ALLOWED_INFINITE_VALUE_NUMERIC,
} from "@/models/resource.model";
import type { INodePoolsNameAndId } from "@/models/node-pool.model";
import type { IProgressColOptions, IStatusColOptions } from "@/models/table.model";
import {
  buildNotReadyTooltip,
  EProjectConditionDeletionReason,
  EProjectConditionNotReadyReason,
  type IProjectTable,
  type IProjectTableFilterBy,
  PROJECT_CONDITION_NOT_READY_REASON_TOOLTIP_MAP,
  ProjectPhaseMap,
} from "@/models/project.model";
import { tableUtil } from "@/utils/table.util";
import type { EEntityPriorityLabel } from "@/models/org-unit.model";
import {
  ENTITY_PRIORITY_DEFAULT_VALUE,
  entityPriorityMap,
  type EOrgUnitRankingMetricValue,
} from "@/models/org-unit.model";
import {
  EOrgUnitOverTimeColumnName,
  EOrgUnitOverTimeValue,
  type IEmptyDepartmentModelConfig,
  type IEmptyProjectModelConfig,
  rankingMetricTypeMap,
  rankingOrderMap,
} from "@/models/org-unit.model";
import { toPercent } from "@/utils/format.util";
import { CLUSTER_COLUMN_FILTER_NAME, type IPaginationFilter, type IFilterBy } from "@/models/filter.model";
import { filterService } from "@/services/filter.service/filter.service";
import type { IDepartmentTableFilterBy } from "@/models/department.model";
import { EGpuResourceState } from "@/models/org-unit-forms.model";

export const orgUnitUtil = {
  getEmptyProjectModel,
  getEmptyResourcesModel,
  getEmptyDepartmentModel,
  getStatusColOptions,
  getNodePoolNameByResource,
  getNodePoolIdByResource,
  convertMemoryValue,
  removeOverQuotaWeightFields,
  getProjectUpdateRequest,
  getDepartmentUpdateRequest,
  updateLimitOrOverQuota,
  enrichResourcesWithCpuAndMemory,
  initResourcesOverQuotaWeight,
  updateSchedulingRules,
  initSchedulingRules,
  updateOverQuotaWeightToDeserved,
  alignCpuAndMemoryLimitsByGpu,
  ensureLimitsNotBelowDeserved,
  getAverageGpuAllocationDataByTimeframe,
  getAverageGpuUtilizationDataByTimeframe,
  getAverageGpuMemoryUtilizationDataByTimeframe,
  getProjectListFilters,
  getDepartmentListFilters,
  getAllocationRatioDisplayValue,
  mapRankingMetricToMetricTypeAndOrder,
  getProgressColOptions,
  isResourceOverQuotaEnabled,
  isOverQuotaValueValid,
  mapStringToPriorityLabel,
  getNodePoolGpuState,
  aggregateGpuStates,
  buildNodePoolBatchPatchRequest,
  buildNodePoolRankBatchPatchRequest,
  buildNodePoolWeightBatchPatchRequest,
  buildNodePoolOrgUnitChangesBatchPatchRequest,
};

function getEmptyProjectModel(config: IEmptyProjectModelConfig): ProjectCreationRequest {
  const emptyResources: Resources[] = getEmptyResourcesModel(
    config.nodePoolsNamesAndIds,
    config.isCpuEnabled,
    config.memoryUnits,
  );
  return {
    clusterId: config.clusterId,
    parentId: config.parentId,
    resources: emptyResources,
    name: "",
    description: "",
    requestedNamespace: "",
    defaultNodePools: config.defaultNodePools,
  };
}

function getEmptyDepartmentModel(config: IEmptyDepartmentModelConfig): DepartmentCreationRequest {
  const emptyResources: Resources[] = getEmptyResourcesModel(
    config.nodePoolsNamesAndIds,
    config.isCpuEnabled,
    config.memoryUnits,
  );
  return {
    clusterId: config.clusterId,
    resources: emptyResources,
    defaultNodePools: config.defaultNodePools,
    name: "",
    description: "",
  };
}

function getEmptyResourcesCpu(isCpuEnabled: boolean): ResourcesCpu | null {
  return isCpuEnabled
    ? {
        deserved: null,
        limit: RESOURCE_MAX_ALLOWED_INFINITE_VALUE,
        overQuotaWeight: OVER_QUOTA_ENABLED_VALUE,
      }
    : null;
}
function getEmptyResourcesMemory(isCpuEnabled: boolean, memoryUnits?: Units): ResourcesMemory | null {
  return isCpuEnabled
    ? {
        deserved: null,
        limit: RESOURCE_MAX_ALLOWED_INFINITE_VALUE,
        overQuotaWeight: OVER_QUOTA_ENABLED_VALUE,
        units: memoryUnits || Units.Mib, // Default to Mib for backwards compatibility
      }
    : null;
}
function getEmptyResourcesModel(
  nodePools: INodePoolsNameAndId[],
  isCpuEnabled = false,
  memoryUnits?: Units,
): Resources[] {
  return nodePools.map((nodePool) => {
    return {
      nodePool: {
        name: nodePool.name,
        id: nodePool.id,
      } as ResourcesNodePool,
      gpu: {
        deserved: 0,
        limit: RESOURCE_MAX_ALLOWED_INFINITE_VALUE,
        overQuotaWeight: OVER_QUOTA_ENABLED_VALUE,
      } as ResourcesGpu,
      cpu: getEmptyResourcesCpu(isCpuEnabled),
      memory: getEmptyResourcesMemory(isCpuEnabled, memoryUnits),
    };
  });
}

function enrichResourcesWithCpuAndMemory(resources: Resources[]): Resources[] {
  return resources.map((resource) => {
    if (!resource.cpu) {
      resource.cpu = getEmptyResourcesCpu(true);
    }
    if (!resource.memory) {
      resource.memory = getEmptyResourcesMemory(true);
    }

    return resource;
  });
}

//eslint-disable-next-line @typescript-eslint/no-explicit-any
function _mapProjectStatusConditionsToMessages(conditions: any[]): string[] {
  return conditions
    .filter(
      (item) =>
        item.status === "False" && item.message && Object.keys(EProjectConditionDeletionReason).includes(item.reason),
    )
    .map((item) => item.message);
}

function _mapMessagesToTooltipText(messages: string[]): string {
  if (messages.length === 0) return "";
  const messagesList = messages
    .map((message) => message)
    .join("")
    .replace(/(?:\r\n|\r|\n)/g, "<br/>")
    .replace("first:", "first:<br>");

  return `
 ${messagesList}<br>
 Either delete the above or try <a href="javascript:void(0)" style="text-decoration: underline;" id="status-col-emit-project-deletion">deleting the project anyway</a>`;
}

function _mapDeletingMessagesToTooltipText(status: ProjectStatus): string {
  const messages = _mapProjectStatusConditionsToMessages(status?.additionalStatusData?.conditions || []);
  return _mapMessagesToTooltipText(messages);
}

function _mapNotReadyPhaseMessageToTooltip(phaseMessage: string): string {
  if (!phaseMessage) return "";

  const allReasons = phaseMessage
    .split(", ")
    .map((r) => r.trim())
    .filter((r): r is EProjectConditionNotReadyReason =>
      PROJECT_CONDITION_NOT_READY_REASON_TOOLTIP_MAP.has(r as EProjectConditionNotReadyReason),
    );

  const reasons =
    allReasons.length > 1
      ? allReasons.filter((r) => r !== EProjectConditionNotReadyReason.RoleBindingsHandlerFailed)
      : allReasons;

  return buildNotReadyTooltip(reasons);
}

function getStatusColOptions(project?: IProjectTable): IStatusColOptions {
  if (!project?.status || !project?.status.phase)
    return {
      status: "-",
      tooltipText: "",
      displayAnimation: false,
    };
  const statusOptions: IStatusColOptions = ProjectPhaseMap[project.status.phase];
  const options = tableUtil.getStatusColOptions(statusOptions, project.status.phaseMessage || "");

  if (project.status.phase === ProjectPhase.Deleting) {
    const tooltipText = _mapDeletingMessagesToTooltipText(project.status);
    if (tooltipText) options.tooltipText = tooltipText;
  } else if (project.status.phase === ProjectPhase.NotReady) {
    options.tooltipText = _mapNotReadyPhaseMessageToTooltip(project.status.phaseMessage || "");
  }

  return options;
}

function getNodePoolNameByResource(resources: Resources): string {
  return resources?.nodePool?.name || "";
}

function getNodePoolIdByResource(resources: Resources): string {
  return resources?.nodePool?.id || "";
}

function removeOverQuotaWeightFields(resources: Resources[]): Resources[] {
  return resources.map((resource) => {
    return {
      ...resource,
      gpu: {
        ...resource.gpu,
        overQuotaWeight: undefined,
      },
      cpu: resource.cpu
        ? {
            ...resource.cpu,
            overQuotaWeight: undefined,
          }
        : undefined,
      memory: resource.memory
        ? {
            ...resource.memory,
            overQuotaWeight: undefined,
          }
        : undefined,
    };
  });
}

function getProjectUpdateRequest(project: Project): ProjectUpdateRequest {
  return {
    description: project.description,
    schedulingRules: project.schedulingRules,
    defaultNodePools: project.defaultNodePools,
    nodeTypes: project.nodeTypes,
    resources: project.resources,
  };
}

function getDepartmentUpdateRequest(department: Department): DepartmentUpdateRequest {
  return {
    description: department.description,
    resources: department.resources,
    schedulingRules: department.schedulingRules,
    nodeTypes: department.nodeTypes,
    defaultNodePools: department.defaultNodePools,
  };
}
/****** Over Quota Weight *****/

/**
 * Determines if the over quota feature is enabled for a given resource.
 *
 * @param {Resources} resource - The resource to check.
 * @param {boolean} isOverQuotaPriorityEnabled - Flag indicating if the over quota priority feature is enabled.
 * @returns {boolean} - Returns `true` if the over quota feature is enabled for the resource, otherwise `false`.
 */
function isResourceOverQuotaEnabled(resource: Resources, isOverQuotaPriorityEnabled: boolean): boolean {
  if (isOverQuotaPriorityEnabled) return resource.gpu?.overQuotaWeight !== EOverQuotaPriority.None;
  return resource.gpu?.overQuotaWeight !== OVER_QUOTA_DISABLED_VALUE;
}

/**
 * Validates if the given value is a valid over quota value for the specified GPU resource.
 *
 * @param {number | null | undefined} value - The value to validate.
 * @param {ResourcesGpu} resourceGpu - The GPU resource to compare against.
 * @param {boolean} isOverQuotaEnabled - Flag indicating if the over quota feature is enabled.
 * @returns {boolean} - Returns `true` if the value is valid, otherwise `false`.
 */
function isOverQuotaValueValid(
  value: number | null | undefined,
  resourceGpu: ResourcesGpu,
  isOverQuotaEnabled: boolean,
): boolean {
  if (isOverQuotaEnabled) {
    return (
      value === RESOURCE_MAX_ALLOWED_INFINITE_VALUE ||
      value === RESOURCE_MAX_ALLOWED_INFINITE_VALUE_NUMERIC ||
      value === undefined ||
      value > resourceGpu.deserved
    );
  }
  return value === resourceGpu.deserved;
}

/**** limit and over quota weight update functions ****/

/**
 * Updates the overQuotaWeight of resources based on their limit.
 *
 * @param {Resources[]} resources - The array of resources to be updated.
 * @returns {Resources[]} - The updated array of resources.
 */
function updateOverQuotaWeightByLimit(resources: Resources[]): Resources[] {
  return resources.map((resource: Resources) => {
    const updatedResource = { ...resource };
    if (updatedResource.gpu.overQuotaWeight === EOverQuotaPriority.None) {
      updatedResource.gpu.limit = updatedResource.gpu.deserved;
      if (updatedResource?.cpu && updatedResource?.memory) {
        updatedResource.cpu.limit = updatedResource.cpu.deserved;
        updatedResource.memory.limit = updatedResource.memory.deserved;
      }
    } else {
      // over quota is enabled
      if (updatedResource.gpu.limit === undefined) {
        updatedResource.gpu.limit = RESOURCE_MAX_ALLOWED_INFINITE_VALUE;
      }
      if (updatedResource?.cpu && updatedResource?.memory) {
        updatedResource.cpu.limit = RESOURCE_MAX_ALLOWED_INFINITE_VALUE;
        updatedResource.memory.limit = RESOURCE_MAX_ALLOWED_INFINITE_VALUE;
      }
    }
    return updatedResource;
  });
}

function initOverQuotaWeightByLimitOqEnabled(resources: Resources[]): Resources[] {
  return resources.map((resource: Resources) => {
    const updatedResource = { ...resource };
    if (updatedResource.gpu.limit === updatedResource.gpu.deserved) {
      updatedResource.gpu.overQuotaWeight = EOverQuotaPriority.None;
      if (updatedResource?.cpu && updatedResource?.memory) {
        updatedResource.cpu.overQuotaWeight = EOverQuotaPriority.None;
        updatedResource.memory.overQuotaWeight = EOverQuotaPriority.None;
      }
    } else if (
      updatedResource.gpu?.limit &&
      updatedResource.gpu?.limit > updatedResource.gpu.deserved &&
      ![EOverQuotaPriority.High, EOverQuotaPriority.Medium, EOverQuotaPriority.Low, EOverQuotaPriority.Lowest].includes(
        updatedResource.gpu?.overQuotaWeight as EOverQuotaPriority,
      )
    ) {
      // over quota is enabled
      updatedResource.gpu.overQuotaWeight = EOverQuotaPriority.Medium;
      if (updatedResource?.cpu && updatedResource?.memory) {
        updatedResource.cpu.overQuotaWeight = EOverQuotaPriority.None;
        updatedResource.memory.overQuotaWeight = EOverQuotaPriority.None;
      }
    }

    return updatedResource;
  });
}

/**
 * Updates the limit or deserved values of resources based on the over quota status.
 *
 * @param {Resources[]} resources - The array of resources to be updated.
 * @returns {Resources[]} - The updated array of resources.
 */
function updateLimitOrDeservedByOverQuota(resources: Resources[]): Resources[] {
  return resources.map((resource) => {
    const updatedResource = { ...resource };
    const isOverQuotaEnabled = updatedResource.gpu.overQuotaWeight === OVER_QUOTA_ENABLED_VALUE;

    /**
     * Updates the limit values to RESOURCE_MAX_ALLOWED_INFINITE_VALUE.
     */
    const updateLimitToOverQuota = () => {
      if (updatedResource.gpu.limit === undefined) {
        updatedResource.gpu.limit = RESOURCE_MAX_ALLOWED_INFINITE_VALUE_NUMERIC;
      }

      //CPU FF is enabled
      if (updatedResource?.cpu && updatedResource?.memory) {
        updatedResource.cpu.limit = RESOURCE_MAX_ALLOWED_INFINITE_VALUE;
        updatedResource.memory.limit = RESOURCE_MAX_ALLOWED_INFINITE_VALUE;
      }
    };

    /**
     * Updates the limit values to the deserved values.
     */
    const updateLimitToDeserved = () => {
      updatedResource.gpu.limit = updatedResource.gpu.deserved;

      //CPU FF is enabled
      if (updatedResource?.cpu && updatedResource?.memory) {
        updatedResource.cpu.limit = updatedResource.cpu?.deserved;
        updatedResource.memory.limit = updatedResource.memory?.deserved;
      }
    };

    if (isOverQuotaEnabled) {
      updateLimitToOverQuota();
    } else {
      updateLimitToDeserved();
    }
    return resource;
  });
}

/** Priority **/
function mapStringToPriorityLabel(value: string): EEntityPriorityLabel {
  return entityPriorityMap[value] || ENTITY_PRIORITY_DEFAULT_VALUE;
}

/**
 * we have two states in our app, over quota weight feature is disable (binary choice enabled/disabled) and over quota weight feature is enabled.
 * for the first state, we have two options for overQuotaWeight, OVER_QUOTA_ENABLED_VALUE and OVER_QUOTA_DISABLED_VALUE, and we update the overQuotaWeight based on the limit, on load, and  update limit or deserved values on submit.
 * for the second state, we have multiple options for overQuotaWeight, (None, Lowest, Low, Medium, High) none means no over quota so limit = deserved.
 * @param resources
 * @param isOverQuotaPriorityEnabled
 */
function updateLimitOrOverQuota(resources: Resources[], isOverQuotaPriorityEnabled: boolean): Resources[] {
  let updatedResources: Resources[];
  if (isOverQuotaPriorityEnabled) {
    updatedResources = updateOverQuotaWeightByLimit(resources);
  } else {
    updatedResources = updateLimitOrDeservedByOverQuota(resources);
  }

  return updatedResources.map((resource) => {
    if (resource?.gpu.limit === RESOURCE_MAX_ALLOWED_INFINITE_VALUE) {
      resource.gpu.limit = null;
    }
    return resource;
  });
}

function initOverQuotaWeightByLimit(resources: Resources[]): Resources[] {
  return resources.map((nodePool: Resources) => {
    const updatedNodePool = { ...nodePool };
    const isOverQuotaEnabled =
      nodePool.gpu.limit === undefined ||
      nodePool.gpu.limit === RESOURCE_MAX_ALLOWED_INFINITE_VALUE ||
      nodePool.gpu.limit === RESOURCE_MAX_ALLOWED_INFINITE_VALUE_NUMERIC ||
      nodePool.gpu.limit > nodePool.gpu.deserved;

    if (isOverQuotaEnabled) {
      updatedNodePool.gpu.overQuotaWeight = OVER_QUOTA_ENABLED_VALUE;
    } else {
      updatedNodePool.gpu.overQuotaWeight = OVER_QUOTA_DISABLED_VALUE;
    }

    return updatedNodePool;
  });
}

// init the over overQuotaWeight based on limit
function initResourcesOverQuotaWeight(resources: Resources[], isOverQuotaPriorityEnabled: boolean): Resources[] {
  if (isOverQuotaPriorityEnabled) {
    return initOverQuotaWeightByLimitOqEnabled(resources);
  } else {
    return initOverQuotaWeightByLimit(resources);
  }
}

/***** SchedulingRules *****/

/**
 * Updates the project scheduling rules based on the parent scheduling rules.
 *
 * The project scheduling rules can only restrict (i.e., have a smaller value than) the parent scheduling rules.
 * If the project did not mutate the parent value, the project value is set to `null`.
 * If the parent value is `null` or undefined, the project value remains unchanged.
 * If the project value is greater than the parent value, the project value is set to the parent value. (this case should not happen due to validation).
 *
 * @param {SchedulingRules} projectSR - The scheduling rules of the project.
 * @param {SchedulingRules} parentSR - The scheduling rules of the parent.
 * @returns {SchedulingRules} - The updated scheduling rules for the project.
 */
function updateSchedulingRules(projectSR: SchedulingRules, parentSR: SchedulingRules): SchedulingRules {
  const updatedSR: SchedulingRules = { ...projectSR };

  for (const key in projectSR) {
    if (Object.prototype.hasOwnProperty.call(projectSR, key) && Object.prototype.hasOwnProperty.call(parentSR, key)) {
      const projectValue = projectSR[key as keyof SchedulingRules];
      const parentValue = parentSR[key as keyof SchedulingRules];

      if (parentValue !== null && parentValue !== undefined && projectValue !== null && projectValue !== undefined) {
        if (projectValue === parentValue) {
          updatedSR[key as keyof SchedulingRules] = null;
        }

        if (projectValue > parentValue) {
          updatedSR[key as keyof SchedulingRules] = parentValue;
        }
      }
    }
  }

  return updatedSR;
}

/**
 * Initializes the scheduling rules for a project based on the parent scheduling rules.
 *
 * This function will copy any scheduling rules from the parent that are not defined in the project.
 *
 * @param {SchedulingRules} projectSR - The scheduling rules of the project.
 * @param {SchedulingRules} parentSR - The scheduling rules of the parent.
 * @returns {SchedulingRules} - The initialized scheduling rules for the project.
 */
function initSchedulingRules(projectSR: SchedulingRules, parentSR: SchedulingRules): SchedulingRules {
  const updatedSR: SchedulingRules = { ...projectSR };
  for (const key in parentSR) {
    if (Object.prototype.hasOwnProperty.call(parentSR, key) && !Object.prototype.hasOwnProperty.call(projectSR, key)) {
      updatedSR[key as keyof SchedulingRules] = parentSR[key as keyof SchedulingRules];
    }
  }
  return updatedSR;
}

function updateOverQuotaWeightToDeserved(resources: Resources[]): Resources[] {
  return resources.map((resource) => {
    const updatedResource = { ...resource };
    updatedResource.gpu.overQuotaWeight = updatedResource.gpu.deserved;

    if (updatedResource?.cpu) {
      updatedResource.cpu.overQuotaWeight = updatedResource.cpu.deserved;
    }
    if (updatedResource?.memory) {
      updatedResource.memory.overQuotaWeight = updatedResource.memory.deserved;
    }

    return updatedResource;
  });
}

/**
 * Align CPU and Memory limits with GPU over-quota state.
 * - If GPU over-quota is enabled (unlimited or limit > deserved), sets CPU/Memory limits to null.
 * - If GPU over-quota is disabled (limit <= deserved), sets CPU/Memory limits to their deserved values.
 */
function alignCpuAndMemoryLimitsByGpu(resources: Resources[]): Resources[] {
  return resources.map((resource) => {
    const updated = { ...resource };
    const gpuLimit = resource.gpu?.limit;
    const gpuDeserved = resource.gpu?.deserved ?? 0;
    const isGpuUnlimited = gpuLimit === null || gpuLimit === undefined || gpuLimit === -1;
    const isGpuOverQuotaEnabled = isGpuUnlimited || gpuLimit > gpuDeserved;

    if (resource.cpu) {
      updated.cpu = {
        ...resource.cpu,
        limit: isGpuOverQuotaEnabled ? null : (resource.cpu.deserved ?? null),
      };
    }

    if (resource.memory) {
      updated.memory = {
        ...resource.memory,
        limit: isGpuOverQuotaEnabled ? null : (resource.memory.deserved ?? null),
        units: resource.memory.units ?? Units.Mb,
      };
    }

    return updated;
  });
}

/**
 * Ensures that limit values are not less than deserved values.
 * When limit is defined (not null/undefined/-1) and is less than deserved,
 * it sets limit equal to deserved to prevent API validation errors.
 *
 * This handles the case: "deserved is more than the limit" API error
 */
function ensureLimitsNotBelowDeserved(resources: Resources[]): Resources[] {
  return resources.map((resource) => {
    const updated = { ...resource };

    // GPU: if limit is defined and less than deserved, set limit = deserved
    if (resource.gpu) {
      const gpuLimit = resource.gpu.limit;
      const gpuDeserved = resource.gpu.deserved ?? 0;
      const isGpuLimitDefined = gpuLimit !== null && gpuLimit !== undefined && gpuLimit !== -1;

      if (isGpuLimitDefined && gpuLimit < gpuDeserved) {
        updated.gpu = {
          ...resource.gpu,
          limit: gpuDeserved,
        };
      }
    }

    // CPU: if limit is defined and less than deserved, set limit = deserved
    if (resource.cpu) {
      const cpuLimit = resource.cpu.limit;
      const cpuDeserved = resource.cpu.deserved ?? 0;
      const isCpuLimitDefined = cpuLimit !== null && cpuLimit !== undefined && cpuLimit !== -1;

      if (isCpuLimitDefined && cpuLimit < cpuDeserved) {
        updated.cpu = {
          ...resource.cpu,
          limit: cpuDeserved,
        };
      }
    }

    // Memory: if limit is defined and less than deserved, set limit = deserved
    if (resource.memory) {
      const memLimit = resource.memory.limit;
      const memDeserved = resource.memory.deserved ?? 0;
      const isMemLimitDefined = memLimit !== null && memLimit !== undefined && memLimit !== -1;

      if (isMemLimitDefined && memLimit < memDeserved) {
        updated.memory = {
          ...resource.memory,
          limit: memDeserved,
        };
      }
    }

    return updated;
  });
}

/**
 * Build BatchPatchNodePoolResourcesRequest for peer project updates.
 * If over quota priority is disabled, overQuotaWeight is omitted.
 * If original limit was unlimited, keep it; otherwise sync limit to new quota.
 */
/**
 * Build batch patch request for quota changes
 * Used in GPU quota step to update peer org unit quotas (projects/departments)
 */
function buildNodePoolBatchPatchRequest(
  peerChanges: Array<{
    orgUnitId: string;
    nodePoolId: string;
    nodePoolName: string;
    gpuQuota: number | null;
    originalLimit: number | null;
    originalOverQuotaWeight: number | null;
  }>,
  isOverQuotaPriorityEnabled: boolean,
): BatchPatchNodePoolResourcesRequest {
  const patchRequests: NodePoolResourcesPatchRequest[] = peerChanges.map((change) => {
    const isUnlimited =
      change.originalLimit === null || change.originalLimit === -1 || change.originalLimit === undefined;

    const overQuotaWeight = isOverQuotaPriorityEnabled ? change.originalOverQuotaWeight : undefined;

    const resources: ResourcesPerNodePoolNullable = {
      gpu: {
        deserved: change.gpuQuota,
        limit: isUnlimited ? change.originalLimit : change.gpuQuota,
        overQuotaWeight,
      } as ResourcesNullableGpu,
    };

    return {
      orgUnitId: change.orgUnitId,
      nodePoolId: change.nodePoolId,
      resources,
    };
  });

  return {
    patchRequests,
  };
}

/**
 * Build batch patch request for rank changes
 * Used in GPU over quota step to update peer org unit ranks (projects/departments)
 */
function buildNodePoolRankBatchPatchRequest(
  rankChanges: Array<{
    orgUnitId: string;
    nodePoolId: string;
    newRank: string;
    originalRank: string;
  }>,
): BatchPatchNodePoolResourcesRequest {
  const patchRequests: NodePoolResourcesPatchRequest[] = rankChanges.map((change) => {
    return {
      orgUnitId: change.orgUnitId,
      nodePoolId: change.nodePoolId,
      rank: change.newRank,
    };
  });

  return {
    patchRequests,
  };
}

/**
 * Build batch patch request for peer org unit weights (projects/departments)
 */
function buildNodePoolWeightBatchPatchRequest(
  weightChanges: Array<{
    orgUnitId: string;
    nodePoolId: string;
    newWeight: number;
    originalWeight: number | null;
  }>,
): BatchPatchNodePoolResourcesRequest {
  const patchRequests: NodePoolResourcesPatchRequest[] = weightChanges.map((change) => {
    return {
      orgUnitId: change.orgUnitId,
      nodePoolId: change.nodePoolId,
      resources: {
        gpu: {
          overQuotaWeight: change.newWeight,
        },
      },
    };
  });

  return {
    patchRequests,
  };
}

/**
 * Build batch patch request for peer org unit changes (combined rank and/or weight)
 */
function buildNodePoolOrgUnitChangesBatchPatchRequest(
  changes: Array<{
    orgUnitId: string;
    nodePoolId: string;
    newRank?: string;
    originalRank?: string;
    newWeight?: number;
    originalWeight?: number | null;
  }>,
): BatchPatchNodePoolResourcesRequest {
  const patchRequests: NodePoolResourcesPatchRequest[] = changes.map((change) => {
    const request: NodePoolResourcesPatchRequest = {
      orgUnitId: change.orgUnitId,
      nodePoolId: change.nodePoolId,
    };

    // Add rank if it changed
    if (change.newRank) {
      request.rank = change.newRank;
    }

    // Add weight if it changed
    if (change.newWeight !== undefined) {
      request.resources = {
        gpu: {
          overQuotaWeight: change.newWeight,
        },
      };
    }

    return request;
  });

  return {
    patchRequests,
  };
}

/** Over time data **/
function getAverageGpuAllocationDataByTimeframe(overtimeData: OvertimeData, timeframe: string): number | undefined {
  switch (timeframe) {
    case EOrgUnitOverTimeValue.TwentyFourHours:
      return overtimeData?.range24hData?.averageGpuAllocation;
    case EOrgUnitOverTimeValue.SevenDays:
      return overtimeData?.range7dData?.averageGpuAllocation;
    case EOrgUnitOverTimeValue.ThirtyDays:
      return overtimeData?.range30dData?.averageGpuAllocation;
    default:
      return undefined;
  }
}

function getAverageGpuUtilizationDataByTimeframe(overtimeData: OvertimeData, timeframe: string): number | undefined {
  switch (timeframe) {
    case EOrgUnitOverTimeValue.TwentyFourHours:
      return overtimeData?.range24hData?.averageGpuUtilization;
    case EOrgUnitOverTimeValue.SevenDays:
      return overtimeData?.range7dData?.averageGpuUtilization;
    case EOrgUnitOverTimeValue.ThirtyDays:
      return overtimeData?.range30dData?.averageGpuUtilization;
    default:
      return undefined;
  }
}

function getAverageGpuMemoryUtilizationDataByTimeframe(
  overtimeData: OvertimeData,
  timeframe: string,
): number | undefined {
  switch (timeframe) {
    case EOrgUnitOverTimeValue.TwentyFourHours:
      return overtimeData?.range24hData?.averageGpuMemoryUtilization || undefined;
    case EOrgUnitOverTimeValue.SevenDays:
      return overtimeData?.range7dData?.averageGpuMemoryUtilization || undefined;
    case EOrgUnitOverTimeValue.ThirtyDays:
      return overtimeData?.range30dData?.averageGpuMemoryUtilization || undefined;
    default:
      return undefined;
  }
}

function _getAvgAllocationFilterKeyByTimeframe(timeframe: string): ProjectFilterSortFields {
  switch (timeframe) {
    case EOrgUnitOverTimeValue.TwentyFourHours:
      return ProjectFilterSortFields.AvgGpuAllocation24h;
    case EOrgUnitOverTimeValue.SevenDays:
      return ProjectFilterSortFields.AvgGpuAllocation7d;
    case EOrgUnitOverTimeValue.ThirtyDays:
      return ProjectFilterSortFields.AvgGpuAllocation30d;
    default:
      return ProjectFilterSortFields.AvgGpuAllocation24h;
  }
}

function _getAvgUtilizationFilterKeyByTimeframe(timeframe: string): ProjectFilterSortFields {
  switch (timeframe) {
    case EOrgUnitOverTimeValue.TwentyFourHours:
      return ProjectFilterSortFields.AvgGpuUtilization24h;
    case EOrgUnitOverTimeValue.SevenDays:
      return ProjectFilterSortFields.AvgGpuUtilization7d;
    case EOrgUnitOverTimeValue.ThirtyDays:
      return ProjectFilterSortFields.AvgGpuUtilization30d;
    default:
      return ProjectFilterSortFields.AvgGpuUtilization24h;
  }
}

function _getAvgGpuMemoryUtilizationFilterKeyByTimeframe(timeframe: string): ProjectFilterSortFields {
  switch (timeframe) {
    case EOrgUnitOverTimeValue.TwentyFourHours:
      return ProjectFilterSortFields.AvgGpuMemoryUtilization24h;
    case EOrgUnitOverTimeValue.SevenDays:
      return ProjectFilterSortFields.AvgGpuMemoryUtilization7d;
    case EOrgUnitOverTimeValue.ThirtyDays:
      return ProjectFilterSortFields.AvgGpuMemoryUtilization30d;
    default:
      return ProjectFilterSortFields.AvgGpuMemoryUtilization24h;
  }
}

function _replaceOverTimeFilterKey(
  filters: IPaginationFilter,
  gpuAllocationTimeframe: string,
  gpuUtilizationTimeframe: string,
  gpuMemoryUtilizationTimeframe: string,
): IPaginationFilter {
  const avgGpuAllocationKey = _getAvgAllocationFilterKeyByTimeframe(gpuAllocationTimeframe);
  const avgGpuUtilizationKey = _getAvgUtilizationFilterKeyByTimeframe(gpuUtilizationTimeframe);
  const avgGpuMemoryUtilizationKey = _getAvgGpuMemoryUtilizationFilterKeyByTimeframe(gpuMemoryUtilizationTimeframe);

  const sortBy = _replaceSortByKey(
    filters.sortBy || undefined,
    avgGpuAllocationKey,
    avgGpuUtilizationKey,
    avgGpuMemoryUtilizationKey,
  );
  const filterBy = _replaceFilterByKeys(
    filters.filterBy,
    avgGpuAllocationKey,
    avgGpuUtilizationKey,
    avgGpuMemoryUtilizationKey,
  );

  return {
    ...filters,
    sortBy,
    filterBy,
  };
}

function _replaceSortByKey(
  sortBy: string | undefined,
  avgGpuAllocationKey: string,
  avgGpuUtilizationKey: string,
  avgGpuMemoryUtilizationKey: string,
): string | undefined {
  if (sortBy === EOrgUnitOverTimeColumnName.AverageGpuAllocation) {
    return avgGpuAllocationKey;
  }
  if (sortBy === EOrgUnitOverTimeColumnName.AverageGpuUtilization) {
    return avgGpuUtilizationKey;
  }
  if (sortBy === EOrgUnitOverTimeColumnName.AverageGpuMemoryUtilization) {
    return avgGpuMemoryUtilizationKey;
  }
  return sortBy;
}

function _replaceFilterByKeys(
  filterBy: string[] | undefined,
  avgGpuAllocationKey: string,
  avgGpuUtilizationKey: string,
  avgGpuMemoryUtilizationKey: string,
): string[] | undefined {
  if (!filterBy) return filterBy;

  return filterBy.map((filter) => {
    if (filter.startsWith(EOrgUnitOverTimeColumnName.AverageGpuAllocation)) {
      return filter.replace(EOrgUnitOverTimeColumnName.AverageGpuAllocation, avgGpuAllocationKey);
    }
    if (filter.startsWith(EOrgUnitOverTimeColumnName.AverageGpuUtilization)) {
      return filter.replace(EOrgUnitOverTimeColumnName.AverageGpuUtilization, avgGpuUtilizationKey);
    }
    if (filter.startsWith(EOrgUnitOverTimeColumnName.AverageGpuMemoryUtilization)) {
      return filter.replace(EOrgUnitOverTimeColumnName.AverageGpuMemoryUtilization, avgGpuMemoryUtilizationKey);
    }
    return filter;
  });
}

function getProjectListFilters(filterBy: IFilterBy): IPaginationFilter {
  const validFilterFields = [
    ...Object.values(ProjectFilterSortFields),
    CLUSTER_COLUMN_FILTER_NAME,
    EOrgUnitOverTimeColumnName.AverageGpuAllocation,
    EOrgUnitOverTimeColumnName.AverageGpuUtilization,
    EOrgUnitOverTimeColumnName.AverageGpuMemoryUtilization,
  ];
  const filters: IPaginationFilter = filterService.mapColumnsFilterToFilterParams(filterBy, validFilterFields, true);
  // @ts-ignore TODO: fix this
  const gpuAllocationTimeframe = (filterBy as IProjectTableFilterBy).avgGpuAllocationTimeframe;
  // @ts-ignore TODO: fix this
  const gpuUtilizationTimeframe = (filterBy as IProjectTableFilterBy).avgGpuUtilizationTimeframe;
  // @ts-ignore TODO: fix this
  const gpuMemoryUtilizationTimeframe = (filterBy as IProjectTableFilterBy).avgGpuMemoryUtilizationTimeframe;

  return _replaceOverTimeFilterKey(
    filters,
    gpuAllocationTimeframe,
    gpuUtilizationTimeframe,
    gpuMemoryUtilizationTimeframe,
  );
}

function getDepartmentListFilters(filterBy: IFilterBy): IPaginationFilter {
  const validFilterFields = [
    ...Object.values(DepartmentFilterSortFields),
    CLUSTER_COLUMN_FILTER_NAME,
    EOrgUnitOverTimeColumnName.AverageGpuAllocation,
    EOrgUnitOverTimeColumnName.AverageGpuUtilization,
    EOrgUnitOverTimeColumnName.AverageGpuMemoryUtilization,
  ];
  const filters: IPaginationFilter = filterService.mapColumnsFilterToFilterParams(filterBy, validFilterFields);
  // @ts-ignore TODO: fix this
  const gpuAllocationTimeframe = (filterBy as IDepartmentTableFilterBy).avgGpuAllocationTimeframe;
  // @ts-ignore TODO: fix this
  const gpuUtilizationTimeframe = (filterBy as IDepartmentTableFilterBy).avgGpuUtilizationTimeframe;
  // @ts-ignore TODO: fix this
  const avgGpuMemoryUtilizationTimeframe = (filterBy as IDepartmentTableFilterBy).avgGpuMemoryUtilizationTimeframe;

  return _replaceOverTimeFilterKey(
    filters,
    gpuAllocationTimeframe,
    gpuUtilizationTimeframe,
    avgGpuMemoryUtilizationTimeframe,
  );
}

/** Tables utils **/

function getAllocationRatioDisplayValue(allocatedGpus: number, totalResources?: TotalResources): string {
  const gpuQuota = totalResources?.gpuQuota || 0;
  if (gpuQuota < 0) return "-";
  return toPercent(allocatedGpus || 0, gpuQuota);
}

/** Ranking utils **/

function mapRankingMetricToMetricTypeAndOrder(
  metricValue: EOrgUnitRankingMetricValue,
  timeFrame: EOrgUnitOverTimeValue,
): { metricType: RankingMetricType; order: RankingOrderType } {
  const metricType = rankingMetricTypeMap[metricValue]?.[timeFrame];
  const order = rankingOrderMap[metricValue];

  if (!metricType || !order) {
    throw new Error("Invalid metric value or time frame");
  }

  return { metricType, order };
}

function getProgressColOptions(value: number | null): IProgressColOptions {
  return {
    value: value,
    color: "soft-blue",
  };
}

//*** V2 forms utils ****/

function convertMemoryValue(
  value: number | null | undefined,
  fromUnit: Units | null | undefined,
  toUnit: Units,
): number | null {
  if (value === null || value === undefined) {
    return null;
  }
  if (!fromUnit || fromUnit === toUnit) {
    return value;
  }

  const valueInMib = convertToMib(value, fromUnit);
  if (valueInMib === null) {
    return null;
  }
  return convertFromMib(valueInMib, toUnit);
}

function convertToMib(value: number, unit: Units): number | null {
  switch (unit) {
    case Units.Gb:
      return value * 1024;
    case Units.Mb:
      return value / 1.048576;
    case Units.Mib:
      return value;
    default:
      return null;
  }
}

function convertFromMib(value: number, unit: Units): number | null {
  switch (unit) {
    case Units.Gb:
      return value / 1024;
    case Units.Mb:
      return value * 1.048576;
    case Units.Mib:
      return value;
    default:
      return null;
  }
}

/**
 * Determines the GPU state for a single nodepool based on deserved and limit values.
 * - limit === 0 -> NoGPUResourcesAreSet
 * - deserved === 0 && (limit > 0 OR limit is unlimited) -> OpportunisticOnly
 * - deserved > 0 && limit === deserved -> DedicatedGPUs
 * - deserved > 0 && (limit > deserved OR limit is unlimited) -> DedicatedAndOpportunisticGPUs
 *
 * Note: limit === -1, null, or undefined indicates unlimited
 * Note: deserved === -1 indicates unlimited dedicated resources
 */
function getNodePoolGpuState(gpu: ResourcesNullableGpu | null | undefined): EGpuResourceState {
  if (!gpu) {
    return EGpuResourceState.NoGPUResourcesAreSet;
  }

  const deserved = gpu.deserved ?? 0;
  const limit = gpu.limit;

  // Check if limit is unlimited (null, undefined, or -1)
  const isLimitUnlimited = limit === null || limit === undefined || limit === -1;

  // Check if deserved is unlimited (-1)
  const isDeservedUnlimited = deserved === -1;

  // limit === 0 means no GPU resources
  if (limit === 0) {
    return EGpuResourceState.NoGPUResourcesAreSet;
  }

  // deserved > 0 or unlimited, with opportunistic capability (limit > deserved or limit is unlimited)
  if ((deserved > 0 || isDeservedUnlimited) && (isLimitUnlimited || limit > deserved)) {
    return EGpuResourceState.DedicatedAndOpportunisticGPUs;
  }

  // deserved > 0 with limit === deserved (dedicated only, no opportunistic)
  if (deserved > 0 && limit === deserved) {
    return EGpuResourceState.DedicatedGPUs;
  }

  // deserved === 0 with some limit available (opportunistic only)
  if (deserved === 0 && (isLimitUnlimited || limit > 0)) {
    return EGpuResourceState.OpportunisticOnly;
  }

  return EGpuResourceState.NoGPUResourcesAreSet;
}

/**
 * Aggregates GPU states from multiple nodepools.
 * Uses priority order: DedicatedAndOpportunistic > Dedicated > Opportunistic > None
 * Returns the highest priority state found across all nodepools.
 *
 * Edge case: if there are no nodepools in DedicatedAndOpportunisticGPUs state,
 * but there is at least one nodepool in OpportunisticOnly state AND at least one in DedicatedGPUs state,
 * returns DedicatedAndOpportunisticGPUs (because the aggregate has both dedicated and opportunistic resources).
 */
function aggregateGpuStates(states: EGpuResourceState[]): EGpuResourceState {
  if (states.length === 0) {
    return EGpuResourceState.NoGPUResourcesAreSet;
  }

  // Edge case: mixed Opportunistic + Dedicated across nodepools = DedicatedAndOpportunistic aggregate
  if (
    states.every((state) => state !== EGpuResourceState.DedicatedAndOpportunisticGPUs) &&
    states.some((state) => state === EGpuResourceState.OpportunisticOnly) &&
    states.some((state) => state === EGpuResourceState.DedicatedGPUs)
  ) {
    return EGpuResourceState.DedicatedAndOpportunisticGPUs;
  }

  return Math.max(...states) as EGpuResourceState;
}
