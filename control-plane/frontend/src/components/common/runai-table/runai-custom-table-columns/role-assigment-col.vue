<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-td @click.stop>
    <q-btn
      flat
      :ripple="false"
      color="primary"
      class="btn-link no-padding"
      :label="displayTextTruncated"
      @click.stop="$emit('custom-cell-event', row)"
      v-if="list.length > 0"
    />
    <div v-else>{{ displayTextTruncated }}</div>
  </q-td>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// models
import type { ITableColumn } from "@/models/table.model";
import type { AccessRule } from "@/swagger-models/authorization-client";
import { subjectTypeOptionsWithSSO } from "@/models/access-rule.model";

// services
import { getFieldContent } from "@/utils/table-format.util";

const subjectTypeOptionsMap: Map<string, string> = new Map(
  subjectTypeOptionsWithSSO.map((option) => [option.value as string, option.label as string]),
);
// Map deprecated "app" to "Service account" for backward compatibility
subjectTypeOptionsMap.set("app", "Service account");

export default defineComponent({
  emits: ["custom-cell-event"],
  props: {
    row: {
      type: Object as PropType<AccessRule>,
      required: true,
    },
    column: {
      type: Object as PropType<ITableColumn>,
      required: true,
    },
    maxLength: {
      type: Number,
      default: 30, // Maximum length of the displayed text (adjust as needed)
    },
  },
  computed: {
    list(): Array<string> {
      return getFieldContent(this.column, this.row) || [];
    },
    displayText(): string {
      const count: Record<string, number> = {};

      for (const item of this.list) {
        count[item] = (count[item] || 0) + 1;
      }

      const result: string[] = [];

      for (const item in count) {
        const formattedItem = subjectTypeOptionsMap.get(item) || item;
        const plural = count[item] > 1 ? "s" : "";
        result.push(`${count[item]} ${formattedItem}${plural}`);
      }

      return result.join(" + ");
    },
    displayTextTruncated(): string {
      if (!this.displayText) return "-";
      if (this.displayText.length <= this.maxLength) {
        return this.displayText;
      }
      return this.displayText.slice(0, this.maxLength) + "...";
    },
  },
});
</script>
