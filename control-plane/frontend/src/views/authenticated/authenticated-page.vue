<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div v-if="appIsReady" class="authenticated-page-wrapper">
    <custom-banner />
    <new-layout>
      <admin-onboarding-modal
        v-if="showAdminOnboardingModal"
        @close="closeOnboardingModal"
        :cluster-id="selectedClusterId"
        :skip-cluster-setup="shouldSkipClusterSetup"
      />
      <template v-if="allowToDisplayBanner">
        <cluster-status-bar />
      </template>
      <runai-fallback-page v-if="isFallback" />
      <template v-else>
        <runai-default-loader v-if="pageLoading" />
        <router-view v-show="!pageLoading" aid="router-content" />
      </template>
      <section class="q-py-md q-ml-md version-info" v-if="!isFullHeight && !hideVersion && !isFullPage">
        <div>
          <div class="app-info" v-if="lastLogin">Last login: {{ lastLogin }}</div>
          <div class="app-info" v-if="appVersion">Version: {{ appVersion }}</div>
        </div>
      </section>
      <admin-message-modal v-if="adminMessage" :message="adminMessage" @close="closeAdminMessage" />
      <system-message-modal v-if="systemMessage" :message="systemMessage" @close="closeSystemMessage" />
      <guided-tour-popup-manager />

      <div class="intro-modal-key" :aid="adminOnboardingModalKey"></div>
    </new-layout>
  </div>
</template>

<script lang="ts">
import { defineComponent } from "vue";

// stores
import { useAppStore } from "@/stores/app.store";
import { useAuthStore } from "@/stores/auth.store";
import { useSettingStore } from "@/stores/setting.store";
import { useClusterStore } from "@/stores/cluster.store";
import { usePermissionStore } from "@/stores/permissions.store";
import { useIdpsStore } from "@/stores/idps.store";
import { useSecuritySettingsStore } from "@/stores/security-settings.store";

// components
import { RunaiFallbackPage } from "@/components/common/runai-fallback-page";
import { RunaiDefaultLoader } from "@/components/common/runai-default-loader";
import { AdminMessageModal } from "@/components/admin-message-modal";
import { SystemMessageModal } from "@/components/system-message-modal";
import { ClusterStatusBar } from "@/components/cluster/cluster-status-bar";
import { NewLayout } from "@/components/infra/runai-layout/new-layout";
import { GuidedTourPopupManager } from "@/components/common/guided-tour-popup";

// services
import { IdleTimer } from "@/services/infra/idle-timer.service/idle-timer.service";
import {
  AI_PRACTITIONER_ROLE_NAME,
  SYSTEM_ADMIN_ROLE_NAME,
} from "@/services/control-plane/team-provisioning.service/team-provisioning.service";
// utils
import { intervalUtil } from "@/utils/interval.util";
import { adminMessageUtil } from "@/utils/admin-message.util/admin-message.util";
import { storageUtil } from "@/utils/storage.util";

// models
import { ForbiddenError, HttpErrorResponse, NotFoundError, UnauthorizedError } from "@/models/http-response.model";
import { type IAdminMessage, DEFAULT_BROWSER_SESSION_TIMEOUT_MINUTES } from "@/models/setting.model";
import type { Idp } from "@/swagger-models/identity-manager-client";
import { Action, ResourceType } from "@/swagger-models/authorization-client";

// constant
import { EIntervalLabels } from "@/models/interval.model";
import { CLUSTER_ID_FILTER, ADMIN_ONBOARDING_MODAL_SHOWN } from "@/common/storage.constant";
import { STORAGE_KEYS } from "@/models/storage.model";
import { OnboardingStep } from "@/components/onboarding/models/admin-onboarding.model";

// router
import { ERROR_ROUTE_NAMES } from "@/router/error.routes/error.routes.names";
import { DASHBOARD_ROUTE_NAMES } from "@/router/dashboard.routes/dashboard.routes.names";
import { LOGIN_ROUTE_NAMES } from "@/router/login.routes/login.routes.names";

import { unleashConfigService } from "@/services/infra/unleash.service/unleash-config.service";
import type { ClustersMinimalData } from "@/swagger-models/cluster-service-client";
import { filterService } from "@/services/filter.service/filter.service";
import AdminOnboardingModal from "@/components/onboarding/admin-onboarding-modal";
import CustomBanner from "@/components/custom-banner/custom-banner.vue";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";
import { ErrorAlert } from "@/utils/error-alert.util";
import { CLUSTER_ROUTE_NAMES } from "@/router/cluster.routes/cluster.routes.names";
import { stepsRequiredPermissions } from "@/components/onboarding/models/admin-onboarding.model";

