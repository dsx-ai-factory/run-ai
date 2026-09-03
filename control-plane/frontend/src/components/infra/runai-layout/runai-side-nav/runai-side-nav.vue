<!--
  SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
  SPDX-License-Identifier: Apache-2.0
-->

<template>
  <q-drawer
    :model-value="true"
    side="left"
    :mini-width="74"
    :width="289"
    :mini="miniState"
    persistent
    behavior="desktop"
    class="runai-side-nav overflow-hidden"
  >
    <section class="entire-side-navigation row no-wrap">
      <section class="left-side-navigation column items-center">
        <div class="runai-corner-logo">
          <runai-distribution-logo
            v-if="!isWhiteLabel"
            composition="vendor"
            width="40px"
            height="32px"
            vendor-width="40px"
            :vendor-height="navigationLogoHeight"
            :vendor-canvas-height="navigationLogoCanvasHeight"
            :vendor-canvas-top="navigationLogoCanvasTop"
            cc-vendor-height="9px"
            gap="0"
            :vendor-asset="navigationLogoAsset"
            :text-color="navigationLogoTextColor"
            alt="NVIDIA"
          />
          <runai-brand-logo v-else :img="mainLogo.src" :height="mainLogo.height" />
        </div>

        <nav class="menu-items-container q-mt-sm" role="list" aria-label="Main navigation">
          <runai-side-nav-menu-item
            v-for="menuItem in subjectList"
            :key="menuItem.value"
            :menu-item="menuItem"
            :sub-menu-items="menuItem.items"
            @item-clicked="selectSubject"
            @item-hovered="setActiveMenuItem"
            @item-out-of-focus="hideActivePopup"
          />
        </nav>
        <div class="q-mt-auto q-mb-md">
          <q-btn
            round
            text-color="white"
            size="12px"
            unelevated
            @click="toggleDrawerState"
            aid="toggle-drawer"
            class="toggle-drawer-btn runai-icon-btn--xsmall"
            :aria-label="miniState ? 'Open sub-menu' : 'Close sub-menu'"
          >
            <q-icon
              class="toggle-btn text-weight-bold"
              :class="{ open: !miniState }"
              name="ph-bold ph-caret-right"
              size="16px"
            />
            <q-tooltip anchor="top right" self="bottom middle" :offset="[15, 5]">
              {{ miniState ? "Open sub-menu" : "Close sub-menu" }}
            </q-tooltip>
          </q-btn>
        </div>
      </section>

      <section class="expanded-section" v-if="showExpandedNavigation" :class="{ 'logo-hidden': !subLogo }">
        <div>
          <div class="runai-corner-logo" v-if="subLogo">
            <img v-if="isWhiteLabel" :src="subLogo" alt="logo" class="sub-logo" />
            <runai-distribution-logo v-else composition="product" width="45px" height="12px" gap="0" class="sub-logo" />
          </div>
          <runai-sub-menu :items="subMenuItems" />
        </div>
      </section>
    </section>
  </q-drawer>
</template>
<script lang="ts">
import { defineComponent } from "vue";
import type { QPopupProxy } from "quasar";

// stores
import { useSettingStore } from "@/stores/setting.store";
import { usePermissionStore } from "@/stores/permissions.store";
import { useClusterStore } from "@/stores/cluster.store";

// models
import type { EMenuSubjectNames } from "../routes-model/routes-model";
import { type IMenuItem, type ISubjectRoutes, menuSubjects } from "../routes-model/routes-model";
import { ResourceType, Action } from "@/swagger-models/authorization-client";
import { DASHBOARD_ROUTE_NAMES } from "@/router/dashboard.routes/dashboard.routes.names";
// cmps
import { RunaiSubMenu } from "@/components/infra/runai-layout/runai-sub-menu";
import { RunaiSideNavMenuItem } from "@/components/infra/runai-layout/runai-side-nav-menu-item";
import { RunaiBrandLogo } from "@/components/common/runai-brand-logo";
import { RunaiDistributionLogo } from "@/components/common/runai-distribution-logo";

// utils
import { deepCopy } from "@/utils/common.util";
import type { SettingKeys } from "@/models/setting.model";
import type { EFeatureToggle } from "@/services/infra/unleash.service/unleash.model";
import { unleashService } from "@/services/infra/unleash.service/unleash.service";
import { useAppStore } from "@/stores/app.store";
import { ReportRoutesName } from "@/router/report.routes/report.routes.names";
import { DATA_SOURCE_ROUTE_NAMES } from "@/router/data-source.routes/data-source.routes.names";
import { AI_APPLICATION_ROUTE_NAMES } from "@/router/ai-application.routes/ai-application.routes.names";
import { MIN_CLUSTER_VERSION_FOR_AI_APPLICATIONS } from "@/common/version.constant";
import {
  CC_DISTRIBUTION_ASSETS_PROFILE,
  DISTRIBUTION_LOGO_ASSETS,
  getDistributionAssetsProfile,
} from "@/services/infra/distribution-assets.service";

