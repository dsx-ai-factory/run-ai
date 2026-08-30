import { defineStore } from "pinia";

// models
import { ResourceType, Action } from "@/swagger-models/authorization-client";
import type { IConfig } from "@/models/config.model";
import { ESidebarState } from "@/models/global.model";
import type { ITheme } from "@/models/theme.model";

// stores
import { usePermissionStore } from "./permissions.store";

// services
import { urlService } from "@/services/url.service/url.service";
import { configService } from "@/services/control-plane/config.service/config.service";
import { orgUnitService } from "@/services/control-plane/org-unit.service/org-unit.service";

// utils
import { storageUtil } from "@/utils/storage.util";

import { ADMIN_ONBOARDING_MODAL_SHOWN } from "@/common/storage.constant";
// routes
import { LANDING_PAGE_NAME } from "@/router/dashboard.routes/dashboard.routes.names";
import { WORKLOAD_ROUTE_NAMES } from "@/router/workloads.routes/workloads.routes.names";

import { CLUSTER_ROUTE_NAMES } from "@/router/cluster.routes/cluster.routes.names";

import { DOCS } from "@/constants/docs";
import { getBootConfig } from "@/services/infra/boot-config.service/boot-config.service";

export const useAppStore = defineStore("App", {
  state: () => ({
    appVersion: getBootConfig().version.display as string,
    config: {} as IConfig,
    backendBaseUrl: urlService.getBackendBaseUrl(),
    pageLoading: false as boolean,
    fallback: false as boolean,
    parentAppOrigin: "" as string,
    lastBackPath: "" as string,
    appLoaded: false as boolean,
    sidebarState: ESidebarState.Mini as ESidebarState,
    isAppHeaderDisabled: false as boolean,
    isReportApiAvailable: false as boolean,
    isOnboardingModalClosed: storageUtil.get<boolean>(ADMIN_ONBOARDING_MODAL_SHOWN) ?? false,
    theme: {
      name: "",
      isWhiteLabel: false,
      color: {
        primaryColor: "",
        secondaryColor: "",
        secondaryColorOpacity: "",
      },
      logo: {
        main: {
          light: { src: "", height: "" },
          dark: { src: "", height: "" },
        },
        sub: "",
      },
    } as ITheme,
    pageTitleOverride: null as string | null,
  }),
  getters: {
    version(): string {
      return this.appVersion;
    },
    supportUrl(): string {
      return DOCS.SUPPORT.URL;
    },
    isMultiTenant(): boolean {
      return this.config.isMultiTenant;
    },
    tenantDomain(): string {
      return this.config.tenantName?.concat(".", window.location.hostname).replace("app.run.ai", "run.ai");
    },
    isProductionEnv(): boolean {
      return !["localhost", "test.run.ai"].some((h) => window.location.hostname.startsWith(h));
    },
    isPageLoading(): boolean {
      return this.pageLoading;
    },
    isFallback(): boolean {
      return this.fallback;
    },
    tenantName(): string {
      return this.config.tenantName;
    },
    isThemed(): boolean {
      return !!this.theme;
    },
    isWhiteLabel(): boolean {
      return this.theme?.isWhiteLabel ?? false;
    },
    mainLogoDark(): { src: string; height: string } {
      return this.theme.logo.main.dark;
    },
    mainLogoLight(): { src: string; height: string } {
      return this.theme.logo.main.light;
    },
  },
  actions: {
    async loadAppData(): Promise<void> {
      try {
        this.config = await configService.get();
      } catch (error: unknown) {
        console.error("Failed to load init data: ", error);
        console.log("re-trying to load init data...");
        this.config = await configService.get();
      }
    },
    setPageLoading(pageLoading: boolean): void {
      this.pageLoading = pageLoading;
    },
    setFallback(fallbackStatus: boolean): void {
      this.fallback = fallbackStatus;
    },
    setLastBackPath(path: string): void {
      this.lastBackPath = path;
    },

    setAppLoaded(appLoaded: boolean): void {
      this.appLoaded = appLoaded;
    },
    setSidebarState(sidebarState: ESidebarState): void {
      this.sidebarState = sidebarState;
    },
    setIsAppHeaderDisabled(val: boolean): void {
      this.isAppHeaderDisabled = val;
    },
    getUserLandingPage(routeName?: string): string | undefined {
      const permissionStore = usePermissionStore();
      if (routeName && routeName !== LANDING_PAGE_NAME) return routeName;

      if (permissionStore.hasReadNodesPermission()) {
        return LANDING_PAGE_NAME;
      } else if (permissionStore.hasPermission(ResourceType.Workloads, Action.Read)) {
        return WORKLOAD_ROUTE_NAMES.WORKLOAD_INDEX;
      } else {
        return CLUSTER_ROUTE_NAMES.CLUSTER_INDEX;
      }
    },
    async loadReportApiAvailability(): Promise<void> {
      try {
        this.isReportApiAvailable = await orgUnitService.isReportApiAvailable();
      } catch (error: unknown) {
        console.error("Failed to load report api availability: ", error);
        this.isReportApiAvailable = false;
      }
    },
    setTheme(theme: ITheme): void {
      this.theme = theme;
    },
    setIsOnboardingModalClosed(isShown: boolean): void {
      this.isOnboardingModalClosed = isShown;
    },
    setPageTitleOverride(title: string | null): void {
      this.pageTitleOverride = title;
    },
  },
});
