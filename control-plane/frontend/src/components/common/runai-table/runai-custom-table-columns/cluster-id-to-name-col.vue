<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-td :aid="column.name" class="cluster-col">
    {{ clusterName }}
  </q-td>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// models
import type { ITableColumn } from "@/models/table.model";
import { useClusterStore } from "@/stores/cluster.store";
import { getFieldContent } from "@/utils/table-format.util";
import { tableDisplayFormatters } from "@/utils/table-display-formatters.util";

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
      clusterStore: useClusterStore(),
    };
  },
  computed: {
    clusterUuid(): string {
      return getFieldContent(this.column, this.row);
    },
    clusterName(): string {
      return tableDisplayFormatters.getClusterName(
        this.clusterUuid,
        this.clusterStore.clusterList,
        this.column,
        this.row,
      );
    },
  },
});
</script>
