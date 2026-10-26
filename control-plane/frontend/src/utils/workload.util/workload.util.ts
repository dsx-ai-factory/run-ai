/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { Connection, Endpoint, Environment, Training, Workspace } from "@/swagger-models/workloads-client";
import type { WorkloadV2 } from "@/swagger-models/workloads-manager-client";
import {
  type HistoryRecord,
  HistoryRecordType,
  Phase,
  Source,
  SourceApi,
  type Workload,
  WorkloadSortFilterFields,
} from "@/swagger-models/workloads-client";
import {
  type Annotation,
  type AssetIdAndKind,
  type AssetsIds,
  type AssetsRef,
  type DatasourceRef,
  type DisplayedJob,
  type DistributedParams,
  type Inference,
  type InferenceV2,
  InternalConnectionType,
  type Label,
  type SpecificRunConnectionInfo,
  type SpecificRunParams,
  type ToolType,
  type TrainingCreationRequestV2,
  type TrainingV2,
  type WorkloadMeta,
  type WorkloadSupportedTypes,
  type WorkspaceCreationRequestV2,
  type WorkspaceV2,
  type ModelInference,
  type ModelInferenceV2,
  type DistributedData,
  type AssetsRefOptional,
  type EnvironmentVariableOfAsset,
} from "@/swagger-models/assets-service-client";
import type { EWorkloadEntity } from "@/models/workload.model";
import {
  EWorkloadAction,
  EWorkloadName,
  type IUIDistributed,
  type IUIDistributedMaster,
  type IUIWorkloadAssets,
  type IUIWorkloadCreation,
  type IUIWorkloadSpecificEnv,
  type IWorkloadMeta,
  WorkloadPhaseMap,
  type WorkloadServicePod,
  EWorkloadType,
  type IWorkloadSupportedRoutes,
  type IToolItem,
  EWorkloadCategory,
} from "@/models/workload.model";
import type { IInferenceList } from "@/models/inference.model";
import type { IUIVolume } from "@/models/data-source.model";
import {
  EHuggingFaceEnvVariableNames,
  NIM_IMAGE_PREFIX,
  NIM_WORKLOAD_LABEL,
  OLD_LLM_IMAGE_FOR_HUGGING_FACE,
  VLLM_WORKLOAD_LABEL,
} from "@/models/inference.model";
import type { IItemizedListItem } from "@/components/common/runai-itemized-list";
import { deepCopy, fallbackDefaultIfNullOrUndefined, omit, pick } from "@/utils/common.util";
import { tableUtil } from "@/utils/table.util";
import { ETableFilters, type IStatusColOptions } from "@/models/table.model";
import { allWorkloadsIconsMap, GENERIC_WORKLOAD_ICON } from "@/common/icons.constant";
import { getHuggingFaceIcon, getNvidiaImageBasedIcon } from "@/common/icon.util";
import type { IWorkspaceList } from "@/models/workspace.model";
import { UI_WORKSPACE_STATUS } from "@/common/status.constant";
import { STORAGE_KEYS } from "@/models/storage.model";
import { filterService } from "@/services/filter.service/filter.service";
import type { IAdvancedFilterModel, IEnumBasedFilterModel } from "@/models/filter.model";
import { EColumnFilterType, EFilterOperator } from "@/models/filter.model";
import type { ITrainingList } from "@/models/training.model";

import type { IWorkloadSpecificRunParams, SpecificRunParamsKeys } from "./workload.util.models";
import { storageUtil } from "@/utils/storage.util";
import { ENVIRONMENT_ROUTE_NAMES } from "@/router/environment.routes/environment.routes.names";
import { COMPUTE_RESOURCE_ROUTE_NAMES } from "@/router/compute-resource.routes/compute-resource.routes.names";
import { DATA_SOURCE_CREATION_ROUTE_MAP } from "@/router/data-source.routes/data-source.routes.names";
import type { Field } from "@/models/workload-phase-messages.model";
import { stringUtil } from "@/utils/string.util";

export interface IUIDistributedUISetup {
  distributedData: DistributedData | null;
  enableEditingMaster?: boolean;
  noMaster?: boolean;
}

export const workloadUtil = {
  getEmptyUIWorkloadCreation,
  getUIWorkloadSpecificEnv,
  getUIWorkloadAssets,
  getWorkloadAssetsFromUIWorkloadAssets,
  getWorkloadSpecificEnvFromUIWorkloadSpecificEnv,
  getWorkspaceCreationRequestV2,
  getTrainingCreationRequestV2,
  convertWorkloadToWorkloadUI,
  getDisabledToolTipText,
  getConnectivityMessage,
  getMissingDomainMessage,
  getWorkloadStatusColOptions,
  getPodStatusColOptions,
  sortPods,
  isWorkloadSourceControlPlane,
  isPhaseUpdateRecordType,
  getWorkloadIcon,
  isNIMWorkload,
  isVLLMWorkload,
  convertConnectionsToToolItems,
  getWorkspaceTrainingStatusColOptions,
  getRunningVsRequestedDisplayedValue,
  convertWorkloadTypesToString,
  isWorkloadSourceCli,
  getWorkloadConnectionsForGrid,
  getUnclaimedServingUrl,
  mapEndpointsToConnections,
  applyUserCreatedWorkloadsFilter,
  convertEnvVarToEnvVarUI,
  convertToWorkload,
  getWorkloadListModalStatusColOptions,
  adjustWorkloadPhaseFilter,
  isRunaiWorkload,
  isWorkloadSubmittedByV2,
  getSupportedRoutes,
  processCommonSpecificEnv,
  _getWorkloadUI,
  _getWorkloadConnectionsForInference,
  _getNimInferenceIcon,
  flattenWorkloadPhaseIssues,
};

