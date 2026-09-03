/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { workloadServiceApi } from "@/services/infra/client-apis/workloads-service-api/workloads-service-api";
import { workloadsManagerApi } from "@/services/infra/client-apis/workloads-manager-api/workloads-manager-api";
import {
  type WorkloadV2CreateRequest,
  type WorkloadV2,
  type WorkloadV2GetResponse,
  type YamlSubmissionSetting,
  type GetYamlSubmission,
  RunState,
  WorkloadsSettingsKeyEnum,
} from "@/swagger-models/workloads-manager-client";
import { httpService } from "@/services/infra/https.service/http.service";

import type { AxiosResponse } from "axios";
import type {
  ElementMetricType,
  WorkloadMetricType,
  MetricsResponse,
  Pod,
  PodMetricType,
  PodVerbosity,
  BatchResponse,
  WorkloadDetailed,
  PendingSchedulingMessage,
  UnifiedWorkload,
  UnifiedWorkloadRequest,
  WorkloadPriority,
  ClusterPriority,
  WorkloadTypeConfig,
  WorkloadCategory,
  Workspace,
  Training,
  DistributedWorkload,
  Inference,
  DistributedInference,
  WorkloadElement,
  WorkloadElementResponse,
} from "@/swagger-models/workloads-client";
import {
  SourceApi,
  WorkloadSortFilterFields,
  WorkloadsSettingsKeyEnum as WorkloadsServiceSettingsKeyEnum,
} from "@/swagger-models/workloads-client";
import type { GetWorkloadsSetting } from "@/swagger-models/workloads-client";
import { API } from "@/common/api.constant";
import type {
  EWorkloadTelemetryGroupBy,
  ICanSubmitWorkload,
  EDistributedFrameworkUrlSlug,
  EWorkloadTelemetryGroupByWithoutBucket,
} from "@/models/workload.model";
import {
  workloadPriorityNameByWorkloadType,
  distributedFrameworkUrlMapping,
  Preemptibility,
  EWorkloadName,
} from "@/models/workload.model";
import { EWorkloadType } from "@/models/workload.model";

import type { IFilterBy, IPaginationFilter } from "@/models/filter.model";
import { controlPlaneService } from "@/services/control-plane/control-plane.service/control-plane.service";
import type {
  GetWorkloadHistory200Response,
  TelemetryResponse,
  Workload,
  WorkloadPatchFields,
  WorkloadTelemetryType,
} from "@/swagger-models/workloads-client";
import type { Phase } from "@/swagger-models/workloads-client";
import { workloadLocalPhaseUpdater } from "@/services/cluster/workload.service/local-status-updater";
import { filterService } from "@/services/filter.service/filter.service";
import { fileService, FileType } from "@/services/file.service/file.service";
import { ClusterState, type ClustersMinimalData } from "@/swagger-models/cluster-service-client";
import type { Project, Resources, QuotaStatusNodePool } from "@/swagger-models/org-unit-service-client";

export const workloadService = {
  list,
  getWorkloadPods,
  getWorkloadHistory,
  getWorkloadById,
  getWorkloadV2ById,
  getWorkloadCrudById,
  getWorkloadsCount,
  getWorkloads,
  getWorkloadsTelemetry,
  getWorkloadsTelemetryCsv,
  countWorkloadsByName,
  getWorkloadMetrics,
  listPods,
  countPods,
  patchWorkloadById,
  getWorkloadPodMetrics,
  getWorkloadPendingSchedulingMessages,
  getUnifiedWorkload,
  canSubmitWorkload,
  // Create
  createWorkloadFromYaml,
  // Delete
  deleteWorkloadV2,
  deleteWorkloadManagerV2,
  deleteMultipleWorkloads,
  // Stop
  stopWorkloadV2,
  stopMultipleWorkloads,
  // Activate
  activateWorkloadV2,
  activateMultipleWorkloads,
  // Priorities
  getWorkloadClusterPriorities,
  getWorkloadPriorities,
  getFallbackWorkloadPriorityByType,
  getDefaultWorkloadPriorityByType,
  // Preemptibility
  getDefaultPreemptibilityByType,
  // Categories
  listCategories,
  // Utilities
  getFrameworkForUrl,
  exceedsAvailableQuota,
  // Workloads settings
  getYamlSubmissionSetting,
  updateYamlSubmissionSetting,
  getApplicationSubmissionSetting,
  updateApplicationSubmissionSetting,
  // Structure elements
  getWorkloadElements,
  getWorkloadElement,
  getWorkloadElementMetrics,
};

