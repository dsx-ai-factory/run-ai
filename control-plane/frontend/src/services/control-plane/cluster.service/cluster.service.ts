/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import compareVersions from "semver-compare";
import type {
  ClustersMinimalData,
  DisplayedCluster,
  NetworkTopologyCreateRequest,
  Node,
  NodeExtended,
  NodeFilterSortFields,
  NodeMetricType,
  NodepoolMinimalSortFields,
  NodeTelemetryType,
  TelemetryResponse,
} from "@/swagger-models/cluster-service-client";
import type { ENodeTelemetryGroupBy } from "@/models/cluster.model";
import { API } from "@/common/api.constant";
import type { AxiosResponse } from "axios";
import { clusterServiceApi } from "@/services/infra/client-apis/cluster-service-api/cluster-service-api";
import type {
  MetricsType,
  MetricsResponse,
  GroupedMetricsResponse,
  ClusterInstallationInfoResponse,
  Nodepool,
  NodepoolCreateFields,
  NodepoolUpdateFields,
  NetworkTopology,
} from "@/swagger-models/cluster-service-client";
import { NodepoolSortFilterFields } from "@/swagger-models/cluster-service-client";
import { httpService } from "@/services/infra/https.service/http.service";
import { fileService, FileType } from "@/services/file.service/file.service";
import { filterUtil } from "@/utils/filter.util/filter.util";

export const clusterService = {
  list,
  listMinimal,
  remove,
  create,
  getOperatorVersions,
  update,
  getClusterMetrics,
  getClusterGroupedMetrics,
  getClusterMetricsCsv,
  getNodeTelemetry,
  getNodeTelemetryCsv,
  getNodepoolMetrics,
  getClusterInstallationInfo,
  getNodepools,
  getNodepool,
  getNodeTelemetries,
  deleteNodepool,
  getTenantNodes,
  createNodepool,
  updateNodepool,
  isNodePoolNameUnique,
  countNodepools,
  countTenantNodes,
  getNodeMetrics,
  getNode,
  getLatestClusterVersion,
  getClusterByUuid,
  getNetworkTopologies,
  getNetworkTopologyLevelsByName,
  getNetworkTopologyCount,
  createNetworkTopology,
  deleteNetworkTopology,
  getNetworkTopologyById,
  getNodepoolsMinimal,
};

const ENDPOINT_V2 = `${API.v1}/clusters`;
const NODE_TELEMETRY_ENDPOINT = `${API.v1}/nodes/telemetry`;

