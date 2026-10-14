/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type {
  WorkloadDetailed,
  Connection,
  EnvironmentVariable,
  Label,
  MasterSpec,
  DistributedSpecSpec,
  TrainingSpecSpec,
  InferenceSpecSpec,
  DistributedWorkload,
  PodNodeAffinity,
  ServingPort,
  WorkloadAllocatedResources,
} from "@/swagger-models/workloads-client";
import type { TItemDetail } from "@/models/workload-creation-drawer-model";
import type {
  IDetailsPanelItem,
  IDetailsPanelItemLink,
  IDetailsPanelNetworkConnection,
  IDetailsPanelSectionDef,
  IDetailsPanelStorageVolume,
  IDetailsPanelStorageVolumeField,
  TDetailsPanelItemGroup,
} from "@/models/details-panel.model";
import type {
  ICliCommandSectionData,
  IComputeSectionData,
  IEnvironmentSectionData,
  IGeneralSectionData,
  INetworkSectionData,
  IStorageSectionData,
} from "@/models/workload-details.model";
import {
  matchAnyItem,
  filterDefined,
  objectMatchesTerm,
  topologyMatches,
  defineSection,
  buildSections,
  type SectionBuilder,
} from "@/utils/details-panel.utils";
import type { CrudWorkload } from "@/services/cluster/workload.service/workload.service";
import type { ICpuCore, ICpuMemory } from "@/models/compute-resource.model";
import { InternalConnectionType } from "@/swagger-models/assets-service-client";
import { dateUtil } from "@/utils/date.util";
import { DATE_FORMATS } from "@/constants/date-formats";
import { LINE_BREAK } from "@/utils/workload-hybrid.util/workload-hybrid.util";
import { getMemoryUnitsForUI, getCpuCoresForUI } from "@/utils/table-format.util";
import { formatK8sResourceMemory } from "@/utils/format.util/format.util";
import { EDistributedReplicaType } from "@/models/workload.model";
import { workloadHybridUtil } from "@/utils/workload-hybrid.util/workload-hybrid.util";
import { PROJECT_ROUTE_NAMES } from "@/router/project.routes/project.routes.names";
import { DEPARTMENT_ROUTE_NAMES } from "@/router/department.routes/department.routes.names";
import { CLUSTER_ROUTE_NAMES } from "@/router/cluster.routes/cluster.routes.names";
import { DepartmentFilterSortFields, ProjectFilterSortFields } from "@/swagger-models/org-unit-service-client";
import { EFilterOperator } from "@/models/filter.model";
import { filterUtil } from "@/utils/filter.util/filter.util";
import { workloadUtil } from "@/utils/workload.util/workload.util";

export const DETAILS_PANEL_SECTIONS = {
  GENERAL: { id: "general", title: "General information" },
  ENVIRONMENT: { id: "environment", title: "Environment" },
  NETWORK: { id: "network", title: "Network" },
  COMPUTE: { id: "compute", title: "Compute" },
  STORAGE: { id: "storage", title: "Data & storage" },
  CLI_COMMAND: { id: "cli-command", title: "CLI Command" },
} as const;

export type WorkloadSpec = MasterSpec | DistributedSpecSpec | TrainingSpecSpec | InferenceSpecSpec;
type WorkloadSpecWithServingPort = WorkloadSpec & { servingPort?: ServingPort | null };
export type ResourceRequestLimit = ICpuCore | ICpuMemory;
interface SubSection {
  subtitle: string;
  content: string[];
}
interface FieldConfig {
  label: string;
  value: string | number | boolean | undefined | null;
  formatter?: (val: string | number | boolean | null | undefined) => string | number;
  suffix?: string;
}

export const workloadDetailsUtil = {
  parseGeneralInfo,
  parseEnvironmentDetails,
  parseNetworkDetails,
  parseComputeDetails,
  parseStorageDetails,
  parseApplicationParameters,
  formatBoolean,
  formatDate,
  formatEnvVars,
  formatLabels,
  formatConnectionContent,
  formatConnectionType,
  formatResourceRequestLimit,
  getSpecSection,
  buildGeneralInfoGroups,
  itemDetailsToSearchableItems,
  buildSearchableSections,
  buildAllocatedResourceItems,
  pair,
};

function _buildFieldsContent(fields: FieldConfig[]): string[] {
  const content: string[] = [];
  fields.forEach(({ label, value, formatter, suffix }) => {
    if (value !== null && value !== undefined && value !== "") {
      const formattedValue = formatter ? formatter(value) : value;
      const suffixStr = suffix ? suffix : "";
      content.push(`${label}: ${formattedValue}${suffixStr}`);
    }
  });
  return content;
}

function _buildItemsWithSeparator<T>(items: T[], builder: (item: T) => string[]): string[] {
  const content: string[] = [];
  items.forEach((item, index) => {
    content.push(...builder(item));
    if (index < items.length - 1) {
      content.push(LINE_BREAK);
    }
  });
  return content;
}

function toItemDetail(content: string[]): TItemDetail[] {
  return content.length ? [{ content }] : [];
}