const WORKLOADS_SERVICE_ENDPOINT = `${API.v1}/workloads`;
const WORKLOADS_TELEMETRY_ENDPOINT = `${API.v1}/workloads/telemetry`;

// Types
type FrameworkUrlType = EDistributedFrameworkUrlSlug | "runai" | "knative";

async function createWorkloadFromYaml(
  request: WorkloadV2CreateRequest,
  dryRun = false,
  returnRawError = false,
): Promise<WorkloadV2> {
  try {
    const response = await workloadsManagerApi.workloadsV2Api.createWorkloadV2(request, dryRun);
    return response.data;
  } catch (error: unknown) {
    if (returnRawError) {
      throw error;
    }
    throw httpService.handleHttpError(error);
  }
}

async function getWorkloads(
  deleted?: boolean,
  offset?: number,
  limit?: number,
  sortOrder?: "asc" | "desc" | undefined,
  sortBy?: WorkloadSortFilterFields,
  filterBy?: string[],
  search?: string,
): Promise<Workload[]> {
  const response = await workloadServiceApi.workloadsApi.getWorkloads(
    deleted,
    offset,
    limit,
    sortOrder,
    sortBy,
    filterBy,
    search,
  );
  return response.data.workloads;
}

async function list(clusterUuid: string, filterBy: IFilterBy = {}, deleted = false): Promise<Workload[]> {
  const validFilterFields = [...Object.values(WorkloadSortFilterFields)];
  filterBy.clusterUuid = clusterUuid;
  const filters: IPaginationFilter = filterService.mapColumnsFilterToFilterParams(filterBy, validFilterFields, true);

  const transformedFilterBy = _transformPriorityFilters(filters.filterBy);

  const workloads = await getWorkloads(
    deleted,
    filters.offset,
    filters.limit,
    filters.sortOrder,
    filters.sortBy as WorkloadSortFilterFields,
    transformedFilterBy,
    filters.search,
  );
  return workloadLocalPhaseUpdater.updateWorkloadsByLocalPhases(workloads);
}

async function getWorkloadsCount(
  filterBy?: Array<string>,
  deleted = false,
  search?: string,
  precise?: boolean,
): Promise<number> {
  const transformedFilterBy = _transformPriorityFilters(filterBy);

  const response = await workloadServiceApi.workloadsApi.countWorkloads(deleted, transformedFilterBy, search, precise);
  return response.data.count;
}

/**
 * Transforms priority filters to use priorityClassName for filtering operations.
 * This ensures that when users filter by priority (not sort), we use the priorityClassName field.
 * This is necessary because filter by priority return multiple results for the selected priorityClassName that we display i the UI filter.
 * @param filterBy Array of filter strings
 * @returns Transformed filterBy array with priority filters converted to priorityClassName
 */
function _transformPriorityFilters(filterBy?: Array<string>): Array<string> | undefined {
  if (!filterBy || filterBy.length === 0) {
    return filterBy;
  }

  return filterBy.map((filter) => {
    if (filter.includes(WorkloadSortFilterFields.Priority)) {
      return filter.replace(WorkloadSortFilterFields.Priority, WorkloadSortFilterFields.PriorityClassName);
    }
    return filter;
  });
}

