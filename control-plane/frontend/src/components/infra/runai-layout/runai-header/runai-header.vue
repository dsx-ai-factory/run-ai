<template>
  <q-header class="runai-header row items-center" :class="{ 'with-shadow': fullWidthPage }">
    <q-toolbar class="header-toolbar">
      <div v-if="fullWidthPage" class="full-page-actions row items-center q-gutter-x-sm q-mr-sm">
        <template v-if="showCloseButton">
          <back-or-close-btn :disable="shouldDisableBackBtn" />
          <div class="separator">
            <q-separator vertical />
          </div>
        </template>
        <template v-if="!hasCustomHeaderLeft">
          <runai-brand-logo v-if="isWhiteLabel" :img="mainLogoDark.src" :height="mainLogoDark.height" />
          <runai-distribution-logo v-else v-bind="distributionLogoPreset" class="full-page-distribution-logo" />
        </template>
      </div>
      <div id="header-left-teleport" />
      <div v-if="!hasCustomHeaderLeft" class="left-actions row items-center">
        <template v-if="subLogo">
          <img v-if="isWhiteLabel" :src="subLogo" alt="logo" class="sub-logo q-mr-sm" />
          <runai-distribution-logo
            v-else
            composition="product"
            width="45px"
            height="12px"
            gap="0"
            class="sub-logo q-mr-sm"
          />
        </template>
        <runai-breadcrumbs :items="breadcrumbsItems" :full-path="breadcrumbsFullPath" :subtitle="breadcrumbsSubtitle" />
        <runai-stage-badge class="q-ml-md" v-if="stageBadge" display-tooltip :badge="stageBadge" />
        <q-badge
          v-if="displayViewOnlyMessage"
          aid="view-only-badge"
          class="view-only-badge q-ml-md q-py-xs q-px-md"
          id="view-only-badge"
          outline
        >
          <q-icon name="ph-bold ph-eye" class="q-mr-sm" size="16px" />
          <span>View only</span>
          <q-tooltip>You can only view this page</q-tooltip>
        </q-badge>
      </div>

      <q-space />

      <div class="right-actions row items-center q-gutter-x-md">
        <help-dropdown class="help-dropdown-container" />
        <div v-if="tenantLogo && !themeLogo" class="tenant-logo">
          <img :src="tenantLogo" alt="tenant logo" />
        </div>
        <div v-else-if="!themeLogo" class="tenant-name text-b-body1">{{ tenantName }}</div>
        <profile-dropdown light-dark-mode />
      </div>
    </q-toolbar>
  </q-header>
</template>
<script lang="ts">
import { defineComponent, type PropType } from "vue";

// cmps
import { HelpDropdown } from "../../top-header/help-dropdown";
import { ProfileDropdown } from "../../top-header/profile-dropdown";
import { RunaiStageBadge } from "@/components/common/runai-stage-badge";
import { RunaiBreadcrumbs } from "@/components/common/runai-breadcrumbs";
import { BackOrCloseBtn } from "@/components/infra/top-header/back-or-close-btn";
import { RunaiBrandLogo } from "@/components/common/runai-brand-logo";
import { DISTRIBUTION_LOGO_PRESETS, RunaiDistributionLogo } from "@/components/common/runai-distribution-logo";

// stores
import { useAuthStore } from "@/stores/auth.store";
import { usePermissionStore } from "@/stores/permissions.store";
import { useAppStore } from "@/stores/app.store";

// models
import type { ResourceType } from "@/swagger-models/authorization-client";
import { Action } from "@/swagger-models/authorization-client";
import type { EBadgeLabel } from "@/models/setting.model";
import { LogoConstants } from "@/models/setting.model";
import type { RunaiRouteRecord } from "vue-router";
import { menuSubjects, type IMenuItem } from "@/components/infra/runai-layout/routes-model/routes-model";

// routes
import { dashboardRoutes } from "@/router/dashboard.routes/dashboard.routes";

// utils
import { runaiHeaderUtil } from "./runai-header.util";
import { imageUtil } from "@/utils/image.util";
import { CLUSTER_ROUTE_NAMES } from "@/router/cluster.routes/cluster.routes.names";
import { useClusterStore } from "@/stores/cluster.store";

