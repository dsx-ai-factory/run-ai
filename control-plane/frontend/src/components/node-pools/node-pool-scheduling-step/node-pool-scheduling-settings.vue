<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <aside class="scheduling-settings" aria-label="Advanced settings">
    <header>
      <span>Advanced settings</span
      ><q-btn flat round dense icon="ph ph-x" aria-label="Hide Advanced settings" @click="emit('close')" />
    </header>
    <div class="settings-body">
      <section v-if="draft.mode === 'guaranteed'">
        <runai-section-title
          title="Weight within rank"
          tooltip-icon="ph ph-info"
          tooltip-text="Set the proportion of unused resources a project or department can get on top of its quota."
        />
        <runai-radio-options
          :model-value="draft.automaticWeights ? 'automatic' : 'custom'"
          :options="weightOptions"
          :disable="readonly"
          @update:model-value="patch({ automaticWeights: $event === 'automatic' })"
        />
      </section>
      <section class="toggle-section">
        <runai-section-title title="Max GPU allocation" tooltip-icon="ph ph-info" :tooltip-text="maxHelp" />
        <runai-toggle
          :model-value="!!draft.maxAllocation"
          :disable="readonly"
          aid="max-allocation-toggle"
          @update:model-value="setMax"
        />
      </section>
      <section>
        <div class="toggle-section">
          <runai-section-title
            title="Time-based fairshare"
            tooltip-icon="ph ph-info"
            width="425px"
            tooltip-text="When enabled, the Scheduler adjusts fairshare scores based on each project's weight and historical usage records. To use this feature, persistent data must be enabled. When disabled, the Scheduler uses the default fairshare, with scheduling fairness based on the current project's weight and resource usage. For more information, see the Time-based fairshare guide."
          />
          <runai-toggle
            :model-value="!!draft.timeFairshare"
            :disable="readonly"
            aid="time-fairshare-toggle"
            @update:model-value="patch({ timeFairshare: $event })"
          />
        </div>
        <div v-if="draft.timeFairshare" class="historical-fields">
          <runai-section-title
            title="Set the weight value that determines how a project's historical resource usage affects its fairshare"
            width="473px"
            tooltip-text="This determines how strongly historical resource usage influences fairshare scores: 0 — History is ignored; fairshare is based on current weight only. 0–1 — History has partial influence; lower values reduce its effect. 1 — History and current weight are balanced equally. >1 — History dominates; higher values amplify its effect."
            ><template #tooltip-content
              >This determines how strongly historical resource usage influences fairshare scores:
              <ul class="historical-weight-help">
                <li>0 — History is ignored; fairshare is based on current weight only.</li>
                <li>0–1 — History has partial influence; lower values reduce its effect.</li>
                <li>1 — History and current weight are balanced equally.</li>
                <li>&gt;1 — History dominates; higher values amplify its effect.</li>
              </ul>
            </template></runai-section-title
          >
          <q-input
            :model-value="draft.historicalWeight"
            @update:model-value="patch({ historicalWeight: number($event) })"
            type="number"
            min="0"
            label="Historical usage weight"
            :disable="readonly"
            :error="invalid(draft.historicalWeight)"
            error-message="Enter a value equal to or above 0"
            no-error-icon
          />
          <runai-section-title
            class="q-mt-lg"
            title="Set the duration of the historical usage window"
            width="372px"
            tooltip-text="This defines the duration of the sliding window used to calculate historical resource usage. At any given moment, the system looks only at usage within this duration from the current time. As time passes, the window slides forward continuously, including new usage and dropping old usage that falls outside the window."
          />
          <div class="duration-grid historical-window">
            <q-input
              v-for="unit in windowUnits"
              :key="unit"
              :model-value="draft.historicalWindow?.[unit] ?? 0"
              @update:model-value="setDuration('historicalWindow', unit, $event)"
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
      </section>
      <section class="runtime-section">
        <runai-section-title
          title="Set the minimum guaranteed runtime"
          tooltip-icon="ph ph-info"
          tooltip-text="This defines the minimum time a preemptible workload is guaranteed to run once it is scheduled in this node pool, before it can be preempted and its resources reallocated to other workloads."
        />
        <div class="duration-grid">
          <q-input
            v-for="unit in runtimeUnits"
            :key="unit"
            :model-value="draft.minimumRuntime?.[unit] ?? 0"
            @update:model-value="setDuration('minimumRuntime', unit, $event)"
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
            tooltip-icon="ph ph-info"
            tooltip-text="When enabled, CPU resources can be set per department or project. For more information, see the Departments guide and Projects guide."
          />
          <runai-toggle
            :model-value="draft.cpuQuota"
            :disable="readonly"
            aid="cpu-quota-toggle"
            @update:model-value="setCpu"
          />
        </section>
        <section class="toggle-section quota-policy-section">
          <runai-section-title
            title="Limit projects from exceeding department quota"
            tooltip-icon="ph ph-info"
            tooltip-text="When enabled, the projects' total quota cannot exceed the department's quota or be less than the total usage by non-preemptible workloads. For more information, see The NVIDIA Run:ai Scheduler: Concepts and Principles guide."
          />
          <runai-toggle
            :model-value="quotaPolicy ?? false"
            disable
            aid="project-quota-policy"
            :tooltip="
              quotaPolicy === undefined
                ? 'Quota policy is unavailable.'
                : 'This organization-wide policy is read-only here.'
            "
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
          ? "You have unsaved changes. If you disable max GPU allocation, they will be lost."
          : "Departments and projects using this node pool will lose any assigned CPU quota."
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
import { RunaiSectionTitle } from "@/components/common/runai-section-title";
import { RunaiRadioOptions } from "@/components/common/runai-radio-options";
import { RunaiToggle } from "@/components/common/runai-toggle";
import { RunaiBaseDialog } from "@/components/common/runai-base-dialog";
import RunaiBtn from "@/components/common/runai-btn/runai-btn.vue";
import type { SchedulingDraft } from "./node-pool-scheduling.model";
const props = defineProps<{ draft: SchedulingDraft; quotaPolicy?: boolean; readonly?: boolean }>();
const emit = defineEmits<{ close: []; patch: [Partial<SchedulingDraft>] }>();
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
const maxHelp =
  "When enabled, the maximum GPU device allocation each department and project can get from this node pool, representing the maximum sum of assigned quota and over-quota GPUs, can be set. When disabled, the maximum is unlimited.";
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
  if (!value && Object.values(props.draft.departments).some((edit) => edit.limit !== undefined))
    confirmation.value = "maxAllocation";
  else patch({ maxAllocation: value });
}
function setCpu(value: boolean) {
  if (!value) confirmation.value = "cpuQuota";
  else patch({ cpuQuota: true });
}
function confirmDisable() {
  if (confirmation.value === "maxAllocation") {
    const departments = Object.fromEntries(
      Object.entries(props.draft.departments).map(([id, edit]) => {
        const { limit: _limit, ...rest } = edit;
        return [id, rest];
      }),
    );
    patch({ maxAllocation: false, departments });
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
header {
  height: 55px;
  padding: 0 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  border-bottom: 1px solid var(--border-subtle);
  font-weight: 500;
}
header .q-btn {
  color: var(--icon-default);
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
.quota-policy-section :deep(.runai-section-title) {
  max-width: 245px;
}
.settings-body > .runtime-section:last-child {
  padding-bottom: 10px !important;
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
.duration-grid :deep(.q-field__bottom) {
  min-height: 16px;
  padding-top: 0;
}
.duration-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 18px 20px;
  margin-top: 12px;
}
.historical-window {
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}
.runtime-section {
  padding-bottom: 36px !important;
}
.historical-fields {
  margin-top: 26px;
}
.duration-grid :deep(.q-field__label) {
  font-size: 16px;
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
