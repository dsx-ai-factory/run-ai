/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { Position, type Edge, type Node } from "@vue-flow/core";
import type {
  AIApplicationResource,
  AiApplicationEndpoint,
  AiApplicationResourceConnection,
  Workload,
} from "@/swagger-models/workloads-client";

export enum EResourceCardType {
  Workload = "workload",
  Secret = "secret",
  Pvc = "pvc",
  ConfigMap = "configmap",
  Endpoint = "endpoint",
}

export enum EResourceKind {
  Workload = "Workload",
  ExternalWorkload = "ExternalWorkload",
  Secret = "Secret",
  PersistentVolumeClaim = "PersistentVolumeClaim",
  ConfigMap = "ConfigMap",
  Endpoint = "Endpoint",
}

export const FILTERABLE_RESOURCE_KINDS = [
  EResourceKind.Workload,
  EResourceKind.Endpoint,
  EResourceKind.Secret,
  EResourceKind.PersistentVolumeClaim,
  EResourceKind.ConfigMap,
];

export const ENDPOINT_NODE_TYPE = "endpoint";
export const ISLAND_NODE_TYPE = "island";
export const CONNECTION_EDGE_TYPE = "connection";
export const WORKLOAD_TOP_HANDLE_ID = "workload-top";
export const WORKLOAD_BOTTOM_HANDLE_ID = "workload-bottom";
export const RESOURCE_TOP_HANDLE_ID = "resource-top";
export const ENDPOINT_BOTTOM_HANDLE_ID = "endpoint-bottom";

export const LEGEND_BOTTOM_BOUNDARY = 70;
export const WORKLOAD_KINDS = new Set<string>([EResourceKind.Workload, EResourceKind.ExternalWorkload]);
export const STORAGE_KINDS = new Set<string>([
  EResourceKind.Secret,
  EResourceKind.PersistentVolumeClaim,
  EResourceKind.ConfigMap,
]);

const CARD_WIDTH = 230;
const CHILD_HORIZONTAL_GAP = 24;
const VERTICAL_GAP = 70;
const MIN_ISLAND_WIDTH = 280;
const ISLAND_PADDING = 20;
const ISLAND_GAP = 30;
const DEFAULT_AVAILABLE_WIDTH = 1200;
const DEFAULT_CARD_HEIGHT = 50;

export const createWorkloadMap = (workloads: Workload[]): Map<string, Workload> => {
  const map = new Map<string, Workload>();
  workloads.forEach((w) => {
    if (w.id) map.set(w.id, w);
  });
  return map;
};

export interface WorkloadIslandResource {
  nodeId: string;
  resource: AIApplicationResource;
}

export interface WorkloadIslandEndpoint {
  nodeId: string;
  endpoint: AiApplicationEndpoint;
  label: string;
}

export interface WorkloadIsland {
  id: string;
  workloadResource: AIApplicationResource & { id: string };
  workload?: Workload;
  resources: WorkloadIslandResource[];
  endpoints: WorkloadIslandEndpoint[];
}

export interface ApplicationGraph {
  islands: WorkloadIsland[];
  unconnectedResources: WorkloadIslandResource[];
}

export interface BuildDiagramResult {
  nodes: Node[];
  edges: Edge[];
}

export const getGraphStructureFingerprint = (graph: ApplicationGraph): string =>
  [
    ...graph.islands.flatMap((island) => [
      island.id,
      ...island.resources.map((resource) => resource.nodeId),
      ...island.endpoints.map((endpoint) => endpoint.nodeId),
    ]),
    ...graph.unconnectedResources.map((resource) => resource.nodeId),
  ].join("|");

const endpointLabel = (endpoint: AiApplicationEndpoint): string => {
  if (endpoint.url) return endpoint.url;

  const protocol = endpoint.protocol ? `${endpoint.protocol.toLowerCase()}://` : "";
  const host = endpoint.host || "Endpoint";
  const port = endpoint.port ? `:${endpoint.port}` : "";
  return `${protocol}${host}${port}${endpoint.path || ""}`;
};

const endpointFingerprint = (endpoint: AiApplicationEndpoint): string =>
  [endpoint.url, endpoint.protocol, endpoint.host, endpoint.port, endpoint.path]
    .filter((value) => value != null)
    .join("|");

const hashString = (value: string): string => {
  let hash = 0;
  for (let index = 0; index < value.length; index += 1) {
    hash = (hash * 31 + value.charCodeAt(index)) | 0;
  }
  return Math.abs(hash).toString(36);
};

