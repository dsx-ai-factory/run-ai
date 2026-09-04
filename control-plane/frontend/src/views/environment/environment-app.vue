<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

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
