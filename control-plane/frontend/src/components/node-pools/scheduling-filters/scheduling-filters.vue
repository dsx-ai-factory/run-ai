<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <runai-advanced-filters
    :filter-by="filterBy"
    :initial-filters-model="filterBy.advancedFilters ?? []"
    :filter-options="filterOptions"
    hide-cluster-filter
    menu-width="208px"
    @update-filter-by="emit('update-filter-by', $event)"
  >
    <template #editor="{ filter, save, close }">
      <runai-free-text-filter
        v-if="filter.type === EColumnFilterType.FreeText"
        :key="filter.name"
        :filter="filter as IFreeTextFilterModel"
        width="282px"
        operator-menu-width="208px"
        :format-operator-label="capitalize"
        max-height="512px"
        class="node-pool-scheduling-filter-popup"
        @save="save"
        @close="close"
      />
      <runai-enum-based-filter
        v-else
        :key="filter.name"
        :filter="filter as IEnumBasedFilterModel"
        width="208px"
        max-height="512px"
        class="node-pool-scheduling-filter-popup"
        @save="save"
        @close="close"
      />
    </template>
  </runai-advanced-filters>
</template>
<script setup lang="ts">
import { computed } from "vue";
import { RunaiAdvancedFilters } from "@/components/common/runai-advanced-filters/index";
import { RunaiFreeTextFilter } from "@/components/common/runai-advanced-filters/runai-free-text-filter/index";
import { RunaiEnumBasedFilter } from "@/components/common/runai-advanced-filters/runai-enum-based-filter/index";
import {
  EColumnFilterType,
  type IAdvancedFilterOption,
  type IFilterBy,
  type IFreeTextFilterModel,
  type IEnumBasedFilterModel,
} from "@/models/filter.model";
import { ENTITY_PRIORITY_SELECT_OPTIONS } from "@/models/org-unit.model";

const props = defineProps<{ filterBy: IFilterBy; ranked: boolean }>();
const emit = defineEmits<{ "update-filter-by": [IFilterBy] }>();
const filterOptions = computed<IAdvancedFilterOption[]>(() => [
  { name: "department", field: "name", label: "Department", type: EColumnFilterType.FreeText },
  ...(props.ranked
    ? [
        {
          name: "rank",
          field: "rank",
          label: "Rank",
          type: EColumnFilterType.EnumBased,
          selectOptions: ENTITY_PRIORITY_SELECT_OPTIONS,
        },
      ]
    : []),
]);
function capitalize(label: string): string {
  return label.charAt(0).toUpperCase() + label.slice(1);
}
</script>
<style scoped lang="scss">
.advanced-filters {
  display: flex;
  flex: 1;
  align-items: center;
  min-width: 0;
  overflow: hidden;
}
</style>

<style lang="scss">
.node-pool-scheduling-filter-popup:has(.checkbox-container[aria-label="Rank"]) {
  padding: 16px 10px;
  border-radius: 3px;
  box-shadow: var(--shadow-s);

  > .q-card__section {
    padding: 0;
  }
  > .q-card__section:first-child {
    height: 32px;
    margin-bottom: 16px;

    > .text-weight-medium {
      font-size: 14px;
      font-weight: 500;
      line-height: 20px;
    }
    > .q-btn {
      width: 32px;
      height: 32px;
      min-height: 32px;
      padding: 0;

      .q-icon {
        font-size: 20px;
      }
    }
  }
  .checkbox-container > span {
    margin-bottom: 16px;
    font-size: 12px;
    font-weight: 500;
    line-height: 18px;
  }
  .checkbox-container > .q-item {
    height: 32px;
    min-height: 32px;
    padding-right: 10px;

    &:not(:last-child) {
      margin-bottom: 7px;
    }
  }
  .q-checkbox {
    width: 100%;
    height: 32px;
    min-height: 32px;
  }
  .q-checkbox__inner {
    font-size: 32px;
  }
  .q-checkbox__label {
    flex: 1;
    min-width: 0;
    padding-left: 8px;
    font-size: 14px;
    font-weight: 400;
    line-height: 19.5px;
  }
  > .q-card__section:last-child {
    height: 38px;
    margin-top: 16px;

    > .q-btn {
      width: 62px;
      height: 38px;
      min-height: 38px;
      padding: 0 10px;
      font-size: 14px;
      font-weight: 500;
      line-height: 20px;
    }
  }
}
</style>
