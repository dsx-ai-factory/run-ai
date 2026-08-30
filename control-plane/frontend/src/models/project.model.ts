import type { IStatusColOptions } from "@/models/table.model";
import type { EQuotaEntity } from "@/models/resource.model";
import type { Scope } from "@/swagger-models/assets-service-client";

import type { ISelectOption } from "@/models/global.model";
import type { Project, QuotaStatusNodePool, Resources } from "@/swagger-models/org-unit-service-client";
import { ProjectPhase } from "@/swagger-models/org-unit-service-client";
import type { IFilterBy } from "@/models/filter.model";
import type { AVG_RESOURCE_TIMEFRAME_FILTER_KEY } from "@/models/org-unit.model";
import DOCS from "@/constants/docs";

export enum EWorkloadNodeAffinity {
  Train = "train",
  Interactive = "interactive",
}

export enum EProjectModalEntity {
  Department = "Department",
}

export enum EProjectConditionDeletionReason {
  DataVolumesDeletionHandlerFailed = "DataVolumesDeletionHandlerFailed",
  WorkloadsDeletionHandlerFailed = "WorkloadsDeletionHandlerFailed",
  SecretsDeletionHandlerFailed = "SecretsDeletionHandlerFailed",
  PvcsDeletionHandlerFailed = "PvcsDeletionHandlerFailed",
}

export enum EProjectConditionNotReadyReason {
  NamespaceGetFailed = "NamespaceGetFailed",
  NamespaceNotFound = "NamespaceNotFound",
  NamespaceLabelMissing = "NamespaceLabelMissing",
  NamespaceUpdateFailed = "NamespaceUpdateFailed",
  NamespaceCreateFailed = "NamespaceCreateFailed",
  NamespaceHandlerFailed = "NamespaceHandlerFailed",
  LimitRangeHandlerFailed = "LimitRangeHandlerFailed",
  QueuesHandlerFailed = "QueuesHandlerFailed",
  RoleBindingsHandlerFailed = "RoleBindingsHandlerFailed",
  SecretsHandlerFailed = "SecretsHandlerFailed",
  PvcsHandlerFailed = "PvcsHandlerFailed",
}

interface IReasonTooltip {
  message: string;
  isHandlerReason: boolean;
}

const getSupportLink = (): string => (DOCS.SUPPORT.URL ? `<a href="${DOCS.SUPPORT.URL}">${DOCS.SUPPORT.LABEL}</a>` : "");
const ensureTerminalPeriod = (message: string): string => (/[.!?]$/.test(message) ? message : `${message}.`);

export const PROJECT_CONDITION_NOT_READY_REASON_TOOLTIP_MAP = new Map<EProjectConditionNotReadyReason, IReasonTooltip>([
  [
    EProjectConditionNotReadyReason.NamespaceGetFailed,
    {
      message: "A namespace couldn't be gotten for the project.<br/>Try creating the project again",
      isHandlerReason: false,
    },
  ],
  [
    EProjectConditionNotReadyReason.NamespaceNotFound,
    {
      message: "A namespace couldn't be found for the project.<br/>Make sure the namespace exists",
      isHandlerReason: false,
    },
  ],
  [
    EProjectConditionNotReadyReason.NamespaceLabelMissing,
    { message: "The 'runai/queue' label is missing from the namespace.<br/>Add the label", isHandlerReason: false },
  ],
  [
    EProjectConditionNotReadyReason.NamespaceUpdateFailed,
    {
      message: "The namespace couldn't be updated for the project.<br/>Try creating the project again",
      isHandlerReason: false,
    },
  ],
  [
    EProjectConditionNotReadyReason.NamespaceCreateFailed,
    {
      message: "The namespace couldn't be created for the project.<br/>Try creating the project again",
      isHandlerReason: false,
    },
  ],
  [
    EProjectConditionNotReadyReason.NamespaceHandlerFailed,
    { message: "The project couldn't be created (NamespaceHandlerFailed).", isHandlerReason: true },
  ],
  [
    EProjectConditionNotReadyReason.LimitRangeHandlerFailed,
    { message: "The project couldn't be created (LimitRangeHandlerFailed).", isHandlerReason: true },
  ],
  [
    EProjectConditionNotReadyReason.QueuesHandlerFailed,
    { message: "The project couldn't be created (QueuesHandlerFailed).", isHandlerReason: true },
  ],
  [
    EProjectConditionNotReadyReason.RoleBindingsHandlerFailed,
    { message: "The project couldn't be created (RoleBindingsHandlerFailed).", isHandlerReason: true },
  ],
  [
    EProjectConditionNotReadyReason.SecretsHandlerFailed,
    { message: "The project couldn't be created (SecretsHandlerFailed).", isHandlerReason: true },
  ],
  [
    EProjectConditionNotReadyReason.PvcsHandlerFailed,
    { message: "The project couldn't be created (PvcsHandlerFailed).", isHandlerReason: true },
  ],
]);

