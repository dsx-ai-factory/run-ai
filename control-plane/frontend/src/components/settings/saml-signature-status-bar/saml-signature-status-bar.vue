<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <notification-alert-bar
    v-if="shouldShow"
    aid="saml-signature-status-bar"
    :variant="ENotificationBarVariant.Error"
    icon-override="ph-bold ph-warning-circle"
    show-h1-content
    title="Your SAML connection isn't validating identity provider signatures,"
  >
    which may impact user verification. To enable validation,
    <template #inline-actions>
      <router-link :to="securitySettingsLocation">reconfigure SAML in your SSO settings.</router-link>
    </template>
  </notification-alert-bar>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import type { RouteLocationRaw } from "vue-router";
import { ENotificationBarVariant, NotificationAlertBar } from "@/components/common/notification-alert-bar";
import { useIdpsStore } from "@/stores/idps.store";
import { SETTINGS_ROUTE_NAMES } from "@/router/settings.routes/settings.routes.names";
import { SECURITY_SECTION_QUERY } from "@/router/settings.routes/settings.routes.constants";

export default defineComponent({
  name: "saml-signature-status-bar",
  components: { NotificationAlertBar },
  data() {
    return {
      idpsStore: useIdpsStore(),
      ENotificationBarVariant,
    };
  },
  computed: {
    // On the settings page itself the security section already shows the same warning inline.
    shouldShow(): boolean {
      return this.idpsStore.shouldShowSamlSignatureBanner && this.$route?.name !== SETTINGS_ROUTE_NAMES.SETTINGS_INDEX;
    },
    securitySettingsLocation(): RouteLocationRaw {
      return { name: SETTINGS_ROUTE_NAMES.SETTINGS_INDEX, query: SECURITY_SECTION_QUERY };
    },
  },
});
</script>
