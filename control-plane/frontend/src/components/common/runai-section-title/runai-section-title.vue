<template>
  <div class="runai-section-title q-mb-sm row items-center justify-between" :class="{ 'no-margin': flat }">
    <div class="row items-center">
      <span v-if="title" class="q-mr-xs" v-text="title"></span>
      <slot v-else name="title-content" />
      <runai-tooltip v-if="tooltipText" :tooltip-text="tooltipText" :width="width" :tooltip-position="tooltipPosition" />
    </div>
    <slot name="title-right-content" />
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { RunaiTooltip } from "@/components/common/runai-tooltip";

export default defineComponent({
  name: "runai-section-title",
  components: {
    RunaiTooltip,
  },
  props: {
    title: {
      type: String as PropType<string>,
      required: false, // if no title is provided, the slot will be used
    },
    tooltipText: {
      type: String as PropType<string>,
      required: false,
    },
    width: {
      type: String as PropType<string>,
      required: false,
      default: "300px",
    },
    tooltipPosition: {
      type: String as PropType<string>,
      required: false,
      default: "right",
    },
    flat: {
      type: Boolean as PropType<boolean>,
      required: false,
      default: false,
    },
  },
});
</script>
