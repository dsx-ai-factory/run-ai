/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { App } from "vue";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { ITheme } from "@/models/theme.model";

const mocks = vi.hoisted(() => ({
  setDistributionFavicon: vi.fn(),
  setTheme: vi.fn(),
}));

vi.mock("@/stores/app.store", () => ({
  useAppStore: () => ({ setTheme: mocks.setTheme }),
}));

vi.mock("@/services/infra/distribution-assets.service", () => ({
  DISTRIBUTION_LOGO_ASSETS: {
    VendorLogoOnDark: "vendor-logo-on-dark.svg",
    VendorLogoOnLight: "vendor-logo-on-light.svg",
  },
  PRODUCT_LOGO_URL: "/distribution-assets/product-logo.svg",
  getDistributionLogoUrl: (asset: string, profile: string) => `/distribution-assets/${profile}/${asset}`,
  getFaviconUrl: (profile: string) => `/distribution-assets/${profile}/favicon.ico`,
  setDistributionFavicon: mocks.setDistributionFavicon,
}));

import { getActiveTheme, theme } from "@/plugins/theme/theme";

const customTheme: ITheme = {
  name: "Custom theme",
  isWhiteLabel: true,
  color: {
    primaryColor: "#123456",
    secondaryColor: "#654321",
    secondaryColorOpacity: "0.25",
    menuColor: "#112233",
    menuSeparatorColor: "1, 2, 3",
    menuSeparatorOpacity: "0.4",
  },
  favicon: "/custom/favicon.ico",
  logo: {
    main: {
      light: { src: "/custom/light.svg", height: "20px" },
      dark: { src: "/custom/dark.svg", height: "18px" },
    },
    sub: "/custom/sub.svg",
  },
};

describe("getActiveTheme", () => {
  it.each(["cc", "ce"] as const)(
    "creates the complete %s profile theme when no custom theme is configured",
    (profile) => {
      expect(getActiveTheme(profile)).toMatchObject({
        favicon: `/distribution-assets/${profile}/favicon.ico`,
        logo: {
          main: {
            light: { src: `/distribution-assets/${profile}/vendor-logo-on-dark.svg`, height: "9px" },
            dark: { src: `/distribution-assets/${profile}/vendor-logo-on-light.svg`, height: "17px" },
          },
          sub: "/distribution-assets/product-logo.svg",
        },
      });
    },
  );

  it.each(["cc", "ce"] as const)("uses a complete custom theme unchanged for the %s profile", (profile) => {
    expect(getActiveTheme(profile, customTheme)).toBe(customTheme);
  });
});

describe("theme", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("applies the final theme supplied by the caller without replacing its settings", () => {
    theme.install({} as App, customTheme);

    expect(document.title).toBe(customTheme.name);
    expect(document.documentElement.style.getPropertyValue("--primary-color")).toBe("#123456");
    expect(mocks.setDistributionFavicon).toHaveBeenCalledWith(customTheme.favicon);
    expect(mocks.setTheme).toHaveBeenCalledWith(customTheme);
  });
});
