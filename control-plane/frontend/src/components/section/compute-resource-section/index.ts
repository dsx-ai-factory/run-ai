/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import ComputeResourceSection from "@/components/section/compute-resource-section/compute-resource-section.vue";
import type {
  IComputeSectionNodePoolData,
  IComputeSectionPolicy,
  IWorkloadComputeSectionOptions,
  IComputeSectionData,
} from "./compute-resource-section.models";
import { ENodePoolsListOrigin } from "./node-pools-resource-section";
import type { IComputeSectionNodePoolsPolicy } from "./node-pools-resource-section";
import type { IAutoScaleData } from "./auto-scale-section";

export {
  ComputeResourceSection,
  type IComputeSectionNodePoolData,
  ENodePoolsListOrigin,
  type IComputeSectionPolicy,
  type IComputeSectionNodePoolsPolicy,
  type IWorkloadComputeSectionOptions,
  type IComputeSectionData,
  type IAutoScaleData,
};
