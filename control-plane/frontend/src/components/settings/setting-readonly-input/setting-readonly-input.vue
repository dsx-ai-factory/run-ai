<template>
  <div :class="{ 'readonly-wrapper': isReadOnly }">
    <slot :readonly="isReadOnly"></slot>
    <q-tooltip v-if="isReadOnly && displayTooltip">This can't be modified based on your permissions</q-tooltip>
  </div>
</template>

<script lang="ts">
import type { PropType } from "vue";
import { defineComponent } from "vue";
import { usePermissionStore } from "@/stores/permissions.store";
import { ResourceType } from "@/swagger-models/authorization-client";

export default defineComponent({
  name: "setting-readonly-input",
  props: {
    resourceType: {
      type: String as PropType<keyof typeof ResourceType>,
      default: "Settings",
    },
    readOnly: {
      type: Boolean as PropType<boolean>,
      default: undefined,
    },
    displayTooltip: {
      type: Boolean as PropType<boolean>,
      default: true,
    },
  },
  computed: {
    isReadOnly(): boolean {
      if (this.readOnly !== undefined) {
        return this.readOnly;
      }

      return usePermissionStore().isReadOnly(ResourceType[this.resourceType]);
    },
  },
});
</script>

<style lang="scss" scoped>
.readonly-wrapper {
  cursor: not-allowed;

  :deep(*) {
    pointer-events: none;
  }
}
</style>
