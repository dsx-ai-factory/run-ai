<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="runai-form-card-section">
    <template v-if="!loading && cardsList.length === 0 && !listQuery?.search">
      <section class="empty-box column justify-center items-center q-pa-md" v-if="emptyMessage">
        <q-icon v-if="iconName" :name="iconName" size="44px" class="text-icon-default" />
        <template v-if="canCreateEntity">
          <p class="empty-message text-center text-italic q-pt-md">{{ emptyMessage }}</p>
          <slot name="add-new">
            <q-btn
              :disable="disableCreateNew"
              class="create-new-btn"
              :aid="addNewBtnAid"
              flat
              color="primary"
              @click="createNew"
            >
              <q-icon left name="ph-bold ph-plus" size="16px"></q-icon>
              <div>new {{ entityName }}</div>
            </q-btn>
          </slot>
        </template>
        <template v-else>
          <slot name="no-roles" />
        </template>
      </section>
    </template>

    <template v-else>
      <section class="header">
        <div v-if="mainMessage" class="subheader q-pa-sm q-pb-md row">
          {{ mainMessage }}
        </div>

        <div v-if="allowActions" class="actions-container row justify-between items-center q-pl-sm q-pb-md">
          <div class="row q-gutter-x-sm">
            <q-select
              class="cards-sort-dropdown"
              v-if="Object.keys(sortOptions).length"
              dense
              outlined
              :options="selectableSortOptions"
              :model-value="relevantSortOption"
              @update:model-value="onSortChanged"
            >
              <template #prepend>
                <q-icon size="16px" name="ph-bold ph-sort-descending" />
              </template>
            </q-select>
            <runai-general-search
              :aid="entityName + '-search-card-input'"
              :search-term="relevantSearchTerm"
              @update-search-term="onSearch"
              :show-search-text="false"
            ></runai-general-search>
          </div>

          <slot name="add-new" v-if="entityName && canCreateEntity">
            <q-btn
              class="add-new-btn"
              flat
              color="primary"
              :aid="addNewBtnAid"
              :disable="disabled || disableCreateNew"
              @click="createNew"
            >
              <q-icon left name="ph-bold ph-plus" size="16px"></q-icon>
              <div>new {{ entityName }}</div>
            </q-btn>
          </slot>
        </div>
      </section>

      <div v-if="(localSearchTerm || listQuery?.search) && list.length === 0" class="column items-center q-pa-xl">
        <q-icon name="ph-bold ph-funnel" size="44px" class="filter-no-results-icon" />
        <div class="q-pt-md">No {{ searchName }} match your filters</div>
      </div>

      <runai-card-list
        v-else
        class="cards-list-container"
        :list="list"
        :loading="loading"
        :multi-select="multiSelect"
        :selected-items="selectedCards"
        :total-count="totalCount"
        :list-query="listQuery"
        @changed="onCardClicked"
        @get-next-batch="onGetNextBatch"
        :disabled="disabled"
      >
        <template #empty-data-display>
          <slot name="empty-data-display" />
        </template>
      </runai-card-list>
    </template>
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// Components
import { RunaiCardList } from "@/components/common/runai-card-list";

// Models
import type { ICardListItem } from "@/components/common/runai-card-list";

// utils
import { spaceToDash, toLowerCase } from "@/utils/string.util";
import type { ISelectOption } from "@/models/global.model";
import { RunaiGeneralSearch } from "@/components/common/runai-general-search";
import { dateUtil } from "@/utils/date.util";
import type { ICardQuery, TCardSortValues } from "../runai-card-list/runai-card-list.model";

export interface ICardSortOption extends ISelectOption {
  label: string;
  value: TCardSortValues;
}

export type TCardSortFunction = (a: ICardListItem, b: ICardListItem) => number;

export interface ICardSupportedSortOptions {
  name?: boolean;
  creation?: boolean;
  recentUsage?: boolean;
}

export type TSortFunctions = TCardSortValues | "default";

const sortFunctionsMap: Record<TSortFunctions, TCardSortFunction> = {
  name: (cardA: ICardListItem, cardB: ICardListItem) => {
    if (!cardB.sortInfo?.name || !cardA.sortInfo?.name) return 0;
    return cardA.sortInfo.name.localeCompare(cardB.sortInfo.name);
  },
  recentUsage: (cardA: ICardListItem, cardB: ICardListItem) =>
    dateUtil.dateToTimestamp(cardB.sortInfo?.recentUsage || undefined) -
    dateUtil.dateToTimestamp(cardA.sortInfo?.recentUsage || undefined),
  creation: (cardA: ICardListItem, cardB: ICardListItem) =>
    dateUtil.dateToTimestamp(cardB.sortInfo?.createdAt || undefined) -
    dateUtil.dateToTimestamp(cardA.sortInfo?.createdAt || undefined),
  default: () => 0,
};

export const START_FROM_SCRATCH_ID = "-1";

const allSortOptions: ICardSortOption[] = [
  { label: "Last used", value: "recentUsage" },
  { label: "Last created", value: "creation" },
  { label: "Name", value: "name" },
];

