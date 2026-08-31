<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-td :aid="column.name" class="description-col" ref="description-cell">
    <span ref="text-span">
      {{ text }}
    </span>
    <q-tooltip v-if="isLongText" max-width="300px" :style="{ wordBreak: 'break-word' }">{{ text }}</q-tooltip>
  </q-td>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// models
import type { ITableColumn } from "@/models/table.model";

// services
import { getFieldContent } from "@/utils/table-format.util";
import type { QTd } from "quasar";

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
  data() {
    return {
      isLongText: false,
    };
  },
  mounted() {
    // Calculating if the text is ellipsis by comparing the width of the text span and the cell
    const textSpan: HTMLSpanElement = this.$refs["text-span"] as HTMLSpanElement;
    const cell: HTMLTableCellElement = (this.$refs["description-cell"] as QTd)?.$el as HTMLTableCellElement;
    this.isLongText = textSpan.offsetWidth > cell.offsetWidth;
  },
  computed: {
    text(): string {
      return getFieldContent(this.column, this.row);
    },
  },
});
</script>

<style lang="scss" scoped>
.description-col {
  max-width: 300px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