const hasResourceId = (resource: AIApplicationResource): resource is AIApplicationResource & { id: string } =>
  Boolean(resource.id);

const getCardHeight = (cardHeights: Record<string, number>, nodeId: string): number =>
  (Reflect.get(cardHeights, nodeId) as number | undefined) || DEFAULT_CARD_HEIGHT;

const resolveResource = (
  connection: AiApplicationResourceConnection,
  resourcesById: Map<string, AIApplicationResource>,
  resourcesByKindAndName: Map<string, AIApplicationResource>,
): AIApplicationResource => {
  const byUid = connection.uid ? resourcesById.get(connection.uid) : undefined;
  if (byUid) return byUid;

  const byKindAndName = resourcesByKindAndName.get(`${connection.kind}:${connection.name}`);
  return (
    byKindAndName || {
      id: connection.uid,
      name: connection.name,
      kind: connection.kind,
    }
  );
};

export const buildApplicationGraph = (
  resources: AIApplicationResource[],
  resourceConnections: AiApplicationResourceConnection[],
  endpoints: AiApplicationEndpoint[],
  workloads: Workload[],
): ApplicationGraph => {
  const workloadMap = createWorkloadMap(workloads);
  const resourcesById = new Map(resources.filter(hasResourceId).map((resource) => [resource.id, resource]));
  const resourcesByKindAndName = new Map(
    resources
      .filter((resource) => resource.kind && resource.name)
      .map((resource) => [`${resource.kind}:${resource.name}`, resource]),
  );
  const workloadResources = resources.filter((resource) => resource.kind && WORKLOAD_KINDS.has(resource.kind));
  const workloadResourceMap = new Map(
    workloadResources.filter(hasResourceId).map((resource) => [resource.id, resource]),
  );

  workloads.forEach((workload) => {
    if (!workload.id || workloadResourceMap.has(workload.id)) return;
    workloadResourceMap.set(workload.id, {
      id: workload.id,
      name: workload.name || "Workload",
      kind: EResourceKind.Workload,
    });
  });

  const referencedResourceKeys = new Set<string>();
  const islands = Array.from(workloadResourceMap.values()).map((workloadResource): WorkloadIsland => {
    const workloadId = workloadResource.id;
    const connectionsForWorkload = resourceConnections.filter((connection) => connection.workloadId === workloadId);
    const resourceOccurrences = new Map<string, number>();
    const islandResources = connectionsForWorkload.map((connection): WorkloadIslandResource => {
      const resource = resolveResource(connection, resourcesById, resourcesByKindAndName);
      const resourceKey = resource.id || `${resource.kind}:${resource.name}`;
      referencedResourceKeys.add(resourceKey);
      const occurrence = resourceOccurrences.get(resourceKey) || 0;
      resourceOccurrences.set(resourceKey, occurrence + 1);
      return {
        nodeId: `${workloadId}:resource:${resourceKey}:${occurrence}`,
        resource,
      };
    });

    const endpointOccurrences = new Map<string, number>();
    const islandEndpoints = endpoints
      .filter((endpoint) => endpoint.workloadId === workloadId)
      .map((endpoint): WorkloadIslandEndpoint => {
        const fingerprint = endpointFingerprint(endpoint);
        const occurrence = endpointOccurrences.get(fingerprint) || 0;
        endpointOccurrences.set(fingerprint, occurrence + 1);
        return {
          nodeId: `${workloadId}:endpoint:${hashString(fingerprint)}:${occurrence}`,
          endpoint,
          label: endpointLabel(endpoint),
        };
      });

    return {
      id: `island:${workloadId}`,
      workloadResource,
      workload: workloadMap.get(workloadId),
      resources: islandResources,
      endpoints: islandEndpoints,
    };
  });

  const unconnectedResources = resources
    .filter((resource) => resource.kind && STORAGE_KINDS.has(resource.kind))
    .filter((resource) => !referencedResourceKeys.has(resource.id || `${resource.kind}:${resource.name}`))
    .map((resource, index) => ({
      nodeId: `unconnected:resource:${resource.id || `${resource.kind}:${resource.name}`}:${index}`,
      resource,
    }));

  return { islands, unconnectedResources };
};

export interface ApplicationGraphFilter {
  selectedKinds: EResourceKind[];
  searchTerm: string;
}

const matchesSearch = (value: string | undefined, normalizedSearchTerm: string): boolean =>
  !normalizedSearchTerm || value?.toLocaleLowerCase().includes(normalizedSearchTerm) === true;

