/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { DATA_SOURCE_ICONS_NAMES, distFrameworkIconsMap, toolTypeIconsMap } from "@/common/icons.constant";
import { getNvidiaImageBasedIcon } from "@/common/icon.util";
import {
  IMAGE_PULL_POLICY_LABELS,
  type IUIConnection,
  toolTypeOptions,
  UID_GID_SOURCE_LABELS,
} from "@/models/environment.model";
import {
  WorkloadCreationDrawerEntities,
  type IDrawerListItem,
  type IDrawerListItemIcon,
  type TItemDetail,
} from "@/models/workload-creation-drawer-model";
import type {
  AssetClusterStatusInfo,
  EmptyDirInstance,
  MpiLauncherCreationPolicy,
} from "@/swagger-models/assets-service-client";
import {
  AssetKind,
  DistributedFramework,
  InternalConnectionType,
  GpuRequestType,
  ToolType,
  type ExtendedResource,
  type ComplianceInfo,
  type ComplianceInfoReason,
  type ComputeAsset,
  type Connection,
  type DatasourceListResponseEntry,
  type EnvironmentAsset,
  type EnvironmentVariableOfAsset,
  type ComputeAssetSpec,
  ImagePullPolicy,
  type HostPathInstance,
  type PvcInstance,
  type NfsInstance,
  type GitInstance,
  type S3Instance,
  type ConfigMapInstance,
  type SecretInstance,
  type PVCAssetSpec,
  type PvcAccessModes,
  type S3AssetSpec,
  type ConfigMapAssetSpec,
  type NFSAssetSpec,
  type HostPathSpec,
  type GitAssetSpec,
  type SecretAssetSpec,
  type DataVolumeInstance,
  AssetSyncStatusEnum,
} from "@/swagger-models/assets-service-client";
import type {
  DistributedCreationRequest,
  DistributedTemplate,
  DistributedTemplateCreationRequest,
  DistributedTemplateUpdateRequest,
  ScopeType,
  SupersetSpec,
  TemplateAssetsIds,
  TemplateMasterSection,
  TrainingTemplate,
  TrainingTemplateCreationRequest,
  TrainingTemplateUpdateRequest,
  UnifiedSpecMaster,
  WorkspaceTemplate,
  WorkspaceTemplateCreationRequest,
  WorkspaceTemplateUpdateRequest,
  InferenceTemplateCreationRequest,
  InferenceTemplateUpdateRequest,
  WorkspaceSpecSpecAllOfStorage,
  ExposedUrl,
  Port,
  RelatedUrl,
  WorkspaceSpecSpecAllOfCompute,
  WorkspaceSpecSpecAllOfSecurity,
  UnifiedWorkload,
  WorkspaceCreationRequest,
  WorkspaceSpecSpec,
  TrainingCreationRequest,
  UnifiedPolicyInfoPerReplicaType,
  RestartPolicy,
  MasterSpec,
} from "@/swagger-models/workloads-client";
import { UidGidSource, AuthorizationType } from "@/swagger-models/workloads-client";
import { fallbackDefaultIfNullOrUndefined, makeId, omit } from "../common.util";
import { policyUtil } from "../policy.util/policy.util";
import {
  Preemptibility,
  type IUIConfigMapInstance,
  type IUIDataVolumeInstance,
  type IUIEmptyDirInstance,
  type IUIEnvironmentFields,
  type IUIGitInstance,
  type IUIHostPathInstance,
  type IUINfsInstance,
  type IUIPvcInstance,
  type IUIS3Instance,
  type IUISecretVolumeInstance,
  type IUIStorageFields,
  type IUISuperWorkloadCreationRequest,
  type IUISuperWorkloadSpec,
  type TUIStorageInstance,
} from "@/models/workload.model";
import { isNullOrUndefined } from "../common.util";
import { UI_EMPTY_DIR_KIND, UI_VOLUME_KIND } from "@/models/data-source.model";
import type { THybridSelectedTemplates } from "@/models/workload-hybrid.model";
import { isBoolean } from "../common.util/common.util";
import { computeResourceUtil } from "@/utils/compute-resource.util";
import { EAccessOptions } from "@/components/environment/connection-access/connection-access-modal/connection-access-modal.model";
import {
  MIN_CLUSTER_VERSION_FOR_COMPLETIONS_AND_PARALLELISM,
  MIN_CLUSTER_VERSION_FOR_PREEMPTIBILITY,
  MIN_CLUSTER_VERSION_FOR_SEMI_PREEMPTIBILITY,
  MIN_CLUSTER_VERSION_FOR_TERMINATION_GRACE_PERIOD,
} from "@/common/version.constant";
import { isNewerVersion } from "@/utils/version.util/version.util";
export const LINE_BREAK = "<br />";

export const workloadHybridUtil = {
  environmentListToDrawerItems,
  computeListToDrawerItems,
  dataSourceListToDrawerItems,
  getHybridWorkspaceModel,
  getHybridTrainingModel,
  getHybridDistributedModel,
  getHybridDistributedMasterModel,
  getHybridBaseWorkloadModel,
  setWorkspaceUIDefaults,
  setTrainingUIDefaults,
  setDistributedUIDefaults,
  getRelevantComputeFields,
  getRelevantStorageFields,
  getRelevantEnvironmentFields,
  getInferenceRelevantEnvironmentFields,
  getDistributedInferenceLeaderRelevantEnvironmentFields,
  getDistributedInferenceLeaderClearedEnvironmentFields,
  combineWorkloadModelWithComputeDetails,
  combineDistributedInferenceLeaderWorkloadModelWithComputeDetails,
  combineDistributedInferenceWorkerWorkloadModelWithComputeDetails,
  combineWorkloadModelWithStorageDetails,
  combineDistributedInferenceLeaderWorkloadModelWithStorageDetails,
  combineDistributedInferenceWorkerWorkloadModelWithStorageDetails,
  combineWorkloadModelWithEnvironmentDetails,
  combineDistributedInferenceLeaderWorkloadModelWithEnvironmentDetails,
  combineDistributedInferenceWorkerWorkloadModelWithEnvironmentDetails,
  combineMasterReplicaWorkloadModelWithEnvironmentDetails,
  combineMasterReplicaWorkloadModelWithComputeDetails,
  combineMasterReplicaWorkloadModelWithStorageDetails,
  prepareWorkloadFlatRequest,
  prepareTrainingFlatRequest,
  prepareDistributedFlatRequest,
  getWorkloadHybridTemplatesSetup,
  getMasterModelForUI,
  setDistributedMasterSpecUIDefaults,
  getEmptyUISuperWorkloadModel,
  getBaseEnvironmentFields,
  getConnectionsField,
  getComputeField,
  getStorageField,
  getSecurityField,
  calculateTotalGpuRequest,
  setBaseWorkloadUIDefaults,

  // templates
  getHybridWorkspaceTemplateModel,
  getHybridTrainingTemplateModel,
  getHybridDistributedTemplateModel,
  prepareWorkspaceTemplateFlatRequest,
  prepareWorkspaceTemplateFlatRequestToUpdate,
  prepareTrainingTemplateFlatRequestToUpdate,
  prepareTrainingTemplateFlatRequest,
  prepareDistributedTemplateFlatRequestToUpdate,
  prepareDistributedTemplateFlatRequest,
  getFirstPageNavigationTemproraryData,
  prepareSingleReplicaWorkloadTemplateFlatCreationRequest,
  prepareSingleReplicaWorkloadTemplateFlatUpdateRequest,
  processWorkloadStorage,

  // for unit tests only
  _getBaseFlatWorkloadSpecData,
  _getCleanedStorageFieldsFromUI,
  _getCleanedComputeFieldsFromUI,
  _getCleanedSecurityFieldsFromUI,
  _getCleanedConnectionsFieldsFromUI,
};

function computeListToDrawerItems(computes: ComputeAsset[]): IDrawerListItem[] {
  return computes.map((compute: ComputeAsset) => ({
    name: compute.meta.name,
    id: compute.meta.id,
    kind: compute.meta.kind,
    description: compute.meta.description || "",
    icons: [{ name: "default", icon: "ph-bold ph-hard-drives" }],
    policy: compute.compliance,
    createdAt: compute.meta.createdAt,
    lastUsed: compute.usageTimes?.lastUsedByWorkload || "",
    disabled: _isDisabled(compute.compliance),
    details: _computeResourceRawToDetails(compute),
  }));
}

function dataSourceListToDrawerItems(
  dataSources: DatasourceListResponseEntry[],
  createdEntityId?: string,
): IDrawerListItem[] {
  return dataSources.map((ds: DatasourceListResponseEntry) => ({
    name: ds.meta.name,
    id: ds.meta.id,
    kind: ds.meta.kind,
    description: ds.meta.description || "",
    icons: _getDataSourceIcon(ds),
    policy: ds.compliance,
    createdAt: ds.meta.createdAt,
    lastUsed: ds.usageTimes?.lastUsedByWorkload || "",
    disabled: _isDisabled(ds.compliance, ds.status),
    disabledInfo: _getDisabledInfo(ds.compliance, ds.status, ds.meta.id, createdEntityId),
    details: _dataSourceRawToDetails(ds),
  }));
}

function environmentListToDrawerItems(environments: EnvironmentAsset[]): IDrawerListItem[] {
  return environments.map((env: EnvironmentAsset) => ({
    name: env.meta.name,
    id: env.meta.id,
    kind: env.meta.kind,
    description: env.meta.description || "",
    icons: _getEnvironmentIcons(env),
    policy: env.compliance,
    createdAt: env.meta.createdAt,
    lastUsed: env.usageTimes?.lastUsedByWorkload || "",
    disabled: _isDisabled(env.compliance),
    details: _environmentRawToDetails(env),
  }));
}

function _getEnvironmentIcons(env: EnvironmentAsset): IDrawerListItemIcon[] {
  const defaultIcon: IDrawerListItemIcon = { name: "default", icon: "environment-gray-new" };

  // Check for image-based icons first (NVIDIA PyTorch, NVIDIA general, etc.)
  const image = env.spec.image;
  const imageBasedIcon = getNvidiaImageBasedIcon(image);
  if (imageBasedIcon) {
    return [{ name: imageBasedIcon, icon: imageBasedIcon }];
  }

  if (!env.spec.connections?.length && !env.meta.workloadSupportedTypes?.distFramework) {
    return [defaultIcon];
  }

  const icons: Array<IDrawerListItemIcon> = [];

  const { distFramework } = env.meta.workloadSupportedTypes || {};
  if (distFramework) {
    // DistFramwork icon should be the first icon.
    icons.push({ name: distFramework, icon: distFrameworkIconsMap[distFramework] });
  }

  env.spec.connections?.forEach((connection: Connection) => {
    const toolType: string | undefined = connection.internalToolInfo?.toolType || connection.externalToolInfo?.toolType;
    icons.push({
      name: connection.name,
      icon: (toolType && toolTypeIconsMap[toolType]) || defaultIcon.icon,
    });
  });

  return icons;
}

function _getDataSourceIcon(ds: DatasourceListResponseEntry): IDrawerListItemIcon[] {
  return [
    {
      name: ds.meta.kind || "",
      icon: (DATA_SOURCE_ICONS_NAMES[ds.meta.kind as keyof object] as string) || DATA_SOURCE_ICONS_NAMES.default,
    },
  ];
}

function _isDisabled(
  compliance: ComplianceInfo | null | undefined,
  status?: AssetClusterStatusInfo | null | undefined,
): boolean {
  if (compliance?.compliance) return false;

  // if status is provided we need to check it.
  if (status?.status === AssetSyncStatusEnum.IssuesFound || status?.status === AssetSyncStatusEnum.Failed) return true;

  if (!compliance) return false;
  return (
    (compliance.reason || []).filter((reason) => {
      return policyUtil.isNonManageableRuleConflict(reason);
    }).length > 0
  );
}

function _getDisabledInfo(
  compliance: ComplianceInfo | null | undefined,
  status: AssetClusterStatusInfo | null | undefined,
  dataSourceId: string,
  createdEntityId?: string,
): string {
  // Check compliance first (higher priority)
  if (compliance && !compliance?.compliance) {
    // if it does not comply due to policy vs if it does not comply for some other reason
    if (compliance?.reason?.some((reason: ComplianceInfoReason) => reason.field)) {
      return "This setup can't be selected based on the policy your administrator set.";
    } else if (compliance?.reason?.length) {
      // Check if this is a newly created data source with issues
      if (dataSourceId === createdEntityId) {
        return "This setup is being created. If it isn't created after a few seconds, contact your administrator.";
      }
      return compliance.reason[0].details;
    }
  }

  // Check status issues
  if (status?.status === AssetSyncStatusEnum.IssuesFound || status?.status === AssetSyncStatusEnum.Failed) {
    return status.message || "This setup can't be selected because it has issues. Contact your administrator.";
  }

  return "";
}

function _hasValue<T>(value: T): boolean {
  return !isNullOrUndefined(value);
}

function _environmentRawToDetails(env: EnvironmentAsset): TItemDetail[] {
  const itemDetails = [];
  env.meta.description && itemDetails.push({ title: "Description", content: [env.meta.description] });
  env.spec.image && itemDetails.push({ title: "Image URL", content: [env.spec.image] });
  env.spec.imagePullPolicy &&
    itemDetails.push({
      title: "Image pull policy",
      content: [IMAGE_PULL_POLICY_LABELS[env.spec.imagePullPolicy] || ""],
    });
  env.meta.workloadSupportedTypes?.distFramework &&
    itemDetails.push(_parseFrameworkDetails(env.meta.workloadSupportedTypes?.distFramework));
  env.spec.connections && itemDetails.push(..._parseConnectionsToDetails(env.spec.connections || []));

  (env.spec.command || env.spec.args) &&
    itemDetails.push({
      title: "Command and arguments",
      content: _parseCommandAndArgsToDetails(env.spec.command || "", env.spec.args || ""),
    });
  env.spec.environmentVariables?.length &&
    itemDetails.push({
      title: "Environment variables",
      content: _parseEnvVarsToDetailsContent(env.spec.environmentVariables || []) || "",
    });
  env.spec.workingDir && itemDetails.push({ title: "Container's working directory ", content: [env.spec.workingDir] });
  env.spec.uidGidSource &&
    itemDetails.push(
      _parseUidGidSource(env.spec.uidGidSource, env.spec.runAsUid, env.spec.runAsGid, env.spec.supplementalGroups),
    );
  env.spec?.capabilities?.length &&
    itemDetails.push({
      title: "Additional Linux capabilities for the container",
      content: [env.spec.capabilities?.join(", ")],
    });
  return itemDetails;
}

