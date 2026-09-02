/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// models
import {
  EAccessOptions,
  type IConnectionAccess,
} from "@/components/environment/connection-access/connection-access-modal/connection-access-modal.model";
import {
  ARGS_PREFIX_FOR_TGI,
  ARGS_PREFIX_FOR_VLLM,
  EHuggingFaceEnvVariableNames,
  EInferenceArchitecture,
  EInferenceType,
  ENimEnvVariableNames,
  HF_HOME_DEFAULT_PATH,
  HOME_DEFAULT_PATH,
  LLM_IMAGE_FOR_HUGGING_FACE,
  LLM_IMAGE_FOR_TGI,
  MAX_REPLICAS_DEFAULT_VALUE,
  MIN_REPLICAS_DEFAULT_VALUE,
  NIM_IMAGE_PREFIX,
  NIM_WORKLOAD_LABEL,
  SERVING_PORT_SERVER_PORT,
} from "@/models/inference.model";
import {
  EWorkloadType,
  type IUISuperWorkloadCreationRequest,
  type IUIInferenceCreation,
  type IUIWorkloadCreation,
  type IUIWorkloadSpecificEnv,
  type IUISuperWorkloadSpec,
  type IUIStorageInstance,
  type IUIStorageFields,
  type IUIPvcInstance,
  type IUIInferenceTemplateCreation,
  EWorkloadCategory,
} from "@/models/workload.model";
import type { IUIWorkloadEnvSectionModel } from "@/components/section/environment-section/environment-section.model";
import type { IComputeSectionData } from "@/components/section/compute-resource-section/compute-resource-section.models";
import type { Resources } from "@/swagger-models/org-unit-service-client";
import {
  type InferenceSpecificRunParams,
  ServingPortAccessAuthorizationTypeEnum,
  type EnvironmentAsset,
  type EnvironmentVariableOfAsset,
  type InferenceCreationRequestV2,
  type SpecificRunServingPortAccessServingPortAccess,
  type PvcInstance,
  type HostPathInstance,
  type NfsInstance,
  type AutoScaling,
  type SpecificRunAutoScalingAutoScaling,
  AssetKind,
  type SupersetSpec,
  type SupersetSpecAllOfStorage,
  type ScopeType,
  type EnvironmentVariableUserCredential,
  ImagePullPolicy,
  UidGidSource,
  type DistributedInferenceLeaderWorkerSpec,
  type DistributedInferenceLeaderSpecFieldsLeader,
  type DistributedInferenceWorkerSpecFieldsWorker,
  type DistributedInferenceServingPortProtocol,
  type DistributedInferenceServingPortAccessAuthorizationTypeEnum,
  DistributedInferenceStartupPolicy,
} from "@/swagger-models/assets-service-client";
import type {
  DistributedInference,
  DistributedInferenceCreationRequest,
  DistributedInferenceSpecSpec,
  UnifiedPolicyInfo,
  UnifiedPolicyInfoPerReplicaTypeRules,
  WorkspaceSpecSpecAllOfCompute,
} from "@/swagger-models/workloads-client/api";
import { ENimAccessModelMethod } from "@/components/workload-creation/hybrid-nim-model-section/nim-access-model/nim-access-model.model";
import {
  Source,
  Phase,
  type Workload,
  type Inference,
  type InferenceCreationRequest,
  type InferenceSpecSpec,
  type EnvironmentVariableSecret,
  type EnvironmentVariable,
  type ServingPort,
  type SupersetServingPort,
} from "@/swagger-models/workloads-client";
import type {
  Inference as InferenceFlat,
  InferenceTemplateCreationRequest,
  InferenceUpdateSpecSpec,
  ServingPortProtocol,
  TemplateAssetsIds,
  UnifiedPolicyInfoPerReplicaType,
  UnifiedWorkload,
} from "@/swagger-models/workloads-client";
import type { IServingEndpointAccessData } from "@/components/section/inference/serving-endpoint-section/serving-endpoint-access.model";
import type { IItemizedListItem } from "@/components/common/runai-itemized-list/runai-itemized-list.model";
import type { IAutoScaleData } from "@/components/section/compute-resource-section/auto-scale-section";
import { workloadUtil, type TWorkloadsToConvert } from "@/utils/workload.util/workload.util";
import type { IUIVolume, TDataSourceKinds } from "@/models/data-source.model";
import type { ClusterDependencyStatus, ClustersMinimalData } from "@/swagger-models/cluster-service-client";

// utils
import { fallbackDefaultIfNullOrUndefined, isEqual, isNumber, omit, pick } from "@/utils/common.util/common.util";
import { workloadCreateFormUtil } from "@/utils/workload-create-form.util/workload-create-form.util";
import { EServerType } from "@/models/inference.model";
import { workloadHybridUtil } from "../workload-hybrid.util/workload-hybrid.util";
import type { InferencePolicyDefaultsAndRulesV2 } from "@/swagger-models/policy-service-client";
import { scaleDownUtil } from "@/components/section/compute-resource-section/auto-scale-section/scale-down-section/scale-down.util";
import { is } from "quasar";
import { thresholdMetricUtils } from "@/components/section/compute-resource-section/auto-scale-section/threshold-metric-section/threshold-metric.util";
import { ERunaiCredentialSource } from "@/components/common/runai-credential-selector-by-source/runai-credential-selector-by-source.model";

export const inferenceWorkloadUtil = {
  getEmptyUIInferenceCreation,
  getEmptyUINimInferenceCreation,
  getInferenceComputeResourceData,
  getEnvironmentRunParamsData,
  getHuggingFaceTokenValue,
  getEnvironmentVariableValue,
  inferCredentialSourceFromEnvVar,
  excludeEnvironmentVariablesByName,
  getInferenceCreationRequestV2,
  convertInferenceToWorkloadUI,
  getInferenceRequestV2SpecificRunParams,
  getHFInferenceWorkloadCreationRequest,
  isLatencyEnableOnCluster,
  isAutoScaleEnableOnCluster,
  convertInferenceToWorkload,
  servingPortInternalToolInfo,
  servingPortAccessAuthorization,
  getUIAutoScaleEmptyState,
  getUIServingPortEmptyState,
  convertAutoScaleUIToFlat,
  convertAutoScaleFlatToUI,
  getInferenceSupportedDataSources,
  getHybridWorkloadUIModel,
  prepareUpdateInferenceFlatRequest,
  replaceEnvironmentVariable,
  prepareInferenceFlatRequest,
  prepareDistributedInferenceFlatRequest,
  prepareInferenceTemplateFlatRequest,
  adjustNimWorkloadProperties,
  adjustHFWorkloadProperties,
  getHFArgsForModel,
  getHFModelNameFromArgs,
  setInferenceNIMUIDefaults,
  setInferenceHuggingFaceUIDefaults,
  getAutoScalingData,
  setInferenceUIDefaults,
  getHybridInferenceModel,
  getHybridDistributedModel,
  getNimModelNameFromImage,
  getEmptyUIInferenceCreationForTemplate,
  getEmptyUISuperWorkloadModel,
  getDistributedInferenceModel,
  setDistributedUIDefaults,
  convertDistributedInferenceToWorkload,

  // private functions
  _getAuthorizationType,
  _getServingPortAccessOption,
  _getFlatWorkloadData,
  _getFlatInferenceWorkloadData,
  _addNimEnvironmentVariables,
  _fillDatasourceData,
  _handleHuggingFaceEnvironmentVariables,
  normalizeAutoScaleData,
  _getFlatInferenceSpecData,
  _addNimEnvironmentVariablesIfNeeded,
  _ensureStoragePathEnvVar,
  removeEmptyStorageArrays,
  _cleanStorageAndSetEnvVar,
  _orderHfEnvironmentVariables,
  _orderNimEnvironmentVariables,
  _ensureDefaultEnvVar,
};

function getHFArgsForModel(image: string, modelName: string): string {
  return image.includes(LLM_IMAGE_FOR_TGI)
    ? `${ARGS_PREFIX_FOR_TGI} ${modelName}`
    : `${ARGS_PREFIX_FOR_VLLM} ${modelName}`;
}

function getHFModelNameFromArgs(args: string): string {
  if (args.includes(ARGS_PREFIX_FOR_TGI)) {
    return args.split(ARGS_PREFIX_FOR_TGI)[1].trim();
  } else if (args.includes(ARGS_PREFIX_FOR_VLLM)) {
    return args.split(ARGS_PREFIX_FOR_VLLM)[1].trim();
  } else {
    return "";
  }
}

