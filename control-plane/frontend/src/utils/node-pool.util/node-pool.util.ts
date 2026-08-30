import {
  EMPTY_PRIORITY_LABEL,
  NodePoolPhaseMap,
  NodePoolStatusMessageMap,
  NodePoolStatusMessagePrefix,
  NodePoolStatusMessageAllowedPrefixesToKeep,
  NRT_NODEPOOL_PREREQ_MESSAGE_PREFIX,
} from "@/models/node-pool.model";
import { docsUtil } from "@/utils/docs.util/docs.util";
import {
  type Nodepool,
  type NodepoolCreateFields,
  type NodepoolUpdateFields,
  NodeTelemetryType,
  type TelemetryResponse,
  type TelemetryResponseValuesInner,
} from "@/swagger-models/cluster-service-client";
import { binPackLabel, spreadLabel } from "@/models/node-pool.model";
import { PlacementStrategy } from "@/swagger-models/cluster-service-client";

import type { IStatusColOptions } from "@/models/table.model";
import { tableUtil } from "@/utils/table.util";
import type { IAdditionalNodePoolTableFields, INodePoolTable } from "@/table-models/node-pool.table-model";
import { ENodeTelemetryGroupBy } from "@/models/cluster.model";
import type { INodeTable } from "@/table-models/node.table-model";
import { deepCopy } from "@/utils/common.util";
import { DEFAULT_NODE_POOL_NAME } from "@/models/resource.model";
import { stringUtil } from "@/utils/string.util";

// Matches the Status legend tooltip width in the Node Pools Figma design (RUN-40875).
const NODE_POOL_STATUS_TOOLTIP_MAX_WIDTH = "300px";

export const nodePoolUtil = {
  placementStrategyText,
  getNodePoolPriorityByName,
  getStatusColOptions,
  enrichNodePoolsWithTelemetry,
  getEmptyNodePool,
  getNodePoolUpdateRequest,
  enrichNodesWithTelemetry,
};

function placementStrategyText(strategy: PlacementStrategy): string {
  return strategy === PlacementStrategy.Binpack ? binPackLabel : spreadLabel;
}

function getNodePoolPriorityByName(defaultNodePools: string[], name: string): number | string {
  const index = defaultNodePools.indexOf(name);
  if (index > -1) {
    return index + 1;
  }
  return EMPTY_PRIORITY_LABEL;
}

function getStatusColOptions(nodePool: INodePoolTable): IStatusColOptions {
  if (!nodePool.phase)
    return {
      status: "-",
      tooltipText: "",
      displayAnimation: false,
      filterKey: "status",
    };
  const statusOptions: IStatusColOptions = NodePoolPhaseMap[nodePool.phase];
  const tooltipText = nodePool.phaseMessage && _getPhaseMessage(nodePool.phaseMessage);
  return {
    ...tableUtil.getStatusColOptions(statusOptions, tooltipText),
    tooltipMaxWidth: NODE_POOL_STATUS_TOOLTIP_MAX_WIDTH,
  };
}

function _getPhaseMessage(phaseMessage: string) {
  if (phaseMessage.startsWith(NRT_NODEPOOL_PREREQ_MESSAGE_PREFIX)) {
    return `${stringUtil.replaceNewLineWithBreaks(phaseMessage)} ${docsUtil.getNumaAwareSchedulingHTMLLink()}`;
  }

  let newMessage = NodePoolStatusMessageMap.get(phaseMessage)?.newMessage;
  if (!newMessage) {
    NodePoolStatusMessagePrefix.forEach((statusMessagePrefix) => {
      if (phaseMessage?.startsWith(statusMessagePrefix.originalMessage)) {
        newMessage = phaseMessage.replace(statusMessagePrefix.originalMessage, statusMessagePrefix.newMessage);
      }
    });
    if (!newMessage) {
      NodePoolStatusMessageAllowedPrefixesToKeep.forEach((statusMessagePrefix) => {
        if (phaseMessage?.startsWith(statusMessagePrefix)) {
          newMessage = phaseMessage;
        }
      });
    }
  }
  if (newMessage) {
    newMessage = stringUtil.replaceNewLineWithBreaks(newMessage);
  }
  return newMessage;
}