function parseGeneralInfo(workloadDetailed: WorkloadDetailed, clusterName: string): TItemDetail[] {
  const content = _buildFieldsContent([
    { label: "Cluster", value: clusterName },
    { label: "Department", value: workloadDetailed.departmentName || "" },
    { label: "Project", value: workloadDetailed.projectName || "" },
    { label: "Category", value: workloadDetailed.category || "" },
    { label: "Type", value: workloadDetailed.type || "" },
    { label: "Priority", value: workloadDetailed.priorityClassName || "" },
    { label: "Preemptible", value: workloadDetailed.preemptible, formatter: formatBoolean },
    { label: "Creation time", value: workloadDetailed.createdAt, formatter: formatDate },
    { label: "Created by", value: workloadDetailed.submittedBy || "" },
  ]);

  return [{ content }];
}

function parseEnvironmentDetails(workloadDetailed: WorkloadDetailed, workloadSpec?: WorkloadSpec): TItemDetail[] {
  const sections: SubSection[] = [];

  // Image URL
  const imageFromCrud = workloadSpec?.image;
  const imageContent = imageFromCrud ? [imageFromCrud] : workloadDetailed.images || [];
  _addSubSection(sections, "Image URL", imageContent);

  const command = workloadSpec?.command ?? workloadDetailed.command ?? "";
  const args = workloadSpec?.args ?? workloadDetailed.arguments ?? "";
  if (command) _addSubSection(sections, "Command", [command]);
  if (args) _addSubSection(sections, "Arguments", [args]);

  // Environment variables — emitted separately by buildEnvironmentVariables as a structured
  // payload, so it's no longer added to the flat key/value rows here.

  // Labels are emitted separately by buildLabels as a structured list of `key:value` strings
  // rendered as code-styled chips inside the Environment card.

  // Application parameters are emitted separately by buildApplicationParameters as a structured
  // list, rendered as a divider-separated sub-section in the Environment card.

  return sections.map(({ subtitle, content }) => ({ subtitle, content }));
}

/**
 * Returns environment variables as a flat list of "name:value" strings, rendered as code-styled
 * chips in the Environment card (same design as Labels). For sourced env vars (secret /
 * configMap / user credential / pod field ref), the value column shows the source path.
 */
function buildEnvironmentVariables(workloadSpec?: WorkloadSpec, workloadDetailed?: WorkloadDetailed): string[] {
  // Spec carries the rich array shape. The mock fixture also stuffs the array into the
  // detailed top-level field — read that as a fallback so the mock works without a spec.
  const raw =
    workloadSpec?.environmentVariables ??
    (workloadDetailed?.environmentVariables as unknown as EnvironmentVariable[] | undefined);
  if (!Array.isArray(raw) || raw.length === 0) return [];

  return raw.map((envVar) => `${envVar.name}:${getEnvVarValueDisplay(envVar)}`);
}

function getEnvVarValueDisplay(envVar: EnvironmentVariable): string {
  if (envVar.value) return envVar.value;
  if (envVar.secret?.name) return `${envVar.secret.key} (${envVar.secret.name})`;
  if (envVar.configMap?.name) return `${envVar.configMap.key} (${envVar.configMap.name})`;
  if (envVar.userCredential?.name) return `${envVar.userCredential.key} (${envVar.userCredential.name})`;
  if (envVar.podFieldRef?.path) return envVar.podFieldRef.path;
  return "";
}

function _addSubSection(sections: SubSection[], subtitle: string, content: string[]): void {
  if (content.length) {
    sections.push({ subtitle, content });
  }
}

function getSpecSection(
  crudWorkload?: CrudWorkload,
  selectedReplicaType?: EDistributedReplicaType,
): WorkloadSpec | undefined {
  if (!crudWorkload) {
    return undefined;
  }

  // Distributed workload (master + worker): independent pod templates at the top level.
  // Each replica's pod-level fields come from its own spec only — no inheritance.
  if ("masterSpec" in crudWorkload && crudWorkload.masterSpec) {
    const distributedWorkload = crudWorkload as DistributedWorkload;

    if (selectedReplicaType === EDistributedReplicaType.Master) {
      return distributedWorkload.masterSpec as WorkloadSpec;
    }

    if (selectedReplicaType === EDistributedReplicaType.Worker && distributedWorkload.spec) {
      return distributedWorkload.spec as WorkloadSpec;
    }
  }

  // DistributedInference (leader + worker): nested inside spec, NOT at the top level.
  // Same no-inheritance rule applies — pick the replica's nested pod template.
  const nestedSpec = (crudWorkload as { spec?: { leader?: unknown; worker?: unknown } }).spec;
  if (nestedSpec && (nestedSpec.leader || nestedSpec.worker)) {
    if (selectedReplicaType === EDistributedReplicaType.Leader && nestedSpec.leader) {
      return nestedSpec.leader as WorkloadSpec;
    }
    if (selectedReplicaType === EDistributedReplicaType.Worker && nestedSpec.worker) {
      return nestedSpec.worker as WorkloadSpec;
    }
  }

  // Check if the workload has a spec property (non-distributed with spec)
  if ("spec" in crudWorkload && crudWorkload.spec) {
    return crudWorkload.spec as WorkloadSpec;
  }

  // For non-distributed types, treat the workload itself as a WorkloadSpec
  return crudWorkload as WorkloadSpec;
}

interface ConnectionWithAccess extends Connection {
  exposedUrlAuthorizationType?: string | null;
  exposedUrlAuthorizedUsers?: string[] | null;
  exposedUrlAuthorizedGroups?: string[] | null;
}

interface SpecPortLike {
  container?: number | null;
  toolName?: string | null;
  toolType?: string | null;
}