export default defineComponent({
  components: {
    RunaiCardList,
    RunaiGeneralSearch,
  },
  emits: ["selected-card-changed", "create-new", "query-changed", "get-next-batch"],
  props: {
    loading: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    emptyMessage: {
      type: String as PropType<string>,
      required: false,
    },
    mainMessage: {
      type: String as PropType<string>,
      required: false,
    },
    entityName: {
      type: String as PropType<string>,
      required: false,
      default: "",
    },
    cardsList: {
      type: Array as PropType<Array<ICardListItem>>,
      required: true,
    },
    selectedCardsIds: {
      type: Array as PropType<Array<string>>,
      required: true,
    },
    iconName: {
      type: String as PropType<string>,
      required: false,
    },
    searchName: {
      type: String as PropType<string>,
      required: true,
    },
    multiSelect: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    allowActions: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    disabled: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    disableCreateNew: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    sortOptions: {
      type: Object as PropType<ICardSupportedSortOptions>,
      default: () => ({}),
    },
    defaultSortOption: {
      type: String as PropType<TCardSortValues>,
      required: false,
    },
    canCreateEntity: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: true,
    },
    listQuery: {
      type: Object as PropType<ICardQuery>,
      required: false,
    },
    totalCount: {
      type: Number as PropType<number>,
      required: false,
    },
  },
  data() {
    return {
      addNewBtnAid: `add-new-${toLowerCase(spaceToDash(this.entityName))}-btn` as string,
      localSearchTerm: "" as string,
      localSelectedSortOption: null as ICardSortOption | null,
    };
  },
  created() {
    this.setDefaultSortOption();
  },
  computed: {
    relevantSearchTerm(): string {
      if (this.listQuery?.search) return this.listQuery.search.toString();
      return this.localSearchTerm;
    },
    relevantSortOption(): ICardSortOption | null {
      if (this.listQuery?.sortBy)
        return allSortOptions.find((sortOpt) => sortOpt.value === this.listQuery?.sortBy) || null;
      return this.localSelectedSortOption;
    },
    selectableSortOptions(): ICardSortOption[] {
      return allSortOptions.filter((sortOption) => !!this.sortOptions[sortOption.value]);
    },
    searchPlaceholder(): string {
      return `Search ${this.searchName}`;
    },
    selectedCards(): Array<string> {
      const ids: Set<string> = new Set(this.selectedCardsIds);
      return this.cardsList
        .filter((card: ICardListItem) => ids.has(String(card.id)) || card.locked)
        .map((card: ICardListItem) => card.id);
    },
    list(): Array<ICardListItem> {
      // if this prop is true, it means the parent is handling the sort and filter (usually mean backend filtering)
      if (this.listQuery) return this.cardsList;

      const sortFunc = this.localSelectedSortOption?.value
        ? sortFunctionsMap[this.localSelectedSortOption.value] || sortFunctionsMap.default
        : sortFunctionsMap.default;

      let startFromScarchCard: ICardListItem | null = null;

      const filteredAndSorted = this.cardsList
        .filter((card: ICardListItem) => {
          if (!this.localSearchTerm) {
            if (card.id === START_FROM_SCRATCH_ID) {
              startFromScarchCard = card;
              return false;
            }
            return true;
          }
          return this.isIncludeSearchTerm(card, this.localSearchTerm);
        })
        .sort(sortFunc);

      if (startFromScarchCard) {
        return [startFromScarchCard, ...filteredAndSorted];
      }

      return filteredAndSorted;
    },
    lockedCards(): Array<ICardListItem> {
      return this.cardsList.filter((card: ICardListItem) => card.locked);
    },
  },
  methods: {
    onCardClicked(selectedCardsIds: Array<string>): void {
      // locked cards are assets imposed by policy.
      // we only make them look as if they are selected but in reality they do not exists within the model.
      // therefore we need to filter them out
      this.$emit(
        "selected-card-changed",
        selectedCardsIds.filter((id: string) => !this.lockedCards.find((card: ICardListItem) => card.id === id)),
      );
    },
    createNew(): void {
      this.$emit("create-new");
    },
    onSearch(term: string | number | null): void {
      if (this.listQuery) {
        this.queryChanged({
          ...this.listQuery,
          search: term === null ? "" : String(term),
        });
        return;
      }

      this.localSearchTerm = term === null ? "" : String(term);
    },
    onSortChanged(sortOption: ICardSortOption | null): void {
      if (this.listQuery) {
        this.queryChanged({
          ...this.listQuery,
          sortBy: sortOption?.value || null,
        });
        return;
      }

      this.localSelectedSortOption = sortOption;
    },
    queryChanged(q: ICardQuery): void {
      this.$emit("query-changed", q);
    },
    onGetNextBatch(amountRequired: number): void {
      if (!this.listQuery) return;

      this.$emit("get-next-batch", amountRequired);
    },
    isIncludeSearchTerm(card: ICardListItem, searchTerm: string): boolean {
      const values: Array<string> = card.searchValues || [];
      const searchTermLowercase: string = searchTerm.toLowerCase();
      return values.some((currVal: string) => currVal.toLowerCase().includes(searchTermLowercase));
    },
    setDefaultSortOption(): void {
      if (this.listQuery?.sortBy) return;

      if (!this.defaultSortOption) this.localSelectedSortOption = allSortOptions[2];
      this.localSelectedSortOption =
        allSortOptions.find((sortOpt) => sortOpt.value === this.defaultSortOption) || allSortOptions[2];
    },
  },
});
</script>

<style lang="scss" scoped>
.runai-form-card-section {
  .search-icon {
    font-size: 16px;
  }
  .cards-sort-dropdown {
    width: 180px;
  }

  .filter-no-results-icon {
    color: #616161;
  }
}
</style>
