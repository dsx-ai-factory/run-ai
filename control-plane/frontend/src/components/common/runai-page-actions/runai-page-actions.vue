<template>
  <section class="runai-page-actions" ref="actions">
    <Transition name="flicker">
      <div class="selected-rows-actions q-px-md" v-if="showSelectedRowsActions && !disableActions">
        <span class="title text-b-subtitle1">{{ selectedRowsActionsLabel }}</span>
        <q-separator :vertical="true" color="white" class="separator q-my-md q-mx-md" />
        <div class="actions-container">
          <slot name="selected-rows-actions" />
        </div>
        <runai-action-button
          v-if="showDetailsBtn"
          :btn-action="isDrawerOpen ? 'hideDetails' : 'showDetails'"
          @click="toggleDetailsDrawer"
        />
        <q-btn
          flat
          round
          color="white"
          icon="ph-bold ph-x"
          class="runai-icon-btn--default"
          @click="$emit('selected-actions-close')"
          aid="close-actions-btn"
          aria-label="Close selection"
        />
      </div>
    </Transition>

    <Transition name="slide-to-top">
      <div class="main-actions" v-if="!showSelectedRowsActions || disableActions">
        <div class="left-actions">
          <template v-if="showPrimaryBtn">
            <slot name="primary-btn">
              <div>
                <q-btn
                  v-permission="{ resourceType, action: Action.Create }"
                  color="primary"
                  class="btn-icon-sm primary-cta-btn runai-icon-btn--default"
                  :icon="primaryBtnIcon"
                  aid="primary-cta-btn"
                  :label="primaryBtnLabel"
                  @click="$emit('create-entity')"
                  :disable="primaryBtnDisable"
                >
                </q-btn>
                <q-tooltip v-if="primaryBtnDisable && primaryToolTipRows">
                  <div v-for="row in primaryToolTipRows" :key="row">
                    {{ row }}
                  </div>
                </q-tooltip>
              </div>
            </slot>
          </template>
        </div>
        <runai-page-filter
          v-if="!advancedFilters"
          :initial-filters-model="columnFilters"
          @update-filter-by="updateFilterBy($event)"
          :filter-options="filterOptions"
          :filter-input-label="filterInputLabel"
          :hide-all-cluster-filter-option="hideAllClusterFilterOption || !isMultiCluster"
          :filter-by="filters"
          :hide-cluster-filter="hideClusterFilter"
        ></runai-page-filter>
        <runai-advanced-filters
          v-else
          @update-filter-by="updateFilterBy($event)"
          :filter-options="advancedFilterOptions"
          :initial-filters-model="advancedColumnFilters"
          :hide-all-cluster-filter-option="hideAllClusterFilterOption || !isMultiCluster"
          :filter-by="filters"
          :hide-cluster-filter="hideClusterFilter"
        />
        <div class="right-actions-container">
          <runai-general-search
            v-if="showSearch"
            :search-term="searchTerm"
            :show-search-text="!hideSearchText"
            :tooltip-text="searchTooltipText"
            @update-search-term="updateSearchTerm"
          ></runai-general-search>
          <q-btn
            v-if="!hideColumnsBtn"
            size="sm"
            class="q-mx-sm runai-icon-btn--default"
            aid="run-page-action-select-columns-btn"
            flat
            icon="ph-bold ph-columns"
            label="Columns"
            stack
          >
            <q-menu transition-show="scale" transition-hide="scale" @before-hide="beforeClose" max-width="1000px">
              <runai-checkbox-list
                :options="allColumnOptions"
                :selected="selectedColumns"
                @changed="updateSelectedColumns"
              />
            </q-menu>
          </q-btn>
          <runai-actions-menu
            class="q-mx-sm"
            v-if="filteredActionMenuItems?.length"
            :action-items="filteredActionMenuItems"
          />
          <q-btn
            v-if="supportRefresh"
            size="sm"
            class="q-mx-sm refresh-data runai-icon-btn--default"
            flat
            icon="ph-bold ph-arrow-clockwise"
            aid="refresh-data-button"
            label="refresh"
            stack
            :disable="refreshingData"
            @click="refreshData"
          >
            <q-tooltip
              :model-value="refreshingData"
              no-parent-event
              anchor="center left"
              self="center right"
              :offset="[0, 0]"
              class="text-no-wrap"
              >Data is up to date</q-tooltip
            >
          </q-btn>

          <q-btn
            v-if="showDetailsBtn"
            size="sm"
            class="q-mx-sm show-details-btn runai-icon-btn--default"
            flat
            icon="ph-bold ph-file-text"
            aid="show-details-button"
            :label="`${isDrawerOpen ? 'hide' : 'show'} details`"
            stack
            @click="toggleDetailsDrawer"
          />
        </div>
      </div>
    </Transition>
    <runai-side-drawer
      @open="isDrawerOpen = true"
      @close="isDrawerOpen = false"
      :is-open="isDrawerOpen"
      v-if="showDetailsBtn"
      :is-empty="!selectedRowsAmount"
      :empty-message="emptyMessage"
      :height-from-top="heightFromTop"
      @resize="onResize"
    >
      <template v-slot:drawer-content>
        <slot name="drawer-content" />
      </template>
    </runai-side-drawer>
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";

