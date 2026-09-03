<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <section class="settings-index q-pt-md">
    <analytics-section :analytics-section-settings="analyticsSectionSettings" v-if="shouldShowAnalyticsSection" />
    <resources-section :resource-settings="resourceSettings" v-if="canViewGeneralSettings" />
    <workloads-section :workload-section-settings="workloadSectionSettings" v-if="shouldShowWorkloadSection" />
    <security-section :security-settings="securitySectionSettings" v-if="canViewSecuritySection" />
    <notifications-section v-if="canViewNotificationsSection" />
    <cluster-authentications-section v-if="canViewGeneralSettings" />
    <branding-section v-if="canViewBrandingSection" />
  </section>
</template>

<script lang="ts">
import { defineComponent } from "vue";

// components
import AnalyticsSection from "@/components/settings/sections/analytics/analytics-section.vue";
import SecuritySection from "@/components/settings/sections/security/security-section.vue";
import NotificationsSection from "@/components/settings/sections/notifications/notifications-section.vue";
import { WorkloadsSection } from "@/components/settings/sections/workloads";
import BrandingSection from "@/components/settings/sections/branding/branding-section.vue";

// stores
import { useAppStore } from "@/stores/app.store";
import { useSettingStore } from "@/stores/setting.store";
import { useAuthStore } from "@/stores/auth.store";
import { usePermissionStore } from "@/stores/permissions.store";
import { useClusterStore } from "@/stores/cluster.store";
// models
import type {
  IAnalyticsSectionSettings,
  IResourceSectionSettings,
  IWorkloadsSectionSettings,
} from "@/models/setting.model";
import { SettingKeys } from "@/models/setting.model";
import type { Setting } from "@/models/setting.model";
import { SettingsKeyEnum, type GetLogoutRedirectUriSettingValue } from "@/swagger-models/identity-manager-client";
import { ResourceType, Action } from "@/swagger-models/authorization-client";

// utils
import { ResourcesSection } from "@/components/settings/sections/resources";
import ClusterAuthenticationsSection from "@/components/settings/sections/cluster-authentication/cluster-autentications-section.vue";
import { unleashService } from "@/services/infra/unleash.service/unleash.service";
import { resolveDocsLinksInHtml } from "@/constants/docs";
import { MIN_CLUSTER_VERSION_FOR_NGC_API_KEY } from "@/common/version.constant";

// Read is included intentionally: users with only Read permission should see
// these sections with controls disabled, not a blank page.
const ALL_CRUD_ACTIONS = [Action.Create, Action.Read, Action.Update, Action.Delete];

export default defineComponent({
  name: "settings-index",
  components: {
    ClusterAuthenticationsSection,
    ResourcesSection,
    WorkloadsSection,
    NotificationsSection,
    SecuritySection,
    AnalyticsSection,
    BrandingSection,
  },
  data() {
    return {
      settingStore: useSettingStore(),
      authStore: useAuthStore(),
      appStore: useAppStore(),
      permissionStore: usePermissionStore(),
      clusterStore: useClusterStore(),
    };
  },
  created() {
    this.appStore.setPageLoading(false);
  },
  computed: {
    analyticsSectionSettings(): IAnalyticsSectionSettings {
      return {
        advancedGpuMetrics: this.getSettingValue(SettingKeys.AdvancedGPUMetrics),
      };
    },
    resourceSettings(): IResourceSectionSettings {
      return {
        projectsOverDepartmentQuota: this.getSettingValue(SettingKeys.LimitQuotaOverSubscription),
        cpuOverQuota: this.getSettingValue(SettingKeys.CPUResourcesQuota),
        overQuotaPriority: this.getSettingValue(SettingKeys.ProjectManualOverQuota),
        gpuOptimization: this.getSettingValue(SettingKeys.GpuOverProvisioning),
      };
    },
    securitySectionSettings(): GetLogoutRedirectUriSettingValue {
      const setting = this.settingStore.settings.get(SettingsKeyEnum.LogoutRedirectUri) as Setting;
      return {
        uri: setting?.value || null,
      };
    },
    workloadSectionSettings(): IWorkloadsSectionSettings {
      const workloadSettings: IWorkloadsSectionSettings = {
        policies: this.getSettingValue(SettingKeys.EnablePolicies),
        models: this.getSettingValue(SettingKeys.EnableModelCatalog),
        nim: this.getSettingValue(SettingKeys.EnableNIM),
        allowGPT2Model: this.getSettingValue(SettingKeys.AllowGPT2Model),
        hybridSubmission: this.getSettingValue(SettingKeys.EnableHybridSubmission),
        templatesV2: this.getSettingValue(SettingKeys.EnableTemplatesV2),
        dataVolumes: this.getSettingValue(SettingKeys.EnableDataVolumes),
        // The yaml submission setting is loaded inside the component, it is implemented in the new settings convention, not in the old backend.
        // This is a placeholder just to keep it in the right order - the value will be loaded inside the component
        yamlSubmission: { key: SettingKeys.YamlSubmission },
        applicationSubmission: { key: SettingKeys.ApplicationSubmission },
        ngcIntegration: this.getSettingValue(SettingKeys.EnableNgcIntegration),
      };
      if (!unleashService.showNgcIntegration()) delete workloadSettings.ngcIntegration;

      return workloadSettings;
    },
    canViewRegistriesSection(): boolean {
      return (
        this.permissionStore.hasSomePermittedActions(ResourceType.Registries, ALL_CRUD_ACTIONS) &&
        this.clusterStore.isAtleastOneClusterSufficientForVersion(MIN_CLUSTER_VERSION_FOR_NGC_API_KEY)
      );
    },
    shouldShowWorkloadSection(): boolean {
      const hasWorkloadSettings = !!Object.values(this.workloadSectionSettings).find((value: Setting) => {
        return value && value.key;
      });
      return (this.canViewGeneralSettings && hasWorkloadSettings) || this.canViewRegistriesSection;
    },
    shouldShowAnalyticsSection(): boolean {
      return (
        this.canViewGeneralSettings &&
        !!Object.values(this.analyticsSectionSettings).find((value: Setting) => {
          return value && value.key;
        })
      );
    },
    canViewSecuritySection(): boolean {
      return this.permissionStore.hasSomePermittedActions(ResourceType.SecuritySettings, ALL_CRUD_ACTIONS);
    },
    canViewGeneralSettings(): boolean {
      return this.permissionStore.hasSomePermittedActions(ResourceType.Settings, ALL_CRUD_ACTIONS);
    },
    canViewNotificationsSection(): boolean {
      return this.permissionStore.hasSomePermittedActions(ResourceType.NotificationSettings, ALL_CRUD_ACTIONS);
    },
    canViewBrandingSection(): boolean {
      return this.permissionStore.hasSomePermittedActions(ResourceType.BrandingSettings, ALL_CRUD_ACTIONS);
    },
  },
  methods: {
    getSettingValue(key: string): Setting | undefined {
      let setting = this.settingStore.settings.get(key);
      if (!setting) {
        // Check if this is a BFF setting (enableNIM, etc.)
        setting = this.settingStore.getBffSettingAsOldFormat(key);
      }

      if (!setting) return undefined;
      if (!setting.description) return setting;
      return { ...setting, description: resolveDocsLinksInHtml(setting.description) };
    },
  },
});
</script>

<style lang="scss">
.settings-container {
  margin: auto;
  width: 800px;

  .runai-expansion-header-content-left {
    font-size: 16px;
    font-weight: 500;
  }
}
</style>
