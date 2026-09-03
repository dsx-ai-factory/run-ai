<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="node-pools-resource-section">
    <div class="q-my-md">
      {{ title }}
      <runai-tooltip :tooltip-text="nodePoolsHeaderInfoText" width="542px" class="node-pools-tooltip-text" />
    </div>
    <runai-draggable-list-order
      :list="list"
      row-key="name"
      main-list-title="Node pool"
      :reset-tooltip-text="resetTooltipText"
      @change="onNodePoolsOrderChanged"
      @reset="onNodePoolsOrderReset"
      :disabled="!editable"
      :disabled-tooltip="disabledTooltip"
    >
      <template #list-item="item">
        <workload-node-pool-list-item
          :node-pool="item.name"
          @remove="onRemove(item)"
          :removeable="list.length > 1"
          :disabled="item.disabled"
        />
      </template>

      <template #footer v-if="editable && moreNodePoolsOptions">
        <runai-dropdown
          color="primary"
          :is-disabled="isAddNodePoolDisabled"
          :disabled-tooltip="addNodePoolDisabledToolTip"
          :options="moreNodePoolsOptions"
          @on-selected="addNodePoolSelected"
        >
          <template #btn>
            <q-icon name="ph-bold ph-plus" size="16px" class="q-mr-sm" />
            <span>NODE POOL</span>
          </template>

          <template #option="option">
            <div>{{ option.name }}</div>
          </template>
        </runai-dropdown>
      </template>
    </runai-draggable-list-order>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// Components
import { RunaiDraggableListOrder, RunaiDropdown } from "@/components/common";
import { WorkloadNodePoolListItem } from "@/components/section/compute-resource-section/workload-node-pool-list-item";
import { RunaiTooltip } from "@/components/common/runai-tooltip";

// models
import { ENodePoolsListOrigin } from "./node-pools-resource-section.models";
import type { IComputeSectionNodePoolsPolicy } from "./node-pools-resource-section.models";

export interface INodePoolsOption {
  id: string;
  name: string;
}

export interface INodePoolsRow {
  name: string;
}

export default defineComponent({
  name: "node-pools-resource-section",
  components: {
    RunaiDraggableListOrder,
    WorkloadNodePoolListItem,
    RunaiDropdown,
    RunaiTooltip,
  },
  emits: ["list-changed", "item-removed"],
  props: {
    selectedNodePools: {
      type: Array as PropType<Array<string>>,
      required: true,
    },
    allNodePools: {
      type: [Array, null] as PropType<Array<string> | null>,
      required: true,
    },
    nodePoolsListOrigin: {
      type: String as PropType<ENodePoolsListOrigin>,
      required: true,
    },
    policy: {
      type: [Object, null] as PropType<IComputeSectionNodePoolsPolicy | null>,
      required: false,
    },
    applyPolicyDefaults: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
    },
    disabled: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    disabledTooltip: {
      type: String as PropType<string>,
      required: false,
      default: "",
    },
  },
  data() {
    return {
      list: [] as Array<INodePoolsRow>,
      originList: [] as Array<INodePoolsRow>,
    };
  },
  created() {
    if (this.policy?.defaults && (this.applyPolicyDefaults || this.policy?.rules?.canEdit === false)) {
      this.list = this.policy?.defaults.map((name: string) => ({ name }));
      this.originList = this.policy?.defaults.map((name: string) => ({ name }));
      this.onNodePoolsOrderChanged(this.list);
    } else {
      this.list = this.selectedNodePools.map((name: string) => ({ name }));
      this.originList = this.selectedNodePools.map((name: string) => ({ name }));
    }
  },
  computed: {
    moreNodePoolsOptions(): Array<INodePoolsOption> | null {
      if (!this.allNodePools) return null;
      const selectedNodePoolsMap = new Set<string>(this.list.map((item: INodePoolsRow) => item.name));
      return this.allNodePools
        .filter((nodePool: string) => !selectedNodePoolsMap.has(nodePool))
        .map((nodePool: string) => ({ id: nodePool, name: nodePool }));
    },
    title(): string {
      return this.editable
        ? "Drag to set the order of priority of the node pools on which the scheduler try to run the workload"
        : "The order of priority of the node pools on which the scheduler try to run the workload";
    },
    resetTooltipText(): string {
      return this.policy?.defaults ? `Reset to policy's default` : `Reset to ${this.nodePoolsListOrigin}'s default`;
    },
    nodePoolsHeaderInfoText(): string {
      const projectInfo = `When a workload is created, the scheduler will try to run it on the first node pool on the list. If the
      node pool doesn't have free resources, the scheduler will move on to the next one until it finds one that's available.
      Drag and drop them to change the order, remove unwanted ones, or reset to the default order defined in the project.`;

      const tempInfo = `When a workload is created, the scheduler will try to run it on the first node pool on the list. If the
      node pool doesn't have free resources, the scheduler will move on to the next one until it finds one that's available.
      Drag and drop them to change the order, remove unwanted ones, or reset to the default order defined in the template.`;

      const lockedAndListFromPolicy = `When a workload is created, the scheduler will try to run it on the first node pool on the list. If the node pool
      doesn't have free resources, the scheduler will move on to the next one until it finds one that's available.
      Based on the policy set by your admin, the list can't be modified.`;

      const listFromPolicy = `When a workload is created, the scheduler will try to run it on the first node pool on the list. If
      the node pool doesn't have free resources, the scheduler will move on to the next one until it finds one that's available.
      Drag and drop them to change the order, remove unwanted ones, or reset to the default order defined in the policy set by your admin.`;

      if (!this.editable && this.policy?.defaults) {
        return lockedAndListFromPolicy;
      } else if (this.policy?.defaults) {
        return listFromPolicy;
      } else if (this.nodePoolsListOrigin === ENodePoolsListOrigin.Template) {
        return tempInfo;
      }

      // default case: this.nodePoolsListOrigin === ENodePoolsListOrigin.Project
      return projectInfo;
    },
    editable(): boolean {
      if (this.policy?.rules?.canEdit !== undefined && this.policy?.rules?.canEdit !== null) {
        return this.policy.rules.canEdit;
      }
      return !this.disabled;
    },
    isAddNodePoolDisabled(): boolean {
      return !this.editable || !this.moreNodePoolsOptions?.length;
    },
    addNodePoolDisabledToolTip(): string {
      if (this.editable) {
        return "No more node pools to add";
      }
      return this.disabledTooltip || "This can't be modified based on the policy your administrator set.";
    },
  },
  methods: {
    addNodePoolSelected(nodePoolSelected: INodePoolsOption): void {
      this.onNodePoolsOrderChanged([...this.list, { name: nodePoolSelected.name }]);
    },
    onNodePoolsOrderChanged(nodePools: Array<INodePoolsRow>): void {
      this.list = nodePools;
      this.$emit(
        "list-changed",
        nodePools.map((nodePool: INodePoolsRow) => nodePool.name),
      );
    },
    onNodePoolsOrderReset(): void {
      this.onNodePoolsOrderChanged(this.originList);
    },
    onRemove(nodePool: INodePoolsRow): void {
      this.onNodePoolsOrderChanged(this.list.filter((item: INodePoolsRow) => nodePool.name !== item.name));
    },
  },
});
</script>
