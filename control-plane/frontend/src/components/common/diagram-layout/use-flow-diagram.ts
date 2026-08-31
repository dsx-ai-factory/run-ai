/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { ref, nextTick } from "vue";
import type { Node } from "@vue-flow/core";
import { useVueFlow } from "@vue-flow/core";
import {
  calculateYPosition,
  calculateInitialViewport,
  measureCardHeights,
  recalculateNodePositions,
} from "./diagram-layout.util";

export interface UseFlowDiagramOptions {
  containerSelector: string;
  topOffset?: number;
}

export function useFlowDiagram(options: UseFlowDiagramOptions) {
  const { containerSelector, topOffset = 0 } = options;

  const { updateNodeInternals, fitView, getNodes, setViewport } = useVueFlow();

  const containerRef = ref<HTMLElement | null>(null);
  const expandedNodes = ref<Set<string>>(new Set());
  const flowNodes = ref<Node[]>([]);
  const cardHeights = ref<Record<string, number>>({});
  let lastNodeOrderByColumn: Record<string, string[]> | undefined;

  const isExpanded = (nodeId: string): boolean => expandedNodes.value.has(nodeId);

  const toggleExpanded = (nodeId: string): void => {
    if (expandedNodes.value.has(nodeId)) {
      expandedNodes.value.delete(nodeId);
    } else {
      expandedNodes.value.add(nodeId);
    }
    expandedNodes.value = new Set(expandedNodes.value);
    nextTick(() => measureAndUpdatePositions(lastNodeOrderByColumn));
  };

  const getNodeYPosition = (
    nodeId: string,
    _columnIndex: number,
    nodeOrderByColumn: Record<string, string[]>,
  ): number => {
    const columnKey = String(_columnIndex);
    const nodesInColumn = nodeOrderByColumn[columnKey] || [];
    return calculateYPosition(nodeId, nodesInColumn, cardHeights.value);
  };

  const fitViewByWidth = (nodeOrderByColumn: Record<string, string[]>): void => {
    const nodes = getNodes.value;
    const fallbackPadding = 0.15;
    if (nodes.length === 0 || !containerRef.value) {
      fitView({ padding: fallbackPadding });
      return;
    }

    const containerWidth = containerRef.value.clientWidth;
    const columnCount = Object.keys(nodeOrderByColumn).length;

    if (columnCount === 0) {
      fitView({ padding: fallbackPadding });
      return;
    }

    const viewport = calculateInitialViewport(containerWidth, columnCount, topOffset);
    setViewport(viewport);
  };

  const measureAndUpdatePositions = (nodeOrderByColumn?: Record<string, string[]>): void => {
    if (nodeOrderByColumn) {
      lastNodeOrderByColumn = nodeOrderByColumn;
    }
    const order = nodeOrderByColumn ?? lastNodeOrderByColumn;
    requestAnimationFrame(() => {
      requestAnimationFrame(() => {
        cardHeights.value = measureCardHeights(containerSelector);
        if (order) {
          const changedIds = recalculateNodePositions(flowNodes.value, (nodeId, colIndex) =>
            getNodeYPosition(nodeId, colIndex, order),
          );
          if (changedIds.length > 0) updateNodeInternals(changedIds);
          fitViewByWidth(order);
        }
      });
    });
  };

  return {
    containerRef,
    expandedNodes,
    flowNodes,
    cardHeights,
    isExpanded,
    toggleExpanded,
    getNodeYPosition,
    fitViewByWidth,
    measureAndUpdatePositions,
    updateNodeInternals,
  };
}
