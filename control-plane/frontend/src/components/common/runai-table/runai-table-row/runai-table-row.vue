<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-tr class="runai-table-row" :class="{ selected: isSelected }" @click="rowClicked">
    <q-td class="q-table--col-auto-width" v-if="isSelectable">
      <q-checkbox
        size="md"
        :model-value="isSelected"
        @click="checkboxClicked"
        :aria-label="`Select row ${row.name || ''}`"
      />
    </q-td>
    <template v-for="col in columns">
      <component
        v-if="col.customCell"
        :key="shouldRerenderCell(col) ? row : 0"
        :aid="col.name"
        :is="col.customCell"
        :row="row"
        :column="col"
        :icon="nameColumnIcon"
        @custom-cell-event="col.customCellEvent && $emit('custom-cell-event', col.customCellEvent)"
      />
      <default-col v-else :row="row" :column="col" :key="col.name"></default-col>
    </template>
  </q-tr>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
// Models
import type { ITableColumn } from "@/models/table.model";
import { ECustomCell } from "@/models/table.model";
// Components
import {
  NameCol,
  ListCol,
  DefaultCol,
  ServiceAccountsStatusCol,
  LinkCol,
  ScopeCol,
  AuditStatusCol,
  OrgTreeCol,
  RoleAssigmentCol,
  StringsListCol,
  StatusCol,
  HtmlCol,
  SourceNameCol,
  ClusterIdToNameCol,
  WorkloadConnectionCol,
  DescriptionCol,
  ToolAccessCol,
  DepartmentQuotaCol,
  WorkloadStatusCol,
  ProgressCol,
  OrgUnitGpuEditableQuotaCol,
  OrgUnitQuotaDistributionCol,
  OrgUnitEditableWeightCol,
  OrgUnitWeightDistributionCol,
  IconButtonCol,
  RoleWithBadgeCol,
  ApplicationDepartmentCol,
  LegacyRoleCol,
  NewRoleCol,
  UserDisplayNameCol,
} from "@/components/common/runai-table/runai-custom-table-columns";

export default defineComponent({
  components: {
    OrgTreeCol,
    LinkCol,
    DefaultCol,
    NameCol,
    ListCol,
    ServiceAccountsStatusCol,
    ScopeCol,
    AuditStatusCol,
    RoleAssigmentCol,
    StringsListCol,
    StatusCol,
    WorkloadStatusCol,
    HtmlCol,
    SourceNameCol,
    ClusterIdToNameCol,
    WorkloadConnectionCol,
    DescriptionCol,
    ToolAccessCol,
    DepartmentQuotaCol,
    ProgressCol,
    OrgUnitGpuEditableQuotaCol,
    OrgUnitQuotaDistributionCol,
    OrgUnitEditableWeightCol,
    OrgUnitWeightDistributionCol,
    IconButtonCol,
    RoleWithBadgeCol,
    ApplicationDepartmentCol,
    LegacyRoleCol,
    NewRoleCol,
    UserDisplayNameCol,
  },
  props: {
    row: {
      type: Object as PropType<any>, // eslint-disable-line @typescript-eslint/no-explicit-any
      required: true,
    },
    columns: {
      type: Array as PropType<Array<ITableColumn>>,
      required: true,
    },
    isSelected: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
    nameColumnIcon: {
      type: String as PropType<string>,
      required: false,
    },
    isSelectable: {
      type: Boolean as PropType<boolean>,
      required: false,
    },
  },
  data() {
    return {
      ECustomCell,
    };
  },
  methods: {
    checkboxClicked(event: MouseEvent): void {
      this.$emit("checkbox-clicked", event);
    },
    rowClicked(event: MouseEvent): void {
      if (event.shiftKey) document.getSelection()?.removeAllRanges();
      this.$emit("row-clicked", event);
    },
    shouldRerenderCell(column: ITableColumn): boolean {
      return column.customCell === ECustomCell.DESCRIPTION_COL;
    },
  },
});
</script>
<style lang="scss" scoped>
.runai-table-row {
  cursor: pointer;

  &.selected {
    td::after {
      background-color: #0000000f;
    }
  }
}
</style>
