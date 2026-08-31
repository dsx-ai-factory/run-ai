<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div></div>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { RouteLocationNormalized } from "vue-router";

// stores
import { useAuthStore } from "@/stores/auth.store";

// models
import { STORAGE_KEYS } from "@/models/storage.model";

// utils
import { storageUtil } from "@/utils/storage.util";

export default defineComponent({
  data() {
    return {
      authStore: useAuthStore(),
    };
  },
  beforeRouteEnter(to: RouteLocationNormalized, from: RouteLocationNormalized) {
    if (from.name) {
      storageUtil.save(STORAGE_KEYS.LOCATION, from.name);
    }
  },
  async created() {
    this.authStore.login();
  },
});
</script>
