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
  let modeBaseline = createSchedulingDraft(initial);
  const hasConfigurationEdits = () => JSON.stringify(draft.value) !== JSON.stringify(modeBaseline);
  function applyMode(mode: SchedulingMode) {
    draft.value = createSchedulingDraft(baseline);
    draft.value.mode = mode;
    draft.value.automaticWeights = mode === "guaranteed";
    draft.value.cpuQuota = false;
    modeBaseline = createSchedulingDraft(draft.value);
    resetRevision.value++;
  }
  function requestMode(mode: SchedulingMode) {
    if (mode === draft.value.mode) return;
    if (hasConfigurationEdits()) pendingMode.value = mode;
    else applyMode(mode);
  }
  function confirmModeChange() {
    if (pendingMode.value) applyMode(pendingMode.value);
    pendingMode.value = undefined;
  }
  function cancelModeChange() {
    pendingMode.value = undefined;
  }
  function acceptSnapshot(snapshot: SchedulingSnapshot) {
    baseline = createSchedulingDraft(snapshot);
    modeBaseline = createSchedulingDraft(snapshot);
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
