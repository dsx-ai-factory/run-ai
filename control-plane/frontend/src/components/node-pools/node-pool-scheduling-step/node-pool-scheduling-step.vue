<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <runai-wizard-layout
    class="node-pool-scheduling-wizard"
    :steps="steps"
    active-step="scheduling"
    width="1000px"
    margin-top="26px"
    left-width="200px"
    right-width="200px"
    hide-right-on-small-screen
  >
    <runai-form-wrapper
      :form-state="{ pool: previewPool, scheduling: draft }"
      full-width
      @leave-page="emit('leave-approved')"
    >
      <runai-expansion-wrapper no-first-top-padding>
        <runai-expansion-item label="Scheduling mode" default-opened disable-closing body-no-padding>
          <template #header
            ><span>Scheduling mode</span
            ><span v-if="!canEditDepartments || !canViewDepartments" class="view-only"
              ><q-icon name="ph ph-eye" /> View only</span
            ></template
          >
          <div class="scheduling-overview">
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
                tooltip-text="The scheduling mode determines each department’s share of GPU capacity in this node pool, including how unused capacity is shared. The selection applies to the projects within each department, too."
              />
              <runai-segmented-button
                :model-value="selectedMode"
                :options="options"
                class="mode-cards"
                spread
                :disable="!canEditDepartments || !canViewDepartments"
                aria-label="Scheduling mode"
                aid="scheduling-mode-options"
                @update:model-value="selectMode"
              >
                <template #option-end="{ selected, option }">
                  <span class="mode-explanation">{{ modeExplanations[String(option.value)] }}</span>
                  <q-icon v-if="selected" name="ph ph-check" size="16px" class="mode-check" aria-hidden="true" />
                </template>
              </runai-segmented-button>
            </div>
          </div>
          <div class="department-scheduling-body">
            <q-banner v-if="loadError" class="text-negative q-ma-md" role="alert"
              >{{ loadError
              }}<template #action
                ><runai-btn aid="retry-departments" flat label="Retry" @click="emit('refresh')" /></template
            ></q-banner>
            <p v-if="telemetryError" class="q-px-md text-b-subtitle2" role="status">{{ telemetryError }}</p>
            <div class="scheduling-content">
              <node-pool-department-table
                :key="`department-draft-${resetRevision}`"
                :rows="rows"
                :baseline-rows="baselineRows"
                :draft="draft"
                :aggregates="aggregates"
                :ranks="ranks"
                :rank-counts="rankCounts"
                :capacity="resources?.gpus"
                :readonly="!canEditDepartments || !canViewDepartments"
                :can-view="canViewDepartments"
                :settings-open="showSettings"
                :totals-complete="totalsComplete"
                :raw-weights-available="pool.schedulingCapabilities?.adjustableWeights === true"
                :can-edit-department="canEditDepartment"
                :loading="!!loading"
                :has-more="hasMore"
                :has-error="!!loadError"
                :search-term="searchTerm"
                :rank-filter="rankFilter"
                :load-quota-context="loadQuotaContext"
                @edit="applyEdit"
                @search="emit('search', $event)"
                @filter-rank="emit('filter-rank', $event)"
                @load-more="emit('load-more')"
                @load-rank="emit('load-rank', $event)"
                @refresh="emit('refresh')"
                @automatic-weights="draft.automaticWeights = $event"
                @advanced-settings="showSettings = !showSettings"
                @new-department="showNewDepartment = true"
              >
                <template #settings
                  ><node-pool-scheduling-settings
                    v-if="showSettings"
                    :draft="draft"
                    :quota-policy="quotaPolicy"
                    :readonly="!canEditDepartments"
                    @close="showSettings = false"
                    @patch="Object.assign(draft, $event)"
                /></template>
              </node-pool-department-table>
            </div>
            <node-pool-new-department-dialog
              v-if="showNewDepartment && canEditDepartments && canViewDepartments"
              :cluster-id="pool.clusterId"
              :cluster-name="pool.clusterName"
              @close="showNewDepartment = false"
            />
          </div>
        </runai-expansion-item>
      </runai-expansion-wrapper>
      <runai-base-dialog
        v-if="pendingMode"
        :model-value="true"
        size="width-sm"
        custom-class="scheduling-confirm"
        @close="cancelModeChange"
      >
        <template #header>Change Mode?</template>
        <template #body><div>You have unsaved changes. If you change the mode, they will be lost.</div></template>
        <template #footer>
          <runai-btn aid="stay-scheduling-mode" flat label="Stay" color="primary" @click="cancelModeChange" />
          <runai-btn aid="confirm-scheduling-mode" label="Change mode" color="primary" @click="confirmModeChange" />
        </template>
      </runai-base-dialog>
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
      <node-pool-wizard-summary scheduling :pool="previewPool" :resources="resources" />
    </template>
  </runai-wizard-layout>