// models
import type { ICheckboxItem } from "@/components/common/runai-checkbox-list/model";
import {
  type IFilterModel,
  type IFilterOption,
  type IFilterBy,
  type IAdvancedFilterOption,
  type IAdvancedFilterModel,
  CLUSTER_COLUMN_FILTER_NAME,
} from "@/models/filter.model";
import type { ITableColumn } from "@/models/table.model";
import type { ActionMenuItem } from "@/components/common/runai-page-actions/model";
import { type ResourceType, Action } from "@/swagger-models/authorization-client";

// components
import { RunaiCheckboxList } from "@/components/common/runai-checkbox-list";
import { RunaiPageFilter } from "@/components/common/runai-page-filter";
import { RunaiGeneralSearch } from "@/components/common/runai-general-search";
import { RunaiActionButton } from "./runai-action-button";
import { RunaiSideDrawer } from "@/components/common/runai-side-drawer";

// stores
import { useClusterStore } from "@/stores/cluster.store";

// services
import { filterService } from "@/services/filter.service/filter.service";
import { RunaiActionsMenu } from "@/components/common/runai-page-actions/runai-actions-menu/";
import { RunaiAdvancedFilters } from "@/components/common/runai-advanced-filters";
import { isEqual } from "@/utils/common.util";
import { formatUtil } from "@/utils/format.util/format.util";
import { MAX_RECORDS_TO_EXPORT } from "@/composables/helpers/use-export-file.mapper";