function connectionKey(port: number | null | undefined, toolType: string | null | undefined): string {
  return `${port ?? ""}-${toolType ?? ""}`;
}

function getEnrichedConnections(
  workloadDetailed: WorkloadDetailed,
  workloadSpec?: WorkloadSpec,
): ConnectionWithAccess[] {
  const allEnvConnections = workloadDetailed.environments?.flatMap((env) => env.connections || []) || [];
  const externalConnections = workloadDetailed.externalConnections || [];
  const exposedUrls = workloadSpec?.exposedUrls ?? [];
  const specPorts = ((workloadSpec as { ports?: SpecPortLike[] | null } | undefined)?.ports ?? []) as SpecPortLike[];
  const servingPort = (workloadSpec as WorkloadSpecWithServingPort)?.servingPort;

  // Clusters may report the ServingPort connection without a URL; fall back to the only
  // workload URL that no other connection claims. A serving port that is not exposed
  // externally has no URL to show.
  const servingPortExposure = servingPort as { exposeExternally?: boolean | null } | null | undefined;
  const isServingPortInternal = servingPortExposure?.exposeExternally === false || !!servingPort?.clusterLocalAccessOnly;
  const servingUrl = isServingPortInternal
    ? undefined
    : workloadUtil.getUnclaimedServingUrl(
        workloadDetailed.urls,
        allEnvConnections,
        exposedUrls.map((exposed) => exposed.url),
      );
  const withServingUrl = (conn: ConnectionWithAccess): ConnectionWithAccess =>
    conn.connectionType === InternalConnectionType.ServingPort && !conn.url && servingUrl
      ? { ...conn, url: servingUrl }
      : conn;

  // When the spec declares no networking entries at all, fall back to the runtime list
  // (covers completed workloads with no spec, externally-wired connections, etc.).
  if (!exposedUrls.length && !specPorts.length && !servingPort) {
    return [...allEnvConnections.map(withServingUrl), ...externalConnections];
  }

  // Build the list from the SPEC so switching Master / Worker reflects each spec. Each spec
  // entry (exposedUrl / port) is one configured access mode and gets one row — when the same
  // tool appears in both arrays (e.g. Jupyter as ExternalUrl + NodePort), both rows are kept.
  const result: ConnectionWithAccess[] = [];

  for (const exposed of exposedUrls) {
    const key = connectionKey(exposed.container, exposed.toolType);
    // Match the ExternalUrl runtime entry for this tool (the one that carries the public URL).
    const matching = allEnvConnections.find(
      (conn) =>
        connectionKey(conn.containerPort, conn.toolType) === key &&
        conn.connectionType === InternalConnectionType.ExternalUrl,
    );
    if (matching) {
      result.push({
        ...matching,
        url: matching.url || exposed.url || undefined,
        exposedUrlAuthorizationType: exposed.authorizationType,
        exposedUrlAuthorizedUsers: exposed.authorizedUsers,
        exposedUrlAuthorizedGroups: exposed.authorizedGroups,
      });
    } else {
      // Spec declares the URL but no runtime data yet — synthesize from the spec.
      result.push({
        name: exposed.toolName || exposed.toolType || `port-${exposed.container ?? ""}`,
        toolType: exposed.toolType ?? "",
        connectionType: InternalConnectionType.ExternalUrl,
        containerPort: exposed.container ?? undefined,
        url: exposed.url ?? undefined,
        authorizationType: exposed.authorizationType ?? undefined,
        authorizedUsers: exposed.authorizedUsers ?? undefined,
        authorizedGroups: exposed.authorizedGroups ?? undefined,
      } as ConnectionWithAccess);
    }
  }

  for (const port of specPorts) {
    const key = connectionKey(port.container, port.toolType);
    const matching = allEnvConnections.find(
      (conn) =>
        connectionKey(conn.containerPort, conn.toolType) === key &&
        conn.connectionType === InternalConnectionType.NodePort,
    );
    if (matching) {
      result.push({ ...matching });
    } else {
      // No runtime entry yet (workload not running) — synthesize from the spec port.
      result.push({
        name: port.toolName || port.toolType || `port-${port.container ?? ""}`,
        toolType: port.toolType ?? "",
        connectionType: InternalConnectionType.NodePort,
        containerPort: port.container ?? undefined,
      } as ConnectionWithAccess);
    }
  }

  if (servingPort) {
    // Prefer the ServingPort connection: a NodePort on the same container port must not take its place.
    const matching =
      allEnvConnections.find((conn) => conn.connectionType === InternalConnectionType.ServingPort) ??
      allEnvConnections.find((conn) => conn.containerPort === servingPort.container);
    if (matching) {
      result.push({
        ...withServingUrl(matching),
        exposedUrlAuthorizationType: servingPort.authorizationType,
        exposedUrlAuthorizedUsers: servingPort.authorizedUsers,
        exposedUrlAuthorizedGroups: servingPort.authorizedGroups,
      });
    } else {
      result.push({
        name: "ServingPort",
        toolType: "custom",
        connectionType: "ServingPort",
        containerPort: servingPort.container ?? undefined,
        url: servingUrl,
        authorizationType: servingPort.authorizationType ?? undefined,
        authorizedUsers: servingPort.authorizedUsers ?? undefined,
        authorizedGroups: servingPort.authorizedGroups ?? undefined,
      } as ConnectionWithAccess);
    }
  }

  // External connections (e.g. Comet, Weights & Biases) are not tied to a spec — append them
  // after the spec-driven connections so they show in both Master and Worker views.
  return [...result, ...externalConnections];
}

