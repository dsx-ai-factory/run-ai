<template>
  <section class="inference-app"><router-view /></section>
</template>

<script lang="ts">
import { defineComponent } from "vue";

// store
import { useInferenceStore } from "@/stores/inference.store";

import { INFERENCE_ROUTE_NAMES } from "@/router/inference.routes/inference.routes.names";

export default defineComponent({
  data() {
    return {
      inferenceStore: useInferenceStore(),
    };
  },
  created() {
    const fromRoute = this.$router.options.history.state.back?.toString() || "";
    if (!fromRoute.includes("/workloads")) {
      return;
    }
    this.inferenceStore.resetCreateInferenceData();
    if (
      this.$route.name !== INFERENCE_ROUTE_NAMES.INFERENCE_NEW &&
      this.$route.query.fromCopyId === undefined &&
      this.$route.query.updateWorkloadId === undefined
    ) {
      this.$router.replace({ name: INFERENCE_ROUTE_NAMES.INFERENCE_NEW });
    }
  },
});
</script>

<style lang="scss" scoped></style>