export function buildNotReadyTooltip(reasons: EProjectConditionNotReadyReason[]): string {
  if (reasons.length === 0) return "";
  const supportLink = getSupportLink();

  if (reasons.length === 1) {
    const entry = PROJECT_CONDITION_NOT_READY_REASON_TOOLTIP_MAP.get(reasons[0]);
    if (!entry) return "";
    if (!supportLink) return ensureTerminalPeriod(entry.message);
    const contactPhrase = entry.isHandlerReason ? "Contact" : "or contact";
    const separator = entry.isHandlerReason ? "<br/>" : " ";
    return `${entry.message}${separator}${contactPhrase} ${supportLink}.`;
  }

  const bodies = reasons.flatMap((r) => {
    const entry = PROJECT_CONDITION_NOT_READY_REASON_TOOLTIP_MAP.get(r);
    return entry ? [entry.message] : [];
  });
  const body = (supportLink ? bodies : bodies.map(ensureTerminalPeriod)).join("<br/><br/>");
  return supportLink ? `${body}<br/><br/>Contact ${supportLink}.` : body;
}

export const ProjectPhaseMap: Record<ProjectPhase, IStatusColOptions> = {
  [ProjectPhase.Creating]: {
    status: "Creating...",
    displayAnimation: true,
  },
  [ProjectPhase.Updating]: {
    status: "Updating...",
    displayAnimation: true,
  },
  [ProjectPhase.Deleting]: {
    status: "Deleting...",
    displayAnimation: true,
    tooltipText: "The project’s subordinates workloads or assets are being deleted",
  },
  [ProjectPhase.Ready]: {
    status: "Ready",
    color: "success",
    displayAnimation: false,
  },
  [ProjectPhase.NotReady]: {
    status: "Not Ready",
    color: "negative",
    displayAnimation: false,
  },
  [ProjectPhase.Initializing]: {
    status: "Initializing...",
    displayAnimation: true,
  },
  [ProjectPhase.Unknown]: {
    status: "Unknown",
    displayAnimation: false,
  },
  [ProjectPhase.Deleted]: {
    status: "Deleted",
    displayAnimation: false,
  },
};

export interface INodeAffinitySelectOption {
  label: string;
  value: string;
  loading: boolean;
  selectedTypes: ISelectOption[];
}

export interface IProjectsModalOptions {
  parentId: string;
  entityName: string;
  entityType: EProjectModalEntity;
  clusterId: string;
}

export interface IProjectTable extends Project {
  rolesNames?: string[];
  averageGpuAllocation?: number;
  averageGpuUtilization?: number;
  averageGpuMemoryUtilization?: number;
}

export interface IProjectResources {
  id: number;
  tenantId: number;
  clusterUuid: string;
  name: string;
  deservedGpus: number;
  allocatedGpus: number;
  createdAt?: string;
}

export interface INodePoolsResourcesRow extends Resources {
  priorityOrder?: number | string;
  allocatedGpus?: number;
  allocatedCpu?: number;
  allocatedMemory?: number;
  overQuotaWeightLabel?: string;
}
export interface IResourceQuota {
  deserved?: number | null;
  maxAllowed?: number | null;
  overQuotaWeight?: number | null;
}

export interface IProjectSectionModel {
  scope: Scope;
  projectId: number | null;
}

export interface INodePoolModalOptions {
  header: string;
  entityType: EQuotaEntity;
  nodePoolQuotaStatuses: QuotaStatusNodePool[];
  resources: Resources[];
  defaultNodePools?: string[];
  loading?: boolean;
  showRedirectBtn?: boolean;
}

export enum EProjectColumnName {
  ProjectName = "name",
  DepartmentName = "department-name",
  Status = "status",
  NodePools = "node-pools",
  AccessRules = "accessRules",
  AllocatedCpu = "allocated-cpu",
  AllocatedMemory = "allocated-memory",
  AllocatedGpus = "allocated-gpus",
  GpuUtilization = "gpu-utilization",
  CpuUtilization = "cpu-utilization",
  MemoryUtilization = "memory-utilization",
  GpuQuota = "gpu-quota",
  CpuQuota = "cpu-quota",
  CpuMemoryQuota = "cpu-memory-quota",
  NodeAffinityTrain = "node-affinity-train",
  NodeAffinityInteractive = "node-affinity-interactive",
  TrainingJobMaxIdleDurationSecs = "training-job-max-idle-duration-secs",
  InteractivePreemptibleJobMaxIdleDurationSecs = "interactive-preemptible-job-max-idle-duration-secs",
  InteractiveJobMaxIdleDurationSecs = "interactive-job-max-idle-duration-secs",
  InteractiveJobTimeLimitSecs = "interactive-job-time-limit-secs",
  TrainingJobTimeLimitSecs = "training-job-time-limit-secs",
  CreatedAt = "created-at",
  Workloads = "workloads",
}

export interface IProjectTableFilterBy extends IFilterBy {
  [AVG_RESOURCE_TIMEFRAME_FILTER_KEY]: string;
}
