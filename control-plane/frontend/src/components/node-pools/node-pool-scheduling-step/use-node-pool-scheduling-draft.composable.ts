/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { ENodePoolSchedulingModeId as ESchedulingMode } from "@/models/node-pool.model";
import { computed, ref } from "vue";
import { createSchedulingDraft, type SchedulingMode, type ISchedulingSnapshot } from "./node-pool-scheduling.model";

export function useNodePoolSchedulingDraft(initial: ISchedulingSnapshot) {
  let baseline = createSchedulingDraft(initial);
  const draft = ref(createSchedulingDraft(initial));
  const pendingMode = ref<SchedulingMode>();
  const isDirty = computed(() => JSON.stringify(draft.value) !== JSON.stringify(baseline));
  function requestMode(mode: SchedulingMode) {
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
  function confirmModeChange() {
    if (pendingMode.value) requestMode(pendingMode.value);
    pendingMode.value = undefined;
  }
  function cancelModeChange() {
    pendingMode.value = undefined;
  }
  function acceptSnapshot(snapshot: ISchedulingSnapshot) {
    baseline = createSchedulingDraft(snapshot);
    draft.value = createSchedulingDraft(snapshot);
    pendingMode.value = undefined;
  }
  return {
    draft,
    isDirty,
    pendingMode,
    requestMode,
    confirmModeChange,
    cancelModeChange,
    acceptSnapshot,
  };
}