function convertWorkloadTypesToString(workloadTypes: WorkloadSupportedTypes): string {
  // It's important to keep the distributed first in this array,
  // when distributed is supported it is the only type that can be supported
  const workloadsList: Array<string> = [
    EWorkloadName.Distributed,
    EWorkloadName.Workspace,
    EWorkloadName.Training,
    EWorkloadName.Inference,
  ];

  const supportedTypes = [];
  if (workloadTypes[EWorkloadName.Distributed]) {
    if (workloadTypes[EWorkloadName.Inference]) {
      // Distributed inference is a standalone workload type.
      // EWorkloadType.DistributedInference matches the label used on the workload list page.
      supportedTypes.push(EWorkloadType.DistributedInference);
    } else {
      supportedTypes.push(EWorkloadName.Training); // Distributed is an architecture type of training
    }
  } else {
    workloadsList.forEach((key: string) => {
      if (workloadTypes[key as keyof WorkloadSupportedTypes]) {
        supportedTypes.push(key);
      }
    });
  }

  const selectedString = supportedTypes.join(", ");
  return selectedString.charAt(0).toUpperCase() + selectedString.slice(1);
}

function getEmptyUIWorkloadCreation(specificEnv?: IUIWorkloadSpecificEnv): IUIWorkloadCreation {
  return {
    name: "",
    projectId: -1,
    namespace: "",
    clusterId: "",
    assets: {
      environment: "",
      compute: "",
    },
    specificEnv: {
      ...specificEnv,
      backoffLimit: 6,
    },
  };
}

function getUIWorkloadAssets(assets: AssetsIds | undefined, uiVolumes?: Array<IUIVolume>): IUIWorkloadAssets {
  return {
    environment: assets?.environment || null,
    compute: assets?.compute || null,
    datasources: assets?.datasources,
    uiVolumes,
  };
}

function getUIWorkloadAssetsRef(
  assets: AssetsRef | AssetsRefOptional | undefined | null,
  uiVolumes?: Array<IUIVolume>,
): IUIWorkloadAssets {
  return {
    environment: assets?.environment?.id || null,
    compute: assets?.compute?.id || null,
    datasources: assets?.datasources?.map((dataSource: DatasourceRef) => {
      const dsAsset: AssetIdAndKind = { id: dataSource.id, kind: dataSource.kind };
      if (dataSource.overrides) dsAsset.overrides = dataSource.overrides;
      return dsAsset;
    }),
    uiVolumes,
  };
}

function getUIWorkloadSpecificEnv(specificEnv: SpecificRunParams | undefined = {}): IUIWorkloadSpecificEnv {
  const workloadSpecificEnv: IUIWorkloadSpecificEnv = {
    args: specificEnv.args,
    command: specificEnv.command,
    runAsUid: specificEnv.runAsUid || null,
    runAsGid: specificEnv.runAsGid || null,
    supplementalGroups: specificEnv.supplementalGroups,
    nodeType: specificEnv.nodeType || null,
    allowOverQuota: specificEnv.allowOverQuota || null,
    nodePools: specificEnv.nodePools ? specificEnv.nodePools : null,
    autoDeletionTimeAfterCompletionSeconds: fallbackDefaultIfNullOrUndefined(
      specificEnv.autoDeletionTimeAfterCompletionSeconds,
      null,
    ),
    connections: specificEnv.connections,
    environmentVariables: specificEnv.environmentVariables || undefined,
    annotations: specificEnv.annotations?.map((ann: Annotation) => {
      const itemized: IItemizedListItem = {
        ...ann,
        name: ann.name || "",
        value: ann.value || "",
        exclude: ann.exclude || undefined,
      };

      return itemized;
    }),
    labels: specificEnv.labels?.map((lbl: Label) => {
      const itemized: IItemizedListItem = {
        ...lbl,
        name: lbl.name || "",
        value: lbl.value || "",
        exclude: lbl.exclude || undefined,
      };

      return itemized;
    }),
    backoffLimit: Number.isInteger(specificEnv.backoffLimit) ? specificEnv.backoffLimit : 6,
    tolerations: specificEnv.tolerations,
    podAffinity: specificEnv.podAffinity,
    parallelism: specificEnv.parallelism,
    completions: specificEnv.completions,
    terminationGracePeriodSeconds: specificEnv.terminationGracePeriodSeconds,
  };

  return workloadSpecificEnv;
}

function convertEnvVarToEnvVarUI(environmentVariables: Array<EnvironmentVariableOfAsset>): Array<IItemizedListItem> {
  return environmentVariables.map((env: EnvironmentVariableOfAsset) => {
    return {
      ...env,
      name: env.name || "",
      value: env.value || "",
      exclude: env.exclude || undefined,
    };
  });
}

function getWorkloadAssetsFromUIWorkloadAssets(assets: IUIWorkloadAssets, workloadVolumes?: Array<string>): AssetsIds {
  if (!assets.environment) throw new Error("assets.environment is required");

  const retAssets: AssetsIds = {
    environment: assets.environment,
    compute: assets.compute,
  };

  if (assets.datasources) {
    retAssets.datasources = assets.datasources;
  }

  if (workloadVolumes) {
    retAssets.workloadVolumes = workloadVolumes;
  }

  return retAssets;
}

function getWorkloadSpecificEnvFromUIWorkloadSpecificEnv(specificEnv: IUIWorkloadSpecificEnv): SpecificRunParams {
  const retSpecificEnv = processCommonSpecificEnv(specificEnv, ["autoScaleData", "servingPortAccess"]);
  return retSpecificEnv;
}

