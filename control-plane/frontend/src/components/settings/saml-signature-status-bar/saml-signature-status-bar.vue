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

<script setup lang="ts">
import { computed } from "vue";
import { useRoute } from "vue-router";
import type { RouteLocationRaw } from "vue-router";
import { ENotificationBarVariant, NotificationAlertBar } from "@/components/common/notification-alert-bar";
import { useIdpsStore } from "@/stores/idps.store";
import { SETTINGS_ROUTE_NAMES } from "@/router/settings.routes/settings.routes.names";
import { SECURITY_SECTION_QUERY } from "@/router/settings.routes/settings.routes.constants";

defineOptions({ name: "SamlSignatureStatusBar" });

const idpsStore = useIdpsStore();
const route = useRoute();

// On the settings page itself the security section already shows the same warning inline.
const shouldShow = computed(
  (): boolean => idpsStore.shouldShowSamlSignatureBanner && route?.name !== SETTINGS_ROUTE_NAMES.SETTINGS_INDEX,
);

const securitySettingsLocation: RouteLocationRaw = {
  name: SETTINGS_ROUTE_NAMES.SETTINGS_INDEX,
  query: SECURITY_SECTION_QUERY,
};
</script>
