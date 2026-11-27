<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <router-view />
</template>

<script lang="ts">
import { defineComponent } from "vue";

import { TRAINING_ROUTE_NAMES } from "@/router/training.routes/training.routes.names";

export default defineComponent({
  created() {
    if (this.$router.options.history.state.back !== "/workloads") return;
    if (this.$route.name !== TRAINING_ROUTE_NAMES.TRAINING_NEW && this.$route.query.fromCopyId === undefined) {
      this.$router.replace({ name: TRAINING_ROUTE_NAMES.TRAINING_NEW });
    }
  },
});
</script>
