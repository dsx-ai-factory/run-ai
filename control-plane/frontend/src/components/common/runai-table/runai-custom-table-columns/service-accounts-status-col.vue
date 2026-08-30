<template>
  <q-td>
    <span :class="cmpClass">{{ status }}</span>
  </q-td>
</template>

<script lang="ts">
import type { ITableColumn } from "@/models/table.model";
import { defineComponent, type PropType } from "vue";
import type { ServiceAccount } from "@/swagger-models/identity-manager-client";

export default defineComponent({
  props: {
    row: {
      type: Object as PropType<ServiceAccount>,
      required: true,
    },
    column: {
      type: Object as PropType<ITableColumn>,
      required: true,
    },
  },
  computed: {
    status(): string {
      return this.row.enabled ? "Active" : "Revoked";
    },
    cmpClass(): string {
      return this.status === "Active" ? "active-clr" : "";
    },
  },
});
</script>