function _parseCommandAndArgsToDetails(command: string, args: string) {
  const output: string[] = [];
  if (command) output.push(`Command: ${command}`);
  if (args) output.push(`Arguments: ${args}`);
  return output;
}

function _parseEnvVarsToDetailsContent(envVars: EnvironmentVariableOfAsset[]) {
  const output: string[] = [];
  return envVars.reduce((acc, envVar) => {
    const { name, value, credential } = envVar;
    acc.push(`Name: ${name}`);
    if (!credential) {
      acc.push(`Source: Custom`);
      acc.push(`Value: ${value}`);
    } else {
      acc.push(`Source: Credential`);
    }
    acc.push(LINE_BREAK);
    return acc;
  }, output);
}

function _parseConnectionsToDetails(connections: Connection[]): TItemDetail[] {
  return connections.map((c: Connection, index: number) => {
    const detail: TItemDetail = {
      title: c.name,
      icon: _getConnectionIcon(c),
      noSeperator: index !== connections.length - 1,
      content: _connectionContent(c),
    };
    return detail;
  });
}

function _connectionContent(c: Connection): string[] {
  const output: string[] = [];

  if (c.isExternal) {
    c.externalToolInfo?.externalUrl && output.push(`External URL: ${c.externalToolInfo?.externalUrl || ""}`);
  } else {
    c.internalToolInfo?.connectionType && output.push(`Connection type: ${c.internalToolInfo?.connectionType || ""}`);
    c.internalToolInfo?.containerPort && output.push(`Container port: ${c.internalToolInfo?.containerPort || ""}`);
    c.internalToolInfo?.externalUrlInfo?.externalUrl &&
      output.push(`Custom URL: ${c.internalToolInfo?.externalUrlInfo?.externalUrl || ""}`);
    c.internalToolInfo?.servingPortInfo?.protocol &&
      output.push(`Protocol: ${c.internalToolInfo?.servingPortInfo?.protocol || ""}`);
  }

  return output;
}

function _getConnectionIcon(c: Connection): string {
  const toolType = c.isExternal ? c.externalToolInfo?.toolType : c.internalToolInfo?.toolType;
  return toolTypeOptions.find((t) => t.value === toolType)?.icon || "default";
}

function _parseFrameworkDetails(distFramework: DistributedFramework): TItemDetail {
  return {
    title: `${distFramework} framework`,
    icon: distFrameworkIconsMap[distFramework] || "",
    content: [],
  };
}

function _parseUidGidSource(
  uidGidSource: UidGidSource,
  uid?: number | null,
  gid?: number | null,
  suppGroups?: string | null,
): TItemDetail {
  const content = [UID_GID_SOURCE_LABELS[uidGidSource]];

  if (uidGidSource === UidGidSource.Custom) {
    uid && content.push(`User ID (UID): ${uid || ""}`);
    gid && content.push(`Group ID (GID): ${gid || ""}`);
    suppGroups && content.push(`Supplementary groups: ${suppGroups}`);
  }

  return {
    title: "UID, GID, and supplementary groups",
    content,
  };
}

function _computeResourceRawToDetails(compute: ComputeAsset): TItemDetail[] {
  const itemDetails = [];

  _hasValue(compute.meta.description) &&
    itemDetails.push({ title: "Description", content: [compute.meta.description || ""] });

  _hasValue(compute.spec.gpuDevicesRequest) &&
    itemDetails.push({
      title: "GPU devices per pod",
      content: [`GPU devices: ${compute.spec.gpuDevicesRequest}`],
    });

  const gpuMemoryContent = _parseGpuMemory(compute.spec);
  if (gpuMemoryContent.length)
    itemDetails.push({
      title: "GPU memory per device",
      content: gpuMemoryContent,
    });

  const cpuCoreContent = _parseCpuCoreRequest(compute.spec);
  if (cpuCoreContent.length) {
    itemDetails.push({
      title: "CPU compute per pod",
      content: cpuCoreContent,
    });
  }

  const cpuMemoryContent = _parseCpuMemory(compute.spec);
  if (cpuMemoryContent.length) {
    itemDetails.push({
      title: "CPU memory per pod",
      content: cpuMemoryContent,
    });
  }

  compute.spec.largeShmRequest &&
    itemDetails.push({
      title: "Increase shared memory size",
      content: ["Enabled"],
    });

  compute.spec?.extendedResources?.length &&
    itemDetails.push({
      title: "Extended resource(s)",
      content: _parseExtendedResourcesToDetailsContent(compute.spec.extendedResources),
    });

  return itemDetails;
}

function _parseGpuMemory(computeSpec: ComputeAssetSpec): string[] {
  const { gpuRequestType, gpuMemoryRequest, gpuMemoryLimit, gpuPortionRequest, gpuPortionLimit } = computeSpec;

  const output: string[] = [];
  if (gpuRequestType === GpuRequestType.Portion) {
    _hasValue(gpuPortionRequest) && output.push(`Request: ${(gpuPortionRequest ?? 0) * 100}%`);
    _hasValue(gpuPortionLimit) && output.push(`Limit: ${(gpuPortionLimit ?? 0) * 100}%`);
  } else if (gpuRequestType === GpuRequestType.Memory) {
    gpuMemoryRequest && output.push(`Request: ${_parseMemoryValue(gpuMemoryRequest)}`);
    gpuMemoryLimit && output.push(`Limit: ${_parseMemoryValue(gpuMemoryLimit)}`);
  }
  return output;
}

function _parseCpuCoreRequest(computeSpec: ComputeAssetSpec): string[] {
  const output: string[] = [];
  const { cpuCoreRequest, cpuCoreLimit } = computeSpec;
  _hasValue(cpuCoreRequest) && output.push(`Request: ${cpuCoreRequest} Cores`);
  _hasValue(cpuCoreLimit) && output.push(`Limit: ${cpuCoreLimit} Cores`);
  return output;
}

function _parseCpuMemory(computeSpec: ComputeAssetSpec): string[] {
  const { cpuMemoryRequest, cpuMemoryLimit } = computeSpec;
  const output: string[] = [];

  cpuMemoryRequest && output.push(`Request: ${_parseMemoryValue(cpuMemoryRequest)}`);
  cpuMemoryLimit && output.push(`Limit: ${_parseMemoryValue(cpuMemoryLimit)}`);

  return output;
}

function _parseExtendedResourcesToDetailsContent(extendedResources: ExtendedResource[]): string[] {
  const output: string[] = [];
  return extendedResources.reduce((acc, eRes) => {
    const { quantity, resource } = eRes;
    acc.push(`Resource: ${resource}`);
    acc.push(`Quantity: ${quantity}`);
    acc.push(LINE_BREAK);
    return acc;
  }, output);
}

function _parseMemoryValue(memoryValue: string): string {
  const unit = memoryValue?.charAt(memoryValue.length - 1);
  const amount = memoryValue.slice(0, memoryValue.length - 1);
  switch (unit) {
    case "M":
      return `${amount} MB`;
    case "G":
      return `${amount} GB`;
    default:
      return memoryValue;
  }
}

function _dataSourceRawToDetails(ds: DatasourceListResponseEntry): TItemDetail[] {
  const dsType = ds?.meta?.kind;

  switch (dsType) {
    case AssetKind.Pvc:
      return _pvcToDetails(ds);
    case AssetKind.Nfs:
      return _nfsToDetails(ds);
    case AssetKind.HostPath:
      return _hostPathToDetails(ds);
    case AssetKind.Git:
      return _gitToDetails(ds);
    case AssetKind.S3:
      return _s3ToDetails(ds);
    case AssetKind.ConfigMap:
      return _configMapToDetails(ds);
    case AssetKind.SecretVolume:
      return _secretVolumeToDetails(ds);
    case AssetKind.DataVolume:
      return _dataVolumeToDetails(ds);
    default:
      return [];
  }
}

function processWorkloadStorage(storage: WorkspaceSpecSpecAllOfStorage): TItemDetail[] {
  const allDetails: TItemDetail[] = [];

  const storageHandlers = {
    hostPath: _hostPathInstanceToDetails,
    nfs: _nfsInstanceToDetails,
    pvc: _pvcInstanceToDetails,
    s3: _s3InstanceToDetails,
    git: _gitInstanceToDetails,
    configMapVolume: _configMapInstanceToDetails,
    secretVolume: _secretVolumeInstanceToDetails,
    dataVolume: _dataVolumeInstanceToDetails,
  };

  (Object.keys(storageHandlers) as Array<keyof typeof storageHandlers>).forEach((key) => {
    const items = storage[key];
    if (items?.length) {
      items.forEach((item) => allDetails.push(...storageHandlers[key](item as never)));
    }
  });

  return allDetails;
}

const readOnlyTitle = "Prevent data modification (the data will be mounted with read-only permissions)";

// Helper for processing DatasourceListResponseEntry (with meta)
function _buildAssetDetails(
  meta: { description?: string | null },
  spec: Record<string, unknown>,
  fields: Array<{ label: string; key: string; formatter?: (val: unknown) => string[] | string | null }>,
): TItemDetail[] {
  const itemDetails: TItemDetail[] = [];

  meta.description && itemDetails.push({ title: "Description", content: [meta.description] });

  fields.forEach(({ label, key, formatter }) => {
    const value = spec[key];
    if (value !== undefined && value !== null) {
      const content = formatter ? formatter(value) : [String(value)];
      if (content !== null) {
        itemDetails.push({ title: label, content: Array.isArray(content) ? content : [content] });
      }
    }
  });

  return itemDetails;
}

// Helper for processing Instance types (no meta)
function _buildInstanceDetails<T>(
  instance: T,
  fields: Array<{ label: string; getValue: (obj: T) => unknown; formatter?: (val: unknown) => string }>,
): TItemDetail[] {
  const content: string[] = [];

  fields.forEach(({ label, getValue, formatter }) => {
    const value = getValue(instance);
    if (value !== undefined && value !== null) {
      const formatted = formatter ? formatter(value) : String(value);
      content.push(`${label}: ${formatted}`);
    }
  });

  return content.length ? [{ content }] : [];
}

function _pvcToDetails(ds: DatasourceListResponseEntry): TItemDetail[] {
  const pvcSpec = ds.spec.pvc as PVCAssetSpec;

  // Flatten claimInfo for easier processing
  const flattenedSpec = {
    ...pvcSpec,
    storageClass: pvcSpec.claimInfo?.storageClass,
    accessModes: pvcSpec.claimInfo?.accessModes,
    size: pvcSpec.claimInfo?.size,
    volumeMode: pvcSpec.claimInfo?.volumeMode,
  };

  return _buildAssetDetails(ds.meta, flattenedSpec as Record<string, unknown>, [
    { label: "Storage class", key: "storageClass" },
    { label: "Access modes", key: "accessModes", formatter: (v) => _pvcAccessModesValue(v as PvcAccessModes) },
    { label: "Claim size", key: "size", formatter: (v) => [String(v) + "B"] },
    { label: "Volume mode", key: "volumeMode" },
    { label: "Container path", key: "path" },
    { label: readOnlyTitle, key: "readOnly", formatter: (v) => (v ? ["Enabled"] : null) },
  ]);
}

function _pvcInstanceToDetails(pvc: PvcInstance): TItemDetail[] {
  return _buildInstanceDetails(pvc, [
    { label: "Type", getValue: () => "PVC" },
    { label: "Storage class", getValue: (pvc) => pvc.claimInfo?.storageClass },
    {
      label: "Access modes",
      getValue: (pvc) => pvc.claimInfo?.accessModes,
      formatter: (v) => (v ? _pvcAccessModesValue(v as PvcAccessModes).join(", ") : ""),
    },
    { label: "Claim size", getValue: (pvc) => pvc.claimInfo?.size, formatter: (v) => String(v) + "B" },
    { label: "Volume mode", getValue: (pvc) => pvc.claimInfo?.volumeMode },
    { label: "Claim name", getValue: (pvc) => pvc.claimName },
    { label: "Container path", getValue: (pvc) => pvc.path },
  ]);
}

function _dataVolumeToDetails(ds: DatasourceListResponseEntry): TItemDetail[] {
  return _buildAssetDetails(ds.meta, {}, []);
}

function _dataVolumeInstanceToDetails(dv: DataVolumeInstance): TItemDetail[] {
  return _buildInstanceDetails(dv, [
    { label: "Type", getValue: () => "Data Volume" },
    { label: "Data Volume ID", getValue: (dv) => dv.id },
    { label: "Container path", getValue: (dv) => dv.mountPath },
  ]);
}

function _pvcAccessModesValue(accessModes: PvcAccessModes): string[] {
  const mapper = {
    readWriteOnce: "Read-write by one node",
    readOnlyMany: "Read-only by many nodes",
    readWriteMany: "Read-write by many nodes",
  };

  const activeAccessModes = [];
  accessModes.readWriteOnce && activeAccessModes.push(mapper.readWriteOnce);
  accessModes.readOnlyMany && activeAccessModes.push(mapper.readOnlyMany);
  accessModes.readWriteMany && activeAccessModes.push(mapper.readWriteMany);

  return [activeAccessModes.join(", ")];
}

function _nfsToDetails(ds: DatasourceListResponseEntry): TItemDetail[] {
  const nfsSpec = ds.spec.nfs as NFSAssetSpec;
  return _buildAssetDetails(ds.meta, nfsSpec as Record<string, unknown>, [
    { label: "NFS server (host name or host IP)", key: "server" },
    { label: "NFS path", key: "path" },
    { label: "Container path", key: "mountPath" },
    { label: readOnlyTitle, key: "readOnly", formatter: (v) => (v ? ["Enabled"] : null) },
  ]);
}

