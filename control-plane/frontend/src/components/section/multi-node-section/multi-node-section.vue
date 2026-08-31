<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item
    class="multi-node-section"
    aid="multi-node-section"
    label="Workload architecture"
    :section-invalid="sectionInvalid"
    :default-opened="isSelectable"
    :disable-opening="!isSelectable"
    :disable-closing="isSelectable"
    hide-expend-icon
  >
    <template #subheader>
      <div class="row items-center justify-between q-pr-sm">
        <span>{{ summary }}</span>
        <slot name="custom-button"></slot>
      </div>
    </template>

    <distributed-selection
      :distributed-model="distributedSelectionModel"
      @update-distributed-model="updateDistributedSelectionModel"
      :form-type="EDistributedFormType.TRAINING"
      :cluster-id="clusterId"
    />

    <section v-if="showNoMasterOptions">
      <distributed-replica-configuration :model-value="noMasterModel" @update:model-value="updateNoMasterModel" />
    </section>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// Components
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";
import { DistributedSelection, type IDistributedSelectionModel } from "./distributed-selection";
import { DistributedReplicaConfiguration } from "./distributed-replica-configuration";

// Models
import { DistributedFramework } from "@/swagger-models/assets-service-client";
import type { IUIDistributed } from "@/models/workload.model";
import { EDistributedFormType } from "./distributed-selection/distributed-selection.vue";

export default defineComponent({
  name: "multi-node-section",
  components: {
    RunaiExpansionItem,
    DistributedSelection,
    DistributedReplicaConfiguration,
  },
  emits: ["update-multi-node-section-model"],
  props: {
    multiNodeSectionModel: {
      type: [Object, null] as PropType<IUIDistributed | null>,
      required: false,
    },
    isSelectable: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    clusterId: {
      type: String as PropType<string>,
      required: true,
    },
  },
  data() {
    return {
      EDistributedFormType: EDistributedFormType,
    };
  },
  computed: {
    distributedSelectionModel(): IDistributedSelectionModel {
      return {
        isDistributed: !!this.multiNodeSectionModel,
        distFramework: this.multiNodeSectionModel?.distFramework || null,
      };
    },
    sectionInvalid(): boolean {
      if (!this.multiNodeSectionModel) return false;
      return this.multiNodeSectionModel && !this.multiNodeSectionModel.distFramework;
    },
    summary(): string {
      if (!this.multiNodeSectionModel?.distFramework) return "Standard";
      return `Distributed (${this.multiNodeSectionModel.distFramework})`;
    },
    showNoMasterOptions(): boolean {
      return (
        !!this.multiNodeSectionModel &&
        (this.multiNodeSectionModel.distFramework === DistributedFramework.PyTorch ||
          this.multiNodeSectionModel.distFramework === DistributedFramework.Tf)
      );
    },
    noMasterModel(): boolean {
      return this.multiNodeSectionModel?.noMaster || false;
    },
  },
  methods: {
    updateNoMasterModel(noMaster: boolean): void {
      if (!this.multiNodeSectionModel) return;
      const multiNodeSectionModel: IUIDistributed = {
        ...this.multiNodeSectionModel,
        noMaster,
      };
      this.updateModel(multiNodeSectionModel);
    },
    updateModel(multiNodeSectionModel: IUIDistributed | null): void {
      this.$emit("update-multi-node-section-model", multiNodeSectionModel);
    },
    updateDistributedSelectionModel(distributedSelection: IDistributedSelectionModel): void {
      const { isDistributed, distFramework } = distributedSelection;
      if (!isDistributed) {
        this.updateModel(null);
      } else {
        const newDistributedModel: IUIDistributed = {
          distFramework: distFramework || undefined,
          numWorkers: 1,
        };

        if (distFramework === DistributedFramework.Jax) newDistributedModel.noMaster = true;

        this.updateModel(newDistributedModel);
      }
    },
  },
});
</script>