function getWorkspaceCreationRequestV2(
  workspace: IUIWorkloadCreation,
  workloadVolumes?: Array<string>,
): WorkspaceCreationRequestV2 {
  if (!workspace.assets.environment) throw "Environment details are missing";
  const workloadCreationRequest: WorkspaceCreationRequestV2 = {
    meta: {
      name: workspace.name,
      projectId: workspace.projectId?.toString(),
      clusterId: workspace.clusterId,
    },
    assets: getWorkloadAssetsFromUIWorkloadAssets(workspace.assets, workloadVolumes),
    specificRunParams: getWorkloadV2RequestSpecificRunParams(workspace.specificEnv),
  };
  return workloadCreationRequest;
}

function getTrainingCreationRequestV2(
  training: IUIWorkloadCreation,
  workloadVolumes?: Array<string>,
  masterWorkloadVolumes?: Array<string>,
): TrainingCreationRequestV2 {
  if (!training.assets.environment) throw "Environment details are missing";
  const workloadCreationRequest: TrainingCreationRequestV2 = {
    meta: {
      name: training.name,
      projectId: training.projectId?.toString(),
      clusterId: training.clusterId,
    },
    assets: getWorkloadAssetsFromUIWorkloadAssets(training.assets, workloadVolumes),
    specificRunParams: getWorkloadV2RequestSpecificRunParams(training.specificEnv),
  };

  if (training.distributed) {
    workloadCreationRequest.distributed = _getDistributedObjectForTrainingV2Request(
      training.distributed,
      masterWorkloadVolumes,
      !training.enableEditingMaster,
    );
  }

  return workloadCreationRequest;
}

function processCommonSpecificEnv(
  specificEnv: IUIWorkloadSpecificEnv,
  omitFields: SpecificRunParamsKeys,
): SpecificRunParams {
  const retSpecificEnv: SpecificRunParams = {
    ...omit(specificEnv, [...omitFields, "environmentVariables", "annotations", "labels"]),
  };

  retSpecificEnv.args ||= null;
  retSpecificEnv.command ||= null;
  retSpecificEnv.runAsUid = retSpecificEnv.runAsUid ? +retSpecificEnv.runAsUid : undefined;
  retSpecificEnv.runAsGid = retSpecificEnv.runAsGid ? +retSpecificEnv.runAsGid : undefined;
  retSpecificEnv.supplementalGroups = Array.isArray(retSpecificEnv.supplementalGroups)
    ? retSpecificEnv.supplementalGroups.join(",")
    : retSpecificEnv.supplementalGroups;

  if (specificEnv.environmentVariables) {
    retSpecificEnv.environmentVariables = specificEnv.environmentVariables;
  }

  if (specificEnv.annotations) {
    retSpecificEnv.annotations = specificEnv.annotations.map((env: IItemizedListItem) =>
      pick(env, "name", "value", "exclude"),
    );
  }

  if (specificEnv.labels) {
    retSpecificEnv.labels = specificEnv.labels.map((env: IItemizedListItem) => pick(env, "name", "value", "exclude"));
  }

  return retSpecificEnv;
}

function getWorkloadV2RequestSpecificRunParams(specificEnv: IUIWorkloadSpecificEnv): IWorkloadSpecificRunParams {
  const retSpecificEnv = processCommonSpecificEnv(specificEnv, ["autoScaleData"]);

  return retSpecificEnv;
}

function _getDistributedObjectForTrainingV2Request(
  distInfo: IUIDistributed,
  masterWorkloadVolumes?: Array<string>,
  sameAsWorker?: boolean,
): DistributedParams {
  if (distInfo.master && !distInfo.noMaster && !distInfo.master.assets.environment)
    throw "Environment details are missing";
  const distParams: DistributedParams = {
    numWorkers: distInfo.numWorkers || 0,
    framework: distInfo.distFramework,
    master: null,
    // if there is no master (only workers) -> do not send master at all
  };

  if (distInfo.master && !distInfo.noMaster) {
    if (sameAsWorker) {
      // if master and worker are equal we only send "sameAsWorker: true"
      distParams.master = { sameAsWorker: true };
      return distParams;
    }
    // master details should only be sent if it is different than the worker (i.e sameAsWorker is falsey)
    distParams.master = {
      assets: getWorkloadAssetsFromUIWorkloadAssets(distInfo.master.assets, masterWorkloadVolumes),
      specificRunParams: getWorkloadV2RequestSpecificRunParams(distInfo.master.specificEnv),
    };
  }

  return distParams;
}

export type TWorkloadsToConvert = WorkspaceV2 & TrainingV2 & InferenceV2;

