<template>
  <span
    class="runai-distribution-logo"
    :style="frameStyle"
    role="img"
    :aria-label="alt"
    :data-composition="composition"
    :data-profile="profile"
    :data-text-color="textColor"
  >
    <svg
      v-if="usesVendorWhiteTextFilter"
      class="logo-color-filter"
      width="0"
      height="0"
      aria-hidden="true"
      focusable="false"
    >
      <defs>
        <filter :id="vendorWhiteTextFilterId" color-interpolation-filters="sRGB">
          <feColorMatrix type="matrix" :values="VENDOR_WHITE_TEXT_COLOR_MATRIX" />
        </filter>
      </defs>
    </svg>
    <span v-if="showsVendorLogo" class="vendor-logo-viewport" :style="vendorViewportStyle">
      <img
        class="vendor-logo-canvas"
        :class="{
          'vendor-logo-horizontal-ce': usesCeHorizontalVendorGeometry,
          'vendor-logo-fallback-cc': usesCcVendorFallback,
        }"
        :src="vendorLogoUrl"
        :style="vendorLogoCanvasStyle"
        alt=""
        draggable="false"
        @error="handleVendorLogoError"
      />
    </span>
    <img
      v-if="showsProductLogo"
      class="product-logo"
      :src="PRODUCT_LOGO_URL"
      :style="productLogoStyle"
      alt=""
      draggable="false"
    />
  </span>
</template>

<script setup lang="ts">
import { computed, ref, useId } from "vue";
import {
  CC_DISTRIBUTION_ASSETS_PROFILE,
  CE_DISTRIBUTION_ASSETS_PROFILE,
  DISTRIBUTION_LOGO_ASSETS,
  PRODUCT_LOGO_URL,
  fallBackDistributionLogo,
  getDistributionAssetsProfile,
  getDistributionLogoUrl,
  type DistributionLogoAsset,
} from "@/services/infra/distribution-assets.service";

defineOptions({
  name: "runai-distribution-logo",
});

type LogoComposition = "vendor" | "product" | "combined";
type LogoTextColor = "default" | "white";

const VENDOR_WHITE_TEXT_COLOR_MATRIX = `
  0 -0.74054054 0 0 1
  0 -0.37837838 0 0 1
  0 -1.37837838 0 0 1
  0 0 0 1 0
`;

const props = withDefaults(
  defineProps<{
    composition?: LogoComposition;
    width: string;
    height: string;
    vendorWidth?: string;
    vendorHeight?: string;
    vendorCanvasHeight?: string;
    vendorCanvasTop?: string;
    productWidth?: string;
    productHeight?: string;
    ccProductWidth?: string;
    ccProductHeight?: string;
    ccVendorHeight?: string;
    gap?: string;
    vendorAsset?: DistributionLogoAsset;
    textColor?: LogoTextColor;
    disableProductOffset?: boolean;
    omitCcWidth?: boolean;
    omitCcVendorViewportHeight?: boolean;
    alt?: string;
  }>(),
  {
    composition: "combined",
    vendorWidth: "96px",
    vendorHeight: "17px",
    vendorCanvasHeight: "366.49%",
    vendorCanvasTop: "-133.24%",
    productWidth: "45px",
    productHeight: "12px",
    ccProductWidth: "",
    ccProductHeight: "",
    ccVendorHeight: "",
    gap: "6px",
    vendorAsset: DISTRIBUTION_LOGO_ASSETS.VendorLogoOnLight,
    textColor: "default",
    disableProductOffset: false,
    omitCcWidth: false,
    omitCcVendorViewportHeight: false,
    alt: "NVIDIA Run:ai",
  },
);

const profile = getDistributionAssetsProfile();
const usesCcVendorGeometry = ref(profile === CC_DISTRIBUTION_ASSETS_PROFILE);
const usesCcVendorFallback = ref(false);
const usesDarkSurfaceFallback = ref(false);
const vendorWhiteTextFilterId = `runai-distribution-logo-white-text-${useId()}`;
const showsVendorLogo = computed(() => props.composition === "vendor" || props.composition === "combined");
const showsProductLogo = computed(() => props.composition === "product" || props.composition === "combined");
const usesWhiteText = computed(() => props.textColor === "white");
const usesCeHorizontalVendorGeometry = computed(
  () =>
    profile === CE_DISTRIBUTION_ASSETS_PROFILE &&
    props.vendorAsset !== DISTRIBUTION_LOGO_ASSETS.VendorMarkOnLight &&
    !usesCcVendorGeometry.value,
);
const usesVendorWhiteTextFilter = computed(
  () =>
    usesWhiteText.value &&
    !usesCeHorizontalVendorGeometry.value &&
    props.vendorAsset !== DISTRIBUTION_LOGO_ASSETS.VendorLogoOnDark &&
    !usesDarkSurfaceFallback.value,
);
const vendorLogoUrl = computed(() =>
  usesCeHorizontalVendorGeometry.value
    ? getDistributionLogoUrl(
        usesWhiteText.value ? DISTRIBUTION_LOGO_ASSETS.VendorLogoOnDark : DISTRIBUTION_LOGO_ASSETS.VendorLogoOnLight,
        profile,
      )
    : getDistributionLogoUrl(props.vendorAsset, profile),
);
const omitsFrameWidth = computed(() => profile === CC_DISTRIBUTION_ASSETS_PROFILE && props.omitCcWidth);
const productWidth = computed(() =>
  profile === CC_DISTRIBUTION_ASSETS_PROFILE && props.ccProductWidth ? props.ccProductWidth : props.productWidth,
);
const productHeight = computed(() =>
  profile === CC_DISTRIBUTION_ASSETS_PROFILE && props.ccProductHeight ? props.ccProductHeight : props.productHeight,
);