function buildConnectionFields(conn: ConnectionWithAccess): string[] {
  return _buildFieldsContent([
    { label: "Name", value: conn.name },
    { label: "Tool", value: conn.toolType },
    { label: "Connection type", value: conn.connectionType, formatter: formatConnectionType },
    { label: "Access", value: getAccessInfo(conn) },
    { label: "Port", value: conn.containerPort },
    { label: "URL", value: conn.url },
  ]);
}

function parseNetworkDetails(workloadDetailed: WorkloadDetailed, workloadSpec?: WorkloadSpec): TItemDetail[] {
  const connectionsToDisplay = getEnrichedConnections(workloadDetailed, workloadSpec);
  if (connectionsToDisplay.length === 0) return [];
  const connectionContent = _buildItemsWithSeparator(connectionsToDisplay, buildConnectionFields);
  return [{ subtitle: "", content: connectionContent }];
}

function isCodeFieldLabel(label: string): boolean {
  const normalized = label.toLowerCase();
  return /(^|\s|-)path$/.test(normalized) || normalized.endsWith("mountpath");
}

/**
 * Returns storage volumes in a view-friendly shape for the new Details panel. Each volume
 * becomes a labelled group (the name from the parser's subtitle) with key/value rows; any
 * label ending in "path" or "mountpath" (e.g. "Container path", "Host path", "NFS path",
 * "MountPath") is flagged for code-style rendering.
 */
function buildStorageVolumes(workloadSpec?: WorkloadSpec): IDetailsPanelStorageVolume[] {
  const details = parseStorageDetails(workloadSpec);
  return details.map((detail) => {
    const fields: IDetailsPanelStorageVolumeField[] = [];
    for (const line of detail.content) {
      if (line === LINE_BREAK) continue;
      const colonIdx = line.indexOf(": ");
      if (colonIdx > 0) {
        const label = line.slice(0, colonIdx);
        const value = line.slice(colonIdx + 2);
        fields.push({ label, value, isCode: isCodeFieldLabel(label) });
      } else if (line.length > 0) {
        fields.push({ label: "", value: line });
      }
    }
    return { name: detail.subtitle ?? "", fields };
  });
}

/**
 * Returns network connections in a view-friendly shape for the new Details panel.
 * Each connection becomes a labelled group (Tool / Connection type / Access / Port / URL)
 * with the connection name as the group header.
 */
function buildNetworkConnections(
  workloadDetailed: WorkloadDetailed,
  workloadSpec?: WorkloadSpec,
): IDetailsPanelNetworkConnection[] {
  return getEnrichedConnections(workloadDetailed, workloadSpec).map((conn) => ({
    name: conn.name ?? "",
    tool: conn.toolType ?? undefined,
    connectionType: formatConnectionType(conn.connectionType) || undefined,
    access: getAccessInfo(conn),
    port: conn.containerPort ?? undefined,
    // URL fallback chain: composed Connection.url → spec-defined exposedUrl.url
    // (handled in getEnrichedConnections) → cluster-synced sub-domain or path stored
    // on the Connection itself.
    url: conn.url || conn.exposedUrlSubDomain || conn.exposedUrlPath || undefined,
  }));
}

function getAccessInfo(conn: ConnectionWithAccess): string {
  // Prefer authorization data from exposedUrl (CRUD spec) over connection data (detailed API)
  const authorizationType = conn.exposedUrlAuthorizationType || conn.authorizationType;
  const authorizedUsers = conn.exposedUrlAuthorizedUsers || conn.authorizedUsers;
  const authorizedGroups = conn.exposedUrlAuthorizedGroups || conn.authorizedGroups;

  if (
    authorizationType === "authorizedUsersOrGroups" ||
    authorizationType === "authorizedUsers" ||
    authorizationType === "authorizedGroups"
  ) {
    const parts: string[] = [];

    if (authorizedUsers?.length) {
      parts.push(`Authorized users - ${authorizedUsers.join(", ")}`);
    }
    if (authorizedGroups?.length) {
      parts.push(`Authorized groups - ${authorizedGroups.join(", ")}`);
    }

    if (parts.length) {
      return parts.join("; ");
    }
  }

  const accessMap: Record<string, string> = {
    authenticatedUsers: "All authenticated users",
    public: "Public",
  };

  return accessMap[authorizationType || ""] || "All users";
}

