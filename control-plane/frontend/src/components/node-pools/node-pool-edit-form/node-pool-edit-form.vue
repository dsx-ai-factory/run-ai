<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="column items-center q-pt-md">
    <runai-form-wrapper :form-state="formData" ref="elNodePoolEditCreateForm">
      <runai-expansion-wrapper>
        <runai-cluster-section
          :disabled="!isNewNodePool"
          entity-name="node pool"
          :cluster-id="formData.clusterId"
          @update:cluster-id="updateClusterId"
        />
        <template v-if="formData.clusterId">
          <node-pool-name-section
            v-model:node-pool-name="formData.name"
            :cluster-id="formData.clusterId"
            :loading="loading"
            :disabled="!isNewNodePool"
            @is-section-invalid="invalidSection.nodePoolName = !$event"
          />
          <node-pool-label-section
            :loading="loading"
            v-model:label-key="formData.labelKey"
            v-model:label-value="formData.labelValue"
            :node-pools="nodePools"
            :disabled="!isNewNodePool"
            @is-section-invalid="invalidSection.nodePoolLabel = !$event"
          />
          <node-pool-placement-strategy-section
            :min-guaranteed-runtime="minGuaranteedRuntime"
            :gpu-placement-strategy="placementStrategyGpu"
            :cpu-placement-strategy="placementStrategyCpu"
            :gpu-device-placement-strategy="placementStrategyGpuDevice"
            :show-gpu-device-placement="showGpuDevicePlacement"
            :allow-time-based-fair-share="allowTimeBasedFairShare"
            :historical-usage-weight="historicalUsageWeight"
            :window-duration="windowDuration"
            :is-cluster-supports-min-guaranteed-runtime="isClusterSupportsMinGuaranteedRuntime"
            :is-cluster-supports-time-based-fair-share="isClusterSupportsTimeBasedFairShare"
            :numa-aware-scheduling="numaAwareScheduling"
            :is-cluster-supports-numa-aware-scheduling="isClusterSupportsNumaAwareScheduling"
            @update:min-guaranteed-runtime="updateMinGuaranteedRuntime"
            @update:gpu-placement-strategy="updatePlacementStrategyGpu"
            @update:cpu-placement-strategy="updatePlacementStrategyCpu"
            @update:gpu-device-placement-strategy="updatePlacementStrategyGpuDevice"
            @update:allow-time-based-fair-share="updateAllowTimeBasedFairShare"
            @update:historical-usage-weight="updateHistoricalUsageWeight"
            @update:window-duration="updateWindowDuration"
            @update:numa-aware-scheduling="updateNumaAwareScheduling"
          />
          <node-pool-network-section
            :cluster-id="formData.clusterId"
            v-model:gpu-network-acceleration-label-key="gpuNetworkAccelerationLabelKey"
            v-model:gpu-network-acceleration-detection="formData.gpuNetworkAccelerationDetection"
            v-model:network-topology-id="formData.networkTopologyId"
          />
          <runai-form-footer
            class="q-mt-md"
            :loading="isNodePoolSubmitting"
            :save-btn-label="isNewNodePool ? 'create node pool' : 'save'"
            :is-form-valid="isFormValid"
            @cancel="$emit('on-cancel')"
            @save="save"
          />
        </template>
      </runai-expansion-wrapper>
    </runai-form-wrapper>
  </div>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";

// util
import { deepCopy, isNullOrUndefined } from "@/utils/common.util";

// model

// cmps
import { RunaiFormWrapper } from "@/components/common/runai-form-wrapper";
import { RunaiExpansionWrapper } from "@/components/common/runai-expansion-wrapper";
import { NodePoolNameSection } from "@/components/node-pools/node-pool-edit-form/node-pool-name-section/";
import { NodePoolLabelSection } from "@/components/node-pools/node-pool-edit-form/node-pool-label-section/";
import { NodePoolPlacementStrategySection } from "@/components/node-pools/node-pool-edit-form/node-pool-placement-strategy-section/";
// const