export default defineComponent({
  name: "runai-header",
  components: {
    RunaiStageBadge,
    BackOrCloseBtn,
    HelpDropdown,
    ProfileDropdown,
    RunaiBreadcrumbs,
    RunaiBrandLogo,
    RunaiDistributionLogo,
  },
  props: {
    breadcrumbsFullPath: {
      type: Boolean as PropType<boolean>,
      required: true,
    },
  },
  data() {
    return {
      appStore: useAppStore(),
      authStore: useAuthStore(),
      permissionStore: usePermissionStore(),
      allSubjects: Object.values(menuSubjects) as IMenuItem[],
      distributionLogoPreset: DISTRIBUTION_LOGO_PRESETS.fullPageHeader,
    };
  },
  computed: {
    subLogo(): string | undefined {
      if (!this.breadcrumbsFullPath || this.fullWidthPage) return undefined;
      return this.appStore.theme?.logo.sub;
    },
    shouldDisableBackBtn(): boolean {
      const { isClustersLoaded, clusters } = useClusterStore();
      const isNewClusterRoute = this.$route.name === CLUSTER_ROUTE_NAMES.CLUSTER_NEW;
      return isNewClusterRoute && isClustersLoaded && !clusters.length;
    },
    breadcrumbsItems(): string[] {
      const pageTitle = this.appStore.pageTitleOverride || this.$route.meta.pageTitle?.toString() || "";
      const path: string[] | null = runaiHeaderUtil.findLabelPath(this.allSubjects, pageTitle);
      return path ? path : [pageTitle];
    },
    breadcrumbsSubtitle(): string | undefined {
      return this.$route.meta.subtitle as string | undefined;
    },
    tenantLogo(): string | undefined {
      if (this.isTransparentPlaceholderLogo(this.authStore.logo)) {
        return undefined;
      }
      return imageUtil.createBase64DataUrl(this.authStore.logo);
    },
    tenantName(): string {
      return this.authStore.tenant?.displayName || this.appStore.config?.tenantName || "";
    },
    stageBadge(): EBadgeLabel | null {
      return (this.$route.meta.stageBadge as EBadgeLabel) || null;
    },
    displayViewOnlyMessage(): boolean {
      if (this.$route.meta.showViewOnlyBadge) return true;
      const currentResourceType: ResourceType | ResourceType[] | null = this.$route.meta.resourceType ?? null;
      if (!currentResourceType) return false;
      const routeName = this.$router.currentRoute.value.name;
      if (
        dashboardRoutes.children &&
        dashboardRoutes.children.some((route: RunaiRouteRecord) => route.name === routeName)
      ) {
        return false;
      }
      return !this.permissionStore.hasSomePermittedActions(currentResourceType, [
        Action.Update,
        Action.Create,
        Action.Delete,
      ]);
    },
    fullWidthPage(): boolean {
      return this.$route.meta.fullWidth as boolean;
    },
    hasCustomHeaderLeft(): boolean {
      return !!this.$route.meta.customHeaderLeft;
    },
    mainLogoDark(): { src: string; height: string } {
      return this.appStore.mainLogoDark;
    },
    themeLogo(): string | undefined {
      return this.isWhiteLabel ? this.mainLogoDark.src : undefined;
    },
    showCloseButton(): boolean {
      return !this.$route.meta.hideCloseButton;
    },
    isWhiteLabel(): boolean {
      return this.appStore.isWhiteLabel;
    },
  },
  methods: {
    isTransparentPlaceholderLogo(logoDataUrl: string | undefined): boolean {
      if (!logoDataUrl) return true;

      const base64Part = logoDataUrl.split(",")[1];
      return base64Part === LogoConstants.TRANSPARENT_PNG_BASE64;
    },
  },
});
</script>
<style scoped lang="scss">
.runai-header {
  height: $header-height;
  background-color: var(--bg-primary-default);
  color: var(--b-text-main);
  border-bottom: 1px solid var(--border-subtle);

  &.with-shadow {
    box-shadow: $runai-table-shadow;
  }

  .header-toolbar {
    padding: 0 var(--space-m);
    height: 100%;
  }

  .full-page-distribution-logo {
    margin-bottom: 2px;

    &[data-profile="ce"] {
      margin-left: 30px;
      overflow: visible;
    }
  }

  .separator {
    height: 50px;
    padding: 9px 0;

    hr {
      height: 100%;
    }
  }

  .view-only-badge {
    color: var(--b-text-sub);
    border-color: var(--border-strong);
  }

  .right-actions {
    .tenant-name {
      border: 1px solid var(--border-subtle);
      border-radius: var(--b-border-radius-radius-xs);
      text-transform: uppercase;
      white-space: nowrap;
      height: 36px;
      display: flex;
      align-items: center;
      padding: 0 var(--space-s);
    }

    .right-btns {
      gap: 15px;
    }

    .tenant-logo {
      height: 36px;
      display: flex;
      justify-content: center;
      align-items: center;
      border: 1px solid var(--border-subtle);
      border-radius: var(--b-border-radius-radius-xs);
      padding: 6px;

      img {
        max-height: 24px;
        max-width: 120px;
      }
    }

    .help-dropdown-container {
      margin-right: -8px;
    }
  }
}
</style>