function parseComputeDetails(workloadDetailed: WorkloadDetailed, workloadCrudSpec?: CrudWorkload): TItemDetail[] {
  const content: string[] = [];

  // Node pools
  if (workloadDetailed.requestedNodePools?.length) {
    content.push(`Node pools: ${workloadDetailed.requestedNodePools.join(", ")}`);
  }

  // Resource configurations
  if (workloadDetailed.workloadRequestedResources) {
    const { gpu, gpuMemory, cpu, cpuMemory } = workloadDetailed.workloadRequestedResources;

    const resourceConfigs = [
      { label: "GPU compute", value: gpu, suffix: "G" },
      { label: "GPU memory", value: gpuMemory },
      { label: "CPU compute", value: cpu, suffix: "cores" },
      { label: "CPU memory", value: cpuMemory },
    ];

    resourceConfigs.forEach(({ label, value, suffix }) => {
      const parts = formatResourceRequestLimit(value, suffix);
      if (parts) content.push(`${label}: ${parts}`);
    });
  }

  // Allocated resources
  for (const item of buildAllocatedResourceItems(workloadDetailed.allocatedResources)) {
    content.push(`${item.label}: ${item.value}`);
  }

  // Master-specific resources for distributed workloads
  if (workloadCrudSpec && "masterSpec" in workloadCrudSpec) {
    const distributedWorkload = workloadCrudSpec as DistributedWorkload;
    const masterCompute = distributedWorkload.masterSpec?.compute;

    if (masterCompute) {
      const masterCpu = {
        request: masterCompute.cpuCoreRequest ?? undefined,
        limit: masterCompute.cpuCoreLimit ?? undefined,
      };
      const masterCpuMemory = {
        request: masterCompute.cpuMemoryRequest ?? undefined,
        limit: masterCompute.cpuMemoryLimit ?? undefined,
      };

      const masterCpuParts = formatResourceRequestLimit(masterCpu, "cores");
      if (masterCpuParts) content.push(`Master CPU compute: ${masterCpuParts}`);

      const masterMemoryParts = formatResourceRequestLimit(masterCpuMemory);
      if (masterMemoryParts) content.push(`Master CPU memory: ${masterMemoryParts}`);
    }
  }

  // Pod configurations
  if (workloadDetailed.requestedPods) {
    const { number, min, max } = workloadDetailed.requestedPods;
    content.push(
      ..._buildFieldsContent([
        { label: "Number of pods", value: number },
        { label: "Min pods", value: min },
        { label: "Max pods", value: max },
      ]),
    );
  }

  return toItemDetail(content);
}

function parseStorageDetails(workloadSpec?: WorkloadSpec): TItemDetail[] {
  if (!workloadSpec?.storage) return [];

  return workloadHybridUtil.processWorkloadStorage(workloadSpec.storage);
}

interface AdditionalFields {
  latestCreatedRevisionName?: string;
}

function parseApplicationParametersContent(workloadDetailed: WorkloadDetailed, workloadSpec?: WorkloadSpec): string[] {
  const content: string[] = [];

  // Get autoscaling from spec
  const spec = workloadSpec as Record<string, unknown> | undefined;
  const autoscaling = spec?.autoscaling as Record<string, unknown> | undefined;

  // Latest created revision name from additionalFields
  const additionalFields = workloadDetailed.additionalFields as AdditionalFields | undefined;
  if (additionalFields?.latestCreatedRevisionName) {
    content.push(`Last created revision: ${additionalFields.latestCreatedRevisionName}`);
  }

  // Autoscaling configuration from CRUD spec
  if (autoscaling) {
    content.push(
      ..._buildFieldsContent([
        { label: "Minimum replicas", value: autoscaling.minReplicas as number },
        { label: "Maximum replicas", value: autoscaling.maxReplicas as number },
        { label: "Initial replicas", value: autoscaling.initialReplicas as number },
        { label: "Metric threshold percentage", value: autoscaling.metricThresholdPercentage as number },
        { label: "Metric", value: autoscaling.metric as string },
        { label: "Metric threshold", value: autoscaling.metricThreshold as number },
      ]),
    );
  }

  return content;
}

function parseApplicationParameters(workloadDetailed: WorkloadDetailed): TItemDetail[] {
  const content = parseApplicationParametersContent(workloadDetailed);
  return content.length ? [{ content }] : [];
}

/**
 * Returns application parameters as a flat key/value list, used as a divider-separated
 * sub-section inside the Environment card per the redesign.
 */
/**
 * Returns workload labels as a flat list of "name:value" strings, rendered as code-styled
 * chips in the Environment card.
 */
function buildLabels(workloadSpec?: WorkloadSpec, workloadDetailed?: WorkloadDetailed): string[] {
  const crudLabels =
    workloadSpec?.labels ?? (workloadDetailed as unknown as { labels?: Label[] } | undefined)?.labels ?? undefined;
  if (!crudLabels?.length) return [];
  return crudLabels.map((label) => `${label.name}:${label.value}`);
}

/**
 * Returns workload annotations as a flat list of "name:value" strings, rendered as code-styled
 * chips in the Environment card (same design as Labels).
 */
function buildAnnotations(workloadSpec?: WorkloadSpec): string[] {
  const annotations = (
    workloadSpec as { annotations?: Array<{ name?: string | null; value?: string | null }> } | undefined
  )?.annotations;
  if (!annotations?.length) return [];
  return annotations.map((a) => `${a.name}:${a.value}`);
}

function buildApplicationParameters(
  workloadDetailed: WorkloadDetailed,
  workloadSpec?: WorkloadSpec,
): IDetailsPanelItem[] {
  const items: IDetailsPanelItem[] = [];
  for (const line of parseApplicationParametersContent(workloadDetailed, workloadSpec)) {
    const idx = line.indexOf(": ");
    if (idx <= 0) continue;
    items.push({ label: line.slice(0, idx), value: line.slice(idx + 2) });
  }
  return items;
}