async function deleteMultipleWorkloads(workloadIds: Array<string>): Promise<BatchResponse> {
  return (await workloadServiceApi.workloadsBatchApi.batchWorkloads({ ids: workloadIds, action: "delete" })).data;
}

async function deleteWorkloadV2(workloadId: string, workloadType: string): Promise<void> {
  try {
    switch (workloadType) {
      case EWorkloadType.Distributed:
        await workloadServiceApi.distributedApi.deleteDistributed(workloadId);
        break;
      case EWorkloadType.Training:
        await workloadServiceApi.trainingsApi.deleteTraining(workloadId);
        break;
      case EWorkloadType.Workspace:
        await workloadServiceApi.workspacesApi.deleteWorkspace(workloadId);
        break;
      case EWorkloadType.Inference:
        await workloadServiceApi.inferencesApi.deleteInference(workloadId);
        break;
      case EWorkloadType.DistributedInference:
        await workloadServiceApi.distributedInferencesApi.deleteDistributedInference(workloadId);
        break;
      default:
        throw new Error(`Unsupported workload type: ${workloadType}`);
    }
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function deleteWorkloadManagerV2(workloadId: string): Promise<void> {
  try {
    await workloadsManagerApi.workloadsV2Api.deleteWorkloadV2(workloadId);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function stopMultipleWorkloads(workloadIds: Array<string>): Promise<BatchResponse> {
  return (await workloadServiceApi.workloadsBatchApi.batchWorkloads({ ids: workloadIds, action: "suspend" })).data;
}

async function stopWorkloadV2(workload: Workload): Promise<void> {
  try {
    if (workload.sourceApi === SourceApi.WorkloadsV2) {
      await workloadsManagerApi.workloadsV2Api.setWorkloadRunState(workload.id, { value: RunState.Suspend });
      return;
    }

    switch (workload.type) {
      case EWorkloadType.Training:
        (await workloadServiceApi.trainingsApi.suspendTraining(workload.id)).data;
        break;
      case EWorkloadType.Workspace:
        (await workloadServiceApi.workspacesApi.suspendWorkspace(workload.id)).data;
        break;
      case EWorkloadType.Distributed:
        (await workloadServiceApi.distributedApi.suspendDistributed(workload.id)).data;
        break;
    }
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function activateMultipleWorkloads(workloadIds: Array<string>): Promise<BatchResponse> {
  return (await workloadServiceApi.workloadsBatchApi.batchWorkloads({ ids: workloadIds, action: "resume" })).data;
}

async function activateWorkloadV2(workload: Workload): Promise<void> {
  try {
    if (workload.sourceApi === SourceApi.WorkloadsV2) {
      await workloadsManagerApi.workloadsV2Api.setWorkloadRunState(workload.id, { value: RunState.Resume });
      return;
    }

    switch (workload.type) {
      case EWorkloadType.Training:
        (await workloadServiceApi.trainingsApi.resumeTraining(workload.id)).data;
        break;
      case EWorkloadType.Workspace:
        (await workloadServiceApi.workspacesApi.resumeWorkspace(workload.id)).data;
        break;
      case EWorkloadType.Distributed:
        (await workloadServiceApi.distributedApi.resumeDistributed(workload.id)).data;
        break;
    }
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getWorkloadPods(
  workloadId: string,
  deleted?: boolean,
  offset?: number,
  limit?: number,
  search?: string,
  fetchAll?: boolean,
  filterBy?: Array<string>,
  sortOrder?: "asc" | "desc",
  sortBy?:
    | "name"
    | "clusterId"
    | "workloadId"
    | "revisionId"
    | "nodeName"
    | "currentNodepool"
    | "task"
    | "createdAt"
    | "completedAt"
    | "deletedAt",
): Promise<Array<Pod>> {
  const maxLimit = 200;

  if (!fetchAll) {
    const currentLimit = limit ? Math.min(limit, maxLimit) : maxLimit;
    const response = await workloadServiceApi.podsApi.getWorkloadPods(
      workloadId,
      deleted,
      offset ?? 0,
      currentLimit,
      search,
      sortOrder,
      sortBy,
      filterBy,
    );
    return response.data.pods;
  }

  // If fetchAll is true, paginate through all pages
  const allPods: Pod[] = [];
  let currentOffset = offset ?? 0;
  let hasMore = true;

  while (hasMore) {
    const response = await workloadServiceApi.podsApi.getWorkloadPods(
      workloadId,
      deleted,
      currentOffset,
      maxLimit,
      search,
      sortOrder,
      sortBy,
      filterBy,
    );

    allPods.push(...response.data.pods);

    // Continue until no more pages
    if (response.data.next) {
      currentOffset = response.data.next;
    } else {
      hasMore = false;
    }
  }

  return allPods;
}

async function getWorkloadElements(workloadId: string): Promise<WorkloadElement[]> {
  const response = await workloadServiceApi.workloadsApi.getWorkloadElements(workloadId);
  return response.data.elements;
}

async function getWorkloadElement(workloadId: string, elementId: string): Promise<WorkloadElementResponse> {
  try {
    const response = await workloadServiceApi.workloadsApi.getWorkloadElement(workloadId, elementId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function listPods(
  deleted?: boolean,
  offset?: number,
  limit?: number,
  sortOrder?: "asc" | "desc",
  sortBy?:
    | "name"
    | "clusterId"
    | "workloadId"
    | "nodeName"
    | "currentNodepool"
    | "createdAt"
    | "completedAt"
    | "deletedAt",
  filterBy?: Array<string>,
  verbosity?: PodVerbosity,
  completed?: string,
  search?: string,
): Promise<Array<Pod>> {
  const response = await workloadServiceApi.podsApi.listPods(
    deleted,
    offset,
    limit,
    sortOrder,
    sortBy,
    filterBy,
    verbosity,
    completed,
    search,
  );
  return response.data.pods;
}

async function countPods(
  deleted?: boolean,
  filterBy?: Array<string>,
  search?: string,
  completed?: string,
): Promise<number> {
  const response = await workloadServiceApi.podsApi.countPods(deleted, filterBy, search, completed);
  return response.data.count;
}

async function getWorkloadById(workloadId: string): Promise<WorkloadDetailed> {
  return (await workloadServiceApi.workloadsApi.getWorkload(workloadId)).data;
}

async function getWorkloadV2ById(workloadId: string): Promise<WorkloadV2GetResponse> {
  try {
    return (await workloadsManagerApi.workloadsV2Api.getWorkloadV2ById(workloadId)).data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

export type CrudWorkload = Workspace | Training | DistributedWorkload | Inference | DistributedInference;

async function getWorkloadCrudById(workloadId: string, workloadType: string): Promise<CrudWorkload> {
  try {
    switch (workloadType) {
      case EWorkloadType.Workspace:
        return (await workloadServiceApi.workspacesApi.getWorkspace(workloadId)).data;
      case EWorkloadType.Training:
        return (await workloadServiceApi.trainingsApi.getTraining(workloadId)).data;
      case EWorkloadType.Distributed:
        return (await workloadServiceApi.distributedApi.getDistributed(workloadId)).data;
      case EWorkloadType.Inference:
        return (await workloadServiceApi.inferencesApi.getInference(workloadId)).data;
      case EWorkloadType.DistributedInference:
        return (await workloadServiceApi.distributedInferencesApi.getDistributedInference(workloadId)).data;
      default:
        throw new Error(`Unsupported workload type: ${workloadType}`);
    }
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function patchWorkloadById(workloadId: string, patchFields: WorkloadPatchFields): Promise<Workload> {
  return controlPlaneService.patch(`${WORKLOADS_SERVICE_ENDPOINT}/${workloadId}`, patchFields);
}

async function getWorkloadHistory(workloadId: string, clusterUuid: string): Promise<GetWorkloadHistory200Response> {
  return controlPlaneService.get(`${WORKLOADS_SERVICE_ENDPOINT}/${workloadId}/history?clusterId=${clusterUuid}`);
}

async function countWorkloadsByName(name: string, clusterId: string, projectId?: number | null): Promise<number> {
  try {
    let filterBy = `name==${name},clusterId==${clusterId}`;
    if (Number.isFinite(projectId)) {
      filterBy = `${filterBy},projectId==${projectId}`;
    }

    const res = await workloadServiceApi.workloadsApi.countWorkloads(false, [filterBy], undefined, true);
    return res.data.count;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

// pending scheduling messages
async function getWorkloadPendingSchedulingMessages(workloadId: string): Promise<PendingSchedulingMessage[]> {
  try {
    const workload = await getWorkloadById(workloadId);
    return workload.pendingSchedulingMessages || [];
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

// new metrics apis
async function getWorkloadMetrics(
  workloadId: string,
  start: string,
  end: string,
  metricType: Array<WorkloadMetricType>,
  numberOfSamples?: number,
): Promise<MetricsResponse> {
  try {
    const res: AxiosResponse<MetricsResponse> = await workloadServiceApi.workloadsApi.getWorkloadMetrics(
      workloadId,
      metricType,
      start,
      end,
      numberOfSamples,
    );
    return res.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getWorkloadPodMetrics(
  workloadId: string,
  podId: string,
  metricType: Array<PodMetricType>,
  start: string,
  end: string,
  numberOfSamples?: number,
): Promise<MetricsResponse> {
  try {
    const res: AxiosResponse<MetricsResponse> = await workloadServiceApi.podsApi.getWorkloadPodMetrics(
      workloadId,
      podId,
      metricType,
      start,
      end,
      numberOfSamples,
    );
    return res.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getWorkloadElementMetrics(
  workloadId: string,
  elementId: string,
  metricType: Array<ElementMetricType>,
  start: string,
  end: string,
  numberOfSamples?: number,
): Promise<MetricsResponse> {
  try {
    const res: AxiosResponse<MetricsResponse> = await workloadServiceApi.workloadsApi.getWorkloadElementMetrics(
      workloadId,
      elementId,
      metricType,
      start,
      end,
      numberOfSamples,
    );
    return res.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

//telemetry
async function getWorkloadsTelemetry(
  telemetryType: WorkloadTelemetryType,
  clusterId?: string,
  nodepoolName?: string,
  departmentId?: string,
  projectId?: string,
  groupBy?: Array<EWorkloadTelemetryGroupByWithoutBucket>,
  phases?: Array<Phase>,
): Promise<TelemetryResponse> {
  try {
    const response = await workloadServiceApi.workloadsApi.getWorkloadsTelemetry(
      telemetryType,
      clusterId,
      nodepoolName,
      departmentId,
      projectId,
      groupBy,
      phases,
    );
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

//telemetry csv
async function getWorkloadsTelemetryCsv(
  telemetryType: WorkloadTelemetryType,
  clusterId?: string,
  nodepoolName?: string,
  departmentId?: string,
  groupBy?: Array<EWorkloadTelemetryGroupBy>,
): Promise<void> {
  try {
    const query: Record<string, string | number | Array<string>> = {
      telemetryType: telemetryType,
    };
    if (clusterId) {
      query["clusterId"] = clusterId;
    }
    if (nodepoolName) {
      query["nodepoolName"] = nodepoolName;
    }
    if (departmentId) {
      query["departmentId"] = departmentId;
    }
    if (groupBy && groupBy.length > 0) {
      query["groupBy"] = groupBy;
    }
    await fileService.downloadFile(WORKLOADS_TELEMETRY_ENDPOINT, FileType.csv, undefined, query, {});
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getUnifiedWorkload(unifiedRequest: UnifiedWorkloadRequest): Promise<UnifiedWorkload> {
  const response = await workloadServiceApi.unifiedApi.getUnifiedWorkload(unifiedRequest);
  return response.data;
}

/**
 * Determines whether a workload can be submitted to the specified cluster.
 *
 * @param {ClustersMinimalData} cluster - The cluster to check for workload submission.
 * @returns {ICanSubmitWorkload | null}
 *   - `null` if the cluster is available for workload submission.
 *   - An object with `info` if submission is not possible and the reason.
 */
function canSubmitWorkload(cluster: ClustersMinimalData): ICanSubmitWorkload | null {
  const isDisconnected = cluster.state === ClusterState.Disconnected || cluster.state === ClusterState.WaitingToConnect;

  if (isDisconnected) {
    return { info: "Cluster is not connected" };
  }

  if (!cluster.domain) {
    return { info: "cluster domain is missing or empty" };
  }

  return null;
}

/** Categories **/
async function listCategories(): Promise<WorkloadCategory[]> {
  try {
    const response = await workloadServiceApi.workloadPropertiesApi.listCategories();
    return response.data.categories;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
/** Priorities **/
async function getWorkloadClusterPriorities(clusterId: string): Promise<ClusterPriority[]> {
  const response = await workloadServiceApi.workloadPropertiesApi.getWorkloadClusterPriorities(clusterId);
  const priorities = response.data.priorities;
  return priorities.filter((priority: ClusterPriority) => priority.source === "Run.ai");
}

async function getWorkloadPriorities(
  sortOrder?: "asc" | "desc",
  sortBy?: "name" | "source" | "value",
  filterBy?: Array<string>,
): Promise<WorkloadPriority[]> {
  const response = await workloadServiceApi.workloadPropertiesApi.getWorkloadPriorities(sortOrder, sortBy, filterBy);
  return response.data.priorities;
}

async function getDefaultWorkloadPriorityByType(workloadType: EWorkloadName): Promise<WorkloadTypeConfig> {
  const response = await workloadServiceApi.workloadPropertiesApi.listWorkloadTypes();
  const workloadTypeName = workloadPriorityNameByWorkloadType[workloadType];
  return response.data.types.find((type: WorkloadTypeConfig) => type.name === workloadTypeName) as WorkloadTypeConfig;
}

async function getYamlSubmissionSetting(): Promise<YamlSubmissionSetting> {
  try {
    const response = await workloadsManagerApi.settingsApi.getWorkloadsSettingByKey(
      WorkloadsSettingsKeyEnum.YamlSubmission,
    );
    const data = response.data as GetYamlSubmission;
    return {
      enabled: data.value?.enabled ?? null,
    };
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateYamlSubmissionSetting(value: boolean): Promise<void> {
  try {
    const response = await workloadsManagerApi.settingsApi.updateWorkloadsSettingByKey(
      WorkloadsSettingsKeyEnum.YamlSubmission,
      {
        enabled: value,
      },
    );
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

export interface IApplicationSubmissionSetting {
  enabled: boolean | null;
}

async function getApplicationSubmissionSetting(): Promise<IApplicationSubmissionSetting> {
  try {
    const response = await workloadServiceApi.settingsApi.getWorkloadsServiceSettingByKey(
      WorkloadsServiceSettingsKeyEnum.ApplicationSubmission,
    );
    const data = response.data as GetWorkloadsSetting;
    return {
      enabled: data.value?.enabled ?? null,
    };
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateApplicationSubmissionSetting(value: boolean): Promise<void> {
  try {
    await workloadServiceApi.settingsApi.updateWorkloadsServiceSettingByKey(
      WorkloadsServiceSettingsKeyEnum.ApplicationSubmission,
      {
        enabled: value,
      },
    );
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

function getFallbackWorkloadPriorityByType(workloadType: EWorkloadName): string {
  const priorities = {
    workspace: "high",
    training: "low",
    inference: "very-high",
    distributed: "low",
    [EWorkloadName.DistributedInference]: "very-high",
  };
  return priorities[workloadType];
}

function getDefaultPreemptibilityByType(workloadType: EWorkloadName): Preemptibility {
  // Use swagger type for type safety - values match backend API exactly
  const defaults: Record<EWorkloadName, Preemptibility> = {
    [EWorkloadName.Workspace]: Preemptibility.NonPreemptible,
    [EWorkloadName.Training]: Preemptibility.Preemptible,
    [EWorkloadName.Inference]: Preemptibility.NonPreemptible,
    [EWorkloadName.Distributed]: Preemptibility.Preemptible,
    [EWorkloadName.DistributedInference]: Preemptibility.NonPreemptible,
  };
  return defaults[workloadType];
}

function getFrameworkForUrl(workloadType: string, distributedFramework?: string): FrameworkUrlType {
  if (workloadType === EWorkloadType.Distributed && distributedFramework) {
    return distributedFrameworkUrlMapping[distributedFramework] || "runai";
  }
  if (workloadType === EWorkloadType.Inference) {
    return "knative";
  }
  return "runai";
}

interface IExceedsAvailableQuotaOptions {
  preemptibility?: Preemptibility | null; // v2.24+ mode
  priorityValue?: number | null; // Legacy mode (< v2.24)
  currProject?: Project;
  selectedNodePoolsResources?: string[];
  totalGpuRequest?: number;
}

function exceedsAvailableQuota(options: IExceedsAvailableQuotaOptions): boolean {
  const { preemptibility, priorityValue, currProject, selectedNodePoolsResources, totalGpuRequest } = options;

  // Determine if non-preemptible
  let isNonPreemptible: boolean;

  if (preemptibility !== undefined && preemptibility !== null) {
    // NEW MODE: Using preemptibility enum (v2.24+)
    isNonPreemptible = preemptibility === Preemptibility.NonPreemptible;
  } else if (priorityValue !== undefined && priorityValue !== null) {
    // OLD MODE: Using priority value (< v2.24)
    isNonPreemptible = priorityValue >= 100;
  } else {
    // No value provided
    return false;
  }

  if (!isNonPreemptible) return false;

  if (!currProject || !selectedNodePoolsResources || selectedNodePoolsResources.length === 0) {
    return false;
  }

  const nodePoolQuotaStatuses = currProject.status?.nodePoolQuotaStatuses;
  if (!nodePoolQuotaStatuses || !currProject.resources) return false;

  const projectResourceByName = new Map<string, Resources>();
  currProject.resources.forEach((resource: Resources) => {
    if (resource.nodePool?.name) {
      projectResourceByName.set(resource.nodePool.name, resource);
    }
  });

  const nodePoolStatusByName = new Map<string, QuotaStatusNodePool>();
  nodePoolQuotaStatuses.forEach((status: QuotaStatusNodePool) => {
    if (status.nodePoolName) {
      nodePoolStatusByName.set(status.nodePoolName, status);
    }
  });

  const gpuRequest = totalGpuRequest || 0;

  // The workload only needs to fit in ONE node pool to run successfully.
  // Show warning only if ALL selected node pools would exceed quota.
  let checkedNodePoolCount = 0;
  let exceededNodePoolCount = 0;

  for (const selectedNodePoolName of selectedNodePoolsResources) {
    const nodePoolResource = projectResourceByName.get(selectedNodePoolName);
    if (!nodePoolResource) continue;

    const nodePoolStatus = nodePoolStatusByName.get(selectedNodePoolName);
    if (!nodePoolStatus) continue;

    checkedNodePoolCount++;

    const quota = nodePoolResource.gpu?.deserved || 0;
    const allocatedNonPreemptible = nodePoolStatus.allocatedNonPreemptible?.gpu || 0;

    if (quota < gpuRequest + allocatedNonPreemptible) {
      exceededNodePoolCount++;
    }
  }

  // Only show warning if we checked at least one node pool and ALL of them exceeded quota
  return checkedNodePoolCount > 0 && exceededNodePoolCount === checkedNodePoolCount;
}
