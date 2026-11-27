<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <router-view />
</template>

<script lang="ts">
import { defineComponent } from "vue";

import { WORKSPACE_ROUTE_NAMES } from "@/router/workspace.routes/workspace.routes.names";

export default defineComponent({
  created() {
    if (this.$router.options.history.state.back !== "/workloads") return;
    if (this.$route.name !== WORKSPACE_ROUTE_NAMES.WORKSPACE_NEW && this.$route.query.fromCopyId === undefined) {
      this.$router.replace({ name: WORKSPACE_ROUTE_NAMES.WORKSPACE_NEW });
    }
  },
});
</script>