import {
  MIN_CLUSTER_VERSION_FOR_MINIMUM_GUARANTEED_RUNTIME,
  MIN_CLUSTER_VERSION_FOR_TIME_BASED_FAIR_SHARE,
  MIN_CLUSTER_VERSION_FOR_GPU_DEVICE_PLACEMENT_STRATEGY,
  MIN_CLUSTER_VERSION_FOR_NUMA_AWARE_SCHEDULING,
} from "@/common/version.constant";
// store
import { useClusterStore } from "@/stores/cluster.store";
import type { Nodepool, NodepoolCreateFields } from "@/swagger-models/cluster-service-client";
import { PlacementStrategy } from "@/swagger-models/cluster-service-client";
import RunaiFormFooter from "@/components/common/runai-form-footer/runai-form-footer.vue";
import { RunaiClusterSection } from "@/components/common/runai-cluster-selection/runai-cluster-section";
import { NodePoolNetworkSection } from "@/components/node-pools/node-pool-edit-form/node-pool-network-section";

export interface ISectionValidation {
  nodePoolScope: boolean;
  nodePoolName: boolean;
  nodePoolLabel: boolean;
  networkSection: boolean;
}
export default defineComponent({
  name: "node-pool-edit-form",
  components: {
    NodePoolNetworkSection,
    RunaiClusterSection,
    RunaiFormFooter,
    NodePoolPlacementStrategySection,
    NodePoolLabelSection,
    NodePoolNameSection,
    RunaiExpansionWrapper,
    RunaiFormWrapper,
  },
  emits: ["on-save", "on-cancel", "update-cluster-id"],
  props: {
    nodePool: {
      type: Object as PropType<Nodepool | NodepoolCreateFields>,
      required: true,
    },
    isNewNodePool: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    isNodePoolSubmitting: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    nodePools: {
      type: Array as PropType<Nodepool[]>,
      default: () => [],
    },
    loading: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      formData: deepCopy(this.nodePool) as NodepoolCreateFields | Nodepool,
      isFormValid: true as boolean,
      clusterStore: useClusterStore(),
      invalidSection: {
        nodePoolName: false as boolean,
        nodePoolLabel: false as boolean,
      } as ISectionValidation,
    };
  },

  computed: {
    placementStrategyGpu(): PlacementStrategy {
      return this.formData.schedulingConfiguration?.gpuPodsPlacementStrategy?.node || PlacementStrategy.Binpack;
    },
    placementStrategyCpu(): PlacementStrategy {
      return this.formData.schedulingConfiguration?.cpuPodsPlacementStrategy?.node || PlacementStrategy.Binpack;
    },
    allowTimeBasedFairShare(): boolean {
      return this.formData.schedulingConfiguration?.timeBasedFairShare?.enabled ?? false;
    },
    historicalUsageWeight(): number | null {
      return this.formData.schedulingConfiguration?.timeBasedFairShare?.historicalUsageWeight ?? null;
    },
    windowDuration(): string {
      return this.formData.schedulingConfiguration?.timeBasedFairShare?.windowDuration ?? "";
    },
    minGuaranteedRuntime(): string {
      const runtime = this.formData.schedulingConfiguration?.minGuaranteedRuntime;
      return runtime === null ? "" : (runtime ?? "");
    },
    isClusterSupportsMinGuaranteedRuntime(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.formData.clusterId,
        MIN_CLUSTER_VERSION_FOR_MINIMUM_GUARANTEED_RUNTIME,
      );
    },
    isClusterSupportsTimeBasedFairShare(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.formData.clusterId,
        MIN_CLUSTER_VERSION_FOR_TIME_BASED_FAIR_SHARE,
      );
    },
    isClusterSupportsGpuDevicePlacement(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.formData.clusterId,
        MIN_CLUSTER_VERSION_FOR_GPU_DEVICE_PLACEMENT_STRATEGY,
      );
    },
    isClusterSupportsNumaAwareScheduling(): boolean {
      return this.clusterStore.isClusterVersionSufficient(
        this.formData.clusterId,
        MIN_CLUSTER_VERSION_FOR_NUMA_AWARE_SCHEDULING,
      );
    },
    numaAwareScheduling(): boolean {
      return this.formData.schedulingConfiguration?.numaAwareScheduling ?? false;
    },
    showGpuDevicePlacement(): boolean {
      return this.isClusterSupportsGpuDevicePlacement;
    },
    placementStrategyGpuDevice(): PlacementStrategy {
      return this.formData.schedulingConfiguration?.gpuPodsPlacementStrategy?.device || PlacementStrategy.Binpack;
    },
    gpuNetworkAccelerationLabelKey: {
      get(): string | undefined {
        return this.formData.gpuNetworkAccelerationLabelKey || undefined;
      },
      set(value: string | undefined): void {
        this.formData.gpuNetworkAccelerationLabelKey = value || null;
      },
    },
  },
  methods: {
    updatePlacementStrategyGpu(value: PlacementStrategy): void {
      if (!this.formData.schedulingConfiguration) {
        this.formData.schedulingConfiguration = {};
      }
      if (!this.formData.schedulingConfiguration.gpuPodsPlacementStrategy) {
        this.formData.schedulingConfiguration.gpuPodsPlacementStrategy = {};
      }
      this.formData.schedulingConfiguration.gpuPodsPlacementStrategy.node = value;
    },
    updatePlacementStrategyCpu(value: PlacementStrategy): void {
      if (!this.formData.schedulingConfiguration) {
        this.formData.schedulingConfiguration = {};
      }
      if (!this.formData.schedulingConfiguration.cpuPodsPlacementStrategy) {
        this.formData.schedulingConfiguration.cpuPodsPlacementStrategy = {};
      }
      this.formData.schedulingConfiguration.cpuPodsPlacementStrategy.node = value;
    },
    updatePlacementStrategyGpuDevice(value: PlacementStrategy): void {
      if (!this.formData.schedulingConfiguration) {
        this.formData.schedulingConfiguration = {};
      }
      if (!this.formData.schedulingConfiguration.gpuPodsPlacementStrategy) {
        this.formData.schedulingConfiguration.gpuPodsPlacementStrategy = {};
      }
      this.formData.schedulingConfiguration.gpuPodsPlacementStrategy.device = value;
    },
    updateAllowTimeBasedFairShare(value: boolean): void {
      if (!this.formData.schedulingConfiguration) {
        this.formData.schedulingConfiguration = {};
      }
      const config = this.formData.schedulingConfiguration;
      if (isNullOrUndefined(config.timeBasedFairShare) || (config.timeBasedFairShare?.enabled ?? false) !== value) {
        config.timeBasedFairShare = value ? { ...(config.timeBasedFairShare || {}), enabled: true } : { enabled: false };
      }
    },
    updateHistoricalUsageWeight(value: number | null): void {
      if (!this.formData.schedulingConfiguration) {
        this.formData.schedulingConfiguration = {};
      }
      if (!this.formData.schedulingConfiguration.timeBasedFairShare) {
        this.formData.schedulingConfiguration.timeBasedFairShare = { enabled: false };
      }
      this.formData.schedulingConfiguration.timeBasedFairShare.historicalUsageWeight = value;
    },
    updateWindowDuration(value: string): void {
      if (!this.formData.schedulingConfiguration) {
        this.formData.schedulingConfiguration = {};
      }
      if (!this.formData.schedulingConfiguration.timeBasedFairShare) {
        this.formData.schedulingConfiguration.timeBasedFairShare = { enabled: false };
      }
      this.formData.schedulingConfiguration.timeBasedFairShare.windowDuration = value || null;
    },
    updateMinGuaranteedRuntime(value: string): void {
      if (!this.formData.schedulingConfiguration) {
        this.formData.schedulingConfiguration = {};
      }
      this.formData.schedulingConfiguration.minGuaranteedRuntime = value || null;
    },
    updateNumaAwareScheduling(value: boolean): void {
      if (!this.formData.schedulingConfiguration) {
        this.formData.schedulingConfiguration = {};
      }
      this.formData.schedulingConfiguration.numaAwareScheduling = value;
    },
    async save(): Promise<void> {
      const isValidationSuccess: boolean = await (this.$refs.elNodePoolEditCreateForm as HTMLFormElement).validate();
      this.isFormValid = isValidationSuccess;
      if (isValidationSuccess) {
        if (
          !this.isClusterSupportsGpuDevicePlacement &&
          this.formData.schedulingConfiguration?.gpuPodsPlacementStrategy
        ) {
          // On clusters that don't support independent device-level placement,
          // omit the device value so the back-end mirrors node-level instead.
          this.formData.schedulingConfiguration.gpuPodsPlacementStrategy.device = undefined;
        }
        if (!this.isClusterSupportsNumaAwareScheduling && this.formData.schedulingConfiguration) {
          this.formData.schedulingConfiguration.numaAwareScheduling = undefined;
        }
        this.$emit("on-save", this.formData);
      }
    },
    async updateClusterId(clusterId: string): Promise<void> {
      if (clusterId) {
        this.$emit("update-cluster-id", clusterId);
      }
      this.formData.clusterId = clusterId;
    },
  },
  watch: {
    formData: {
      handler() {
        this.isFormValid = true;
      },
      deep: true,
    },
  },
});
</script>
