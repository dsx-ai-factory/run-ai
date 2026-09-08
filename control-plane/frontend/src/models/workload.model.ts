/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IStatusColOptions, ITableColumn } from "./table.model";
import type { SupersetSpec, SupersetWorkload, WorkloadSortFilterFields } from "@/swagger-models/workloads-client";
import {
  type Pod as WorkloadServicePod,
  Phase,
  type Connection,
  Preemptibility,
} from "@/swagger-models/workloads-client";
import type { MnnvlMode } from "@/swagger-models/workloads-manager-client";
import type {
  SpecificRunConnectionInfo,
  AssetIdAndKind,
  ToolType,
  DistributedData,
  DistributedFramework,
  Toleration,
  PodAffinity,
  EnvironmentVariableOfAsset,
  HostPathInstance,
  NfsInstance,
  S3Instance,
  GitInstance,
  SecretInstance,
  ConfigMapInstance,
  PvcInstance,
  DataVolumeInstance,
  ServingConfiguration,
} from "@/swagger-models/assets-service-client";
import type { ENimAccessModelMethod } from "@/components/workload-creation/hybrid-nim-model-section/nim-access-model/nim-access-model.model";

import type { IAutoScaleData } from "@/components/section/compute-resource-section";
import type { IItemizedListItem } from "@/components/common/runai-itemized-list";
import type { IUIVolume, TUIDataSourceInstancesKinds } from "@/models/data-source.model";
import type { EInferenceArchitecture, EInferenceType, EServerType } from "@/models/inference.model";
import type { IEnumBasedFilterSelectOption, IPaginationFilter } from "@/models/filter.model";
import type { IUIConnection } from "./environment.model";
import type { IServingEndpointAccessData } from "@/components/section/inference/serving-endpoint-section/serving-endpoint-access.model";
import type { ScopeType } from "@/swagger-models/authorization-client";
import type { EmptyDirInstance } from "@/swagger-models/policy-service-client";

export type { WorkloadServicePod };
export { Preemptibility };

export enum EIdleWorkloadMaxIdleDuration {
  training = "trainingJobMaxIdleDurationSecs",
  interactivePreemptible = "interactivePreemptibleJobMaxIdleDurationSecs",
  interactive = "interactiveJobMaxIdleDurationSecs",
}

export enum EWorkloadTimeLimit {
  training = "trainingJobTimeLimitSecs",
  interactive = "interactiveJobTimeLimitSecs",
}

export enum EWorkloadName {
  Workspace = "workspace",
  Training = "training",
  Inference = "inference",
  Distributed = "distributed",
  DistributedInference = "distributed-inference",
}

export enum EWorkloadType {
  Workspace = "Workspace",
  Training = "Training",
  Inference = "Inference",
  Distributed = "Distributed",
  DistributedInference = "DistributedInference",
  // Old clusters return different types for distributed workloads
  DistributedTraining = "DistributedTraining",
  TFJob = "TFJob",
}

export enum EWorkloadTypeSupportedInAssetsService {
  Workspace = "Workspace",
  Training = "Training",
  Inference = "Inference",
  Distributed = "Distributed",
  DistributedInference = "DistributedInference",
}

export const EWorkloadTypeTooltip = {
  [EWorkloadType.Workspace]:
    "Interactive development environment for building and testing. Recommended for lightweight experimentation and debugging.",
  [EWorkloadType.Training]:
    "Workload for standard or distributed training models. Recommended for resource-intensive model development.",
  [EWorkloadType.Inference]: "Deployment of an AI model for serving via an API. Recommended for production use.",
} as const;

export const workloadPriorityNameByWorkloadType = {
  [EWorkloadName.Workspace]: "InteractiveWorkload",
  [EWorkloadName.Training]: "TrainingWorkload",
  [EWorkloadName.Inference]: "InferenceWorkload",
  [EWorkloadName.Distributed]: "DistributedWorkload",
  [EWorkloadName.DistributedInference]: "DistributedInferenceWorkload",
};

export const workloadPreemptibilityOptions = [
  { label: "Preemptible", value: Preemptibility.Preemptible },
  { label: "Semi-preemptible", value: Preemptibility.SemiPreemptible },
  { label: "Non-preemptible", value: Preemptibility.NonPreemptible },
];

