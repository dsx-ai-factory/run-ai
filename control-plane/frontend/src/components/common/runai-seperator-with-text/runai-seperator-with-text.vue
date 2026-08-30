<template>
  <div class="seperator-with-text row fit no-wrap q-gutter-x-sm justify-start items-center">
    <q-separator class="col" />
    <div class="col-grow">{{ text }}</div>
    <q-separator inset class="col" />
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from "vue";
export default defineComponent({
  props: {
    text: {
      type: String as PropType<string>,
      required: true,
    },
  },
});
</script>

<style scoped lang="scss"></style>
