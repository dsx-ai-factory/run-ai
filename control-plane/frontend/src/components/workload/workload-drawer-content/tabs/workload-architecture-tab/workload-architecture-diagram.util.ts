import { Position, type Node, type Edge } from "@vue-flow/core";
import type { Workload, WorkloadElement, Pod, WorkloadElementResponse } from "@/swagger-models/workloads-client";
import { DEFAULT_ENDPOINT_HANDLE_ID } from "@/components/common/diagram-layout/diagram-layout.util";
import { workloadUtil } from "@/utils/workload.util/workload.util";

export const NODE_TYPE = "element";
export const ENDPOINT_NODE_TYPE = "endpoint";
export const HORIZONTAL_EDGE_TYPE = "horizontal-circle";

export const WORKLOAD_RIGHT_HANDLE_ID = "right";
export const ENDPOINT_LEFT_HANDLE_ID = DEFAULT_ENDPOINT_HANDLE_ID;

const CARD_WIDTH = 200;
const HORIZONTAL_GAP = 40;
const ROW_GAP = 60;
const DEFAULT_CARD_HEIGHT = 50;
const ENDPOINT_HORIZONTAL_GAP = 60;
const ENDPOINT_VERTICAL_GAP = 12;

export enum EArchitectureNodeKind {
  Workload = "workload",
  Element = "element",
  Endpoint = "endpoint",
}

export interface NodeResourceInfo {
  gpuComputeAllocation: number | null;
  gpuMemoryAllocation: string | null;
  cpuComputeAllocation: number | null;
  cpuMemoryAllocation: string | null;
}

export interface ArchitectureNodeData {
  name: string;
  kind: EArchitectureNodeKind;
  parentId: string | null;
  elementKind: string | null;
  kubernetesName: string | null;
  rowIndex: number;
  phase: string | null;
  workloadType: string | null;
  resources: NodeResourceInfo | null;
  podsDisplay: string | null;
}

export interface EndpointNodeData {
  url: string;
}

interface TreeNode {
  id: string;
  data: ArchitectureNodeData;
  children: TreeNode[];
}

export type WorkloadElementDetailsById = Record<string, WorkloadElementResponse>;

const countPodsPerElement = (pods: Pod[]): Map<string, number> => {
  const counts = new Map<string, number>();

  for (const pod of pods) {
    const elements = pod.elements;
    if (!elements?.length) continue;

    let deepest = elements[0];
    for (const el of elements) {
      if (el.hierarchyLevel > deepest.hierarchyLevel) {
        deepest = el;
      }
    }

    const current = counts.get(deepest.elementId) ?? 0;
    counts.set(deepest.elementId, current + 1);
  }

  return counts;
};

const getWorkloadPodsDisplay = (workload: Workload): string | null => {
  const value = workloadUtil.getRunningVsRequestedDisplayedValue(workload);
  return value === "-" ? null : value;
};

const mapAllocatedResources = (
  allocated?: Workload["allocatedResources"] | WorkloadElementResponse["allocatedResources"] | null,
): NodeResourceInfo | null => {
  if (!allocated) return null;

  return {
    gpuComputeAllocation: allocated.gpu ?? null,
    gpuMemoryAllocation: allocated.gpuMemory ?? null,
    cpuComputeAllocation: allocated.cpu ?? null,
    cpuMemoryAllocation: allocated.cpuMemory ?? null,
  };
};

const buildTree = (
  workload: Workload,
  elements: WorkloadElement[],
  pods: Pod[],
  elementDetailsById: WorkloadElementDetailsById,
): TreeNode => {
  const workloadId = workload.id ?? "workload-root";
  const podCountsByElement = countPodsPerElement(pods);

  const root: TreeNode = {
    id: workloadId,
    data: {
      name: workload.name ?? "Workload",
      kind: EArchitectureNodeKind.Workload,
      parentId: null,
      elementKind: null,
      kubernetesName: null,
      rowIndex: 0,
      phase: workload.phase ?? null,
      workloadType: workload.type ?? null,
      resources: mapAllocatedResources(workload.allocatedResources),
      podsDisplay: getWorkloadPodsDisplay(workload),
    },
    children: [],
  };

  const nodeMap = new Map<string, TreeNode>();
  nodeMap.set(workloadId, root);

  for (const element of elements) {
    const podCount = podCountsByElement.get(element.id) ?? null;
    const elementDetails = elementDetailsById[element.id];
    const node: TreeNode = {
      id: element.id,
      data: {
        name: element.name,
        kind: EArchitectureNodeKind.Element,
        parentId: element.parentId ?? workloadId,
        elementKind: element.kubernetesRef?.gvk?.kind ?? null,
        kubernetesName: element.kubernetesRef?.name ?? null,
        rowIndex: element.hierarchyLevel,
        phase: null,
        workloadType: null,
        resources: mapAllocatedResources(elementDetails?.allocatedResources),
        podsDisplay: podCount !== null ? String(podCount) : null,
      },
      children: [],
    };
    nodeMap.set(element.id, node);
  }

  for (const [, node] of nodeMap) {
    if (node === root) continue;
    const parentId = node.data.parentId;
    const parent = parentId ? nodeMap.get(parentId) : root;
    if (parent) {
      parent.children.push(node);
    } else {
      root.children.push(node);
    }
  }

  return root;
};

