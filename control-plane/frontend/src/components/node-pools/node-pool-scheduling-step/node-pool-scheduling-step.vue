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
    @step-click="onStepClick"
  >
    <runai-form-wrapper
      :form-state="{ pool: previewPool, scheduling: draft, saving }"
      :prevent-leave="saving || creatingDepartment"
      full-width
      @leave-page="emit('leave-approved')"
    >
      <fieldset :disabled="saving" class="scheduling-fields">
        <runai-expansion-wrapper no-first-top-padding>
          <runai-expansion-item
            class="scheduling-mode-section"
            label="Scheduling mode"
            default-opened
            disable-closing
            body-no-padding
          >
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
                <div
                  class="mode-hover-target"
                  aid="scheduling-mode-hover-target"
                  role="presentation"
                  @mouseenter="openModeHelp"
                  @mouseover="hoverModeHelp"
                  @mouseleave="leaveModeHelp"
                  @focusin="focusModeHelp"
                  @focusout="blurModeHelp"
                >
                  <runai-segmented-button
                    :model-value="selectedMode"
                    :options="options"
                    class="mode-cards"
                    :class="{ 'mode-cards--expanded': modeHelpExpanded, 'mode-cards--notice': showGuaranteedNotice }"
                    spread
                    :disable="saving || !canEditDepartments || !canViewDepartments"
                    aria-label="Scheduling mode"
                    aid="scheduling-mode-options"
                    @keydown.esc.stop="modeHelpDismissed = true"
                    @update:model-value="selectMode"
                  >
                    <template #option-end="{ selected, option }">
                      <q-icon class="mode-icon" :name="modeIcons[String(option.value)]" size="20px" aria-hidden="true" />
                      <div class="mode-details">
                        <q-slide-transition :duration="511">
                          <div v-show="modeHelpExpanded" class="mode-explanation-transition">
                            <div class="mode-explanation-content">
                              <span class="mode-explanation">{{ modeExplanations[String(option.value)] }}</span>
                            </div>
                          </div>
                        </q-slide-transition>
                        <q-slide-transition :duration="511">
                          <div
                            v-show="option.value === ESchedulingMode.Guaranteed && showGuaranteedNotice"
                            class="mode-warning-transition"
                          >
                            <div class="mode-warning-container">
                              <notification-alert-bar
                                class="mode-guarantee-warning"
                                :variant="ENotificationBarVariant.Default"
                                icon-override="ph ph-lock"
                                >Mode can't be changed after saving</notification-alert-bar
                              >
                            </div>
                          </div>
                        </q-slide-transition>
                      </div>
                      <q-icon
                        v-if="selected || option.disable || saving || !canEditDepartments || !canViewDepartments"
                        name="ph-bold ph-check"
                        size="16px"
                        class="mode-check"
                        :class="{ 'mode-check--hover-only': !selected }"
                        aria-hidden="true"
                      />
                    </template>
                  </runai-segmented-button>
                </div>
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
                v-if="showNewDepartment && ((canEditDepartments && canViewDepartments) || creatingDepartment)"
                :cluster-id="pool.clusterId"
                :cluster-name="pool.clusterName"
                @close="showNewDepartment = false"
                @created="emit('refresh')"
                @saving="onDepartmentSaving"
              />
            </div>
          </runai-expansion-item>
        </runai-expansion-wrapper>
      </fieldset>
    </runai-form-wrapper>
    <template #content-footer>
      <div class="scheduling-footer">
        <p v-if="saveError || validationError" class="text-negative" role="alert">{{ saveError || validationError }}</p>
        <div class="row justify-end q-gutter-md">
          <runai-btn
            aid="scheduling-save-close"
            label="Save & close"
            flat
            color="primary"
            :loading="saving"
            :disable="saveDisabled"
            @click="save('close')"
          />
          <runai-btn
            aid="scheduling-save-continue"
            label="Save & continue"
            color="primary"
            :loading="saving"
            :disable="saveDisabled"
            @click="save('continue')"
          />
        </div>
      </div>
    </template>
    <template #right-sidebar>
      <node-pool-wizard-summary :pool="previewPool" :resources="resources" />
    </template>
  </runai-wizard-layout>
</template>

<script setup lang="ts">
import { ENodePoolSchedulingMode as ESchedulingMode } from "@/models/node-pool.model";

