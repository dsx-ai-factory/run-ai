import { shallowMount } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";

const storeState = vi.hoisted(() => ({
  isWhiteLabel: false,
}));

const distributionState = vi.hoisted(() => ({
  profile: "cc",
}));

vi.mock("@/stores/app.store", () => ({
  useAppStore: () => ({
    isWhiteLabel: storeState.isWhiteLabel,
    isReportApiAvailable: true,
    mainLogoLight: { src: "/distribution-assets/cc/vendor-logo-on-dark.svg", height: "9px" },
    theme: { logo: { sub: "/distribution-assets/product-logo.svg" } },
  }),
}));

vi.mock("@/stores/setting.store", () => ({
  useSettingStore: () => ({
    isDataVolumeEnabled: false,
    isFeatureEnabled: () => true,
  }),
}));

vi.mock("@/stores/permissions.store", () => ({
  usePermissionStore: () => ({
    hasOverviewPermission: () => true,
    hasSomePermittedActions: () => true,
    isActionGrantedForAnyResource: () => true,
    isActionGrantedForEveryResource: () => true,
  }),
}));

vi.mock("@/stores/cluster.store", () => ({
  useClusterStore: () => ({
    isAtleastOneClusterSufficientForVersion: () => true,
  }),
}));

vi.mock("@/services/infra/unleash.service/unleash.service", () => ({
  unleashService: { isFeatureEnabled: () => true },
}));

vi.mock("@/services/infra/distribution-assets.service", () => ({
  CC_DISTRIBUTION_ASSETS_PROFILE: "cc",
  DISTRIBUTION_LOGO_ASSETS: {
    VendorLogoOnDark: "vendor-logo-on-dark.svg",
    VendorMarkOnLight: "vendor-mark-on-light.svg",
  },
  getDistributionAssetsProfile: () => distributionState.profile,
}));

import { RunaiDistributionLogo } from "@/components/common/runai-distribution-logo";
import { RunaiBrandLogo } from "@/components/common/runai-brand-logo";
import RunaiSideNav from "./runai-side-nav.vue";

function mountSideNav() {
  return shallowMount(RunaiSideNav, {
    props: { initialState: false },
    global: {
      mocks: {
        $route: { meta: { parentSubject: "analytics" } },
        $router: { push: vi.fn() },
      },
      stubs: {
        QDrawer: { template: "<aside><slot /></aside>" },
        QBtn: { template: "<button><slot /></button>" },
        QIcon: true,
        QTooltip: true,
      },
    },
  });
}

describe("runai side navigation branding", () => {
  beforeEach(() => {
    storeState.isWhiteLabel = false;
    distributionState.profile = "cc";
  });

  it("renders distribution vendor and product logos", () => {
    const wrapper = mountSideNav();

    expect(wrapper.findAllComponents(RunaiDistributionLogo)).toHaveLength(2);
  });

  it("uses the compact white source asset for the CC navigation rail", () => {
    const wrapper = mountSideNav();
    const vendorLogo = wrapper.findAllComponents(RunaiDistributionLogo)[0];

    expect(vendorLogo.props("vendorAsset")).toBe("vendor-logo-on-dark.svg");
    expect(vendorLogo.props("textColor")).toBe("default");
    expect(vendorLogo.props("vendorHeight")).toBe("9px");
  });

  it("uses the CE dark-surface asset directly instead of recoloring the green mark", () => {
    distributionState.profile = "ce";
    const wrapper = mountSideNav();
    const vendorLogo = wrapper.findAllComponents(RunaiDistributionLogo)[0];

    expect(vendorLogo.props("vendorAsset")).toBe("vendor-logo-on-dark.svg");
    expect(vendorLogo.props("textColor")).toBe("white");
    expect(vendorLogo.props("vendorHeight")).toBe("9px");
    expect(vendorLogo.props("vendorCanvasHeight")).toBe("9px");
    expect(vendorLogo.props("vendorCanvasTop")).toBe("0");
  });

  it("preserves white-label branches with theme URLs backed by distribution assets", () => {
    storeState.isWhiteLabel = true;
    const wrapper = mountSideNav();

    expect(wrapper.findComponent(RunaiDistributionLogo).exists()).toBe(false);
    expect(wrapper.findComponent(RunaiBrandLogo).props("img")).toBe("/distribution-assets/cc/vendor-logo-on-dark.svg");
    expect(wrapper.find("img.sub-logo").attributes("src")).toBe("/distribution-assets/product-logo.svg");
  });
});