function convertWorkloadToWorkloadUI(workload: TWorkloadsToConvert, uiVolumes?: Array<IUIVolume>): IUIWorkloadCreation {
  const workloadMeta: IWorkloadMeta = {
    name: workload.meta.name,
    projectId: +workload.meta.projectId,
    namespace: "",
    clusterId: workload.meta.clusterId,
  };

  const dataSources: AssetIdAndKind[] =
    workload.assets.datasources?.map((ds: DatasourceRef): AssetIdAndKind => {
      const dsAsset: AssetIdAndKind = {
        id: ds.id,
        kind: ds.kind,
      };
      if (ds.overrides) dsAsset.overrides = ds.overrides;
      return dsAsset;
    }) || [];

  const assets = {
    environment: workload.assets.environment.id,
    compute: workload.assets.compute?.id || null,
    datasources: dataSources,
  };

  // For copied workload we need to reset the connection
  const specificEnv: SpecificRunParams = workload.specificRunParams ? deepCopy(workload.specificRunParams) : {};
  if (specificEnv.connections) {
    specificEnv.connections = _emptyConnectionsCustomInfo(specificEnv.connections);
  }

  let distributedSetup: IUIDistributedUISetup | undefined;

  if (workload.distributed) {
    distributedSetup = { distributedData: workload.distributed };
    if (workload.distributed.master && Object.keys(workload.distributed.master).length) {
      if (
        workload.distributed.master.sameAsWorker ||
        !Object.keys(workload.distributed.master.specificRunParams || {}).length
      ) {
        distributedSetup.enableEditingMaster = false;
      } else {
        // if we have data which is specific for master we make enableEditingMaster true by default
        distributedSetup.enableEditingMaster = true;
      }
    } else {
      distributedSetup.noMaster = true;
    }
  }

  return _getWorkloadUI(workloadMeta, assets, specificEnv, uiVolumes, distributedSetup);
}

function _getWorkloadUI(
  meta: IWorkloadMeta,
  assets: AssetsIds,
  specificEnv: SpecificRunParams = {},
  uiVolumes?: Array<IUIVolume>,
  distributedSetup?: IUIDistributedUISetup,
): IUIWorkloadCreation {
  const workloadUI: IUIWorkloadCreation = {
    name: meta.name,
    namespace: meta.namespace,
    clusterId: meta.clusterId,
    projectId: meta.projectId,
    assets: getUIWorkloadAssets(assets, uiVolumes),
    specificEnv: getUIWorkloadSpecificEnv(specificEnv),
  };

  if (distributedSetup?.distributedData) {
    let master: IUIDistributedMaster | null;
    if (!distributedSetup.distributedData.master || distributedSetup.noMaster) {
      master = null;
    } else {
      master = {
        assets: getUIWorkloadAssetsRef(distributedSetup.distributedData.master?.assets),
        specificEnv: getUIWorkloadSpecificEnv(distributedSetup.distributedData.master?.specificRunParams || {}),
      };
    }

    workloadUI.distributed = {
      numWorkers: distributedSetup.distributedData.numWorkers,
      distFramework: distributedSetup.distributedData.framework,
      master,
      noMaster: distributedSetup.noMaster,
    };

    workloadUI.enableEditingMaster = distributedSetup.enableEditingMaster;
  }

  return workloadUI;
}

function _emptyConnectionsCustomInfo(connections: Array<SpecificRunConnectionInfo>): Array<SpecificRunConnectionInfo> {
  return connections.map((connection: SpecificRunConnectionInfo) => {
    if (connection.nodePort) {
      return {
        ...connection,
        nodePort: null,
      };
    } else if (connection.externalUrl) {
      return {
        ...connection,
        externalUrl: null,
      };
    }
    return connection;
  });
}

function getDisabledToolTipText(action: EWorkloadAction, entity: EWorkloadEntity): string {
  switch (action) {
    case EWorkloadAction.activate:
      return `The selected ${entity} is already active or in the process of being activated`;
    case EWorkloadAction.run:
      return `The selected ${entity} is already running or being prepared to run`;
    case EWorkloadAction.stop:
      return `The selected ${entity} is already stopped or in the process of being stopped`;
    case EWorkloadAction.connect:
      return `To connect, the selected ${entity} must be running`;
    default:
      return "Action for the selected workload is not authorized";
  }
}

function getConnectivityMessage(errorCode?: number): string {
  const unauthorizedProblem = "The API server is not configured correctly. Contact your administrator"; // when error code is 401
  const connectivityProblem = (errorCode?: number): string =>
    errorCode
      ? `There are issues with your connection to the cluster. Make sure you're using your organization's VPN, or contact your administrator (error code: ${errorCode})` // when error code is 403, 404 or 50X
      : "There are issues with your connection to the cluster. Contact your administrator.";
  switch (errorCode) {
    case 401:
      return unauthorizedProblem;
    case -1:
      return connectivityProblem();
    default:
      return connectivityProblem(errorCode);
  }
}

function getMissingDomainMessage(): string {
  return "The cluster domain is not defined. Contact your administrator.";
}

function getWorkspaceTrainingStatusColOptions(job?: DisplayedJob, tooltipText = ""): IStatusColOptions {
  const statusOptions = UI_WORKSPACE_STATUS[job?.status || ""];
  if (job?.msSinceLastStatusUpdate) {
    statusOptions.statusUpdatedTimeInMs = Date.now() - job?.msSinceLastStatusUpdate;
  }
  return tableUtil.getStatusColOptions(statusOptions, tooltipText);
}

//new workloads table
function getWorkloadStatusColOptions(workload: Workload, phaseMessage?: string): IStatusColOptions {
  const statusOptions: IStatusColOptions = { ...WorkloadPhaseMap[workload.phase] };
  statusOptions.statusUpdatedTimeInMs = workload.phaseUpdatedAt
    ? new Date(workload.phaseUpdatedAt).getTime()
    : undefined;
  return tableUtil.getStatusColOptions(statusOptions, phaseMessage);
}
function getWorkloadListModalStatusColOptions(phase: Phase): IStatusColOptions {
  if (!WorkloadPhaseMap[phase]) {
    return tableUtil.getStatusColOptions({ status: phase, displayAnimation: false });
  }
  const statusOptions: IStatusColOptions = { ...WorkloadPhaseMap[phase] };
  return tableUtil.getStatusColOptions(statusOptions);
}

