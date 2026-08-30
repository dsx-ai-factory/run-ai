import type { App } from "vue";
import { beforeEach, describe, expect, it, vi } from "vitest";

const profileState = vi.hoisted(() => ({ profile: "ce" }));

const mocks = vi.hoisted(() => ({
  faviconUrl: "/distribution-assets/ce/favicon.ico",
  getDistributionLogoUrl: vi.fn((asset: string, profile: string) => `/distribution-assets/${profile}/${asset}`),
  setDistributionFavicon: vi.fn(),
  setTheme: vi.fn(),
}));

vi.mock("@/stores/app.store", () => ({
  useAppStore: () => ({ setTheme: mocks.setTheme }),
}));

vi.mock("@/services/infra/distribution-assets.service", () => ({
  CC_DISTRIBUTION_ASSETS_PROFILE: "cc",
  CE_DISTRIBUTION_ASSETS_PROFILE: "ce",
  CC_FAVICON_URL: "/distribution-assets/cc/favicon.ico",
  DISTRIBUTION_LOGO_ASSETS: {
    VendorLogoOnDark: "vendor-logo-on-dark.svg",
    VendorLogoOnLight: "vendor-logo-on-light.svg",
    VendorMarkOnLight: "vendor-mark-on-light.svg",
  },
  PRODUCT_LOGO_URL: "/distribution-assets/product-logo.svg",
  getDistributionAssetsProfile: () => profileState.profile,
  getDistributionLogoUrl: mocks.getDistributionLogoUrl,
  getFaviconUrl: () => mocks.faviconUrl,
  setDistributionFavicon: mocks.setDistributionFavicon,
}));

import { theme } from "./theme";

describe("theme", () => {
  beforeEach(() => {
    profileState.profile = "ce";
    vi.clearAllMocks();
  });

  it("uses the resolved distribution assets in the document and app store theme", () => {
    theme.install({} as App);

    expect(mocks.setDistributionFavicon).toHaveBeenCalledWith(mocks.faviconUrl);
    expect(mocks.setTheme).toHaveBeenCalledWith(
      expect.objectContaining({
        favicon: mocks.faviconUrl,
        logo: {
          main: {
            light: { src: "/distribution-assets/ce/vendor-logo-on-dark.svg", height: "9px" },
            dark: { src: "/distribution-assets/ce/vendor-logo-on-light.svg", height: "17px" },
          },
          sub: "/distribution-assets/product-logo.svg",
        },
      }),
    );
  });

  it("uses the compact CC vendor wordmark for the sidebar theme slot", () => {
    profileState.profile = "cc";

    theme.install({} as App);

    expect(mocks.setTheme).toHaveBeenCalledWith(
      expect.objectContaining({
        logo: expect.objectContaining({
          main: expect.objectContaining({
            light: { src: "/distribution-assets/cc/vendor-logo-on-dark.svg", height: "9px" },
          }),
        }),
      }),
    );
  });
});