export enum EDistributedReplicaType {
  Master = "Master",
  Worker = "Worker",
  Leader = "Leader",
}

export enum EWorkloadFormType {
  Workspace = "workspace",
  Training = "training",
  Distributed = "distributed",
  Template = "template",
  Deployment = "deployment",
  Inference = "inference",
  DistributedInference = "distributed-inference",
}

export enum EWorkloadEntity {
  workspace = "workspace",
  training = "training",
  workload = "workload",
  deployment = "deployment",
}

export enum EExternalWorkloadType {
  SeldonDeployment = "SeldonDeployment",
  VirtualMachineInstance = "VirtualMachineInstance",
  TFJob = "TFJob",
  PyTorchJob = "PyTorchJob",
  XGBoostJob = "XGBoostJob",
  MPIJob = "MPIJob",
  Notebook = "Notebook",
  Job = "Job",
  StatefulSet = "StatefulSet",
  ReplicaSet = "ReplicaSet",
  Pod = "Pod",
  AmlJob = "AmlJob",
  Service = "Service",
  CronJob = "CronJob",
  DevWorkspace = "DevWorkspace",
  RayCluster = "RayCluster",
  RayJob = "RayJob",
  RayService = "RayService",
  TaskRun = "TaskRun",
  PipelineRun = "PipelineRun",
  Workflow = "Workflow",
  ScheduledWorkflow = "ScheduledWorkflow",
  NVIDIANIM = "NVIDIA-NIM",
  JAXJob = "JAXJob",
}

export enum EWorkloadAction {
  activate = "activate",
  run = "run",
  stop = "stop",
  connect = "connect",
  delete = "delete",
  clone = "clone",
  update = "update",
}

export enum EWorkloadNodeArchitecture {
  Distributed = "Distributed",
  Standard = "Standard",
}

export enum EWorkloadErrorMessage {
  NotRunaiWorkloadMessage = "This action is not supported for the selected workload. To perform the action, go to the application you used to create the workload.",
  FailedToLoadWorkloads = "Failed to load workloads",
  FailedToLoadWorkloadsCount = "Failed to load workloads count",
  FailedToLoadWorkloadPriorities = "Failed to load workload priorities",
  FailedToLoadPendingSchedulingMessages = "Failed to load pending scheduling messages",
  FailedToLoadPads = "Failed to load pods",
  FailedToRunWorkload = "Failed to start workload",
  FailedToDeleteWorkloads = "Failed to delete workloads",
  FailedToStopWorkloads = "Failed to stop workload",
  NoDeletePermission = "You don't have permission to delete this workload.",
  DeleteWorkloadIsNotSupported = "This action is not supported for the selected workload. To perform it, make sure you are the creator of the workload or go to the application used to create it.",
  WorkloadIsDeleting = "The selected workload is already deleted or in the process of being deleted",
  NoActionPermission = "You don't have permission to perform this action.",
  InferenceCantBeCreatedKnativeAndLwsMissing = "Inference can't be created because Knative and Leader Worker Set (LWS) aren't installed on the selected cluster. Contact your administrator.",
  ProjectNoLongerExists = "A workload whose project no longer exists cannot be copied.",
  WorkloadsCreatedUsingTheFlexibleAndSpecificCluster = "This action is only supported for inference workloads created using the flexible form in clusters version 2.19 or higher",
  WorkloadFailedCantBeUpdated = "The selected workload failed and can't be updated",
  WorkloadFailedCantBeRun = "To re-run a failed workload, first make a copy",
  RunActionNotAuthorized = "The selected workload can't be set to run. You're either not authorized to perform this action or need to go to the application from which they were created to perform the action.",
  StopActionNotAuthorized = "The selected workload can't be stopped. You're either not authorized to perform this action or need to go to the application from which they were created to perform the action.",
  WorkloadCreatingCantBeUpdated = "The selected workload is creating and can't be updated",
  WorkloadDeletingCantBeUpdated = "The selected workload is deleting and can't be updated",
  WorkloadTerminatingCantBeUpdated = "The selected workload is terminating and can't be updated",
  CliWorkloadsCantBeUpdated = "This action is not supported for the selected workload because it was created via the CLI. To perform the action, go to the CLI to update the workload.",
  RequiresFlexibleSubmissionForm = "This action is not supported for the selected workload. To perform it, ask your administrator to enable the flexible workload submission form.",
  ActionNotSupported = "This action is not supported for the selected workload",
}