function getNimModelNameFromImage(image: string): string {
  return image.replace(`${NIM_IMAGE_PREFIX}/`, "");
}
export interface IWorkloadCreateOptions {
  storage?: {
    dataSource: PvcInstance | HostPathInstance | NfsInstance;
    kind: keyof SupersetSpecAllOfStorage;
  };
  genericSecretName?: string;
  addHomeDir?: boolean;
}

function convertInferenceToWorkload(inferenceWorkload: Inference, tenantId: number): Workload {
  return {
    category: EWorkloadCategory.Inference,
    source: Source.ControlPlane,
    conditions: [{ status: "", type: "", lastTransitionTime: null }],
    priorityClassName: "",
    type: EWorkloadType.DistributedInference,
    name: inferenceWorkload.name,
    id: inferenceWorkload.workloadId,
    clusterId: inferenceWorkload.clusterId,
    projectName: "",
    projectId: inferenceWorkload.projectId,
    departmentName: "",
    departmentId: inferenceWorkload.departmentId || "",
    namespace: "",
    createdAt: inferenceWorkload.createdAt,
    phase: inferenceWorkload.actualPhase || Phase.Creating,
    k8sPhase: "",
    tenantId: tenantId,
    runningPods: 0,
    phaseUpdatedAt: "",
    k8sPhaseUpdatedAt: "",
    updatedAt: "",
    deletedAt: null,
    submittedBy: inferenceWorkload.createdBy,
    priority: null,
  };
}

function convertDistributedInferenceToWorkload(
  distributedInferenceWorkload: DistributedInference,
  tenantId: number,
): Workload {
  return {
    category: EWorkloadCategory.Inference,
    source: Source.ControlPlane,
    conditions: [{ status: "", type: "", lastTransitionTime: null }],
    priorityClassName: "",
    type: EWorkloadType.DistributedInference,
    name: distributedInferenceWorkload.name,
    id: distributedInferenceWorkload.workloadId,
    clusterId: distributedInferenceWorkload.clusterId,
    projectName: "",
    projectId: distributedInferenceWorkload.projectId,
    departmentName: "",
    departmentId: distributedInferenceWorkload.departmentId || "",
    namespace: "",
    createdAt: distributedInferenceWorkload.createdAt,
    phase: distributedInferenceWorkload.actualPhase || Phase.Creating,
    k8sPhase: "",
    tenantId: tenantId,
    runningPods: 0,
    phaseUpdatedAt: "",
    k8sPhaseUpdatedAt: "",
    updatedAt: "",
    deletedAt: null,
    submittedBy: distributedInferenceWorkload.createdBy,
    priority: null,
  };
}

function getEmptyUINimInferenceCreation(specificEnv?: IUIWorkloadSpecificEnv): IUIInferenceCreation {
  const uiWorkload: IUIWorkloadCreation = getEmptyUIInferenceCreation(specificEnv);
  const modelCreateUi: IUIInferenceCreation = {
    ...uiWorkload,
    image: "",
    serverType: EServerType.Nim,
    accessFromStorage: false,
    accessModelMethod: ENimAccessModelMethod.Manually,
    architecture: EInferenceArchitecture.Standard,
  };
  return modelCreateUi;
}

function getEmptyUIInferenceCreationForTemplate(specificEnv?: IUIWorkloadSpecificEnv): IUIInferenceTemplateCreation {
  const uiWorkload: IUIInferenceCreation = getEmptyUIInferenceCreation(specificEnv);

  return {
    ...uiWorkload,
    serverType: EServerType.Nim,
    templateName: "",
    templateDescription: "",
  };
}

function getEmptyUIInferenceCreation(specificEnv?: IUIWorkloadSpecificEnv): IUIInferenceCreation {
  const uiWorkload: IUIWorkloadCreation = workloadUtil.getEmptyUIWorkloadCreation(specificEnv);
  const uiInferenceWorkload: IUIInferenceCreation = {
    ...uiWorkload,
    specificEnv: {
      ...omit(uiWorkload.specificEnv, ["backoffLimit"]),
      autoScaleData: getUIAutoScaleEmptyState(),
      servingPortAccess: getUIServingPortEmptyState(),
    },
    serverType: EServerType.Nim,
    architecture: EInferenceArchitecture.Standard,
  };

  return uiInferenceWorkload;
}

function getUIAutoScaleEmptyState(): IAutoScaleData {
  return {
    minReplicas: MIN_REPLICAS_DEFAULT_VALUE,
    maxReplicas: MAX_REPLICAS_DEFAULT_VALUE,
    thresholdMetric: undefined,
    thresholdValue: undefined,
    scaleToZeroRetentionSeconds: undefined,
  };
}

function getUIServingPortEmptyState(): IServingEndpointAccessData {
  return {
    access: {
      authorizedUsers: null,
      authorizedGroups: null,
      accessOption: EAccessOptions.PUBLIC,
    },
    port: SERVING_PORT_SERVER_PORT,
    protocol: "http",
  };
}

function getInferenceComputeResourceData(
  workload: IUIWorkloadCreation,
  projectNodepools: Array<Resources> | undefined | null = [],
): IComputeSectionData {
  const computeData: IComputeSectionData = workloadCreateFormUtil.getComputeResourceData(workload, projectNodepools);
  computeData.autoScaleData = workload.specificEnv.autoScaleData;
  return computeData;
}

function getEnvironmentRunParamsData(workload: IUIWorkloadCreation): IUIWorkloadEnvSectionModel {
  const envSectionData: IUIWorkloadEnvSectionModel = workloadCreateFormUtil.getEnvironmentRunParamsData(workload);
  envSectionData.servingPortAccess = workload.specificEnv.servingPortAccess;
  return envSectionData;
}

function getHuggingFaceTokenValue(environmentVariables: Array<EnvironmentVariableOfAsset> = []): string {
  const tokenEnv: EnvironmentVariableOfAsset | undefined = environmentVariables.find(
    (env: EnvironmentVariableOfAsset) => env.name === EHuggingFaceEnvVariableNames.HF_TOKEN && !env.credential,
  );

  return tokenEnv?.value || "";
}

function getEnvironmentVariableValue(
  keyName: string,
  environmentVariables: Array<EnvironmentVariableOfAsset> = [],
): string {
  const modelEnv: EnvironmentVariableOfAsset | undefined = environmentVariables.find(
    (env: EnvironmentVariableOfAsset) => env.name === keyName,
  );

  return modelEnv?.value || "";
}

function inferCredentialSourceFromEnvVar(env: EnvironmentVariable | undefined): ERunaiCredentialSource | undefined {
  if (!env || env.value) {
    return undefined;
  }
  if (env.secret?.name) {
    return ERunaiCredentialSource.SharedSecret;
  }
  if (env.userCredential?.name) {
    return ERunaiCredentialSource.MyCredentials;
  }
  return undefined;
}

function replaceEnvironmentVariable(
  environmentVariables: Array<EnvironmentVariable> | undefined | null,
  name: string,
  value?: string,
  secret?: EnvironmentVariableSecret,
  userCredentials?: EnvironmentVariableUserCredential,
): Array<EnvironmentVariable> {
  let envs = environmentVariables || [];
  envs = envs.filter((env: EnvironmentVariable) => env.name !== name);

  if (secret) {
    envs.push({ name, secret });
  } else if (userCredentials) {
    envs.push({ name, userCredential: userCredentials });
  } else if (value) {
    envs.push({ name, value });
  }

  return envs;
}

function excludeEnvironmentVariablesByName(
  keyNames: string | Array<string>,
  environmentVariables: Array<EnvironmentVariableOfAsset | EnvironmentVariable> = [],
): Array<EnvironmentVariableOfAsset | EnvironmentVariable> {
  const keyNamesArray: Array<string> = Array.isArray(keyNames) ? keyNames : [keyNames];
  const keyNamesMap = new Set<string>(keyNamesArray);
  return environmentVariables.filter(
    (env: EnvironmentVariableOfAsset | EnvironmentVariable) => env.name && !keyNamesMap.has(env.name),
  );
}

