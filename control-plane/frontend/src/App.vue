<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <runai-app-loader v-if="!appLoaded" />
  <router-view class="app-main" v-if="preAppDataLoaded" />
</template>

<script lang="ts">
import { defineComponent } from "vue";

// stores
import { useAppStore } from "./stores/app.store";
import { useAuthStore } from "./stores/auth.store";

// components
import { RunaiAppLoader } from "./components/common/runai-app-loader";

// models
import { ERROR_ROUTE_NAMES } from "@/router/error.routes/error.routes.names";

// utils
import { storageUtil } from "@/utils/storage.util";
import { HAD_SESSION } from "@/common/storage.constant";

export default defineComponent({
  components: {
    RunaiAppLoader,
  },
  data() {
    return {
      appStore: useAppStore(),
      authStore: useAuthStore(),
      preAppDataLoaded: false as boolean,
    };
  },
  async created() {
    this.appStore.setAppLoaded(false);

    try {
      this.preAppDataLoaded = false;
      await this.loadAppConfiguration();
      this.preAppDataLoaded = true;
    } catch (error: unknown) {
      console.error("Something went wrong during the app's initial loading. ", error);
      await this.$router.replace({ name: ERROR_ROUTE_NAMES.BAD_GATEWAY_ERROR });
    } finally {
      this.preAppDataLoaded = true;
    }
  },
  computed: {
    appLoaded(): boolean {
      return this.appStore.appLoaded;
    },
  },
  methods: {
    redirectToCanonicalUrlIfNeeded(): void {
      const { isMultiTenant, redirectUrl } = this.appStore.config;
      if (!isMultiTenant || !redirectUrl) return;

      const canonicalOrigin = new URL(redirectUrl).origin;
      if (canonicalOrigin === window.location.origin) return;

      const canonicalUrl = redirectUrl + window.location.pathname + window.location.search + window.location.hash;
      window.location.replace(canonicalUrl);
    },

    async loadAppConfiguration(): Promise<void> {
      if (!this.authStore.accessToken && storageUtil.get(HAD_SESSION)) {
        try {
          await this.authStore.updateAccessToken();
        } catch (error) {
          // Continue - authenticated-page will redirect to login
        }
      }

      await this.appStore.loadAppData();
      this.redirectToCanonicalUrlIfNeeded();
      if (this.appStore.tenantName) {
        await this.authStore.loadAuthProviderConfig(this.appStore.config.authHost);
      }
    },
  },
});
</script>
<style scoped lang="scss"></style>