function formatResourceRequestLimit(resource?: ResourceRequestLimit | null, suffix?: string): string | null {
  if (!resource) return null;

  const parts: string[] = [];
  const { request, limit } = resource;

  if (request !== null && request !== undefined) {
    const formattedRequest = formatResourceValue(request, suffix);
    parts.push(`Request ${formattedRequest}`);
  }
  if (limit !== null && limit !== undefined) {
    const formattedLimit = formatResourceValue(limit, suffix);
    parts.push(`Limit ${formattedLimit}`);
  }

  return parts.length ? parts.join("  ") : null;
}

function formatResourceValue(value: number | string, suffix?: string): string {
  if (typeof value === "string") {
    // Memory values like "1G", "500M" - use existing utility
    return getMemoryUnitsForUI(value);
  }

  if (suffix === "cores") {
    // CPU cores - use shared utility for cores/millicores formatting
    return getCpuCoresForUI(value);
  }

  if (suffix === "G") {
    // GPU values
    return `${value}G`;
  }

  // Default: just the value
  return suffix ? `${value} ${suffix}` : value.toString();
}

/**
 * Shared by the workload-level and element-level Compute sections — both surface the same four
 * allocation fields from a `WorkloadAllocatedResources`-shaped object.
 */
function buildAllocatedResourceItems(allocated?: WorkloadAllocatedResources | null): IDetailsPanelItem[] {
  const items: IDetailsPanelItem[] = [];

  if (allocated?.gpu !== null && allocated?.gpu !== undefined) {
    items.push({ label: "GPU compute allocation", value: String(allocated.gpu) });
  }
  if (allocated?.gpuMemory) {
    items.push({ label: "GPU memory allocation", value: formatK8sResourceMemory(allocated.gpuMemory) });
  }
  if (allocated?.cpu !== null && allocated?.cpu !== undefined) {
    items.push({ label: "CPU compute allocation", value: getCpuCoresForUI(allocated.cpu) });
  }
  if (allocated?.cpuMemory) {
    items.push({ label: "CPU memory allocation", value: formatK8sResourceMemory(allocated.cpuMemory) });
  }

  return items;
}

function formatBoolean(value: string | number | boolean | null | undefined): string {
  if (value === null || value === undefined) return "";
  if (typeof value === "boolean") return value ? "Yes" : "No";
  return String(value);
}

function formatDate(date: string | number | boolean | null | undefined): string {
  if (!date || typeof date !== "string") return "";
  return dateUtil.dateFormat(new Date(date), DATE_FORMATS.STANDARD_DATE);
}

function formatEnvVars(envVars: Record<string, string>): string[] {
  return Object.entries(envVars).map(([key, value]) => `${key}: ${value}`);
}

function formatLabels(labels: Label[] | Record<string, string> | string[] | undefined): string[] {
  if (!labels) return [];

  if (Array.isArray(labels)) {
    return labels.map((label) => {
      if (typeof label === "string") {
        // If already formatted or just a string, wrap in brackets
        return label.startsWith("[") ? label : `[${label}]`;
      }
      if (typeof label === "object" && label !== null) {
        // Handle label objects with name and value properties
        const labelObj = label as Label;
        if (labelObj.name && labelObj.value) {
          return `[${labelObj.name}:${labelObj.value}]`;
        }
      }
      return `[${JSON.stringify(label)}]`;
    });
  }
  if (typeof labels === "object" && labels !== null) {
    return Object.entries(labels).map(([k, v]) => `[${k}:${v}]`);
  }
  return [];
}

function formatConnectionContent(connection: Connection): string[] {
  const content = _buildFieldsContent([
    { label: "Tool", value: connection.toolType },
    { label: "Connection type", value: connection.connectionType, formatter: formatConnectionType },
    { label: "Container port", value: connection.containerPort },
    { label: "URL", value: connection.url },
    { label: "Access", value: connection.authorizationType },
  ]);

  return content.length ? content : ["No connection details available"];
}

function formatConnectionType(type: string | number | boolean | null | undefined): string {
  if (!type || typeof type !== "string") return "";

  const connectionTypeMap: Record<string, string> = {
    [InternalConnectionType.ExternalUrl]: "External URL",
    [InternalConnectionType.NodePort]: "Node port",
    [InternalConnectionType.LoadBalancer]: "Load Balancer",
    [InternalConnectionType.ServingPort]: "Serving endpoint",
  };

  return connectionTypeMap[type] || type;
}

interface IBuildGeneralInfoOptions {
  projectLink?: IDetailsPanelItemLink;
  clusterLink?: IDetailsPanelItemLink;
  departmentLink?: IDetailsPanelItemLink;
}

/**
 * Builds a router-link to the Projects grid pre-filtered by the workload's project name.
 * Returns undefined when the workload doesn't have a project name.
 */
function buildProjectLink(workloadDetailed: WorkloadDetailed): IDetailsPanelItemLink | undefined {
  const projectName = workloadDetailed.projectName;
  if (!projectName) return undefined;
  return {
    text: projectName,
    to: {
      name: PROJECT_ROUTE_NAMES.PROJECT_INDEX,
      query: {
        advancedFilters: `[${filterUtil.getFreeTextFilterQuery(ProjectFilterSortFields.Name, projectName, "Project", EFilterOperator.Equals)}]`,
      },
    },
  };
}

/**
 * Builds a router-link to the Clusters grid pre-filtered by the cluster name.
 * Returns undefined when no cluster name is available.
 */
