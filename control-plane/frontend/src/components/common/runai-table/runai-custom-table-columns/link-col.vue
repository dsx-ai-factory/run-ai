<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-td>
    <q-btn
      v-if="!isTextOnly"
      flat
      :ripple="false"
      color="primary"
      class="btn-link no-padding"
      :label="displayText"
      @click.stop="$emit('custom-cell-event', row)"
    >
      <q-tooltip v-if="tooltipText" :anchor="tooltipAnchor" :offset="tooltipOffset">{{ tooltipText }}</q-tooltip>
    </q-btn>
    <span v-else>{{ displayText }}</span>
  </q-td>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// models
import { ECustomCell, type ICustomCellLinkColConfig, type ITableColumn } from "@/models/table.model";

// services
import { getFieldContent } from "@/utils/table-format.util";
import { getCustomCellConfig } from "@/utils/table-format.util/table-format.util";
import type { TQuasarTooltipAnchor } from "@/models/global.model";

export default defineComponent({
  emits: ["custom-cell-event"],
  props: {
    row: {
      type: Object as PropType<any>, // eslint-disable-line @typescript-eslint/no-explicit-any
      required: true,
    },
    column: {
      type: Object as PropType<ITableColumn>,
      required: true,
    },
  },
  computed: {
    displayText(): string {
      return getFieldContent(this.column, this.row);
    },
    isTextOnly(): boolean {
      return !!(getCustomCellConfig(this.column, ECustomCell.LINK_COL) as ICustomCellLinkColConfig)?.isTextOnly;
    },
    tooltipText(): string | undefined {
      return (getCustomCellConfig(this.column, ECustomCell.LINK_COL) as ICustomCellLinkColConfig)?.tooltipText;
    },
    tooltipAnchor(): TQuasarTooltipAnchor {
      return (
        ((getCustomCellConfig(this.column, ECustomCell.LINK_COL) as ICustomCellLinkColConfig)
          ?.tooltipAnchor as TQuasarTooltipAnchor) || "center right"
      );
    },
    tooltipOffset(): [number, number] {
      return (
        (getCustomCellConfig(this.column, ECustomCell.LINK_COL) as ICustomCellLinkColConfig)?.tooltipOffset || [120, 0]
      );
    },
  },
});
</script>
