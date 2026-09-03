<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <router-view />
</template>

<script lang="ts">
import { useTrainingStore } from "@/stores/training.store";
import { defineComponent } from "vue";

export default defineComponent({
  async created() {
    await useTrainingStore().loadActionPermissions();
  },
});
</script>