function _nfsInstanceToDetails(nfs: NfsInstance): TItemDetail[] {
  return _buildInstanceDetails(nfs, [
    { label: "Type", getValue: () => "NFS" },
    { label: "NFS server (host name or host IP)", getValue: (nfs) => nfs.server },
    { label: "NFS path", getValue: (nfs) => nfs.path },
    { label: "Container path", getValue: (nfs) => nfs.mountPath },
    { label: "Read only", getValue: (nfs) => nfs.readOnly, formatter: (v) => (v ? "Enabled" : "") },
  ]);
}

function _hostPathToDetails(ds: DatasourceListResponseEntry): TItemDetail[] {
  const hostPathSpec = ds.spec.hostPath as HostPathSpec;
  return _buildAssetDetails(ds.meta, hostPathSpec as Record<string, unknown>, [
    { label: "Host path", key: "path" },
    { label: "Container path", key: "mountPath" },
    { label: readOnlyTitle, key: "readOnly", formatter: (v) => (v ? ["Enabled"] : null) },
  ]);
}

function _hostPathInstanceToDetails(hp: HostPathInstance): TItemDetail[] {
  return _buildInstanceDetails(hp, [
    { label: "Type", getValue: () => "Host Path" },
    { label: "Host path", getValue: (hp) => hp.path },
    { label: "Container path", getValue: (hp) => hp.mountPath },
    { label: "Read only", getValue: (hp) => hp.readOnly, formatter: (v) => (v ? "Enabled" : "") },
  ]);
}

function _gitToDetails(ds: DatasourceListResponseEntry): TItemDetail[] {
  const gitSpec = ds.spec.git as GitAssetSpec;
  return _buildAssetDetails(ds.meta, gitSpec as Record<string, unknown>, [
    { label: "Repository URL", key: "repository" },
    { label: "Revision", key: "revision" },
    { label: "Branch", key: "branch" },
    { label: "Container path", key: "path" },
  ]);
}

function _gitInstanceToDetails(git: GitInstance): TItemDetail[] {
  return _buildInstanceDetails(git, [
    { label: "Type", getValue: () => "Git" },
    { label: "Repository URL", getValue: (git) => git.repository },
    { label: "Revision", getValue: (git) => git.revision },
    { label: "Branch", getValue: (git) => git.branch },
    { label: "Container path", getValue: (git) => git.path },
  ]);
}

function _s3ToDetails(ds: DatasourceListResponseEntry): TItemDetail[] {
  const s3Spec = ds.spec.s3 as S3AssetSpec;
  return _buildAssetDetails(ds.meta, s3Spec as Record<string, unknown>, [
    { label: "S3 service URL", key: "url" },
    { label: "Bucket name", key: "bucket" },
    { label: "Container path", key: "path" },
  ]);
}

function _s3InstanceToDetails(s3: S3Instance): TItemDetail[] {
  return _buildInstanceDetails(s3, [
    { label: "Type", getValue: () => "S3" },
    { label: "S3 service URL", getValue: (s3) => s3.url },
    { label: "Bucket name", getValue: (s3) => s3.bucket },
    { label: "Container path", getValue: (s3) => s3.path },
  ]);
}

function _configMapToDetails(ds: DatasourceListResponseEntry): TItemDetail[] {
  const configMapSpec = ds.spec.config_map as ConfigMapAssetSpec;
  return _buildAssetDetails(ds.meta, configMapSpec as Record<string, unknown>, [
    { label: "ConfigMap name", key: "configMap" },
    { label: "Container path", key: "mountPath" },
  ]);
}

function _configMapInstanceToDetails(cm: ConfigMapInstance): TItemDetail[] {
  return _buildInstanceDetails(cm, [
    { label: "Type", getValue: () => "Config Map" },
    { label: "ConfigMap name", getValue: (cm) => cm.configMap },
    { label: "Container path", getValue: (cm) => cm.mountPath },
  ]);
}

function _secretVolumeToDetails(ds: DatasourceListResponseEntry): TItemDetail[] {
  const secretSpec = ds.spec.secret as SecretAssetSpec;
  return _buildAssetDetails(ds.meta, secretSpec as Record<string, unknown>, [
    { label: "Container path", key: "mountPath" },
  ]);
}

function _secretVolumeInstanceToDetails(secret: SecretInstance): TItemDetail[] {
  return _buildInstanceDetails(secret, [
    { label: "Type", getValue: () => "Secret" },
    { label: "Secret name", getValue: (secret) => secret.secret },
    { label: "Container path", getValue: (secret) => secret.mountPath },
  ]);
}

function getHybridBaseWorkloadModel(unifiedWorkload: UnifiedWorkload): IUISuperWorkloadCreationRequest {
  return {
    name: unifiedWorkload.meta?.name || "",
    projectId: unifiedWorkload.meta?.projectId || "",
    clusterId: unifiedWorkload.meta?.clusterId || "",
    spec: {
      command: unifiedWorkload.spec?.worker?.command,
      args: unifiedWorkload.spec?.worker?.args,
      image: unifiedWorkload.spec?.worker?.image,
      imagePullPolicy: unifiedWorkload.spec?.worker?.imagePullPolicy,
      imagePullSecrets: unifiedWorkload.spec?.worker?.imagePullSecrets,
      workingDir: unifiedWorkload.spec?.worker?.workingDir,
      createHomeDir: unifiedWorkload.spec?.worker?.createHomeDir,
      nodeType: unifiedWorkload.spec?.worker?.nodeType,
      nodePools: unifiedWorkload.spec?.worker?.nodePools,
      podAffinity: unifiedWorkload.spec?.worker?.podAffinity,
      tty: unifiedWorkload.spec?.worker?.tty,
      stdin: unifiedWorkload.spec?.worker?.stdin,
      environmentVariables: unifiedWorkload.spec?.worker?.environmentVariables,
      annotations: unifiedWorkload.spec?.worker?.annotations,
      labels: unifiedWorkload.spec?.worker?.labels,
      tolerations: unifiedWorkload.spec?.worker?.tolerations,
      connections: getConnectionsField(
        unifiedWorkload.spec?.worker?.ports || [],
        unifiedWorkload.spec?.worker?.exposedUrls || [],
        unifiedWorkload.spec?.worker?.relatedUrls || [],
      ),
      compute: getComputeField(unifiedWorkload.spec?.worker?.compute),
      storage: getStorageField(unifiedWorkload.spec?.worker?.storage),
      security: getSecurityField(unifiedWorkload.spec?.worker?.security),
    },
  };
}

function getHybridWorkspaceModel(unifiedWorkload: UnifiedWorkload): IUISuperWorkloadCreationRequest {
  const workload = getHybridBaseWorkloadModel(unifiedWorkload);
  workload.spec.backoffLimit = unifiedWorkload.spec?.worker?.backoffLimit;
  workload.spec.priorityClass = unifiedWorkload.spec?.worker?.priorityClass;
  workload.spec.preemptibility = unifiedWorkload.spec?.worker?.preemptibility;
  workload.spec.terminateAfterPreemption = unifiedWorkload.spec?.worker?.terminateAfterPreemption;
  workload.spec.autoDeletionTimeAfterCompletionSeconds =
    unifiedWorkload.spec?.worker?.autoDeletionTimeAfterCompletionSeconds;
  workload.spec.terminationGracePeriodSeconds = unifiedWorkload.spec?.worker?.terminationGracePeriodSeconds;
  return workload;
}

function getHybridTrainingModel(unifiedWorkload: UnifiedWorkload): IUISuperWorkloadCreationRequest {
  return {
    name: unifiedWorkload.meta?.name || "",
    projectId: unifiedWorkload.meta?.projectId || "",
    clusterId: unifiedWorkload.meta?.clusterId || "",
    spec: {
      command: unifiedWorkload.spec?.worker?.command,
      args: unifiedWorkload.spec?.worker?.args,
      image: unifiedWorkload.spec?.worker?.image,
      imagePullPolicy: unifiedWorkload.spec?.worker?.imagePullPolicy,
      workingDir: unifiedWorkload.spec?.worker?.workingDir,
      createHomeDir: unifiedWorkload.spec?.worker?.createHomeDir,
      nodeType: unifiedWorkload.spec?.worker?.nodeType,
      nodePools: unifiedWorkload.spec?.worker?.nodePools,
      podAffinity: unifiedWorkload.spec?.worker?.podAffinity,
      tty: unifiedWorkload.spec?.worker?.tty,
      stdin: unifiedWorkload.spec?.worker?.stdin,
      environmentVariables: unifiedWorkload.spec?.worker?.environmentVariables,
      annotations: unifiedWorkload.spec?.worker?.annotations,
      labels: unifiedWorkload.spec?.worker?.labels,
      tolerations: unifiedWorkload.spec?.worker?.tolerations,
      parallelism: unifiedWorkload.spec?.worker?.parallelism,
      completions: unifiedWorkload.spec?.worker?.completions,
      terminateAfterPreemption: unifiedWorkload.spec?.worker?.terminateAfterPreemption,
      autoDeletionTimeAfterCompletionSeconds: unifiedWorkload.spec?.worker?.autoDeletionTimeAfterCompletionSeconds,
      backoffLimit: unifiedWorkload.spec?.worker?.backoffLimit,
      connections: getConnectionsField(
        unifiedWorkload.spec?.worker?.ports || [],
        unifiedWorkload.spec?.worker?.exposedUrls || [],
        unifiedWorkload.spec?.worker?.relatedUrls || [],
      ),
      terminationGracePeriodSeconds: unifiedWorkload.spec?.worker?.terminationGracePeriodSeconds,
      priorityClass: unifiedWorkload.spec?.worker?.priorityClass,
      preemptibility: unifiedWorkload.spec?.worker?.preemptibility,
      compute: getComputeField(unifiedWorkload.spec?.worker?.compute),
      storage: getStorageField(unifiedWorkload.spec?.worker?.storage),
      security: getSecurityField(unifiedWorkload.spec?.worker?.security),
    },
  };
}
function getHybridDistributedModel(
  unifiedWorkload: UnifiedWorkload,
  noMaster = false,
  basedOnExistingWorkload = false,
): IUISuperWorkloadCreationRequest {
  const uiDist: IUISuperWorkloadCreationRequest = {
    name: unifiedWorkload.meta?.name || "",
    projectId: unifiedWorkload.meta?.projectId || "",
    clusterId: unifiedWorkload.meta?.clusterId || "",
    masterSpecSameAsWorker: unifiedWorkload.spec?.masterSpecSameAsWorker,
    spec: {
      command: unifiedWorkload.spec?.worker?.command,
      args: unifiedWorkload.spec?.worker?.args,
      image: unifiedWorkload.spec?.worker?.image,
      imagePullPolicy: unifiedWorkload.spec?.worker?.imagePullPolicy,
      workingDir: unifiedWorkload.spec?.worker?.workingDir,
      createHomeDir: unifiedWorkload.spec?.worker?.createHomeDir,
      nodeType: unifiedWorkload.spec?.worker?.nodeType,
      nodePools: unifiedWorkload.spec?.worker?.nodePools,
      podAffinity: unifiedWorkload.spec?.worker?.podAffinity,
      tty: unifiedWorkload.spec?.worker?.tty,
      stdin: unifiedWorkload.spec?.worker?.stdin,
      environmentVariables: unifiedWorkload.spec?.worker?.environmentVariables,
      annotations: unifiedWorkload.spec?.worker?.annotations,
      labels: unifiedWorkload.spec?.worker?.labels,
      tolerations: unifiedWorkload.spec?.worker?.tolerations,
      terminateAfterPreemption: unifiedWorkload.spec?.worker?.terminateAfterPreemption,
      priorityClass: unifiedWorkload.spec?.worker?.priorityClass,
      preemptibility: unifiedWorkload.spec?.worker?.preemptibility,
      autoDeletionTimeAfterCompletionSeconds: unifiedWorkload.spec?.worker?.autoDeletionTimeAfterCompletionSeconds,
      backoffLimit: unifiedWorkload.spec?.worker?.backoffLimit,
      numWorkers: unifiedWorkload.spec?.worker?.numWorkers,
      segmentSize: unifiedWorkload.spec?.worker?.segmentSize,
      cleanPodPolicy: unifiedWorkload.spec?.worker?.cleanPodPolicy,
      connections: getConnectionsField(
        unifiedWorkload.spec?.worker?.ports || [],
        unifiedWorkload.spec?.worker?.exposedUrls || [],
        unifiedWorkload.spec?.worker?.relatedUrls || [],
      ),
      terminationGracePeriodSeconds: unifiedWorkload.spec?.worker?.terminationGracePeriodSeconds,
      compute: getComputeField(unifiedWorkload.spec?.worker?.compute),
      storage: getStorageField(unifiedWorkload.spec?.worker?.storage),
      security: getSecurityField(unifiedWorkload.spec?.worker?.security),
      //  careful here
      distributedFramework: unifiedWorkload.spec?.worker?.distributedFramework,
      sshAuthMountPath: unifiedWorkload.spec?.worker?.sshAuthMountPath,
      mpiLauncherCreationPolicy: unifiedWorkload.spec?.worker?.mpiLauncherCreationPolicy,
      minReplicas: unifiedWorkload.spec?.worker?.minReplicas,
      maxReplicas: unifiedWorkload.spec?.worker?.maxReplicas,
    },
  };

  if (noMaster) return uiDist;
  if (basedOnExistingWorkload && !unifiedWorkload.spec?.master) return uiDist;

  return {
    ...uiDist,
    masterSpec: getMasterModelForUI(
      unifiedWorkload.spec?.master || {},
      unifiedWorkload.spec?.worker?.sshAuthMountPath || undefined,
      unifiedWorkload.spec?.worker?.mpiLauncherCreationPolicy || undefined,
    ),
  };
}