import { computed, onBeforeUnmount, ref, watch } from "vue";
import { buildSchedulingSave } from "@/utils/node-pool-scheduling.util/node-pool-scheduling-save.util";
import { useNodePoolSchedulingSave } from "@/composables/node-pool/use-node-pool-scheduling-save.composable";
import { extractErrorMessages } from "@/utils/error-alert.util";
import { schedulingDurationToFields } from "@/utils/node-pool-scheduling.util/scheduling-duration.util";
import NotificationAlertBar from "@/components/common/notification-alert-bar/notification-alert-bar.vue";
import { ENotificationBarVariant } from "@/components/common/notification-alert-bar/notification-alert-bar.model";
import { NodePoolDepartmentTable } from "@/components/node-pools/node-pool-department-table";
import { NodePoolSchedulingSettings } from "@/components/node-pools/node-pool-scheduling-settings";
import { NodePoolNewDepartmentDialog } from "@/components/node-pools/node-pool-new-department-dialog";
import { useNodePoolSchedulingDraft } from "@/composables/node-pool/use-node-pool-scheduling-draft.composable";
import {
  toSchedulingCapabilities,
  isDepartmentLimitBelowQuota,
  type IDepartmentSchedulingEdit,
  type ISchedulingTotals,
  type IDepartmentValidationContext,
} from "@/models/node-pool-scheduling.model";
import type { DepartmentNodePoolResources } from "@/swagger-models/org-unit-service-client";
import RunaiWizardLayout from "@/components/common/runai-wizard-layout/runai-wizard-layout.vue";
import RunaiFormWrapper from "@/components/common/runai-form-wrapper/runai-form-wrapper.vue";
import RunaiExpansionWrapper from "@/components/common/runai-expansion-wrapper/runai-expansion-wrapper.vue";
import RunaiExpansionItem from "@/components/common/runai-expansion-item/runai-expansion-item.vue";
import RunaiSectionTitle from "@/components/common/runai-section-title/runai-section-title.vue";
import RunaiBtn from "@/components/common/runai-btn/runai-btn.vue";
import { RunaiSegmentedButton } from "@/components/common/runai-segmented-button";
import NodePoolWizardSummary from "../node-pool-wizard-summary/node-pool-wizard-summary.vue";
import { isSupportedSchedulingMode, getSchedulingMode, type INodePoolSummaryResources } from "@/models/node-pool.model";
import type { Nodepool } from "@/swagger-models/cluster-service-client";

