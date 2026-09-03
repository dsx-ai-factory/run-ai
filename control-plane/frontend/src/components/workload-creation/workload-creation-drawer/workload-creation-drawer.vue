<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section>
    <q-drawer
      class="workload-creation-drawer shadow-4"
      :model-value="isOpen"
      side="right"
      overlay
      :width="drawerWidth"
      behavior="desktop"
      @hide="resetSelectedItems"
    >
      <section class="entire-content full-height">
        <section class="extended-section full-width" v-if="showExtendedSection && displayedItem">
          <drawer-item-compliance-section
            v-if="showComplianceSection"
            :displayed-item="displayedItem"
            :entity="entity"
            :distributed-replica-type="displayReplicaType ?? complyTo.complyToReplicaType"
            :policy-details="policyDetails"
            @close="closeExtendedSection"
          />
          <drawer-item-details-section
            v-else-if="showDetailsSection"
            :displayed-item="displayedItem"
            :entity="entity"
            @close="closeExtendedSection"
          />
        </section>
        <section class="primary-section full-height">
          <workload-creation-drawer-skeleton v-if="loadingData" @close="closeDrawer" />
          <workload-creation-drawer-content
            v-else
            :items="items"
            :entity="entity"
            :displayed-item-id="itemId"
            :loading-error="loadingError"
            :is-multi-select="isMultiSelect"
            :selected-items="selectedItems"
            :created-entity-id="createdEntityId"
            :allowed-creation-ds-options="allowedCreationDsOptions"
            @close="closeDrawer"
            @show-details="openDetailsSection"
            @show-policy="openPolicySection"
            @refresh-data="loadEntities"
            @select="onSelection"
            @multi-select="onMultiSelection"
            @create-new="$emit('create-new', $event)"
          />
        </section>
      </section>
    </q-drawer>
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";
import { WorkloadCreationDrawerContent } from "@/components/workload-creation/workload-creation-drawer-content";
import { WorkloadCreationDrawerEntities } from "@/models/workload-creation-drawer-model";
import { ExtendedSections, type IDrawerListItem } from "@/models/workload-creation-drawer-model";
import { WorkloadCreationDrawerSkeleton } from "../workload-creation-drawer-skeleton";
import type { IPolicyViewDetails } from "@/utils/policy.util/policy-view.util";
import { DrawerItemComplianceSection } from "../drawer-item-compliance-section";
import { DrawerItemDetailsSection } from "../drawer-item-details-section";
import { workloadCreationDrawerEntityMapper } from "../workload-creation-drawer-content/content-mapper";
import type { IAssetsFilter } from "@/models/filter.model";
import { AssetKind } from "@/swagger-models/assets-service-client";
import type { ReplicaType } from "@/swagger-models/workloads-client";
import { useClusterStore } from "@/stores/cluster.store";
import { DATA_SOURCE_ASSET_KINDS_TO_NAMES, type TDataSourceKinds } from "@/models/data-source.model";

const drawerFullWidth = 909;
const drawerInitialWidth = 372;

