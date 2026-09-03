/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import type { App } from "vue";
import { useAppStore } from "@/stores/app.store";
import type { ITheme } from "@/models/theme.model";
import {
  DISTRIBUTION_LOGO_ASSETS,
  PRODUCT_LOGO_URL,
  getDistributionLogoUrl,
  getFaviconUrl,
  setDistributionFavicon,
  type DistributionAssetsProfile,
} from "@/services/infra/distribution-assets.service";

// Replace undefined with a complete custom theme for a downstream build.
const ACTIVE_THEME: ITheme | undefined = undefined;

const DEFAULT_THEME = {
  name: "Run:ai",
  isWhiteLabel: false,
  color: {
    primaryColor: "#65a800",
    secondaryColor: "#65a800",
    secondaryColorOpacity: "0.15",
    menuColor: "#000000",
    menuSeparatorColor: "255, 255, 255",
    menuSeparatorOpacity: "0.3",
  },
} satisfies Omit<ITheme, "favicon" | "logo">;

/** Returns the complete custom theme when configured, or the selected profile theme otherwise. */
export function getActiveTheme(
  profile: DistributionAssetsProfile,
  activeTheme: ITheme | undefined = ACTIVE_THEME,
): ITheme {
  if (activeTheme) return activeTheme;

  return {
    ...DEFAULT_THEME,
    favicon: getFaviconUrl(profile),
    logo: {
      main: {
        light: {
          src: getDistributionLogoUrl(DISTRIBUTION_LOGO_ASSETS.VendorLogoOnDark, profile),
          height: "9px",
        },
        dark: {
          src: getDistributionLogoUrl(DISTRIBUTION_LOGO_ASSETS.VendorLogoOnLight, profile),
          height: "17px",
        },
      },
      sub: PRODUCT_LOGO_URL,
    },
  };
}

function hexToRgb(hex: string): string {
  if (!/^#?[0-9A-Fa-f]{6}$/.test(hex)) {
    console.error("Invalid hex color:", hex);
    return "0, 0, 0";
  }
  hex = hex.replace(/^#/, "");
  const bigint = parseInt(hex, 16);
  const r = (bigint >> 16) & 255;
  const g = (bigint >> 8) & 255;
  const b = bigint & 255;
  return `${r}, ${g}, ${b}`;
}

function setColors(theme: ITheme): void {
  document.documentElement.style.setProperty("--primary-color", theme.color.primaryColor);
  document.documentElement.style.setProperty("--primary-color-rgb", hexToRgb(theme.color.primaryColor));
  document.documentElement.style.setProperty("--secondary-color", theme.color.secondaryColor);
  document.documentElement.style.setProperty("--secondary-color-rgb", hexToRgb(theme.color.secondaryColor));
  document.documentElement.style.setProperty("--secondary-color-opacity", theme.color.secondaryColorOpacity);
  document.documentElement.style.setProperty("--menu-color", theme.color.menuColor);
  document.documentElement.style.setProperty("--menu-separator-color", theme.color.menuSeparatorColor);
  document.documentElement.style.setProperty("--menu-separator-opacity", theme.color.menuSeparatorOpacity);
}

function setTitle(theme: ITheme): void {
  document.title = theme.name;
}

export const theme = {
  install(_app: App, finalTheme: ITheme) {
    const appStore = useAppStore();

    setTitle(finalTheme);
    setDistributionFavicon(finalTheme.favicon);
    setColors(finalTheme);
    appStore.setTheme(finalTheme);
  },
};