import { useRoleStore } from "@/stores/role.store";
import { useNotificationStore } from "@/stores/notifications.store";

export default defineComponent({
  components: {
    AdminOnboardingModal,
    CustomBanner,
    ClusterStatusBar,
    RunaiFallbackPage,
    RunaiDefaultLoader,
    AdminMessageModal,
    SystemMessageModal,
    NewLayout,
    GuidedTourPopupManager,
  },
  data() {
    return {
      appStore: useAppStore(),
      authStore: useAuthStore(),
      settingStore: useSettingStore(),
      clusterStore: useClusterStore(),
      permissionStore: usePermissionStore(),
      notificationsStore: useNotificationStore(),
      idpsStore: useIdpsStore(),
      securitySettingsStore: useSecuritySettingsStore(),
      roleStore: useRoleStore(),
      idleTimerService: IdleTimer.getInstance() as IdleTimer,
      adminMessage: null as IAdminMessage | null,
      systemMessage: null as IAdminMessage | null,
      systemMessageShownThisSession: false,
      adminMessageShownThisSession: false,
      appIsReady: false as boolean,
      showAdminOnboardingModal: false as boolean,
    };
  },
  async created() {
    try {
      this.appStore.setAppLoaded(false);
      this.appIsReady = false;
      if (this.isTokenMissing) {
        await this.redirectToLogin();
        return;
      }

      await this.loadAuthenticatedData();

      this.authStore.updateUserLastLoggedIn();
      this.authStore.initCrossTabLogoutDetection();

      await this.initSecuritySettings();

      await this.startAppIntervals();
      await this.initUnleash();
    } catch (error: unknown) {
      console.error(error);
      if (error instanceof HttpErrorResponse) {
        if (error.statusCode === UnauthorizedError.statusCode) {
          await this.redirectToLogin();
        } else if (error.statusCode === ForbiddenError.statusCode) {
          await this.$router.replace({ name: ERROR_ROUTE_NAMES.NO_ROLES });
        } else if (error.statusCode > 500) {
          await this.$router.replace({ name: ERROR_ROUTE_NAMES.BAD_GATEWAY_ERROR });
        } else if (error.statusCode === NotFoundError.statusCode) {
          await this.$router.replace({ name: ERROR_ROUTE_NAMES.NOT_FOUND });
        }
      }
    } finally {
      this.appStore.setAppLoaded(true);
      this.appIsReady = true;
    }
  },
  computed: {
    appVersion(): string {
      return this.appStore.version;
    },
    lastLogin(): string {
      return this.authStore.userLastLogin;
    },
    adminOnboardingModalKey(): string {
      return ADMIN_ONBOARDING_MODAL_SHOWN;
    },
    clusters(): Array<ClustersMinimalData> {
      return this.clusterStore.clusters;
    },
    isTokenMissing(): boolean {
      return !this.authStore.accessToken;
    },
    allowToDisplayBanner(): boolean {
      return this.hasAccessToken && !this.pageLoading && !this.isFallback;
    },
    pageLoading(): boolean {
      return this.appStore.isPageLoading;
    },
    isFallback(): boolean {
      return this.appStore.isFallback;
    },
    isFullHeight(): boolean {
      return !!this.$route.meta.fullHeight;
    },
    isFullPage(): boolean {
      return !!this.$route.meta.fullPage;
    },
    hideVersion(): boolean {
      return !!this.$route?.meta.hideVersion;
    },
    hasAccessToken(): boolean {
      return !!this.authStore.accessToken;
    },
    hasMinTeamSetupPermissions(): boolean {
      return this.permissionStore.isActionGrantedForEveryResource(
        [ResourceType.Project, ResourceType.AccessRules, ResourceType.Users],
        Action.Create,
      );
    },

    shouldSkipClusterSetup(): boolean {
      // Show only team setup if there are clusters but no projects
      // This logic assumes showAdminOnboardingModal is set due to no projects (see displayOnboardingModalIfNeeded)
      return this.clusters.length > 0;
    },
    selectedClusterId(): string {
      return this.clusterStore.selectedClusterId;
    },
  },
  methods: {
    async loadAuthenticatedData(): Promise<void> {
      await this.loadAppData();
    },

    async loadIdps(): Promise<Idp[] | void> {
      if (this.permissionStore.hasPermission(ResourceType.SecuritySettings, Action.Read))
        return this.idpsStore.loadIdps();
    },
    async loadAppData(): Promise<void> {
      await this.authStore.loadUserInfo();
      await this.permissionStore.loadPermissions();
      if (!this.permissionStore.permissionsList.length) {
        throw new ForbiddenError("Must have mandatory permissions to view app");
      }
      await Promise.all([
        this.clusterStore.loadClusters(),
        this.authStore.loadTenant(),
        this.settingStore.loadSetting(),
        this.notificationsStore.loadNotificationChannels(),
      ]);
      this.initSelectedClusterId();
      await this.loadSupplementaryData();
      await this.changeLandingPageIfNeeded();
      if (!this.isRoutePermitted()) {
        await this.$router.replace({ name: ERROR_ROUTE_NAMES.NO_ROLES });
      } else if (this.clusters.length === 0 && this.clusterStore.canCreateCluster) {
        // waiting for clusters to load, If there are no clusters toggle the cluster setup modal
        await this.$router.replace({ name: CLUSTER_ROUTE_NAMES.CLUSTER_NEW });

        if (await this.hasOnboardingStepsToShow()) {
          this.showAdminOnboardingModal = true;
        } else {
          this.appStore.setIsOnboardingModalClosed(true);
        }
      } else {
        await this.displayOnboardingModalIfNeeded();
      }
    },
    isRoutePermitted(): boolean {
      if (
        this.$route.name === DASHBOARD_ROUTE_NAMES.RESOURCES_INDEX &&
        !this.permissionStore.isActionGrantedForEveryResource(
          [ResourceType.Cluster, ResourceType.Department, ResourceType.Project],
          Action.Read,
        )
      ) {
        return false;
      }
      if (!this.$route.meta.requiresAuth || !this.$route.meta.minPermittedActions || !this.$route.meta.resourceType)
        return true;

      return this.permissionStore.hasSomePermittedActions(
        this.$route.meta.resourceType,
        this.$route.meta.minPermittedActions,
      );
    },
    async redirectToLogin(): Promise<void> {
      await this.$router.replace({ name: LOGIN_ROUTE_NAMES.LOGIN_PAGE });
    },
    async changeLandingPageIfNeeded(): Promise<void> {
      const landingPage = this.appStore.getUserLandingPage(this.$route.name?.toString());
      if (landingPage === this.$route.name) return;
      await this.$router.replace({ name: landingPage });
    },
    closeAdminMessage(dismiss: boolean): void {
      if (dismiss && this.adminMessage) adminMessageUtil.dismissAdminMessage(this.adminMessage.id);
      this.adminMessageShownThisSession = true;
      this.adminMessage = null;
    },
    closeSystemMessage(): void {
      if (!this.systemMessage) return;

      adminMessageUtil.dismissSystemMessage(this.systemMessage.id);
      this.systemMessageShownThisSession = true;
      this.systemMessage = null;
    },
    async loadAdminMessage(): Promise<void> {
      if (this.adminMessageShownThisSession) return;
      const adminMessage = await this.settingStore.loadAdminMessage();
      if (!adminMessage) return;
      this.adminMessage = adminMessageUtil.getAdminMessage(adminMessage);
    },
    async loadSystemMessage(): Promise<void> {
      if (this.systemMessageShownThisSession) return;
      const raw = await this.settingStore.loadSystemMessage();
      if (!raw) return;
      this.systemMessage = adminMessageUtil.getSystemMessage(raw);
    },
    async startAppIntervals(): Promise<void> {
      await this.loadSystemMessage();
      await this.loadAdminMessage();
      intervalUtil.startInterval(EIntervalLabels.AdminMessage, async () => {
        await this.loadSystemMessage();
        await this.loadAdminMessage();
      });
    },
    async initUnleash(): Promise<void> {
      await unleashConfigService.init(this.authStore.tenant.id.toString());
    },
    stopAppIntervals(): void {
      intervalUtil.stopInterval(EIntervalLabels.AdminMessage);
    },
    //ONBOARDING MODAL
    async displayOnboardingModalIfNeeded(): Promise<void> {
      try {
        const modalAlreadyShown = storageUtil.get<boolean | null>(this.adminOnboardingModalKey);
        if (!this.hasMinTeamSetupPermissions || modalAlreadyShown) {
          this.appStore.setIsOnboardingModalClosed(true);
          return;
        }
        const numberOfProjects = await orgUnitService.countProjects({});
        if (numberOfProjects === 0 && this.hasMinTeamSetupPermissions) {
          this.showAdminOnboardingModal = true;
          storageUtil.save(this.adminOnboardingModalKey, true);
        } else {
          this.appStore.setIsOnboardingModalClosed(true);
        }
      } catch (e) {
        new ErrorAlert({ generalMessage: ErrorAlert.failedLoadingMessage("projects") });
        this.appStore.setIsOnboardingModalClosed(true);
      }
    },
    closeOnboardingModal(): void {
      this.showAdminOnboardingModal = false;
      storageUtil.save(this.adminOnboardingModalKey, true);
      this.appStore.setIsOnboardingModalClosed(true);
    },
    async hasOnboardingStepsToShow(): Promise<boolean> {
      let hasAIPractitionerRole = false;
      let hasSystemAdminRole = false;
      try {
        await this.roleStore.loadRoles();
        hasAIPractitionerRole = this.roleStore.roleList.some((role) => role.name === AI_PRACTITIONER_ROLE_NAME);
        hasSystemAdminRole = this.roleStore.roleList.some((role) => role.name === SYSTEM_ADMIN_ROLE_NAME);
      } catch {
        //do nothing meanwhile
      }
      const stepsPermissions = [
        {
          stepId: OnboardingStep.ClusterSetup,
          permissionChecks: stepsRequiredPermissions[OnboardingStep.ClusterSetup],
          additionalConditions: [this.clusters.length === 0],
        },
        {
          stepId: OnboardingStep.SsoSetup,
          permissionChecks: stepsRequiredPermissions[OnboardingStep.SsoSetup],
          additionalConditions: [
            !storageUtil.get(STORAGE_KEYS.ONBOARDING_SKIPPED_SSO),
            !this.authStore.isSSO,
            hasSystemAdminRole,
          ],
        },
        {
          stepId: OnboardingStep.EmailServerSetup,
          permissionChecks: stepsRequiredPermissions[OnboardingStep.EmailServerSetup],
          additionalConditions: [
            !storageUtil.get(STORAGE_KEYS.ONBOARDING_SKIPPED_EMAIL_SERVER_SETUP),
            !this.appStore.isMultiTenant,
            hasSystemAdminRole,
          ],
        },
        {
          stepId: OnboardingStep.TeamSetup,
          permissionChecks: stepsRequiredPermissions[OnboardingStep.TeamSetup],
          additionalConditions: [!storageUtil.get(STORAGE_KEYS.ONBOARDING_SKIPPED_TEAM_SETUP), hasAIPractitionerRole],
        },
      ];

      return stepsPermissions.some((stepPermission) => {
        const hasRequiredPermissions = this.permissionStore.hasAllPermissions(stepPermission.permissionChecks || []);
        const hasAdditionalConditions = stepPermission.additionalConditions.every((condition) => Boolean(condition));
        return hasRequiredPermissions && hasAdditionalConditions;
      });
    },
    /**
     * @description Initializes the selected cluster ID.
     * This function should be called only on first load.
     */
    initSelectedClusterId(): void {
      const urlClusterId = this.$route.query.clusterUuid as string | null;

      const storageClusterId = storageUtil.get<string | null>(CLUSTER_ID_FILTER);
      if (!storageClusterId) {
        filterService.clearAllTablesFilters();
      }
      if (urlClusterId) {
        this.clusterStore.setSelectedClusterById(urlClusterId);
      } else if (storageClusterId) {
        this.clusterStore.setSelectedClusterById(storageClusterId);
      } else {
        this.clusterStore.setSelectedClusterById(this.clusterStore.defaultClusterId);
      }
      console.info(`%cSet active cluster ${this.clusterStore.selectedClusterName}`, "color: green;");
    },
    /**
     * Loads supplementary data required for the application.
     * Note: The function calls within this method are not critical for the app to function.
     * If any of these calls fail, the app will still continue to operate.
     *
     * @returns {void}
     */
    async loadSupplementaryData(): Promise<void> {
      try {
        this.appStore.loadReportApiAvailability();
        this.loadIdps();
        await this.authStore.loadUserOrgUnits();
      } catch (e) {
        console.error(`Error loading registries, org units or idps: ${e}`);
      }
    },
    async initSecuritySettings(): Promise<void> {
      try {
        const settings = await this.securitySettingsStore.getSettings();
        const sessionTimeout = settings?.browserSessionTimeout;
        const idleTimeout = sessionTimeout?.value?.idleTimeout;
        if (idleTimeout !== undefined && idleTimeout !== null) {
          await this.idleTimerService.updateTimeout(idleTimeout);
        } else {
          await this.idleTimerService.updateTimeout(DEFAULT_BROWSER_SESSION_TIMEOUT_MINUTES);
        }
      } catch (error) {
        console.error("Failed to initialize security settings", error);
        await this.idleTimerService.updateTimeout(DEFAULT_BROWSER_SESSION_TIMEOUT_MINUTES);
      }
    },
  },
  unmounted(): void {
    this.stopAppIntervals();
  },
});
</script>
<style scoped lang="scss">
.version-info {
  display: flex;
  flex: 1;
  align-items: flex-end;

  .app-info {
    font-size: 12px;
  }
}

.intro-modal-key {
  display: none;
}
</style>