function _mapTelemetryTypeToNodePoolProperty(telemetryType: NodeTelemetryType): string {
  switch (telemetryType) {
    case NodeTelemetryType.TotalGpus:
      return "totalGpus";
    case NodeTelemetryType.FreeGpus:
      return "freeGpus";
    case NodeTelemetryType.ReadyGpus:
      return "readyGpus";
    case NodeTelemetryType.TotalGpuMemoryBytes:
      return "totalGpuMemoryBytes";
    case NodeTelemetryType.AllocatedGpus:
      return "allocatedGpus";
    case NodeTelemetryType.TotalCpuCores:
      return "totalCpuCores";
    case NodeTelemetryType.TotalCpuMemoryBytes:
      return "totalCpuMemoryBytes";
    case NodeTelemetryType.AllocatedCpuCores:
      return "allocatedCpuCores";
    case NodeTelemetryType.AllocatedCpuMemoryBytes:
      return "allocatedCpuMemoryBytes";
    default:
      return "";
  }
}

function _getTelemetryGroupValue(
  telemetryResponse: TelemetryResponseValuesInner,
  groupBy: ENodeTelemetryGroupBy,
): string | undefined {
  return telemetryResponse.groups?.find((group) => group.key === groupBy)?.value;
}

function _parseAndEnrichTelemetry(
  entity: INodePoolTable | INodeTable,
  telemetry: TelemetryResponse,
  telemetryData: TelemetryResponse,
  clusterId: string,
  groupBy: ENodeTelemetryGroupBy,
): void {
  telemetryData.values
    .filter(
      (telemetryResponse) =>
        _getTelemetryGroupValue(telemetryResponse, groupBy) === entity.name &&
        _getTelemetryGroupValue(telemetryResponse, ENodeTelemetryGroupBy.ClusterId) === clusterId,
    )
    .forEach((telemetryResponse) => {
      const property = _mapTelemetryTypeToNodePoolProperty(telemetry.type as NodeTelemetryType);
      entity.additionalFields = entity.additionalFields || ({} as IAdditionalNodePoolTableFields);
      entity.additionalFields[property as keyof typeof entity.additionalFields] = parseFloat(telemetryResponse.value);
    });
}

/**
 * Enriches the given node pools with telemetry data.
 *
 * @param {INodePoolTable[]} nodePools - The array of node pools to be enriched.
 * @param {TelemetryResponse[]} telemetryData - The telemetry data to enrich the node pools with.
 * @returns {INodePoolTable[]} - The enriched node pools.
 */
function enrichNodePoolsWithTelemetry(
  nodePools: INodePoolTable[],
  telemetryData: TelemetryResponse[],
): INodePoolTable[] {
  const enrichedNodePools = deepCopy(nodePools);
  return enrichedNodePools.map((nodePool) => {
    telemetryData.forEach((telemetry) =>
      _parseAndEnrichTelemetry(nodePool, telemetry, telemetry, nodePool.clusterId, ENodeTelemetryGroupBy.Nodepool),
    );
    return nodePool;
  });
}

/**
 * Enriches the given nodes with telemetry data.
 *
 * @param {INodeTable[]} nodes - The array of nodes to be enriched.
 * @param {TelemetryResponse[]} telemetryData - The telemetry data to enrich the node pools with.
 * @returns {INodePoolTable[]} - The enriched node pools.
 */
function enrichNodesWithTelemetry(nodes: INodeTable[], telemetryData: TelemetryResponse[]): INodeTable[] {
  return nodes.map((node) => {
    telemetryData.forEach((telemetry) =>
      _parseAndEnrichTelemetry(node, telemetry, telemetry, node.clusterUuid, ENodeTelemetryGroupBy.Node),
    );
    return node;
  });
}

function getEmptyNodePool(clusterId?: string): NodepoolCreateFields {
  return {
    name: "",
    labelKey: "",
    labelValue: "",
    schedulingConfiguration: {
      gpuPodsPlacementStrategy: {
        node: PlacementStrategy.Binpack,
      },
      cpuPodsPlacementStrategy: {
        node: PlacementStrategy.Binpack,
      },
    },
    clusterId: clusterId || "",
  };
}

function getNodePoolUpdateRequest(nodePool: Nodepool): NodepoolUpdateFields {
  const isDefaultNodePool = nodePool.name === DEFAULT_NODE_POOL_NAME;
  const commonFields = {
    schedulingConfiguration: nodePool.schedulingConfiguration,
    ...(nodePool.gpuNetworkAccelerationDetection && {
      gpuNetworkAccelerationDetection: nodePool.gpuNetworkAccelerationDetection,
    }),
    ...(nodePool.gpuNetworkAccelerationLabelKey && {
      gpuNetworkAccelerationLabelKey: nodePool.gpuNetworkAccelerationLabelKey,
    }),
    ...(nodePool.networkTopologyId && {
      networkTopologyId: nodePool.networkTopologyId,
    }),
  };
  if (isDefaultNodePool) return commonFields;
  return {
    ...commonFields,
    labelKey: nodePool.labelKey,
    labelValue: nodePool.labelValue,
    networkTopologyId: nodePool.networkTopologyId,
  };
}
