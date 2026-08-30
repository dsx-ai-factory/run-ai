<template>
  <node-index />
</template>
<script lang="ts">
import { defineComponent } from "vue";
//cmps
import { NodeIndex } from "@/views/node/";
export default defineComponent({
  name: "node-app",
  components: { NodeIndex },
});
</script>

<style scoped lang="scss"></style>
