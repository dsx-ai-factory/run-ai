<template>
  <section class="runai-brand-logo">
    <img :height="height" :src="img" alt="logo" />
  </section>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";

export default defineComponent({
  props: {
    height: {
      type: String as PropType<string>,
      default: "34",
    },
    img: {
      type: String as PropType<string>,
      required: true,
    },
  },
});
</script>
<style lang="scss" scoped>
.runai-brand-logo {
  display: flex;
  align-items: center;
  justify-content: center;
}
.svg-large {
  height: 50px;
}
</style>