// api calls
async function list(): Promise<Array<DisplayedCluster>> {
  try {
    const response = await clusterServiceApi.clustersApi.getClusters();
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function listMinimal(): Promise<Array<ClustersMinimalData>> {
  try {
    const response = await clusterServiceApi.clustersApi.getClustersMinimalData();
    return response.data.clusters || [];
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getClusterByUuid(uuid: string): Promise<DisplayedCluster> {
  try {
    const response = await clusterServiceApi.clustersApi.getClusterByUuid(uuid);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function remove(uuid: string): Promise<void> {
  try {
    await clusterServiceApi.clustersApi.deleteCluster(uuid);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function update(uuid: string, updatedCluster: DisplayedCluster): Promise<void> {
  try {
    await clusterServiceApi.clustersApi.updateCluster(uuid, { name: updatedCluster.name });
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function create(name: string): Promise<DisplayedCluster> {
  try {
    const response = await clusterServiceApi.clustersApi.createCluster({ name });
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getOperatorVersions(): Promise<Record<string, string>> {
  try {
    const response = await clusterServiceApi.clustersApi.getClusterVersions();
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

function getClusterInstallationInfo(
  clusterId: string,
  version: string,
  remoteClusterUrl?: string,
): Promise<AxiosResponse<ClusterInstallationInfoResponse>> {
  try {
    return clusterServiceApi.clustersApi.getClusterInstallInfoByUuid(clusterId, version, remoteClusterUrl);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

// ==== Nodepools ====

async function getNodepoolMetrics(
  clusterUuid: string,
  nodepoolName: string,
  start: string,
  end: string,
  metricType: Array<MetricsType>,
  numberOfSamples?: number,
): Promise<MetricsResponse> {
  const res: AxiosResponse<MetricsResponse> = await clusterServiceApi.nodePoolsApi.getNodepoolMetrics(
    clusterUuid,
    nodepoolName,
    start,
    end,
    metricType,
    numberOfSamples,
  );
  return res.data;
}

async function getClusterMetrics(
  clusterUuid: string,
  start: string,
  end: string,
  metricType: Array<MetricsType>,
  numberOfSamples?: number,
): Promise<MetricsResponse> {
  const res: AxiosResponse<MetricsResponse> = await clusterServiceApi.clustersApi.getClusterMetrics(
    clusterUuid,
    start,
    end,
    metricType,
    numberOfSamples,
  );

  return res.data;
}

async function getClusterGroupedMetrics(
  clusterUuid: string,
  start: string,
  end: string,
  metricType: Array<MetricsType>,
  numberOfSamples?: number,
  groupBy?: "Category" | "Nodepool" | "Project" | "Department",
  nodepoolName?: string,
): Promise<GroupedMetricsResponse> {
  const res: AxiosResponse<GroupedMetricsResponse> = await clusterServiceApi.clustersApi.getClusterMetrics(
    clusterUuid,
    start,
    end,
    metricType,
    numberOfSamples,
    groupBy,
    nodepoolName,
  );

  return res.data;
}

async function getClusterMetricsCsv(
  clusterUuid: string,
  start: string,
  end: string,
  metricType: Array<MetricsType>,
  numberOfSamples?: number,
): Promise<void> {
  const query: Record<string, string | number | Array<string>> = {
    metricType: metricType,
    start: start,
    end: end,
  };
  if (numberOfSamples) {
    query["numberOfSamples"] = numberOfSamples;
  }
  await fileService.downloadFile(ENDPOINT_V2 + "/" + clusterUuid + "/metrics", FileType.csv, undefined, query, {});
}

async function getNodeTelemetry(
  telemetryType: NodeTelemetryType,
  clusterId?: string,
  nodepoolName?: string,
  groupBy?: Array<ENodeTelemetryGroupBy>,
): Promise<TelemetryResponse> {
  try {
    const response = await clusterServiceApi.nodesApi.getNodeTelemetry(telemetryType, clusterId, nodepoolName, groupBy);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getNodeMetrics(
  nodeId: string,
  metricType: Array<NodeMetricType>,
  start: string,
  end: string,
  numberOfSamples?: number,
): Promise<MetricsResponse> {
  const res: AxiosResponse<MetricsResponse> = await clusterServiceApi.nodesApi.getNodeMetrics(
    nodeId,
    metricType,
    start,
    end,
    numberOfSamples,
  );
  return res.data;
}

async function getNode(nodeId: string): Promise<NodeExtended> {
  try {
    const response: AxiosResponse<NodeExtended> = await clusterServiceApi.nodesApi.getNode(nodeId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getNodeTelemetryCsv(
  telemetryType: NodeTelemetryType,
  clusterId?: string,
  nodepoolName?: string,
  groupBy?: Array<ENodeTelemetryGroupBy>,
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
    if (groupBy && groupBy.length > 0) {
      query["groupBy"] = groupBy;
    }
    await fileService.downloadFile(NODE_TELEMETRY_ENDPOINT, FileType.csv, undefined, query, {});
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getNodepools(
  filterBy?: Array<string>,
  sortBy?: NodepoolSortFilterFields,
  sortOrder?: "asc" | "desc",
  offset?: number,
  limit?: number,
): Promise<Nodepool[]> {
  try {
    const response = await clusterServiceApi.nodePoolsApi.getNodepools(filterBy, sortBy, sortOrder, offset, limit);
    return response.data.nodepools || [];
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function countNodepools(filterBy?: Array<string>): Promise<number> {
  try {
    const response = await clusterServiceApi.nodePoolsApi.countNodepools(filterBy);
    return response.data.count;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getNodepool(nodepoolId: string): Promise<Nodepool> {
  try {
    const response = await clusterServiceApi.nodePoolsApi.getNodepool(nodepoolId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function deleteNodepool(nodepoolId: string): Promise<void> {
  try {
    await clusterServiceApi.nodePoolsApi.deleteNodepool(nodepoolId);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createNodepool(nodepool: NodepoolCreateFields): Promise<Nodepool> {
  try {
    const response = await clusterServiceApi.nodePoolsApi.createNodepool(nodepool);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function updateNodepool(nodepoolId: string, nodepool: NodepoolUpdateFields): Promise<Nodepool> {
  try {
    const response = await clusterServiceApi.nodePoolsApi.updateNodepool(nodepoolId, nodepool);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getNodeTelemetries(
  telemetryTypes: NodeTelemetryType[],
  clusterId?: string,
  nodepoolName?: string,
  groupBy?: ENodeTelemetryGroupBy[],
): Promise<TelemetryResponse[]> {
  try {
    const _getTelemetry = async (telemetryType: NodeTelemetryType) =>
      getNodeTelemetry(telemetryType, clusterId, nodepoolName, groupBy);

    const response = await Promise.allSettled(telemetryTypes.map(_getTelemetry));

    return response
      .filter((r) => r.status === "fulfilled" && r.value.values !== null)
      .map((result) => (result as PromiseFulfilledResult<TelemetryResponse>).value);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getTenantNodes(
  filterBy?: Array<string>,
  sortBy?: NodeFilterSortFields,
  sortOrder?: "asc" | "desc",
  offset?: number,
  limit?: number,
  search?: string,
): Promise<Node[]> {
  try {
    const response = await clusterServiceApi.nodesApi.getTenantNodes(filterBy, sortBy, sortOrder, offset, limit, search);
    return response.data.nodes;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function countTenantNodes(filterBy?: Array<string>, search?: string): Promise<number> {
  try {
    const response = await clusterServiceApi.nodesApi.countTenantNodes(filterBy, search);
    return response.data.count;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function isNodePoolNameUnique(clusterId: string, nodepoolName: string): Promise<boolean> {
  try {
    const filterBy = [
      filterUtil.getEqualsFilterString(NodepoolSortFilterFields.ClusterId, clusterId),
      filterUtil.getEqualsFilterString(NodepoolSortFilterFields.Name, nodepoolName),
    ];
    const nodepools = await getNodepools(filterBy);
    return nodepools.length === 0;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getLatestClusterVersion(): Promise<string> {
  const versions = await getOperatorVersions();

  const versionEntries = Object.entries(versions);
  if (versionEntries.length > 0) {
    const sortedVersions = versionEntries.sort((a, b) => {
      const versionA = a[1];
      const versionB = b[1];
      return compareVersions(versionB, versionA);
    });

    return sortedVersions[0][1];
  }
  throw new Error("No cluster versions found");
}

async function getNetworkTopologies(filterBy?: Array<string>): Promise<NetworkTopology[]> {
  try {
    const response = await clusterServiceApi.networkTopologiesApi.getNetworkTopologies(filterBy);
    return response.data.networkTopologies || [];
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

/**
 * Fetches the ordered level list for a single named network topology, scoped to a cluster.
 * Returns an empty array (never throws) when the topology isn't found or the lookup fails —
 * callers use this to enrich a details panel where topology levels are supplementary.
 */
async function getNetworkTopologyLevelsByName(name: string, clusterId: string): Promise<string[]> {
  try {
    const topologies = await getNetworkTopologies([
      filterUtil.getEqualsFilterString("name", name),
      filterUtil.getEqualsFilterString("clusterId", clusterId),
    ]);
    return topologies.length > 0 ? (topologies[0].levels ?? []) : [];
  } catch {
    return [];
  }
}

async function getNetworkTopologyCount(filterBy?: Array<string>): Promise<number> {
  try {
    const response = await clusterServiceApi.networkTopologiesApi.getNetworkTopologiesCount(filterBy);
    return response.data.count;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function createNetworkTopology(networkTopology: NetworkTopologyCreateRequest): Promise<NetworkTopology> {
  try {
    const response = await clusterServiceApi.networkTopologiesApi.createNetworkTopology(networkTopology);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function deleteNetworkTopology(networkTopologyId: string): Promise<void> {
  try {
    await clusterServiceApi.networkTopologiesApi.deleteNetworkTopology(networkTopologyId);
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getNetworkTopologyById(networkTopologyId: string): Promise<NetworkTopology> {
  try {
    const response = await clusterServiceApi.networkTopologiesApi.getNetworkTopologyById(networkTopologyId);
    return response.data;
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}

async function getNodepoolsMinimal(
  filterBy?: Array<string>,
  sortBy?: NodepoolMinimalSortFields,
  sortOrder?: "asc" | "desc",
  offset?: number,
  limit?: number,
) {
  try {
    const response = await clusterServiceApi.nodePoolsApi.getNodepoolsMinimal(
      filterBy,
      sortBy,
      sortOrder,
      offset,
      limit,
    );
    return response.data.nodepools || [];
  } catch (err: unknown) {
    throw httpService.handleHttpError(err);
  }
}
