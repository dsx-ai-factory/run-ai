<template>
  <q-td v-if="progressColOptions.value !== null" :class="column.classes">
    <div class="progress-col">
      <div class="display-value-wrapper">{{ displayValue }}</div>
      <q-linear-progress
        size="13px"
        :value="progressPercentage"
        :color="progressColOptions.color"
        :aria-label="`${column.label}: ${displayValue}`"
      />
    </div>
  </q-td>
  <q-td v-else>-</q-td>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
import type { IProgressColOptions, ITableColumn } from "@/models/table.model";
import { getFieldContent } from "@/utils/table-format.util";
export default defineComponent({
  name: "progress-col",
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
    progressColOptions(): IProgressColOptions {
      return getFieldContent(this.column, this.row);
    },
    color(): string {
      return this.progressColOptions.color || "soft-blue";
    },
    displayValue(): string {
      if (this.progressColOptions.value === null) {
        return "";
      }
      return `${this.progressColOptions.value}%`;
    },
    progressPercentage(): number {
      if (!this.progressColOptions.value) {
        return 0;
      }
      return this.progressColOptions.value / 100;
    },
  },
});
</script>

<style scoped lang="scss">
.progress-col {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 5px;
}
</style>
