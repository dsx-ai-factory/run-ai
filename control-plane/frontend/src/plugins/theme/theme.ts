import type { App } from "vue";
import { useAppStore } from "@/stores/app.store";
import type { ITheme } from "@/models/theme.model";
import {
  CC_FAVICON_URL,
  CC_DISTRIBUTION_ASSETS_PROFILE,
  DISTRIBUTION_LOGO_ASSETS,
  PRODUCT_LOGO_URL,
  getDistributionAssetsProfile,
  getDistributionLogoUrl,
  getFaviconUrl,
  setDistributionFavicon,
} from "@/services/infra/distribution-assets.service";

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

const RUNAI_THEME: ITheme = {
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
  favicon: CC_FAVICON_URL,
  logo: {
    main: {
      light: {
        src: getDistributionLogoUrl(DISTRIBUTION_LOGO_ASSETS.VendorLogoOnDark, CC_DISTRIBUTION_ASSETS_PROFILE),
        height: "9px",
      },
      dark: {
        src: getDistributionLogoUrl(DISTRIBUTION_LOGO_ASSETS.VendorLogoOnLight, CC_DISTRIBUTION_ASSETS_PROFILE),
        height: "17px",
      },
    },
    sub: PRODUCT_LOGO_URL,
  },
};

export const theme = {
  install(_app: App) {
    const appStore = useAppStore();
    const profile = getDistributionAssetsProfile();
    const resolvedTheme: ITheme = {
      ...RUNAI_THEME,
      favicon: getFaviconUrl(),
      logo: {
        ...RUNAI_THEME.logo,
        main: {
          light: {
            ...RUNAI_THEME.logo.main.light,
            src: getDistributionLogoUrl(DISTRIBUTION_LOGO_ASSETS.VendorLogoOnDark, profile),
            height: "9px",
          },
          dark: {
            ...RUNAI_THEME.logo.main.dark,
            src: getDistributionLogoUrl(DISTRIBUTION_LOGO_ASSETS.VendorLogoOnLight, profile),
          },
        },
      },
    };

    setTitle(resolvedTheme);
    setDistributionFavicon(resolvedTheme.favicon);
    setColors(resolvedTheme);
    appStore.setTheme(resolvedTheme);
  },
};
