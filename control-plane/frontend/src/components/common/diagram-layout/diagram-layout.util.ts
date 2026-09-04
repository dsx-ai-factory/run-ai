/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { Node } from "@vue-flow/core";

export const COLUMN_WIDTH = 250;
export const COLUMN_GAP = 20;
export const NODE_GAP = 15;
export const DEFAULT_CARD_HEIGHT = 80;

export const HORIZONTAL_PADDING = 12;
export const MIN_ZOOM = 0.3;
export const MAX_ZOOM = 1.5;
export const DEFAULT_ENDPOINT_HANDLE_ID = "left";

export const calculateXPosition = (columnIndex: number): number => {
  return columnIndex * (COLUMN_WIDTH + COLUMN_GAP);
};

export const calculateYPosition = (
  nodeId: string,
  nodesInColumn: string[],
  cardHeights: Record<string, number>,
): number => {
  const nodeIndex = nodesInColumn.indexOf(nodeId);
  if (nodeIndex <= 0) return 0;

  let y = 0;
  for (let i = 0; i < nodeIndex; i++) {
    const prevNodeId = nodesInColumn[i];
    const height = cardHeights[prevNodeId] || DEFAULT_CARD_HEIGHT;
    y += height + NODE_GAP;
  }
  return y;
};

export const calculateContentWidth = (columnCount: number): number => {
  if (columnCount <= 0) return 0;
  return columnCount * COLUMN_WIDTH + (columnCount - 1) * COLUMN_GAP;
};

export const calculateZoomToFitWidth = (containerWidth: number, columnCount: number): number => {
  const contentWidth = calculateContentWidth(columnCount);
  if (contentWidth <= 0) return 1;

  const availableWidth = containerWidth - HORIZONTAL_PADDING * 2;
  const zoomToFit = availableWidth / contentWidth;

  return Math.max(MIN_ZOOM, Math.min(MAX_ZOOM, zoomToFit));
};

export interface ViewportConfig {
  x: number;
  y: number;
  zoom: number;
}

export const calculateInitialViewport = (containerWidth: number, columnCount: number, topOffset = 0): ViewportConfig => {
  const zoom = calculateZoomToFitWidth(containerWidth, columnCount);

  return {
    x: HORIZONTAL_PADDING,
    y: topOffset,
    zoom,
  };
};

export const measureCardHeights = (containerSelector: string): Record<string, number> => {
  const cards = document.querySelectorAll<HTMLElement>(`${containerSelector} [data-node-id]`);
  const heights: Record<string, number> = {};

  cards.forEach((card) => {
    const nodeId = card.getAttribute("data-node-id");
    if (nodeId) heights[nodeId] = card.offsetHeight;
  });

  return heights;
};

export const recalculateNodePositions = (
  flowNodes: Node[],
  getNodeYPosition: (nodeId: string, columnIndex: number) => number,
): string[] => {
  const changedIds: string[] = [];

  flowNodes.forEach((node) => {
    const columnIndex = (node.data?.columnIndex as number) ?? 0;
    const newY = getNodeYPosition(node.id, columnIndex);

    if (Math.abs(node.position.y - newY) > 1) {
      node.position = { x: calculateXPosition(columnIndex), y: newY };
      changedIds.push(node.id);
    }
  });

  return changedIds;
};
