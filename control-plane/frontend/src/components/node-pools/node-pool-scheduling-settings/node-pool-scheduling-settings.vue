<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <aside
    class="scheduling-settings"
    :class="{ 'scheduling-settings--compact': draft.mode === ESchedulingMode.Proportional }"
    aria-label="Advanced settings"
  >
    <header>
      <span>Advanced settings</span
      ><q-btn
        ref="closeButton"
        flat
        round
        dense
        icon="ph-bold ph-x"
        aria-label="Hide Advanced settings"
        @click="emit('close')"
      >
        <q-tooltip class="node-pool-settings-hide-help">Hide Advanced settings</q-tooltip>
      </q-btn>
    </header>
    <div class="settings-body">
      <section v-if="draft.mode === ESchedulingMode.Guaranteed">
        <scheduling-section-title
          title="Weight within rank"
          interactive-tooltip
          tooltip-icon="ph-bold ph-info"
          tooltip-text="For more information, see The NVIDIA Run:ai Scheduler: Concepts and Principles guide"
        >
          <template #tooltip-content>
            <span class="node-pool-settings-help node-pool-settings-help--standard">
              For more information, see
              <a :href="DOCS.PLATFORM_MANAGEMENT.SCHEDULING_CONCEPTS" target="_blank" rel="noopener noreferrer"
                >The NVIDIA Run:ai Scheduler: Concepts and Principles guide</a
              >
            </span>
          </template>
        </scheduling-section-title>
        <q-option-group
          class="weight-options runai-option-group"
          :model-value="draft.automaticWeights ? 'automatic' : 'custom'"
          :options="weightOptions"
          type="radio"
          aria-label="Weight within rank"
          :disable="readonly"
          @update:model-value="patch({ automaticWeights: $event === 'automatic' })"
        >
          <template #label="option">
            <div class="weight-option-label" :aid="`radio-option-${option.label}`">
              <span class="text-b-subtitle2">{{ option.label }}</span>
              <scheduling-help-tooltip
                :tooltip-text="option.toolTip"
                width="374px"
                popup-class="node-pool-settings-option-help"
                @click.stop
              />
            </div>
          </template>
        </q-option-group>
      </section>
      <section v-if="isSupportedSchedulingMode(draft.mode)" class="toggle-section">
        <scheduling-section-title title="Max GPU allocation" tooltip-icon="ph-bold ph-info" :tooltip-text="maxHelp">
          <template #tooltip-content>
            <span class="node-pool-settings-help node-pool-settings-help--standard">{{ maxHelp }}</span>
          </template>
        </scheduling-section-title>
        <runai-toggle
          :model-value="!!draft.maxAllocation"
          :disable="readonly || checkingLimits || (!canResetResources && (savedMaxAllocation ?? draft.maxAllocation))"
          aid="max-allocation-toggle"
          @update:model-value="setMax"
        />
      </section>
      <p v-if="maxAllocationError" role="alert" class="text-negative q-mt-sm">{{ maxAllocationError }}</p>
      <section>
        <div class="toggle-section">
          <scheduling-section-title
            title="Time-based fairshare"
            interactive-tooltip
            tooltip-icon="ph-bold ph-info"
            width="427px"
            tooltip-text="When enabled, the Scheduler adjusts fairshare scores based on each project's weight and historical usage records. To use this feature, persistent data must be enabled. When disabled, the Scheduler uses the default fairshare, with scheduling fairness based on the current project's weight and resource usage. For more information, see the Time-based fairshare guide."
          >
            <template #tooltip-content>
              <span class="node-pool-settings-help node-pool-settings-help--fairshare">
                When enabled, the Scheduler adjusts fairshare scores based on each project's weight and historical usage
                records.<br />
                To use this feature,
                <a :href="DOCS.NODE_POOLS.PERSISTENT_DATA" target="_blank" rel="noopener noreferrer">persistent data</a>
                must be enabled.<br />
                When disabled, the Scheduler uses the default fairshare, with scheduling fairness based on the current
                project's weight and resource usage. For more information, see the
                <a :href="DOCS.NODE_POOLS.TIME_BASED_FAIR_SHARE" target="_blank" rel="noopener noreferrer"
                  >Time-based fairshare guide</a
                >.
              </span>
            </template>
          </scheduling-section-title>
          <runai-toggle
            :model-value="!!draft.timeFairshare"
            :disable="readonly"
            aid="time-fairshare-toggle"
            @update:model-value="patch({ timeFairshare: $event })"
          />
        </div>
        <div v-if="draft.timeFairshare" class="historical-fields">
          <div class="historical-weight-group">
            <scheduling-section-title
              flat
              class="historical-weight-title"
              title="Set the weight value that determines how a project's historical resource usage affects its fairshare"
              width="475px"
              tooltip-text="This determines how strongly historical resource usage influences fairshare scores: 0 — History is ignored; fairshare is based on current weight only. 0–1 — History has partial influence; lower values reduce its effect. 1 — History and current weight are balanced equally. >1 — History dominates; higher values amplify its effect."
              ><template #tooltip-content>
                <div class="node-pool-settings-help node-pool-settings-help--historical">
                  This determines how strongly historical resource usage influences fairshare scores:
                  <ul class="historical-weight-help">
                    <li>0 — History is ignored; fairshare is based on current weight only.</li>
                    <li>0–1 — History has partial influence; lower values reduce its effect.</li>
                    <li>1 — History and current weight are balanced equally.</li>
                    <li>&gt;1 — History dominates; higher values amplify its effect.</li>
                  </ul>
                </div>
              </template></scheduling-section-title
            >
            <div class="historical-weight-field">
              <q-input
                stack-label
                class="historical-weight-input"
                @input.capture="trackNumericValidity($event, 'historicalWeight')"
                :model-value="draft.historicalWeight"
                @update:model-value="patch({ historicalWeight: number($event, 'historicalWeight') })"
                type="number"
                inputmode="decimal"
                step="any"
                @keydown="numericKeydown($event)"
                @paste="numericPaste($event)"
                @beforeinput="numericBeforeInput($event)"
                min="0"
                label="Historical usage weight"
                :disable="readonly"
                :error="invalid(draft.historicalWeight)"
                error-message="Enter a value equal to or above 0.0"
                no-error-icon
              />
            </div>
          </div>
          <div class="historical-window-group">
            <scheduling-section-title
              flat
              title="Set the duration of the historical usage window"
              width="374px"
              :tooltip-text="windowHelp"
            >
              <template #tooltip-content>
                <span class="node-pool-settings-help node-pool-settings-help--standard">{{ windowHelp }}</span>
              </template>
            </scheduling-section-title>
            <div class="duration-grid historical-window">
              <q-input
                v-for="unit in windowUnits"
                :key="unit"
                @input.capture="trackNumericValidity($event, `historicalWindow.${unit}`)"
                :model-value="draft.historicalWindow?.[unit] ?? 0"
                @update:model-value="setDuration('historicalWindow', unit, $event)"
                stack-label
                type="number"
                inputmode="decimal"
                step="any"
                @keydown="numericKeydown($event)"
                @paste="numericPaste($event)"
                @beforeinput="numericBeforeInput($event)"
                min="0"
                :label="unit"
                :disable="readonly"
                :error="invalid(draft.historicalWindow?.[unit])"
                :error-message="`Enter a timeframe of at least 0 ${unit.toLowerCase()}`"
                no-error-icon
              />
            </div>
          </div>
        </div>
      </section>
      <section class="runtime-section">
        <scheduling-section-title
          title="Set the minimum guaranteed runtime"
          tooltip-icon="ph-bold ph-info"
          :tooltip-text="runtimeHelp"
        >
          <template #tooltip-content>
            <span class="node-pool-settings-help node-pool-settings-help--standard">{{ runtimeHelp }}</span>
          </template>
        </scheduling-section-title>
        <div class="duration-grid">
          <q-input
            v-for="unit in runtimeUnits"
            :key="unit"
            @input.capture="trackNumericValidity($event, `minimumRuntime.${unit}`)"
            :model-value="draft.minimumRuntime?.[unit] ?? 0"
            @update:model-value="setDuration('minimumRuntime', unit, $event)"
            stack-label
            type="number"
            inputmode="decimal"
            step="any"
            @keydown="numericKeydown($event)"
            @paste="numericPaste($event)"
            @beforeinput="numericBeforeInput($event)"
            min="0"
            :label="unit"
            :disable="readonly"
            :error="invalid(draft.minimumRuntime?.[unit])"
            error-message="Enter a timeframe of at least 0 seconds"
            no-error-icon
          />
        </div>
      </section>
      <template v-if="draft.mode === ESchedulingMode.Guaranteed">
        <section class="toggle-section">
          <scheduling-section-title
            title="CPU quota"
            interactive-tooltip
            tooltip-icon="ph-bold ph-info"
            tooltip-text="When enabled, CPU resources can be set per department or project. For more information, see the Departments guide and Projects guide."
          >
            <template #tooltip-content>
              <span class="node-pool-settings-help node-pool-settings-help--cpu">
                When enabled, CPU resources can be set per department or project. For more information, see the
                <a :href="DOCS.PLATFORM_MANAGEMENT.DEPARTMENTS" target="_blank" rel="noopener noreferrer"
                  >Departments guide</a
                >
                and
                <a :href="DOCS.PLATFORM_MANAGEMENT.PROJECTS" target="_blank" rel="noopener noreferrer">Projects guide</a
                >.
              </span>
            </template>
          </scheduling-section-title>
          <runai-toggle
            :model-value="draft.cpuQuota"
            :disable="readonly || (!canResetResources && (savedCpuQuota ?? draft.cpuQuota))"
            aid="cpu-quota-toggle"
            @update:model-value="setCpu"
          />
        </section>
      </template>
    </div>
    <runai-base-dialog
      v-if="confirmation"
      :model-value="true"
      custom-class="scheduling-confirm"
      size="width-sm"
      @close="confirmation = undefined"
    >
      <template #header>{{
        confirmation === "cpuQuota" ? "Disable CPU Quota?" : "Disable Max GPU Allocation?"
      }}</template>
      <template #body>{{
        confirmation === "maxAllocation"
          ? 'Disabling this will reset the max GPU allocation of all departments in this node pool to "Unlimited"'
          : confirmation === "unsavedLimits"
            ? "You have unsaved changes. If you disable max GPU allocation, they will be lost."
            : "The CPU compute and CPU memory quotas of departments and projects in this node pool will be removed. Their limit will be set to Auto, and workloads can use up to the node pool’s maximum available CPU resources. Running workloads are not affected."
      }}</template>
      <template #footer
        ><runai-btn
          aid="keep-setting-enabled"
          flat
          color="primary"
          label="Keep enabled"
          @click="confirmation = undefined" /><runai-btn
          color="primary"
          :label="confirmation === 'maxAllocation' ? 'Reset & disable' : 'Disable'"
          aid="confirm-disable-setting"
          @click="confirmDisable"
      /></template>
    </runai-base-dialog>
  </aside>
