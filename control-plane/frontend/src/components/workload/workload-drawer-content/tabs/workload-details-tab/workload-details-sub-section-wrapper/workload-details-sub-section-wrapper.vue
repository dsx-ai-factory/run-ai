<template>
  <div class="sub-section-wrapper q-pt-sm">
    <div class="sub-section-wrapper-header flex items-center">
      <div class="text-weight-bold text-subtitle1 q-mr-xs">{{ title }}</div>
      <runai-tooltip v-if="tooltipText" :tooltip-text="tooltipText" />
    </div>
    <div v-if="!isEmpty" class="sub-section-wrapper-body q-pt-sm">
      <slot></slot>
    </div>
  </div>
  <div v-if="isEmpty" class="sub-section-wrapper-empty">
    {{ emptyMessage }}
  </div>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";
import { RunaiTooltip } from "@/components/common/runai-tooltip/";

export default defineComponent({
  name: "workload-details-sub-section-wrapper",
  components: { RunaiTooltip },
  props: {
    title: {
      type: String as PropType<string>,
      required: true,
    },
    isEmpty: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    tooltipText: {
      type: String as PropType<string>,
      required: false,
    },
    emptyMessage: {
      type: String as PropType<string>,
      required: false,
    },
  },
});
</script>

<style scoped lang="scss">
.sub-section-wrapper-empty {
  height: 120px;
  display: flex;
  justify-content: center;
  flex-direction: column;
  align-items: center;
}
.sub-section-wrapper-empty,
.sub-section-wrapper-body {
  font-size: 12px;
}
</style>
