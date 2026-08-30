<template>
  <q-td>
    <q-btn
      flat
      :ripple="false"
      color="primary"
      class="btn-link no-padding"
      :label="displayText"
      :class="{ active: isActive }"
      @click.stop="
        $emit('custom-cell-event', row);
        isActive = true;
      "
    >
      <runai-org-tree-with-pop-over
        v-if="isActive"
        @hide="isActive = false"
        :selected="[{ type: row.scopeType, id: row.scopeId }]"
        readonly
      />
    </q-btn>
  </q-td>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// models
import type { ITableColumn } from "@/models/table.model";

// services
import { getFieldContent } from "@/utils/table-format.util";
import { RunaiOrgTreeWithPopOver } from "@/components/common/runai-org-tree/with-popover";
import type { AccessRule } from "@/swagger-models/authorization-client";

export default defineComponent({
  components: { RunaiOrgTreeWithPopOver },
  emits: ["custom-cell-event"],
  props: {
    row: {
      type: Object as PropType<AccessRule>, // eslint-disable-line @typescript-eslint/no-explicit-any
      required: true,
    },
    column: {
      type: Object as PropType<ITableColumn>,
      required: true,
    },
  },
  data() {
    return {
      isActive: false,
    };
  },
  computed: {
    displayText(): string {
      return getFieldContent(this.column, this.row);
    },
  },
});
</script>
<style scoped lang="scss">
.active {
  text-decoration: underline;
}
</style>
