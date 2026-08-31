<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-expansion-item
    class="notifications-section settings-container"
    label="Cluster authentication"
    subheader="Enable workload submission via Kubectl"
  >
    <section class="researcher-authentication">
      <div class="q-pb-md">Use the following to enable users' workload submissions via Kubectl</div>
      <div>
        <server-config v-if="serverConfigInfo" :server-config="serverConfigInfo" />
      </div>
    </section>
  </runai-expansion-item>
</template>

<script lang="ts">
import { defineComponent } from "vue";
// Stores
import { useAuthStore } from "@/stores/auth.store";
// Models
import type { ITenantServerConfigInfo } from "@/models/tenant.model";
// Components
import { ServerConfig } from "./server-config";
import { RunaiExpansionItem } from "@/components/common/runai-expansion-item";

export default defineComponent({
  name: "researcher-authentication",
  components: {
    RunaiExpansionItem,
    ServerConfig,
  },
  data() {
    return {
      authStore: useAuthStore(),
      serverConfigInfo: null as ITenantServerConfigInfo | null,
      loading: false as boolean,
    };
  },
  async created(): Promise<void> {
    this.loading = true;
    try {
      const configInfo = await this.authStore.loadTenantConfigInfo();
      this.serverConfigInfo = configInfo.serverConfigInfo;
    } catch (err: unknown) {
      console.error(err);
    } finally {
      this.loading = false;
    }
  },
});
</script>
