/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { ENodePoolSchedulingMode as ESchedulingMode } from "@/models/node-pool.model";
import { isSupportedSchedulingMode } from "@/models/node-pool.model";
import { ref } from "vue";
import { createSchedulingDraft, type ISchedulingSnapshot } from "@/models/node-pool-scheduling.model";

export function useNodePoolSchedulingDraft(initial: ISchedulingSnapshot) {
  const draft = ref(createSchedulingDraft(initial));
  function requestMode(mode: ESchedulingMode) {
    if (!isSupportedSchedulingMode(mode)) return;
    // Preserve other edits; maximum allocations belong only to Guaranteed quota.
    if (mode !== ESchedulingMode.Guaranteed) {
      const departments = Object.fromEntries(
        Object.entries(draft.value.departments).map(([id, edit]) => {
          const next = { ...edit };
          delete next.limit;
          return [id, next];
        }),
      );
      draft.value = { ...draft.value, mode, maxAllocation: false, departments };
      return;
    }
    draft.value.mode = mode;
  }
  return { draft, requestMode };
}
