<template>
  <q-td class="audit-status-col" :class="getStatusClass">
    {{ status }}
  </q-td>
</template>

<script lang="ts">
import type { ITableColumn } from "@/models/table.model";
import { getFieldContent } from "@/utils/table-format.util";
import { defineComponent, type PropType } from "vue";
import type { AuditLog } from "@/swagger-models/audit-service-client";
import { EEventStatus } from "@/models/audit.model";

const STATUS_CLASSES = {
  SUCCESS: "success",
  FAILED: "failed",
};

export default defineComponent({
  props: {
    row: {
      type: Object as PropType<AuditLog>,
      required: true,
    },
    column: {
      type: Object as PropType<ITableColumn>,
      required: true,
    },
  },
  computed: {
    status(): string {
      return getFieldContent(this.column, this.row);
    },
    getStatusClass(): string {
      return this.status === EEventStatus.Failed ? STATUS_CLASSES.FAILED : STATUS_CLASSES.SUCCESS;
    },
  },
});
</script>
<style lang="scss" scoped>
.success {
  color: $success;
}

.failed {
  color: $negative;
}
</style>
