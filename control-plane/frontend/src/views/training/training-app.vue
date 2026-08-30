<template>
  <router-view />
</template>

<script lang="ts">
import { useTrainingStore } from "@/stores/training.store";
import { defineComponent } from "vue";

export default defineComponent({
  async created() {
    await useTrainingStore().loadActionPermissions();
  },
});
</script>
