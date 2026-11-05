<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <runai-wizard-layout
    :steps="steps"
    active-step="scheduling"
    width="1000px"
    margin-top="26px"
    left-width="200px"
    right-width="200px"
    hide-right-on-small-screen
  >
    <runai-form-wrapper :form-state="previewPool" full-width @leave-page="emit('leave-approved')">
      <runai-expansion-wrapper no-first-top-padding>
        <runai-expansion-item label="Scheduling mode" default-opened disable-closing>
          <div class="capacity-label text-b-subtitle1">Total node pool capacity</div>
          <div class="capacity q-mt-xs" aid="pool-gpu-capacity">
            <template v-if="resources?.gpus != null">
              <span class="text-h5 text-weight-medium">{{ resources.gpus }}</span> <span>GPUs</span>
            </template>
            <span v-else class="text-h5 text-weight-medium">Unknown</span>
          </div>
          <div class="mode-selector">
            <runai-section-title
              title="Set how GPU capacity is split between departments in this node pool"
              tooltip-text="Choose how this node pool distributes GPU capacity between departments."
            />
            <runai-segmented-button
              :model-value="selectedMode"
              :options="options"
              class="mode-cards"
              spread
              aria-label="Scheduling mode"
              aid="scheduling-mode-options"
              @update:model-value="selectMode"
            >
              <template #option-end="{ selected }">
                <q-icon v-if="selected" name="ph ph-check" size="16px" class="mode-check" aria-hidden="true" />
              </template>
            </runai-segmented-button>
          </div>
        </runai-expansion-item>
      </runai-expansion-wrapper>
    </runai-form-wrapper>
    <template #content-footer>
      <div class="scheduling-footer">
        <!-- Temporary review phase: selection is preview-only until department editing and saving are added. -->
        <p class="text-b-body1 preview-notice" role="status">Changes are not saved in this preview.</p>
        <div class="row justify-end q-gutter-md">
          <runai-btn aid="scheduling-save-close" label="Save & close" flat color="primary" :disable="true" />
          <runai-btn aid="scheduling-save-continue" label="Save & continue" color="primary" :disable="true" />
        </div>
      </div>
    </template>
    <template #right-sidebar>
      <node-pool-wizard-summary :pool="previewPool" :resources="resources" />
    </template>
  </runai-wizard-layout>
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import RunaiWizardLayout from "@/components/common/runai-wizard-layout/runai-wizard-layout.vue";
import RunaiFormWrapper from "@/components/common/runai-form-wrapper/runai-form-wrapper.vue";
import RunaiExpansionWrapper from "@/components/common/runai-expansion-wrapper/runai-expansion-wrapper.vue";
import RunaiExpansionItem from "@/components/common/runai-expansion-item/runai-expansion-item.vue";
import RunaiSectionTitle from "@/components/common/runai-section-title/runai-section-title.vue";
import RunaiBtn from "@/components/common/runai-btn/runai-btn.vue";
import { RunaiSegmentedButton } from "@/components/common/runai-segmented-button";
import NodePoolWizardSummary from "../node-pool-wizard-summary/node-pool-wizard-summary.vue";
import {
  getSchedulingMode,
  type INodePoolSummaryResources,
} from "../node-pool-wizard-summary/node-pool-wizard-summary.model";
import type { Nodepool, SchedulingCapabilities } from "@/swagger-models/cluster-service-client";
import { PROPORTIONAL_CAPABILITIES } from "../node-pool-create-flow/node-pool-create-flow.model";

const props = defineProps<{ pool: Nodepool; resources?: INodePoolSummaryResources }>();
const emit = defineEmits<{ "leave-approved": [] }>();
const steps = [
  { name: "identity", title: "Cluster & name", done: true, disabled: true },
  { name: "scheduling", title: "Scheduling mode", done: false },
  { name: "topology", title: "Topology & placement", done: false, disabled: true },
];
const options = [
  {
    value: "proportional",
    label: "Proportional",
    description: "Weight-based distribution",
    icon: "node-pool-proportional",
    aid: "mode-proportional",
  },
  {
    value: "ranked",
    label: "Ranked",
    description: "Rank-based distribution",
    icon: "node-pool-ranked",
    aid: "mode-ranked",
  },
  {
    value: "guaranteed",
    label: "Guaranteed quota",
    description: "Guaranteed quota, redistribute unused",
    icon: "node-pool-guaranteed",
    aid: "mode-guaranteed",
  },
];
const presets = new Map<string, SchedulingCapabilities>([
  ["proportional", { ...PROPORTIONAL_CAPABILITIES }],
  ["ranked", { adjustableWeights: true, quota: false, ranks: true, cpuQuota: false }],
  ["guaranteed", { adjustableWeights: false, quota: true, ranks: true, cpuQuota: false }],
]);
const capabilities = ref(props.pool.schedulingCapabilities ? { ...props.pool.schedulingCapabilities } : undefined);
const selectedMode = computed(() => {
  const mode = getSchedulingMode(capabilities.value);
  return options.find((option) => option.label === mode)?.value ?? null;
});
const previewPool = computed(() => ({ ...props.pool, schedulingCapabilities: capabilities.value }));
function selectMode(value: string | number | boolean | null) {
  const preset = typeof value === "string" ? presets.get(value) : undefined;
  if (preset) capabilities.value = { ...preset };
}
</script>

<style scoped lang="scss">
.capacity-label,
.preview-notice {
  color: var(--b-text-sub);
}
.capacity {
  display: flex;
  align-items: baseline;
  gap: 10px;
}
.mode-selector {
  margin-top: 48px;
  margin-bottom: 20px;
}
.mode-cards {
  align-items: stretch;
  border: 0;
  padding: 0;
  gap: 7px;

  :deep(.runai-segment) {
    position: relative;
    min-width: 0;
    border: 1px solid var(--border-subtle);
    border-radius: 5px;
    padding: 10px 16px;
    gap: 10px;
    text-align: left;
  }
  :deep(.runai-segment--selected) {
    border-color: var(--border-pressed);
    box-shadow: inset 0 0 0 2px var(--border-pressed);
  }
  :deep(.runai-segment__label) {
    font-weight: 500;
    padding-right: 8px;
  }
  :deep(.runai-segment__description) {
    color: var(--b-text-sub);
    margin-top: 3px;
    line-height: 18px;
  }
  :deep(.runai-segment__text) {
    align-items: flex-start;
  }
  :deep(.runai-segment__icon path) {
    fill: currentColor;
  }
}
.mode-check {
  position: absolute;
  top: 0;
  right: 0;
  width: 20px;
  height: 20px;
  border-radius: 3px;
  background: var(--border-pressed);
  color: white;
}
.scheduling-footer {
  padding: 24px 0 8px;
}
.preview-notice {
  text-align: right;
}
@media (max-width: 767px) {
  .mode-cards {
    flex-direction: column;
    :deep(.runai-segment) {
      width: 100%;
    }
  }
}
</style>