</template>

<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { dateUtil } from "@/utils/date.util/date.util";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import NodePoolDepartmentTable from "./node-pool-department-table.vue";
import NodePoolSchedulingSettings from "./node-pool-scheduling-settings.vue";
import NodePoolNewDepartmentDialog from "./node-pool-new-department-dialog.vue";
import { useNodePoolSchedulingDraft } from "./use-node-pool-scheduling-draft.composable";
import {
  toSchedulingCapabilities,
  type SchedulingMode,
  type DepartmentSchedulingEdit,
  type SchedulingTotals,
  type DepartmentValidationContext,
} from "./node-pool-scheduling.model";
import type { DepartmentNodePoolResources } from "@/swagger-models/org-unit-service-client";
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
} from "@/models/node-pool.model";
import type { Nodepool } from "@/swagger-models/cluster-service-client";

const props = withDefaults(
  defineProps<{
    pool: Nodepool;
    resources?: INodePoolSummaryResources;
    rows?: DepartmentNodePoolResources[];
    baselineRows?: DepartmentNodePoolResources[];
    aggregates?: Record<string, SchedulingTotals>;
    ranks?: string[];
    rankCounts?: Record<string, number>;
    canViewDepartments?: boolean;
    canEditDepartments?: boolean;
    totalsComplete?: boolean;
    canEditDepartment?: (id: string) => boolean;
    loading?: boolean;
    loadError?: string;
    telemetryError?: string;
    hasMore?: boolean;
    searchTerm?: string;
    rankFilter?: string;
    quotaPolicy?: boolean;
    loadQuotaContext?: (id: string) => Promise<DepartmentValidationContext>;
  }>(),
  {
    rows: () => [],
    aggregates: () => ({}),
    ranks: () => [],
    canViewDepartments: true,
    canEditDepartments: true,
    totalsComplete: true,
  },
);
const emit = defineEmits<{
  "leave-approved": [];
  refresh: [];
  search: [string];
  "filter-rank": [string | undefined];
  "load-more": [];
  "load-rank": [string];
}>();
const showSettings = ref(false);
const showNewDepartment = ref(false);
const modeExplanations: Record<string, string> = {
  proportional: "GPU resources are distributed proportionally according to each department’s and project’s weight.",
  ranked:
    "Higher-ranked departments and projects receive GPU resources first. Remaining capacity goes to lower-ranked ones, in rank order.",
  guaranteed: "Each department has a guaranteed GPU quota. Any unused quota is redistributed by rank and weight.",
};
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
function durationFields(value: string | null | undefined) {
  const seconds = dateUtil.parseDurationStringToSeconds(value ?? "");
  return {
    Days: Math.floor(seconds / 86400),
    Hours: Math.floor((seconds % 86400) / 3600),
    Minutes: Math.floor((seconds % 3600) / 60),
    Seconds: seconds % 60,
  };
}
const savedMode = getSchedulingMode(props.pool.schedulingCapabilities);
const mode = (
  { Proportional: "proportional", Ranked: "ranked", "Guaranteed quota": "guaranteed" } as Record<string, SchedulingMode>
)[savedMode];
const { draft, pendingMode, resetRevision, requestMode, confirmModeChange, cancelModeChange } =
  useNodePoolSchedulingDraft({
    mode,
    automaticWeights: props.pool.schedulingCapabilities?.adjustableWeights === false,
    cpuQuota: props.pool.schedulingCapabilities?.cpuQuota === true,
    maxAllocation: false,
    timeFairshare: props.pool.schedulingConfiguration?.timeBasedFairShare?.enabled === true,
    historicalWeight: props.pool.schedulingConfiguration?.timeBasedFairShare?.historicalUsageWeight ?? null,
    historicalWindow: durationFields(props.pool.schedulingConfiguration?.timeBasedFairShare?.windowDuration),
    minimumRuntime: durationFields(props.pool.schedulingConfiguration?.minGuaranteedRuntime),
    capabilities: props.pool.schedulingCapabilities,
    departments: {},
  });