export enum EWorkloadModalEntity {
  Project = "Project",
  Department = "Department",
  Nodepool = "Nodepool",
}

export enum EWorkloadTelemetryGroupBy {
  ClusterId = "ClusterId",
  DepartmentId = "DepartmentId",
  ProjectId = "ProjectId",
  Type = "Type",
  CurrentNodepools = "CurrentNodepools",
  RequestedNodepools = "RequestedNodepools",
  Phase = "Phase",
  Category = "Category",
  Bucket = "Bucket",
}

export type EWorkloadTelemetryGroupByWithoutBucket = Exclude<EWorkloadTelemetryGroupBy, "Bucket">;

export interface EWorkloadTelemetryPointMeta {
  entityId: string;
  groupByKey: EWorkloadTelemetryGroupBy;
}

export interface DrillupEventWorkloadTelemetryObject {
  seriesOptions: {
    data: {
      drilldown: boolean;
      meta: EWorkloadTelemetryPointMeta;
      name: string;
      y: number;
    }[];
  };
}

export interface DrilldownEventWorkloadTelemetryObject {
  point: { meta: EWorkloadTelemetryPointMeta };
  seriesOptions: {
    data: {
      drilldown: boolean;
      name: string;
      y: number;
    }[];
  };
}

export interface ETelemetrySeriesPoint {
  meta: EWorkloadTelemetryPointMeta;
  name?: string;
  y?: number;
}

export interface ETelemetryDrilldownEventObject {
  point: ETelemetrySeriesPoint;
  drilldown: unknown;
}

export enum EWorkloadPodPhase {
  Running = "Running",
}

export type ExtendedPhase = (typeof ExtendedPhase)[keyof typeof ExtendedPhase];

export const ExtendedPhase = {
  ...Phase,
  Succeeded: "Succeeded",
  Deleted: "Deleted",
} as const;

export const WorkloadPhaseMap: Record<ExtendedPhase, IStatusColOptions> = {
  [ExtendedPhase.Creating]: {
    status: "Creating...",
    displayAnimation: true,
  },
  [ExtendedPhase.Initializing]: {
    status: "Initializing",
    displayAnimation: false,
  },
  [ExtendedPhase.Updating]: {
    status: "Updating...",
    displayAnimation: true,
  },
  [ExtendedPhase.Pending]: {
    status: "Pending",
    color: "warning",
    displayAnimation: false,
  },
  [ExtendedPhase.Running]: {
    status: "Running",
    displayAnimation: false,
    color: "success",
  },
  [ExtendedPhase.Stopped]: {
    status: "Stopped",
    displayAnimation: false,
  },
  [ExtendedPhase.Degraded]: {
    status: "Running with issues",
    color: "warning",
    displayAnimation: false,
  },
  [ExtendedPhase.Failed]: {
    status: "Failed",
    color: "negative",
    displayAnimation: false,
  },
  [ExtendedPhase.Completed]: {
    status: "Completed",
    displayAnimation: false,
  },
  [ExtendedPhase.Deleting]: {
    status: "Deleting...",
    displayAnimation: true,
    filterKey: "status",
  },
  [ExtendedPhase.Deleted]: {
    status: "Deleted",
    displayAnimation: false,
    filterKey: "status",
  },
  [ExtendedPhase.Unknown]: {
    status: "Unknown",
    displayAnimation: false,
  },
  [ExtendedPhase.Stopping]: {
    status: "Stopping...",
    displayAnimation: true,
  },
  [ExtendedPhase.Terminating]: {
    status: "Terminating...",
    displayAnimation: true,
  },
  [ExtendedPhase.Succeeded]: {
    status: "Succeeded",
    color: "success",
    displayAnimation: false,
  },
  [ExtendedPhase.Resuming]: {
    status: "Resuming...",
    displayAnimation: true,
  },
};