const frameStyle = computed(() => ({
  width: omitsFrameWidth.value ? undefined : props.width,
  height: props.height,
  flex: omitsFrameWidth.value ? undefined : `0 0 ${props.width}`,
  gap: profile === CC_DISTRIBUTION_ASSETS_PROFILE && props.composition === "combined" ? "6px" : props.gap,
}));
const vendorViewportStyle = computed(() => ({
  width: profile === CC_DISTRIBUTION_ASSETS_PROFILE ? "auto" : props.vendorWidth,
  height:
    profile === CC_DISTRIBUTION_ASSETS_PROFILE
      ? props.omitCcVendorViewportHeight
        ? undefined
        : props.ccVendorHeight || props.productHeight
      : props.vendorHeight,
  flex: profile === CC_DISTRIBUTION_ASSETS_PROFILE ? "0 0 auto" : `0 0 ${props.vendorWidth}`,
  "--cc-vendor-height": props.ccVendorHeight || props.productHeight,
}));
const vendorLogoFilter = computed(() => {
  return usesVendorWhiteTextFilter.value ? `url(#${vendorWhiteTextFilterId})` : undefined;
});
const vendorLogoHeight = computed(() => {
  if (usesCeHorizontalVendorGeometry.value) return props.vendorHeight;
  if (usesCcVendorGeometry.value) return props.ccVendorHeight || props.productHeight;
  return props.vendorCanvasHeight;
});
const vendorLogoTop = computed(() => {
  if (usesCeHorizontalVendorGeometry.value || usesCcVendorGeometry.value) return undefined;
  return props.vendorCanvasTop;
});
const vendorLogoCanvasStyle = computed(() => ({
  filter: vendorLogoFilter.value,
  height: vendorLogoHeight.value,
  top: vendorLogoTop.value,
}));
const productLogoStyle = computed(() => ({
  width: productWidth.value,
  height: productHeight.value,
  flex: `0 0 ${productWidth.value}`,
  filter: usesWhiteText.value ? "brightness(0) invert(1)" : undefined,
  transform:
    profile === "ce" && props.composition === "combined" && !props.disableProductOffset ? "translateY(1px)" : undefined,
}));

function handleVendorLogoError(event: Event): void {
  usesDarkSurfaceFallback.value =
    props.vendorAsset === DISTRIBUTION_LOGO_ASSETS.VendorLogoOnDark ||
    props.vendorAsset === DISTRIBUTION_LOGO_ASSETS.VendorMarkOnLight;
  if (profile !== "cc" && event.currentTarget instanceof HTMLImageElement) {
    usesCcVendorGeometry.value = true;
    usesCcVendorFallback.value = true;
  }
  fallBackDistributionLogo(event, props.vendorAsset);
}
</script>

<style scoped lang="scss">
.runai-distribution-logo {
  position: relative;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  line-height: 0;
  overflow: hidden;
}

.logo-color-filter {
  position: absolute;
  width: 0;
  height: 0;
  pointer-events: none;
}

.product-logo {
  display: block;
  width: auto;
  max-width: 100%;
  object-fit: contain;
}

.vendor-logo-viewport {
  position: relative;
  display: block;
  overflow: hidden;
}

.runai-distribution-logo[data-profile="ce"] .vendor-logo-viewport {
  overflow: visible;
}

// The verified NVIDIA SVG uses an unmodified 2560×1440 canvas. Its horizontal
// logo occupies y=523.54106 through y=916.459, so crop the transparent canvas
// in the fixed viewport without changing the downloaded bytes.
.vendor-logo-canvas {
  position: absolute;
  left: 50%;
  display: block;
  width: auto;
  max-width: none;
  transform: translateX(-50%);
}

.runai-distribution-logo[data-profile="cc"] .vendor-logo-canvas,
.vendor-logo-canvas.vendor-logo-horizontal-ce,
.vendor-logo-canvas.vendor-logo-fallback-cc {
  position: static;
  width: auto;
  height: var(--cc-vendor-height);
  margin-left: auto;
  transform: none;
}

.runai-distribution-logo[data-profile="cc"] .vendor-logo-viewport,
.runai-distribution-logo:has(.vendor-logo-horizontal-ce) .vendor-logo-viewport,
.runai-distribution-logo:has(.vendor-logo-fallback-cc) .vendor-logo-viewport {
  display: flex;
  align-items: center;
  justify-content: flex-end;
}

.runai-distribution-logo:has(.vendor-logo-fallback-cc) .vendor-logo-viewport {
  width: auto !important;
  height: var(--cc-vendor-height) !important;
  flex: 0 0 auto !important;
}

.runai-distribution-logo[data-profile="cc"][data-composition="vendor"] .vendor-logo-canvas {
  margin-right: auto;
}
</style>
