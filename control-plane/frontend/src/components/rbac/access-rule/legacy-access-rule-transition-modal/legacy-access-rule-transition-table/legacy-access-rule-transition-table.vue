<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-table
    :rows="rows"
    :columns="columns"
    :loading="loading ?? false"
    :filter-by="filterBy"
    :selected="selected"
    :get-row-key="(row) => row.id"
    selection="single"
    hide-pagination
    :bordered="false"
    no-shadow
    sticky-header
    :style="{ maxHeight: '416px' }"
    @update:selected="emit('update:selected', $event as ILegacyRoleTransitionRow[])"
  />
</template>

<script setup lang="ts">
import { computed, ref } from "vue";
import type { RoleV1 } from "@/swagger-models/authorization-client/api";
import type { ILegacyRoleTransitionRow } from "@/models/legacy-access-rule-transition.model";
import type { IFilterBy } from "@/models/filter.model";
import { ECustomCell, type ICustomCellNewRoleColConfig } from "@/models/table.model";
import { ERoleStage, ROLE_STAGE_CONFIG } from "@/models/roles.model";
import { legacyAccessRuleTransitionColumns } from "@/table-models/legacy-access-rule-transition.table-model";
import RunaiTable from "@/components/common/runai-table/runai-table.vue";

defineOptions({ name: "legacy-access-rule-transition-table" });

const props = defineProps<{
  rows: ILegacyRoleTransitionRow[];
  availableRoles: RoleV1[];
  loading?: boolean;
  selected: ILegacyRoleTransitionRow[];
}>();

const emit = defineEmits<{
  "update:selected": [value: ILegacyRoleTransitionRow[]];
}>();

const filterBy = ref<IFilterBy>({} as IFilterBy);

const roleOptions = computed(() =>
  props.availableRoles
    .filter((r) => {
      const key = Object.keys(ROLE_STAGE_CONFIG).find((k) => k.toLowerCase() === r.name.toLowerCase());
      return ROLE_STAGE_CONFIG[key ?? ""]?.stage !== ERoleStage.Legacy;
    })
    .map((r) => r.name),
);

const columns = computed(() =>
  legacyAccessRuleTransitionColumns.map((col) => {
    if (col.name !== "newRole") return col;
    return {
      ...col,
      customCellConfig: {
        [ECustomCell.NEW_ROLE_COL]: {
          roleOptions: roleOptions.value,
          getReviewRoleId: (name: string) => props.availableRoles.find((r) => r.name === name)?.id,
        } satisfies ICustomCellNewRoleColConfig,
      },
    };
  }),
);
</script>
