<template>
  <router-view />
</template>

<script lang="ts">
import { useDataVolumeStore } from "@/stores/data-volume.store";
import { defineComponent } from "vue";

export default defineComponent({
  name: "data-volume-app",
  async created() {
    await useDataVolumeStore().loadActionPermissions();
  },
});
</script>
