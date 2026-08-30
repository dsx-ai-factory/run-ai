<template>
  <q-td>
    <div class="row items-center no-wrap">
      {{ formattedText }}
      <q-tooltip v-if="showTooltip">{{ text }}</q-tooltip>

      <template v-if="iconConfig">
        <q-space />
        <q-icon :name="iconConfig.name" :color="iconConfig.color" :size="iconConfig.size || '16px'">
          <q-tooltip v-if="iconConfig.tooltip">{{ iconConfig.tooltip }}</q-tooltip>
        </q-icon>
      </template>
    </div>
  </q-td>
</template>
<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";
import type { ICellIconConfig, ITableColumn } from "@/models/table.model";
import { MAX_STRING_LENGTH_BEFORE_TRUNCATION } from "@/models/global.model";
import { getFieldContent } from "@/utils/table-format.util";
import { formatArrayStrings } from "@/utils/table-format.util";

export default defineComponent({
  name: "strings-list-col",
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
    content(): Array<string> {
      return getFieldContent(this.column, this.row);
    },
    text(): string | null {
      if (!this.content) return null;
      return this.content.join(", ");
    },
    showTooltip(): boolean {
      if (!this.content || !this.text) return false;
      // Show tooltip if the joined text (with separators) exceeds the truncation limit
      return this.text.length > MAX_STRING_LENGTH_BEFORE_TRUNCATION;
    },
    formattedText(): string {
      return formatArrayStrings(this.content) || "-";
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

<style scoped lang="scss"></style>
