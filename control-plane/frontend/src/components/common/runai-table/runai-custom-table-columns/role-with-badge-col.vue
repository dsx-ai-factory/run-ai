<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-td>
    <div class="row items-center no-wrap q-gutter-sm">
      <q-btn
        v-if="!isTextOnly"
        flat
        :ripple="false"
        color="primary"
        class="btn-link no-padding"
        :label="roleName"
        @click.stop="emit('custom-cell-event', props.row)"
      />
      <span v-else>{{ roleName }}</span>
      <role-stage-badge :role-name="roleName" :hidden-stages="hiddenStages" />
    </div>
  </q-td>
</template>

<script setup lang="ts">
import { computed } from "vue";

// models
import type { ITableColumn } from "@/models/table.model";
import { ECustomCell, type ICustomCellRoleWithBadgeColConfig } from "@/models/table.model";

// services
import { getFieldContent, getCustomCellConfig } from "@/utils/table-format.util/table-format.util";

// components
import { RoleStageBadge } from "@/components/common/role-stage-badge";

export interface Props {
  row: any; // eslint-disable-line @typescript-eslint/no-explicit-any
  column: ITableColumn;
}

const props = defineProps<Props>();

const emit = defineEmits<{
  (e: "custom-cell-event", row: any): void; // eslint-disable-line @typescript-eslint/no-explicit-any
}>();

const config = computed(
  () =>
    getCustomCellConfig(props.column, ECustomCell.ROLE_WITH_BADGE_COL) as ICustomCellRoleWithBadgeColConfig | undefined,
);

const roleName = computed(() => getFieldContent(props.column, props.row));

const isTextOnly = computed(() => !!config.value?.isTextOnly);

const hiddenStages = computed(() => config.value?.hiddenStages);
</script>