function getHFInferenceWorkloadCreationRequest(
  workload: IUIInferenceCreation,
  compute: WorkspaceSpecSpecAllOfCompute,
  workloadCreateOptions?: IWorkloadCreateOptions,
): InferenceCreationRequest {
  const workloadCreationRequest: InferenceCreationRequest = _getFlatInferenceWorkloadData(workload, compute);

  // hf specific params
  if (!workloadCreationRequest.spec) {
    workloadCreationRequest.spec = {};
  }

  if (!workload.serverType) {
    throw new Error("Missing server type: The workload object must include a valid 'serverType' property.");
  }

  const modelFlag = workload.serverType === EServerType.TGI ? "--model-id" : "--model";
  workloadCreationRequest.spec.args = `${modelFlag} ${workload.modelName} ${
    workloadCreationRequest.spec.args || ""
  }`.trim();
  workloadCreationRequest.spec.image = workload.image;
  workloadCreationRequest.spec.servingPort = {
    protocol: (workload.specificEnv.servingPortAccess?.protocol as ServingPortProtocol | undefined) || "http",
    container: workload.specificEnv.servingPortAccess?.port || SERVING_PORT_SERVER_PORT,
    clusterLocalAccessOnly: workload.specificEnv.servingPortAccess?.clusterLocalAccessOnly,
    ...servingPortAccessAuthorization(workload.specificEnv.servingPortAccess?.access),
  };

  if (workloadCreateOptions?.storage) {
    workloadCreationRequest.spec = _fillDatasourceData(
      workloadCreationRequest.spec,
      workloadCreateOptions.storage.dataSource,
      workloadCreateOptions.storage.kind,
      EHuggingFaceEnvVariableNames.HF_HOME,
    );
  }

  if (workloadCreationRequest.spec.compute && workloadCreationRequest.spec.compute.gpuDevicesRequest === 0) {
    workloadCreationRequest.spec.compute.gpuDevicesRequest = undefined;
  }

  workloadCreationRequest.spec.environmentVariables = _handleHuggingFaceEnvironmentVariables(
    workloadCreationRequest.spec.environmentVariables || [],
    workload.token,
    {
      genericSecretName: workloadCreateOptions?.genericSecretName,
      addHomeDir: workloadCreateOptions?.addHomeDir,
    },
  );

  if (workload.specificEnv.servingConfiguration) {
    workloadCreationRequest.spec.servingConfiguration = workload.specificEnv.servingConfiguration;
  }

  return workloadCreationRequest;
}

function getEmptyUISuperWorkloadModel({
  scopeType,
  scopeId,
  name,
  clusterId,
  description,
}: {
  scopeType?: ScopeType;
  scopeId?: string;
  name?: string;
  clusterId?: string;
  description?: string;
}): IUISuperWorkloadCreationRequest {
  return {
    name: name || "",
    projectId: "",
    clusterId: clusterId || "",
    scopeType: scopeType,
    scopeId: scopeId,
    description: description || "",
    spec: { connections: [] },
  };
}

/**
 *
 * @param environmentVariables - the environment variables to add to the workload
 * @param token - the token to add to the workload
 * @param options
 *  - addHomeDir: add the hardcoded home dir env variable to the workload
 *  - genericSecretName: add the token from a generic secret. if provided, the token will be added to the HF_TOKEN env variable and will remove the token from the environment variables
 * @returns
 */
function _handleHuggingFaceEnvironmentVariables(
  environmentVariables: Array<EnvironmentVariable>,
  token?: string,
  options?: {
    genericSecretName?: string;
    addHomeDir?: boolean;
  },
): Array<EnvironmentVariable> {
  let envVars = environmentVariables;

  if (token) {
    envVars = excludeEnvironmentVariablesByName(EHuggingFaceEnvVariableNames.HF_TOKEN, envVars);
    envVars.push({
      name: EHuggingFaceEnvVariableNames.HF_TOKEN,
      value: token || "",
    });
  }

  if (options?.genericSecretName) {
    envVars = excludeEnvironmentVariablesByName(EHuggingFaceEnvVariableNames.HF_TOKEN, envVars);
    envVars.push({
      name: EHuggingFaceEnvVariableNames.HF_TOKEN,
      secret: {
        name: options.genericSecretName,
        key: EHuggingFaceEnvVariableNames.HF_TOKEN,
      },
    });
  }

  const isEnvExist = envVars.find(
    (env: EnvironmentVariableOfAsset) => env.name === EHuggingFaceEnvVariableNames.HF_HOME,
  );

  if (!isEnvExist) {
    envVars.push({
      name: EHuggingFaceEnvVariableNames.HF_HOME,
      value: HF_HOME_DEFAULT_PATH,
    });
  }

  // on openshift, we need to add the home dir env variable, since the default path is not writable
  // and lead to permission error on /.cache since the default HOME is "/"
  if (options?.addHomeDir) {
    envVars.push({
      name: EHuggingFaceEnvVariableNames.HOME,
      value: HOME_DEFAULT_PATH,
    });
  }

  return envVars;
}

function _fillDatasourceData(
  nimWorkloadSpec: InferenceSpecSpec,
  storage: PvcInstance | HostPathInstance | NfsInstance | undefined,
  kind: keyof SupersetSpecAllOfStorage | undefined,
  envVarName: string,
): InferenceSpecSpec {
  if (kind) {
    nimWorkloadSpec.storage = {
      [kind]: [storage],
    };
  }

  nimWorkloadSpec.environmentVariables = excludeEnvironmentVariablesByName(
    envVarName,
    nimWorkloadSpec.environmentVariables || [],
  );
  const path: string | undefined | null = storage?.path;

  if (path) {
    nimWorkloadSpec.environmentVariables?.push({ name: envVarName, value: path });
  }

  return nimWorkloadSpec;
}

function _addNimEnvironmentVariables(
  environmentVariables: Array<EnvironmentVariableOfAsset>,
): Array<EnvironmentVariableOfAsset> {
  environmentVariables = excludeEnvironmentVariablesByName(
    [
      ENimEnvVariableNames.OutlinesCacheDir,
      ENimEnvVariableNames.NimServerPort,
      ENimEnvVariableNames.NimJsonlLogging,
      ENimEnvVariableNames.NimLogLevel,
    ],
    environmentVariables,
  );

  environmentVariables?.push(
    { name: ENimEnvVariableNames.OutlinesCacheDir, value: "/tmp/outlines" },
    { name: ENimEnvVariableNames.NimServerPort, value: String(SERVING_PORT_SERVER_PORT) },
    { name: ENimEnvVariableNames.NimJsonlLogging, value: "1" },
    { name: ENimEnvVariableNames.NimLogLevel, value: "INFO" },
  );

  return environmentVariables;
}

function _getFlatInferenceWorkloadData(
  workload: IUIInferenceCreation,
  computeFields: WorkspaceSpecSpecAllOfCompute,
  environment?: EnvironmentAsset,
): InferenceCreationRequest {
  const specificEnv = workload.specificEnv;
  const workloadCreationRequest: InferenceCreationRequest = _getFlatWorkloadData(workload, computeFields, environment);

  if (workloadCreationRequest.spec && specificEnv.autoScaleData) {
    workloadCreationRequest.spec.autoscaling = convertAutoScaleUIToFlat(specificEnv.autoScaleData);
  }

  return workloadCreationRequest;
}

function convertAutoScaleUIToFlat(autoScaleData: IAutoScaleData): AutoScaling {
  return {
    minReplicas: autoScaleData.minReplicas,
    maxReplicas: autoScaleData.maxReplicas,
    scaleToZeroRetentionSeconds: autoScaleData.scaleToZeroRetentionSeconds || undefined,
    metric: autoScaleData.thresholdMetric || undefined,
    metricThreshold: autoScaleData.thresholdValue || undefined,
  };
}

function convertAutoScaleFlatToUI(autoScaleData: AutoScaling): IAutoScaleData {
  return {
    minReplicas: autoScaleData.minReplicas === null ? undefined : autoScaleData.minReplicas,
    maxReplicas: autoScaleData.maxReplicas === null ? undefined : autoScaleData.maxReplicas,
    scaleToZeroRetentionSeconds: autoScaleData.scaleToZeroRetentionSeconds || undefined,
    thresholdMetric: autoScaleData.metric || undefined,
    thresholdValue: autoScaleData.metricThreshold || undefined,
  };
}