function getMasterModelForUI(
  unifiedMaster: UnifiedSpecMaster,
  sshAuthMountPath?: string,
  mpiLauncherCreationPolicy?: MpiLauncherCreationPolicy | null,
): IUISuperWorkloadSpec {
  return {
    command: unifiedMaster?.command,
    args: unifiedMaster?.args,
    image: unifiedMaster?.image,
    imagePullPolicy: unifiedMaster?.imagePullPolicy,
    workingDir: unifiedMaster?.workingDir,
    createHomeDir: unifiedMaster?.createHomeDir,
    nodeType: unifiedMaster?.nodeType,
    nodePools: unifiedMaster?.nodePools,
    tty: unifiedMaster?.tty,
    stdin: unifiedMaster?.stdin,
    environmentVariables: unifiedMaster?.environmentVariables,
    annotations: unifiedMaster?.annotations,
    labels: unifiedMaster?.labels,
    tolerations: unifiedMaster?.tolerations,
    terminateAfterPreemption: unifiedMaster?.terminateAfterPreemption,
    priorityClass: unifiedMaster?.priorityClass,
    preemptibility: unifiedMaster?.preemptibility,
    autoDeletionTimeAfterCompletionSeconds: unifiedMaster?.autoDeletionTimeAfterCompletionSeconds,
    backoffLimit: unifiedMaster?.backoffLimit,
    terminationGracePeriodSeconds: unifiedMaster?.terminationGracePeriodSeconds,
    connections: getConnectionsField(
      unifiedMaster?.ports || [],
      unifiedMaster?.exposedUrls || [],
      unifiedMaster?.relatedUrls || [],
    ),
    compute: getComputeField(unifiedMaster?.compute),
    storage: getStorageField(unifiedMaster?.storage),
    security: getSecurityField(unifiedMaster?.security),
    sshAuthMountPath,
    mpiLauncherCreationPolicy,
  };
}

// this function ONLY serves as a form model for the master step to use the useHybridTemplates composable
function getHybridDistributedMasterModel(unified: UnifiedWorkload): IUISuperWorkloadCreationRequest {
  return {
    name: unified.meta?.name || "",
    projectId: unified.meta?.projectId || "",
    clusterId: unified.meta?.clusterId || "",
    spec: {
      command: unified.spec?.master?.command,
      args: unified.spec?.master?.args,
      image: unified.spec?.master?.image,
      imagePullPolicy: unified.spec?.master?.imagePullPolicy,
      workingDir: unified.spec?.master?.workingDir,
      createHomeDir: unified.spec?.master?.createHomeDir,
      nodeType: unified.spec?.master?.nodeType,
      nodePools: unified.spec?.master?.nodePools,
      tty: unified.spec?.master?.tty,
      stdin: unified.spec?.master?.stdin,
      environmentVariables: unified.spec?.master?.environmentVariables,
      annotations: unified.spec?.master?.annotations,
      labels: unified.spec?.master?.labels,
      tolerations: unified.spec?.master?.tolerations,
      terminateAfterPreemption: unified.spec?.master?.terminateAfterPreemption,
      priorityClass: unified.spec?.master?.priorityClass,
      preemptibility: unified.spec?.master?.preemptibility,
      autoDeletionTimeAfterCompletionSeconds: unified.spec?.master?.autoDeletionTimeAfterCompletionSeconds,
      backoffLimit: unified.spec?.master?.backoffLimit,
      connections: getConnectionsField(
        unified.spec?.master?.ports || [],
        unified.spec?.master?.exposedUrls || [],
        unified.spec?.master?.relatedUrls || [],
      ),
      terminationGracePeriodSeconds: unified.spec?.master?.terminationGracePeriodSeconds,
      compute: getComputeField(unified.spec?.master?.compute),
      storage: getStorageField(unified.spec?.master?.storage),
      security: getSecurityField(unified.spec?.master?.security),
    },
  };
}

function getComputeField(
  compute: WorkspaceSpecSpecAllOfCompute | undefined | null,
): WorkspaceSpecSpecAllOfCompute | undefined {
  if (!compute) return {};
  compute = { ...compute };

  if (compute.gpuPortionRequest === 1) {
    delete compute.gpuPortionRequest;
    delete compute.gpuPortionLimit;
    delete compute.gpuRequestType;
  }

  return {
    gpuDevicesRequest: compute.gpuDevicesRequest,
    gpuRequestType: compute.gpuRequestType,
    gpuComputeMode: compute.gpuComputeMode,
    gpuPortionRequest: compute.gpuPortionRequest,
    gpuPortionLimit: compute.gpuPortionLimit,
    gpuMemoryRequest: compute.gpuMemoryRequest,
    gpuMemoryLimit: compute.gpuMemoryLimit,
    cpuCoreRequest: compute.cpuCoreRequest,
    cpuCoreLimit: compute.cpuCoreLimit,
    cpuMemoryRequest: compute.cpuMemoryRequest,
    cpuMemoryLimit: compute.cpuMemoryLimit,
    largeShmRequest: compute.largeShmRequest,
    extendedResources: compute.extendedResources,
  };
}

function getStorageField(storage: WorkspaceSpecSpecAllOfStorage | undefined | null): IUIStorageFields {
  if (!storage) return {};
  const uiStorage: IUIStorageFields = {};
  uiStorage.pvc = storage.pvc
    ?.filter((p) => p.existingPvc)
    .map((pvc: PvcInstance) => ({
      id: makeId(),
      createdAt: Date.now(),
      type: AssetKind.Pvc,
      ...pvc,
    }));

  uiStorage.volumes = storage.pvc?.filter((p) => !p.existingPvc).map((pvc: PvcInstance) => pvcToUIVolumeInstance(pvc));

  uiStorage.nfs = storage.nfs?.map((nfs: NfsInstance) => ({
    id: makeId(),
    createdAt: Date.now(),
    type: AssetKind.Nfs,
    ...nfs,
  }));

  uiStorage.hostPath = storage.hostPath?.map((hostPath: HostPathInstance) => ({
    id: makeId(),
    createdAt: Date.now(),
    type: AssetKind.HostPath,
    ...hostPath,
  }));

  uiStorage.git = storage.git?.map((git: GitInstance) => ({
    id: makeId(),
    createdAt: Date.now(),
    type: AssetKind.Git,
    ...git,
  }));

  uiStorage.s3 = storage.s3?.map((s3: S3Instance) => ({
    id: makeId(),
    createdAt: Date.now(),
    type: AssetKind.S3,
    ...s3,
  }));

  uiStorage.configMapVolume = storage.configMapVolume?.map((configMap: ConfigMapInstance) => ({
    id: makeId(),
    createdAt: Date.now(),
    type: AssetKind.ConfigMap,
    ...configMap,
  }));

  uiStorage.secretVolume = storage.secretVolume?.map((secret: SecretInstance) => ({
    id: makeId(),
    createdAt: Date.now(),
    type: AssetKind.SecretVolume,
    ...secret,
  }));

  uiStorage.dataVolume = storage.dataVolume?.map((dv: DataVolumeInstance) => ({
    id: makeId(),
    createdAt: Date.now(),
    type: AssetKind.DataVolume,
    dataVolumeId: dv.id,
    mountPath: dv.mountPath,
  }));

  uiStorage.emptyDirVolume = storage.emptyDirVolume?.map((emptyDir: EmptyDirInstance) => ({
    id: makeId(),
    createdAt: Date.now(),
    type: UI_EMPTY_DIR_KIND,
    ...emptyDir,
  }));

  return uiStorage;
}

function pvcToUIVolumeInstance(pvc: PvcInstance): IUIPvcInstance {
  return {
    id: makeId(),
    createdAt: Date.now(),
    type: UI_VOLUME_KIND,
    name: pvc.name || "",
    claimName: "claim-" + makeId(10),
    path: pvc.path || "",
    claimInfo: pvc.claimInfo || {},
    ephemeral: pvc.ephemeral || false,
  };
}

function getStorageFieldFromUI(storage: IUIStorageFields): WorkspaceSpecSpecAllOfStorage {
  const storageFields: WorkspaceSpecSpecAllOfStorage = {};

  storageFields.pvc = storage.pvc?.map((pvc: IUIPvcInstance) => ({
    ...omit(pvc, ["id", "type", "createdAt", "name"]),
    // only remove the name property if its falsy
    ...(pvc.name ? { name: pvc.name } : {}),
  }));

  if (storage.volumes?.length) {
    const pvcVolumes = storage.volumes?.map((pvc: IUIPvcInstance) => ({
      ...omit(pvc, ["id", "type", "createdAt", "name"]),
      // only remove the name property if its falsy
      ...(pvc.name ? { name: pvc.name } : {}),
    }));

    if (storageFields.pvc) storageFields.pvc = [...storageFields.pvc, ...pvcVolumes];
    else storageFields.pvc = pvcVolumes;
  }

  storageFields.nfs = storage.nfs?.map((nfs: IUINfsInstance) => ({
    ...omit(nfs, ["id", "type", "createdAt"]),
  }));

  storageFields.hostPath = storage.hostPath?.map((hostPath: IUIHostPathInstance) => ({
    ...omit(hostPath, ["id", "type", "createdAt"]),
  }));

  storageFields.git = storage.git?.map((git: IUIGitInstance) => ({
    ...omit(git, ["id", "type", "createdAt"]),
  }));

  storageFields.s3 = storage.s3?.map((s3: IUIS3Instance) => ({
    ...omit(s3, ["id", "type", "createdAt"]),
  }));

  storageFields.configMapVolume = storage.configMapVolume?.map((configMap: IUIConfigMapInstance) => ({
    ...omit(configMap, ["id", "type", "createdAt"]),
    subPath: configMap.subPath || undefined,
  }));

  storageFields.secretVolume = storage.secretVolume?.map((secret: IUISecretVolumeInstance) => ({
    ...omit(secret, ["id", "type", "createdAt"]),
  }));

  storageFields.emptyDirVolume = storage.emptyDirVolume?.map((emptyDir: IUIEmptyDirInstance) => ({
    ...omit(emptyDir, ["id", "type", "createdAt"]),
  }));

  storageFields.dataVolume = storage.dataVolume?.map((dv: IUIDataVolumeInstance) => ({
    id: dv.dataVolumeId,
    mountPath: dv.mountPath,
  }));
  return storageFields;
}

export function _getCleanedStorageFieldsFromUI(storage: IUIStorageFields): WorkspaceSpecSpecAllOfStorage | undefined {
  const storageFields: WorkspaceSpecSpecAllOfStorage = {};

  type IUISource =
    | IUIPvcInstance
    | IUINfsInstance
    | IUIHostPathInstance
    | IUIGitInstance
    | IUIS3Instance
    | IUIConfigMapInstance
    | IUISecretVolumeInstance
    | IUIEmptyDirInstance;

  function removeFields<T extends IUISource>(
    arr?: Array<T>,
    emptyStringFieldsToRemove: (keyof T)[] = [],
  ): Array<Omit<T, "id" | "type" | "createdAt" | keyof T>> | undefined {
    if (!arr || arr.length === 0) return undefined;
    return arr.map((item) => {
      return omit(item, ["id", "type", "createdAt", ...emptyStringFieldsToRemove]);
    });
  }

  storageFields.pvc =
    storage.pvc && storage.pvc.length > 0
      ? storage.pvc.map((pvc: IUIPvcInstance) => ({
          ...omit(pvc, ["id", "type", "createdAt", "name"]),
          ...(pvc.name ? { name: pvc.name } : {}),
        }))
      : undefined;
  storageFields.nfs = removeFields<IUINfsInstance>(storage.nfs);
  storageFields.hostPath = removeFields<IUIHostPathInstance>(storage.hostPath);
  storageFields.git = removeFields<IUIGitInstance>(storage.git);
  storageFields.s3 = removeFields<IUIS3Instance>(storage.s3);
  storageFields.configMapVolume = removeFields<IUIConfigMapInstance>(storage.configMapVolume);
  storageFields.secretVolume = removeFields<IUISecretVolumeInstance>(storage.secretVolume);
  storageFields.emptyDirVolume = removeFields<IUIEmptyDirInstance>(storage.emptyDirVolume);

  if (storage.volumes?.length) {
    const pvcVolumes = storage.volumes.map((pvc: IUIPvcInstance) => ({
      ...omit(pvc, ["id", "type", "createdAt", "name"]),
      ...(pvc.name ? { name: pvc.name } : {}),
    }));
    if (storageFields.pvc && pvcVolumes) storageFields.pvc = [...storageFields.pvc, ...pvcVolumes];
    else storageFields.pvc = pvcVolumes;
  }

  storageFields.dataVolume =
    storage.dataVolume && storage.dataVolume.length > 0
      ? storage.dataVolume?.map((dv: IUIDataVolumeInstance) => ({
          id: dv.dataVolumeId,
          mountPath: dv.mountPath,
        }))
      : undefined;

  Object.keys(storageFields).forEach((key: string) => {
    if (!storageFields[key as keyof WorkspaceSpecSpecAllOfStorage]) {
      delete storageFields[key as keyof WorkspaceSpecSpecAllOfStorage];
    }
  });
  return Object.keys(storageFields).length > 0 ? storageFields : undefined;
}

export function _getCleanedComputeFieldsFromUI(
  compute: WorkspaceSpecSpecAllOfCompute,
): WorkspaceSpecSpecAllOfCompute | undefined {
  const retCompute = getComputeFieldFromUI(compute);
  Object.keys(retCompute).forEach((key: string) => {
    if (!retCompute[key as keyof WorkspaceSpecSpecAllOfCompute]) {
      delete retCompute[key as keyof WorkspaceSpecSpecAllOfCompute];
    }
  });

  return Object.keys(retCompute).length > 0 ? retCompute : undefined;
}

export function _getCleanedSecurityFieldsFromUI(
  security?: WorkspaceSpecSpecAllOfSecurity | null,
): WorkspaceSpecSpecAllOfSecurity | undefined {
  const retSecurity = getSecurityFieldForCreationRequest(security);

  return Object.keys(retSecurity).length > 0 ? retSecurity : undefined;
}

export function _getCleanedConnectionsFieldsFromUI(connections?: Array<IUIConnection>): IURLAndPorts {
  const retConnections: IURLAndPorts = _connectionsToFlatRequest(connections);

  Object.keys(retConnections).forEach((key: string) => {
    if (retConnections[key as keyof IURLAndPorts]?.length === 0) {
      retConnections[key as keyof IURLAndPorts] = undefined;
    }
  });

  return retConnections;
}

