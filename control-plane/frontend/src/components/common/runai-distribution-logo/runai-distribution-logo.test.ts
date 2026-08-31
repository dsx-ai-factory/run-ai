/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { mount } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";

const distributionState = vi.hoisted(() => ({
  profile: "ce",
  fallBackDistributionLogo: vi.fn(),
}));

vi.mock("@/services/infra/distribution-assets.service", () => ({
  CC_DISTRIBUTION_ASSETS_PROFILE: "cc",
  CE_DISTRIBUTION_ASSETS_PROFILE: "ce",
  DISTRIBUTION_LOGO_ASSETS: {
    VendorLogoOnLight: "vendor-logo-on-light.svg",
    VendorLogoOnDark: "vendor-logo-on-dark.svg",
    VendorMarkOnLight: "vendor-mark-on-light.svg",
  },
  PRODUCT_LOGO_URL: "/distribution-assets/product-logo.svg",
  getDistributionAssetsProfile: () => distributionState.profile,
  getDistributionLogoUrl: (asset: string, profile: string) => `/distribution-assets/${profile}/${asset}`,
  fallBackDistributionLogo: distributionState.fallBackDistributionLogo,
}));

import RunaiDistributionLogo from "./runai-distribution-logo.vue";

describe("runai distribution logo", () => {
  beforeEach(() => {
    distributionState.profile = "ce";
    vi.clearAllMocks();
  });

  it("selects the CE split-logo assets", () => {
    const wrapper = mount(RunaiDistributionLogo, {
      props: {
        width: "156px",
        height: "17px",
      },
    });

    expect(wrapper.attributes("data-profile")).toBe("ce");
    expect(wrapper.find(".vendor-logo-canvas").attributes("src")).toBe(
      "/distribution-assets/ce/vendor-logo-on-light.svg",
    );
    expect(wrapper.find(".product-logo").attributes("src")).toBe("/distribution-assets/product-logo.svg");
  });

  it("renders only the product wordmark for product composition", () => {
    const wrapper = mount(RunaiDistributionLogo, {
      props: {
        composition: "product",
        width: "45px",
        height: "12px",
        gap: "0",
      },
    });

    expect(wrapper.find(".vendor-logo-canvas").exists()).toBe(false);
    expect(wrapper.find(".product-logo").exists()).toBe(true);
  });

  it("selects the CC split-logo assets", () => {
    distributionState.profile = "cc";
    const wrapper = mount(RunaiDistributionLogo, {
      props: {
        width: "200px",
        height: "23.4px",
        vendorWidth: "124px",
        vendorHeight: "23.4px",
        productWidth: "53px",
        productHeight: "14px",
        gap: "7px",
      },
    });

    expect(wrapper.attributes("data-profile")).toBe("cc");
    expect(wrapper.find(".vendor-logo-canvas").attributes("src")).toBe(
      "/distribution-assets/cc/vendor-logo-on-light.svg",
    );
    expect(wrapper.find(".product-logo").attributes("src")).toBe("/distribution-assets/product-logo.svg");
  });

  it("keeps the CC dark-surface wordmark unfiltered", () => {
    distributionState.profile = "cc";
    const wrapper = mount(RunaiDistributionLogo, {
      props: {
        composition: "vendor",
        width: "40px",
        height: "9px",
        vendorAsset: "vendor-logo-on-dark.svg",
        textColor: "white",
      },
    });

    const vendorLogo = wrapper.find(".vendor-logo-canvas");
    expect(vendorLogo.attributes("src")).toBe("/distribution-assets/cc/vendor-logo-on-dark.svg");
    expect(wrapper.find("feColorMatrix").exists()).toBe(false);
    expect(vendorLogo.attributes("style")).not.toContain("filter");
  });

  it("delegates CE image errors to the one-shot matching-asset fallback", async () => {
    const wrapper = mount(RunaiDistributionLogo, {
      props: {
        composition: "vendor",
        width: "124px",
        height: "23px",
        vendorWidth: "124px",
        vendorHeight: "23px",
      },
    });

    await wrapper.find(".vendor-logo-canvas").trigger("error");

    expect(distributionState.fallBackDistributionLogo).toHaveBeenCalledOnce();
    expect(distributionState.fallBackDistributionLogo).toHaveBeenCalledWith(
      expect.any(Event),
      "vendor-logo-on-light.svg",
    );
  });

  it("selects the CE dark-surface asset for white logo text", () => {
    const wrapper = mount(RunaiDistributionLogo, {
      props: {
        width: "200px",
        height: "19px",
        vendorWidth: "124px",
        vendorHeight: "19px",
        productWidth: "53px",
        productHeight: "14px",
        textColor: "white",
      },
    });

    expect(wrapper.attributes("data-text-color")).toBe("white");
    expect(wrapper.find("feColorMatrix").exists()).toBe(false);
    expect(wrapper.find(".vendor-logo-canvas").attributes("src")).toBe(
      "/distribution-assets/ce/vendor-logo-on-dark.svg",
    );
  });

  it("preserves fallback geometry when the compact CE navigation wordmark fails", async () => {
    const wrapper = mount(RunaiDistributionLogo, {
      props: {
        composition: "vendor",
        width: "40px",
        height: "32px",
        vendorWidth: "40px",
        vendorHeight: "9px",
        vendorCanvasHeight: "9px",
        vendorCanvasTop: "0",
        ccVendorHeight: "9px",
        vendorAsset: "vendor-logo-on-dark.svg",
        textColor: "white",
      },
    });

    const vendorLogo = wrapper.find(".vendor-logo-canvas");
    expect(vendorLogo.attributes("src")).toBe("/distribution-assets/ce/vendor-logo-on-dark.svg");
    expect(vendorLogo.classes()).toContain("vendor-logo-horizontal-ce");
    expect(vendorLogo.attributes("style")).not.toContain("filter");

    await vendorLogo.trigger("error");

    expect(vendorLogo.classes()).not.toContain("vendor-logo-horizontal-ce");
    expect(vendorLogo.classes()).toContain("vendor-logo-fallback-cc");
  });

  it("keeps the dark-surface fallback unfiltered when the CE navigation asset fails", async () => {
    const wrapper = mount(RunaiDistributionLogo, {
      props: {
        composition: "vendor",
        width: "40px",
        height: "32px",
        vendorWidth: "40px",
        vendorHeight: "32px",
        vendorCanvasHeight: "45px",
        vendorCanvasTop: "-6.5px",
        vendorAsset: "vendor-mark-on-light.svg",
        textColor: "white",
      },
    });

    const vendorLogo = wrapper.find(".vendor-logo-canvas");
    expect(vendorLogo.attributes("style")).toContain("filter");

    await vendorLogo.trigger("error");

    expect(vendorLogo.attributes("style")).not.toContain("filter");
    expect(vendorLogo.classes()).toContain("vendor-logo-fallback-cc");
  });

  it("selects the CE vertical navigation asset", () => {
    const wrapper = mount(RunaiDistributionLogo, {
      props: {
        composition: "vendor",
        width: "40px",
        height: "32px",
        vendorWidth: "40px",
        vendorHeight: "32px",
        vendorCanvasHeight: "45px",
        vendorCanvasTop: "-6.5px",
        vendorAsset: "vendor-mark-on-light.svg",
        textColor: "white",
      },
    });

    expect(wrapper.find(".vendor-logo-canvas").attributes("src")).toBe(
      "/distribution-assets/ce/vendor-mark-on-light.svg",
    );
  });
});
