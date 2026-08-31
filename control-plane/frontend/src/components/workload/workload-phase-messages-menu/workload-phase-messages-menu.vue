<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-menu
    @hide="$emit('hide')"
    model-value
    class="workload-phase-messages-menu"
    aid="workload-phase-messages-menu"
    anchor="bottom middle"
    self="top middle"
  >
    <div class="q-pa-md">The workload failed to run, due to the following reasons:</div>
    <q-separator />
    <q-table table-class="runai-table" :columns="columns" :rows="messages" hide-bottom flat />
  </q-menu>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { ITableColumn } from "@/models/table.model";
import type { Field } from "@/models/workload-phase-messages.model";

export default defineComponent({
  name: "workload-phase-messages-menu",
  emits: ["hide"],
  props: {
    columns: {
      type: Array as PropType<ITableColumn[]>,
      required: true,
    },
    messages: {
      type: Array as PropType<Field[]>,
      required: true,
    },
  },
});
</script>
<style lang="scss">
.workload-phase-messages-menu {
  .q-table {
    td {
      max-width: 500px;
      text-wrap: wrap;
    }
  }
}
</style>