const collectAllNodes = (node: TreeNode, result: TreeNode[]): void => {
  result.push(node);
  for (const child of node.children) {
    collectAllNodes(child, result);
  }
};

const computeRowYOffsets = (allTreeNodes: TreeNode[], cardHeights: Record<string, number>): Map<number, number> => {
  const maxHeightByRow = new Map<number, number>();

  for (const node of allTreeNodes) {
    const row = node.data.rowIndex;
    const height = cardHeights[node.id] || DEFAULT_CARD_HEIGHT;
    const current = maxHeightByRow.get(row) ?? 0;
    maxHeightByRow.set(row, Math.max(current, height));
  }

  const rows = Array.from(maxHeightByRow.keys()).sort((a, b) => a - b);
  const offsets = new Map<number, number>();

  let currentY = 0;
  for (const row of rows) {
    offsets.set(row, currentY);
    currentY += (maxHeightByRow.get(row) ?? DEFAULT_CARD_HEIGHT) + ROW_GAP;
  }

  return offsets;
};

const assignPositions = (
  node: TreeNode,
  startX: number,
  rowYOffsets: Map<number, number>,
  nodes: Node[],
  edges: Edge[],
): number => {
  const y = rowYOffsets.get(node.data.rowIndex) ?? 0;

  if (node.children.length === 0) {
    nodes.push({
      id: node.id,
      type: NODE_TYPE,
      position: { x: startX, y },
      data: { ...node.data },
    });
    return startX + CARD_WIDTH + HORIZONTAL_GAP;
  }

  let currentX = startX;
  for (const child of node.children) {
    currentX = assignPositions(child, currentX, rowYOffsets, nodes, edges);

    edges.push({
      id: `edge-${node.id}-${child.id}`,
      source: node.id,
      target: child.id,
      type: "circle",
    });
  }

  const firstChildNode = nodes.find((n) => n.id === node.children[0]?.id);
  const lastChildNode = nodes.find((n) => n.id === node.children[node.children.length - 1]?.id);

  const childrenStartX = firstChildNode?.position.x ?? startX;
  const childrenEndX = (lastChildNode?.position.x ?? startX) + CARD_WIDTH;
  const centerX = childrenStartX + (childrenEndX - childrenStartX) / 2 - CARD_WIDTH / 2;

  nodes.push({
    id: node.id,
    type: NODE_TYPE,
    position: { x: centerX, y },
    data: { ...node.data },
  });

  return currentX;
};

export interface BuildResult {
  nodes: Node[];
  edges: Edge[];
}

const getWorkloadEndpoints = (workload: Workload): string[] => {
  const endpoints = workload.endpoints;
  if (!endpoints?.length) return [];
  return endpoints
    .map((endpoint) => endpoint.url)
    .filter((url): url is string => typeof url === "string" && url.length > 0);
};

const appendEndpointNodesAndEdges = (
  workloadId: string,
  workloadNode: Node | undefined,
  endpoints: string[],
  nodes: Node[],
  edges: Edge[],
  cardHeights: Record<string, number>,
): void => {
  if (!workloadNode || endpoints.length === 0) return;

  const endpointStartX = workloadNode.position.x + CARD_WIDTH + ENDPOINT_HORIZONTAL_GAP;
  let nextY = workloadNode.position.y;

  endpoints.forEach((url, index) => {
    const endpointId = `${workloadId}-endpoint-${index}`;

    nodes.push({
      id: endpointId,
      type: ENDPOINT_NODE_TYPE,
      position: { x: endpointStartX, y: nextY },
      data: { url } as EndpointNodeData,
      targetPosition: Position.Left,
    });

    edges.push({
      id: `edge-${workloadId}-${endpointId}`,
      source: workloadId,
      sourceHandle: WORKLOAD_RIGHT_HANDLE_ID,
      target: endpointId,
      targetHandle: ENDPOINT_LEFT_HANDLE_ID,
      type: HORIZONTAL_EDGE_TYPE,
    });

    const endpointHeight = cardHeights[endpointId] || DEFAULT_CARD_HEIGHT;
    nextY += endpointHeight + ENDPOINT_VERTICAL_GAP;
  });
};

export const buildDiagram = (
  workload: Workload,
  elements: WorkloadElement[],
  pods: Pod[],
  cardHeights: Record<string, number> = {},
  elementDetailsById: WorkloadElementDetailsById = {},
): BuildResult => {
  const tree = buildTree(workload, elements, pods, elementDetailsById);

  const allTreeNodes: TreeNode[] = [];
  collectAllNodes(tree, allTreeNodes);

  const rowYOffsets = computeRowYOffsets(allTreeNodes, cardHeights);

  const nodes: Node[] = [];
  const edges: Edge[] = [];
  assignPositions(tree, 0, rowYOffsets, nodes, edges);

  const workloadId = workload.id ?? "workload-root";
  const workloadNode = nodes.find((n) => n.id === workloadId);
  const endpoints = getWorkloadEndpoints(workload);
  appendEndpointNodesAndEdges(workloadId, workloadNode, endpoints, nodes, edges, cardHeights);

  return { nodes, edges };
};
