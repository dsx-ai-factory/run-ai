<template>
  <router-view />
</template>

<script lang="ts">
import { defineComponent } from "vue";

// store
import { useTrainingStore } from "@/stores/training.store";

import { TRAINING_ROUTE_NAMES } from "@/router/training.routes/training.routes.names";

export default defineComponent({
  data() {
    return {
      trainingStore: useTrainingStore(),
    };
  },
  created() {
    if (this.$router.options.history.state.back !== "/workloads") return;
    this.trainingStore.resetCreateTrainingData();
    if (this.$route.name !== TRAINING_ROUTE_NAMES.TRAINING_NEW && this.$route.query.fromCopyId === undefined) {
      this.$router.replace({ name: TRAINING_ROUTE_NAMES.TRAINING_NEW });
    }
  },
});
</script>