export default defineComponent({
  name: "workload-creation-drawer",
  components: {
    WorkloadCreationDrawerContent,
    WorkloadCreationDrawerSkeleton,
    DrawerItemComplianceSection,
    DrawerItemDetailsSection,
  },
  emits: ["close", "select", "select-multiple", "create-new"],
  props: {
    isOpen: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    complyTo: {
      type: Object as PropType<IAssetsFilter>,
      required: true,
    },
    entity: {
      type: [String, null] as PropType<WorkloadCreationDrawerEntities | null>,
      required: true,
    },
    policyDetails: {
      type: Object as PropType<IPolicyViewDetails>,
      required: false,
      default: () => ({}),
    },
    displayReplicaType: {
      type: String as PropType<ReplicaType>,
      required: false,
    },
    createdEntityId: {
      type: [String, null] as PropType<string | null>,
      required: false,
      default: "",
    },
    clusterId: {
      type: String as PropType<string>,
      required: false,
    },
    allowedDataSourceOptions: {
      type: Object as PropType<Set<TDataSourceKinds>>,
      required: false,
    },
    applyItemModificationsAfterLoaded: {
      type: Function as PropType<
        (entity: WorkloadCreationDrawerEntities, items: IDrawerListItem[], clusterId?: string) => IDrawerListItem[]
      >,
      required: false,
    },
  },
  data() {
    return {
      clusterStore: useClusterStore(),
      itemId: "" as string,
      displayedSection: "" as ExtendedSections,
      selectedItems: [] as { id: string; name: string; kind: AssetKind }[],
      loadingData: false,
      loadingError: false,
      loadedItems: {
        [WorkloadCreationDrawerEntities.Environment]: [] as IDrawerListItem[],
        [WorkloadCreationDrawerEntities.Compute]: [] as IDrawerListItem[],
        [WorkloadCreationDrawerEntities.DataSource]: [] as IDrawerListItem[],
        [WorkloadCreationDrawerEntities.ModelStorage]: [] as IDrawerListItem[],
      } as Record<WorkloadCreationDrawerEntities, IDrawerListItem[]>,
    };
  },
  mounted(): void {
    document.addEventListener("keydown", this.handleEscapeKey);
  },
  computed: {
    items(): IDrawerListItem[] {
      if (!this.entity) {
        return [];
      }
      return this.loadedItems[this.entity] || [];
    },
    drawerWidth(): number {
      return this.showExtendedSection ? drawerFullWidth : drawerInitialWidth;
    },
    showExtendedSection(): boolean {
      return !!this.itemId && !!this.displayedSection && !!this.displayedItem;
    },
    displayedItem(): IDrawerListItem | undefined {
      return this.items.find((item) => item.id === this.itemId);
    },
    showComplianceSection(): boolean {
      return this.displayedSection === ExtendedSections.POLICY;
    },
    showDetailsSection(): boolean {
      return this.displayedSection === ExtendedSections.DETAILS;
    },
    isMultiSelect(): boolean {
      return this.entity === WorkloadCreationDrawerEntities.DataSource;
    },
    allowedCreationDsOptions(): Array<string> | undefined {
      let allowedCreationOptions: string[] | undefined;
      if (this.allowedDataSourceOptions) {
        allowedCreationOptions = [...this.allowedDataSourceOptions].map((option) => {
          return DATA_SOURCE_ASSET_KINDS_TO_NAMES[option];
        });
      }
      return allowedCreationOptions;
    },
  },
  methods: {
    closeDrawer(): void {
      this.closeExtendedSection();
      this.$emit("close");
    },
    handleEscapeKey(event: KeyboardEvent): void {
      if (event.key === "Escape" && this.isOpen) {
        this.closeDrawer();
      }
    },
    openPolicySection(itemId: string): void {
      if (this.itemId === itemId && this.displayedSection === ExtendedSections.POLICY) {
        this.closeExtendedSection();
        return;
      }
      this.itemId = itemId;
      this.displayedSection = ExtendedSections.POLICY;
    },
    openDetailsSection(itemId: string): void {
      if (this.itemId === itemId && this.displayedSection === ExtendedSections.DETAILS) {
        this.closeExtendedSection();
        return;
      }
      this.itemId = itemId;
      this.displayedSection = ExtendedSections.DETAILS;
    },
    closeExtendedSection(): void {
      this.itemId = "";
      this.displayedSection = "" as ExtendedSections;
    },
    async loadEntities(forceLoad = false) {
      if (!this.entity || (this.loadedItems[this.entity].length && !forceLoad)) {
        return;
      }
      try {
        if (!forceLoad) {
          this.loadingData = true;
          this.loadingError = false;
        }

        let loadedItems = await workloadCreationDrawerEntityMapper[this.entity].loadFunction(this.complyTo);
        if (this.applyItemModificationsAfterLoaded) {
          loadedItems = this.applyItemModificationsAfterLoaded(this.entity, loadedItems, this.clusterId);
        }

        this.loadedItems[this.entity] = loadedItems;

        const createdItem = this.loadedItems[this.entity].find((item) => item.id === this.createdEntityId);
        if (
          createdItem?.kind === AssetKind.Pvc &&
          !createdItem?.policy?.compliance &&
          createdItem.policy?.reason?.some((reason) => !reason.field)
        ) {
          setTimeout(() => {
            this.loadEntities(true);
          }, 2000);
        }
      } catch (e) {
        this.loadingError = true;
        console.error(e);
      } finally {
        this.loadingData = false;
      }
    },
    onSelection(itemIdAndName: { id: string; name: string; kind: AssetKind }): void {
      if (!this.isMultiSelect) {
        this.$emit("select", itemIdAndName);
        return;
      }

      const existingItemIndex = this.selectedItems.findIndex((item) => item.id === itemIdAndName.id);
      if (existingItemIndex !== -1) {
        this.selectedItems = this.selectedItems.filter((item) => item.id !== itemIdAndName.id);
        return;
      }
      this.selectedItems.push(itemIdAndName);
    },
    onMultiSelection(): void {
      this.$emit("select-multiple", this.selectedItems);
    },
    resetSelectedItems(): void {
      this.selectedItems = [];
    },
    resetLoadedItems(): void {
      this.loadedItems = {
        [WorkloadCreationDrawerEntities.Environment]: [] as IDrawerListItem[],
        [WorkloadCreationDrawerEntities.Compute]: [] as IDrawerListItem[],
        [WorkloadCreationDrawerEntities.DataSource]: [] as IDrawerListItem[],
        [WorkloadCreationDrawerEntities.ModelStorage]: [] as IDrawerListItem[],
      };
    },
  },
  watch: {
    entity: {
      handler(newVal: WorkloadCreationDrawerEntities, oldVal: WorkloadCreationDrawerEntities): void {
        if (newVal !== oldVal) {
          this.loadingError = false;
          this.resetSelectedItems();
          this.loadEntities();
          this.closeExtendedSection();
        }
      },
      immediate: true,
    },
    complyTo: {
      handler(newVal: IAssetsFilter, oldVal: IAssetsFilter): void {
        // if the scope has changed, reset the selected items and loaded items
        if (
          newVal.clusterUuid !== oldVal.clusterUuid ||
          newVal.departmentId !== oldVal.departmentId ||
          newVal.projectId !== oldVal.projectId
        ) {
          this.resetSelectedItems();
          this.resetLoadedItems();
        }
      },
    },
  },
  unmounted(): void {
    document.removeEventListener("keydown", this.handleEscapeKey);
  },
});
</script>
<style lang="scss" scoped>
.workload-creation-drawer {
  .entire-content {
    display: flex;
    flex-direction: row;
    flex-wrap: nowrap;
    .extended-section {
      display: flex;
      flex-direction: column;
      border-right: 1px $black-10 solid;
    }
    .primary-section {
      min-width: 372px;
      max-width: 372px;
    }
  }
}
</style>
