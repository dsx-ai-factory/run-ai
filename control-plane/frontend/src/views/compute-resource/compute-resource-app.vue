<template>
  <router-view />
</template>

<script lang="ts">
import { defineComponent } from "vue";
import { useComputeResourceStore } from "@/stores/compute-resource.store";

export default defineComponent({
  async created() {
    await useComputeResourceStore().loadActionPermissions();
  },
});
</script>
