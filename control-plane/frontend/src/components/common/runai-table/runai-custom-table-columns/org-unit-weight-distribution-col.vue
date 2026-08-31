<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-td class="org-unit-weight-distribution-col">
    <org-unit-weight-distribution-chart
      v-if="displayChart"
      :entity-name="row.orgUnitName"
      :weight="row.weight ?? 0"
      :other-entities-weight="otherEntitiesWeight"
      :entity-type="entityType"
    />
    <div v-else class="text-center">-</div>
  </q-td>
</template>

<script setup lang="ts">
import { computed } from "vue";
import type { ITableColumn } from "@/models/table.model";
import { ECustomCell } from "@/models/table.model";
import { getCustomCellConfig } from "@/utils/table-format.util/table-format.util";
import { OrgUnitWeightDistributionChart } from "@/components/org-unit/v2/common/org-unit-weight-distribution-chart";
import type { ICustomCellWeightDistributionColConfig } from "@/table-models/org-unit.table-model";
import { EQuotaEntity } from "@/models/resource.model";

export interface IWeightDistributionColRow {
  id: string;
  orgUnitName: string;
  weight: number | null;
  serverWeight?: number | null;
}

defineOptions({ name: "OrgUnitWeightDistributionCol" });

const props = defineProps<{
  column: ITableColumn;
  row: IWeightDistributionColRow;
}>();

const config = computed(() => {
  return getCustomCellConfig(props.column, ECustomCell.WEIGHT_DISTRIBUTION_COL) as
    | ICustomCellWeightDistributionColConfig
    | undefined;
});

const otherEntitiesWeight = computed(() => {
  const totalWeightInRank = config.value?.totalWeightInRank ?? 0;
  const currentWeight = props.row.weight ?? 0;
  return Math.max(totalWeightInRank - currentWeight, 0);
});

const entityType = computed(() => config.value?.entityType ?? EQuotaEntity.project);
const displayChart = computed(() => {
  return props.row.weight !== null && props.row.weight !== -1 && props.row.weight !== undefined;
});
</script>

<style scoped lang="scss">
.org-unit-weight-distribution-col {
  width: 300px;
  padding: 8px 20px;
}
</style>