export const filterApplicationGraph = (
  graph: ApplicationGraph,
  { selectedKinds, searchTerm }: ApplicationGraphFilter,
): ApplicationGraph => {
  const selectedKindSet = new Set(selectedKinds);
  const normalizedSearchTerm = searchTerm.trim().toLocaleLowerCase();
  const workloadKindSelected = selectedKindSet.has(EResourceKind.Workload);

  const islands = graph.islands.flatMap((island): WorkloadIsland[] => {
    const workloadMatchesSearch = matchesSearch(island.workloadResource.name, normalizedSearchTerm);
    const showWorkloadMatch = workloadKindSelected && workloadMatchesSearch;
    const showAllSelectedChildren = Boolean(normalizedSearchTerm) && showWorkloadMatch;
    const resources = island.resources.filter(
      ({ resource }) =>
        selectedKindSet.has(resource.kind as EResourceKind) &&
        (showAllSelectedChildren || matchesSearch(resource.name, normalizedSearchTerm)),
    );
    const endpoints = island.endpoints.filter(
      ({ label }) =>
        selectedKindSet.has(EResourceKind.Endpoint) &&
        (showAllSelectedChildren || matchesSearch(label, normalizedSearchTerm)),
    );

    if (!showWorkloadMatch && resources.length === 0 && endpoints.length === 0) return [];
    return [{ ...island, resources, endpoints }];
  });

  const unconnectedResources = graph.unconnectedResources.filter(
    ({ resource }) =>
      selectedKindSet.has(resource.kind as EResourceKind) && matchesSearch(resource.name, normalizedSearchTerm),
  );

  return { islands, unconnectedResources };
};

const childrenWidth = (count: number): number =>
  count > 0 ? count * CARD_WIDTH + (count - 1) * CHILD_HORIZONTAL_GAP : 0;

const maxNodeHeight = (nodeIds: string[], cardHeights: Record<string, number>): number =>
  nodeIds.reduce((height, nodeId) => Math.max(height, getCardHeight(cardHeights, nodeId)), 0);