function _getFlatWorkloadData(
  workload: IUIInferenceCreation,
  computeFields: WorkspaceSpecSpecAllOfCompute,
  environment?: EnvironmentAsset,
): InferenceCreationRequest {
  const specificEnv = workload.specificEnv || {};
  const workloadCreationRequest: InferenceCreationRequest = {
    name: workload.name,
    projectId: workload.projectId.toString(),
    clusterId: workload.clusterId,
    spec: {
      command: specificEnv.command || undefined,
      args: specificEnv.args || undefined,
      nodeType: specificEnv.nodeType || undefined,
      nodePools: specificEnv.nodePools,
      podAffinity: specificEnv.podAffinity || undefined,
      environmentVariables: specificEnv.environmentVariables,
      annotations: specificEnv.annotations?.map((env: IItemizedListItem) => pick(env, "name", "value", "exclude")),
      labels: specificEnv.labels?.map((env: IItemizedListItem) => pick(env, "name", "value", "exclude")),
      tolerations: specificEnv.tolerations,
      compute: computeFields,
    },
  };

  if (environment) {
    workloadCreationRequest.spec = {
      ...workloadCreationRequest.spec,
      image: environment.spec.image || undefined,
      imagePullPolicy: environment.spec.imagePullPolicy || undefined,
      workingDir: environment.spec.workingDir || undefined,
      createHomeDir: environment.spec.createHomeDir || undefined,
    };
  }
  return workloadCreationRequest;
}

/// ---->
function getInferenceCreationRequestV2(
  inference: IUIWorkloadCreation,
  supportingServingPortAccess = false,
  workloadVolumes?: Array<string>,
): InferenceCreationRequestV2 {
  if (!inference.assets.environment) throw "Environment details are missing";
  const workloadCreationRequest: InferenceCreationRequestV2 = {
    meta: {
      name: inference.name,
      projectId: inference.projectId.toString(),
      clusterId: inference.clusterId,
    },
    assets: workloadUtil.getWorkloadAssetsFromUIWorkloadAssets(inference.assets, workloadVolumes),
    specificRunParams: getInferenceRequestV2SpecificRunParams(inference.specificEnv, supportingServingPortAccess),
  };
  return workloadCreationRequest;
}

function normalizeAutoScaleData(
  autoScalingSpec: SpecificRunAutoScalingAutoScaling | null | undefined,
  policy?: InferencePolicyDefaultsAndRulesV2,
): IAutoScaleData {
  let autoScaling: AutoScaling = {
    minReplicas: autoScalingSpec?.minReplicas,
    maxReplicas: autoScalingSpec?.maxReplicas,
    metric: autoScalingSpec?.thresholdMetric,
    metricThreshold: autoScalingSpec?.thresholdValue,
    scaleToZeroRetentionSeconds: autoScalingSpec?.scaleToZeroRetentionSeconds,
  };

  autoScaling = getAutoScalingData(autoScaling, policy as unknown as UnifiedPolicyInfoPerReplicaType | undefined);
  const autoScaleData: IAutoScaleData = {
    minReplicas: autoScaling.minReplicas === null ? undefined : autoScaling.minReplicas,
    maxReplicas: autoScaling.maxReplicas === null ? undefined : autoScaling.maxReplicas,
    thresholdMetric: autoScaling.metric === null ? undefined : autoScaling.metric,
    thresholdValue: autoScaling.metricThreshold === null ? undefined : autoScaling.metricThreshold,
    scaleToZeroRetentionSeconds:
      autoScaling.scaleToZeroRetentionSeconds === null ? undefined : autoScaling.scaleToZeroRetentionSeconds,
  };

  return autoScaleData;
}

function convertInferenceToWorkloadUI(
  workload: TWorkloadsToConvert,
  inferenceType?: EInferenceType,
  supportingServingPortAccess = false,
  uiVolumes?: Array<IUIVolume>,
  modelName?: string,
  policy?: InferencePolicyDefaultsAndRulesV2,
  templateId?: string,
  architecture?: EInferenceArchitecture,
): IUIInferenceCreation {
  const baseWorkload: IUIWorkloadCreation = workloadUtil.convertWorkloadToWorkloadUI(workload, uiVolumes);
  const inferenceUI: IUIInferenceCreation = {
    ...baseWorkload,
    inferenceType,
    templateId,
    architecture: EInferenceArchitecture.Standard,
    serverType: EServerType.Custom,
  };

  if (modelName) {
    inferenceUI.modelName = modelName;
  }

  inferenceUI.specificEnv.autoScaleData = normalizeAutoScaleData(workload.specificRunParams?.autoScaling, policy);

  if (supportingServingPortAccess) {
    inferenceUI.specificEnv.servingPortAccess = {
      access: servingPortInternalToolInfo(workload.specificRunParams?.servingPortAccess),
    };
  }

  // delete inferenceUI.specificEnv.allowOverQuota;
  return {
    ...inferenceUI,
    inferenceType: inferenceType || EInferenceType.AssetsBased,
    architecture: architecture || EInferenceArchitecture.Standard,
  };
}

function servingPortInternalToolInfo(
  servingPortAccess:
    | SpecificRunServingPortAccessServingPortAccess
    | ServingPort
    | SupersetServingPort
    | undefined
    | null,
): IConnectionAccess {
  return {
    accessOption: _getServingPortAccessOption(servingPortAccess),
    authorizedGroups: servingPortAccess?.authorizedGroups,
    authorizedUsers: servingPortAccess?.authorizedUsers,
  };
}

function _getServingPortAccessOption(
  servingPortAccess:
    | SpecificRunServingPortAccessServingPortAccess
    | ServingPort
    | SupersetServingPort
    | undefined
    | null,
): EAccessOptions {
  let authorizationType: EAccessOptions = EAccessOptions.PUBLIC;
  if (!servingPortAccess) return authorizationType;
  switch (servingPortAccess.authorizationType) {
    case ServingPortAccessAuthorizationTypeEnum.Public:
      authorizationType = EAccessOptions.PUBLIC;
      break;
    case ServingPortAccessAuthorizationTypeEnum.AuthenticatedUsers:
      authorizationType = EAccessOptions.EVERYONE;
      break;
    case ServingPortAccessAuthorizationTypeEnum.AuthorizedUsersOrGroups:
      if (servingPortAccess.authorizedGroups) {
        authorizationType = EAccessOptions.GROUPS;
      } else {
        authorizationType = EAccessOptions.SPECIFIC_USERS;
      }
  }

  return authorizationType;
}

function getInferenceRequestV2SpecificRunParams(
  specificEnv: IUIWorkloadSpecificEnv,
  supportingServingPortAccess = false,
): InferenceSpecificRunParams {
  const retSpecificEnv: InferenceSpecificRunParams = workloadUtil.processCommonSpecificEnv(specificEnv, [
    "allowOverQuota",
    "backoffLimit",
    "autoScaleData",
    "servingPortAccess",
  ]);

  retSpecificEnv.autoScaling = specificEnv.autoScaleData;

  if (supportingServingPortAccess && specificEnv.servingPortAccess?.access) {
    retSpecificEnv.servingPortAccess = servingPortAccessAuthorization(specificEnv.servingPortAccess.access);
  }

  return retSpecificEnv;
}

function isLatencyEnableOnCluster(currCluster: ClustersMinimalData): boolean {
  if (!currCluster.dependenciesStatus?.optional?.knative) return false;
  const dependencyStatus: Record<string, ClusterDependencyStatus> | undefined =
    currCluster.dependenciesStatus.optional.knative.components;
  return (dependencyStatus && dependencyStatus["hpa"] && dependencyStatus["hpa"].available) || false;
}

function isAutoScaleEnableOnCluster(currCluster: ClustersMinimalData): boolean {
  if (!currCluster.dependenciesStatus?.optional?.knative) return false;
  const dependencyStatus: Record<string, ClusterDependencyStatus> | undefined =
    currCluster.dependenciesStatus.optional.knative.components?.knative?.components;

  return (
    (dependencyStatus &&
      dependencyStatus["enable-scale-to-zero"] &&
      dependencyStatus["enable-scale-to-zero"].available) ||
    false
  );
}

function servingPortAccessAuthorization(
  servingPortAccess: IConnectionAccess | null | undefined,
): SpecificRunServingPortAccessServingPortAccess | ServingPort {
  return {
    authorizationType: _getAuthorizationType(servingPortAccess),
    authorizedGroups: servingPortAccess?.authorizedGroups,
    authorizedUsers: servingPortAccess?.authorizedUsers,
  };
}