watch(
  () => draft.value.mode,
  (mode, previous) => {
    if (mode === "proportional" && previous && previous !== "proportional") emit("filter-rank", undefined);
  },
);
const selectedMode = computed(() => draft.value.mode ?? null);
const previewPool = computed(() => ({
  ...props.pool,
  schedulingCapabilities: draft.value.mode ? toSchedulingCapabilities(draft.value) : draft.value.capabilities,
}));
function selectMode(value: string | number | boolean | null) {
  if (value === "proportional" || value === "ranked" || value === "guaranteed") requestMode(value);
}
function applyEdit(edit: DepartmentSchedulingEdit) {
  if (!props.canEditDepartments) return;
  draft.value.departments[edit.id] = { ...draft.value.departments[edit.id], ...edit };
}
</script>

<style scoped lang="scss">
.node-pool-scheduling-wizard :deep(.wizard-step-item) {
  padding-top: 7px;
  padding-bottom: 7px;
}
.node-pool-scheduling-wizard :deep(.wizard-step-title) {
  white-space: nowrap;
  font-weight: 400;
}
.node-pool-scheduling-wizard :deep(.wizard-step--disabled .wizard-step-item) {
  opacity: 1 !important;
}
.scheduling-overview {
  padding: 18px 19px 40px;
}
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
  margin-top: 36px;
  margin-bottom: 0;
}
.node-pool-scheduling-wizard :deep(.runai-expansion-header) {
  height: 46px;
}
.mode-selector :deep(.runai-section-title) {
  margin-bottom: 10px;
}
.mode-cards {
  align-items: stretch;
  border: 0;
  padding: 0;
  gap: 7px;

  :deep(.runai-segment) {
    position: relative;
    min-width: 0;
    min-height: 61px;
    height: 61px;
    overflow: visible;
    border: 1px solid var(--border-subtle);
    border-radius: 5px;
    padding: 10px 16px;
    gap: 10px;
    text-align: left;
  }
  :deep(.runai-segment--selected) {
    border-color: var(--border-pressed);
    box-shadow: inset 0 0 0 1px var(--border-pressed);
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
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 40px 0 8px;
}
.preview-notice {
  flex: 1;
  margin: 0;
  text-align: left;
  font-size: 12px;
}
.scheduling-content {
  display: block;
}
.scheduling-content > .department-section {
  flex: 1;
  min-width: 0;
}
.mode-explanation {
  display: none;
  position: absolute;
  top: 57px;
  left: -1px;
  right: -1px;
  padding: 4px 14px 7px;
  min-height: 64px;
  white-space: normal;
  background: white;
  border: 1px solid var(--border-subtle);
  border-top: 0;
  border-radius: 0 0 4px 4px;
  color: var(--b-text-sub);
  font-size: 12px;
  line-height: 16.8px;
  z-index: 5;
}
.mode-cards :deep(.runai-segment:hover),
.mode-cards :deep(.runai-segment:focus-visible) {
  border-radius: 4px 4px 0 0;
}
.mode-cards :deep(.runai-segment:hover) .mode-explanation,
.mode-cards :deep(.runai-segment:focus-visible) .mode-explanation {
  display: block;
}
.mode-cards :deep(.runai-segment--selected) .mode-explanation {
  border: 2px solid var(--border-pressed);
  border-top: 0;
  left: -2px;
  right: -2px;
}
.view-only {
  margin-left: 10px;
  border: 1px solid var(--border-default);
  color: var(--b-text-sub);
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 2px;
}
@media (min-width: 1240px) and (max-width: 1439px) {
  .node-pool-scheduling-wizard {
    width: 1220px;
  }
  .node-pool-scheduling-wizard :deep(.runai-panels-layout) {
    grid-template-columns: 200px minmax(0, 1fr);
  }
  .node-pool-scheduling-wizard :deep(.panel-right) {
    display: none !important;
  }
}
.mode-cards :deep(.runai-segment--disabled) {
  background: #eee;
}
.mode-cards :deep(.runai-segment--selected.runai-segment--disabled) {
  border-color: var(--border-pressed);
  color: var(--interactive-secondary-default-text);
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
