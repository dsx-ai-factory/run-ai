<template>
  <router-view />
</template>

<script lang="ts">
import { useWorkspaceStore } from "@/stores/workspace.store";
import { defineComponent } from "vue";

export default defineComponent({
  async created() {
    await useWorkspaceStore().loadActionPermissions();
  },
});
</script>
