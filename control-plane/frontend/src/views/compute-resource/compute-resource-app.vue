<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

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
