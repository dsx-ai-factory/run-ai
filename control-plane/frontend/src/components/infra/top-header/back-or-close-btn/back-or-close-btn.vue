<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-btn
    :disable="disable"
    @click="onLeave"
    :icon="btnIcon"
    class="runai-icon-btn--default"
    flat
    round
    aid="header-leave-page-button"
    aria-label="Leave this page"
  >
    <q-tooltip v-if="!disable" class="white" anchor="center right" self="center left" :offset="[10, 10]">
      Leave this page
    </q-tooltip>
  </q-btn>
</template>
<script lang="ts">
import { defineComponent } from "vue";

// stores
import { useAppStore } from "@/stores/app.store";
import { useAuthStore } from "@/stores/auth.store";

// services
import { requestToLeave } from "@/services/infra/router.service/router.service";

export default defineComponent({
  name: "back-or-close-btn",
  props: {
    disable: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      appStore: useAppStore(),
      authStore: useAuthStore(),
    };
  },
  computed: {
    btnIcon(): string {
      return this.$route.meta.supportBack ? "ph-bold ph-arrow-left" : "ph-bold ph-x";
    },
    askToLeave(): boolean {
      return !!this.$route.meta.requestToLeave;
    },
  },
  methods: {
    async onLeave() {
      if (!this.askToLeave || this.appStore.isFallback) {
        this.redirect();
        return;
      }
      const confirm = await requestToLeave();
      if (confirm) this.redirect();
    },

    redirect(): void {
      if (this.$route?.meta?.backPageName) {
        if (this.$route?.query?.previousRoute) {
          this.$router.push({ name: this.$route.query.previousRoute as string });
        } else {
          const paramId = (this.$router.options.history.state.appData as Record<string, string>)?.paramId;

          this.$router.push({
            name: this.$route?.meta?.backPageName as string,
            query: this.$route?.query,
            params: paramId ? { id: paramId } : undefined,
          });
        }
      } else if (this.appStore.lastBackPath) {
        this.$router.push(this.appStore.lastBackPath);
      } else {
        const landingPage = this.appStore.getUserLandingPage();
        this.$router.push({ name: landingPage });
      }
    },
  },
});
</script>

<style lang="scss">
[aid="header-leave-page-button"] .q-icon:before {
  color: $black-54;
}
</style>
