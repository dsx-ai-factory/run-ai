/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { ArrayRules } from "@/swagger-models/policy-service-client";

export enum ENodePoolsListOrigin {
  Project = "project",
  Template = "template",
}

export interface IComputeSectionNodePoolsPolicy {
  rules?: ArrayRules;
  defaults?: Array<string>;
}