function _getBaseFlatWorkloadSpecData(workloadSpec: IUISuperWorkloadSpec): SupersetSpec {
  if (!workloadSpec.image) {
    throw new Error(`Invalid workloadSpec configuration: image is required but was not provided.`);
  }

  const storage = _getCleanedStorageFieldsFromUI(workloadSpec.storage || {});
  const compute = _getCleanedComputeFieldsFromUI(workloadSpec.compute || {});
  const security = _getCleanedSecurityFieldsFromUI(workloadSpec.security);
  const connections = _getCleanedConnectionsFieldsFromUI(workloadSpec.connections);

  const retWorkloadSpec: SupersetSpec = {
    command: workloadSpec.command || undefined,
    args: workloadSpec.args || undefined,
    image: workloadSpec.image,
    imagePullPolicy: workloadSpec.imagePullPolicy || undefined,
    imagePullSecrets: workloadSpec.imagePullSecrets?.length ? workloadSpec.imagePullSecrets : undefined,
    workingDir: workloadSpec.workingDir || undefined,
    createHomeDir: workloadSpec.createHomeDir || undefined,
    probes: Object.keys(workloadSpec.probes || {}).length ? workloadSpec.probes : undefined,
    nodeType: workloadSpec.nodeType || undefined,
    nodePools: workloadSpec.nodePools?.length ? workloadSpec.nodePools : undefined,
    podAffinity: workloadSpec.podAffinity || undefined,
    tty: isBoolean(workloadSpec.tty) ? workloadSpec.tty : undefined,
    stdin: isBoolean(workloadSpec.stdin) ? workloadSpec.tty : undefined,
    environmentVariables: workloadSpec.environmentVariables?.length ? workloadSpec.environmentVariables : undefined,
    annotations: workloadSpec.annotations?.length ? workloadSpec.annotations : undefined,
    labels: workloadSpec.labels?.length ? workloadSpec.labels : undefined,
    tolerations: workloadSpec.tolerations?.length ? workloadSpec.tolerations : undefined,
    restartPolicy: (workloadSpec.restartPolicy as RestartPolicy | null | undefined) || undefined,
    priorityClass: workloadSpec.priorityClass || undefined,
    preemptibility: workloadSpec.preemptibility || undefined,
    storage,
    compute,
    security,
    ...connections,
  };
  return retWorkloadSpec;
}

function getComputeFieldFromUI(compute: WorkspaceSpecSpecAllOfCompute): WorkspaceSpecSpecAllOfCompute {
  return {
    ...omit(compute, ["gpuMemoryPortionRequest", "gpuMemoryPortionLimit"]),
    gpuPortionRequest: compute.gpuPortionRequest === 1 ? undefined : compute.gpuPortionRequest,
    gpuRequestType: compute.gpuPortionRequest === 1 ? undefined : compute.gpuRequestType,
  };
}

function getSecurityField(
  security: WorkspaceSpecSpecAllOfSecurity | undefined | null,
): WorkspaceSpecSpecAllOfSecurity | undefined {
  if (!security) return {};

  return {
    uidGidSource: security.uidGidSource,
    capabilities: security.capabilities,
    seccompProfileType: security.seccompProfileType,
    runAsNonRoot: security.runAsNonRoot,
    readOnlyRootFilesystem: security.readOnlyRootFilesystem,
    runAsGid: security.runAsGid,
    runAsUid: security.runAsUid,
    supplementalGroups: security.supplementalGroups,
    allowPrivilegeEscalation: security.allowPrivilegeEscalation,
    hostIpc: security.hostIpc,
    hostNetwork: security.hostNetwork,
  };
}

function getConnectionsField(ports: Port[], exposedUrls: ExposedUrl[], relatedUrls: RelatedUrl[]): IUIConnection[] {
  const connections: IUIConnection[] = [];
  ports.forEach((port: Port) => {
    connections.push(convertPortToUiConnection(port));
  });

  exposedUrls.forEach((url: ExposedUrl) => {
    connections.push(convertExposedUrlToUiConnection(url));
  });

  relatedUrls.forEach((url: RelatedUrl) => {
    connections.push(convertRelatedUrlToUiConnection(url));
  });

  return connections;
}

function _convertAuthorizationTypeToAccessOption(
  authorizationType: AuthorizationType | null | undefined,
): EAccessOptions | undefined {
  if (!authorizationType) return undefined;

  switch (authorizationType) {
    case AuthorizationType.AuthenticatedUsers:
      return EAccessOptions.EVERYONE;
    case AuthorizationType.AuthorizedUsers:
      return EAccessOptions.SPECIFIC_USERS;
    case AuthorizationType.AuthorizedGroups:
      return EAccessOptions.GROUPS;
    default:
      return undefined;
  }
}

function convertPortToUiConnection(port: Port): IUIConnection {
  return {
    id: makeId(),
    name: port.toolName || "",
    toolType: port.toolType as ToolType,
    containerPort: port.container || undefined,
    isCustomPort: !!port.external,
    nodePort: port.external,
    connectionType: (port.serviceType as InternalConnectionType) || InternalConnectionType.NodePort,
  };
}

function convertExposedUrlToUiConnection(url: ExposedUrl): IUIConnection {
  return {
    id: makeId(),
    name: url.name || url.toolName || "",
    toolType: url.toolType as ToolType,
    connectionType: InternalConnectionType.ExternalUrl,
    isCustomUrl: !!url.url,
    externalUrl: url.url,
    authorizedUsers: url.authorizedUsers,
    authorizedGroups: url.authorizedGroups,
    containerPort: url.container || undefined,
    accessOption: _convertAuthorizationTypeToAccessOption(url.authorizationType as AuthorizationType),
  };
}

function convertRelatedUrlToUiConnection(url: RelatedUrl): IUIConnection {
  return {
    id: makeId(),
    toolType: url.type as ToolType,
    name: url.name || "",
    externalUrl: url.url,
    connectionType: InternalConnectionType.ExternalUrl,
  };
}

function setBaseWorkloadUIDefaults(
  uiModel: IUISuperWorkloadCreationRequest,
  isSSO: boolean,
  policy?: UnifiedPolicyInfoPerReplicaType | null,
): IUISuperWorkloadCreationRequest {
  // checking specific values with the policy before setting
  let imagePullPolicyUIDefault: ImagePullPolicy | null = ImagePullPolicy.IfNotPresent;
  if (Array.isArray(policy?.rules?.imagePullPolicy?.options)) {
    if (!policy.rules.imagePullPolicy.options.map((o) => o.value).includes(ImagePullPolicy.IfNotPresent)) {
      imagePullPolicyUIDefault = null;
    }
  }

  let uidGidSourceDefault: UidGidSource | null = isSSO ? UidGidSource.FromIdpToken : UidGidSource.FromTheImage;
  if (Array.isArray(policy?.rules?.security?.uidGidSource?.options)) {
    if (!policy.rules.security.uidGidSource.options.map((o) => o.value).includes(uidGidSourceDefault)) {
      uidGidSourceDefault = null;
    }
  }

  return {
    ...uiModel,
    spec: {
      ...uiModel.spec,
      image: fallbackDefaultIfNullOrUndefined(uiModel.spec.image, ""),
      imagePullPolicy: fallbackDefaultIfNullOrUndefined(uiModel.spec.imagePullPolicy, imagePullPolicyUIDefault),
      environmentVariables: fallbackDefaultIfNullOrUndefined(uiModel.spec.environmentVariables, []),
      connections: fallbackDefaultIfNullOrUndefined(uiModel.spec.connections, []),
      security: {
        ...uiModel.spec.security,
        uidGidSource: fallbackDefaultIfNullOrUndefined(uiModel.spec?.security?.uidGidSource, uidGidSourceDefault),
        capabilities: fallbackDefaultIfNullOrUndefined(uiModel.spec?.security?.capabilities, []),
      },
    },
  };
}

function setWorkspaceUIDefaults(
  uiModel: IUISuperWorkloadCreationRequest,
  isSSO: boolean,
  policy?: UnifiedPolicyInfoPerReplicaType | null,
  clusterVersion?: string | null,
): IUISuperWorkloadCreationRequest {
  const workloadUI: IUISuperWorkloadCreationRequest = setBaseWorkloadUIDefaults(uiModel, isSSO, policy);

  const versionDefaults = getUIDefaultsByClusterVersion(clusterVersion || "", uiModel.spec);
  return {
    ...workloadUI,
    spec: {
      ...workloadUI.spec,
      backoffLimit: fallbackDefaultIfNullOrUndefined(uiModel.spec?.backoffLimit, 6),
      preemptibility: versionDefaults.preemptibility,
      terminationGracePeriodSeconds: versionDefaults.terminationGracePeriodSeconds,
      autoDeletionTimeAfterCompletionSeconds: fallbackDefaultIfNullOrUndefined(
        uiModel.spec?.autoDeletionTimeAfterCompletionSeconds,
        60 * 60 * 24 * 30, // 30 days in seconds
      ),
    },
  };
}

function setTrainingUIDefaults(
  uiModel: IUISuperWorkloadCreationRequest,
  isSSO: boolean,
  policy?: UnifiedPolicyInfoPerReplicaType | null,
  clusterVersion?: string | null,
): IUISuperWorkloadCreationRequest {
  // checking specific values with the policy before setting
  let imagePullPolicyUIDefault: ImagePullPolicy | null = ImagePullPolicy.IfNotPresent;
  if (Array.isArray(policy?.rules?.imagePullPolicy?.options)) {
    if (!policy.rules.imagePullPolicy.options.map((o) => o.value).includes(ImagePullPolicy.IfNotPresent)) {
      imagePullPolicyUIDefault = null;
    }
  }

  let uidGidSourceDefault: UidGidSource | null = isSSO ? UidGidSource.FromIdpToken : UidGidSource.FromTheImage;
  if (Array.isArray(policy?.rules?.security?.uidGidSource?.options)) {
    if (!policy.rules.security.uidGidSource.options.map((o) => o.value).includes(uidGidSourceDefault)) {
      uidGidSourceDefault = null;
    }
  }

  const versionDefaults = getUIDefaultsByClusterVersion(clusterVersion || "", uiModel.spec);

  return {
    ...uiModel,
    spec: {
      ...uiModel.spec,
      completions: versionDefaults.completions,
      parallelism: versionDefaults.parallelism,
      image: fallbackDefaultIfNullOrUndefined(uiModel.spec.image, ""),
      imagePullPolicy: fallbackDefaultIfNullOrUndefined(uiModel.spec.imagePullPolicy, imagePullPolicyUIDefault),
      environmentVariables: fallbackDefaultIfNullOrUndefined(uiModel.spec.environmentVariables, []),
      autoDeletionTimeAfterCompletionSeconds: fallbackDefaultIfNullOrUndefined(
        uiModel.spec?.autoDeletionTimeAfterCompletionSeconds,
        60 * 60 * 24 * 30, // 30 days in seconds
      ),
      backoffLimit: fallbackDefaultIfNullOrUndefined(uiModel.spec?.backoffLimit, 6),
      preemptibility: versionDefaults.preemptibility,
      terminationGracePeriodSeconds: versionDefaults.terminationGracePeriodSeconds,
      connections: fallbackDefaultIfNullOrUndefined(uiModel.spec.connections, []),

      security: {
        ...uiModel.spec.security,
        uidGidSource: fallbackDefaultIfNullOrUndefined(uiModel.spec?.security?.uidGidSource, uidGidSourceDefault),
        capabilities: fallbackDefaultIfNullOrUndefined(uiModel.spec?.security?.capabilities, []),
      },
    },
  };
}

function setDistributedUIDefaults(
  uiModel: IUISuperWorkloadCreationRequest,
  isSSO: boolean,
  policy?: UnifiedPolicyInfoPerReplicaType | null,
  clusterVersion?: string | null,
): IUISuperWorkloadCreationRequest {
  // checking specific values with the policy before setting
  let imagePullPolicyUIDefault: ImagePullPolicy | null = ImagePullPolicy.IfNotPresent;
  if (Array.isArray(policy?.rules?.imagePullPolicy?.options)) {
    if (!policy.rules.imagePullPolicy.options.map((o) => o.value).includes(ImagePullPolicy.IfNotPresent)) {
      imagePullPolicyUIDefault = null;
    }
  }

  let uidGidSourceDefault: UidGidSource | null = isSSO ? UidGidSource.FromIdpToken : UidGidSource.FromTheImage;
  if (Array.isArray(policy?.rules?.security?.uidGidSource?.options)) {
    if (!policy.rules.security.uidGidSource.options.map((o) => o.value).includes(uidGidSourceDefault)) {
      uidGidSourceDefault = null;
    }
  }

  const versionDefaults = getUIDefaultsByClusterVersion(clusterVersion || "", uiModel.spec);

  return {
    ...uiModel,
    spec: {
      ...uiModel.spec,

      image: fallbackDefaultIfNullOrUndefined(uiModel.spec.image, ""),
      imagePullPolicy: fallbackDefaultIfNullOrUndefined(uiModel.spec.imagePullPolicy, imagePullPolicyUIDefault),
      environmentVariables: fallbackDefaultIfNullOrUndefined(uiModel.spec.environmentVariables, []),
      autoDeletionTimeAfterCompletionSeconds: fallbackDefaultIfNullOrUndefined(
        uiModel.spec?.autoDeletionTimeAfterCompletionSeconds,
        60 * 60 * 24 * 30, // 30 days in seconds
      ),
      backoffLimit: fallbackDefaultIfNullOrUndefined(uiModel.spec?.backoffLimit, 6),
      preemptibility: versionDefaults.preemptibility,
      terminationGracePeriodSeconds: versionDefaults.terminationGracePeriodSeconds,
      connections: fallbackDefaultIfNullOrUndefined(uiModel.spec.connections, []),

      security: {
        ...uiModel.spec.security,
        uidGidSource: fallbackDefaultIfNullOrUndefined(uiModel.spec?.security?.uidGidSource, uidGidSourceDefault),
        capabilities: fallbackDefaultIfNullOrUndefined(uiModel.spec?.security?.capabilities, []),
      },
    },
  };
}