</template>
<script setup lang="ts">
import { ENodePoolSchedulingMode as ESchedulingMode, isSupportedSchedulingMode } from "@/models/node-pool.model";

import { numericKeydown, numericPaste, numericBeforeInput } from "@/utils/numeric-input.util/numeric-input.util";
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { QBtn } from "quasar";
import { extractErrorMessages } from "@/utils/error-alert.util";
import { DOCS } from "@/constants/docs/index";
import { SchedulingSectionTitle } from "@/components/node-pools/scheduling-section-title";
import { SchedulingHelpTooltip } from "@/components/node-pools/scheduling-help-tooltip";
import { RunaiToggle } from "@/components/common/runai-toggle/index";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog/index";
import RunaiBtn from "@/components/common/runai-btn/runai-btn.vue";
import type { SchedulingDraft } from "@/models/node-pool-scheduling.model";
const props = defineProps<{
  draft: SchedulingDraft;
  readonly?: boolean;
  canResetResources?: boolean;
  loadHasGpuLimits?: () => Promise<boolean>;
  savedMaxAllocation?: boolean;
  savedCpuQuota?: boolean;
}>();
const emit = defineEmits<{ close: []; patch: [Partial<SchedulingDraft>] }>();
const closeButton = ref<QBtn>();
onMounted(() => closeButton.value?.$el.focus({ preventScroll: true }));
const weightOptions = [
  {
    label: "Automatic",
    value: "automatic",
    toolTip:
      "A project or department gets a proportion of unused resources relative to its quota, and it can't be modified.",
  },
  {
    label: "Custom",
    value: "custom",
    toolTip: "Set the proportion of unused resources a project or department can get on top of its quota.",
  },
];
const maxHelp = computed(() =>
  props.draft.mode === ESchedulingMode.Guaranteed
    ? 'When enabled, the maximum GPU device allocation each department and project can get from this node pool, representing the maximum sum of assigned quota and over-quota GPUs, can be set. When disabled, the maximum is "unlimited."'
    : "Set the maximum number of GPUs each department and project can use from this node pool. When disabled, the maximum is unlimited.",
);
const runtimeHelp =
  "This defines the minimum time a preemptible workload is guaranteed to run once it is scheduled in this node pool, before it can be preempted and its resources reallocated to other workloads.";
