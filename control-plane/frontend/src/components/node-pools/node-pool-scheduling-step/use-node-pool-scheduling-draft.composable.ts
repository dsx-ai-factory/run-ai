/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
import { computed, ref } from "vue";
import { createSchedulingDraft, type SchedulingMode, type SchedulingSnapshot } from "./node-pool-scheduling.model";

export function useNodePoolSchedulingDraft(initial: SchedulingSnapshot) {
  let baseline = createSchedulingDraft(initial);
  const draft = ref(createSchedulingDraft(initial));
  const pendingMode = ref<SchedulingMode>();
  const resetRevision = ref(0);
  const isDirty = computed(() => JSON.stringify(draft.value) !== JSON.stringify(baseline));
  function requestMode(mode: SchedulingMode) {
    // Nov 12 requirements supersede the original discard-on-mode-change design.
    // Keep inactive values in the draft; request mapping omits inactive fields.
    draft.value.mode = mode;
  }
  function confirmModeChange() {
    if (pendingMode.value) requestMode(pendingMode.value);
    pendingMode.value = undefined;
  }
  function cancelModeChange() {
    pendingMode.value = undefined;
  }
  function acceptSnapshot(snapshot: SchedulingSnapshot) {
    baseline = createSchedulingDraft(snapshot);
    draft.value = createSchedulingDraft(snapshot);
    pendingMode.value = undefined;
  }
  return {
    draft,
    isDirty,
    pendingMode,
    resetRevision,
    requestMode,
    confirmModeChange,
    cancelModeChange,
    acceptSnapshot,
  };
}
