<template>
  <q-td :aid="column.name" class="html-col" v-dompurify-html="fieldContent" />
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// models
import type { ITableColumn } from "@/models/table.model";

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
    fieldContent(): string {
      return getFieldContent(this.column, this.row);
    },
  },
});
</script>