function _getAuthorizationType(
  servingPortAccess: IConnectionAccess | null | undefined,
): ServingPortAccessAuthorizationTypeEnum | undefined {
  let authorizationType: ServingPortAccessAuthorizationTypeEnum | undefined = undefined;

  if (!servingPortAccess?.accessOption) return authorizationType;

  switch (servingPortAccess.accessOption) {
    case EAccessOptions.PUBLIC:
      // set to undefined since no need to send it to the server
      break;
    case EAccessOptions.EVERYONE:
      authorizationType = ServingPortAccessAuthorizationTypeEnum.AuthenticatedUsers;
      break;
    case EAccessOptions.GROUPS:
    case EAccessOptions.SPECIFIC_USERS:
      authorizationType = ServingPortAccessAuthorizationTypeEnum.AuthorizedUsersOrGroups;
  }

  return authorizationType;
}

/**
 * Retrieves the list of data source types supported for inference workloads.
 *
 * By default, the function includes the following data sources:
 * - PVC (Persistent Volume Claim)
 * - Git repository
 * - ConfigMap
 *
 * If `isVersion219OrHigher` is `true`, additional data sources supported from cluster version 2.19 are included:
 * - HostPath
 * - NFS (Network File System)
 * - DataVolume (when not in model store only mode)
 *
 * @param {Object} options - Configuration options.
 * @param {boolean} options.isVersion219OrHigher - Whether the cluster is version 2.19 or later.
 * @returns {TDataSourceKinds[]} A list of supported data source type names.
 */
function getInferenceSupportedDataSources(options: {
  isVersion219OrHigher: boolean;
  modelStoreSupportOnly?: boolean;
}): TDataSourceKinds[] {
  const BASE_SUPPORTED_DATA_SOURCES: TDataSourceKinds[] = [AssetKind.Pvc];
  const EXTENDED_DATA_SOURCES: TDataSourceKinds[] = [AssetKind.Git, AssetKind.ConfigMap];
  const VERSION_219_DATA_SOURCES: TDataSourceKinds[] = [AssetKind.HostPath, AssetKind.Nfs];
  const VERSION_219_EXTENDED_DATA_SOURCES: TDataSourceKinds[] = [AssetKind.DataVolume];

  const supportedDataSources: TDataSourceKinds[] = [...BASE_SUPPORTED_DATA_SOURCES];
  if (!options.modelStoreSupportOnly) {
    supportedDataSources.push(...EXTENDED_DATA_SOURCES);
  }

  if (options.isVersion219OrHigher) {
    supportedDataSources.push(...VERSION_219_DATA_SOURCES);

    if (!options.modelStoreSupportOnly) {
      supportedDataSources.push(...VERSION_219_EXTENDED_DATA_SOURCES);
    }
  }

  return supportedDataSources;
}

function getHybridWorkloadUIModel(inference: InferenceFlat): IUISuperWorkloadCreationRequest {
  const uiModel: IUISuperWorkloadCreationRequest = {
    ...inference,
    spec: {
      ...inference.spec,
      connections: workloadHybridUtil.getConnectionsField(
        inference.spec?.ports || [],
        inference.spec?.exposedUrls || [],
        inference.spec?.relatedUrls || [],
      ),
      compute: workloadHybridUtil.getComputeField(inference.spec?.compute),
      storage: workloadHybridUtil.getStorageField(inference.spec?.storage),
      security: workloadHybridUtil.getSecurityField(inference.spec?.security),
    },
  };

  // Bug fix: When the user sets `minReplicas` to 0 in the API but does not specify
  // `scaleToZeroRetentionSeconds`, Knative defaults this value to 1 minute.
  // To ensure consistency, we explicitly set it to 60 seconds in the UI.
  if (uiModel.spec.autoscaling?.minReplicas === 0 && uiModel.spec.autoscaling.scaleToZeroRetentionSeconds == undefined) {
    uiModel.spec.autoscaling.scaleToZeroRetentionSeconds = 60;
  }

  return uiModel;
}

function prepareUpdateInferenceFlatRequest(
  inference: InferenceFlat,
  modelUi: IUISuperWorkloadCreationRequest,
): InferenceUpdateSpecSpec {
  const originSpec = inference.spec;
  const newSpec = modelUi.spec;

  const compute: WorkspaceSpecSpecAllOfCompute | undefined = workloadHybridUtil._getCleanedComputeFieldsFromUI(
    newSpec.compute || {},
  );
  return {
    command: originSpec?.command !== newSpec.command ? newSpec.command : undefined,
    args: originSpec?.args !== newSpec.args ? newSpec.args : undefined,
    image: originSpec?.image !== newSpec.image ? newSpec.image : undefined,
    imagePullPolicy: originSpec?.imagePullPolicy !== newSpec.imagePullPolicy ? newSpec.imagePullPolicy : undefined,
    workingDir: originSpec?.workingDir !== newSpec.workingDir ? newSpec.workingDir : undefined,
    createHomeDir: originSpec?.createHomeDir !== newSpec.createHomeDir ? newSpec.createHomeDir : undefined,
    probes: isEqual(originSpec?.probes, newSpec.probes) ? undefined : newSpec.probes,
    environmentVariables: isEqual(originSpec?.environmentVariables, newSpec.environmentVariables)
      ? undefined
      : newSpec.environmentVariables,
    autoscaling: isEqual(originSpec?.autoscaling, newSpec.autoscaling) ? undefined : newSpec.autoscaling,
    compute: isEqual(compute, originSpec?.compute) ? undefined : compute,
    nodePools: isEqual(originSpec?.nodePools, newSpec.nodePools) ? undefined : modelUi.spec.nodePools,
    nodeType: originSpec?.nodeType !== newSpec.nodeType ? undefined : modelUi.spec.nodeType,
    servingConfiguration: isEqual(originSpec?.servingConfiguration, newSpec.servingConfiguration)
      ? undefined
      : newSpec.servingConfiguration,
    imagePullSecrets: isEqual(originSpec?.imagePullSecrets, newSpec.imagePullSecrets)
      ? undefined
      : newSpec.imagePullSecrets,
  };
}

function _ensureStoragePathEnvVar(
  environmentVariables: Array<EnvironmentVariableOfAsset | EnvironmentVariable>,
  storage: IUIStorageFields | undefined,
  envVarName: string,
): Array<EnvironmentVariableOfAsset | EnvironmentVariable> {
  let envs = environmentVariables || [];
  envs = excludeEnvironmentVariablesByName(envVarName, envs);
  let path: string | undefined | null;
  if (storage?.nfs && storage.nfs[0]?.path) {
    path = storage.nfs[0].path;
  } else if (storage?.pvc && storage.pvc[0]?.path) {
    path = storage.pvc[0].path;
  } else if (storage?.hostPath && storage.hostPath[0]?.path) {
    path = storage.hostPath[0].path;
  }
  if (path) {
    envs.push({ name: envVarName, value: path });
  }
  return envs;
}

function removeEmptyStorageArrays(storage: IUIStorageFields | undefined): IUIStorageFields | undefined {
  if (!storage) return storage;
  const cleanedStorage = { ...storage };
  Object.keys(cleanedStorage).forEach((key) => {
    const arr = cleanedStorage[key as keyof IUIStorageFields] as Array<IUIStorageInstance> | undefined;
    if (!arr?.length) {
      delete cleanedStorage[key as keyof IUIStorageFields];
    }
  });
  return cleanedStorage;
}

// this function is used to order the Nim environment variables in the correct order
// this is necessary only because the e2e tests expects the environment variables in the correct order
function _orderNimEnvironmentVariables(
  environmentVariables: Array<EnvironmentVariableOfAsset | EnvironmentVariable>,
): Array<EnvironmentVariableOfAsset | EnvironmentVariable> {
  const NIM_ENV_VAR_ORDER: string[] = [
    ENimEnvVariableNames.OutlinesCacheDir,
    ENimEnvVariableNames.NimServerPort,
    ENimEnvVariableNames.NimJsonlLogging,
    ENimEnvVariableNames.NimLogLevel,
    ENimEnvVariableNames.NimModelProfile,
    ENimEnvVariableNames.NgcApiKey,
    ENimEnvVariableNames.NimCachePath,
  ];

  const byName = new Map<string, EnvironmentVariableOfAsset | EnvironmentVariable>();
  for (const env of environmentVariables) {
    if (env.name) byName.set(env.name, env);
  }
  const ordered: Array<EnvironmentVariableOfAsset | EnvironmentVariable> = [];
  for (const name of NIM_ENV_VAR_ORDER) {
    const env = byName.get(name);
    if (env) {
      ordered.push(env);
      byName.delete(name);
    }
  }
  byName.forEach((env) => ordered.push(env));
  return ordered;
}