export default defineComponent({
  name: "runai-side-nav",
  components: {
    RunaiSideNavMenuItem,
    RunaiSubMenu,
    RunaiBrandLogo,
    RunaiDistributionLogo,
  },
  emits: ["drawer-state-changed"],
  props: {
    initialState: {
      type: Boolean,
      default: true,
    },
  },
  data() {
    return {
      settingStore: useSettingStore(),
      permissionStore: usePermissionStore(),
      clusterStore: useClusterStore(),
      appStore: useAppStore(),
      activeMenuItem: null as QPopupProxy | null,
      miniState: true,
    };
  },
  created() {
    this.miniState = this.initialState;
  },
  computed: {
    isWhiteLabel(): boolean {
      return this.appStore.isWhiteLabel;
    },
    isOssDistribution(): boolean {
      return getDistributionAssetsProfile() === CC_DISTRIBUTION_ASSETS_PROFILE;
    },
    navigationLogoAsset() {
      // The navigation rail is always dark. Select the source artwork intended for
      // dark surfaces instead of depending on an SVG filter to recolor the CE mark.
      return DISTRIBUTION_LOGO_ASSETS.VendorLogoOnDark;
    },
    navigationLogoTextColor(): "default" | "white" {
      return this.isOssDistribution ? "default" : "white";
    },
    navigationLogoHeight(): string {
      return "9px";
    },
    navigationLogoCanvasHeight(): string {
      return "9px";
    },
    navigationLogoCanvasTop(): string {
      return "0";
    },
    subLogo(): string | undefined {
      return this.appStore.theme.logo.sub || undefined;
    },
    mainLogo(): { src: string; height: string } {
      return this.appStore.mainLogoLight;
    },
    availableSubjects(): IMenuItem[] {
      return [
        this.filterDashboardItems(deepCopy(menuSubjects.analytics)),
        this.filterWorkloadManagerItems(deepCopy(menuSubjects.workloadManager)),
        this.filterHiddenItems(deepCopy(menuSubjects.resources)),
        this.filterHiddenItems(deepCopy(menuSubjects.organization)),
        this.filterHiddenItems(deepCopy(menuSubjects.access)),
        this.filterHiddenItems(deepCopy(menuSubjects.policies)),
        this.filterHiddenItems(deepCopy(menuSubjects.admin)),
      ].filter((subject) => subject.items.length > 0); // filter out subjects by empty items
    },
    currentSubject(): EMenuSubjectNames {
      return this.$route.meta.parentSubject as EMenuSubjectNames;
    },
    subjectList(): IMenuItem[] {
      return this.availableSubjects.map((item) => ({
        ...item,
        isActive: this.currentSubject === item.value,
      }));
    },
    showExpandedNavigation(): boolean {
      return !this.miniState;
    },
    subMenuItems(): ISubjectRoutes[] {
      return this.subjectList.find((item) => item.isActive)?.items || [];
    },
    themeLogo(): string | undefined {
      return this.appStore.mainLogoLight.src;
    },
  },
  methods: {
    setActiveMenuItem(item: QPopupProxy): void {
      if (!this.miniState) return;
      this.hideActivePopup();
      this.activeMenuItem = item;
      item.show();
    },
    hideActivePopup(): void {
      if (!this.activeMenuItem) return;
      this.activeMenuItem.hide && this.activeMenuItem.hide();
      this.activeMenuItem = null;
    },
    selectSubject(menuItem: IMenuItem): void {
      this.hideActivePopup();
      this.$router.push({ name: menuItem.items[0].name }); // default to first item in subject
    },
    toggleDrawerState(): void {
      this.miniState = !this.miniState;
      this.$emit("drawer-state-changed", this.miniState);
    },
    filterHiddenItems(subject: IMenuItem): IMenuItem {
      const itemsInSubject = subject.items.filter((item) => {
        if (item.subItems) {
          item.subItems = item.subItems.filter((subItem) => {
            if (subItem.subItems) {
              subItem.subItems = subItem.subItems?.filter((subItem) => {
                return !this.isItemHidden(subItem);
              });
              return subItem.subItems.length > 0;
            }
            return !this.isItemHidden(subItem);
          });

          return item.subItems.length > 0;
        } else {
          return !this.isItemHidden(item);
        }
      });

      return {
        ...subject,
        items: itemsInSubject,
      };
    },
    isItemHidden(item: ISubjectRoutes): boolean {
      if ((item.minPermittedActions && !item.resourceType) || (!item.minPermittedActions && item.resourceType)) {
        throw new Error("minPermittedActions and resourceType must be defined together");
      }

      if (item.hiddenForBranded && this.isWhiteLabel) return true;

      // Check for anyOfResourceTypes (OR logic) - user needs the action on ANY of the resource types
      if (item.anyOfResourceTypes && item.anyOfResourceTypesAction) {
        const hasAnyPermission = this.permissionStore.isActionGrantedForAnyResource(
          item.anyOfResourceTypes,
          item.anyOfResourceTypesAction,
        );
        if (!hasAnyPermission) return true;
      } else {
        // Fallback to original logic - resourceType + minPermittedActions
        const hasPermissions =
          item.minPermittedActions &&
          item.resourceType &&
          this.permissionStore.hasSomePermittedActions(item.resourceType, item.minPermittedActions);

        if (!hasPermissions) return true;
      }

      const isFeatureFlagOff = item.featureFlagName && !this.isFeatureFlagOn(item.featureFlagName);
      const isFeatureToggleOff = item.featureToggleName && !this.isFeatureToggleOn(item.featureToggleName);

      return !!(isFeatureFlagOff || isFeatureToggleOff);
    },
    filterDashboardItems(subject: IMenuItem): IMenuItem {
      const items = subject.items.filter((item) => {
        if (item.subItems) {
          item.subItems = item.subItems.filter((subItem) => {
            switch (subItem.name) {
              case DASHBOARD_ROUTE_NAMES.RESOURCES_INDEX:
                return (
                  !this.isItemHidden(subItem) &&
                  this.permissionStore.isActionGrantedForEveryResource(
                    [ResourceType.Cluster, ResourceType.Department, ResourceType.Project],
                    Action.Read,
                  )
                );
              case DASHBOARD_ROUTE_NAMES.OVERVIEW_INDEX:
                return this.permissionStore.hasOverviewPermission() && !this.isItemHidden(subItem);
              default:
                return !this.isItemHidden(subItem);
            }
          });

          return item.subItems.length > 0;
        }
        if (item.name == ReportRoutesName.ReportIndex) {
          return this.appStore.isReportApiAvailable && !this.isItemHidden(item);
        }
        return !this.isItemHidden(item);
      });

      return {
        ...subject,
        items,
      };
    },
    filterWorkloadManagerItems(subject: IMenuItem): IMenuItem {
      const filteredSubject = this.filterHiddenItems(deepCopy(subject));
      if (this.settingStore.isDataVolumeEnabled) {
        const assetsItem = filteredSubject.items.find((item) => item.name === "assets");
        if (assetsItem) {
          assetsItem.subItems = assetsItem.subItems?.filter((subItem) => {
            return subItem.name !== DATA_SOURCE_ROUTE_NAMES.DATA_SOURCE_INDEX;
          });
        }
      }

      // Filter AI Applications if no cluster supports it (< 2.23)
      filteredSubject.items = filteredSubject.items.filter((item) => {
        return this.shouldShowAIApplicationsItem(item);
      });

      return filteredSubject;
    },
    shouldShowAIApplicationsItem(item: ISubjectRoutes): boolean {
      if (item.name === AI_APPLICATION_ROUTE_NAMES.AI_APPLICATION_INDEX) {
        // Check both Unleash flag and cluster version
        const isFeatureEnabled = !item.featureToggleName || this.isFeatureToggleOn(item.featureToggleName);
        const hasCompatibleCluster = this.clusterStore.isAtleastOneClusterSufficientForVersion(
          MIN_CLUSTER_VERSION_FOR_AI_APPLICATIONS,
        );
        return isFeatureEnabled && hasCompatibleCluster;
      }
      return true;
    },
    isFeatureFlagOn(featureFlagName: SettingKeys): boolean {
      return this.settingStore.isFeatureEnabled(featureFlagName);
    },
    isFeatureToggleOn(featureName: EFeatureToggle): boolean {
      return unleashService.isFeatureEnabled(featureName);
    },
  },
});
</script>