function setDistributedMasterSpecUIDefaults(
  masterSpec: IUISuperWorkloadSpec,
  isSSO: boolean,
  policy?: UnifiedPolicyInfoPerReplicaType | null,
  clusterVersion?: string | null,
): IUISuperWorkloadSpec {
  // checking specific values with the policy before setting
  let imagePullPolicyUIDefault: ImagePullPolicy | null = ImagePullPolicy.IfNotPresent;
  if (Array.isArray(policy?.rules?.imagePullPolicy?.options)) {
    if (!policy.rules.imagePullPolicy.options.map((o) => o.value).includes(ImagePullPolicy.IfNotPresent)) {
      imagePullPolicyUIDefault = null;
    }
  }

  let uidGidSourceDefault: UidGidSource | null = isSSO ? UidGidSource.FromIdpToken : UidGidSource.FromTheImage;
  if (Array.isArray(policy?.rules?.security?.uidGidSource?.options)) {
    if (!policy.rules.security.uidGidSource.options.map((o) => o.value).includes(uidGidSourceDefault)) {
      uidGidSourceDefault = null;
    }
  }

  const versionDefaults = getUIDefaultsByClusterVersion(clusterVersion || "", masterSpec);

  return {
    ...masterSpec,
    image: fallbackDefaultIfNullOrUndefined(masterSpec.image, ""),
    imagePullPolicy: fallbackDefaultIfNullOrUndefined(masterSpec.imagePullPolicy, imagePullPolicyUIDefault),
    environmentVariables: fallbackDefaultIfNullOrUndefined(masterSpec.environmentVariables, []),
    autoDeletionTimeAfterCompletionSeconds: fallbackDefaultIfNullOrUndefined(
      masterSpec?.autoDeletionTimeAfterCompletionSeconds,
      60 * 60 * 24 * 30, // 30 days in seconds
    ),
    backoffLimit: fallbackDefaultIfNullOrUndefined(masterSpec?.backoffLimit, 6),
    preemptibility: versionDefaults.preemptibility,
    terminationGracePeriodSeconds: versionDefaults.terminationGracePeriodSeconds,
    connections: fallbackDefaultIfNullOrUndefined(masterSpec.connections, []),

    security: {
      ...masterSpec.security,
      uidGidSource: fallbackDefaultIfNullOrUndefined(masterSpec?.security?.uidGidSource, uidGidSourceDefault),
      capabilities: fallbackDefaultIfNullOrUndefined(masterSpec?.security?.capabilities, []),
    },
  };
}

function getUIDefaultsByClusterVersion(clusterVersion: string, uiModelSpec: IUISuperWorkloadSpec): SupersetSpec {
  const DEFAULT_VALUES = {
    COMPLETIONS: 1,
    PARALLELISM: null,
    TERMINATION_GRACE_PERIOD_SECONDS: 30,
  } as const;

  const supportRunsAndParallelism = isNewerVersion(clusterVersion, MIN_CLUSTER_VERSION_FOR_COMPLETIONS_AND_PARALLELISM);
  const supportPreemptibility = isNewerVersion(clusterVersion, MIN_CLUSTER_VERSION_FOR_PREEMPTIBILITY);
  const supportSemiPreemptibility = isNewerVersion(clusterVersion, MIN_CLUSTER_VERSION_FOR_SEMI_PREEMPTIBILITY);
  const supportTerminationGracePeriod = isNewerVersion(clusterVersion, MIN_CLUSTER_VERSION_FOR_TERMINATION_GRACE_PERIOD);

  return {
    completions: supportRunsAndParallelism
      ? fallbackDefaultIfNullOrUndefined(uiModelSpec.completions, DEFAULT_VALUES.COMPLETIONS)
      : null,
    parallelism: supportRunsAndParallelism
      ? fallbackDefaultIfNullOrUndefined(uiModelSpec.parallelism, DEFAULT_VALUES.PARALLELISM)
      : null,
    terminationGracePeriodSeconds: supportTerminationGracePeriod
      ? fallbackDefaultIfNullOrUndefined(
          uiModelSpec.terminationGracePeriodSeconds,
          DEFAULT_VALUES.TERMINATION_GRACE_PERIOD_SECONDS,
        )
      : null,
    preemptibility:
      supportPreemptibility &&
      (uiModelSpec.preemptibility !== Preemptibility.SemiPreemptible || supportSemiPreemptibility)
        ? uiModelSpec.preemptibility
        : undefined,
  };
}

function getRelevantComputeFields(model: IUISuperWorkloadCreationRequest): IUISuperWorkloadSpec["compute"] {
  return {
    ...model.spec.compute,
  };
}

function getRelevantStorageFields(model: IUISuperWorkloadCreationRequest): IUISuperWorkloadSpec["storage"] {
  return model.spec.storage;
}

function getBaseEnvironmentFields(spec: IUISuperWorkloadSpec): IUIEnvironmentFields {
  return {
    command: spec.command,
    args: spec.args,
    image: spec.image,
    imagePullPolicy: spec.imagePullPolicy,
    workingDir: spec.workingDir,
    createHomeDir: spec.createHomeDir,
    environmentVariables: spec.environmentVariables,
    connections: spec.connections,
    security: spec.security,
  };
}

function getRelevantEnvironmentFields(model: IUISuperWorkloadCreationRequest): IUIEnvironmentFields {
  const environmentFields: IUIEnvironmentFields = getBaseEnvironmentFields(model.spec);
  return environmentFields;
}

function getInferenceRelevantEnvironmentFields(model: IUISuperWorkloadCreationRequest): IUIEnvironmentFields {
  const environmentFields: IUIEnvironmentFields = getBaseEnvironmentFields(model.spec);
  environmentFields.servingPort = {
    protocol: model.spec.servingPort?.protocol,
    container: model.spec.servingPort?.container,
    authorizationType: model.spec.servingPort?.authorizationType,
    authorizedGroups: model.spec.servingPort?.authorizedGroups,
    authorizedUsers: model.spec.servingPort?.authorizedUsers,
  };

  return environmentFields;
}

function getDistributedInferenceLeaderRelevantEnvironmentFields(
  model: IUISuperWorkloadCreationRequest,
): IUIEnvironmentFields {
  const environmentFields: IUIEnvironmentFields = getBaseEnvironmentFields(model.spec.leader || {});
  environmentFields.servingPort = {
    protocol: model.spec.leader?.servingPort?.protocol,
    container: model.spec.leader?.servingPort?.container,
    authorizationType: model.spec.leader?.servingPort?.authorizationType,
    authorizedGroups: model.spec.leader?.servingPort?.authorizedGroups,
    authorizedUsers: model.spec.leader?.servingPort?.authorizedUsers,
  };

  return environmentFields;
}

function getDistributedInferenceLeaderClearedEnvironmentFields(
  model: IUISuperWorkloadCreationRequest,
): IUIEnvironmentFields {
  const environmentFields: IUIEnvironmentFields = getBaseEnvironmentFields(model.spec.leader || {});
  environmentFields.servingPort = {
    protocol: model.spec.servingPort?.protocol || "http",
    container: model.spec.servingPort?.container,
    authorizationType: model.spec.servingPort?.authorizationType,
    authorizedGroups: model.spec.servingPort?.authorizedGroups,
    authorizedUsers: model.spec.servingPort?.authorizedUsers,
  };

  return environmentFields;
}

function combineWorkloadModelWithComputeDetails(
  initialWorkloadModel: IUISuperWorkloadCreationRequest,
  computeFields: IUISuperWorkloadSpec["compute"],
): IUISuperWorkloadCreationRequest {
  return {
    ...initialWorkloadModel,
    spec: {
      ...initialWorkloadModel.spec,
      compute: computeFields,
    },
  };
}

function combineDistributedInferenceLeaderWorkloadModelWithComputeDetails(
  initialWorkloadModel: IUISuperWorkloadCreationRequest,
  computeFields: IUISuperWorkloadSpec["compute"],
): IUISuperWorkloadCreationRequest {
  return {
    ...initialWorkloadModel,
    spec: { ...initialWorkloadModel.spec, leader: { ...initialWorkloadModel.spec.leader, compute: computeFields } },
  };
}

function combineDistributedInferenceWorkerWorkloadModelWithComputeDetails(
  initialWorkloadModel: IUISuperWorkloadCreationRequest,
  computeFields: IUISuperWorkloadSpec["compute"],
): IUISuperWorkloadCreationRequest {
  return {
    ...initialWorkloadModel,
    spec: { ...initialWorkloadModel.spec, worker: { ...initialWorkloadModel.spec.worker, compute: computeFields } },
  };
}

function getStorageFields(initialStorage?: IUIStorageFields, storageFields?: IUIStorageFields): IUIStorageFields {
  return {
    pvc: combineExistingWithNewStorageFields(initialStorage?.pvc, storageFields?.pvc),
    nfs: combineExistingWithNewStorageFields(initialStorage?.nfs, storageFields?.nfs),
    hostPath: combineExistingWithNewStorageFields(initialStorage?.hostPath, storageFields?.hostPath),
    git: combineExistingWithNewStorageFields(initialStorage?.git, storageFields?.git),
    s3: combineExistingWithNewStorageFields(initialStorage?.s3, storageFields?.s3),
    configMapVolume: combineExistingWithNewStorageFields(
      initialStorage?.configMapVolume,
      storageFields?.configMapVolume,
    ),
    secretVolume: combineExistingWithNewStorageFields(initialStorage?.secretVolume, storageFields?.secretVolume),
    dataVolume: combineExistingWithNewStorageFields(initialStorage?.dataVolume, storageFields?.dataVolume),
    volumes: initialStorage?.volumes,
  };
}

function combineWorkloadModelWithStorageDetails(
  initialWorkloadModel: IUISuperWorkloadCreationRequest,
  storageFields?: IUIStorageFields,
): IUISuperWorkloadCreationRequest {
  return {
    ...initialWorkloadModel,
    spec: {
      ...initialWorkloadModel.spec,
      storage: getStorageFields(initialWorkloadModel.spec.storage, storageFields),
    },
  };
}

function combineDistributedInferenceLeaderWorkloadModelWithStorageDetails(
  initialWorkloadModel: IUISuperWorkloadCreationRequest,
  storageFields: IUIStorageFields,
): IUISuperWorkloadCreationRequest {
  return {
    ...initialWorkloadModel,
    spec: {
      ...initialWorkloadModel.spec,
      leader: {
        ...initialWorkloadModel.spec.leader,
        storage: getStorageFields(initialWorkloadModel.spec.leader?.storage, storageFields),
      },
    },
  };
}

function combineDistributedInferenceWorkerWorkloadModelWithStorageDetails(
  initialWorkloadModel: IUISuperWorkloadCreationRequest,
  storageFields: IUIStorageFields,
): IUISuperWorkloadCreationRequest {
  return {
    ...initialWorkloadModel,
    spec: {
      ...initialWorkloadModel.spec,
      worker: {
        ...initialWorkloadModel.spec.worker,
        storage: getStorageFields(initialWorkloadModel.spec.worker?.storage, storageFields),
      },
    },
  };
}

function combineWorkloadModelWithEnvironmentDetails(
  initialWorkloadModel: IUISuperWorkloadCreationRequest,
  envFields: IUIEnvironmentFields,
): IUISuperWorkloadCreationRequest {
  return {
    ...initialWorkloadModel,
    spec: {
      ...initialWorkloadModel.spec,
      ...envFields,
    },
  };
}

function combineDistributedInferenceLeaderWorkloadModelWithEnvironmentDetails(
  initialWorkloadModel: IUISuperWorkloadCreationRequest,
  envFields: IUIEnvironmentFields,
): IUISuperWorkloadCreationRequest {
  return {
    ...initialWorkloadModel,
    spec: {
      ...initialWorkloadModel.spec,
      servingPort: envFields.servingPort,
      leader: {
        ...initialWorkloadModel.spec.leader,
        ...omit(envFields, ["servingPort"]),
      },
    },
  };
}

function combineDistributedInferenceWorkerWorkloadModelWithEnvironmentDetails(
  initialWorkloadModel: IUISuperWorkloadCreationRequest,
  envFields: IUIEnvironmentFields,
): IUISuperWorkloadCreationRequest {
  return {
    ...initialWorkloadModel,
    spec: {
      ...initialWorkloadModel.spec,
      worker: {
        ...initialWorkloadModel.spec.worker,
        ...envFields,
      },
    },
  };
}

function combineExistingWithNewStorageFields(
  initialStorageModel: Array<TUIStorageInstance> = [],
  newStorageModel: Array<TUIStorageInstance> = [],
): Array<TUIStorageInstance> {
  const initialStorageNames = new Set(initialStorageModel.filter((s) => !!s.name).map((s) => s.name));
  return [...initialStorageModel, ...newStorageModel.filter((s) => !initialStorageNames.has(s.name))].map((s) => ({
    ...s,
    createdAt: s.createdAt || Date.now(),
  }));
}

function combineMasterReplicaWorkloadModelWithComputeDetails(
  initialMasterSpec: IUISuperWorkloadSpec,
  computeFields: IUISuperWorkloadSpec["compute"],
): IUISuperWorkloadSpec {
  return {
    ...initialMasterSpec,
    compute: computeFields,
  };
}

function combineMasterReplicaWorkloadModelWithStorageDetails(
  initialMasterSpec: IUISuperWorkloadSpec,
  storageFields?: IUIStorageFields,
): IUISuperWorkloadSpec {
  return {
    ...initialMasterSpec,
    storage: {
      pvc: combineExistingWithNewStorageFields(initialMasterSpec?.storage?.pvc, storageFields?.pvc),
      nfs: combineExistingWithNewStorageFields(initialMasterSpec?.storage?.nfs, storageFields?.nfs),
      hostPath: combineExistingWithNewStorageFields(initialMasterSpec?.storage?.hostPath, storageFields?.hostPath),
      git: combineExistingWithNewStorageFields(initialMasterSpec?.storage?.git, storageFields?.git),
      s3: combineExistingWithNewStorageFields(initialMasterSpec?.storage?.s3, storageFields?.s3),
      configMapVolume: combineExistingWithNewStorageFields(
        initialMasterSpec?.storage?.configMapVolume,
        storageFields?.configMapVolume,
      ),
      secretVolume: combineExistingWithNewStorageFields(
        initialMasterSpec?.storage?.secretVolume,
        storageFields?.secretVolume,
      ),
      dataVolume: combineExistingWithNewStorageFields(initialMasterSpec?.storage?.dataVolume, storageFields?.dataVolume),
      volumes: initialMasterSpec.storage?.volumes,
    },
  };
}

function combineMasterReplicaWorkloadModelWithEnvironmentDetails(
  initialMasterSpec: IUISuperWorkloadSpec,
  envFields: IUIEnvironmentFields,
): IUISuperWorkloadSpec {
  return {
    ...initialMasterSpec,
    ...envFields,
  };
}

