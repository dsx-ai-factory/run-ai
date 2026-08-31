<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-td :aid="column.name" class="user-display-name-col" :class="column.classes" :style="column.style">
    <runai-user-display-name v-if="shouldRenderDisplayName" :username="username" />
    <span v-else>{{ fallbackText }}</span>
  </q-td>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { RunaiUserDisplayName } from "@/components/common/runai-user-display-name";
import { ECustomCell, type ICustomCellUserDisplayNameColConfig, type ITableColumn } from "@/models/table.model";
import { getCustomCellConfig, getField } from "@/utils/table-format.util";

const props = defineProps<{
  /* eslint-disable-next-line @typescript-eslint/no-explicit-any */
  row: any;
  column: ITableColumn;
}>();

const username = computed((): string => {
  const val = getField(props.column, props.row);
  return typeof val === "string" ? val : "";
});

const fallbackText = computed((): string => username.value || "-");

const shouldRenderDisplayName = computed((): boolean => {
  if (!username.value) return false;
  const config = getCustomCellConfig(props.column, ECustomCell.USER_DISPLAY_NAME_COL) as
    | ICustomCellUserDisplayNameColConfig
    | undefined;
  if (config?.skipWhen && config.skipWhen(props.row)) return false;
  return true;
});
</script>