const props = withDefaults(
  defineProps<{
    pool: Nodepool;
    resources?: INodePoolSummaryResources;
    rows?: DepartmentNodePoolResources[];
    baselineRows?: DepartmentNodePoolResources[];
    aggregates?: Record<string, ISchedulingTotals>;
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
    loadQuotaContext?: (id: string) => Promise<IDepartmentValidationContext>;
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
  saved: [pool: Nodepool, destination?: "close" | "continue"];
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
  [ESchedulingMode.Proportional]:
    "GPU resources are distributed proportionally according to each department's and project's weight.",
  [ESchedulingMode.Ranked]:
    "Higher-ranked departments and projects receive GPU resources first. Remaining capacity goes to lower-ranked ones, in rank order.",
  [ESchedulingMode.Guaranteed]:
    "Each department has a guaranteed GPU quota. Any unused quota is redistributed by rank and weight.",
};
const steps = computed(() => [
  {
    name: "identity",
    title: "Cluster & name",
    done: true,
    disabled: true,
  },
  { name: "scheduling", title: "Scheduling mode", done: false },
  { name: "topology", title: "Topology & placement", done: false, disabled: saveDisabled.value },
]);
function onStepClick({ name }: { name: string }) {
  if (name === "topology") void save("continue");
}
const modePermissionTooltip = computed(() =>
  !props.canViewDepartments
    ? "You do not have permission to view departments."
    : !props.canEditDepartments
      ? "You do not have permission to edit departments."
      : undefined,
);
const options = computed(() => [
  {
    value: ESchedulingMode.Proportional,
    disable: props.pool.schedulingCapabilities?.quota === true,
    disabledTooltip:
      modePermissionTooltip.value ??
      (props.pool.schedulingCapabilities?.quota
        ? "An existing Guaranteed quota pool cannot switch to a lower scheduling mode."
        : undefined),
    label: "Proportional",
    description: "Weight-based distribution",
    aid: "mode-proportional",
  },
  {
    value: ESchedulingMode.Ranked,
    disable: props.pool.schedulingCapabilities?.quota === true,
    disabledTooltip:
      modePermissionTooltip.value ??
      (props.pool.schedulingCapabilities?.quota
        ? "An existing Guaranteed quota pool cannot switch to a lower scheduling mode."
        : undefined),
    label: "Ranked",
    description: "Rank-based distribution",
    aid: "mode-ranked",
  },
  {
    value: ESchedulingMode.Guaranteed,
    label: "Guaranteed quota",
    description: "Guaranteed quota, redistribute unused",
    disabledTooltip: modePermissionTooltip.value,
    aid: "mode-guaranteed",
  },
]);
const modeIcons: Record<string, string> = {
  [ESchedulingMode.Proportional]: "ph ph-chart-pie-slice",
  [ESchedulingMode.Ranked]: "ph ph-ranking",
  [ESchedulingMode.Guaranteed]: "ph ph-shield-check",
};
const savedMode = getSchedulingMode(props.pool.schedulingCapabilities);
const mode = isSupportedSchedulingMode(savedMode) ? savedMode : undefined;
const { draft, requestMode } = useNodePoolSchedulingDraft({
  mode,
  automaticWeights:
    mode === ESchedulingMode.Guaranteed ? props.pool.schedulingCapabilities?.adjustableWeights === false : true,
  cpuQuota: props.pool.schedulingCapabilities?.cpuQuota === true,
  maxAllocation: props.pool.schedulingCapabilities?.limits === true,
  timeFairshare: props.pool.schedulingConfiguration?.timeBasedFairShare?.enabled === true,
  historicalWeight: props.pool.schedulingConfiguration?.timeBasedFairShare?.historicalUsageWeight ?? 1,
  historicalWindow: schedulingDurationToFields(
    props.pool.schedulingConfiguration?.timeBasedFairShare?.windowDuration ?? "7d",
    false,
  ),
  minimumRuntime: schedulingDurationToFields(props.pool.schedulingConfiguration?.minGuaranteedRuntime),
  capabilities: props.pool.schedulingCapabilities,
  departments: {},
});
watch(
  () => draft.value.mode,
  (mode, previous) => {
    if (mode === ESchedulingMode.Proportional && previous && previous !== ESchedulingMode.Proportional)
      emit("filter-rank", undefined);
  },
);
const selectedMode = computed(() => draft.value.mode ?? null);
const modeCardsHovered = ref(false);
const modeCardsFocused = ref(false);
const hoveredMode = ref<string>();
const modeHelpDismissed = ref(false);
const unsavedGuaranteedMode = computed(
  () => draft.value.mode === ESchedulingMode.Guaranteed && !props.pool.schedulingCapabilities?.quota,
);
const modeHelpExpanded = computed(
  () => !modeHelpDismissed.value && (modeCardsHovered.value || modeCardsFocused.value || unsavedGuaranteedMode.value),
);
const showGuaranteedNotice = computed(
  () =>
    modeHelpExpanded.value &&
    !props.pool.schedulingCapabilities?.quota &&
    (hoveredMode.value === ESchedulingMode.Guaranteed || draft.value.mode === ESchedulingMode.Guaranteed),
);
let modeHelpCloseTimer: ReturnType<typeof setTimeout> | undefined;
onBeforeUnmount(() => clearTimeout(modeHelpCloseTimer));
function openModeHelp() {
  clearTimeout(modeHelpCloseTimer);
  modeCardsHovered.value = true;
  modeHelpDismissed.value = false;
}
function hoverModeHelp(event: MouseEvent) {
  clearTimeout(modeHelpCloseTimer);
  modeCardsHovered.value = true;
  const aid = (event.target as Element).closest("[role=radio]")?.getAttribute("aid");
  hoveredMode.value = options.value.find((option) => option.aid === aid)?.value;
}
function leaveModeHelp() {
  clearTimeout(modeHelpCloseTimer);
  modeHelpCloseTimer = setTimeout(() => {
    modeCardsHovered.value = false;
    hoveredMode.value = undefined;
  }, 150);
}
function focusModeHelp(event: FocusEvent) {
  modeCardsFocused.value = (event.target as Element).matches(":focus-visible");
  modeHelpDismissed.value = false;
}
function blurModeHelp(event: FocusEvent) {
  if (!(event.currentTarget as Element).contains(event.relatedTarget as Node | null)) modeCardsFocused.value = false;
}
const previewPool = computed(() => ({
  ...props.pool,
  schedulingCapabilities: draft.value.mode ? toSchedulingCapabilities(draft.value) : draft.value.capabilities,
}));
function selectMode(value: string | number | boolean | null) {
  if (saving.value || (props.pool.schedulingCapabilities?.quota && value !== ESchedulingMode.Guaranteed)) return;
  if (value === ESchedulingMode.Proportional || value === ESchedulingMode.Ranked || value === ESchedulingMode.Guaranteed)
    requestMode(value);
}
function applyEdit(edit: IDepartmentSchedulingEdit) {
  if (saving.value || !props.canEditDepartments) return;
  draft.value.departments[edit.id] = { ...draft.value.departments[edit.id], ...edit };
}

const saving = ref(false);
const creatingDepartment = ref(false);
function onDepartmentSaving(value: boolean) {
  creatingDepartment.value = value;
  emit("saving", saving.value || value);
}
const saveError = ref("");
const saver = useNodePoolSchedulingSave();
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
      if (capabilities.quota && capabilities.limits && isDepartmentLimitBelowQuota(limit, quota))
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
const saveDisabled = computed(
  () =>
    saving.value ||
    creatingDepartment.value ||
    props.loading ||
    !props.canSavePool ||
    !props.canEditDepartments ||
    !!validationError.value,
);
async function save(destination: "close" | "continue" = "close") {
  if (saveDisabled.value) return;
  saving.value = true;
  emit("saving", true);
  saveError.value = "";
  try {
    if (draft.value.mode === ESchedulingMode.Guaranteed) {
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
    emit("saved", pool, destination);
  } catch (error) {
    if (saver.savedPool) emit("pool-saved", saver.savedPool);
    saveError.value =
      (saver.poolSettingsSaved
        ? "Pool settings were saved, but some department changes remain unsaved. "
        : saver.savedPool
          ? "Current pool settings were not saved. Some changes were saved in an earlier attempt. "
          : "") + extractErrorMessages(error, "Could not save scheduling. Your changes are preserved.").join("; ");
  } finally {
    saving.value = false;
    emit("saving", false);
  }
}
</script>

<style scoped lang="scss">
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
  font-weight: 500;
  line-height: 20px !important;
}
.node-pool-scheduling-wizard :deep(.wizard-step--disabled .wizard-step-title) {
  color: var(--text-disabled) !important;
}
.node-pool-scheduling-wizard :deep([aria-current="step"] .wizard-step-title) {
  color: var(--b-text-main) !important;
}
.node-pool-scheduling-wizard :deep(.wizard-step--completed .wizard-step-indicator),
.node-pool-scheduling-wizard :deep(.wizard-step--selected .wizard-step-indicator) {
  background-color: var(--feedback-success) !important;
}
.node-pool-scheduling-wizard :deep(.wizard-step--default.wizard-step--disabled .wizard-step-indicator) {
  background-color: var(--text-disabled) !important;
}
.node-pool-scheduling-wizard :deep(.wizard-step-icon) {
  color: var(--text-inverse) !important;
}
.node-pool-scheduling-wizard :deep(.wizard-step-number) {
  font-weight: 800 !important;
  line-height: 16.8px;
  left: 50%;
}
.node-pool-scheduling-wizard :deep(.wizard-step-divider) {
  margin-left: 27.5px;
}
.node-pool-scheduling-wizard :deep(.wizard-step--disabled .wizard-step-item) {
  opacity: 1 !important;
}
.scheduling-overview {
  container-type: inline-size;
  padding: 18px 20px 40px;
}
.capacity-label {
  color: var(--b-text-sub);
}
.capacity {
  display: flex;
  align-items: flex-end;
  gap: 10px;
  margin-top: 2px;
}
.mode-selector {
  margin-top: 40px;
  margin-bottom: 0;
}
.scheduling-mode-section :deep(> .q-expansion-item__container) {
  border: 0;
}
.scheduling-mode-section :deep(.runai-expansion-header) {
  height: 48px;
  min-height: 48px;
  background: var(--bg-primary-default);
  box-shadow: inset 0 -1px var(--border-subtle);
}
.scheduling-mode-section :deep(.runai-expansion-divider) {
  height: 0;
}
.mode-selector :deep(.runai-section-title) {
  margin-right: -3px;
  margin-bottom: 10px;
  margin-left: -3px;
}
.capacity-number {
  font-size: 24px;
  font-weight: 600;
  line-height: 32px;
  letter-spacing: 0.0703px;
  color: var(--b-text-main);
}
.capacity-unit {
  font-size: 14px;
  font-weight: 500;
  line-height: 20px;
  height: 22px;
  letter-spacing: -0.1504px;
}
.mode-hover-target {
  margin-inline: -20px;
  padding-inline: 20px;
}
.mode-cards {
  align-items: flex-start;
  border: 0;
  padding: 0;
  gap: 7px;

  :deep(.runai-segment) {
    display: grid;
    grid-template-columns: 20px minmax(0, 1fr);
    align-content: start;
    position: relative;
    min-width: 0;
    min-height: 61px;
    height: auto;
    overflow: visible;
    border: 1px solid var(--border-subtle);
    border-radius: 5px;
    padding: 9px 15px;
    column-gap: 10px;
    row-gap: 0;
    text-align: left;
    background: var(--bg-primary-default);
    transition: none;
  }
  :deep(.runai-segment:hover:not(.runai-segment--selected):not(.runai-segment--disabled)) {
    background: var(--bg-primary-hover);
  }
  :deep(.runai-segment--selected:hover:not(.runai-segment--disabled)) {
    background: var(--bg-primary-pressed);
  }
  :deep(.mode-icon) {
    grid-column: 1;
    color: var(--b-text-sub);
    grid-row: 1;
    align-self: center;
  }
  :deep(.runai-segment--selected) {
    border-color: var(--border-pressed);
    box-shadow: inset 0 0 0 2px var(--border-pressed);
  }
  :deep(.runai-segment__label) {
    white-space: normal;
    font-weight: 500;
    line-height: 20px;
    padding-right: 8px;
  }
  :deep(.runai-segment__description) {
    font-weight: 500;
    color: var(--b-text-sub);
    margin-top: 3px;
    line-height: 18px;
  }
  :deep(.runai-segment__text) {
    grid-column: 2;
    grid-row: 1;
    min-height: 41px;
    align-items: flex-start;
  }
  :deep(.runai-segment--selected:not(.runai-segment--disabled) .mode-icon) {
    color: var(--border-pressed);
  }
}
.mode-check {
  position: absolute;
  top: 1px;
  // Source MAX inset is 2.33334px; the card border contributes 1px.
  right: 1.33334px;
  width: 20px;
  height: 20px;
  border-radius: 3px;
  background: var(--border-pressed);
  color: var(--text-inverse);
}
.scheduling-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 40px 0 8px;
}
.scheduling-footer > .row {
  margin-left: auto;
}
.scheduling-content {
  display: block;
}
.scheduling-content > .department-section {
  flex: 1;
  min-width: 0;
}
.mode-explanation {
  grid-column: 1 / -1;
  white-space: normal;
  color: var(--b-text-sub);
  font-size: 12px;
  font-weight: 400;
  line-height: 16.8px;
}
.mode-details {
  grid-column: 1 / -1;
}
.mode-explanation-content,
.mode-warning-container {
  padding-top: 10px;
}
.mode-cards :deep(.runai-segment:not([aid="mode-guaranteed"])) .mode-explanation-content {
  min-height: 61px;
}
.mode-explanation {
  display: block;
}
.mode-cards {
  align-items: stretch;
}
@media (prefers-reduced-motion: reduce) {
  .mode-explanation-transition,
  .mode-warning-transition {
    transition-duration: 0.001ms !important;
  }
}
.mode-guarantee-warning {
  min-height: 40px;
  padding: 10px 16px;
  border-radius: 0;
  text-align: left;
  animation: none;
  background-color: var(--b-bg-secondary-default);
}
.view-only {
  margin-left: 10px;
  border: 1px solid var(--border-default);
  color: var(--b-text-sub);
  font-size: 12px;
  padding: 2px 8px;
  border-radius: 2px;
}
.mode-cards :deep(.runai-segment--disabled) {
  background: var(--bg-primary-hover);
  opacity: 1 !important;
}
.mode-cards :deep(.runai-segment--disabled:not(.runai-segment--selected) .runai-segment__label) {
  color: var(--text-disabled);
}
.mode-cards :deep(.runai-segment--disabled .runai-segment__description),
.mode-cards :deep(.runai-segment--disabled .mode-explanation) {
  color: var(--text-tertiary);
}
.mode-cards :deep(.runai-segment--disabled .mode-icon) {
  color: var(--text-tertiary);
}
.mode-check--hover-only {
  visibility: hidden;
}
.mode-cards :deep(.runai-segment--disabled:hover) {
  border-color: var(--border-pressed);
  box-shadow: inset 0 0 0 2px var(--border-pressed);
}
.mode-cards :deep(.runai-segment--disabled:hover .runai-segment__label) {
  color: var(--interactive-secondary-hover-text);
}
.mode-cards :deep(.runai-segment--disabled:hover .mode-icon) {
  color: var(--interactive-secondary-hover-text);
}
.mode-cards :deep(.runai-segment--disabled:hover .mode-check--hover-only) {
  visibility: visible;
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
