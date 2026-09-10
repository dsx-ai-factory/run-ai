<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-table
    ref="tableRef"
    class="runai-table"
    :rows="rows"
    :columns="columns"
    :visible-columns="filterBy.displayedColumns"
    :pagination="filterBy"
    :filter="searchTerm"
    @update:pagination="updateFilters"
    :selection="selectionMode"
    :selected="selected"
    :row-key="getRowKey"
    :class="tableClasses"
    :style="style"
    :binary-state-sort="true"
    :hide-selected-banner="true"
    table-class="runai-table"
    rows-per-page-label="Rows per page"
    :loading="loading"
    :hide-pagination="hidePagination"
    :flat="noShadow"
    :bordered="bordered"
    :square="!roundedCorners"
    @request="fetchPartialPagedData"
    :rows-per-page-options="rowsPerPageOptions"
    @update:selected="headerSelectionToggle"
    :hide-header="hideHeader"
  >
    <template v-if="overrideTableHeader" v-slot:header="props">
      <runai-table-header
        :props="props"
        :filter-by="filterBy"
        :is-selected="selected.length === rows.length && !loading"
        :is-selectable="!disableSelection"
        :selection="selectionMode"
        @checkbox-clicked="selectAllRows"
        @custom-header-event="onCustomHeaderEvent"
      />
    </template>
    <template v-slot:top-row="{ cols }">
      <runai-table-row
        v-for="(pinnedRow, pinnedPosition) in pinnedRows"
        :key="getRowKey(pinnedRow)"
        :columns="cols"
        :row="pinnedRow"
        :name-column-icon="getRowIcon(pinnedRow)"
        :is-selected="selectedRowKeys.has(getRowKey(pinnedRow))"
        :is-selectable="!disableSelection"
        @checkbox-clicked="
          isSingleSelect
            ? onSelect(getRowKey(pinnedRow), pinnedRow, pinnedPosition)
            : onMultipleSelect($event, {
                key: getRowKey(pinnedRow),
                row: pinnedRow,
                rowIndex: pinnedPosition,
                selected: selectedRowKeys.has(getRowKey(pinnedRow)),
              })
        "
        @row-clicked="onRowClicked($event, pinnedRow, pinnedPosition)"
        @custom-cell-event="$emit($event.emitName, pinnedRow)"
      />
    </template>

    <template v-slot:body="props">
      <runai-table-row
        v-if="!isPinned(props.row)"
        :key="props.key"
        :columns="props.cols"
        :row="props.row"
        :name-column-icon="getRowIcon(props.row)"
        :is-selectable="!disableSelection"
        @custom-cell-event="$emit($event.emitName, props.row)"
        @checkbox-clicked="
          isSingleSelect
            ? onSelect(props.key, props.row, visibleIndex(props.rowIndex))
            : onMultipleSelect($event, { ...props, rowIndex: visibleIndex(props.rowIndex) })
        "
        @row-clicked="onRowClicked($event, props.row, visibleIndex(props.rowIndex))"
        :is-selected="props.selected"
      />
    </template>

    <template v-slot:no-data>
      <slot v-if="!loading && !pinnedRows.length" name="no-data"></slot>
    </template>
  </q-table>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// Components
import { RunaiTableRow } from "@/components/common/runai-table/runai-table-row";
import { RunaiTableHeader } from "./runai-table-header";
// Models
import type { IFilterBy } from "@/models/filter.model";
import type { ITableColumn } from "@/models/table.model";

// Utils
import { isEqual } from "@/utils/common.util";

// Rows are entity-shaped and differ per index page, so the table stays row-agnostic.
export type TableRow = any; // eslint-disable-line @typescript-eslint/no-explicit-any

export type RowProps = {
  row: TableRow;
  key: string | number;
  rowIndex: number;
  selected: boolean;
};

