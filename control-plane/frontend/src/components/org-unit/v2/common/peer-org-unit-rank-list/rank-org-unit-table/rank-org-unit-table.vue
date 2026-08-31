<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-table
    class="rank-org-unit-rank-table"
    :rows="rowsWithPendingChanges"
    :columns="columns"
    :filter-by="filterBy"
    :loading="false"
    :get-row-key="getRowKey"
    v-model:selected="selectedOrgUnits"
    modal-view
    :bordered="false"
    hide-header
    no-shadow
    :rows-per-page-options="[50]"
    selection="multiple"
    @update-filters="handleUpdateFilters"
  >
    <template #no-data>
      <runai-table-no-data
        v-if="!loading"
        :entity-name="entityType"
        :filter-by="filterBy"
        :show-create-btn="false"
        :show-filter-icon-and-button="false"
      />
    </template>
  </runai-table>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { RunaiTable } from "@/components/common/runai-table";
import { RunaiTableNoData } from "@/components/common/runai-table/runai-table-no-data";
import type { IFilterBy } from "@/models/filter.model";
import { ECustomCell, type ICellIconConfig, type ITableColumn } from "@/models/table.model";
import type { DepartmentNodePoolResources, ProjectNodePoolResources } from "@/swagger-models/org-unit-service-client";
import { EQuotaEntity } from "@/models/resource.model";

// Union type to support both project and department node pool resources
export type OrgUnitNodePoolResources = ProjectNodePoolResources | DepartmentNodePoolResources;

// Columns displayed in this table
const DISPLAYED_COLUMNS = ["name", "weight", "distribution"];

export interface Props {
  orgUnits: OrgUnitNodePoolResources[];
  entityType?: EQuotaEntity;
  loading?: boolean;
  selected?: OrgUnitNodePoolResources[];
  nodePoolId: string;
  pendingWeightChanges?: Map<string, { newWeight: number; originalWeight: number | null }>;
  totalWeightInRank: number; // Passed from parent (already fetched)
  isOverQuotaPriorityEnabled?: boolean;
  currentOrgUnitId?: string; // ID of the current org unit being edited (for star indicator)
  parentFilterBy?: IFilterBy; // FilterBy from parent (includes searchTerm, rowsNumber for pagination)
}

const props = withDefaults(defineProps<Props>(), {
  entityType: EQuotaEntity.project,
  orgUnits: () => [],
  loading: false,
  selected: () => [],
  pendingWeightChanges: () => new Map(),
  totalWeightInRank: 0,
  isOverQuotaPriorityEnabled: true,
  currentOrgUnitId: undefined,
  parentFilterBy: undefined,
});

const emit = defineEmits<{
  (
    e: "weight-changed",
    payload: {
      orgUnitId: string;
      nodePoolId: string;
      weight: number;
      originalWeight: number | null;
    },
  ): void;
  (e: "update-filters", filterBy: IFilterBy): void;
}>();

const selectedOrgUnits = defineModel<OrgUnitNodePoolResources[]>("selected", { default: () => [] });

// Forward filter updates to parent for server-side pagination
const handleUpdateFilters = (newFilterBy: IFilterBy) => {
  emit("update-filters", newFilterBy);
};

const entityLabel = computed(() => (props.entityType === "department" ? "Department" : "Project"));

// Merge parent's filterBy with local displayedColumns
// This ensures the no-data component shows the correct search message
const filterBy = computed<IFilterBy>(() => ({
  ...props.parentFilterBy,
  displayedColumns: DISPLAYED_COLUMNS,
}));

const getRowKey = (row: OrgUnitNodePoolResources) => row.id;

// Transform rows to include pending weight changes
const rowsWithPendingChanges = computed(() => {
  return props.orgUnits.map((orgUnit) => {
    const serverWeight = orgUnit.resources?.gpu?.overQuotaWeight ?? null;
    const deserved = orgUnit.resources?.gpu?.deserved ?? null;
    const pendingChange = props.pendingWeightChanges?.get(orgUnit.orgUnitId);
    const weight = pendingChange ? pendingChange.newWeight : serverWeight;
    const isCurrentOrgUnit = props.currentOrgUnitId === orgUnit.orgUnitId;

    return {
      ...orgUnit,
      weight, // Display weight (pending or server)
      serverWeight, // Original weight from server
      deserved, // GPU deserved value (used when disabled)
      isCurrentOrgUnit, // Flag for star indicator
    };
  });
});

// Disabled tooltip based on entity type
const disabledTooltip = computed(() => {
  return `The ${props.entityType}'s weight is proportional to its assigned GPU quota. To set a custom value, contact your administrator.`;
});

// Cell icon for current org unit indicator
const currentOrgUnitIcon = (row: { isCurrentOrgUnit?: boolean }): ICellIconConfig | undefined => {
  if (row.isCurrentOrgUnit) {
    return {
      name: "ph-fill ph-star",
      color: "black-54",
      size: "16px",
    };
  }
  return undefined;
};

const columns = computed<ITableColumn[]>(() => [
  {
    name: "name",
    label: entityLabel.value,
    field: "orgUnitName",
    align: "left" as const,
    sortable: false,
    display: true,
    style: "width: 33%;",
    cellIcon: currentOrgUnitIcon,
    customCell: ECustomCell.STRINGS_LIST_COL,
    format: (orgUnitName) => [orgUnitName],
  },
  {
    name: "weight",
    label: "Weight",
    field: "weight",
    align: "left" as const,
    sortable: false,
    display: true,
    style: "width: 33%; ",
    customCell: ECustomCell.ORG_UNIT_EDITABLE_WEIGHT_COL,
    customCellConfig: {
      [ECustomCell.ORG_UNIT_EDITABLE_WEIGHT_COL]: {
        nodePoolId: props.nodePoolId,
        onSave: (payload: { orgUnitId: string; nodePoolId: string; weight: number; originalWeight: number | null }) => {
          emit("weight-changed", payload);
        },
        disabled: !props.isOverQuotaPriorityEnabled,
        disabledTooltip: disabledTooltip.value,
      },
    },
  },
  {
    name: "distribution",
    label: "Weight distribution",
    field: "weight",
    align: "left" as const,
    sortable: false,
    display: true,
    style: "width: 33%",
    customCell: ECustomCell.WEIGHT_DISTRIBUTION_COL,
    customCellConfig: {
      [ECustomCell.WEIGHT_DISTRIBUTION_COL]: {
        totalWeightInRank: props.totalWeightInRank,
        entityType: props.entityType,
      },
    },
  },
]);
</script>

<style lang="scss">
.rank-org-unit-rank-table {
  .runai-table-row td:first-child {
    padding-left: 50px;
  }
  .runai-table-row td:nth-child(2) {
    width: 230px; // fit the org unit name long text overflow
  }
}
</style>

<style scoped lang="scss">
// Smooth entrance animation for table rows
:deep(.runai-table-row) {
  animation: row-fade-slide-in 0.3s ease-out;
}

@keyframes row-fade-slide-in {
  from {
    opacity: 0;
    transform: translateX(-8px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}
</style>