interface IWorkloadMetadataAnnotations {
  workspaceId: string;
}

export interface IWorkloadMetadata {
  name: string;
  namespace: string;
  creationTimestamp?: null;
  annotations: IWorkloadMetadataAnnotations;
}

export interface IWorkloadSpec {
  environment: {
    items: Record<string, { value: string }>;
  };
  name: { value: string };
  image: { value: string };
  username: { value: string };
  command?: { value: string };
  arguments?: { value: string };
  volumes?: {
    items: Record<string, unknown>;
  };
  cpu?: { value: string };
  gpu?: { value: string };
  memory?: { value: string };
  nodeType?: { value: string };
  gitSync?: {
    items: Record<string, unknown>;
  };
  largeShm: {
    value: boolean;
  };
  imagePullPolicy: {
    value: string;
  };
  annotations: {
    items: Record<string, unknown>;
  };
  preemptible?: { value: string };
}

export interface IWorkload {
  apiVersion: string;
  kind: string;
  metadata: IWorkloadMetadata;
  spec: IWorkloadSpec;
  masterSpec?: IWorkloadSpec;
}

export interface IWorkloadCreate {
  metadata: IWorkloadMetadata;
  spec: IWorkloadSpec;
  status?: Record<string, unknown>;
  masterSpec?: IWorkloadSpec;
}

export interface IWorkloadResponse {
  metadata: IWorkloadMetadata;
  spec: IWorkloadSpec;
  status?: Record<string, unknown>;
}

export interface IWorkloadDurationOption {
  value: EIdleWorkloadMaxIdleDuration | string;
  label: string;
  duration: number | null | undefined;
  inheritedDuration?: number | null;
}

export interface IWorkloadListModal {
  name: string;
  type: string;
  status: IStatusColOptions;
}

export interface IUIDistributedMaster {
  assets: IUIWorkloadAssets;
  specificEnv: IUIWorkloadSpecificEnv;
}

export interface IUIDistributed extends Omit<DistributedData, "master" | "framework" | "numWorkers"> {
  numWorkers?: number;
  master?: IUIDistributedMaster | null;
  distFramework?: DistributedFramework | null;
  noMaster?: boolean;
}

export interface IUIWorkloadCreation {
  name: string;
  projectId: number;
  namespace: string;
  clusterId: string;
  assets: IUIWorkloadAssets;
  specificEnv: IUIWorkloadSpecificEnv;
  distributed?: IUIDistributed | null;
  enableEditingMaster?: boolean;
}

export interface IUIInferenceCreation extends IUIWorkloadCreation {
  inferenceType?: EInferenceType;
  modelName?: string;
  image?: string;
  accessModelMethod?: ENimAccessModelMethod;
  accessFromStorage?: boolean;
  profileNameOrHash?: string;
  gatedModel?: boolean;
  token?: string;
  genericSecretId?: string;
  genericSecretName?: string;
  serverType: EServerType;
  templateId?: string | null;
  architecture: EInferenceArchitecture;
}

export interface IUIInferenceTemplateCreation extends IUIInferenceCreation {
  templateName: string;
  templateDescription: string;
}

export interface IUIWorkloadAssets {
  environment: string | null;
  compute: string | null;
  datasources?: Array<AssetIdAndKind>;
  uiVolumes?: Array<IUIVolume>;
}

export interface IUIWorkloadSpecificEnv {
  command?: string | null;
  args?: string | null;
  environmentVariables?: Array<EnvironmentVariableOfAsset>;
  allowOverQuota?: boolean | null;
  connections?: Array<SpecificRunConnectionInfo> | null;
  runAsGid?: number | null;
  runAsUid?: number | null;
  supplementalGroups?: string | null;
  nodeType?: string | null;
  nodePools?: Array<string> | null;
  annotations?: Array<IItemizedListItem>;
  labels?: Array<IItemizedListItem>;
  autoDeletionTimeAfterCompletionSeconds?: number | null;
  autoScaleData?: IAutoScaleData;
  backoffLimit?: number | null;
  tolerations?: Array<Toleration> | null;
  podAffinity?: PodAffinity | null;
  completions?: number | null;
  parallelism?: number | null;
  servingPortAccess?: IServingEndpointAccessData | null;
  terminationGracePeriodSeconds?: number | null;
  servingConfiguration?: ServingConfiguration | null;
}

