<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-base-dialog
    model-value
    @hide="onHide"
    @close="onHide"
    :no-padding="true"
    size="width-sm"
    max-height="70vh"
    class="runai-org-tree-with-popover"
  >
    <template #header> {{ title }} </template>
    <template #body>
      <div class="row">
        <q-input
          :clearable="!!searchTerm"
          clear-icon="ph-bold ph-x"
          class="col-12"
          v-model="searchTerm"
          :placeholder="placeholderSearch"
          aid="org-tree-search-input"
          :input-style="{ paddingInline: '16px' }"
        >
          <template v-slot:append>
            <q-icon name="ph-bold ph-magnifying-glass" size="20px" class="q-ml-sm q-mr-lg q-pr-xs" />
          </template>
        </q-input>
      </div>
      <q-separator />
      <div class="popover-body">
        <runai-org-tree
          :selected="selected"
          :readonly="readonly"
          @update:selected="onUpdatedSelected"
          :search-term="searchTerm"
          :no-results-label="noResultsLabel"
          :allowed-scopes="allowedScopes"
          :scopes-supported-versions="scopesSupportedVersions"
          :disable-tenant="disableTenant"
          :disable-cluster="disableCluster"
          :forced-disabled-scopes="forcedDisabledScopes"
          :hidden-scope-types="hiddenScopeTypes"
          :tick-strategy="tickStrategy"
          :selected-cluster-id="selectedClusterId"
          :multiple="multiple"
        />
      </div>
    </template>
    <template #footer>
      <section class="row items-center q-py-sm" style="width: 100%">
        <span v-if="showSelectionError" class="text-negative q-ml-sm" aid="org-tree-selection-error">
          Select a scope
        </span>
        <div class="q-ml-auto">
          <q-btn
            aid="org-tree-modal-cancel-btn"
            flat
            color="primary"
            label="Cancel"
            class="q-mr-sm"
            @click="onHide"
            v-if="!readonly"
          />
          <q-btn aid="org-tree-modal-select-btn" color="primary" :label="selectButtonCaption" @click="onSelect" />
        </div>
      </section>
    </template>
  </runai-base-dialog>
</template>

<script lang="ts">
// stores
import { defineComponent, type PropType } from "vue";

// models
import type { IOrgTreeNodeId } from "@/models/org-tree.model";

// cmps
import { RunaiBaseDialog } from "../../runai-base-dialog";
import { RunaiOrgTree } from "@/components/common/runai-org-tree";
import type { PermittedScopes, ScopeType } from "@/swagger-models/authorization-client";
import type { TTreeDisabledScopes, TTreeSupportedVersions } from "@/models/global.model";

// services

export default defineComponent({
  components: { RunaiOrgTree, RunaiBaseDialog },
  emits: {
    "update:selected": (selected: IOrgTreeNodeId) => selected,
    "update:selected-nodes": (selected: IOrgTreeNodeId[]) => selected,
    hide: () => true,
  },
  props: {
    title: {
      type: String as PropType<string>,
      default: "Scope",
    },
    selected: {
      type: Array as PropType<IOrgTreeNodeId[]>,
      default: {} as IOrgTreeNodeId,
    },
    readonly: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    placeholderSearch: {
      type: String as PropType<string>,
      default: `Search for a cluster, department, or project`,
    },
    noResultsLabel: {
      type: String as PropType<string>,
      default: "No cluster, department, or project match your filters",
    },
    cover: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    allowedScopes: {
      type: Object as PropType<PermittedScopes>,
    },
    scopesSupportedVersions: {
      type: Object as PropType<TTreeSupportedVersions>,
      default: () => ({}),
    },
    disableTenant: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    disableCluster: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    forcedDisabledScopes: {
      type: Object as PropType<TTreeDisabledScopes>,
      default: () => ({}),
    },
    hiddenScopeTypes: {
      type: Object as PropType<Record<ScopeType, boolean>>,
      default: () => ({}),
    },
    tickStrategy: {
      type: String as PropType<"leaf" | "leaf-filtered" | "none" | "strict" | undefined>,
      default: "strict",
    },
    selectedClusterId: {
      type: String as PropType<string>,
      required: false,
    },
    multiple: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    allowEmptySelection: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      selectedNodes: [] as IOrgTreeNodeId[],
      searchTerm: "" as string,
      showSelectionError: false as boolean,
    };
  },
  computed: {
    selectButtonCaption(): string {
      return this.readonly ? "Close" : "Apply";
    },
    isSelectionEmpty(): boolean {
      return !this.allowEmptySelection && !this.selectedNodes.length;
    },
  },
  methods: {
    onSelect(): void {
      if (!this.readonly && this.isSelectionEmpty) {
        this.showSelectionError = true;
        return;
      }
      if (this.multiple) {
        this.$emit("update:selected-nodes", this.selectedNodes);
      } else {
        this.$emit("update:selected", this.selectedNodes[0]);
      }
      this.onHide();
    },
    onHide(): void {
      this.searchTerm = "";
      this.selectedNodes = [] as IOrgTreeNodeId[];
      this.showSelectionError = false;
      this.$emit("hide");
    },
    onUpdatedSelected(selected: IOrgTreeNodeId[]): void {
      this.selectedNodes = selected;
      if (selected.length) this.showSelectionError = false;
    },
  },
  watch: {
    selected: {
      handler(newValue: IOrgTreeNodeId[]): void {
        this.selectedNodes = newValue;
      },
      immediate: true,
      deep: true,
    },
  },
});
</script>

<style lang="scss">
.popover-body {
  min-height: 72px;
  max-height: calc(70vh - 184px);
}
</style>
