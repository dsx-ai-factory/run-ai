<template>
  <q-td>
    <q-btn
      flat
      :ripple="false"
      color="primary"
      class="btn-link no-padding"
      :label="displayText"
      @click.stop="$emit('custom-cell-event', row)"
      v-if="list.length > 0 && !isString"
    />
    <div v-else>{{ displayText }}</div>
  </q-td>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// models
import type { ITableColumn } from "@/models/table.model";

// services
import { getFieldContent } from "@/utils/table-format.util";
import { tableDisplayFormatters } from "@/utils/table-display-formatters.util";

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
    list(): Array<string> | string {
      return getFieldContent(this.column, this.row) || [];
    },
    displayText(): string {
      return tableDisplayFormatters.getListPreview(this.list);
    },
    isString(): boolean {
      return typeof this.list === "string";
    },
  },
});
</script>