export interface IWorkloadMeta {
  name: string;
  namespace: string;
  clusterId: string;
  projectId: number;
}

export interface IWorkloadPodsModalOptions {
  columns: ITableColumn[];
  header: string;
  displayRedirection?: boolean;
  getPods: (filters: IPaginationFilter) => Promise<WorkloadServicePod[]>;
  getPodsCount: (filters: IPaginationFilter) => Promise<number>;
  clusterId?: string;
}
export interface IWorkloadConnectionsModalOptions {
  connections: Connection[];
  header: string;
  workloadDetails?: {
    currentUser: string;
    isRunning: boolean;
    name: string;
    type: string;
    projectName: string;
  };
}

export interface IWorkloadListModalOptions {
  entityFilter: string;
  //in order to support the table to pick the correct filter need to supply both filterName and entityName
  filterName: WorkloadSortFilterFields;
  entityName: string;
  entityType: EWorkloadModalEntity;
  columns: ITableColumn[];
  clusterId: string;
}

export interface IWorkloadListModalRow {
  name: string;
  status: IStatusColOptions;
  type: IDisplayWorkloadType;
}

export type IDisplayWorkloadType = "Workspace" | "Training" | "Inference" | "Distributed";

const workloadFilterDisplayedPhasesSet: Set<string> = new Set([
  Phase.Creating,
  Phase.Initializing,
  Phase.Pending,
  Phase.Running,
  Phase.Degraded,
  Phase.Stopped,
  Phase.Failed,
  Phase.Completed,
  Phase.Unknown,
]);
export const workloadPhaseFilterOptions: IEnumBasedFilterSelectOption[] = Object.keys(Phase).map((key) => ({
  label: key,
  value: key,
  hidden: !workloadFilterDisplayedPhasesSet.has(key),
}));

export interface IWorkloadSupportedRoutes {
  environment: {
    name: string;
    query: { fromPage: string };
  };
  compute: {
    name: string;
    query: { fromPage: string };
  };
  datasource: {
    pages: Record<string, string>;
  };
  onLeave: {
    name: string;
  };
}

export interface IToolItem {
  toolType: ToolType;
  toolName: string;
  url: string;
  authorizedUsers?: string[] | null;
}

export type TUIStorageInstance =
  | IUINfsInstance
  | IUIHostPathInstance
  | IUIGitInstance
  | IUIS3Instance
  | IUIConfigMapInstance
  | IUISecretVolumeInstance
  | IUIPvcInstance;

export interface IUIHostPathInstance extends HostPathInstance {
  id: string;
  type: TUIDataSourceInstancesKinds;
  createdAt?: number;
}

export interface IUINfsInstance extends NfsInstance {
  id: string;
  type: TUIDataSourceInstancesKinds;
  createdAt?: number;
}

export interface IUIGitInstance extends GitInstance {
  id: string;
  type: TUIDataSourceInstancesKinds;
  createdAt?: number;
}

export interface IUIS3Instance extends S3Instance {
  id: string;
  type: TUIDataSourceInstancesKinds;
  createdAt?: number;
}

export interface IUIConfigMapInstance extends ConfigMapInstance {
  id: string;
  type: TUIDataSourceInstancesKinds;
  createdAt?: number;
}

export interface IUISecretVolumeInstance extends SecretInstance {
  id: string;
  type: TUIDataSourceInstancesKinds;
  createdAt?: number;
}

export interface IUIPvcInstance extends PvcInstance {
  id: string;
  type: TUIDataSourceInstancesKinds;
  createdAt?: number;
}

export interface IUIDataVolumeInstance extends DataVolumeInstance {
  id: string;
  type: TUIDataSourceInstancesKinds;
  createdAt?: number;
  dataVolumeId?: string | null;
}

export interface IUIEmptyDirInstance extends EmptyDirInstance {
  id: string;
  type: TUIDataSourceInstancesKinds;
  createdAt?: number;
}

