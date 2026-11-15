/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { IStatusColOptions } from "@/models/table.model";
import { CLUSTER_STATUS_TOOLTIP } from "@/models/cluster.model";
import { splitCamelCaseAndCapitalize } from "@/utils/string.util/string.util";
import type {
  ClusterDependenciesStatus,
  ClusterDependencyStatus,
  ClusterDisplayedStatus,
  ClustersMinimalData,
} from "@/swagger-models/cluster-service-client";
import { ClusterState, ClusterReportedStatusPlatformTypeEnum } from "@/swagger-models/cluster-service-client";
import type { ClusterReportedStatusOperandsValue } from "@/swagger-models/cluster-service-client";
import { DOCS } from "@/constants/docs";
import { docsUtil } from "@/utils/docs.util/docs.util";

export const clusterUtil = {
  getStatusColOptions,
  addSessionCreatedCluster,
  removeSessionCreatedCluster,
  getNonFunctioningServices,
  getMissingPrerequisites,
  getDistribution,
  isKnativeInstalledInCluster,
  isLwsInstalledInCluster,
  findClusterById,
  formatClusterNameWithVersion,
  isClusterReachable,
};

// Degraded clusters are still live, so features that don't depend on what is degraded stay available.
const REACHABLE_CLUSTER_STATES: ReadonlySet<ClusterState> = new Set<ClusterState>([
  ClusterState.Connected,
  ClusterState.MissingPrerequisites,
  ClusterState.ServiceIssues,
]);

function isClusterReachable(state?: ClusterState): boolean {
  return !!state && REACHABLE_CLUSTER_STATES.has(state);
}

function isKnativeInstalledInCluster(cluster: ClustersMinimalData): boolean {
  return cluster.dependenciesStatus?.optional?.knative?.available || false;
}

function isLwsInstalledInCluster(cluster: ClustersMinimalData): boolean {
  return cluster.dependenciesStatus?.optional?.lws?.available || false;
}

function getStatusColOptions(status?: ClusterDisplayedStatus): IStatusColOptions {
  if (!status) {
    return {
      filterKey: "status",
      status: "-",
      displayAnimation: false,
      colStyle: { "min-width": "150px" },
    };
  }
  return {
    status: _getStatus(status.state as ClusterState),
    tooltipText: _getTooltipText(status),
    color: _statusColor(status.state as ClusterState),
    displayAnimation: false,
    filterKey: "status",
    colStyle: { "min-width": "150px" },
  };
}

function _getStatus(status: ClusterState): string {
  if (!Object.keys(ClusterState).includes(status)) return "-";
  return splitCamelCaseAndCapitalize(status).trim();
}

function getClusterTooltipSupportLinks(): string {
  const supportListItem = DOCS.SUPPORT.URL ? `<li style="list-style: disc;">${docsUtil.getSupportHTMLLink()}</li>` : "";
  return (
    "<br>" +
    "<div style='padding-left: 20px'>" +
    '<ul style="list-style: disc;">' +
    `<li style="list-style: disc;">${docsUtil.getTroubleshootingHTMLLink(DOCS.TROUBLESHOOTING.CLUSTERS)}</li>` +
    supportListItem +
    "</ul>" +
    "</div>"
  );
}

function _getTooltipText(status: ClusterDisplayedStatus): string {
  switch (status.state) {
    case ClusterState.WaitingToConnect:
      return CLUSTER_STATUS_TOOLTIP.WaitingToConnect + getClusterTooltipSupportLinks();
    case ClusterState.Connected:
      return CLUSTER_STATUS_TOOLTIP.Connected;
    case ClusterState.Disconnected:
      return CLUSTER_STATUS_TOOLTIP.Disconnected + getClusterTooltipSupportLinks();
    case ClusterState.ServiceIssues:
      return CLUSTER_STATUS_TOOLTIP.ServiceIssues + getClusterTooltipSupportLinks();
    case ClusterState.MissingPrerequisites:
      return getMissingPrerequisitesText(status);
    case ClusterState.Unknown:
      return DOCS.SUPPORT.URL
        ? CLUSTER_STATUS_TOOLTIP.Unknown + docsUtil.getSupportHTMLLink()
        : "The cluster status couldn't be determined.";
    default:
      return "-";
  }
}

function _statusColor(status: ClusterState): string {
  switch (status) {
    case ClusterState.Connected:
      return "success";
    case ClusterState.Unknown:
    case ClusterState.WaitingToConnect:
      return "normal";
    default:
      return "warning";
  }
}

const storageKey = "sessionCreatedClustersIds";

function _setSessionCreatedClusters(newClustersUuidList: string[]): void {
  sessionStorage.setItem(storageKey, JSON.stringify(newClustersUuidList));
}

function getSessionCreatedClusters(): string[] {
  const storageInfo: string | null = sessionStorage.getItem(storageKey);
  if (!storageInfo) return [];
  return JSON.parse(storageInfo);
}

