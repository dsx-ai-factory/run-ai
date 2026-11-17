<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <aside
    class="scheduling-settings"
    :class="{ 'scheduling-settings--compact': draft.mode === 'proportional' }"
    aria-label="Advanced settings"
  >
    <header>
      <span>Advanced settings</span
      ><q-btn flat round dense icon="ph-bold ph-x" aria-label="Hide Advanced settings" @click="emit('close')">
        <q-tooltip class="node-pool-settings-hide-help">Hide Advanced settings</q-tooltip>
      </q-btn>
    </header>
    <div class="settings-body">
      <section v-if="draft.mode === 'guaranteed'">
        <runai-section-title
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
        </runai-section-title>
        <runai-radio-options
          :model-value="draft.automaticWeights ? 'automatic' : 'custom'"
          :options="weightOptions"
          :disable="readonly"
          @update:model-value="patch({ automaticWeights: $event === 'automatic' })"
        />
      </section>
      <section v-if="draft.mode === 'guaranteed'" class="toggle-section">
        <runai-section-title title="Max GPU allocation" tooltip-icon="ph-bold ph-info" :tooltip-text="maxHelp">
          <template #tooltip-content>
            <span class="node-pool-settings-help node-pool-settings-help--standard">{{ maxHelp }}</span>
          </template>
        </runai-section-title>
        <runai-toggle
          :model-value="!!draft.maxAllocation"
          :disable="readonly || (!canResetResources && (savedMaxAllocation ?? draft.maxAllocation))"
          aid="max-allocation-toggle"
          @update:model-value="setMax"
        />
      </section>
      <section>
        <div class="toggle-section">
          <runai-section-title
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
          </runai-section-title>
          <runai-toggle
            :model-value="!!draft.timeFairshare"
            :disable="readonly"
            aid="time-fairshare-toggle"
            @update:model-value="patch({ timeFairshare: $event })"
          />
        </div>
        <div v-if="draft.timeFairshare" class="historical-fields">
          <div class="historical-weight-group">
            <runai-section-title
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
              </template></runai-section-title
            >
            <div class="historical-weight-field">
              <q-input
                stack-label
                class="historical-weight-input"
                :model-value="draft.historicalWeight"
                @update:model-value="patch({ historicalWeight: number($event) })"
                type="number"
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
            <runai-section-title
              flat
              title="Set the duration of the historical usage window"
              width="374px"
              :tooltip-text="windowHelp"
            >
              <template #tooltip-content>
                <span class="node-pool-settings-help node-pool-settings-help--standard">{{ windowHelp }}</span>
              </template>
            </runai-section-title>
            <div class="duration-grid historical-window">
              <q-input
                v-for="unit in windowUnits"
                :key="unit"
                :model-value="draft.historicalWindow?.[unit] ?? 0"
                @update:model-value="setDuration('historicalWindow', unit, $event)"
                stack-label
                type="number"
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
        <runai-section-title
          title="Set the minimum guaranteed runtime"
          tooltip-icon="ph-bold ph-info"
          :tooltip-text="runtimeHelp"
        >
          <template #tooltip-content>
            <span class="node-pool-settings-help node-pool-settings-help--standard">{{ runtimeHelp }}</span>
          </template>
        </runai-section-title>
        <div class="duration-grid">
          <q-input
            v-for="unit in runtimeUnits"
            :key="unit"
            :model-value="draft.minimumRuntime?.[unit] ?? 0"
            @update:model-value="setDuration('minimumRuntime', unit, $event)"
            stack-label
            type="number"
            min="0"
            :label="unit"
            :disable="readonly"
            :error="invalid(draft.minimumRuntime?.[unit])"
            error-message="Enter a timeframe of at least 0 seconds"
            no-error-icon
          />
        </div>
      </section>
      <template v-if="draft.mode === 'guaranteed'">
        <section class="toggle-section">
          <runai-section-title
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
          </runai-section-title>
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
        confirmation === "maxAllocation" ? "Disable Max GPU Allocation?" : "Disable CPU Quota?"
      }}</template>
      <template #body>{{
        confirmation === "maxAllocation"
          ? "On save, the backend removes existing department and project limits. Unsaved values remain in this draft."
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
          label="Disable"
          aid="confirm-disable-setting"
          @click="confirmDisable"
      /></template>
    </runai-base-dialog>
  </aside>