function getSecurityFieldForCreationRequest(
  security: WorkspaceSpecSpecAllOfSecurity | undefined | null,
): WorkspaceSpecSpecAllOfSecurity {
  if (security?.uidGidSource !== UidGidSource.FromIdpToken) return { ...security };

  // Omit identity values even when loaded from a template or existing workload.
  // The server applies policy defaults, then the submitting user's token claims.
  return {
    ...security,
    runAsUid: undefined,
    runAsGid: undefined,
    supplementalGroups: undefined,
  };
}

function prepareWorkloadFlatRequest(model: IUISuperWorkloadCreationRequest): WorkspaceCreationRequest {
  const storage = getStorageFieldFromUI(model.spec.storage || {});
  const compute = getComputeFieldFromUI(model.spec.compute || {});
  const security = getSecurityFieldForCreationRequest(model.spec.security);

  model = _removeTemplateFieldsFromModel(model);

  return {
    ...model,
    spec: {
      ...omit(model.spec, ["connections"]),
      ..._connectionsToFlatRequest(model.spec.connections),
      restartPolicy: (model.spec.restartPolicy as RestartPolicy | null | undefined) || undefined,
      nodeType: model.spec.nodeType || undefined,
      storage,
      security,
      compute,
    },
  };
}

function prepareTrainingFlatRequest(model: IUISuperWorkloadCreationRequest): TrainingCreationRequest {
  const storage = getStorageFieldFromUI(model.spec.storage || {});
  const compute = getComputeFieldFromUI(model.spec.compute || {});
  const security = getSecurityFieldForCreationRequest(model.spec.security);

  model = _removeTemplateFieldsFromModel(model);

  return {
    ...model,
    spec: {
      ...omit(model.spec, ["connections"]),
      ..._connectionsToFlatRequest(model.spec.connections),
      restartPolicy: (model.spec.restartPolicy as RestartPolicy | null | undefined) || undefined,
      nodeType: model.spec.nodeType || undefined,
      storage,
      compute,
      security,
    },
  };
}

function prepareDistributedFlatRequest(model: IUISuperWorkloadCreationRequest): DistributedCreationRequest {
  const storage = getStorageFieldFromUI(model.spec.storage || {});
  const compute = getComputeFieldFromUI(model.spec.compute || {});
  const security = getSecurityFieldForCreationRequest(model.spec.security);

  model = _removeTemplateFieldsFromModel(model);

  const request: DistributedCreationRequest = {
    ...model,
    spec: {
      ...omit(model.spec, ["connections"]),
      ..._connectionsToFlatRequest(model.spec.connections),
      restartPolicy: (model.spec.restartPolicy as RestartPolicy | null) || undefined,
      nodeType: model.spec.nodeType || undefined,
      storage,
      security,
      compute,
    },
    // at the moment this data is junk that remains after the spread operator, we will initialize it later on if its needed
    masterSpec: undefined,
  };

  // if sameAsWorker is falsy AND masterSpec exists, we need prepare it as well
  if (!request.masterSpecSameAsWorker && model.masterSpec) {
    const security = getSecurityFieldForCreationRequest(model.masterSpec.security);

    // note: we need to remove sshAuthMountPath from the masterSpec since it's not supported in the API
    request.masterSpec = {
      ..._omitIrrelevantMasterSpecFields(model.masterSpec),
      ..._connectionsToFlatRequest(model.masterSpec.connections),
      restartPolicy: (model.masterSpec.restartPolicy as RestartPolicy | null | undefined) || undefined,
      nodeType: model.masterSpec.nodeType || undefined,
      storage: getStorageFieldFromUI(model.masterSpec.storage || {}),
      compute: _stripMpiMasterGpuFields(
        getComputeFieldFromUI(model.masterSpec.compute || {}),
        model.spec.distributedFramework,
      ),
      security,
    };
  }
  return request;
}

interface IURLAndPorts {
  exposedUrls: WorkspaceSpecSpec["exposedUrls"];
  ports: WorkspaceSpecSpec["ports"];
  relatedUrls: WorkspaceSpecSpec["relatedUrls"];
}

function _convertAccessOptionToAuthorizationType(
  accessOption: string | null | undefined,
): AuthorizationType | undefined {
  if (!accessOption) return undefined;

  switch (accessOption) {
    case EAccessOptions.EVERYONE:
      return AuthorizationType.AuthenticatedUsers;
    case EAccessOptions.SPECIFIC_USERS:
    case EAccessOptions.PRIVATE:
      return AuthorizationType.AuthorizedUsers;
    case EAccessOptions.GROUPS:
      return AuthorizationType.AuthorizedGroups;
    case EAccessOptions.PUBLIC:
      // PUBLIC is not supported for exposedUrls, only for serving ports
      return undefined;
    default:
      return undefined;
  }
}

function _connectionsToFlatRequest(connections?: Array<IUIConnection>): IURLAndPorts {
  const urlsAndPorts: IURLAndPorts = {
    exposedUrls: [],
    ports: [],
    relatedUrls: [],
  };

  connections?.forEach((connection: IUIConnection) => {
    if (connection.connectionType === InternalConnectionType.NodePort) {
      const port: Port = {
        container: connection.containerPort,
        external: connection.nodePort,
        toolName: connection.name,
        toolType: connection.toolType,
        serviceType: InternalConnectionType.NodePort,
      };
      urlsAndPorts.ports && urlsAndPorts.ports.push(port);
    } else if (connection.connectionType === InternalConnectionType.LoadBalancer) {
      const port: Port = {
        container: connection.containerPort,
        toolName: connection.name,
        toolType: connection.toolType,
        serviceType: InternalConnectionType.LoadBalancer,
      };
      urlsAndPorts.ports && urlsAndPorts.ports.push(port);
    } else if (connection.connectionType === InternalConnectionType.ExternalUrl) {
      if (connection.toolType === ToolType.Wandb || connection.toolType === ToolType.Comet) {
        const relatedUrl: RelatedUrl = {
          name: connection.name,
          url: connection.externalUrl,
          type: connection.toolType,
        };
        urlsAndPorts.relatedUrls && urlsAndPorts.relatedUrls.push(relatedUrl);
      } else {
        const exposedUrl: ExposedUrl = {
          toolName: connection.name,
          toolType: connection.toolType,
          url: connection.externalUrl,
          container: connection.containerPort,
          authorizedUsers: connection.authorizedUsers,
          authorizedGroups: connection.authorizedGroups,
          authorizationType: _convertAccessOptionToAuthorizationType(connection.accessOption),
        };
        urlsAndPorts.exposedUrls && urlsAndPorts.exposedUrls.push(exposedUrl);
      }
    }
  });

  return urlsAndPorts;
}

function getWorkloadHybridTemplatesSetup(): THybridSelectedTemplates {
  return {
    [WorkloadCreationDrawerEntities.Environment]: {
      name: "",
      id: "",
      defaultData: null as IUIEnvironmentFields | null,
      clearedState: null as IUIEnvironmentFields | null,
      renderKey: 1001,
      sectionKey: 2000,
    },
    [WorkloadCreationDrawerEntities.DataSource]: {
      name: "",
      id: "",
      defaultData: null as IUISuperWorkloadSpec["storage"] | null,
      clearedState: null as IUISuperWorkloadSpec["storage"] | null,
      sectionKey: 2000,
    },
    [WorkloadCreationDrawerEntities.ModelStorage]: {
      name: "",
      id: "",
      defaultData: null as IUISuperWorkloadSpec["storage"] | null,
      clearedState: null as IUISuperWorkloadSpec["storage"] | null,
      sectionKey: 2000,
    },
    [WorkloadCreationDrawerEntities.Compute]: {
      name: "",
      id: "",
      defaultData: null as IUISuperWorkloadSpec["compute"] | null,
      clearedState: null as IUISuperWorkloadSpec["compute"] | null,
      renderKey: 1001,
      sectionKey: 2000,
    },
  };
}

function getEmptyUISuperWorkloadModel(): IUISuperWorkloadCreationRequest {
  return {
    name: "",
    projectId: "",
    clusterId: "",
    spec: { connections: [] },
  };
}

function getHybridWorkspaceTemplateModel(template: WorkspaceTemplate): IUISuperWorkloadCreationRequest {
  return {
    name: template.meta?.name || "",
    projectId: template.meta?.projectId?.toString() || "",
    scopeType: template.meta?.scopeType,
    scopeId: template.meta?.scopeId,
    departmentId: template.meta?.departmentId || undefined,
    clusterId: template.meta?.clusterId || "",
    description: template.meta?.description || "",
    spec: {
      command: template.spec?.command,
      args: template.spec?.args,
      image: template.spec?.image,
      imagePullPolicy: template.spec?.imagePullPolicy,
      workingDir: template.spec?.workingDir,
      createHomeDir: template.spec?.createHomeDir,
      nodeType: template.spec?.nodeType,
      nodePools: template.spec?.nodePools,
      podAffinity: template.spec?.podAffinity,
      tty: template.spec?.tty,
      stdin: template.spec?.stdin,
      environmentVariables: template.spec?.environmentVariables,
      annotations: template.spec?.annotations,
      labels: template.spec?.labels,
      tolerations: template.spec?.tolerations,
      backoffLimit: template.spec?.backoffLimit,
      priorityClass: template.spec?.priorityClass,
      preemptibility: template.spec?.preemptibility,
      terminateAfterPreemption: template.spec?.terminateAfterPreemption,
      autoDeletionTimeAfterCompletionSeconds: template.spec?.autoDeletionTimeAfterCompletionSeconds,
      terminationGracePeriodSeconds: template.spec?.terminationGracePeriodSeconds,
      connections: getConnectionsField(
        template.spec?.ports || [],
        template.spec?.exposedUrls || [],
        template.spec?.relatedUrls || [],
      ),
      compute: getComputeField(template.spec?.compute),
      storage: getStorageField(template.spec?.storage),
      security: getSecurityField(template.spec?.security),
    },
  };
}

function getHybridTrainingTemplateModel(template: TrainingTemplate): IUISuperWorkloadCreationRequest {
  return {
    name: template.meta?.name || "",
    projectId: template.meta?.projectId?.toString() || "",
    scopeType: template.meta?.scopeType,
    scopeId: template.meta?.scopeId,
    departmentId: template.meta?.departmentId || undefined,
    clusterId: template.meta?.clusterId || "",
    description: template.meta?.description || "",
    spec: {
      command: template.spec?.command,
      args: template.spec?.args,
      image: template.spec?.image,
      imagePullPolicy: template.spec?.imagePullPolicy,
      workingDir: template.spec?.workingDir,
      createHomeDir: template.spec?.createHomeDir,
      nodeType: template.spec?.nodeType,
      nodePools: template.spec?.nodePools,
      podAffinity: template.spec?.podAffinity,
      tty: template.spec?.tty,
      stdin: template.spec?.stdin,
      environmentVariables: template.spec?.environmentVariables,
      annotations: template.spec?.annotations,
      labels: template.spec?.labels,
      tolerations: template.spec?.tolerations,
      backoffLimit: template.spec?.backoffLimit,
      priorityClass: template.spec?.priorityClass,
      preemptibility: template.spec?.preemptibility,
      terminateAfterPreemption: template.spec?.terminateAfterPreemption,
      autoDeletionTimeAfterCompletionSeconds: template.spec?.autoDeletionTimeAfterCompletionSeconds,
      terminationGracePeriodSeconds: template.spec?.terminationGracePeriodSeconds,
      completions: template.spec?.completions,
      parallelism: template.spec?.parallelism,
      connections: getConnectionsField(
        template.spec?.ports || [],
        template.spec?.exposedUrls || [],
        template.spec?.relatedUrls || [],
      ),
      compute: getComputeField(template.spec?.compute),
      storage: getStorageField(template.spec?.storage),
      security: getSecurityField(template.spec?.security),
    },
  };
}

function getHybridDistributedTemplateModel(
  template: DistributedTemplate,
  noMaster = false,
  basedOnExistingWorkload = false,
): IUISuperWorkloadCreationRequest {
  const uiDist: IUISuperWorkloadCreationRequest = {
    name: template.meta?.name || "",
    projectId: template.meta?.projectId?.toString() || "",
    scopeType: template.meta?.scopeType,
    scopeId: template.meta?.scopeId,
    departmentId: template.meta?.departmentId || undefined,
    clusterId: template.meta?.clusterId || "",
    description: template.meta?.description || "",
    masterSpecSameAsWorker: template.masterSpecSameAsWorker,
    spec: {
      command: template.spec?.command,
      args: template.spec?.args,
      image: template.spec?.image,
      imagePullPolicy: template.spec?.imagePullPolicy,
      workingDir: template.spec?.workingDir,
      createHomeDir: template.spec?.createHomeDir,
      nodeType: template.spec?.nodeType,
      nodePools: template.spec?.nodePools,
      podAffinity: template.spec?.podAffinity,
      tty: template.spec?.tty,
      stdin: template.spec?.stdin,
      environmentVariables: template.spec?.environmentVariables,
      annotations: template.spec?.annotations,
      labels: template.spec?.labels,
      tolerations: template.spec?.tolerations,
      terminateAfterPreemption: template.spec?.terminateAfterPreemption,
      autoDeletionTimeAfterCompletionSeconds: template.spec?.autoDeletionTimeAfterCompletionSeconds,
      terminationGracePeriodSeconds: template.spec?.terminationGracePeriodSeconds,
      backoffLimit: template.spec?.backoffLimit,
      numWorkers: template.spec?.numWorkers,
      segmentSize: template.spec?.segmentSize,
      cleanPodPolicy: template.spec?.cleanPodPolicy,
      connections: getConnectionsField(
        template.spec?.ports || [],
        template.spec?.exposedUrls || [],
        template.spec?.relatedUrls || [],
      ),
      compute: getComputeField(template.spec?.compute),
      storage: getStorageField(template.spec?.storage),
      security: getSecurityField(template.spec?.security),
      //  careful here
      distributedFramework: template.spec?.distributedFramework,
      sshAuthMountPath: template.spec?.sshAuthMountPath,
      mpiLauncherCreationPolicy: template.spec?.mpiLauncherCreationPolicy,
      minReplicas: template.spec?.minReplicas,
      maxReplicas: template.spec?.maxReplicas,
    },
  };

  if (noMaster) return uiDist;
  if (basedOnExistingWorkload && !template.master?.spec) return uiDist;

  return {
    ...uiDist,
    masterSpec: getMasterModelForUI(
      template.master?.spec || {},
      template.spec?.sshAuthMountPath || undefined,
      template.spec?.mpiLauncherCreationPolicy || undefined,
    ),
  };
}