function buildClusterLink(clusterName: string): IDetailsPanelItemLink | undefined {
  if (!clusterName) return undefined;
  return {
    text: clusterName,
    to: {
      name: CLUSTER_ROUTE_NAMES.CLUSTER_INDEX,
      query: {
        advancedFilters: `[${filterUtil.getFreeTextFilterQuery("name", clusterName, "Cluster", EFilterOperator.Equals)}]`,
      },
    },
  };
}

/**
 * Builds a router-link to the Departments grid pre-filtered by the workload's department name.
 * Returns undefined when the workload doesn't have a department name.
 */
function buildDepartmentLink(workloadDetailed: WorkloadDetailed): IDetailsPanelItemLink | undefined {
  const departmentName = workloadDetailed.departmentName;
  if (!departmentName) return undefined;
  return {
    text: departmentName,
    to: {
      name: DEPARTMENT_ROUTE_NAMES.DEPARTMENT_INDEX,
      query: {
        advancedFilters: `[${filterUtil.getFreeTextFilterQuery(DepartmentFilterSortFields.Name, departmentName, "Department", EFilterOperator.Equals)}]`,
      },
    },
  };
}

/**
 * Builds the General-information groups in the order specified by the redesign:
 *  Group 1: Type / Category / Priority / Preemptible
 *  Group 2: Project / Cluster / Department
 *  Group 3: Creation time / Created by
 * Each group is dropped from the output if empty.
 */
function buildGeneralInfoGroups(
  workloadDetailed: WorkloadDetailed,
  clusterName: string,
  options: IBuildGeneralInfoOptions = {},
): TDetailsPanelItemGroup[] {
  const typeGroup = filterDefined<IDetailsPanelItem>([
    pair("Type", workloadDetailed.type),
    pair("Category", workloadDetailed.category),
    pair("Priority", workloadDetailed.priorityClassName),
    workloadDetailed.preemptible !== null && workloadDetailed.preemptible !== undefined
      ? { label: "Preemptible", value: formatBoolean(workloadDetailed.preemptible) }
      : null,
  ]);

  const projectItem: IDetailsPanelItem | null = workloadDetailed.projectName
    ? {
        label: "Project",
        value: options.projectLink?.text || workloadDetailed.projectName,
        link: options.projectLink,
      }
    : null;

  const clusterItem: IDetailsPanelItem | null = clusterName
    ? { label: "Cluster", value: options.clusterLink?.text || clusterName, link: options.clusterLink }
    : null;

  const departmentItem: IDetailsPanelItem | null = workloadDetailed.departmentName
    ? {
        label: "Department",
        value: options.departmentLink?.text || workloadDetailed.departmentName,
        link: options.departmentLink,
      }
    : null;

  const orgGroup = filterDefined<IDetailsPanelItem>([projectItem, clusterItem, departmentItem]);

  const auditGroup = filterDefined<IDetailsPanelItem>([
    workloadDetailed.createdAt ? { label: "Creation time", value: formatDate(workloadDetailed.createdAt) } : null,
    pair("Created by", workloadDetailed.submittedBy),
  ]);

  return [typeGroup, orgGroup, auditGroup].filter((group) => group.length > 0);
}

/**
 * Converts a TItemDetail[] (legacy parser output) into searchable items.
 * - Subtitled details become a single item: subtitle = label, content lines = value array.
 * - Untitled details have each "Label: value" line split into its own item.
 */
function itemDetailsToSearchableItems(details: TItemDetail[]): IDetailsPanelItem[] {
  const items: IDetailsPanelItem[] = [];
  for (const detail of details) {
    if (detail.subtitle) {
      const lines = detail.content.filter((line) => line !== LINE_BREAK);
      if (lines.length === 0) continue;
      items.push({ label: detail.subtitle, value: lines.length === 1 ? lines[0] : lines });
      continue;
    }
    for (const line of detail.content) {
      if (line === LINE_BREAK) continue;
      const colonIdx = line.indexOf(": ");
      if (colonIdx > 0) {
        items.push({ label: line.slice(0, colonIdx), value: line.slice(colonIdx + 2) });
      } else if (line.length > 0) {
        items.push({ label: "", value: line });
      }
    }
  }
  return items;
}

function pair(label: string, value: string | undefined | null): IDetailsPanelItem | null {
  if (!value) return null;
  return { label, value };
}

interface IBuildSearchableSectionsInput {
  workloadDetailed: WorkloadDetailed;
  workloadCrud?: CrudWorkload;
  currentSpec?: WorkloadSpec;
  clusterName: string;
  /** Full ordered list of topology levels (broadest → most specific). */
  topologyLevels?: string[];
  /** Node affinity for the currently selected spec (rendered inside the Compute card). */
  nodeAffinity?: PodNodeAffinity;
}

function buildGeneralSection(input: IBuildSearchableSectionsInput): IDetailsPanelSectionDef | undefined {
  const groups = buildGeneralInfoGroups(input.workloadDetailed, input.clusterName, {
    projectLink: buildProjectLink(input.workloadDetailed),
    clusterLink: buildClusterLink(input.clusterName),
    departmentLink: buildDepartmentLink(input.workloadDetailed),
  });
  const data: IGeneralSectionData = { groups };
  return defineSection(DETAILS_PANEL_SECTIONS.GENERAL, {
    isEmpty: !groups.length,
    data,
    customMatch: (term: string) => matchAnyItem(groups.flat(), term),
  });
}