</template>
<script setup lang="ts">
import { ref } from "vue";
import { DOCS } from "@/constants/docs";
import { RunaiSectionTitle } from "@/components/common/runai-section-title";
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import { RunaiToggle } from "@/components/common/runai-toggle";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import RunaiBtn from "@/components/common/runai-btn/runai-btn.vue";
import type { SchedulingDraft } from "./node-pool-scheduling.model";
const props = defineProps<{
  draft: SchedulingDraft;
  readonly?: boolean;
  canResetResources?: boolean;
  savedMaxAllocation?: boolean;
  savedCpuQuota?: boolean;
}>();
const emit = defineEmits<{ close: []; patch: [Partial<SchedulingDraft>] }>();
const weightOptions = [
  {
    label: "Automatic",
    value: "automatic",
    toolTipClass: "node-pool-settings-option-help",
    toolTipWidth: "374px",
    toolTip:
      "A project or department gets a proportion of unused resources relative to its quota, and it can't be modified.",
  },
  {
    label: "Custom",
    value: "custom",
    toolTipClass: "node-pool-settings-option-help",
    toolTipWidth: "374px",
    toolTip: "Set the proportion of unused resources a project or department can get on top of its quota.",
  },
];
const maxHelp =
  'When enabled, the maximum GPU device allocation each department and project can get from this node pool, representing the maximum sum of assigned quota and over-quota GPUs, can be set. When disabled, the maximum is "unlimited."';
const runtimeHelp =
  "This defines the minimum time a preemptible workload is guaranteed to run once it is scheduled in this node pool, before it can be preempted and its resources reallocated to other workloads.";
const windowHelp =
  "This defines the duration of the sliding window used to calculate historical resource usage. At any given moment, the system looks only at usage within this duration from the current time. As time passes, the window slides forward continuously, including new usage and dropping old usage that falls outside the window.";
const runtimeUnits = ["Days", "Hours", "Minutes", "Seconds"];
const windowUnits = ["Days", "Hours", "Minutes"];
const confirmation = ref<"maxAllocation" | "cpuQuota">();
function patch(value: Partial<SchedulingDraft>) {
  if (!props.readonly) emit("patch", value);
}
function invalid(value: number | null | undefined) {
  return value != null && (!Number.isFinite(value) || value < 0);
}
function number(value: string | number | null) {
  return value === "" || value === null ? null : Number(value);
}
function setDuration(field: "historicalWindow" | "minimumRuntime", unit: string, value: string | number | null) {
  patch({ [field]: { ...props.draft[field], [unit]: number(value) } });
}
function setMax(value: boolean) {
  if (
    props.draft.mode !== "guaranteed" ||
    props.readonly ||
    (!value && (props.savedMaxAllocation ?? props.draft.maxAllocation) && !props.canResetResources)
  )
    return;
  if (!value && (props.savedMaxAllocation ?? props.draft.maxAllocation)) confirmation.value = "maxAllocation";
  else patch({ maxAllocation: value });
}
function setCpu(value: boolean) {
  if (props.readonly || (!value && (props.savedCpuQuota ?? props.draft.cpuQuota) && !props.canResetResources)) return;
  if (!value && (props.savedCpuQuota ?? props.draft.cpuQuota)) confirmation.value = "cpuQuota";
  else patch({ cpuQuota: value });
}
function confirmDisable() {
  if (confirmation.value === "maxAllocation") {
    // The backend resets stored limits on save. Keep inactive draft limits so
    // switching the toggle back on restores the user's unsaved values.
    patch({ maxAllocation: false });
  } else patch({ cpuQuota: false });
  confirmation.value = undefined;
}
</script>
<style scoped lang="scss">
.scheduling-settings {
  width: 380px;
  flex: 0 0 380px;
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
  color: rgba(0, 0, 0, 0.54);
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
.settings-body :deep(.runai-radio-options) {
  margin-top: 7px;
}
.settings-body :deep(.runai-section .q-field) {
  display: none;
}
.settings-body :deep(.runai-section-title) {
  font-size: 14px;
  font-weight: 400;
  line-height: 20px;
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
  grid-template-columns: 1fr 1fr;
  gap: 0 20px;
  align-items: center;
  margin-top: 10px;
}
.historical-window {
  align-items: start;
  grid-template-columns: repeat(3, 1fr);
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
  padding-left: 20px;
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
.runai-tooltip:has(> .node-pool-settings-help),
.runai-tooltip.node-pool-settings-option-help,
.q-tooltip.node-pool-settings-hide-help {
  padding: 10px;
  border-radius: 3px;
  background: white;
  color: rgba(0, 0, 0, 0.87);
  box-shadow: -1px 0 7px rgba(0, 0, 0, 0.15);
  font-family: Roboto, sans-serif;
  font-size: 12px;
  font-weight: 400;
  line-height: 16.8px;
}
.runai-tooltip:has(> .node-pool-settings-help--standard),
.runai-tooltip.node-pool-settings-option-help {
  width: 374px;
  max-width: min(374px, calc(100vw - 20px)) !important;
}
.runai-tooltip:has(> .node-pool-settings-help--fairshare) {
  width: 427px;
  max-width: min(427px, calc(100vw - 20px)) !important;
}
.runai-tooltip:has(> .node-pool-settings-help--cpu) {
  width: 429px;
  max-width: min(429px, calc(100vw - 20px)) !important;
}
.runai-tooltip:has(> .node-pool-settings-help--historical) {
  width: 475px;
  max-width: min(475px, calc(100vw - 20px)) !important;
}
.q-tooltip.node-pool-settings-hide-help {
  width: 147px;
}
</style>
