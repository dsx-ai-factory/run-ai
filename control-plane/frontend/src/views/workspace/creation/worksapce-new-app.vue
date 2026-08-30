<template>
  <router-view />
</template>

<script lang="ts">
import { defineComponent } from "vue";

// store
import { useWorkspaceStore } from "@/stores/workspace.store";

import { WORKSPACE_ROUTE_NAMES } from "@/router/workspace.routes/workspace.routes.names";

export default defineComponent({
  data() {
    return {
      workspaceStore: useWorkspaceStore(),
    };
  },
  created() {
    if (this.$router.options.history.state.back !== "/workloads") return;
    this.workspaceStore.resetCreateWorkspaceData();
    if (this.$route.name !== WORKSPACE_ROUTE_NAMES.WORKSPACE_NEW && this.$route.query.fromCopyId === undefined) {
      this.$router.replace({ name: WORKSPACE_ROUTE_NAMES.WORKSPACE_NEW });
    }
  },
});
</script>
