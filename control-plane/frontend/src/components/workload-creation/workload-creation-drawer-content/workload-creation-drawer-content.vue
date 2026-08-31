<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <overlay-drawer-section-wrapper @close="$emit('close')">
    <template #header-title>
      <div>{{ title }}</div>
    </template>
    <template #header-main>
      <q-item class="row items-center q-pa-md full-width">
        <q-input
          dense
          v-model="filter"
          aid="search-input-drawer"
          class="search-input q-mr-md"
          placeholder="Search"
          :clearable="!!filter"
          clear-icon="ph-bold ph-x"
        >
          <template v-slot:prepend>
            <q-icon name="ph-bold ph-magnifying-glass" size="20px" />
          </template>
        </q-input>

        <runai-dropdown-btn-select
          :options="sortOptions"
          :selected-value="sortBy"
          :tooltip="sortBtnTooltip"
          dropdown-icon="ph-bold ph-sort-descending"
          @on-select="onSortSelected"
          class="sort-btn square-btn"
          custom-classes="btn-icon-md"
        />
        <q-tooltip target=".sort-btn">{{ sortBtnTooltip }}</q-tooltip>

        <runai-dropdown
          v-if="creationDropdownOptions.length"
          class="square-btn new-btn-dropdown add-new-btn"
          custom-classes="items-center full-width full-height btn-icon-md"
          icon="ph-bold ph-plus"
          color="black-54"
          :fit="false"
          :options="creationDropdownOptions"
          @on-selected="$emit('create-new', $event?.value)"
          aid="add-new-dropdown-btn"
        />
        <runai-btn
          v-else
          class="square-btn add-new-btn btn-icon-md"
          icon="ph-bold ph-plus"
          flat
          @click="$emit('create-new')"
          color="black-54"
          aid="add-new-btn"
        />
        <q-tooltip target=".add-new-btn" anchor="bottom left" :offset="[26, 14]"> New {{ entityName }} </q-tooltip>
      </q-item>
    </template>

    <template #main>
      <q-list class="items-list">
        <q-item-section aid="loading-error" class="text-center q-pa-md" v-if="loadingError">
          <div>For some reason {{ entityName }}s can't be displayed right now.</div>
          <div><a href="#" @click.prevent="$emit('refresh-data')">Try reloading</a> or contact your admin.</div>
        </q-item-section>
        <section class="list-no-data" v-else-if="!filteredItems.length">
          <q-item-section class="text-center q-pa-md">
            <div class="text-black-54">
              <q-icon v-if="filter" name="ph-bold ph-funnel" size="44px" class="q-mb-md" />
              <q-icon v-else-if="emptyIcon" :name="emptyIcon" size="44px" class="q-mb-md" />
              <q-item-label>{{ emptyText }}</q-item-label>
            </div>
          </q-item-section>
        </section>
        <workload-drawer-item
          class="mb-1px"
          v-for="item in filteredItems"
          :key="item.name"
          :item="item"
          :focused="item.id === displayedItemId"
          :selected="isSelected(item.id)"
          @details-clicked="onDetailsClicked"
          @policy-clicked="onPolicyClicked"
          @item-clicked="onItemClicked"
        />
      </q-list>
    </template>
    <template #footer v-if="isMultiSelect">
      <q-item :clickable="false" class="multi-select-footer q-pa-md">
        <q-item-section>
          <q-item-label class="selected-counter" :class="{ 'failed-clr': noSelectionWarning }">
            {{ selectedItems.length }} Selected
          </q-item-label>
        </q-item-section>
        <q-item-section side>
          <runai-btn aid="multi-select-btn" color="primary" label="load" @click="onMultiSelect" />
        </q-item-section>
      </q-item>
    </template>
  </overlay-drawer-section-wrapper>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";
import type { WorkloadCreationDrawerEntities } from "@/models/workload-creation-drawer-model";
import { type IDrawerListItem } from "@/models/workload-creation-drawer-model";
import { WorkloadDrawerItem } from "@/components/workload-creation/workload-drawer-item";
import { RunaiDropdown } from "@/components/common";
import { workloadCreationDrawerEntityMapper } from "./content-mapper";
import { OverlayDrawerSectionWrapper } from "@/components/workload-creation/overlay-drawer-section-wrapper";
import { RunaiDropdownBtnSelect } from "@/components/common/runai-dropdown-btn-select";
import { dateUtil } from "@/utils/date.util";
import type { ISelectOption } from "@/models/global.model";
import type { AssetKind } from "@/swagger-models/assets-service-client";
import { RunaiBtn } from "@/components/common/runai-btn";
import { usePermissionStore } from "@/stores/permissions.store";
import type { TDataSourceKinds } from "@/models/data-source.model";
import { DS_KIND_TO_RESOURCE_TYPE } from "@/models/data-source.model";
import { Action } from "@/swagger-models/authorization-client";

export type TSortBy = "name" | "lastUsed" | "lastCreated" | "default";
export type TSortFunction = (a: IDrawerListItem, b: IDrawerListItem) => number;

const sortOptions: ISelectOption[] = [
  { label: "Last used", value: "lastUsed" },
  { label: "Last created", value: "lastCreated" },
  { label: "Name", value: "name" },
];

