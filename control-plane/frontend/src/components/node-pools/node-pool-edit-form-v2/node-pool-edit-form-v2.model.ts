/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import type { NodepoolCreateFields } from "@/swagger-models/cluster-service-client";

export interface INodePoolFormDraft extends NodepoolCreateFields {
  description: string;
}
export enum ENodePoolFormDestination {
  Close = "close",
  Continue = "continue",
}
