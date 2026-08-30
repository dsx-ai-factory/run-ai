<template>
  <section :class="['runai-drawer-nodata', wrapperPaddingClass]">
    <div v-for="(message, index) of messages" class="row justify-center" :key="index">{{ message }}</div>
    <slot name="additional-content" />
    <div class="row justify-center q-pt-xl">
      <q-icon :name="iconName" class="empty-content-icon" :class="iconClass" />
    </div>
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { PropType } from "vue";

export default defineComponent({
  props: {
    messages: {
      type: Array as PropType<Array<string>>,
      required: true,
    },
    iconName: {
      type: String as PropType<string>,
      required: true,
    },
    wrapperPaddingClass: {
      type: String as PropType<string>,
      default: "q-pt-xl",
    },
    iconClass: {
      type: String as PropType<string>,
      default: "",
    },
  },
});
</script>

<style lang="scss" scoped>
.runai-drawer-nodata {
  height: 100%;
  .empty-content-icon {
    font-size: var(--icon-size-empty-state, 44px);
  }
}
</style>