function getPodStatusColOptions(pod: WorkloadServicePod): IStatusColOptions {
  const statusOptions: IStatusColOptions = tableUtil.getStatusColOptions(WorkloadPhaseMap[pod.k8sPhase as Phase]);
  statusOptions.statusUpdatedTimeInMs = pod.k8sPhaseUpdatedAt ? new Date(pod.k8sPhaseUpdatedAt).getTime() : undefined;
  if (statusOptions.status === "-" && pod.k8sPhase) {
    statusOptions.status = pod.k8sPhase;
  }
  return statusOptions;
}

function sortPods(pods: WorkloadServicePod[]): WorkloadServicePod[] {
  return [...pods].sort((a, b) => {
    const nameA = a.name || "";
    const nameB = b.name || "";

    const isMasterA = nameA.includes("-master-");
    const isMasterB = nameB.includes("-master-");

    if (isMasterA && !isMasterB) return -1;
    if (!isMasterA && isMasterB) return 1;

    // Extract numeric suffix for natural sorting (e.g., "worker-10" -> 10)
    const extractNumber = (name: string): number => {
      const match = name.match(/-(\d+)$/);
      return match ? parseInt(match[1], 10) : -1;
    };

    const numA = extractNumber(nameA);
    const numB = extractNumber(nameB);

    if (numA !== -1 && numB !== -1) {
      return numA - numB;
    }

    return nameA.localeCompare(nameB);
  });
}

