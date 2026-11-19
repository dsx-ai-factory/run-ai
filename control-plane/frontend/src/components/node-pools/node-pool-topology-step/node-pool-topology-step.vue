<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <runai-wizard-layout
    class="node-pool-topology-wizard"
    responsive
    :steps="steps"
    active-step="topology"
    width="1000px"
    margin-top="26px"
    left-width="200px"
    right-width="200px"
    hide-right-on-small-screen
    :disable-navigation="busy"
    @step-click="onStepClick"
  >
    <runai-form-wrapper
      ref="form"
      :form-state="draft"
      :prevent-leave="busy"
      full-width
      @leave-page="emit('leave-approved')"
    >
      <runai-expansion-wrapper role="presentation" no-first-top-padding>
        <runai-expansion-item
          :label="canSave ? 'Network & placement' : 'Network & placement View only'"
          default-opened
          disable-closing
          body-no-padding
        >
          <template #header
            ><span>Network &amp; placement</span
            ><span v-if="!canSave" class="view-only"><q-icon name="ph-fill ph-eye" /> View only</span></template
          >
          <fieldset class="topology-fields" :disabled="busy || !canSave">
            <network-topology-selection
              v-if="supports.networkTopology && (canReadTopologies || topologyCreating)"
              class="topology-selection"
              tooltip-width="300px"
              :cluster-id="pool.clusterId"
              :network-topology-id="draft.networkTopologyId"
              :network-topology-name="selectedTopologyName"
              :disable="busy || !canSave || !canReadTopologies"
              :can-create="canCreateTopology && canSave"
              @select="draft.networkTopologyId = $event"
              @loaded="topologies = $event"
              @creating="topologyCreating = $event"
            />
            <p v-else class="text-b-subtitle2" role="status">
              {{
                canReadTopologies
                  ? "Network topology selection is not supported by this cluster."
                  : "You do not have permission to view network topologies."
              }}
              The saved topology is preserved.
            </p>
            <section class="mnnvl-section">
              <runai-radio-options
                aid="topology-mnnvl"
                separate-tooltips
                title-tooltip-width="300px"
                tooltip-width="300px"
                title="Define whether MNNVL is present on this node pool"
                :title-tooltip="mnnvlTooltip"
                :options="NODE_POOL_NETWORK_DISCOVERY_OPTIONS"
                :model-value="effectiveDetection"
                :disable="busy || !canSave || !supports.mnnvl"
                @update:model-value="draft.gpuNetworkAccelerationDetection = $event"
              />
              <p v-if="!supports.mnnvl" class="text-b-body1" role="status">
                MNNVL detection settings require Run:ai cluster version 2.21 or later.
              </p>
            </section>
            <runai-section
              v-if="supports.mnnvl && draft.gpuNetworkAccelerationDetection !== Detection.DontUse"
              class="mnnvl-label-section"
              title="Set the node’s label used to detect and enable MNNVL acceleration"
            >
              <runai-name-validation
                aid="topology-mnnvl-label"
                v-model="draft.gpuNetworkAccelerationLabelKey"
                class="mnnvl-label"
                label="Label"
                stack-label
                :counter="false"
                :maxlength="317"
                :placeholder="DEFAULT_GPU_NETWORK_ACCELERATION_LABEL_KEY"
                :read-only="busy || !canSave || !supports.mnnvl"
                :rules="[validateTopologyLabelKey]"
                :aria-invalid="validationError ? 'true' : undefined"
                :aria-describedby="
                  validationError
                    ? 'node-pool-topology-label-help node-pool-topology-save-error'
                    : 'node-pool-topology-label-help'
                "
              />
              <p id="node-pool-topology-label-help" class="label-help text-b-body1">
                Use letters, numbers, '-', '_' or '.', with an optional DNS prefix followed by '/'.
              </p>
            </runai-section>
            <p class="guide-link text-b-subtitle2">
              For more information, see the
              <a :href="DOCS.TOPOLOGIES.GB200" target="_blank" rel="noopener noreferrer"
                >Using GB200 NVL72 and Multi-Node NVLink Domains guide</a
              >
            </p>
            <section v-if="supports.numa" class="numa-section">
              <div class="numa-control">
                <runai-toggle
                  aid="topology-numa"
                  v-model="draft.numaAwareScheduling"
                  label="NUMA-aware scheduling"
                  left-label
                  :disable="busy || !canSave"
                /><runai-tooltip :tooltip-text="numaTooltip" width="300px" tooltip-position="right">
                  <template #content
                    ><span class="numa-tooltip-content">{{ numaTooltip }}</span></template
                  >
                </runai-tooltip>
              </div>
              <p class="guide-link text-b-subtitle2">
                For more information on the prerequisites and configuration requirements, see the
                <a :href="DOCS.NODE_POOLS.NUMA_AWARE_SCHEDULING" target="_blank" rel="noopener noreferrer"
                  >Configuring NUMA-Aware Scheduling guide</a
                >
              </p>
            </section>
            <runai-folding-sub-section
              label="Placement strategy"
              aid="topology-placement-toggle"
              class="placement-section"
            >
              <section class="placement-group">
                <h3 class="text-b-subtitle1">GPU pods</h3>
                <runai-radio-options
                  aid="topology-gpu-node-placement"
                  separate-tooltips
                  title-tooltip-width="300px"
                  tooltip-width="300px"
                  title="Set the node placement strategy"
                  title-tooltip="Specify how GPU pods are placed across nodes."
                  :options="placementOptions.gpu"
                  v-model="draft.gpuNodePlacement"
                  horizontal
                  :disable="busy || !canSave"
                />
                <runai-radio-options
                  v-if="supports.gpuDevice"
                  aid="topology-gpu-device-placement"
                  separate-tooltips
                  title-tooltip-width="300px"
                  tooltip-width="300px"
                  title="Set the device placement strategy"
                  title-tooltip="Specify how GPU pods are placed across GPU devices within a node when using GPU fractions."
                  :options="placementOptions.gpuDevice"
                  v-model="draft.gpuDevicePlacement"
                  horizontal
                  :disable="busy || !canSave"
                />
              </section>
              <section class="placement-group">
                <h3 class="text-b-subtitle1">CPU pods</h3>
                <runai-radio-options
                  aid="topology-cpu-node-placement"
                  separate-tooltips
                  title-tooltip-width="300px"
                  tooltip-width="300px"
                  title="Set the node placement strategy"
                  title-tooltip="Specify how CPU pods are placed across nodes."
                  :options="placementOptions.cpu"
                  v-model="draft.cpuNodePlacement"
                  horizontal
                  :disable="busy || !canSave"
                />
              </section>
            </runai-folding-sub-section>
          </fieldset>
        </runai-expansion-item>
      </runai-expansion-wrapper>
    </runai-form-wrapper>
    <template #content-footer
      ><div class="topology-footer">
        <p
          v-if="saveError || validationError"
          id="node-pool-topology-save-error"
          class="text-negative"
          :role="saveError ? 'alert' : undefined"
        >
          {{ saveError || validationError }}
        </p>
        <div class="row q-gutter-md">
          <runai-btn
            aid="topology-cancel"
            flat
            label="Cancel"
            color="primary"
            :disable="busy"
            @click="emit('cancel')"
          /><runai-btn
            aid="topology-save-close"
            label="Save & close"
            color="primary"
            :loading="saving"
            :disable="busy || !canSave || !!validationError"
            @click="save"
          />
        </div></div
    ></template>
    <template #right-sidebar
      ><node-pool-wizard-summary
        :pool="previewPool"
        :resources="resources"
        :show-numa="supports.numa"
        :show-gpu-device="supports.gpuDevice"
    /></template>
  </runai-wizard-layout>