<style lang="scss">
.runai-side-nav {
  box-shadow: var(--shadow-s);
}
</style>
<style scoped lang="scss">
.entire-side-navigation {
  $mini-width: 74px;
  $rest-width: 215px;
  height: 100%;

  .runai-corner-logo {
    padding: 0;
    min-height: $header-height;
    height: $header-height;
    display: flex;
    justify-content: center;
    align-items: center;
    border-bottom: 1px solid rgba(var(--menu-separator-color), var(--menu-separator-opacity));
    width: $mini-width;
  }

  .left-side-navigation {
    @extend .bg-menu-clr;
    height: 100%;
    width: $mini-width;
    display: flex;
    justify-content: space-between;
    flex-direction: column;
    align-items: center;
    flex-wrap: nowrap;

    .menu-items-container {
      overflow: auto;
      display: flex;
      flex-direction: column;
      gap: 5px;

      &::-webkit-scrollbar {
        width: 8px;
        height: 8px;
      }

      &::-webkit-scrollbar-thumb {
        background-color: $primary;
        border-radius: 4px;
      }
    }
  }

  .toggle-drawer-btn {
    background-color: $side-nav-selected-issue-background;
  }

  .toggle-btn {
    transform: rotate(0deg);
    transition: all 0.3s ease-in-out;

    &.open {
      transform: rotate(-180deg);
    }
  }

  .expanded-section {
    display: flex;
    justify-content: center;
    background-color: var(--nav-secondary-bg-default);
    border-left: 1px solid var(--border-subtle);
    width: $rest-width;
    flex-wrap: nowrap;
    overflow-y: auto;
    overflow-x: hidden;

    &::-webkit-scrollbar {
      width: 8px;
      height: 8px;
    }

    &::-webkit-scrollbar-thumb {
      background-color: var(--icon-default);
      border-radius: 4px;
    }

    &.logo-hidden {
      padding-top: $header-height;
    }
  }
}
</style>