function prepareSingleReplicaWorkloadTemplateFlatCreationRequest<
  T extends TrainingTemplateCreationRequest | WorkspaceTemplateCreationRequest,
>(model: IUISuperWorkloadCreationRequest, linkedAssetsIds?: TemplateAssetsIds): T {
  const storage = getStorageFieldFromUI(model.spec.storage || {});
  const compute = getComputeFieldFromUI(model.spec.compute || {});
  const security = { ...model.spec.security };

  const req = {
    meta: {
      name: model.name,
      scopeType: model.scopeType as ScopeType,
      scopeId: model.scopeId,
      description: model.description || undefined,
    },
    spec: {
      ...omit(model.spec, ["connections"]),
      image: model.spec.image || undefined,
      ..._connectionsToFlatRequest(model.spec.connections),
      storage,
      security,
      compute: linkedAssetsIds?.compute ? undefined : compute,
    },
    assets:
      linkedAssetsIds?.compute || linkedAssetsIds?.environment
        ? { environment: linkedAssetsIds.environment, compute: linkedAssetsIds.compute }
        : undefined,
  } as T;

  return clearAssetBoundFieldsFromTemplateModel(req);
}

function prepareSingleReplicaWorkloadTemplateFlatUpdateRequest<
  T extends TrainingTemplateUpdateRequest | WorkspaceTemplateUpdateRequest | InferenceTemplateUpdateRequest,
>(model: IUISuperWorkloadCreationRequest, linkedAssetsIds?: TemplateAssetsIds): T {
  const storage = getStorageFieldFromUI(model.spec.storage || {});
  const compute = getComputeFieldFromUI(model.spec.compute || {});
  const security = { ...model.spec.security };

  const requestRaw: T = {
    meta: {
      name: model.name,
      description: model.description || undefined,
    },
    spec: {
      ...omit(model.spec, ["connections"]),
      image: model.spec.image || undefined,
      ..._connectionsToFlatRequest(model.spec.connections),
      storage,
      security,
      compute,
    },
    assets:
      linkedAssetsIds?.compute || linkedAssetsIds?.environment
        ? { environment: linkedAssetsIds.environment, compute: linkedAssetsIds.compute }
        : undefined,
  } as T;

  return clearAssetBoundFieldsFromTemplateModel(requestRaw);
}

function prepareWorkspaceTemplateFlatRequest(
  model: IUISuperWorkloadCreationRequest,
  linkedAssetsIds?: TemplateAssetsIds,
): WorkspaceTemplateCreationRequest {
  return prepareSingleReplicaWorkloadTemplateFlatCreationRequest<WorkspaceTemplateCreationRequest>(
    model,
    linkedAssetsIds,
  );
}

function prepareWorkspaceTemplateFlatRequestToUpdate(
  model: IUISuperWorkloadCreationRequest,
  linkedAssetsIds?: TemplateAssetsIds,
): WorkspaceTemplateUpdateRequest {
  return prepareSingleReplicaWorkloadTemplateFlatUpdateRequest<WorkspaceTemplateUpdateRequest>(model, linkedAssetsIds);
}

function prepareTrainingTemplateFlatRequest(
  model: IUISuperWorkloadCreationRequest,
  linkedAssetsIds?: TemplateAssetsIds,
): TrainingTemplateCreationRequest {
  return prepareSingleReplicaWorkloadTemplateFlatCreationRequest<TrainingTemplateCreationRequest>(
    model,
    linkedAssetsIds,
  );
}

function prepareTrainingTemplateFlatRequestToUpdate(
  model: IUISuperWorkloadCreationRequest,
  linkedAssetsIds?: TemplateAssetsIds,
): TrainingTemplateUpdateRequest {
  return prepareSingleReplicaWorkloadTemplateFlatUpdateRequest<TrainingTemplateUpdateRequest>(model, linkedAssetsIds);
}

function prepareDistributedTemplateFlatRequestToUpdate(
  model: IUISuperWorkloadCreationRequest,
  linkedAssetsIds?: { worker: TemplateAssetsIds; master: TemplateAssetsIds },
): DistributedTemplateUpdateRequest {
  const storage = getStorageFieldFromUI(model.spec.storage || {});
  const compute = getComputeFieldFromUI(model.spec.compute || {});
  const security = { ...model.spec.security };

  const requestRaw: DistributedTemplateUpdateRequest = {
    meta: {
      name: model.name,
      description: model.description || undefined,
    },
    masterSpecSameAsWorker: model.masterSpecSameAsWorker,
    spec: {
      ...omit(model.spec, ["connections"]),
      image: model.spec.image || undefined,
      ..._connectionsToFlatRequest(model.spec.connections),
      storage,
      security,
      compute,
      restartPolicy: (model.spec.restartPolicy as RestartPolicy | null) || undefined,
    },
    assets:
      linkedAssetsIds?.worker?.compute || linkedAssetsIds?.worker?.environment
        ? { environment: linkedAssetsIds.worker.environment, compute: linkedAssetsIds.worker.compute }
        : undefined,
  };

  const clearedRequest = clearAssetBoundFieldsFromTemplateModel(requestRaw);

  if (!clearedRequest.masterSpecSameAsWorker && model.masterSpec) {
    const security = { ...model.masterSpec.security };

    const masterRaw = {
      spec: {
        ..._omitIrrelevantMasterSpecFields(model.masterSpec),
        image: model.masterSpec.image || undefined,
        ..._connectionsToFlatRequest(model.masterSpec.connections),
        storage: getStorageFieldFromUI(model.masterSpec.storage || {}),
        compute: _stripMpiMasterGpuFields(
          getComputeFieldFromUI(model.masterSpec.compute || {}),
          model.spec.distributedFramework,
        ),
        security,
        restartPolicy: (model.masterSpec.restartPolicy as RestartPolicy | null) || undefined,
      },
      assets:
        linkedAssetsIds?.master?.compute || linkedAssetsIds?.master?.environment
          ? { environment: linkedAssetsIds.master.environment, compute: linkedAssetsIds.master.compute }
          : undefined,
    };
    clearedRequest.master = clearAssetBoundFieldsFromTemplateModel(masterRaw);
  }

  return clearedRequest;
}

function prepareDistributedTemplateFlatRequest(
  model: IUISuperWorkloadCreationRequest,
  linkedAssetsIds?: { worker: TemplateAssetsIds; master: TemplateAssetsIds },
): DistributedTemplateCreationRequest {
  const storage = getStorageFieldFromUI(model.spec.storage || {});
  const compute = getComputeFieldFromUI(model.spec.compute || {});
  const security = { ...model.spec.security };

  const requestRaw: DistributedTemplateCreationRequest = {
    meta: {
      name: model.name,
      scopeType: model.scopeType as ScopeType,
      scopeId: model.scopeId,
      description: model.description || undefined,
    },
    masterSpecSameAsWorker: model.masterSpecSameAsWorker,
    spec: {
      ...omit(model.spec, ["connections"]),
      ..._connectionsToFlatRequest(model.spec.connections),
      image: model.spec.image || undefined,
      storage,
      security,
      compute,
      restartPolicy: (model.spec.restartPolicy as RestartPolicy | null) || undefined,
    },
    assets:
      linkedAssetsIds?.worker?.compute || linkedAssetsIds?.worker?.environment
        ? { environment: linkedAssetsIds.worker.environment, compute: linkedAssetsIds.worker.compute }
        : undefined,
  };

  const clearedRequest = clearAssetBoundFieldsFromTemplateModel(requestRaw);

  // if sameAsWorker is falsy AND masterSpec exists, we need prepare it as well
  if (!clearedRequest.masterSpecSameAsWorker && model.masterSpec) {
    const security = { ...model.masterSpec.security };

    // note: we need to remove sshAuthMountPath from the masterSpec since it's not supported in the API
    const masterRaw = {
      spec: {
        ..._omitIrrelevantMasterSpecFields(model.masterSpec),
        ..._connectionsToFlatRequest(model.masterSpec.connections),
        image: model.masterSpec.image || undefined,
        storage: getStorageFieldFromUI(model.masterSpec.storage || {}),
        compute: _stripMpiMasterGpuFields(
          getComputeFieldFromUI(model.masterSpec.compute || {}),
          model.spec.distributedFramework,
        ),
        security,
        restartPolicy: (model.masterSpec.restartPolicy as RestartPolicy | null) || undefined,
      } as MasterSpec,
      assets:
        linkedAssetsIds?.master?.compute || linkedAssetsIds?.master?.environment
          ? { environment: linkedAssetsIds.master.environment, compute: linkedAssetsIds.master.compute }
          : undefined,
    };

    clearedRequest.master = clearAssetBoundFieldsFromTemplateModel(masterRaw);
  }
  return clearedRequest;
}

function calculateTotalGpuRequest(options: {
  specCompute?: WorkspaceSpecSpecAllOfCompute | null;
  masterCompute?: WorkspaceSpecSpecAllOfCompute | null;
  numWorkers?: number | null;
  replicas?: { min?: number | null; max?: number | null };
}): number | undefined {
  const { specCompute, masterCompute, numWorkers, replicas } = options;

  if (!specCompute) {
    return;
  }

  const shouldCalculate = specCompute.gpuRequestType === GpuRequestType.Portion || !specCompute.gpuRequestType;

  if (!shouldCalculate) {
    return;
  }

  let totalGpuRequest = 0;

  // Calculate worker GPU request
  if (replicas) {
    // Inference case with replicas
    totalGpuRequest = computeResourceUtil.totalGpuResourcesNumeric(specCompute, undefined, replicas);
  } else {
    // Distributed training case with numWorkers
    totalGpuRequest = computeResourceUtil.totalGpuResourcesNumeric(specCompute, numWorkers || undefined);
  }

  // Add master GPU request if applicable (distributed training)
  if (masterCompute) {
    const masterShouldCalculate =
      masterCompute && (masterCompute.gpuRequestType === GpuRequestType.Portion || !masterCompute.gpuRequestType);

    if (masterShouldCalculate && masterCompute.gpuRequestType === GpuRequestType.Portion) {
      const masterGpuRequest = computeResourceUtil.totalGpuResourcesNumeric(masterCompute);
      totalGpuRequest += masterGpuRequest;
    }
  }

  return totalGpuRequest;
}

function _removeTemplateFieldsFromModel(model: IUISuperWorkloadCreationRequest): IUISuperWorkloadCreationRequest {
  delete model.scopeType;
  delete model.scopeId;
  delete model.departmentId;

  return model;
}

function getFirstPageNavigationTemproraryData(
  model: IUISuperWorkloadCreationRequest,
): Record<string, string | null | undefined> {
  return {
    name: model.name,
    description: model.description,
    clusterId: model.clusterId,
    scopeType: model.scopeType,
    scopeId: model.scopeId,
    distFramework: model.spec.distributedFramework,
    noMaster: model.masterSpec ? "false" : "true",
  };
}

function clearAssetBoundFieldsFromTemplateModel<
  T extends
    | TrainingTemplateCreationRequest
    | WorkspaceTemplateCreationRequest
    | DistributedTemplateCreationRequest
    | TemplateMasterSection
    | InferenceTemplateCreationRequest,
>(model: T): T {
  if (!model.assets?.environment && !model.assets?.compute) {
    return model;
  }

  if (model.assets?.compute && model.spec?.compute) {
    model.spec.compute = undefined;
  }

  if (model.assets?.environment && model.spec) {
    model.spec = omit(model.spec as SupersetSpec, envAssetFieldsKeys) as typeof model.spec;
    if (model.spec?.security) {
      model.spec.security = omit(model.spec.security as WorkspaceSpecSpecAllOfSecurity, envFieldsKeysInSecurity);
    }
  }

  return model;
}

// GPU fields are forbidden on the master replica (launcher) of an MPI workload and are
// rejected by the API. We strip them before submission to avoid a 400 error.
const MPI_MASTER_FORBIDDEN_GPU_FIELDS: (keyof WorkspaceSpecSpecAllOfCompute)[] = [
  "gpuDevicesRequest",
  "gpuRequestType",
  "gpuPortionRequest",
  "gpuPortionLimit",
  "gpuMemoryRequest",
  "gpuMemoryLimit",
];

function _stripMpiMasterGpuFields(
  compute: WorkspaceSpecSpecAllOfCompute,
  distributedFramework: DistributedFramework | null | undefined,
): WorkspaceSpecSpecAllOfCompute {
  if (distributedFramework !== DistributedFramework.Mpi) return compute;
  return omit(compute, MPI_MASTER_FORBIDDEN_GPU_FIELDS);
}

function _omitIrrelevantMasterSpecFields(masterSpec: IUISuperWorkloadSpec): IUISuperWorkloadSpec {
  // these fields should only be set on the worker spec as they apply to both worker and master
  // connections are the exception as they are not used by the UI only and are not supported by the API
  return omit(masterSpec, [
    "cleanPodPolicy",
    "sshAuthMountPath",
    "mpiLauncherCreationPolicy",
    "backoffLimit",
    "autoDeletionTimeAfterCompletionSeconds",
    "terminationGracePeriodSeconds",
    "priorityClass",
    "preemptibility",
    "segmentSize",
    "connections",
  ]);
}

const envAssetFieldsKeys: (keyof SupersetSpec)[] = [
  "command",
  "args",
  "environmentVariables",
  "image",
  "imagePullPolicy",
  "workingDir",
  "createHomeDir",
  "probes",
  "tty",
  "stdin",
  "relatedUrls",
  "ports",
  "exposedUrls",
  "servingPort",
];

const envFieldsKeysInSecurity: (keyof WorkspaceSpecSpecAllOfSecurity)[] = [
  "capabilities",
  "allowPrivilegeEscalation",
  "hostIpc",
  "hostNetwork",
  "seccompProfileType",
  "runAsNonRoot",
  "readOnlyRootFilesystem",
  "runAsUid",
  "runAsGid",
  "supplementalGroups",
  "uidGidSource",
];