export default defineComponent({
  components: {
    RunaiAdvancedFilters,
    RunaiActionsMenu,
    RunaiCheckboxList,
    RunaiPageFilter,
    RunaiGeneralSearch,
    RunaiSideDrawer,
    RunaiActionButton,
  },
  emits: [
    "columns-changed",
    "selected-actions-close",
    "filters-changed",
    "create-entity",
    "show-details",
    "drawer-resized",
    "drawer-toggled",
    "export-csv",
    "refresh-data",
  ],
  props: {
    advancedFilters: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    columns: {
      type: Array as PropType<ITableColumn[]>,
      required: true,
    },
    filters: {
      type: Object as PropType<IFilterBy>,
      required: true,
    },
    selectedRowsAmount: {
      type: Number as PropType<number>,
      required: false,
    },
    showPrimaryBtn: {
      type: Boolean as PropType<boolean>,
      default: true,
      required: false,
    },
    showSearch: {
      type: Boolean as PropType<boolean>,
      default: true,
      required: false,
    },
    hideColumnsBtn: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    hideSearchText: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    primaryBtnLabel: {
      type: String as PropType<string>,
      required: false,
    },
    primaryBtnIcon: {
      type: String as PropType<string>,
      default: "ph-bold ph-plus",
    },
    primaryBtnDisable: {
      type: Boolean as PropType<boolean>,
      default: false,
      required: false,
    },
    primaryTooltipText: {
      type: String as PropType<string>,
      required: false,
    },
    filterInputLabel: {
      type: String as PropType<string>,
      required: false,
    },
    actionMenuItems: {
      type: Array as PropType<ActionMenuItem[]>,
      required: false,
    },
    hideExportCsv: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    showDetailsBtn: {
      type: Boolean as PropType<boolean>,
      default: false,
      required: false,
    },
    emptyMessage: {
      type: String as PropType<string>,
      default: "Select a row to see it's details",
      required: false,
    },
    drawerDarkMode: {
      type: Boolean as PropType<boolean>,
      default: false,
      required: false,
    },
    disableActions: {
      type: Boolean as PropType<boolean>,
      default: false,
      required: false,
    },
    supportRefresh: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    hideAllClusterFilterOption: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    hideClusterFilter: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    searchTooltipText: {
      type: String as PropType<string>,
      default: "",
    },
  },
  data() {
    return {
      selectedColumns: [] as Array<string>,
      isDrawerOpen: false as boolean,
      heightFromTop: 0 as number,
      clusterStore: useClusterStore(),
      refreshingData: false,
    };
  },
  created() {
    this.selectedColumns = this.columns
      .filter((col: ITableColumn) => this.displayedColumns.has(col.name) || col.mandatory)
      .map((col) => col.name);
  },
  mounted() {
    this.setHeightFromTop();
    addEventListener("resize", this.setHeightFromTop);
  },
  computed: {
    isMultiCluster(): boolean {
      return this.clusterStore.isMultiCluster;
    },
    allColumnOptions(): Array<ICheckboxItem> {
      return this.columns.map((col: ITableColumn) => ({
        label: col.label,
        value: col.name,
        disable: col.mandatory,
      }));
    },
    Action(): typeof Action {
      return Action;
    },
    resourceType(): ResourceType {
      return this.$route?.meta.resourceType as ResourceType;
    },
    filteredActionMenuItems(): ActionMenuItem[] {
      const csvOption: ActionMenuItem = {
        label: "Download as CSV",
        action: () => this.$emit("export-csv"),
        ...(this.filters.rowsNumber != null &&
          this.filters.rowsNumber >= MAX_RECORDS_TO_EXPORT && {
            tooltip: `Export to CSV is limited to ${formatUtil.formatNumberWithCommas(MAX_RECORDS_TO_EXPORT)} rows`,
          }),
      };
      if (this.actionMenuItems?.length) {
        if (this.hideExportCsv) {
          return this.actionMenuItems;
        }
        return [csvOption, ...this.actionMenuItems];
      }
      if (this.hideExportCsv) {
        return [];
      }
      return [csvOption];
    },
    columnFilters(): Array<IFilterModel> {
      return this.filters.columnFilters || ([] as Array<IFilterModel>);
    },
    advancedColumnFilters(): Array<IAdvancedFilterModel> {
      return this.filters.advancedFilters || [];
    },
    searchTerm(): string {
      return this.filters.searchTerm || "";
    },
    filterOptions(): Array<IFilterOption> {
      return filterService.getFilterOptions(this.columns.filter((c) => !c.hideFilter));
    },
    advancedFilterOptions(): Array<IAdvancedFilterOption> {
      return filterService.getAdvancedFilterOptions(this.columns.filter((c) => !c.hideFilter));
    },
    selectedRowsActionsLabel(): string {
      return this.selectedRowsAmount + " selected";
    },
    showSelectedRowsActions(): boolean {
      return !!this.selectedRowsAmount;
    },
    displayedColumns(): Set<string> {
      return new Set(this.filters.displayedColumns);
    },
    primaryToolTipRows(): string[] | null {
      if (!this.primaryTooltipText) return null;
      return this.primaryTooltipText.split("\n");
    },
    clusterFilter(): IFilterModel | undefined {
      return { field: "", name: CLUSTER_COLUMN_FILTER_NAME, term: this.filters.clusterUuid || "" };
    },
  },
  methods: {
    refreshData(): void {
      this.$emit("refresh-data");
      this.refreshingData = true;
      setTimeout(() => (this.refreshingData = false), 5000);
    },
    updateSelectedColumns(selectedColumns: Array<string>): void {
      this.selectedColumns = selectedColumns;
    },
    beforeClose(): void {
      if (isEqual(this.selectedColumns, this.filters.displayedColumns)) return;
      this.updateFilterBy({ ...this.filters, displayedColumns: this.selectedColumns });
    },
    updateFilterBy(filterBy: IFilterBy): void {
      this.$emit("filters-changed", filterBy);
    },
    updateSearchTerm(searchTerm: string): void {
      this.updateFilterBy({ ...this.filters, searchTerm });
    },
    toggleDetailsDrawer(): void {
      this.setHeightFromTop();
      this.isDrawerOpen = !this.isDrawerOpen;
    },
    openDrawer(): void {
      this.setHeightFromTop();
      this.isDrawerOpen = true;
    },
    setHeightFromTop(): void {
      this.$nextTick(() => {
        const actionPageElement = this.$refs.actions as HTMLElement;
        if (actionPageElement) {
          this.heightFromTop = actionPageElement.getBoundingClientRect().bottom;
        }
        // This condition is to handle a glitch which occurs when the page is refreshed and the side drawer is open.
        if (this.heightFromTop === 0) {
          setTimeout(() => {
            this.setHeightFromTop();
          }, 50);
        }
      });
    },
    onResize(width: string): void {
      this.$emit("drawer-resized", width);
    },
  },
  watch: {
    "clusterStore.clusterAlert": {
      handler() {
        this.setHeightFromTop();
      },
      deep: true,
      immediate: true,
    },
    isDrawerOpen: {
      handler(val: boolean) {
        this.$emit("drawer-toggled", val);
      },
    },
  },
  unmounted() {
    removeEventListener("resize", this.setHeightFromTop);
  },
});
</script>