export default defineComponent({
  components: {
    RunaiTableRow,
    RunaiTableHeader,
  },
  props: {
    rows: {
      type: Array as PropType<TableRow[]>,
      required: true,
    },
    columns: {
      type: Array as PropType<ITableColumn[]>,
      required: true,
    },
    filterBy: {
      type: Object as PropType<IFilterBy>,
      required: true,
    },
    loading: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
    selected: {
      type: Array as PropType<TableRow[]>,
      required: false,
      default: () => [],
    },
    disableSelection: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    topRow: {
      type: [Object, null] as PropType<TableRow | null>,
      required: false,
      default: null,
    },
    topRows: {
      type: Array as PropType<TableRow[]>,
      required: false,
      default: () => [],
    },
    getRowKey: {
      type: Function as PropType<(row: TableRow) => string | number>,
      required: false,
      default: () => "",
    },
    getRowIcon: {
      type: Function as PropType<(row: TableRow) => string>,
      required: false,
      default: () => "",
    },
    stickyColumns: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    hidePagination: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    noShadow: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
    modalView: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    stickyHeader: {
      // When using sticky header, height/max-height must be provided as well in the style prop
      type: Boolean as PropType<boolean>,
      default: false,
    },
    style: {
      type: Object as PropType<Record<string, string>>,
      default: () => ({}),
    },
    bordered: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
    isServerSidePagination: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    rowsPerPageOptions: {
      type: Array as PropType<number[]>,
      default: () => [0, 5, 7, 10, 15, 20, 25, 50],
    },
    selection: {
      type: String as PropType<"none" | "single" | "multiple">,
      default: "single",
    },
    overrideTableHeader: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    roundedCorners: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    hideHeader: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  data() {
    return {
      lastSelectedIndex: null as null | number,
    };
  },
  mounted() {
    this.$nextTick(() => this.applySortableHeaderA11y());
  },
  computed: {
    searchTerm(): string | undefined {
      return this.modalView ? this.filterBy.searchTerm : undefined;
    },
    pinnedRows(): TableRow[] {
      if (this.topRows.length) return this.topRows;
      return this.topRow ? [this.topRow] : [];
    },
    pinnedRowKeys(): Set<string | number> {
      return new Set(this.pinnedRows.map((row) => this.getRowKey(row)));
    },
    selectedRowKeys(): Set<string | number> {
      return new Set((this.selected ?? []).map((row: TableRow) => this.getRowKey(row)));
    },
    // Pinned rows render above the body and are skipped by it, so the raw rows order
    // is not what the user sees. Shift-range selection has to walk this order instead.
    visibleRows(): TableRow[] {
      return [...this.pinnedRows, ...this.rows.filter((row: TableRow) => !this.isPinned(row))];
    },
    // Positional rather than keyed: getRowKey defaults to "" on the pages that never
    // pass it, which would collapse every row onto one index.
    bodyVisibleIndexes(): number[] {
      let nextIndex = this.pinnedRows.length;
      return this.rows.map((row: TableRow) => (this.isPinned(row) ? -1 : nextIndex++));
    },
    isSingleSelect(): boolean {
      return this.selection === "single";
    },
    tableClasses(): Record<string, boolean> {
      return {
        "sticky-column-table": this.stickyColumns,
        "sticky-header": this.stickyHeader,
        "no-shadow": this.noShadow,
      };
    },
    selectionMode(): "none" | "single" | "multiple" {
      return this.disableSelection ? "none" : this.selection;
    },
  },
  methods: {
    isPinned(row: TableRow): boolean {
      return this.pinnedRowKeys.has(this.getRowKey(row));
    },
    visibleIndex(rowIndex: number): number {
      return this.bodyVisibleIndexes[rowIndex] ?? rowIndex;
    },
    rowAtIndex(index: number): TableRow {
      return this.visibleRows[index];
    },
    headerSelectionToggle(selectedRows: readonly TableRow[]): void {
      this.$emit("update:selected", selectedRows);
    },
    selectAllRows(): void {
      if (this.selected.length === this.rows.length) {
        this.$emit("update:selected", []);
      } else {
        this.$emit("update:selected", this.rows);
      }
    },
    onRowClicked(event: MouseEvent, row: TableRow, rowIndex: number): void {
      if (this.isSingleSelect || (!event.shiftKey && !event.metaKey && !event.ctrlKey)) {
        // Click on the row itself without any modifier keys acts as regular single selection.
        this.onSelect(this.getRowKey(row), row, rowIndex);
      } else {
        // Behave as regular multiple selection
        const isRowSelected = this.selected.find((currRow) => this.getRowKey(currRow) === this.getRowKey(row));
        this.onMultipleSelect(
          { shiftKey: event.shiftKey, metaKey: event.metaKey, ctrlKey: event.ctrlKey },
          { key: this.getRowKey(row), row, rowIndex: rowIndex, selected: isRowSelected },
        );
      }
    },
    onSelect(key: string | number, row: TableRow, rowIndex: number): void {
      if (this.selected.find((currRow) => this.getRowKey(currRow) === key)) {
        this.lastSelectedIndex = null;
        this.$emit("update:selected", []);
      } else {
        this.lastSelectedIndex = rowIndex;
        this.$emit("update:selected", [row]);
      }
    },
    onMultipleSelect(event: { shiftKey: boolean; metaKey: boolean; ctrlKey: boolean }, props: RowProps) {
      // If shift is not pressed, select/unselect the row
      if (!event.shiftKey) {
        if (!props.selected) {
          this.$emit("update:selected", [...this.selected, props.row]);
          this.lastSelectedIndex = props.rowIndex;
        } else {
          this.$emit(
            "update:selected",
            this.selected.filter((currRow) => this.getRowKey(currRow) !== props.key),
          );
          this.lastSelectedIndex = null;
        }
      } else {
        if (this.lastSelectedIndex === null) {
          // if the row isn't selected and there is no last selected row, select it
          if (!props.selected) {
            this.lastSelectedIndex = props.rowIndex;
            this.$emit("update:selected", [...this.selected, props.row]);
            return;
          }
        }

        // if the row is selected, unselect it
        if (props.selected) {
          this.lastSelectedIndex = null;
          this.$emit(
            "update:selected",
            this.selected.filter((currRow) => this.getRowKey(currRow) !== props.key),
          );
        }

        // if the row isn't selected, select it and all rows between the last selected row and this row
        this.handleRangeSelection(props.selected, props.rowIndex);
      }
    },
    updateFilters(newFilterBy: IFilterBy): void {
      if (isEqual(this.filterBy, newFilterBy)) return;

      if (this.isServerSidePagination) {
        this.fetchPartialPagedData(newFilterBy as { pagination: IFilterBy });
      } else {
        this.$emit("update-filters", { ...this.filterBy, ...newFilterBy });
      }
    },
    fetchPartialPagedData(props: { pagination: IFilterBy }) {
      this.$emit("update-filters", props.pagination);
    },
    handleRangeSelection(isSelected: boolean, rowIndex: number) {
      if (this.lastSelectedIndex === null) return;
      const rowsToAdd: Array<TableRow> = [];
      const rowKeysToRemove: Array<string | number> = [];

      const start = Math.min(this.lastSelectedIndex, rowIndex);
      const end = Math.max(this.lastSelectedIndex, rowIndex);
      for (let i = start; i <= end; i++) {
        const row = this.rowAtIndex(i);
        if (!row) continue;
        const key = this.getRowKey(row);
        if (isSelected) {
          rowKeysToRemove.push(key);
        } else if (!this.selected.find((currRow) => this.getRowKey(currRow) === key)) {
          rowsToAdd.push(row);
        }
      }

      if (isSelected) {
        this.$emit(
          "update:selected",
          this.selected.filter((currRow) => !rowKeysToRemove.includes(this.getRowKey(currRow))),
        );
      } else {
        this.$emit("update:selected", [...this.selected, ...rowsToAdd]);
      }
    },
    onCustomHeaderEvent(event: { emitName: string; value: unknown }) {
      this.$emit(event.emitName, event.value);
    },
    // a11y: Quasar's <q-th> has inheritAttrs:false so :tabindex/:aria-sort bindings
    // don't reach the underlying <th>. We imperatively set them after render.
    applySortableHeaderA11y(): void {
      const tableEl = (this.$refs.tableRef as { $el?: HTMLElement })?.$el;
      if (!tableEl || typeof tableEl.querySelectorAll !== "function") return;
      tableEl.querySelectorAll<HTMLElement>("th.sortable").forEach((th) => {
        th.setAttribute("tabindex", "0");
        const isSorted = th.classList.contains("sorted");
        const isDesc = isSorted ? (this.filterBy as { descending?: boolean }).descending : undefined;
        th.setAttribute("aria-sort", isSorted ? (isDesc ? "descending" : "ascending") : "none");
        if (!th.dataset.a11yKeydown) {
          th.addEventListener("keydown", (e: KeyboardEvent) => {
            if (e.key === "Enter" || e.key === " ") {
              e.preventDefault();
              th.click();
            }
          });
          th.dataset.a11yKeydown = "true";
        }
      });
    },
  },
  watch: {
    filterBy: {
      deep: true,
      handler() {
        this.$nextTick(() => this.applySortableHeaderA11y());
      },
    },
    rows() {
      this.$nextTick(() => this.applySortableHeaderA11y());
    },
  },
});
</script>
<style lang="scss" scoped></style>
