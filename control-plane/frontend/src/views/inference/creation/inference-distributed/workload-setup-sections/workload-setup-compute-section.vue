<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item
    class="compute-resources-section"
    label="Compute resources"
    default-opened
    aid="compute-resources-section"
  >
    <template #subheader>
      <span>{{ summary }}</span>
    </template>

    <div class="compute-section-content q-pb-sm q-mb-sm">
      <div class="q-mb-md">Set the number of replicas for your workload</div>

      <div class="row no-wrap">
        <div class="left-panel">
          <div class="col-3">
            <div class="row">
              <policy-number-input
                type="number"
                label="Replicas"
                stack-label
                :model-value="model.replicas"
                @update:model-value="onReplicasChanged"
                :min-value="1"
                :disable="false"
                :policy-rules="config.replicasPolicyRules"
                custom-ref="replicasInput"
                aid="replicas-input"
              />
            </div>
          </div>
        </div>
        <div class="right-panel">
          <distributed-gpu-resources-summary
            :leader-compute="model.leaderCompute"
            :worker-compute="model.workerCompute"
            :replicas="model.replicas"
            :num-workers="model.workers"
          />
        </div>
      </div>
    </div>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// components
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { PolicyNumberInput } from "@/components/common/policy-number-input";
import DistributedGpuResourcesSummary from "./distributed-gpu-resources-summary.vue";

// models
import type { IComputeSectionModel, IComputeSectionConfig } from "./workload-setup-sections.types";

export enum EEmitName {
  UpdateModel = "update:model",
}

export default defineComponent({
  name: "workload-setup-compute-section",
  components: {
    RunaiExpansionItem,
    PolicyNumberInput,
    DistributedGpuResourcesSummary,
  },
  emits: {
    [EEmitName.UpdateModel]: (model: IComputeSectionModel) => !!model,
  },
  props: {
    model: {
      type: Object as PropType<IComputeSectionModel>,
      required: true,
    },
    config: {
      type: Object as PropType<IComputeSectionConfig>,
      default: () => ({}),
    },
  },
  computed: {
    summary(): string {
      return `Replicas: ${this.model.replicas}`;
    },
  },
  methods: {
    onReplicasChanged(replicas: number | string): void {
      const value = Number(replicas);
      const clamped = Number.isNaN(value) || value < 1 ? 1 : value;
      this.$emit(EEmitName.UpdateModel, { ...this.model, replicas: clamped });
    },
  },
});
</script>

<style lang="scss" scoped>
.compute-section-content {
  .left-panel {
    flex-basis: calc(100% - $distributed-gpu-summary-width);
  }

  .right-panel {
    flex-basis: $distributed-gpu-summary-width;
  }
}
</style>
