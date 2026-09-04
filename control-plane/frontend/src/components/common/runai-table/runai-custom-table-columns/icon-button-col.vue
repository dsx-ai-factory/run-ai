<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-td :aid="column.name" class="icon-button-col">
    <q-btn
      flat
      round
      dense
      padding="sm"
      :icon="iconConfig.icon"
      :color="buttonColor"
      :size="iconConfig.size || 'sm'"
      class="runai-icon-btn--xsmall"
      :disable="isDisabled"
      :aria-label="tooltipText || 'Action'"
      @click.stop="handleClick"
    >
      <q-tooltip v-if="tooltipText" class="tool-tip-dark">{{ tooltipText }}</q-tooltip>
    </q-btn>
  </q-td>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// models
import { ECustomCell, type ICustomCellIconButtonColConfig, type ITableColumn } from "@/models/table.model";

// utils
import { getCustomCellConfig } from "@/utils/table-format.util/table-format.util";

export default defineComponent({
  emits: ["custom-cell-event"],
  props: {
    row: {
      type: Object as PropType<unknown>,
      required: true,
    },
    column: {
      type: Object as PropType<ITableColumn>,
      required: true,
    },
  },
  computed: {
    iconConfig(): ICustomCellIconButtonColConfig {
      const config = getCustomCellConfig(this.column, ECustomCell.ICON_BUTTON_COL) as
        | ICustomCellIconButtonColConfig
        | undefined;
      return (
        config ?? {
          icon: "ph-bold ph-question",
        }
      );
    },
    isDisabled(): boolean {
      if (this.iconConfig.isDisabled) {
        return this.iconConfig.isDisabled(this.row);
      }
      return false;
    },
    tooltipText(): string {
      if (this.iconConfig.getTooltip) {
        return this.iconConfig.getTooltip(this.row);
      }
      return "";
    },
    buttonColor(): string {
      return this.iconConfig.color || "b-sub";
    },
  },
  methods: {
    handleClick(): void {
      this.$emit("custom-cell-event", this.row);
    },
  },
});
</script>

<style lang="scss" scoped>
.icon-button-col {
  text-align: center;
}
</style>