function _cleanStorageAndSetEnvVar(model: IUISuperWorkloadCreationRequest, envVarName: string): void {
  model.spec.storage = removeEmptyStorageArrays(model.spec.storage);
  if (Object.keys(model.spec.storage || {}).length === 0) return;
  model.spec.environmentVariables = _ensureStoragePathEnvVar(
    model.spec.environmentVariables || [],
    model.spec.storage,
    envVarName,
  );
}

function _ensureDefaultEnvVar(
  environmentVariables: Array<EnvironmentVariableOfAsset | EnvironmentVariable>,
  name: string,
  value: string,
): Array<EnvironmentVariableOfAsset | EnvironmentVariable> {
  const exists = environmentVariables.some((env) => env.name === name);
  if (!exists) {
    environmentVariables.push({ name, value });
  }
  return environmentVariables;
}

// this function is used to order the Hugging Face environment variables in the correct order
// this is necessary only because the e2e tests expects the environment variables in the correct order
function _orderHfEnvironmentVariables(
  environmentVariables: Array<EnvironmentVariableOfAsset | EnvironmentVariable>,
): Array<EnvironmentVariableOfAsset | EnvironmentVariable> {
  const HF_ENV_VAR_ORDER: string[] = [EHuggingFaceEnvVariableNames.HF_HOME, EHuggingFaceEnvVariableNames.HF_TOKEN];
  const byName = new Map<string, EnvironmentVariableOfAsset | EnvironmentVariable>();
  for (const env of environmentVariables) {
    if (env.name) byName.set(env.name, env);
  }
  const ordered: Array<EnvironmentVariableOfAsset | EnvironmentVariable> = [];
  for (const name of HF_ENV_VAR_ORDER) {
    const env = byName.get(name);
    if (env) {
      ordered.push(env);
      byName.delete(name);
    }
  }
  byName.forEach((env) => ordered.push(env));
  return ordered;
}

function adjustHFWorkloadProperties(model: IUISuperWorkloadCreationRequest): IUISuperWorkloadCreationRequest {
  _cleanStorageAndSetEnvVar(model, EHuggingFaceEnvVariableNames.HF_HOME);
  model.spec.environmentVariables = _ensureDefaultEnvVar(
    model.spec.environmentVariables || [],
    EHuggingFaceEnvVariableNames.HF_HOME,
    HF_HOME_DEFAULT_PATH,
  );
  model.spec.environmentVariables = _orderHfEnvironmentVariables(model.spec.environmentVariables ?? []);
  return model;
}

function adjustNimWorkloadProperties(model: IUISuperWorkloadCreationRequest): IUISuperWorkloadCreationRequest {
  _cleanStorageAndSetEnvVar(model, ENimEnvVariableNames.NimCachePath);
  model.spec.environmentVariables = _orderNimEnvironmentVariables(model.spec.environmentVariables ?? []);
  model.spec.compute = {
    ...model.spec.compute,
    largeShmRequest: true,
  };
  return model;
}

function setInferenceHuggingFaceUIDefaults(
  uiModel: IUISuperWorkloadCreationRequest,
  isSSO: boolean,
  policy?: UnifiedPolicyInfoPerReplicaType | null,
): IUISuperWorkloadCreationRequest {
  const workloadUI: IUISuperWorkloadCreationRequest = setInferenceUIDefaults(uiModel, isSSO, policy);

  workloadUI.spec.servingPort = {
    ...workloadUI.spec.servingPort,
    protocol: "http",
    container: SERVING_PORT_SERVER_PORT,
  };

  workloadUI.spec.image = LLM_IMAGE_FOR_HUGGING_FACE;

  return workloadUI;
}

function setInferenceNIMUIDefaults(
  uiModel: IUISuperWorkloadCreationRequest,
  isSSO: boolean,
  policy?: UnifiedPolicyInfoPerReplicaType | null,
): IUISuperWorkloadCreationRequest {
  const workloadUI: IUISuperWorkloadCreationRequest = setInferenceUIDefaults(uiModel, isSSO, policy);

  workloadUI.spec.environmentVariables = _addNimEnvironmentVariablesIfNeeded(
    workloadUI.spec.environmentVariables || [],
    policy,
    workloadUI.spec.servingPort?.container,
  );

  if (!workloadUI.spec.servingPort) {
    workloadUI.spec.servingPort = {
      protocol: "http",
    };
  }

  workloadUI.spec.security = {
    ...workloadUI.spec.security,
    runAsGid: 1000,
    runAsUid: 1000,
  };

  if (!workloadUI.spec.labels?.some((label) => label.name === NIM_WORKLOAD_LABEL)) {
    workloadUI.spec.labels = [...(workloadUI.spec.labels || []), { name: NIM_WORKLOAD_LABEL, value: "true" }];
  }

  if (!workloadUI.spec.servingPort.container) {
    workloadUI.spec.servingPort.container = SERVING_PORT_SERVER_PORT;
  }
  return workloadUI;
}

function setInferenceUIDefaults(
  uiModel: IUISuperWorkloadCreationRequest,
  isSSO: boolean,
  policy?: UnifiedPolicyInfoPerReplicaType | null,
): IUISuperWorkloadCreationRequest {
  const workloadUI: IUISuperWorkloadCreationRequest = workloadHybridUtil.setBaseWorkloadUIDefaults(
    uiModel,
    isSSO,
    policy,
  );

  if (!workloadUI.spec.servingPort) {
    workloadUI.spec.servingPort = {
      protocol: "http",
    };
  }

  if (!workloadUI.spec.servingPort.protocol) {
    workloadUI.spec.servingPort = {
      ...workloadUI.spec.servingPort,
      protocol: "http",
    };
  }

  workloadUI.spec.autoscaling = getAutoScalingData(uiModel.spec.autoscaling || undefined, policy);

  return {
    ...workloadUI,
    spec: {
      ...workloadUI.spec,
    },
  };
}

function getHybridInferenceModel(
  unifiedWorkload: UnifiedWorkload,
  clearEnvironmentVariables = false,
): IUISuperWorkloadCreationRequest {
  const workload: IUISuperWorkloadCreationRequest = workloadHybridUtil.getHybridBaseWorkloadModel(unifiedWorkload);
  workload.spec.priorityClass = unifiedWorkload.spec?.worker?.priorityClass;
  workload.spec.preemptibility = unifiedWorkload.spec?.worker?.preemptibility;
  workload.spec.servingPort = unifiedWorkload.spec?.worker?.servingPort;
  workload.spec.autoscaling = unifiedWorkload.spec?.worker?.autoscaling;
  workload.spec.servingConfiguration = unifiedWorkload.spec?.worker?.servingConfiguration;

  // Bug fix: When the user sets `minReplicas` to 0 in the API but does not specify
  // `scaleToZeroRetentionSeconds`, Knative defaults this value to 1 minute.
  // To ensure consistency, we explicitly set it to 60 seconds in the UI.
  if (
    workload.spec.autoscaling?.minReplicas === 0 &&
    workload.spec.autoscaling.scaleToZeroRetentionSeconds == undefined
  ) {
    workload.spec.autoscaling.scaleToZeroRetentionSeconds = 60;
  }

  if (workload.spec.storage?.volumes) {
    workload.spec.storage.volumes = workload.spec.storage.volumes.map((volume: IUIPvcInstance) => {
      delete volume.ephemeral;
      return volume;
    });
  }

  // Clear environment variables if requested
  if (clearEnvironmentVariables && workload.spec.environmentVariables) {
    workload.spec.environmentVariables = workload.spec.environmentVariables.map((env) => {
      if (
        (env.name === ENimEnvVariableNames.NgcApiKey && !env.secret && !env.userCredential) ||
        (env.name === EHuggingFaceEnvVariableNames.HF_TOKEN && !env.secret && !env.userCredential)
      )
        return { ...env, value: "" };
      return env;
    });
  }

  return workload;
}