</template>
<script setup lang="ts">
import { computed, ref, watch } from "vue";
import RunaiWizardLayout from "@/components/common/runai-wizard-layout/runai-wizard-layout.vue";
import RunaiFormWrapper from "@/components/common/runai-form-wrapper/runai-form-wrapper.vue";
import RunaiExpansionWrapper from "@/components/common/runai-expansion-wrapper/runai-expansion-wrapper.vue";
import RunaiExpansionItem from "@/components/common/runai-expansion-item/runai-expansion-item.vue";
import RunaiSection from "@/components/common/runai-section/runai-section.vue";
import RunaiRadioOptions from "@/components/common/runai-radio-options/runai-radio-options.vue";
import RunaiNameValidation from "@/components/common/runai-name-validation/runai-name-validation.vue";
import RunaiToggle from "@/components/common/runai-toggle/runai-toggle.vue";
import RunaiTooltip from "@/components/common/runai-tooltip/runai-tooltip.vue";
import RunaiFoldingSubSection from "@/components/common/runai-folding-sub-section/runai-folding-sub-section.vue";
import RunaiBtn from "@/components/common/runai-btn/runai-btn.vue";
import NetworkTopologySelection from "@/components/node-pools/node-pool-edit-form/node-pool-network-section/network-topology-selection/network-topology-selection.vue";
import NodePoolWizardSummary from "@/components/node-pools/node-pool-wizard-summary/node-pool-wizard-summary.vue";
import {
  DEFAULT_GPU_NETWORK_ACCELERATION_LABEL_KEY,
  NODE_POOL_NETWORK_DISCOVERY_OPTIONS,
  PlacementOptions,
  type INodePoolSummaryResources,
} from "@/models/node-pool.model";
import {
  NodepoolGPUNetworkAccelerationDetection as Detection,
  type NetworkTopology,
  type Nodepool,
} from "@/swagger-models/cluster-service-client";
import { DOCS } from "@/constants/docs";
import { extractErrorMessages } from "@/utils/error-alert.util";
import { topologyDraftFromPool, buildTopologyPatch, validateTopologyLabelKey } from "./node-pool-topology.model";
import { saveTopologyPool } from "./node-pool-topology.service";
const props = defineProps<{
  pool: Nodepool;
  resources?: INodePoolSummaryResources;
  supports: { mnnvl: boolean; numa: boolean; gpuDevice: boolean; networkTopology: boolean };
  canSave: boolean;
  canReadTopologies: boolean;
  canCreateTopology: boolean;
}>();
const emit = defineEmits<{ saved: [Nodepool]; saving: [boolean]; "leave-approved": []; cancel: []; back: [] }>();
const draft = ref(topologyDraftFromPool(props.pool));
const effectiveDetection = computed(() =>
  props.supports.mnnvl ? draft.value.gpuNetworkAccelerationDetection : Detection.Auto,
);
const saving = ref(false);
const topologyCreating = ref(false);
const busy = computed(() => saving.value || topologyCreating.value);
watch(busy, (value) => emit("saving", value), { flush: "sync" });
const saveError = ref("");
const form = ref<InstanceType<typeof RunaiFormWrapper>>();
const topologies = ref<NetworkTopology[]>([]);
const selectedTopologyName = computed(
  () =>
    topologies.value.find((t) => t.id === draft.value.networkTopologyId)?.name ||
    (draft.value.networkTopologyId === props.pool.networkTopologyId ? props.pool.networkTopologyName : undefined) ||
    "",
);
const previewPool = computed(() => ({
  ...props.pool,
  networkTopologyId: draft.value.networkTopologyId,
  networkTopologyName: selectedTopologyName.value,
  gpuNetworkAccelerationDetection: effectiveDetection.value,
  schedulingConfiguration: {
    ...props.pool.schedulingConfiguration,
    numaAwareScheduling: draft.value.numaAwareScheduling,
    gpuPodsPlacementStrategy: { node: draft.value.gpuNodePlacement, device: draft.value.gpuDevicePlacement },
    cpuPodsPlacementStrategy: { node: draft.value.cpuNodePlacement },
  },
}));
const validationError = computed(() => {
  try {
    buildTopologyPatch(draft.value, props.supports);
    return "";
  } catch (error) {
    return extractErrorMessages(error).join("; ");
  }
});
const steps = [
  { name: "identity", title: "Cluster & name", done: true, disabled: true },
  { name: "scheduling", title: "Scheduling mode", done: true },
  { name: "topology", title: "Topology & placement", done: false },
];
const mnnvlTooltip =
  "If this node pool is configured as Multi-Node NVLink (MNNVL) capable, NVIDIA Run:ai leverages its high-bandwidth GPU interconnects to accelerate workloads.";
