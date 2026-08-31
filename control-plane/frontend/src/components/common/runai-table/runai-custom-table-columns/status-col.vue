<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-td :style="statusOptions.colStyle">
    <runai-status :options="statusOptions" @tooltip-clicked="handleToolTipClick" />
  </q-td>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";

// utils
import { getFieldContent } from "@/utils/table-format.util";
// models
import type { IStatusColOptions } from "@/models/table.model";
import type { ITableColumn } from "@/models/table.model";
import { RunaiStatus } from "@/components/common/runai-status/";

export default defineComponent({
  components: { RunaiStatus },
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
    statusOptions(): IStatusColOptions {
      return getFieldContent(this.column, this.row);
    },
  },
  methods: {
    handleToolTipClick(event: Event) {
      const target = event.target as HTMLInputElement;
      // emit event only if the specific element was clicked (id starts with "status-col-emit")
      if (target.id?.startsWith("status-col-emit")) {
        this.$emit("custom-cell-event");
      }
    },
  },
});
</script>
