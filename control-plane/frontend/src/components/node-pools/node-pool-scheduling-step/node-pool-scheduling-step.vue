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
    <runai-form-wrapper
      :form-state="{ pool: previewPool, departments: draft.departments }"
      full-width
      @leave-page="emit('leave-approved')"
    >
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
        <div v-if="!canViewDepartments" class="q-pa-lg" role="status">
          You do not have permission to view departments.
        </div>
        <template v-else>
          <q-banner v-if="loadError" class="text-negative q-ma-md" role="alert"
            >{{ loadError
            }}<template #action
              ><runai-btn aid="retry-departments" flat label="Retry" @click="emit('refresh')" /></template
          ></q-banner>
          <p v-if="telemetryError" class="q-px-md text-b-subtitle2" role="status">{{ telemetryError }}</p>
          <node-pool-department-table
            :key="`department-draft-${resetRevision}`"
            :rows="rows"
            :draft="draft"
            :aggregates="aggregates"
            :ranks="ranks"
            :rank-counts="rankCounts"
            :capacity="resources?.gpus"
            :readonly="!canEditDepartments"
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
          />
          <p v-if="draft.mode === 'guaranteed'" class="q-pa-md text-b-subtitle2">
            Limit projects from exceeding department quota:
            {{ quotaPolicy === undefined ? "Unknown" : quotaPolicy ? "On" : "Off" }}. This organization-wide policy is
            read-only here.
          </p>
        </template>
      </runai-expansion-wrapper>
      <runai-base-dialog v-if="pendingMode" :model-value="true" size="width-sm" @close="cancelModeChange">
        <template #header>Change Mode?</template>
        <template #body
          ><div class="q-px-md q-py-lg">
            You have unsaved changes. If you change the mode, they will be lost.
          </div></template
        >
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
      <node-pool-wizard-summary :pool="previewPool" :resources="resources" />
    </template>
  </runai-wizard-layout>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import NodePoolDepartmentTable from "./node-pool-department-table.vue";
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
} from "../node-pool-wizard-summary/node-pool-wizard-summary.model";
import type { Nodepool } from "@/swagger-models/cluster-service-client";

const props = withDefaults(
  defineProps<{
    pool: Nodepool;
    resources?: INodePoolSummaryResources;
    rows?: DepartmentNodePoolResources[];
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
const savedMode = getSchedulingMode(props.pool.schedulingCapabilities);
const mode = (
  { Proportional: "proportional", Ranked: "ranked", "Guaranteed quota": "guaranteed" } as Record<string, SchedulingMode>
)[savedMode];
const { draft, pendingMode, resetRevision, requestMode, confirmModeChange, cancelModeChange } =
  useNodePoolSchedulingDraft({
    mode,
    automaticWeights: props.pool.schedulingCapabilities?.adjustableWeights === false,
    cpuQuota: props.pool.schedulingCapabilities?.cpuQuota === true,
    capabilities: props.pool.schedulingCapabilities,
    departments: {},
  });
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