const windowHelp =
  "This defines the duration of the sliding window used to calculate historical resource usage. At any given moment, the system looks only at usage within this duration from the current time. As time passes, the window slides forward continuously, including new usage and dropping old usage that falls outside the window.";
const runtimeUnits = ["Days", "Hours", "Minutes", "Seconds"];
const windowUnits = ["Days", "Hours", "Minutes"];
const confirmation = ref<"maxAllocation" | "unsavedLimits" | "cpuQuota">();
function patch(value: Partial<SchedulingDraft>) {
  if (!props.readonly) emit("patch", value);
}
function invalid(value: number | null | undefined) {
  return value != null && (!Number.isFinite(value) || value < 0);
}
const invalidNumericInputs = new Map<string, boolean>();
function trackNumericValidity(event: Event, key: string) {
  if (event.target instanceof HTMLInputElement) invalidNumericInputs.set(key, event.target.validity.badInput);
}
function number(value: string | number | null, key: string) {
  if (invalidNumericInputs.get(key)) return NaN;
  return value === "" || value === null ? null : Number(value);
}
function setDuration(field: "historicalWindow" | "minimumRuntime", unit: string, value: string | number | null) {
  patch({ [field]: { ...props.draft[field], [unit]: number(value, `${field}.${unit}`) } });
}
const checkingLimits = ref(false);
const maxAllocationError = ref("");
let limitCheckVersion = 0;
function invalidateLimitCheck() {
  limitCheckVersion++;
  checkingLimits.value = false;
  maxAllocationError.value = "";
}
watch(
  [() => props.draft.mode, () => props.draft.maxAllocation, () => props.readonly, () => props.canResetResources],
  () => {
    invalidateLimitCheck();
    if (confirmation.value !== "cpuQuota") confirmation.value = undefined;
  },
);
onBeforeUnmount(invalidateLimitCheck);
function disableWithoutSavedLimits() {
  if (Object.values(props.draft.departments).some((edit) => edit.limit !== undefined)) {
    confirmation.value = "unsavedLimits";
  } else patch({ maxAllocation: false });
}
async function setMax(value: boolean) {
  if (
    !isSupportedSchedulingMode(props.draft.mode) ||
    props.readonly ||
    checkingLimits.value ||
    (!value && (props.savedMaxAllocation ?? props.draft.maxAllocation) && !props.canResetResources)
  )
    return;
  maxAllocationError.value = "";
  if (value) {
    patch({ maxAllocation: true });
    return;
  }
  if (!(props.savedMaxAllocation ?? props.draft.maxAllocation)) {
    disableWithoutSavedLimits();
    return;
  }
  const version = ++limitCheckVersion;
  checkingLimits.value = true;
  try {
    if (!props.loadHasGpuLimits) throw new Error("Failed to check GPU allocation limits. Try again.");
    const hasLimits = await props.loadHasGpuLimits();
    if (version !== limitCheckVersion) return;
    if (hasLimits) confirmation.value = "maxAllocation";
    else disableWithoutSavedLimits();
  } catch (error) {
    if (version === limitCheckVersion) {
      maxAllocationError.value = extractErrorMessages(error, "Failed to check GPU allocation limits. Try again.").join(
        " ",
      );
    }
  } finally {
    if (version === limitCheckVersion) checkingLimits.value = false;
  }
}
function setCpu(value: boolean) {
  if (props.readonly || (!value && (props.savedCpuQuota ?? props.draft.cpuQuota) && !props.canResetResources)) return;
  if (!value && (props.savedCpuQuota ?? props.draft.cpuQuota)) confirmation.value = "cpuQuota";
  else patch({ cpuQuota: value });
}
function confirmDisable() {
  if (confirmation.value === "cpuQuota") patch({ cpuQuota: false });
  else {
    const departments = { ...props.draft.departments };
    for (const [id, edit] of Object.entries(departments)) {
      if (edit.limit === undefined) continue;
      const remaining = { ...edit };
      delete remaining.limit;
      if (Object.keys(remaining).length === 1) delete departments[id];
      else departments[id] = remaining;
    }
    patch({ maxAllocation: false, departments });
  }
  confirmation.value = undefined;
}
</script>
<style scoped lang="scss">
.scheduling-settings {
  width: 100%;
  min-width: 0;
  flex: 0 1 auto;
  border-left: 1px solid var(--border-subtle);
}
.scheduling-settings--compact {
  align-self: flex-start;
  box-shadow: inset 0 -1px var(--border-subtle);
}
header {
  height: 55px;
  padding: 0 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  box-shadow: inset 0 -1px var(--border-subtle);
  font-weight: 500;
  line-height: 20px;
}
header .q-btn {
  color: var(--icon-default);
  width: 32px;
  min-width: 32px;
  height: 32px;
  min-height: 32px;
}
header :deep(.q-icon) {
  font-size: 20px;
}
.settings-body {
  padding: 20px;
}
.settings-body > section {
  padding: 26px 0;
  border-bottom: 1px solid var(--border-subtle);
}
.settings-body > section:first-child {
  padding-top: 0;
}
.settings-body > .toggle-section:first-child {
  padding-top: 26px;
}
.settings-body > section:last-child {
  border-bottom: 0;
  padding-bottom: 0;
}
.settings-body > .toggle-section:last-child {
  padding-bottom: 26px;
  border-bottom: 1px solid var(--border-subtle);
}
.toggle-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  min-height: 31px;
}
.toggle-section :deep(.runai-section-title) {
  flex: 1;
  min-width: 0;
  margin-bottom: 0;
}
.toggle-section :deep(.q-toggle) {
  max-height: 31px;
}
.toggle-section :deep(.runai-section-title > .row) {
  display: inline;
}
.toggle-section :deep(.runai-section-title > .row > span) {
  display: inline;
}
.toggle-section :deep(.runai-section-title > .row > .q-btn) {
  vertical-align: middle;
  height: 20px;
  width: 20px;
  min-height: 20px;
  padding: 0;
}
.settings-body > .runtime-section:last-child {
  padding-bottom: 0 !important;
}
.weight-options {
  margin-top: 7px;
}
.weight-option-label {
  display: flex;
  align-items: center;
  gap: 0;
  height: 40px;
  margin-right: 16px;
}
.weight-options :deep(.q-radio__label) {
  padding-left: var(--space-xs);
}
.weight-options :deep(.q-radio.disabled) {
  opacity: 1;
}
.weight-options :deep(.q-radio.disabled .q-radio__inner) {
  color: var(--icon-disabled);
}
.settings-body :deep(.runai-section-title) {
  font-size: 14px;
  font-weight: 400;
  line-height: 20px;
}
.duration-grid :deep(.q-field) {
  min-width: 0;
}
.duration-grid :deep(.q-field--with-bottom) {
  padding: 5px 0 29px;
}
.duration-grid :deep(input[type="number"]::-webkit-inner-spin-button) {
  opacity: 1;
}
.duration-grid :deep(.q-field__bottom) {
  min-height: 16px;
  padding-top: 0;
}
.duration-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0 20px;
  align-items: center;
  margin-top: 10px;
}
.historical-window {
  align-items: start;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 20px;
  margin-top: 0;
}
.runtime-section {
  padding-bottom: 26px !important;
}
.historical-fields {
  margin-top: 26px;
  border-top: 1px solid var(--border-subtle);
  padding-top: 26px;
  display: flex;
  flex-direction: column;
  gap: 26px;
}
.historical-weight-group,
.historical-window-group {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.historical-weight-field {
  width: 192px;
  max-width: 100%;
  align-self: flex-start;
  padding: 5px 0;
}
.historical-weight-input {
  width: 100%;
  padding-bottom: 24px;
}
.historical-weight-input :deep(.q-field__control) {
  height: 53px;
  min-height: 53px;
}
.historical-weight-input :deep(.q-field__label) {
  top: 0;
  transform: none;
  font-size: 12px;
  line-height: 16.8px;
}
.historical-weight-input :deep(.q-field__native) {
  min-height: 0;
  padding: 22px 0 10px;
  font-size: 14px;
  line-height: 19.5px;
}
.historical-weight-input :deep(input[type="number"]::-webkit-inner-spin-button) {
  opacity: 1;
}
.historical-fields :deep(.runai-section-title) {
  line-height: 19.5px;
}
.historical-fields :deep(.runai-section-title > .row) {
  flex-wrap: nowrap;
  width: 100%;
  justify-content: space-between;
  gap: 7px;
}
.historical-fields :deep(.runai-section-title > .row > span) {
  flex: 1;
  min-width: 0;
}
.historical-weight-title :deep(> .row > span) {
  max-width: 280px;
}
.historical-fields :deep(.runai-section-title .q-btn) {
  flex-shrink: 0;
  height: 32px;
  width: 32px;
  min-height: 32px;
  padding: 0;
}
.historical-fields :deep(.runai-section-title .q-icon) {
  font-size: 20px;
}
.settings-body :deep(.q-field--error) {
  padding-bottom: 0;
}
.settings-body :deep(.q-field--error .q-field__bottom) {
  position: static;
  transform: none;
  bottom: auto;
  min-height: 0;
  padding: 7px 0 0;
  font-size: 12px;
  line-height: 16.8px;
}
.settings-body :deep(.q-field--error .q-field__messages),
.settings-body :deep(.q-field--error .q-field__messages > div) {
  font-size: 12px;
  line-height: 16.8px;
}
.duration-grid :deep(.q-field__control) {
  height: 53px;
  min-height: 53px;
}
.duration-grid :deep(.q-field__label) {
  top: 0;
  transform: none;
  font-size: 12px;
  line-height: 16.8px;
}
.duration-grid :deep(.q-field__native) {
  min-height: 0;
  padding: 22px 0 10px;
  font-size: 14px;
  line-height: 19.5px;
}
.duration-grid :deep(.q-field--error) {
  padding: 5px 0;
}
.runtime-section > :deep(.runai-section-title) {
  margin-bottom: 0;
  min-height: 32px;
}
</style>
<style lang="scss">
.historical-weight-help {
  margin: 0;
  padding-inline-start: 20px;

  > li {
    list-style-type: disc;
  }
}
.scheduling-confirm {
  width: 500px !important;
}
.scheduling-confirm .runai-base-dialog-header {
  height: 64px;
  padding: 16px;
}
.scheduling-confirm .runai-base-dialog-body {
  padding: 24px 16px;
  min-height: 72px;
}
.scheduling-confirm .runai-base-dialog-footer {
  min-height: 70px;
  padding: 16px;
}
</style>

<style lang="scss">
.scheduling-help-popup:has(> .node-pool-settings-help),
.scheduling-help-popup.node-pool-settings-option-help,
.q-tooltip.node-pool-settings-hide-help {
  padding: 10px;
  border-radius: 3px;
  background: var(--bg-primary-default);
  color: var(--b-text-main);
  box-shadow: var(--shadow-s);
  font-family: Roboto, sans-serif;
  font-size: 12px;
  font-weight: 400;
  line-height: 16.8px;
}
.scheduling-help-popup:has(> .node-pool-settings-help--standard),
.scheduling-help-popup.node-pool-settings-option-help {
  width: 374px;
  max-width: min(374px, calc(100vw - 20px)) !important;
}
.scheduling-help-popup:has(> .node-pool-settings-help--fairshare) {
  width: 427px;
  max-width: min(427px, calc(100vw - 20px)) !important;
}
.scheduling-help-popup:has(> .node-pool-settings-help--cpu) {
  width: 429px;
  max-width: min(429px, calc(100vw - 20px)) !important;
}
.scheduling-help-popup:has(> .node-pool-settings-help--historical) {
  width: 475px;
  max-width: min(475px, calc(100vw - 20px)) !important;
}
.q-tooltip.node-pool-settings-hide-help {
  width: 147px;
}
</style>