export const buildDiagram = (
  graph: ApplicationGraph,
  cardHeights: Record<string, number> = {},
  availableWidth = DEFAULT_AVAILABLE_WIDTH,
): BuildDiagramResult => {
  const nodes: Node[] = [];
  const edges: Edge[] = [];
  const workloadIslandSizes = graph.islands.map((island) => {
    const workloadNodeId = `${island.id}:workload`;
    const endpointNodeIds = island.endpoints.map((endpoint) => endpoint.nodeId);
    const resourceNodeIds = island.resources.map((resource) => resource.nodeId);
    if (endpointNodeIds.length === 0 && resourceNodeIds.length === 0) {
      return { width: CARD_WIDTH, height: getCardHeight(cardHeights, workloadNodeId) };
    }

    const endpointsRowWidth = childrenWidth(endpointNodeIds.length);
    const resourcesRowWidth = childrenWidth(resourceNodeIds.length);
    return {
      width: Math.max(MIN_ISLAND_WIDTH, Math.max(endpointsRowWidth, resourcesRowWidth) + ISLAND_PADDING * 2),
      height:
        ISLAND_PADDING * 2 +
        getCardHeight(cardHeights, workloadNodeId) +
        (endpointNodeIds.length > 0 ? VERTICAL_GAP + maxNodeHeight(endpointNodeIds, cardHeights) : 0) +
        (resourceNodeIds.length > 0 ? VERTICAL_GAP + maxNodeHeight(resourceNodeIds, cardHeights) : 0),
    };
  });
  const unconnectedResourceSizes = graph.unconnectedResources.map((resource) => ({
    width: CARD_WIDTH,
    height: getCardHeight(cardHeights, resource.nodeId),
  }));
  const islandSizes = [...workloadIslandSizes, ...unconnectedResourceSizes];
  const islandPositions: Array<{ x: number; y: number }> = [];
  let rowX = 0;
  let rowY = LEGEND_BOTTOM_BOUNDARY;
  let rowHeight = 0;

  islandSizes.forEach((size) => {
    if (rowX > 0 && rowX + size.width > availableWidth) {
      rowX = 0;
      rowY += rowHeight + ISLAND_GAP;
      rowHeight = 0;
    }
    islandPositions.push({ x: rowX, y: rowY });
    rowX += size.width + ISLAND_GAP;
    rowHeight = Math.max(rowHeight, size.height);
  });

  graph.islands.forEach((island, islandIndex) => {
    const workloadId = island.workloadResource.id;
    const workloadNodeId = `${island.id}:workload`;
    const workloadHeight = getCardHeight(cardHeights, workloadNodeId);
    const endpointCount = island.endpoints.length;
    const resourceCount = island.resources.length;
    const hasConnections = endpointCount > 0 || resourceCount > 0;
    const islandPosition = islandPositions.at(islandIndex) || { x: 0, y: LEGEND_BOTTOM_BOUNDARY };
    const islandSize = workloadIslandSizes.at(islandIndex) || { width: MIN_ISLAND_WIDTH, height: 0 };
    const contentTop = hasConnections ? islandPosition.y + ISLAND_PADDING : islandPosition.y;
    const endpointY = contentTop;
    const endpointHeight =
      endpointCount > 0
        ? maxNodeHeight(
            island.endpoints.map((endpoint) => endpoint.nodeId),
            cardHeights,
          )
        : 0;
    const workloadY = endpointCount > 0 ? endpointY + endpointHeight + VERTICAL_GAP : contentTop;
    const resourceY = workloadY + workloadHeight + (resourceCount > 0 ? VERTICAL_GAP : 0);
    const workloadX = hasConnections ? islandPosition.x + (islandSize.width - CARD_WIDTH) / 2 : islandPosition.x;
    let endpointX = islandPosition.x + (islandSize.width - childrenWidth(endpointCount)) / 2;
    let resourceX = islandPosition.x + (islandSize.width - childrenWidth(resourceCount)) / 2;

    if (hasConnections) {
      nodes.push({
        id: island.id,
        type: ISLAND_NODE_TYPE,
        position: islandPosition,
        data: {},
        style: { width: `${islandSize.width}px`, height: `${islandSize.height}px` },
        selectable: false,
        draggable: false,
        zIndex: 0,
      });
    }

    nodes.push({
      id: workloadNodeId,
      type: EResourceKind.Workload,
      position: { x: workloadX, y: workloadY },
      data: {
        name: island.workloadResource.name,
        kind: island.workloadResource.kind,
        resourceId: workloadId,
        workloadId,
        phase: island.workload?.phase,
      },
      sourcePosition: Position.Bottom,
      zIndex: 3,
    });

    island.endpoints.forEach((endpoint) => {
      nodes.push({
        id: endpoint.nodeId,
        type: ENDPOINT_NODE_TYPE,
        position: { x: endpointX, y: endpointY },
        data: { url: endpoint.label },
        targetPosition: Position.Bottom,
        zIndex: 3,
      });
      endpointX += CARD_WIDTH + CHILD_HORIZONTAL_GAP;
    });

    island.resources.forEach(({ nodeId, resource }) => {
      nodes.push({
        id: nodeId,
        type: resource.kind,
        position: { x: resourceX, y: resourceY },
        data: {
          name: resource.name,
          kind: resource.kind,
          resourceId: resource.id,
        },
        targetPosition: Position.Top,
        zIndex: 3,
      });
      resourceX += CARD_WIDTH + CHILD_HORIZONTAL_GAP;
    });

    island.endpoints.forEach((endpoint) => {
      edges.push({
        id: `edge:${workloadNodeId}:${endpoint.nodeId}`,
        source: workloadNodeId,
        sourceHandle: WORKLOAD_TOP_HANDLE_ID,
        target: endpoint.nodeId,
        targetHandle: ENDPOINT_BOTTOM_HANDLE_ID,
        type: CONNECTION_EDGE_TYPE,
        zIndex: 2,
      });
    });
    island.resources.forEach((resource) => {
      edges.push({
        id: `edge:${workloadNodeId}:${resource.nodeId}`,
        source: workloadNodeId,
        sourceHandle: WORKLOAD_BOTTOM_HANDLE_ID,
        target: resource.nodeId,
        targetHandle: RESOURCE_TOP_HANDLE_ID,
        type: CONNECTION_EDGE_TYPE,
        zIndex: 2,
      });
    });
  });

  graph.unconnectedResources.forEach(({ nodeId, resource }, index) => {
    const position = islandPositions.at(graph.islands.length + index) || { x: 0, y: LEGEND_BOTTOM_BOUNDARY };
    nodes.push({
      id: nodeId,
      type: resource.kind,
      position,
      data: {
        name: resource.name,
        kind: resource.kind,
        resourceId: resource.id,
      },
      zIndex: 3,
    });
  });

  return { nodes, edges };
};