function getHybridDistributedModel(unifiedWorkload: UnifiedWorkload): IUISuperWorkloadCreationRequest {
  const leader = unifiedWorkload.spec?.worker?.leader || {};
  const worker = unifiedWorkload.spec?.worker?.worker || {};

  const uiDist: IUISuperWorkloadCreationRequest = {
    name: unifiedWorkload.meta?.name || "",
    projectId: unifiedWorkload.meta?.projectId || "",
    clusterId: unifiedWorkload.meta?.clusterId || "",
    spec: {
      leader: {
        annotations: leader?.annotations || null,
        args: leader?.args || null,
        command: leader?.command || null,
        createHomeDir: leader?.createHomeDir || null,
        environmentVariables: leader?.environmentVariables || null,
        image: leader?.image || null,
        imagePullPolicy: leader?.imagePullPolicy || null,
        imagePullSecrets: leader?.imagePullSecrets || null,
        labels: leader?.labels || null,
        nodeAffinityRequired: leader?.nodeAffinityRequired || null,
        nodeType: leader?.nodeType || null,
        podAffinity: leader?.podAffinity || null,
        probes: leader?.probes || null,
        tolerations: leader?.tolerations || null,
        workingDir: leader?.workingDir || null,
        compute: workloadHybridUtil.getComputeField(leader?.compute),
        storage: workloadHybridUtil.getStorageField(leader?.storage),
        security: workloadHybridUtil.getSecurityField(leader?.security),
        servingPort: unifiedWorkload.spec?.worker?.servingPort,
      },
      worker: {
        annotations: worker?.annotations || null,
        args: worker?.args || null,
        command: worker?.command || null,
        createHomeDir: worker?.createHomeDir || null,
        environmentVariables: worker?.environmentVariables || null,
        image: worker?.image || null,
        imagePullPolicy: worker?.imagePullPolicy || null,
        imagePullSecrets: worker?.imagePullSecrets || null,
        labels: worker?.labels || null,
        nodeAffinityRequired: worker?.nodeAffinityRequired || null,
        nodeType: worker?.nodeType || null,
        podAffinity: worker?.podAffinity || null,
        probes: worker?.probes || null,
        tolerations: worker?.tolerations || null,
        workingDir: worker?.workingDir || null,
        compute: workloadHybridUtil.getComputeField(worker?.compute),
        storage: workloadHybridUtil.getStorageField(worker?.storage),
        security: workloadHybridUtil.getSecurityField(worker?.security),
      },
    },
  };

  return uiDist;
}

function getDistributedSpec(
  defaultUuidGidSource: UidGidSource | null,
  spec?: IUISuperWorkloadSpec,
  rules?: UnifiedPolicyInfoPerReplicaTypeRules,
): IUISuperWorkloadSpec {
  let imagePullPolicyUIDefault: ImagePullPolicy | null = spec?.imagePullPolicy || ImagePullPolicy.IfNotPresent;
  if (Array.isArray(rules?.imagePullPolicy?.options)) {
    if (!rules.imagePullPolicy.options.map((o) => o.value).includes(imagePullPolicyUIDefault)) {
      imagePullPolicyUIDefault = null;
    }
  }

  let uidGidSourceDefault = defaultUuidGidSource;
  if (Array.isArray(rules?.security?.uidGidSource?.options)) {
    if (!rules.security.uidGidSource.options.map((o) => o.value).includes(uidGidSourceDefault)) {
      uidGidSourceDefault = null;
    }
  }

  const retSpec: IUISuperWorkloadSpec = {
    ...spec,
    imagePullPolicy: imagePullPolicyUIDefault,
    security: {
      ...(spec?.security || {}),
      uidGidSource: fallbackDefaultIfNullOrUndefined(spec?.security?.uidGidSource, uidGidSourceDefault),
      capabilities: fallbackDefaultIfNullOrUndefined(spec?.security?.capabilities, []),
    },
  };

  return retSpec;
}

function setDistributedUIDefaults(
  uiModel: IUISuperWorkloadCreationRequest,
  isSSO: boolean,
  policy?: UnifiedPolicyInfo | null,
): IUISuperWorkloadCreationRequest {
  const uidGidSourceDefault: UidGidSource | null = isSSO ? UidGidSource.FromIdpToken : UidGidSource.FromTheImage;
  const leaderSpec = getDistributedSpec(
    uidGidSourceDefault,
    uiModel.spec.leader,
    policy?.worker?.rules?.leader || undefined,
  );
  const workerSpec = getDistributedSpec(
    uidGidSourceDefault,
    uiModel.spec.worker,
    policy?.worker?.rules?.worker || undefined,
  );

  if (uiModel.spec.workers === undefined) {
    uiModel.spec.workers = 1;
  }
  if (uiModel.spec.replicas === undefined) {
    uiModel.spec.replicas = 1;
  }

  if (!uiModel.spec.startupPolicy) {
    uiModel.spec.startupPolicy = DistributedInferenceStartupPolicy.LeaderCreated;
  }

  // Set default protocol for serving port
  if (!uiModel.spec.servingPort?.protocol) {
    uiModel.spec.servingPort = {
      ...uiModel.spec.servingPort,
      protocol: "http",
    };
  }

  return {
    ...uiModel,
    spec: {
      ...uiModel.spec,
      leader: leaderSpec,
      worker: workerSpec,
    },
  };
}

// Moshe: need to write unit tests for this function
function extractLeaderWorkerSpec(spec: DistributedInferenceLeaderWorkerSpec | undefined | null): IUISuperWorkloadSpec {
  return {
    ...spec,
    compute: workloadHybridUtil.getComputeField(spec?.compute),
    storage: workloadHybridUtil.getStorageField(spec?.storage),
    security: workloadHybridUtil.getSecurityField(spec?.security),
  };
}

function getServingPortFieldForDistributed(
  servingPort: SupersetServingPort | null | undefined,
): SupersetServingPort | null | undefined {
  if (!servingPort) return undefined;
  return {
    ...omit(servingPort, ["port"]),
    container: servingPort.port,
  };
}

// Moshe: need to write unit tests for this function
function getDistributedInferenceModel(unifiedWorkload: UnifiedWorkload): IUISuperWorkloadCreationRequest {
  const spec = unifiedWorkload.spec?.worker || {};

  const servingPort = getServingPortFieldForDistributed(spec.servingPort);

  const uiDist: IUISuperWorkloadCreationRequest = {
    name: unifiedWorkload.meta?.name || "",
    projectId: unifiedWorkload.meta?.projectId || "",
    clusterId: unifiedWorkload.meta?.clusterId || "",
    spec: {
      ...omit(spec, ["storage", "servingPort"]),
      leader: extractLeaderWorkerSpec(spec.leader),
      worker: extractLeaderWorkerSpec(spec.worker),
      servingPort,
    },
  };
  return uiDist;
}

function _addNimEnvironmentVariablesIfNeeded(
  environmentVariables: Array<EnvironmentVariableOfAsset>,
  policy?: UnifiedPolicyInfoPerReplicaType | null,
  servingPortContainerFromSpec?: number | null,
): Array<EnvironmentVariableOfAsset> {
  const servingPortContainer =
    servingPortContainerFromSpec != null
      ? String(servingPortContainerFromSpec)
      : policy?.defaults?.servingPort?.container?.toString() || String(SERVING_PORT_SERVER_PORT);

  const env: Record<string, EnvironmentVariableOfAsset> = {
    [ENimEnvVariableNames.OutlinesCacheDir]: { name: ENimEnvVariableNames.OutlinesCacheDir, value: "/tmp/outlines" },
    [ENimEnvVariableNames.NimServerPort]: {
      name: ENimEnvVariableNames.NimServerPort,
      value: servingPortContainer,
    },
    [ENimEnvVariableNames.NimJsonlLogging]: { name: ENimEnvVariableNames.NimJsonlLogging, value: "1" },
    [ENimEnvVariableNames.NimLogLevel]: { name: ENimEnvVariableNames.NimLogLevel, value: "INFO" },
  };

  Object.keys(env).forEach((envName: string) => {
    if (!environmentVariables.some((env) => env.name === envName)) {
      environmentVariables.push(env[envName]);
    }
  });

  return environmentVariables;
}