function addSessionCreatedCluster(newClusterCreatedUuid: string): void {
  const createdClustersUuids: string[] = getSessionCreatedClusters();
  const newList: string[] = [...createdClustersUuids, newClusterCreatedUuid];
  _setSessionCreatedClusters(newList);
}

function removeSessionCreatedCluster(clusterUuid: string): void {
  const createdClustersUuids: string[] = getSessionCreatedClusters();
  const updatedList: string[] = createdClustersUuids.filter((uuid) => uuid !== clusterUuid);
  _setSessionCreatedClusters(updatedList);
}

function getNonFunctioningServices(operands: { [key: string]: ClusterReportedStatusOperandsValue }): string[] {
  return (
    Object.entries(operands)
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      .filter(([_, operand]) => operand.unreadyThresholdCrossed)

      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      .map(([key, _]) => key)
  );
}

function getMissingPrerequisites(dependencies: ClusterDependenciesStatus): string[] {
  return (
    Object.entries(dependencies.required ?? {})
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      .filter(([_, dependency]) => !dependency.available)

      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      .map(([key, _]) => key)
  );
}

// The gpu-stack dependency covers both the NVIDIA stack and GPU fractioning. Naming the
// dependency alone would not say which of them is unsupported, so the fractioning reason the
// cluster reports is rendered instead.
function getGpuFractioningPrerequisiteReason(dependencies: ClusterDependenciesStatus): string | undefined {
  const gpuFractioning = dependencies.required?.["gpu-stack"]?.components?.["gpu-fractioning"];
  if (!gpuFractioning || gpuFractioning.available) return undefined;
  return gpuFractioning.reason;
}

// A dependency can report itself available while a component nested under it is not — the
// operator does exactly that for nvidia, whose own status stays true when nvidia-metrics fails.
// So the whole subtree has to be walked, not just the direct children.
function hasFailureOtherThanFractioning(components: { [key: string]: ClusterDependencyStatus }): boolean {
  return Object.entries(components).some(([name, component]) => {
    if (name === "gpu-fractioning") return false;
    if (!component.available) return true;
    return component.components ? hasFailureOtherThanFractioning(component.components) : false;
  });
}

// gpu-stack is reported by name alone, which says nothing a reader can act on. When fractioning
// is its only failing part, the dedicated section carries the reason instead; when anything else
// under gpu-stack is also failing, the name has to stay or that failure goes unreported.
function isFractioningOnlyGpuStackFailure(dependencies: ClusterDependenciesStatus): boolean {
  const components = dependencies.required?.["gpu-stack"]?.components;
  if (!components) return false;

  return !hasFailureOtherThanFractioning(components);
}

function getMissingPrerequisitesText(status: ClusterDisplayedStatus): string {
  const dependencies = status.dependencies as ClusterDependenciesStatus;
  const gpuFractioningReason = getGpuFractioningPrerequisiteReason(dependencies);

  const suppressGpuStackName = Boolean(gpuFractioningReason) && isFractioningOnlyGpuStackFailure(dependencies);

  const missingPrerequisites = clusterUtil
    .getMissingPrerequisites(dependencies)
    .filter((name) => !(suppressGpuStackName && name === "gpu-stack"));

  const sections: string[] = [];

  if (missingPrerequisites.length) {
    sections.push(
      CLUSTER_STATUS_TOOLTIP.MissingPrerequisites.replace("{{missing_prerequisit}}", missingPrerequisites.join(", ")),
    );
  }

  if (gpuFractioningReason) {
    sections.push(CLUSTER_STATUS_TOOLTIP.MissingGpuFractioningPrerequisites.replace("{{reason}}", gpuFractioningReason));
  }

  if (!sections.length) {
    sections.push(CLUSTER_STATUS_TOOLTIP.MissingPrerequisites.replace("{{missing_prerequisit}}", "none"));
  }

  return [...sections, CLUSTER_STATUS_TOOLTIP.PrerequisitesGuide].filter(Boolean).join("<br><br>");
}

function getDistribution(distribution: ClusterReportedStatusPlatformTypeEnum): string {
  switch (distribution) {
    case ClusterReportedStatusPlatformTypeEnum.Aks:
    case ClusterReportedStatusPlatformTypeEnum.Eks:
    case ClusterReportedStatusPlatformTypeEnum.Gke:
      return distribution.toUpperCase();
    case ClusterReportedStatusPlatformTypeEnum.Vanilla:
      return "Vanilla";
    case ClusterReportedStatusPlatformTypeEnum.Openshift:
      return "OpenShift";
    default:
      return "-";
  }
}

function findClusterById(clusters: ClustersMinimalData[], clusterId: string): ClustersMinimalData | undefined {
  return clusters.find((cluster) => cluster.uuid === clusterId);
}

function formatClusterNameWithVersion(cluster: ClustersMinimalData | undefined | null): string {
  if (!cluster) return "";
  return cluster.version && cluster.version.trim() ? `${cluster.name} (v${cluster.version})` : cluster.name;
}
