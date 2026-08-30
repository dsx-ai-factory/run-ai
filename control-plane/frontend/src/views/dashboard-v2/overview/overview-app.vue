<template>
  <overview-index />
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { OverviewIndex } from "@/views/dashboard-v2/overview/index";

export default defineComponent({
  name: "overview-app",
  components: { OverviewIndex },
});
</script>

<style scoped lang="scss"></style>