function getAutoScalingData(autoscaling?: AutoScaling, policy?: UnifiedPolicyInfoPerReplicaType | null): AutoScaling {
  const retAutoscaling: AutoScaling = {};

  retAutoscaling.minReplicas = getAutoscalingMinReplicasInitialValue(autoscaling, policy, MIN_REPLICAS_DEFAULT_VALUE);
  retAutoscaling.maxReplicas = getAutoscalingMaxReplicasInitialValue(
    autoscaling,
    policy,
    MAX_REPLICAS_DEFAULT_VALUE,
    retAutoscaling.minReplicas !== undefined && retAutoscaling.minReplicas !== null
      ? retAutoscaling.minReplicas
      : MIN_REPLICAS_DEFAULT_VALUE,
  );

  if (retAutoscaling.minReplicas === 0) {
    retAutoscaling.scaleToZeroRetentionSeconds = scaleDownUtil.getScaleToZeroRetentionSecondsDefaultValue(
      autoscaling?.scaleToZeroRetentionSeconds || undefined,
      policy?.defaults?.autoscaling,
      policy?.rules?.autoscaling,
    );
  }

  if (is.number(retAutoscaling.minReplicas) && is.number(retAutoscaling.maxReplicas)) {
    if (
      retAutoscaling.minReplicas === retAutoscaling.maxReplicas ||
      (retAutoscaling.minReplicas === 0 && retAutoscaling.maxReplicas === 1)
    ) {
      retAutoscaling.metric = undefined;
      retAutoscaling.metricThreshold = undefined;
    } else {
      retAutoscaling.metric = thresholdMetricUtils.getDefaultMetricOption(
        autoscaling?.metric || undefined,
        policy?.defaults?.autoscaling,
        policy?.rules?.autoscaling,
      );
      retAutoscaling.metricThreshold = thresholdMetricUtils.getThresholdMetricDefaultValue(
        autoscaling?.metricThreshold || undefined,
        policy?.defaults?.autoscaling,
        policy?.rules?.autoscaling,
      );
    }
  }

  return retAutoscaling;
}

function getAutoscalingMinReplicasInitialValue(
  autoScaling: AutoScaling | null | undefined,
  policy: UnifiedPolicyInfoPerReplicaType | null | undefined,
  minReplicas: number | undefined,
): number | undefined {
  const mergedMinReplicas = autoScaling?.minReplicas;
  // Note: there is a validation in the backend to check if the default minReplicas
  // is greater or equals to the minReplicas in the policy rule
  const policyRulesMinReplicas = policy?.rules?.autoscaling?.minReplicas?.min;
  if (mergedMinReplicas !== undefined && mergedMinReplicas !== null) return mergedMinReplicas;
  if (policyRulesMinReplicas !== undefined && policyRulesMinReplicas !== null) return policyRulesMinReplicas;
  return minReplicas;
}

function getAutoscalingMaxReplicasInitialValue(
  autoScaling: AutoScaling | null | undefined,
  policy: UnifiedPolicyInfoPerReplicaType | null | undefined,
  maxReplicas: number,
  minReplicas: number,
): number | undefined {
  const mergedMaxReplicas = autoScaling?.maxReplicas;
  // Note: there is a validation in the backend to check if the default maxReplicas
  // is greater or equals to the min in maxReplicas in the policy rule
  const policyRulesMaxReplicas = policy?.rules?.autoscaling?.maxReplicas?.min;
  if (mergedMaxReplicas !== undefined && mergedMaxReplicas !== null) return mergedMaxReplicas;
  if (policyRulesMaxReplicas !== undefined && policyRulesMaxReplicas !== null && policyRulesMaxReplicas > minReplicas) {
    return policyRulesMaxReplicas;
  }
  return Math.max(maxReplicas || 1, minReplicas);
}

function prepareInferenceFlatRequest(model: IUISuperWorkloadCreationRequest): InferenceCreationRequest {
  const specFields = _getFlatInferenceSpecData(model.spec);

  return {
    ...model,
    spec: specFields,
  };
}

function _getFlatInferenceSpecData(workloadSpec: IUISuperWorkloadSpec): InferenceSpecSpec {
  _validateServingPortRequiredFields(workloadSpec.servingPort);
  _validateAutoScalingRequiredFields(workloadSpec.autoscaling);

  const baseWorkloadSpec: SupersetSpec = workloadHybridUtil._getBaseFlatWorkloadSpecData(workloadSpec);

  const retWorkloadSpec: InferenceSpecSpec = {
    ...baseWorkloadSpec,
    ...pick(workloadSpec, "minReplicas", "maxReplicas", "servingPort", "autoscaling", "servingConfiguration"),
  } as InferenceSpecSpec;

  return retWorkloadSpec;
}

function _getDistributedLeaderWorkerSpecData(
  leaderWorkerSpec: IUISuperWorkloadSpec,
): DistributedInferenceLeaderSpecFieldsLeader | DistributedInferenceWorkerSpecFieldsWorker {
  const storage = workloadHybridUtil._getCleanedStorageFieldsFromUI(leaderWorkerSpec.storage || {});
  const compute = workloadHybridUtil._getCleanedComputeFieldsFromUI(leaderWorkerSpec.compute || {});
  const security = workloadHybridUtil._getCleanedSecurityFieldsFromUI(leaderWorkerSpec.security);

  return {
    ...omit(leaderWorkerSpec, ["connections"]),
    storage,
    compute,
    security,
  };
}

function _getDistributedSpecData(spec: IUISuperWorkloadSpec): Omit<DistributedInferenceSpecSpec, "leader" | "worker"> {
  const servingPort = spec.servingPort || { protocol: null };

  let authorizationType = servingPort.authorizationType;

  if (authorizationType === ServingPortAccessAuthorizationTypeEnum.AuthorizedUsersOrGroups) {
    if (servingPort.authorizedUsers) {
      authorizationType = ServingPortAccessAuthorizationTypeEnum.AuthorizedUsers;
    } else if (servingPort.authorizedGroups) {
      authorizationType = ServingPortAccessAuthorizationTypeEnum.AuthorizedGroups;
    } else {
      throw new Error(
        "Invalid servingPort configuration: authorizedUsers or authorizedGroups are required when authorizationType is AuthorizedUsersOrGroups.",
      );
    }
  }

  return {
    servingPort: {
      ...omit(servingPort, ["container", "clusterLocalAccessOnly"]),
      port: servingPort.container,
      protocol: servingPort.protocol as unknown as DistributedInferenceServingPortProtocol | null,
      authorizationType: authorizationType as unknown as DistributedInferenceServingPortAccessAuthorizationTypeEnum,
      authorizedUsers: servingPort.authorizedUsers || undefined,
      authorizedGroups: servingPort.authorizedGroups || undefined,
    },
    priorityClass: spec.priorityClass || null,
    preemptibility: spec.preemptibility || undefined,
    startupPolicy: spec.startupPolicy,
    nodePools: spec.nodePools ?? undefined,
  };
}

function prepareDistributedInferenceFlatRequest(
  model: IUISuperWorkloadCreationRequest,
  workers: number,
): DistributedInferenceCreationRequest {
  _validateServingPortRequiredFields(model.spec.servingPort);
  const specFields = _getDistributedSpecData(model.spec);
  const leaderSpecFields = _getDistributedLeaderWorkerSpecData(model.spec.leader || {});
  const workerSpecFields = _getDistributedLeaderWorkerSpecData(model.spec.worker || {});
  return {
    ...model,
    spec: {
      ...specFields,
      workers,
      replicas: model.spec.replicas ?? 1,
      leader: leaderSpecFields,
      worker: workerSpecFields,
    },
  };
}

function prepareInferenceTemplateFlatRequest(
  model: IUISuperWorkloadCreationRequest,
  linkedAssetsIds?: TemplateAssetsIds,
): InferenceTemplateCreationRequest {
  model.spec.storage = removeEmptyStorageArrays(model.spec.storage);
  return workloadHybridUtil.prepareSingleReplicaWorkloadTemplateFlatCreationRequest<InferenceTemplateCreationRequest>(
    model,
    linkedAssetsIds,
  );
}

function _validateServingPortRequiredFields(
  servingPort?: SupersetServingPort | null,
): asserts servingPort is SupersetServingPort {
  if (!servingPort) {
    throw new Error("Invalid servingPort configuration: servingPort is required but was not provided.");
  }

  if (!servingPort.protocol) {
    throw new Error("Invalid servingPort configuration: protocol is required but was not provided.");
  }

  if (!servingPort.container) {
    throw new Error("Invalid servingPort configuration: container is required but was not provided.");
  }
}

function _validateAutoScalingRequiredFields(autoscaling?: AutoScaling | null): asserts autoscaling is AutoScaling {
  if (!autoscaling) {
    throw new Error("Invalid autoscaling configuration: autoscaling is required but was not provided.");
  }

  if (!isNumber(autoscaling.minReplicas)) {
    throw new Error(
      `Invalid autoscaling configuration: minReplicas must be a number, but received ${autoscaling.minReplicas}.`,
    );
  }

  if (!isNumber(autoscaling.maxReplicas)) {
    throw new Error(
      `Invalid autoscaling configuration: maxReplicas must be a number, but received ${autoscaling.maxReplicas}.`,
    );
  }
}
