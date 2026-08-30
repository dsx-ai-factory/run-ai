<template>
  <router-view />
</template>

<script lang="ts">
import { useWorkloadTemplateStore } from "@/stores/workload-template.store";
import { defineComponent } from "vue";

export default defineComponent({
  async created() {
    await useWorkloadTemplateStore().loadActionPermissions();
  },
});
</script>
