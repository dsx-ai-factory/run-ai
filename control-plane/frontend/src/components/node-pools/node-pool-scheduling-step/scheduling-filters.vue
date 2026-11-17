<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->
<template>
  <div class="advanced-filters">
    <runai-advanced-filter-summary :filters="appliedFilters" @select="edit" @remove="remove" />
    <q-btn class="add-filter-btn q-ml-xs" flat no-caps label="Add filter">
      <q-menu transition-show="scale" transition-hide="scale" :style="{ width: '208px' }">
        <q-list>
          <q-item v-for="name in availableFilters" :key="name" clickable v-close-popup @click="add(name)">
            <q-item-section>{{ name === "department" ? "Department" : "Rank" }}</q-item-section>
          </q-item>
        </q-list>
      </q-menu>
      <runai-page-filter-wrapper
        v-if="editor"
        display
        :title="editor.label"
        :max-width="editor.name === 'department' ? '282px' : '208px'"
        :style="{ width: editor.name === 'department' ? '282px' : '208px' }"
        max-height="512px"
        class="node-pool-scheduling-filter-popup"
        @save="apply"
        @close="editor = undefined"
      >
        <template #body>
          <template v-if="editor.type === EColumnFilterType.FreeText">
            <q-select
              aid="department-filter-operator"
              v-model="editor.value"
              outlined
              emit-value
              map-options
              :options="operators"
              menu-shrink
              :popup-content-style="{ width: '208px' }"
              class="q-mb-lg"
            />
            <p class="text-black-54 filter-input-label">The following value:</p>
            <q-input aid="department-filter-term" outlined v-model="editor.term" label="Department" />
          </template>
          <div v-else class="checkbox-container" role="group" aria-label="Rank">
            <span class="text-black-54 q-mb-md">Matches any:</span>
            <q-item v-for="option in editor.selectOptions" :key="String(option.value)" dense role="presentation">
              <q-checkbox class="no-padding" v-model="editor.selectedValues" :val="option.value" :label="option.label" />
            </q-item>
          </div>
        </template>
      </runai-page-filter-wrapper>
    </q-btn>
  </div>
</template>
<script setup lang="ts">
import { computed, ref } from "vue";
import { RunaiAdvancedFilterSummary } from "@/components/common/runai-advanced-filters/runai-advanced-filter-summary";
import { RunaiPageFilterWrapper } from "@/components/common/runai-advanced-filters/runai-page-filter-wrapper";
import {
  EColumnFilterType,
  EFilterOperator,
  type IAdvancedFilterModel,
  type IFilterBy,
  type IFreeTextFilterModel,
  type IEnumBasedFilterModel,
} from "@/models/filter.model";
import { ENTITY_PRIORITY_SELECT_OPTIONS } from "@/models/org-unit.model";

type SchedulingFilter =
  | (IFreeTextFilterModel & { type: EColumnFilterType.FreeText })
  | (IEnumBasedFilterModel & { type: EColumnFilterType.EnumBased });
const props = defineProps<{ filterBy: IFilterBy; ranked: boolean }>();
const emit = defineEmits<{ "update-filter-by": [IFilterBy] }>();
const editor = ref<SchedulingFilter>();
const operators = [
  { label: "Contains", value: EFilterOperator.Contains },
  { label: "Does not contain", value: EFilterOperator.DoesNotContain },
  { label: "Equals", value: EFilterOperator.Equals },
  { label: "Does not equal", value: EFilterOperator.NotEquals },
];
const appliedFilters = computed(() => props.filterBy.advancedFilters ?? []);
const availableFilters = computed(() =>
  (props.ranked ? ["department", "rank"] : ["department"]).filter(
    (name) => !appliedFilters.value.some((filter) => filter.name === name),
  ),
);
function edit(filter: IAdvancedFilterModel) {
  if (filter.type === EColumnFilterType.FreeText)
    editor.value = { ...(filter as IFreeTextFilterModel), type: EColumnFilterType.FreeText };
  else if (filter.type === EColumnFilterType.EnumBased) {
    const rank = filter as IEnumBasedFilterModel;
    editor.value = { ...rank, type: EColumnFilterType.EnumBased, selectedValues: [...rank.selectedValues] };
  }
}
function add(name: string) {
  editor.value =
    name === "department"
      ? { name, label: "Department", type: EColumnFilterType.FreeText, term: "", value: EFilterOperator.Contains }
      : {
          name,
          label: "Rank",
          type: EColumnFilterType.EnumBased,
          selectOptions: ENTITY_PRIORITY_SELECT_OPTIONS,
          selectedValues: [],
        };
}
function remove(name: string) {
  emit("update-filter-by", {
    ...props.filterBy,
    advancedFilters: appliedFilters.value.filter((filter) => filter.name !== name),
  });
}
function apply() {
  const filter = editor.value;
  if (!filter) return;
  if (filter.type !== EColumnFilterType.FreeText || filter.term) {
    const saved =
      filter.type === EColumnFilterType.EnumBased
        ? { ...filter, selectedValues: [...filter.selectedValues] }
        : { ...filter };
    emit("update-filter-by", {
      ...props.filterBy,
      advancedFilters: [...appliedFilters.value.filter((item) => item.name !== filter.name), saved],
    });
  }
  editor.value = undefined;
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
.filter-input-label {
  margin-bottom: 5px;
}
.checkbox-container {
  display: flex;
  flex-direction: column;
  .q-item {
    padding: 0;
  }
}
</style>

<style lang="scss">
.node-pool-scheduling-filter-popup:has(.checkbox-container[aria-label="Rank"]) {
  padding: 16px 10px;
  border-radius: 3px;
  box-shadow: -1px 0 7px rgba(0, 0, 0, 0.15);

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
