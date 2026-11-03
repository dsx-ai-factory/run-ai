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
    title="Security risk: Identity provider signatures aren't being validated."
  >
    Login responses can't be confirmed as authentic, which may allow unauthorized access to user accounts. To enable
    validation,
    <template #inline-actions>
      <router-link :to="securitySettingsLocation">reconfigure SAML in SSO settings.</router-link>
    </template>
  </notification-alert-bar>
</template>

<script setup lang="ts">
import { computed } from "vue";
import { useRoute } from "vue-router";
import type { RouteLocationRaw } from "vue-router";
import { ENotificationBarVariant, NotificationAlertBar } from "@/components/common/notification-alert-bar";
import { useIdpsStore } from "@/stores/idps.store";
import { usePageBannerStore } from "@/stores/page-banner.store";
import { SETTINGS_ROUTE_NAMES } from "@/router/settings.routes/settings.routes.names";
import { SECURITY_SECTION_QUERY } from "@/router/settings.routes/settings.routes.constants";

defineOptions({ name: "SamlSignatureStatusBar" });

const idpsStore = useIdpsStore();
const pageBannerStore = usePageBannerStore();
const route = useRoute();

// Not needed on the settings page (the security section shows the same warning inline) nor on the full-width and
// full-page screens (create/edit flows, error pages). Any other banner on the page takes priority; this one shows
// once they are gone.
const hiddenOnThisPage = computed(
  (): boolean =>
    route?.name === SETTINGS_ROUTE_NAMES.SETTINGS_INDEX || !!route?.meta?.fullWidth || !!route?.meta?.fullPage,
);

const shouldShow = computed(
  (): boolean => idpsStore.shouldShowSamlSignatureBanner && !hiddenOnThisPage.value && !pageBannerStore.hasActiveBanner,
);

const securitySettingsLocation: RouteLocationRaw = {
  name: SETTINGS_ROUTE_NAMES.SETTINGS_INDEX,
  query: SECURITY_SECTION_QUERY,
};
</script>