const numaTooltip =
  "When enabled, the Scheduler places workloads so that GPU, CPU, and CPU memory are allocated within the same NUMA node where possible, minimizing cross-NUMA fragmentation and improving performance. There are cluster prerequisites for enabling this feature.\n\nWhen disabled, the Scheduler places workloads using aggregate node capacity only, without considering NUMA topology.";
// Step 3 copy differs from the legacy Scheduling section; keep it scoped here.
const placementOptions = {
  gpu: PlacementOptions.gpu.map((o, i) => ({
    ...o,
    toolTip: i
      ? "Distribute GPU pods across as many nodes as possible to balance the load across nodes."
      : "Place as many GPU pods as possible on each node to use fewer nodes and maximize available capacity.",
  })),
  gpuDevice: PlacementOptions.gpuDevice.map((o, i) => ({
    ...o,
    toolTip: i
      ? "Distribute GPU pods across as many GPU devices as possible to balance the load across devices."
      : "Place as many GPU pods as possible on each GPU device to use fewer devices and maximize available capacity, minimizing GPU fraction fragmentation.",
  })),
  cpu: PlacementOptions.cpu.map((o, i) => ({
    ...o,
    toolTip: i
      ? "Distribute CPU pods across as many nodes as possible to balance the load across nodes."
      : "Place as many CPU pods as possible on each node to use fewer nodes and maximize available capacity.",
  })),
};
function onStepClick({ name }: { name: string }) {
  if (!busy.value && name === "scheduling") emit("back");
}
async function save() {
  if (busy.value || !props.canSave || validationError.value) return;
  saving.value = true;
  saveError.value = "";
  try {
    if (!(await form.value?.validate())) return;
    const saved = await saveTopologyPool(props.pool.id, buildTopologyPatch(draft.value, props.supports));
    saving.value = false;
    emit("saved", saved);
  } catch (error) {
    saveError.value = extractErrorMessages(
      error,
      "Could not save topology and placement. Your changes are preserved.",
    ).join("; ");
  } finally {
    saving.value = false;
  }
}
</script>
<style scoped lang="scss">
.topology-fields {
  border: 0;
  margin: 0;
  padding: 20px 16px;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 26px;
}
.topology-selection {
  margin-bottom: 0;
}
.topology-selection :deep(.runai-select) {
  width: 250px;
  max-width: 100%;
  flex: none;
}
.mnnvl-label {
  width: 510px;
  max-width: 100%;
}
.label-help {
  margin: -12px 0 0;
  color: var(--b-text-sub);
}
.guide-link {
  margin: 0;
  font-style: italic;
}
.guide-link a {
  text-decoration: underline;
}
.numa-tooltip-content {
  white-space: pre-line;
}
.numa-section {
  margin-top: 26px;
}
.numa-control {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 8px;
}
.placement-section {
  margin-top: 4px;
}
.placement-group {
  display: flex;
  flex-direction: column;
  gap: 20px;
  margin-top: 26px;
}
.placement-group :deep(.runai-radio-options > span) {
  width: 100%;
}
.placement-group :deep(.horizontal-options),
.placement-group :deep(.horizontal-options > div) {
  margin-left: 0;
}
.placement-group h3 {
  font-size: 14px;
  line-height: 20px;
  margin: 0;
}
.topology-footer {
  flex-wrap: wrap;
  padding: 40px 0 8px;
  display: flex;
  justify-content: flex-end;
  gap: 20px;
  align-items: center;
}
.topology-footer > p {
  flex: 1;
  margin: 0;
}
.view-only {
  font-size: 12px;
  color: var(--b-text-sub);
  margin-left: auto;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.node-pool-topology-wizard :deep(.wizard-step-title) {
  white-space: nowrap;
  font-weight: 500;
}
.node-pool-topology-wizard :deep(.wizard-step-item) {
  padding-top: 7px;
  padding-bottom: 7px;
}
.node-pool-topology-wizard :deep(.runai-option-group) {
  max-width: 100%;
}
@media (prefers-reduced-motion: reduce) {
  .node-pool-topology-wizard :deep(.folding-icon) {
    transition: none;
  }
}
</style>
