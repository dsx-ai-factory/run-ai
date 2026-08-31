<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-td :aid="column.name" class="default-col" :class="column.classes" :style="column.style">
    <div class="row items-center no-wrap">
      <span>{{ text }}</span>
      <q-icon
        v-if="iconConfig"
        :name="iconConfig.name"
        :color="iconConfig.color"
        :size="iconConfig.size || '16px'"
        class="q-ml-auto"
      >
        <q-tooltip v-if="iconConfig.tooltip">{{ iconConfig.tooltip }}</q-tooltip>
      </q-icon>
    </div>
  </q-td>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// models
import type { ITableColumn, ICellIconConfig } from "@/models/table.model";

// services
import { getFieldContent } from "@/utils/table-format.util";

export default defineComponent({
  props: {
    column: {
      type: Object as PropType<ITableColumn>,
      required: true,
    },
    row: {
      type: Object as PropType<any>, // eslint-disable-line @typescript-eslint/no-explicit-any
      required: true,
    },
  },
  computed: {
    text(): string {
      return getFieldContent(this.column, this.row);
    },
    iconConfig(): ICellIconConfig | undefined {
      if (this.column.cellIcon) {
        return this.column.cellIcon(this.row);
      }
      return undefined;
    },
  },
});
</script>
