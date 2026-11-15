<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <runai-wizard-layout
    class="node-pool-scheduling-wizard"
    responsive
    :steps="steps"
    active-step="scheduling"
    width="1000px"
    margin-top="26px"
    left-width="200px"
    right-width="200px"
    hide-right-on-small-screen
  >
    <runai-form-wrapper
      :form-state="{ pool: previewPool, scheduling: draft, saving }"
      full-width
      @leave-page="emit('leave-approved')"
    >
      <fieldset :disabled="saving" class="scheduling-fields">
        <runai-expansion-wrapper no-first-top-padding>
          <runai-expansion-item label="Scheduling mode" default-opened disable-closing body-no-padding>
            <template #header
              ><span>Scheduling mode</span
              ><span v-if="!canEditDepartments || !canViewDepartments" class="view-only"
                ><q-icon name="ph-fill ph-eye" /> View only</span
              ></template
            >
            <div class="scheduling-overview">
              <div class="capacity-label text-b-subtitle1">Total node pool capacity</div>
              <div class="capacity" aid="pool-gpu-capacity">
                <template v-if="resources?.gpus != null">
                  <span class="capacity-number">{{ resources.gpus }}</span> <span class="capacity-unit">GPUs</span>
                </template>
                <span v-else class="capacity-number">Unknown</span>
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
                  :disable="saving || !canEditDepartments || !canViewDepartments"
                  aria-label="Scheduling mode"
                  aid="scheduling-mode-options"
                  @update:model-value="selectMode"
                >
                  <template #option-end="{ selected, option }">
                    <span class="mode-explanation">{{ modeExplanations[String(option.value)] }}</span>
                    <q-icon v-if="selected" name="ph-bold ph-check" size="16px" class="mode-check" aria-hidden="true" />
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
                  :pool-id="pool.id"
                  :rows="rows"
                  :baseline-rows="baselineRows"
                  :draft="draft"
                  :aggregates="aggregates"
                  :ranks="ranks"
                  :rank-counts="rankCounts"
                  :capacity="resources?.gpus"
                  :readonly="saving || !canEditDepartments || !canViewDepartments"
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
                  @close-settings="showSettings = false"
                  @new-department="showNewDepartment = true"
                >
                  <template #settings
                    ><node-pool-scheduling-settings
                      v-if="showSettings"
                      :draft="draft"
                      :readonly="saving || !canEditDepartments"
                      :can-reset-resources="canResetResources"
                      :saved-max-allocation="!!pool.schedulingCapabilities?.limits"
                      :saved-cpu-quota="!!pool.schedulingCapabilities?.cpuQuota"
                      @close="showSettings = false"
                      @patch="!saving && Object.assign(draft, $event)"
                  /></template>
                </node-pool-department-table>
              </div>
              <node-pool-new-department-dialog
                v-if="showNewDepartment && canEditDepartments && canViewDepartments"
                :cluster-id="pool.clusterId"
                :cluster-name="pool.clusterName"
                @close="showNewDepartment = false"
                @created="emit('refresh')"
                @saving="emit('saving', $event)"
              />
            </div>
          </runai-expansion-item>
        </runai-expansion-wrapper>
      </fieldset>
    </runai-form-wrapper>
    <template #content-footer>
      <div class="scheduling-footer">
        <p v-if="saveError || validationError" class="text-negative" role="alert">{{ saveError || validationError }}</p>
        <p class="text-b-body1 preview-notice" role="status">
          Topology & placement is not available yet. Use Save & close to save scheduling.
        </p>
        <div class="row justify-end q-gutter-md">
          <runai-btn
            aid="scheduling-save-close"
            label="Save & close"
            flat
            color="primary"
            :loading="saving"
            :disable="saving || loading || !canSavePool || !canEditDepartments || !!validationError"
            @click="save"
          />
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
import { createSchedulingSaver, buildSchedulingSave } from "./node-pool-scheduling-save";
import { patchSchedulingPool, patchSchedulingDepartments } from "./node-pool-scheduling.service";
import { extractErrorMessages } from "@/utils/error-alert.util";
import { dateUtil } from "@/utils/date.util/date.util";
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
import { getSchedulingMode, type INodePoolSummaryResources } from "@/models/node-pool.model";
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
    canSavePool?: boolean;
    canResetResources?: boolean;
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
  saved: [Nodepool];
  "pool-saved": [Nodepool];
  saving: [boolean];
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
const options = computed(() => [
  {
    value: "proportional",
    disable: props.pool.schedulingCapabilities?.quota === true,
    disabledTooltip: "An existing Guaranteed quota pool cannot switch to a lower scheduling mode.",
    label: "Proportional",
    description: "Weight-based distribution",
    icon: "node-pool-proportional",
    aid: "mode-proportional",
  },
  {
    value: "ranked",
    disable: props.pool.schedulingCapabilities?.quota === true,
    disabledTooltip: "An existing Guaranteed quota pool cannot switch to a lower scheduling mode.",
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
]);
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
const { draft, resetRevision, requestMode } = useNodePoolSchedulingDraft({
  mode,
  automaticWeights: props.pool.schedulingCapabilities?.adjustableWeights === false,
  cpuQuota: props.pool.schedulingCapabilities?.cpuQuota === true,
  maxAllocation: props.pool.schedulingCapabilities?.limits === true,
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
  if (saving.value || (props.pool.schedulingCapabilities?.quota && value !== "guaranteed")) return;
  if (value === "proportional" || value === "ranked" || value === "guaranteed") requestMode(value);
}
function applyEdit(edit: DepartmentSchedulingEdit) {
  if (saving.value || !props.canEditDepartments) return;
  draft.value.departments[edit.id] = { ...draft.value.departments[edit.id], ...edit };
}

const saving = ref(false);
const saveError = ref("");
const saver = createSchedulingSaver({ patchPool: patchSchedulingPool, patchDepartments: patchSchedulingDepartments });
const validationError = computed(() => {
  try {
    buildSchedulingSave(props.pool, draft.value);
    const capabilities = toSchedulingCapabilities(draft.value);
    if (
      !props.canResetResources &&
      ((props.pool.schedulingCapabilities?.cpuQuota && !capabilities.cpuQuota) ||
        (props.pool.schedulingCapabilities?.limits && !capabilities.limits))
    )
      return "Turning off CPU quota or Max GPU allocation requires update permission for all departments and projects in this cluster.";
    for (const edit of Object.values(draft.value.departments)) {
      if (props.canEditDepartment?.(edit.id) === false)
        return "You do not have permission to update one of the edited departments.";
      const original = (props.baselineRows ?? props.rows).find((row) => row.orgUnitId === edit.id)?.resources.gpu;
      const quota = edit.quota ?? original?.deserved;
      const limit = edit.limit === undefined ? original?.limit : edit.limit;
      if (capabilities.quota && capabilities.limits && limit != null && limit !== -1 && quota != null && limit < quota)
        return "Maximum allocation must be at least guaranteed GPUs.";
    }
    return "";
  } catch (error) {
    return extractErrorMessages(error).join("; ");
  }
});
watch(
  draft,
  () => {
    saveError.value = "";
  },
  { deep: true },
);
async function save() {
  if (saving.value || props.loading || !props.canSavePool || !props.canEditDepartments || validationError.value) return;
  saving.value = true;
  emit("saving", true);
  saveError.value = "";
  try {
    if (draft.value.mode === "guaranteed") {
      for (const edit of Object.values(draft.value.departments)) {
        if (edit.quota === undefined) continue;
        if (!props.loadQuotaContext) throw new Error("Project quota validation is unavailable. Try again.");
        const context = await props.loadQuotaContext(edit.id);
        if (context.enforceChildQuota && (context.childQuotaSum == null || edit.quota < context.childQuotaSum))
          throw new Error("Guaranteed GPUs cannot be below the child project quota sum.");
      }
    }
    const pool = await saver.save(props.pool, draft.value);
    saving.value = false;
    emit("saving", false);
    emit("saved", pool);
  } catch (error) {
    if (saver.savedPool) emit("pool-saved", saver.savedPool);
    saveError.value =
      (saver.savedPool ? "Pool settings were saved, but some department changes remain unsaved. " : "") +
      extractErrorMessages(error, "Could not save scheduling. Your changes are preserved.").join("; ");
  } finally {
    saving.value = false;
    emit("saving", false);
  }
}
</script>

<style scoped lang="scss">
.node-pool-scheduling-wizard :deep([aria-current="step"] .wizard-step-title) {
  font-weight: 700;
}

.scheduling-fields {
  border: 0;
  padding: 0;
  margin: 0;
  min-width: 0;
}
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
  container-type: inline-size;
  padding: 18px 19px 40px;
}
.capacity-label,
.preview-notice {
  color: var(--b-text-sub);
}
.capacity {
  display: flex;
  align-items: flex-end;
  gap: 10px;
  margin-top: 2px;
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
.capacity-number {
  font-size: 24px;
  font-weight: 600;
  line-height: 32px;
  letter-spacing: 0.0703px;
  color: #0f172b;
}
.capacity-unit {
  font-size: 14px;
  font-weight: 500;
  line-height: 20px;
  height: 22px;
  letter-spacing: -0.1504px;
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
    height: auto;
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
    white-space: normal;
    font-weight: 500;
    padding-right: 8px;
  }
  :deep(.runai-segment__description) {
    font-weight: 500;
    color: var(--b-text-sub);
    margin-top: 3px;
    line-height: 18px;
  }
  :deep(.runai-segment__text) {
    align-items: flex-start;
  }
  :deep(.runai-segment--selected:not(.runai-segment--disabled) .runai-segment__icon path) {
    fill: #689f38;
    fill-opacity: 1;
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
  top: calc(100% - 4px);
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
@container (max-width: 520px) {
  .mode-cards {
    flex-direction: column;
    :deep(.runai-segment) {
      width: 100%;
    }
  }
}
</style>
