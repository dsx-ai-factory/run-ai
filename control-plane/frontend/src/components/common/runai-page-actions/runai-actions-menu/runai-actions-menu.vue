<template>
  <runai-dropdown
    size="sm"
    aid="run-page-action-select-columns-btn"
    flat
    icon="ph-bold ph-dots-three-vertical"
    :custom-classes="`btn-icon-md ${largeIcon ? 'runai-icon-btn--large' : 'runai-icon-btn--default'}`"
    :label="hideLabel ? '' : 'More'"
    stack
    :round="round"
    :options="actionItems"
    @on-selected="(item) => item.action()"
  />
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

import { RunaiDropdown } from "@/components/common";
import type { ActionMenuItem } from "@/components/common/runai-page-actions/model";

export default defineComponent({
  name: "runai-actions-menu",
  components: { RunaiDropdown },
  props: {
    actionItems: {
      type: Array as PropType<ActionMenuItem[]>,
      required: false,
    },
    hideLabel: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    round: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    largeIcon: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
});
</script>
<style scoped lang="scss"></style>