function isWorkloadSourceControlPlane(workload: Workload): boolean {
  return workload.source === Source.ControlPlane;
}
function isWorkloadSourceCli(workload: Workload): boolean {
  return workload.source === Source.Cli;
}
function isRunaiWorkload(workload: Workload): boolean {
  return isWorkloadSourceCli(workload) || isWorkloadSourceControlPlane(workload);
}
function isWorkloadSubmittedByV2(workload: Workload): boolean {
  return workload.sourceApi === SourceApi.WorkloadsV2;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function isWorkload(workload: any): workload is Workload {
  return workload.name !== undefined;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function _isWorkspaceList(workload: any): workload is IWorkspaceList {
  return !!workload.job;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function _isTrainingList(workload: any): workload is ITrainingList {
  return !!workload.job;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function _isWorkspaceV2(workload: any): workload is WorkspaceV2 {
  return workload.meta.workloadId !== undefined;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function _isTrainingV2(workload: any): workload is TrainingV2 {
  return workload.meta.workloadId !== undefined;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function _isInferenceV2(workload: any): workload is InferenceV2 | ModelInferenceV2 {
  return workload.meta.workloadId !== undefined;
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function _isHybridWorkload(workload: any): workload is Workspace | Training {
  return workload.workloadId !== undefined;
}

function isPhaseUpdateRecordType(historyRecord: HistoryRecord): boolean {
  return historyRecord.meta.type === HistoryRecordType.PhaseUpdate;
}

function convertToWorkload(
  workload:
    | Workload
    | IWorkspaceList
    | ITrainingList
    | WorkspaceV2
    | TrainingV2
    | IInferenceList
    | Inference
    | InferenceV2
    | ModelInference
    | ModelInferenceV2
    | Workspace
    | Training
    | WorkloadV2,
  workloadType?: EWorkloadType,
): Workload {
  if (_isHybridWorkload(workload)) {
    return _convertHybridToWorkload(workload);
  } else if (isWorkload(workload)) {
    return workload;
  } else if (_isWorkloadV2(workload)) {
    return _convertWorkloadV2ToWorkload(workload);
  } else if (workloadType === EWorkloadType.Inference) {
    return _isInferenceV2(workload)
      ? _convertWorkloadMetaToWorkload(workload.meta)
      : _convertInferenceV1WorkloadToWorkload(workload as Inference | ModelInference);
  } else if (_isWorkspaceV2(workload) || _isTrainingV2(workload) || _isInferenceV2(workload)) {
    return _convertWorkloadMetaToWorkload(workload.meta);
  } else if (_isTrainingList(workload) || _isWorkspaceList(workload)) {
    return _convertDisplayedJobToWorkload(workload.job as DisplayedJob);
  } else {
    throw new Error(`Failed to convert workload to Workload ${workload}`);
  }
}

function _convertInferenceV1WorkloadToWorkload(workload: Inference | ModelInference): Workload {
  return {
    category: EWorkloadCategory.Inference,
    priority: null,
    source: Source.ControlPlane,
    conditions: [{ status: "", type: "", lastTransitionTime: null }],
    priorityClassName: "",
    type: String(EWorkloadType.Inference),
    name: workload.meta.name,
    id: workload.meta.id,
    clusterId: workload.meta.clusterId,
    projectName: "",
    projectId: String(workload.meta.projectId),
    departmentName: "",
    departmentId: "",
    namespace: "",
    createdAt: "",
    phase: Phase.Creating,
    k8sPhase: "",
    tenantId: 0,
    runningPods: 0,
    phaseUpdatedAt: "",
    k8sPhaseUpdatedAt: "",
    updatedAt: "",
    deletedAt: null,
    submittedBy: workload.meta.createdBy,
  };
}

function _isWorkloadV2(workload: unknown): workload is WorkloadV2 {
  return typeof workload === "object" && workload !== null && "metadata" in workload && "workloadManifest" in workload;
}

function _convertWorkloadV2ToWorkload(workload: WorkloadV2): Workload {
  const metadata = workload.metadata;
  return {
    source: Source.ControlPlane,
    conditions: [{ status: "", type: "", lastTransitionTime: null }],
    priorityClassName: metadata.priority || "",
    priority: null, // WorkloadV2 uses string priority names, not numeric values
    category: metadata.category || "",
    type: metadata.gvk?.kind || "",
    name: metadata.name,
    id: metadata.id,
    clusterId: metadata.clusterId || "",
    projectName: metadata.projectName || "",
    projectId: metadata.projectId,
    departmentName: metadata.departmentName || "",
    departmentId: metadata.departmentId || "",
    namespace: "",
    createdAt: metadata.createdAt || "",
    phase: Phase.Creating,
    k8sPhase: "",
    tenantId: metadata.tenantId || 0,
    runningPods: 0,
    phaseUpdatedAt: "",
    k8sPhaseUpdatedAt: "",
    updatedAt: "",
    deletedAt: null,
    submittedBy: metadata.createdBy || "",
  };
}

function _convertDisplayedJobToWorkload(displayedJob: DisplayedJob): Workload {
  return {
    category: EWorkloadCategory.Workspace,
    priority: null,
    source: Source.ControlPlane,
    conditions: [{ status: displayedJob.status || "", type: "", lastTransitionTime: null }],
    priorityClassName: "",
    type: displayedJob.jobType || "",
    name: displayedJob.jobName || "",
    id: displayedJob.JobId || "",
    clusterId: displayedJob.clusterId || "",
    projectName: displayedJob.project || "",
    projectId: "",
    namespace: "",
    createdAt: displayedJob.creationTime || "",
    phase: Phase.Creating,
    k8sPhase: "",
    tenantId: 0,
    runningPods: 0,
    departmentId: "",
    departmentName: "",
    phaseUpdatedAt: "",
    k8sPhaseUpdatedAt: "",
    updatedAt: "",
    deletedAt: null,
    submittedBy: displayedJob.user,
  };
}

function _convertWorkloadMetaToWorkload(workloadMeta: WorkloadMeta): Workload {
  return {
    category: EWorkloadCategory.Workspace,
    priority: null,
    phase: Phase.Creating,
    id: workloadMeta.workloadId,
    name: workloadMeta.name,
    clusterId: workloadMeta.clusterId,
    createdAt: workloadMeta.createdAt,
    departmentId: workloadMeta.departmentId || "",
    projectId: workloadMeta.projectId,
    submittedBy: workloadMeta.createdBy,
    conditions: [],
    deletedAt: null,
    departmentName: "",
    k8sPhase: "",
    k8sPhaseUpdatedAt: "",
    namespace: "",
    phaseUpdatedAt: "",
    priorityClassName: "",
    projectName: "",
    runningPods: 0,
    source: Source.ControlPlane,
    tenantId: 0,
    type: "",
    updatedAt: "",
  };
}

function _convertHybridToWorkload(workload: Workspace | Training): Workload {
  return {
    category: EWorkloadCategory.Workspace,
    priority: null,
    phase: Phase.Creating,
    id: workload.workloadId,
    name: workload.name,
    clusterId: workload.clusterId,
    createdAt: workload.createdAt,
    departmentId: workload.departmentId || "",
    projectId: workload.projectId,
    submittedBy: workload.createdBy,
    conditions: [],
    deletedAt: null,
    departmentName: "",
    k8sPhase: "",
    k8sPhaseUpdatedAt: "",
    namespace: "",
    phaseUpdatedAt: "",
    priorityClassName: "",
    projectName: "",
    runningPods: 0,
    source: Source.ControlPlane,
    tenantId: 0,
    type: "",
    updatedAt: "",
  };
}

function _getDistributedFrameworkIcon(workload: Workload): undefined | string {
  if (!workload.distributedFramework) return undefined;
  return allWorkloadsIconsMap[workload.distributedFramework]
    ? allWorkloadsIconsMap[workload.distributedFramework]
    : undefined;
}

function _getDistributedInferenceIcon(workload: Workload): undefined | string {
  if (workload.type !== EWorkloadType.DistributedInference) return undefined;
  return GENERIC_WORKLOAD_ICON;
}

function _getToolTypeIcon(workload: Workload): undefined | string {
  const connections = workload.environments?.[0]?.connections;
  const connection = connections?.find(
    (conn) => conn.toolType && conn.connectionType !== InternalConnectionType.ServingPort,
  );

  if (!connection?.toolType) return undefined;
  return allWorkloadsIconsMap[connection.toolType];
}

function _getDefaultIcon(workload: Workload): string {
  return allWorkloadsIconsMap[workload.type] || allWorkloadsIconsMap.Unknown;
}

function isNIMWorkload(workload: Workload): boolean {
  const additionalFields = workload.additionalFields;
  const explicitNimValue = additionalFields?.nimWorkload ?? additionalFields?.[NIM_WORKLOAD_LABEL];
  if (explicitNimValue === "true" || explicitNimValue === true) return true;

  const explicitVllmValue = additionalFields?.vllmWorkload ?? additionalFields?.[VLLM_WORKLOAD_LABEL];
  if (explicitVllmValue === "true" || explicitVllmValue === true) return false;

  const imageName = workload.images?.[0];
  return !!(imageName && imageName.includes(NIM_IMAGE_PREFIX));
}

function isVLLMWorkload(workload: Workload): boolean {
  if (workload.type !== EWorkloadType.Inference) return false;

  const additionalFields = workload.additionalFields;
  const explicitNimValue = additionalFields?.nimWorkload ?? additionalFields?.[NIM_WORKLOAD_LABEL];
  if (explicitNimValue === "true" || explicitNimValue === true) return false;

  const explicitVllmValue = additionalFields?.vllmWorkload ?? additionalFields?.[VLLM_WORKLOAD_LABEL];
  return explicitVllmValue === "true" || explicitVllmValue === true;
}

function _getNimInferenceIcon(workload: Workload): string | undefined {
  return isNIMWorkload(workload) ? allWorkloadsIconsMap.NvidiaNim : undefined;
}

function _getNvcfIcon(workload: Workload): string | undefined {
  return workload.type === "NVCF Function" ? allWorkloadsIconsMap.Nvidia : undefined;
}

function _getHuggingFaceInferenceIcon(workload: Workload): string | undefined {
  if (workload.type !== EWorkloadType.Inference) return undefined;
  if (!workload.images || !workload.images[0]) return undefined;

  const workloadImage = workload.images[0];

  const isFlatWorkload = !workload?.environments?.[0]?.id;
  if (isFlatWorkload) {
    return getHuggingFaceIcon(workload.type, workloadImage);
  }

  const isRunaiModelExist: string | undefined =
    workload.environmentVariables && workload.environmentVariables[EHuggingFaceEnvVariableNames.RUNAI_MODEL];
  const isRunaiModelNameExist: string | undefined =
    workload.environmentVariables && workload.environmentVariables[EHuggingFaceEnvVariableNames.RUNAI_MODEL_NAME];
  if (workloadImage.includes(OLD_LLM_IMAGE_FOR_HUGGING_FACE) && isRunaiModelExist && !isRunaiModelNameExist) {
    return allWorkloadsIconsMap.HuggingFace;
  } else {
    return undefined;
  }
}

function getWorkloadIcon(workload: Workload): string {
  // Check for image-based icons first (NVIDIA PyTorch, NVIDIA general, etc.)
  const image = workload.images?.[0];
  const imageBasedIcon = getNvidiaImageBasedIcon(image);
  if (imageBasedIcon) {
    return imageBasedIcon;
  }

  const distributedIcon = _getDistributedFrameworkIcon(workload);
  if (distributedIcon) return distributedIcon;

  const distributedInferenceIcon = _getDistributedInferenceIcon(workload);
  if (distributedInferenceIcon) return distributedInferenceIcon;

  const nimIcon = _getNimInferenceIcon(workload);
  if (nimIcon) return nimIcon;

  const nvcfIcon = _getNvcfIcon(workload);
  if (nvcfIcon) return nvcfIcon;

  const huggingFaceIcon = _getHuggingFaceInferenceIcon(workload);
  if (huggingFaceIcon) return huggingFaceIcon;

  const toolTypeIcon = _getToolTypeIcon(workload);
  return toolTypeIcon ? toolTypeIcon : _getDefaultIcon(workload);
}

function convertConnectionsToToolItems(connections: Connection[]): IToolItem[] {
  return connections.map((connection: Connection) => ({
    toolType: connection.toolType as ToolType,
    toolName: connection.name,
    url: connection.url || "",
    authorizedUsers: connection.authorizedUsers || undefined,
  }));
}

function getRunningVsRequestedDisplayedValue(workload: Workload): string {
  if (workload.requestedPods?.number) {
    return `${workload.runningPods}/${workload.requestedPods.number}`;
  } else if (workload.requestedPods?.min !== undefined && workload.requestedPods.max) {
    return `${workload.runningPods}/${workload.requestedPods.min}-${workload.requestedPods.max}`;
  } else if (workload.requestedPods?.parallelism) {
    return `${workload.runningPods}/${workload.requestedPods.parallelism}`;
  } else if (workload.requestedPods?.completions && !workload.requestedPods?.parallelism) {
    return `${workload.runningPods}/1`;
  }
  return "-";
}

/**
 *
 * @param workload - current workload
 * @param connections current connections in the workload
 * @param clusterDomain For inference workloads we set the connections URL to the external URL. This sets the same URL for all the connections. This need to be changed in the future.
 * @returns Connection[] - updated connections
 */
function _getWorkloadConnectionsForInference(
  workload: Workload,
  connections: Connection[],
  clusterDomain?: string,
): Connection[] {
  let currConnections = [...connections];

  const servingPortConnection: Connection | undefined = connections.find(
    (c: Connection) => c.connectionType === InternalConnectionType.ServingPort,
  );

  if (!servingPortConnection && workload.urls && workload.urls.length > connections.length) {
    const servingUrl = getUnclaimedServingUrl(workload.urls, connections);
    if (servingUrl) {
      currConnections.push({
        url: servingUrl,
        name: "Serving Endpoint",
        toolType: "custom",
        connectionType: InternalConnectionType.ServingPort,
      });
    }
  }

  if (clusterDomain) {
    currConnections = currConnections.map((c: Connection) => {
      if (c.connectionType !== InternalConnectionType.ExternalUrl) return c;
      const workloadName = workload.name;
      const projectName = workload.projectName;
      return { ...c, url: `${clusterDomain}/${projectName}/${workloadName}` };
    });
  }

  return currConnections;
}

/**
 * Clusters may report the ServingPort connection without a URL. The serving URL is then the only
 * entry of `urls` that no connection (and none of `knownUrls`, e.g. the spec's exposed URLs) claims.
 * The order of `urls` differs between workload types, so when none or several entries are unclaimed
 * this returns undefined rather than guessing, to avoid showing a tool URL as the serving endpoint.
 *
 * @param urls - the workload URLs
 * @param connections - the workload connections
 * @param knownUrls - other URLs that belong to known connections
 * @returns string | undefined - the serving URL, if it can be identified
 */
function getUnclaimedServingUrl(
  urls: string[] | null | undefined,
  connections: Connection[],
  knownUrls: Array<string | null | undefined> = [],
): string | undefined {
  const claimedUrls = new Set([...connections.map((c: Connection) => c.url), ...knownUrls].filter(Boolean));
  const unclaimedUrls = (urls ?? []).filter((url: string) => !claimedUrls.has(url));
  return unclaimedUrls.length === 1 ? unclaimedUrls[0] : undefined;
}
/**
 *
 * @param workload
 * @param clusterDomain For inference workloads we set the connections URL to the external URL. This sets the same URL for all the connections. This need to be changed in the future.
 * @returns Connection[]
 */
function mapEndpointsToConnections(endpoints: Endpoint[]): Connection[] {
  return endpoints.map((endpoint) => ({
    name: endpoint.host || "",
    toolType: "",
    connectionType: "",
    url: endpoint.url,
    containerPort: endpoint.port ? endpoint.port : undefined,
  }));
}

function getWorkloadConnectionsForGrid(workload: Workload, clusterDomain?: string): Connection[] {
  let connections: Connection[];
  if (workload.environments?.length) {
    connections = workload.environments
      .reduce((acc: Connection[], env: Environment) => {
        acc = [...acc, ...(env.connections || [])];
        return acc;
      }, [])
      .filter((c) => {
        // Old clusters(older than 2.17) doesn't report the url for the serving port connection
        // In this case we want to filter it from the connections and add it later hard coded.
        return c.connectionType !== InternalConnectionType.ServingPort || !!c.url;
      });
  } else if (workload.endpoints?.length) {
    // Slim 2.25+ projection — used only when environments[].connections[] is absent.
    // Lacks name/toolType/connectionType/authorizationType metadata.
    connections = mapEndpointsToConnections(workload.endpoints);
  } else {
    connections = workload.externalConnections || [];
  }

  return workload.type === EWorkloadType.Inference
    ? _getWorkloadConnectionsForInference(workload, connections, clusterDomain)
    : connections;
}

/**
 * Applies a filter to the workloads based on the user who created them.
 * This function is used to filter the workloads displayed to the user,
 * showing only those that were created by the currently logged-in user.
 *
 * @param {string} userName - The name of the user for whom to filter the workloads.
 */
function applyUserCreatedWorkloadsFilter(shouldApplyFilter: boolean, userName: string): void {
  const isFirstLogin = storageUtil.get(STORAGE_KEYS.IS_FIRST_LOGIN);

  if (isFirstLogin) {
    const filterBy = filterService.loadFilters(window.location, ETableFilters.WORKLOAD, {});
    if (shouldApplyFilter) {
      const submittedByFilter = {
        term: userName,
        value: EFilterOperator.Equals,
        label: "Created by",
        name: WorkloadSortFilterFields.SubmittedBy,
        type: EColumnFilterType.FreeText,
      };
      filterService.setColumnAdvancedFilter(filterBy, submittedByFilter, ETableFilters.WORKLOAD);
    } else {
      filterBy.advancedFilters = [];
    }
    filterService.saveFilters(ETableFilters.WORKLOAD, filterBy);
  }
  storageUtil.save(STORAGE_KEYS.IS_FIRST_LOGIN, false);
}

/*
 * Adjust the workload phase filter to include all the relevant hidden phases,
 * for e.g. when Running is selected, also include Resuming and Updating
 * always include Terminating
 */
function adjustWorkloadPhaseFilter(filters: IAdvancedFilterModel[]): IAdvancedFilterModel[] {
  const updatedFilters: IAdvancedFilterModel[] = deepCopy(filters);
  const phaseFilter = updatedFilters.find((f) => f.name === WorkloadSortFilterFields.Phase) as IEnumBasedFilterModel;

  if (!phaseFilter) return updatedFilters;

  const phaseMapping: Record<string, string[]> = {
    [Phase.Running]: [Phase.Resuming, Phase.Updating],
    [Phase.Stopped]: [Phase.Stopping],
  };

  for (const phase in phaseMapping) {
    if (phaseFilter.selectedValues.includes(phase as keyof typeof phaseMapping)) {
      phaseFilter.selectedValues.push(...phaseMapping[phase as keyof typeof phaseMapping]);
    } else {
      phaseFilter.selectedValues = phaseFilter.selectedValues.filter(
        (v: string | boolean) => !phaseMapping[phase as keyof typeof phaseMapping].includes(v as string),
      );
    }
  }

  if (!phaseFilter.selectedValues.includes(Phase.Terminating)) {
    phaseFilter.selectedValues.push(Phase.Terminating);
  }

  return updatedFilters;
}

function getSupportedRoutes(fromPage: string, leavePage: string): IWorkloadSupportedRoutes | undefined {
  return {
    environment: {
      name: ENVIRONMENT_ROUTE_NAMES.ENVIRONMENT_NEW,
      query: { fromPage },
    },
    compute: {
      name: COMPUTE_RESOURCE_ROUTE_NAMES.COMPUTE_RESOURCE_NEW,
      query: { fromPage },
    },
    datasource: {
      pages: DATA_SOURCE_CREATION_ROUTE_MAP,
    },
    onLeave: {
      name: leavePage,
    },
  };
}

/*** Phase messages ***/
function flattenWorkloadPhaseIssues(issues: string): Field[] {
  if (!issues) return [];

  let parsedObject;
  try {
    parsedObject = JSON.parse(issues);
  } catch (error) {
    console.error("Invalid JSON input:", error);
    return [];
  }

  const result: Field[] = [];

  for (const key in parsedObject) {
    if (parsedObject[key] && Array.isArray(parsedObject[key])) {
      parsedObject[key].forEach((item: Omit<Field, "parentName">) => {
        result.push({ ...item, parentName: stringUtil.capitalizeString(key) });
      });
    }
  }

  return result;
}
