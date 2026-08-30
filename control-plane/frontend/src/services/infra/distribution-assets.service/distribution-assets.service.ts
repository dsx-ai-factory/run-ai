import { getBootConfig } from "@/services/infra/boot-config.service/boot-config.service";

export const CC_DISTRIBUTION_ASSETS_PROFILE = "cc";
export const CE_DISTRIBUTION_ASSETS_PROFILE = "ce";
export const CC_FAVICON_URL = "/distribution-assets/cc/favicon.ico";
export const CE_FAVICON_URL = "/distribution-assets/ce/favicon.ico";
export const PRODUCT_LOGO_URL = "/distribution-assets/product-logo.svg";

export const DISTRIBUTION_LOGO_ASSETS = {
  VendorLogoOnLight: "vendor-logo-on-light.svg",
  VendorLogoOnDark: "vendor-logo-on-dark.svg",
  VendorMarkOnLight: "vendor-mark-on-light.svg",
} as const;

export type DistributionAssetsProfile = typeof CC_DISTRIBUTION_ASSETS_PROFILE | typeof CE_DISTRIBUTION_ASSETS_PROFILE;
export type DistributionLogoAsset = (typeof DISTRIBUTION_LOGO_ASSETS)[keyof typeof DISTRIBUTION_LOGO_ASSETS];

export function getDistributionAssetsProfile(): DistributionAssetsProfile {
  return getBootConfig().distributionAssets.profile === CE_DISTRIBUTION_ASSETS_PROFILE
    ? CE_DISTRIBUTION_ASSETS_PROFILE
    : CC_DISTRIBUTION_ASSETS_PROFILE;
}

export function getFaviconUrl(profile: DistributionAssetsProfile = getDistributionAssetsProfile()): string {
  return profile === CE_DISTRIBUTION_ASSETS_PROFILE ? CE_FAVICON_URL : CC_FAVICON_URL;
}

export function getDistributionLogoUrl(
  asset: DistributionLogoAsset,
  profile: DistributionAssetsProfile = getDistributionAssetsProfile(),
): string {
  return `/distribution-assets/${profile}/${asset}`;
}

export function fallBackDistributionLogo(event: Event, asset: DistributionLogoAsset): void {
  const image = event.currentTarget;
  if (!(image instanceof HTMLImageElement)) return;

  const fallbackAsset =
    asset === DISTRIBUTION_LOGO_ASSETS.VendorMarkOnLight ? DISTRIBUTION_LOGO_ASSETS.VendorLogoOnDark : asset;
  const fallbackUrl = getDistributionLogoUrl(fallbackAsset, CC_DISTRIBUTION_ASSETS_PROFILE);
  if (new URL(image.src, window.location.origin).pathname === fallbackUrl) {
    image.onerror = null;
    return;
  }

  image.onerror = null;
  image.src = fallbackUrl;
}

export function setDistributionFavicon(faviconUrl: string = getFaviconUrl()): void {
  let link: HTMLLinkElement | null = document.querySelector('link[rel="icon"]');
  if (!link) {
    link = document.createElement("link");
    link.rel = "icon";
    document.head.appendChild(link);
  }

  const faviconLink = link;
  faviconLink.onerror =
    faviconUrl === CC_FAVICON_URL
      ? null
      : () => {
          faviconLink.onerror = null;
          faviconLink.href = CC_FAVICON_URL;
        };
  faviconLink.href = faviconUrl;
}
