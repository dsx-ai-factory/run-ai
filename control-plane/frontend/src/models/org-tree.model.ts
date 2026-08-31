/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { QTreeNode } from "quasar";
import type { ScopeType } from "@/swagger-models/authorization-client";

// Aligned with the icons catalog (`runaiIcons.entity.*`) — Yuval UX review.
// Cluster uses the filled custom runai-cluster glyph (catalog: entity.cluster).
// Tenant (Organization) uses `ph-network` — matches Figma nav/organization.
export enum EOrgTreeNodeIcons {
  Tenant = "ph-fill ph-network",
  Cluster = "runai-filled-cluster",
  Department = "ph-fill ph-garage",
  Project = "ph-fill ph-suitcase",
}
export interface IOrgTreeNode extends QTreeNode {
  id: string;
  path: string;
  type: ScopeType;
  children: IOrgTreeNode[];
}

export interface IOrgTreeNodeId {
  path?: string;
  id: string;
  type: ScopeType;
}