<style lang="scss" scoped>
.runai-page-actions {
  padding: 9px 0 9px var(--space-m);
  background-color: var(--bg-primary-default, #{$white});
  overflow: hidden;
  position: relative;
  min-height: $runai-page-actions-min-height;
  border-bottom: 1px solid var(--border-subtle, #{$grey-4});

  .main-actions {
    display: flex;
    justify-content: space-between;
    align-items: center;

    .left-actions,
    .right-actions-container {
      display: flex;

      .primary-cta-btn {
        flex-shrink: 0;
        margin-right: var(--space-m);
      }
    }

    // Action-bar utility icons (search, columns, more, refresh, show/hide details)
    // render in icon-default color.
    .right-actions-container {
      color: var(--icon-default, #707070);

      :deep(.q-btn) {
        color: var(--icon-default, #707070);
      }

      :deep(.q-icon) {
        color: var(--icon-default, #707070);
      }
    }

    .runai-page-filter {
      flex: 1;
    }

    .show-details-button {
      width: 103px;
    }
  }

  .selected-rows-actions {
    background-color: var(--interactive-primary-default-bg);
    position: absolute;
    left: 0;
    top: 0;
    right: 0;
    bottom: 0;
    z-index: 1;
    display: flex;
    align-items: center;
    color: var(--text-inverse, #{$grey-0});

    .title {
      color: inherit;
    }

    .actions-container {
      flex: 1;

      overflow-x: auto;

      &::-webkit-scrollbar {
        width: 8px;
        height: 8px;
      }

      &::-webkit-scrollbar-thumb {
        background-color: $black-35;
        border-radius: 4px;
      }

      &::-webkit-scrollbar-track {
        background: transparent;
      }

      :deep(> .row.items-center) {
        flex-wrap: nowrap;
        white-space: nowrap;
      }
    }
  }

  .separator {
    opacity: 0.2;
  }
}
</style>
