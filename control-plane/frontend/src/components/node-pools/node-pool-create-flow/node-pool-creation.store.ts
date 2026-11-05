/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { defineStore } from "pinia";
import type { Nodepool } from "@/swagger-models/cluster-service-client";
import type { NodePoolCreationDraft } from "./node-pool-create-flow.model";

// Keep the created identity and draft available when the scheduling step takes over.
export const useNodePoolCreationStore = defineStore("node-pool-creation", {
  state: () => ({ handoff: null as { pool: Nodepool; draft: NodePoolCreationDraft } | null }),
});
