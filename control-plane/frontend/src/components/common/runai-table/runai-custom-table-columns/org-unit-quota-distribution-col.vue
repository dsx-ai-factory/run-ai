<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-td class="org-unit-quota-distribution-col">
    <org-unit-quota-distribution-chart
      :entity-name="row.name"
      :entity-type="config?.entityType || EQuotaEntity.project"
      :assigned-quota="isEntityUnlimited ? 0 : (row.gpuQuota ?? 0)"
      :parent-name="config?.parentName || ''"
      :node-pool-name="config?.nodePoolName || ''"
      :parent-available-quota="config?.totalQuota ?? 0"
      :other-entities-assigned-quota="otherEntitiesQuota"
      :is-unlimited="config?.isUnlimited ?? false"
      :is-assigned-quota-unlimited="isEntityUnlimited"
      :is-other-entities-quota-unlimited="isOtherEntitiesUnlimited"
      :height="20"
    />
  </q-td>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { ITableColumn } from "@/models/table.model";
import { ECustomCell } from "@/models/table.model";
import { getCustomCellConfig } from "@/utils/table-format.util/table-format.util";
import { OrgUnitQuotaDistributionChart } from "@/components/org-unit/v2/common/org-unit-quota-distribution-chart";
import type { ICustomCellQuotaDistributionColConfig } from "@/table-models/org-unit.table-model";
import { EQuotaEntity } from "@/models/resource.model";
import { DEFAULT_DEPARTMENT_NAME } from "@/models/department.model";

export interface IQuotaDistributionColRow {
  id: string;
  name: string;
  gpuQuota: number | null;
  serverGpuQuota?: number | null;
}

export interface Props {
  column: ITableColumn;
  row: IQuotaDistributionColRow;
}

const props = defineProps<Props>();

// Make config reactive
const config = computed(() => {
  return getCustomCellConfig(props.column, ECustomCell.QUOTA_DISTRIBUTION_COL) as
    | ICustomCellQuotaDistributionColConfig
    | undefined;
});

// Check if this entity has unlimited quota (-1)
// Only applies to "default" department
const isEntityUnlimited = computed(() => {
  if (config.value?.entityType !== EQuotaEntity.department) return false;
  return props.row.name === DEFAULT_DEPARTMENT_NAME && props.row.gpuQuota === -1;
});

// Check if other entities include unlimited quota
// True when: we're NOT the default department AND the default department is unlimited
const isOtherEntitiesUnlimited = computed(() => {
  if (config.value?.entityType !== EQuotaEntity.department) return false;
  // If this IS the default department, other entities can't be unlimited
  if (props.row.name === DEFAULT_DEPARTMENT_NAME) return false;
  // Otherwise, check if default department is unlimited (from config)
  return config.value?.isDefaultDepartmentUnlimited ?? false;
});

// Calculate other entities' quota (excluding this row's project)
// Formula: totalAssignedQuota - thisProjectQuota
const otherEntitiesQuota = computed(() => {
  const totalAssigned = config.value?.totalAssignedQuota ?? 0;
  // Treat -1 (unlimited) as 0 for calculation
  const thisEntityQuota = props.row.gpuQuota === -1 ? 0 : (props.row.gpuQuota ?? 0);

  return Math.max(0, totalAssigned - thisEntityQuota);
});
</script>

<style scoped lang="scss">
.org-unit-quota-distribution-col {
  width: 300px;
  padding: 8px 20px;
}
</style>