const CODE_ITEM_LABELS = new Set(["Command", "Arguments"]);

function buildEnvironmentSection(input: IBuildSearchableSectionsInput): IDetailsPanelSectionDef | undefined {
  const { workloadDetailed, currentSpec } = input;
  const items = itemDetailsToSearchableItems(parseEnvironmentDetails(workloadDetailed, currentSpec)).map((it) =>
    CODE_ITEM_LABELS.has(it.label) ? { ...it, isCode: true } : it,
  );
  const environmentVariables = buildEnvironmentVariables(currentSpec, workloadDetailed);
  const applicationParameters = buildApplicationParameters(workloadDetailed, currentSpec);
  const labels = buildLabels(currentSpec, workloadDetailed);
  const annotations = buildAnnotations(currentSpec);

  const data: IEnvironmentSectionData = {
    items,
    environmentVariables,
    applicationParameters,
    labels,
    annotations,
  };
  return defineSection(DETAILS_PANEL_SECTIONS.ENVIRONMENT, {
    isEmpty:
      !items.length &&
      !environmentVariables.length &&
      !applicationParameters.length &&
      !labels.length &&
      !annotations.length,
    data,
    // Section is visible only when at least one rendered value contains the term — mirrors what
    // the section component filters in place. Walk values only (not JSON keys) to avoid false
    // positives like "label" / "value".
    customMatch: (term: string) =>
      matchAnyItem(items, term) ||
      matchAnyItem(applicationParameters, term) ||
      environmentVariables.some((s) => includesCI(s, term)) ||
      labels.some((s) => includesCI(s, term)) ||
      annotations.some((s) => includesCI(s, term)),
  });
}

function buildNetworkSection(input: IBuildSearchableSectionsInput): IDetailsPanelSectionDef | undefined {
  const { workloadDetailed, currentSpec, topologyLevels = [] } = input;
  const connections = buildNetworkConnections(workloadDetailed, currentSpec);
  const topology = workloadDetailed.topology?.name ? workloadDetailed.topology : undefined;
  const data: INetworkSectionData = { connections, topology, topologyLevels: topology ? topologyLevels : [] };
  return defineSection(DETAILS_PANEL_SECTIONS.NETWORK, {
    isEmpty: !connections.length && !topology,
    data,
    customMatch: (term: string) =>
      connections.some((c) => objectMatchesTerm(c, term)) ||
      (topology ? topologyMatches(topology, data.topologyLevels, term) : false),
  });
}

function buildComputeSection(input: IBuildSearchableSectionsInput): IDetailsPanelSectionDef | undefined {
  const { workloadDetailed, workloadCrud, nodeAffinity } = input;
  const items = itemDetailsToSearchableItems(parseComputeDetails(workloadDetailed, workloadCrud));
  const hasNodeAffinity =
    !!nodeAffinity &&
    ((nodeAffinity.required?.nodeSelectorTerms?.length ?? 0) > 0 ||
      (nodeAffinity.preferred?.nodeSelectorTerms?.length ?? 0) > 0);

  const data: IComputeSectionData = { items, nodeAffinity: hasNodeAffinity ? nodeAffinity : undefined };
  return defineSection(DETAILS_PANEL_SECTIONS.COMPUTE, {
    isEmpty: !items.length && !hasNodeAffinity,
    data,
    // Walk only the rendered values (not JSON keys) so "preferred" / "required" keys don't make
    // the card show up with an empty body for unrelated search terms.
    customMatch: (term: string) =>
      matchAnyItem(items, term) || (hasNodeAffinity ? objectMatchesTerm(nodeAffinity, term) : false),
  });
}

function buildStorageSection(input: IBuildSearchableSectionsInput): IDetailsPanelSectionDef | undefined {
  const volumes = buildStorageVolumes(input.currentSpec);
  const data: IStorageSectionData = { volumes };
  return defineSection(DETAILS_PANEL_SECTIONS.STORAGE, {
    isEmpty: !volumes.length,
    data,
    customMatch: (term: string) => volumes.some((v) => objectMatchesTerm(v, term)),
  });
}

function buildCliCommandSection(input: IBuildSearchableSectionsInput): IDetailsPanelSectionDef | undefined {
  const command = input.workloadDetailed.additionalFields?.cliCommand;
  const data: ICliCommandSectionData = { command: command ?? "" };
  return defineSection(DETAILS_PANEL_SECTIONS.CLI_COMMAND, {
    isEmpty: !command,
    data,
    customMatch: (term: string) => includesCI(command ?? "", term),
  });
}

// Order here determines the on-screen order of the accordion cards.
const SECTION_BUILDERS: SectionBuilder<IBuildSearchableSectionsInput>[] = [
  buildGeneralSection,
  buildEnvironmentSection,
  buildNetworkSection,
  buildComputeSection,
  buildStorageSection,
  buildCliCommandSection,
];

/**
 * Orchestrates the workload-detail data into the section-def shape consumed by
 * `<runai-details-panel>`. Each section has its own builder that returns undefined
 * when the section has no data to show; empty sections are filtered out.
 */
export function buildSearchableSections(input: IBuildSearchableSectionsInput): IDetailsPanelSectionDef[] {
  return buildSections(SECTION_BUILDERS, input);
}

function includesCI(haystack: string, term: string): boolean {
  if (!term) return false;
  return haystack.toLowerCase().includes(term.toLowerCase());
}
