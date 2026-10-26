/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import {
  GpuFractioningConditionType,
  GpuFractioningNotManagedPrerequisite,
  GpuFractioningNotManagedReason,
  type INode,
  NodeConditionMap,
  NodeStatusMap,
  NrtHealthyConditionType,
} from "@/models/node.model";
import {
  type MetricsResponse,
  type Node,
  NodeMetricType,
  type NodeStatusConditionDetails,
} from "@/swagger-models/cluster-service-client";
import type { IStatusColOptions } from "@/models/table.model";
import { tableUtil } from "@/utils/table.util";
import { docsUtil } from "@/utils/docs.util/docs.util";
import { stringUtil } from "@/utils/string.util";
import type { IGpuModalRow } from "@/table-models/node.table-model";

const NODE_STATUS_TOOLTIP_MAX_WIDTH = "300px";
const GPU_FRACTIONING_PREREQUISITES_INTRO = "The following prerequisites are not met on the node:";
const GPU_FRACTIONING_PREREQUISITES_OUTCOME = "As a result, GPU fractions are not supported on this node.";

export const nodeUtil = {
  enrichNodesWithInfoFromCluster,
  getStatusColOptions,
  parseNodeGpuMeasurements,
};

function enrichNodesWithInfoFromCluster(nodes: Array<INode>, clusterServiceNodes: Node[]): Array<INode> {
  if (clusterServiceNodes.length === 0) {
    return nodes;
  }
  return clusterServiceNodes.map((csNode: Node) => {
    const node: INode = nodes.find((node: INode) => node.node === csNode.name) || _getEmptyNode();
    return {
      ...node,
      node: csNode.name,
      nodePool: csNode.nodePool,
      gpuType: csNode?.gpuInfo?.gpuType,
      totalGpus: csNode?.gpuInfo?.gpuCount?.toString(),
      status: csNode.status,
      conditions: csNode?.conditions,
      clusterId: csNode.clusterUuid,
    };
  });
}

function getStatusColOptions(node: Node): IStatusColOptions {
  const statusOptions: IStatusColOptions = NodeStatusMap[node.status];
  const tooltipText: string = getConditionsAsText(node);

  return {
    ...tableUtil.getStatusColOptions(statusOptions, tooltipText),
    tooltipMaxWidth: NODE_STATUS_TOOLTIP_MAX_WIDTH,
  };
}

function getConditionsAsText(node: Node): string {
  let conditions = "";
  node.conditions?.forEach((condition: NodeStatusConditionDetails) => {
    const message = _getConditionMessage(condition);
    if (message) {
      // Each condition is its own multi-line block ending in a consequence and a guide link, so
      // they need a blank line between them or two blocks read as one paragraph.
      conditions += (conditions ? "<br>" : "") + message + "<br>";
    }
  });
  return conditions;
}

function _getConditionMessage(condition: NodeStatusConditionDetails): string | undefined {
  // The NRT-health message text depends on the specific unmet prerequisite (reason), so it's
  // rendered from the backend-provided message rather than a single static NodeConditionMap entry.
  if (condition.type === NrtHealthyConditionType) {
    return condition.message
      ? `${stringUtil.replaceNewLineWithBreaks(condition.message)} ${docsUtil.getNumaAwareSchedulingHTMLLink()}`
      : undefined;
  }
  // The GPU fractioning operator supplies its own reason and message per failing node, so the
  // unmet prerequisite is rendered from the backend and only the surrounding copy is fixed.
  if (condition.type === GpuFractioningConditionType) {
    const unmetPrerequisite = _getGpuFractioningPrerequisite(condition);
    return [
      GPU_FRACTIONING_PREREQUISITES_INTRO,
      unmetPrerequisite,
      GPU_FRACTIONING_PREREQUISITES_OUTCOME,
      docsUtil.getPrerequisitesHTMLLink(),
    ]
      .filter((line): line is string => Boolean(line))
      .join("<br>");
  }
  return NodeConditionMap[condition.type]?.message;
}

// The backend states the unmanaged case as a full sentence, which repeats the frame's own
// consequence line, so that one prerequisite is phrased here instead.
function _getGpuFractioningPrerequisite(condition: NodeStatusConditionDetails): string | undefined {
  if (condition.reason === GpuFractioningNotManagedReason) {
    return GpuFractioningNotManagedPrerequisite;
  }
  return condition.message
    ? stringUtil.replaceNewLineWithBreaks(condition.message)
    : NodeConditionMap[condition.type]?.message;
}

function _getEmptyNode(): INode {
  return {
    clusterId: "",
    id: "",
    node: "",
    nodeStatus: "",
    utilization: "",
    gpuType: "",
    totalGpus: "",
    allocatedGpus: "",
    usedGpus: "",
    totalCpus: "",
    allocatedCpus: "",
    usedCpus: "",
    totalGpuMemory: "",
    usedGpuMemory: "",
    totalCpuMemory: "",
    allocatedMemory: "",
    usedCpuMemory: "",
    swapCpuMemory: "",
    nodePool: "",
    gpuDevices: [],
    conditions: [],
  };
}

/*** Metric Util ***/

function parseNodeGpuMeasurements(metricsResponse: MetricsResponse): IGpuModalRow[] {
  const gpuMap = new Map<string, Partial<IGpuModalRow>>();

  metricsResponse.measurements.forEach((measurement) => {
    const gpuIndex = measurement.labels?.gpu;
    if (!gpuIndex) return;
    if (!gpuMap.has(gpuIndex)) {
      gpuMap.set(gpuIndex, {});
    }
    const gpuData = gpuMap.get(gpuIndex) as Partial<IGpuModalRow>;
    const value = measurement.values?.[0]?.value;
    if (value === undefined) return;
    switch (measurement.type) {
      case NodeMetricType.GpuAllocationPerGpu:
        gpuData.allocatedCompute = Number(value);
        break;
      case NodeMetricType.GpuMemoryAllocationBytesPerGpu:
        gpuData.allocatedMemory = Number(value);
        break;
      case NodeMetricType.GpuMemoryUsageBytesPerGpu:
        gpuData.usedMemory = Number(value);
        break;
      case NodeMetricType.GpuIdleTimeSecondsPerGpu:
        gpuData.idleTime = value;
        break;
      case NodeMetricType.GpuUtilizationPerGpu:
        gpuData.computeUtilization = Number(value);
        break;
      case NodeMetricType.GpuMemoryUtilizationPerGpu:
        gpuData.memoryUtilization = Number(value);
        break;
    }
  });

  return Array.from(gpuMap.entries())
    .map(([index, data]) => ({
      index: Number(index),
      allocatedCompute: data.allocatedCompute || 0,
      allocatedMemory: data.allocatedMemory || 0,
      usedMemory: data.usedMemory || 0,
      computeUtilization: data.computeUtilization || 0,
      memoryUtilization: data.memoryUtilization || 0,
      idleTime: data.idleTime !== undefined ? String(data.idleTime) : "",
    }))
    .sort((a, b) => a.index - b.index);
}