export interface IUIStorageFields {
  nfs?: IUINfsInstance[];
  hostPath?: IUIHostPathInstance[];
  git?: IUIGitInstance[];
  s3?: IUIS3Instance[];
  configMapVolume?: IUIConfigMapInstance[];
  secretVolume?: IUISecretVolumeInstance[];
  pvc?: IUIPvcInstance[];
  volumes?: IUIPvcInstance[];
  dataVolume?: IUIDataVolumeInstance[];
  emptyDirVolume?: IUIEmptyDirInstance[];
}

export type IUIStorageInstance =
  | IUINfsInstance
  | IUIHostPathInstance
  | IUIGitInstance
  | IUIS3Instance
  | IUIConfigMapInstance
  | IUISecretVolumeInstance
  | IUIPvcInstance
  | IUIDataVolumeInstance;

export type IUIWorkloadStorageFields = IUIStorageFields;

export type IUIDistributedWorkloadSpec = Omit<IUISuperWorkloadSpec, "leader" | "worker">;

export interface IUISuperWorkloadSpec
  extends Omit<SupersetSpec, "ports" | "exposedUrls" | "relatedUrls" | "storage" | "leader" | "worker"> {
  leader?: IUIDistributedWorkloadSpec;
  worker?: IUIDistributedWorkloadSpec;
  connections?: Array<IUIConnection>;
  storage?: IUIWorkloadStorageFields;
  mnnvl?: MnnvlMode | null;
}

export interface IUISuperWorkloadCreationRequest
  extends Omit<
    SupersetWorkload,
    "spec" | "workloadId" | "requestedName" | "createdBy" | "createdAt" | "desiredPhase" | "masterSpec"
  > {
  spec: IUISuperWorkloadSpec;
  masterSpec?: IUISuperWorkloadSpec;
  // template fields
  scopeType?: ScopeType;
  scopeId?: string | null;
  description?: string | null;
  departmentId?: string;
  // templateId for workload submission from templates
  templateId?: string | null;
}

export type IUIEnvironmentFields = Omit<
  IUISuperWorkloadSpec,
  | "compute"
  | "storage"
  | "nodeType"
  | "nodePools"
  | "podAffinity"
  | "annotations"
  | "labels"
  | "tolerations"
  | "terminateAfterPreemption"
  | "autoDeletionTimeAfterCompletionSeconds"
  | "backoffLimit"
  | "tty"
  | "stdin"
>;

export interface ICanSubmitWorkload {
  info: string; // this prop for debugging
  statusCode?: number;
}

/** Drawer Navigation **/
export type TWorkloadDrawerTab = "logs" | "terminal" | "metrics" | "pods" | "event history" | "details" | "structure";

export interface IWorkloadDrawerNavigationPayload {
  tab: TWorkloadDrawerTab;
  podName?: string;
}

export interface IOpenWorkloadDrawerTabPayload extends IWorkloadDrawerNavigationPayload {
  workloadId: string;
}

/** Categories **/
export enum EWorkloadCategory {
  Workspace = "Build",
  Training = "Train",
  // eslint-disable-next-line @typescript-eslint/no-duplicate-enum-values -- Distributed intentionally shares category with Training
  Distributed = "Train",
  Inference = "Deploy",
}

export enum EDistributedFrameworkUrlSlug {
  XGBoost = "xgboost",
  PyTorch = "pytorch",
  MPI = "mpi",
  TensorFlow = "tf",
  JAX = "jax",
}

export const distributedFrameworkUrlMapping: Record<string, EDistributedFrameworkUrlSlug> = {
  [EExternalWorkloadType.XGBoostJob]: EDistributedFrameworkUrlSlug.XGBoost,
  [EExternalWorkloadType.PyTorchJob]: EDistributedFrameworkUrlSlug.PyTorch,
  [EExternalWorkloadType.MPIJob]: EDistributedFrameworkUrlSlug.MPI,
  [EExternalWorkloadType.TFJob]: EDistributedFrameworkUrlSlug.TensorFlow,
  [EExternalWorkloadType.JAXJob]: EDistributedFrameworkUrlSlug.JAX,
};

export interface IYamlWorkloadFormState {
  yamlContent: string;
  priorityClass: string | null;
  preemptibility: Preemptibility | null;
  mnnvl: MnnvlMode | null;
}
