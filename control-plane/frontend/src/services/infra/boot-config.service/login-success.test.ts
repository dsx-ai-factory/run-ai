/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { readFileSync } from "node:fs";
import vm from "node:vm";
import { describe, expect, it, vi } from "vitest";

const html = readFileSync("public/static-cli/login-success.html", "utf8");
const script = html.match(/<script>([\s\S]*?)<\/script>/)?.[1];

async function renderLoginSuccess(
  url?: string,
  profile = "cc",
  options: { deferFetch?: boolean; failedImages?: string[]; rejectFetch?: boolean } = {},
) {
  const supportLink = { href: "" };
  const supportFooter = { hidden: true };
  const loginSuccessContainer = { classList: { remove: vi.fn() } };
  const distributionLogo = { classList: { add: vi.fn() }, setAttribute: vi.fn() };
  const vendorLogo = {
    className: "vendor-logo-canvas",
    src: "",
  };
  const productLogo = { src: "" };
  const failedImages = new Set(options.failedImages);
  let resolveFetch: ((response: { ok: boolean; json: () => Promise<unknown> }) => void) | undefined;
  let timeoutCallback: (() => void) | undefined;

  class MockImage {
    onload: (() => void) | null = null;
    onerror: (() => void) | null = null;
    private source = "";

    set src(value: string) {
      this.source = value;
      queueMicrotask(() => {
        if (failedImages.has(value)) this.onerror?.();
        else this.onload?.();
      });
    }

    get src() {
      return this.source;
    }
  }

  const configResponse = {
    ok: true,
    json: () => Promise.resolve({ distributionAssets: { profile }, support: { url } }),
  };
  const fetch = options.rejectFetch
    ? vi.fn().mockRejectedValue(new Error("boot config unavailable"))
    : options.deferFetch
      ? vi.fn(
          () =>
            new Promise<{ ok: boolean; json: () => Promise<unknown> }>((resolve) => {
              resolveFetch = resolve;
            }),
        )
      : vi.fn().mockResolvedValue(configResponse);
  const setTimeoutMock = vi.fn((callback: () => void) => {
    timeoutCallback = callback;
    return 1;
  });

  vm.runInNewContext(script as string, {
    AbortController,
    Image: MockImage,
    URL,
    clearTimeout,
    document: {
      getElementById: vi.fn((id: string) => {
        if (id === "support-link") return supportLink;
        if (id === "support-footer") return supportFooter;
        if (id === "login-success-container") return loginSuccessContainer;
        if (id === "distribution-assets-logo") return distributionLogo;
        if (id === "vendor-logo") return vendorLogo;
        if (id === "product-logo") return productLogo;
        return null;
      }),
    },
    fetch,
    setTimeout: setTimeoutMock,
  });
  await new Promise((resolve) => setTimeout(resolve, 0));
  expect(fetch).toHaveBeenCalledOnce();

  return {
    supportLink,
    supportFooter,
    removedUnavailableClass: loginSuccessContainer.classList.remove.mock.calls,
    distributionLogo,
    resolveFetch() {
      resolveFetch?.(configResponse);
    },
    vendorLogo,
    productLogo,
    setTimeout: setTimeoutMock,
    triggerTimeout() {
      timeoutCallback?.();
    },
  };
}

describe("CLI login success", () => {
  it("shows support only from a timely response with a safe URL", async () => {
    for (const safeUrl of ["https://support.example.com/help", "/support/help"]) {
      await expect(renderLoginSuccess(safeUrl)).resolves.toMatchObject({
        supportLink: { href: safeUrl },
        supportFooter: { hidden: false },
        removedUnavailableClass: [["support-unavailable"]],
      });
    }

    for (const unsafeUrl of [
      undefined,
      "javascript:alert(1)",
      "data:text/html,test",
      "//evil.example.com",
      " https://support.example.com",
    ]) {
      await expect(renderLoginSuccess(unsafeUrl)).resolves.toMatchObject({
        supportLink: { href: "" },
        supportFooter: { hidden: true },
        removedUnavailableClass: [],
      });
    }

    const lateResponse = await renderLoginSuccess("https://support.example.com/help", "ce", {
      deferFetch: true,
    });
    lateResponse.triggerTimeout();
    lateResponse.resolveFetch();
    await new Promise((resolve) => setTimeout(resolve, 0));

    expect(lateResponse.supportLink.href).toBe("");
    expect(lateResponse.supportFooter.hidden).toBe(true);
    expect(lateResponse.removedUnavailableClass).toEqual([]);
  });

  it("keeps a fixed empty logo slot until boot config resolves", () => {
    expect(html).not.toContain('rel="preload"');
    expect(html).not.toContain('src="/distribution-assets/ce/vendor-logo-on-light.svg"');
    expect(html).toContain('id="distribution-assets-logo"');
    expect(html).toContain('aria-busy="true"');
    expect(html).toContain("const BOOT_CONFIG_TIMEOUT_MS = 4500");
  });

  it("renders the resolved profile after both images load", async () => {
    const ce = await renderLoginSuccess(undefined, "ce");

    expect(ce.vendorLogo.src).toBe("/distribution-assets/ce/vendor-logo-on-light.svg");
    expect(ce.vendorLogo.className).toContain("profile-ce");
    expect(ce.productLogo.src).toBe("/distribution-assets/product-logo.svg");
    expect(ce.distributionLogo.classList.add).toHaveBeenCalledWith("is-ready");
    expect(ce.distributionLogo.setAttribute).toHaveBeenCalledWith("aria-busy", "false");

    const unknown = await renderLoginSuccess(undefined, "custom");
    expect(unknown.vendorLogo.src).toBe("/distribution-assets/cc/vendor-logo-on-light.svg");
    expect(unknown.vendorLogo.className).toContain("profile-cc");
  });

  it("falls back to CC when config or the selected vendor is unavailable", async () => {
    const failedVendor = await renderLoginSuccess(undefined, "ce", {
      failedImages: ["/distribution-assets/ce/vendor-logo-on-light.svg"],
    });

    expect(failedVendor.vendorLogo.src).toBe("/distribution-assets/cc/vendor-logo-on-light.svg");
    expect(failedVendor.vendorLogo.className).toContain("profile-cc");

    const failedConfig = await renderLoginSuccess(undefined, "ce", { rejectFetch: true });
    expect(failedConfig.vendorLogo.src).toBe("/distribution-assets/cc/vendor-logo-on-light.svg");
    expect(failedConfig.vendorLogo.className).toContain("profile-cc");
  });

  it("keeps the complete logo hidden when the product image is unavailable", async () => {
    const { distributionLogo, productLogo, vendorLogo } = await renderLoginSuccess(undefined, "ce", {
      failedImages: ["/distribution-assets/product-logo.svg"],
    });

    expect(vendorLogo.src).toBe("");
    expect(productLogo.src).toBe("");
    expect(distributionLogo.classList.add).not.toHaveBeenCalled();
  });
});
