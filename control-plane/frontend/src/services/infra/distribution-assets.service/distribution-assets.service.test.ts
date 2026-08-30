import { beforeEach, describe, expect, it, vi } from "vitest";

const bootConfigState = vi.hoisted(() => ({ profile: "cc" }));

vi.mock("@/services/infra/boot-config.service/boot-config.service", () => ({
  getBootConfig: () => ({ distributionAssets: { profile: bootConfigState.profile } }),
}));

import {
  CC_DISTRIBUTION_ASSETS_PROFILE,
  CC_FAVICON_URL,
  CE_DISTRIBUTION_ASSETS_PROFILE,
  CE_FAVICON_URL,
  DISTRIBUTION_LOGO_ASSETS,
  PRODUCT_LOGO_URL,
  fallBackDistributionLogo,
  getDistributionAssetsProfile,
  getDistributionLogoUrl,
  getFaviconUrl,
  setDistributionFavicon,
  type DistributionAssetsProfile,
} from "./distribution-assets.service";

describe("distribution assets service", () => {
  beforeEach(() => {
    bootConfigState.profile = CC_DISTRIBUTION_ASSETS_PROFILE;
    document.querySelectorAll('link[rel="icon"]').forEach((link) => link.remove());
  });

  it.each([
    [CC_DISTRIBUTION_ASSETS_PROFILE, CC_FAVICON_URL],
    [CE_DISTRIBUTION_ASSETS_PROFILE, CE_FAVICON_URL],
  ])("selects the favicon for the %s profile", (profile, expectedUrl) => {
    bootConfigState.profile = profile;

    expect(getDistributionAssetsProfile()).toBe(profile);
    expect(getFaviconUrl()).toBe(expectedUrl);
  });

  it.each(["", "custom", "../ce"])("uses the CC profile for an unknown value %s", (profile) => {
    bootConfigState.profile = profile;

    expect(getDistributionAssetsProfile()).toBe(CC_DISTRIBUTION_ASSETS_PROFILE);
    expect(getFaviconUrl()).toBe(CC_FAVICON_URL);
  });

  it("falls back once to the default favicon when the CE asset is unavailable", () => {
    bootConfigState.profile = CE_DISTRIBUTION_ASSETS_PROFILE;
    setDistributionFavicon();

    const link = document.querySelector<HTMLLinkElement>('link[rel="icon"]');
    expect(link?.getAttribute("href")).toBe(CE_FAVICON_URL);

    link?.dispatchEvent(new Event("error"));
    expect(link?.getAttribute("href")).toBe(CC_FAVICON_URL);
    expect(link?.onerror).toBeNull();
  });

  it.each([
    [CC_DISTRIBUTION_ASSETS_PROFILE, "/distribution-assets/cc/vendor-logo-on-light.svg"],
    [CE_DISTRIBUTION_ASSETS_PROFILE, "/distribution-assets/ce/vendor-logo-on-light.svg"],
  ])("selects a stable role-based logo path for the %s profile", (profile, expectedUrl) => {
    expect(
      getDistributionLogoUrl(DISTRIBUTION_LOGO_ASSETS.VendorLogoOnLight, profile as DistributionAssetsProfile),
    ).toBe(expectedUrl);
  });

  it("uses the tightly cropped SVG for the Run:ai product wordmark", () => {
    expect(PRODUCT_LOGO_URL).toBe("/distribution-assets/product-logo.svg");
    expect(DISTRIBUTION_LOGO_ASSETS.VendorLogoOnDark).toBe("vendor-logo-on-dark.svg");
  });

  it("falls back once to the matching CC logo", () => {
    const image = document.createElement("img");
    image.src = getDistributionLogoUrl(DISTRIBUTION_LOGO_ASSETS.VendorLogoOnLight, CE_DISTRIBUTION_ASSETS_PROFILE);
    image.onerror = vi.fn();

    fallBackDistributionLogo({ currentTarget: image } as unknown as Event, DISTRIBUTION_LOGO_ASSETS.VendorLogoOnLight);

    expect(image.getAttribute("src")).toBe("/distribution-assets/cc/vendor-logo-on-light.svg");
    expect(image.onerror).toBeNull();
  });
});
