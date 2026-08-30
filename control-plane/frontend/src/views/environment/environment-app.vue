<template>
  <router-view />
</template>

<script lang="ts">
import { useEnvironmentStore } from "@/stores/environment.store";
import { defineComponent } from "vue";

export default defineComponent({
  async created() {
    await useEnvironmentStore().loadActionPermissions();
  },
});
</script>
