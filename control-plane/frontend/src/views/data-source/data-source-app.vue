<template>
  <router-view />
</template>

<script lang="ts">
import { useDataSourceStore } from "@/stores/data-source.store";
import { defineComponent } from "vue";

export default defineComponent({
  async created() {
    await useDataSourceStore().loadActionPermissions();
  },
});
</script>