const sortFunctionsMap: Record<TSortBy, TSortFunction> = {
  name: (cardA: IDrawerListItem, cardB: IDrawerListItem) => {
    if (!cardB.name || !cardA.name) return 0;
    return cardA.name.localeCompare(cardB.name);
  },
  lastUsed: (cardA: IDrawerListItem, cardB: IDrawerListItem) =>
    dateUtil.dateToTimestamp(cardB.lastUsed) - dateUtil.dateToTimestamp(cardA.lastUsed),
  lastCreated: (cardA: IDrawerListItem, cardB: IDrawerListItem) =>
    dateUtil.dateToTimestamp(cardB.createdAt) - dateUtil.dateToTimestamp(cardA.createdAt),
  default: () => 0,
};

export default defineComponent({
  name: "workload-creation-drawer-content",
  components: {
    WorkloadDrawerItem,
    OverlayDrawerSectionWrapper,
    RunaiDropdownBtnSelect,
    RunaiDropdown,
    RunaiBtn,
  },
  emits: ["close", "show-details", "show-policy", "create-new", "refresh-data", "select", "multi-select"],
  props: {
    entity: {
      type: [String, null] as PropType<WorkloadCreationDrawerEntities | null>,
      required: true,
    },
    items: {
      type: Array as PropType<IDrawerListItem[]>,
      required: false,
      default: () => [],
    },
    displayedItemId: {
      type: String as PropType<string>,
      required: false,
      default: "",
    },
    loadingError: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    isMultiSelect: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    selectedItems: {
      type: Array as PropType<{ id: string; name: string }[]>,
      default: () => [],
    },
    createdEntityId: {
      type: [String, null] as PropType<string | null>,
      required: false,
      default: "",
    },
    allowedCreationDsOptions: {
      type: Array as PropType<Array<string>>,
      required: false,
    },
  },
  data() {
    return {
      filter: "" as string,
      sortOptions,
      sortBy: "lastUsed" as TSortBy,
      noSelectionWarning: false,
      permissionStore: usePermissionStore(),
    };
  },
  computed: {
    filteredItems(): IDrawerListItem[] {
      const sortedAndFiltered = this.items
        .filter((item) => item.name.includes(this.filter))
        .sort(sortFunctionsMap[this.sortBy]);
      if (!this.createdEntityId || this.filter) return sortedAndFiltered;
      const createdEntity = sortedAndFiltered.find((item) => item.id === this.createdEntityId);
      return createdEntity
        ? [createdEntity, ...sortedAndFiltered.filter((item) => item.id !== this.createdEntityId)]
        : sortedAndFiltered;
    },
    title(): string {
      if (!this.entity) return "";
      return (
        workloadCreationDrawerEntityMapper[this.entity]?.alternativeTitle?.() ||
        workloadCreationDrawerEntityMapper[this.entity]?.title ||
        ""
      );
    },
    entityName(): string {
      if (!this.entity) return "";
      return workloadCreationDrawerEntityMapper[this.entity]?.entityLabel || "";
    },
    emptyIcon(): string {
      if (!this.entity) return "";
      return workloadCreationDrawerEntityMapper[this.entity]?.defaultIcon || "";
    },
    emptyText(): string {
      return this.filter
        ? `No ${this.entityName}s match your filter`
        : `Looks like you don't have any ${this.entityName}s yet...`;
    },
    sortBtnTooltip(): string {
      const label = this.sortOptions.find((option) => option.value === this.sortBy)?.label;
      return `Sorted by ${label?.toLocaleLowerCase() || ""}`;
    },
    creationDropdownOptions(): Array<ISelectOption> {
      if (!this.entity) return [];

      const creationOptions = workloadCreationDrawerEntityMapper[this.entity]?.creationOptions || [];

      const optionsWithPermissions = creationOptions.map((option) => {
        const resourceType = DS_KIND_TO_RESOURCE_TYPE[option.value as TDataSourceKinds];
        option.disable = !this.permissionStore.hasPermission(resourceType, Action.Create);
        return option;
      });

      if (!this.allowedCreationDsOptions) return optionsWithPermissions;

      const set = new Set(this.allowedCreationDsOptions);
      return optionsWithPermissions.filter((option: ISelectOption) => set.has(option.label));
    },
  },
  methods: {
    onItemClicked(itemIdAndName: { id: string; name: string; kind: AssetKind }): void {
      if (this.noSelectionWarning) this.noSelectionWarning = false;
      this.$emit("select", itemIdAndName);
    },
    onDetailsClicked(id: string): void {
      this.$emit("show-details", id);
    },
    onPolicyClicked(id: string): void {
      this.$emit("show-policy", id);
    },
    onSortSelected(sortBy: TSortBy): void {
      this.sortBy = sortBy;
    },
    onMultiSelect(): void {
      if (!this.selectedItems.length) {
        this.noSelectionWarning = true;
        return;
      }
      this.$emit("multi-select");
    },
    isSelected(itemId: string): boolean {
      return !!this.selectedItems?.some((item) => item.id === itemId);
    },
  },
});
</script>
<style lang="scss" scoped>
.search-input {
  flex: 1;

  :deep(.q-field__focusable-action) {
    font-size: 18px;
  }
}
.square-btn {
  height: 36px;
  width: 36px;
}

.multi-select-footer {
  border-top: 1px $black-10 solid;
  height: 69px;
  .selected-counter {
    transition: color 0.3s;
  }
}
.mb-1px {
  margin-bottom: 1px;
}
</style>
