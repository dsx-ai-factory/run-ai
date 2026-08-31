/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { ENodePoolsListOrigin, IComputeSectionNodePoolsPolicy } from "./node-pools-resource-section";
import type { IAutoScaleData } from "./auto-scale-section";
import type { PodAffinity, Toleration } from "@/swagger-models/assets-service-client";

export interface IComputeSectionNodePoolData {
  defaultNodePools: Array<string>;
  allNodePoolsOptions: Array<string>;
  nodePoolsListOrigin: ENodePoolsListOrigin;
}

export interface IComputeSectionPolicy {
  nodePools?: IComputeSectionNodePoolsPolicy;
}

export interface IWorkloadComputeSectionOptions {
  cardsTitle?: string;
  applyPolicyDefaults?: boolean;
  hideNodePools?: boolean;
  hideCardsActions?: boolean;
  autoScale?: boolean;
  showTopologySection?: boolean;
  allowAnyEffect?: boolean;
  latencyEnabled?: boolean;
  autoScaleEnabled?: boolean; // this flag enable/ disable to chose the numbers of replicas
}

export interface IComputeSectionData {
  computeResourceId: string | null;
  nodeType: string | null;
  nodePools?: IComputeSectionNodePoolData | null;
  minWorkers?: number | null;
  workers?: number | null;
  autoScaleData?: IAutoScaleData;
  tolerations?: Array<Toleration> | null;
  podAffinity?: PodAffinity | null;
}
