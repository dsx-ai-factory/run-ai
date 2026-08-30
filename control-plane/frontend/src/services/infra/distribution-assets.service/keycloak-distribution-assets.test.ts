import { existsSync, readFileSync } from "node:fs";
import vm from "node:vm";
import { describe, expect, it, vi } from "vitest";

const script = readFileSync("../keycloak-extensions/runai-theme/login/resources/js/distribution-assets.js", "utf8");
const styles = readFileSync("../keycloak-extensions/runai-theme/login/resources/css/nv-login.css", "utf8");

class FakeElement {
  className = "";
  children: FakeElement[] = [];
  attributes = new Map<string, string>();
  style = { backgroundImage: "" };
  parent: FakeElement | null = null;
  classList = {
    add: (className: string) => {
      this.className = `${this.className} ${className}`.trim();
    },
    contains: (className: string) => this.className.split(" ").includes(className),
    remove: (className: string) => {
      this.className = this.className
        .split(" ")
        .filter((value) => value && value !== className)
        .join(" ");
    },
  };

  appendChild(child: FakeElement) {
    child.parent = this;
    this.children.push(child);
  }

  querySelector(selector: string) {
    if (selector === ".distribution-assets-login-logo") {
      return this.children.find((child) => child.className === "distribution-assets-login-logo") || null;
    }
    return null;
  }

  remove() {
    if (!this.parent) return;
    this.parent.children = this.parent.children.filter((child) => child !== this);
    this.parent = null;
  }

  setAttribute(name: string, value: string) {
    this.attributes.set(name, value);
  }
}

class FakeLink extends FakeElement {
  href = "";
  rel = "";
  onerror: (() => void) | null = null;
}

class FakeImage extends FakeElement {
  static instances: FakeImage[] = [];
  alt = "";
  draggable = true;
  onload: (() => void) | null = null;
  onerror: (() => void) | null = null;
  src = "";

  constructor() {
    super();
    FakeImage.instances.push(this);
  }
}

function renderKeycloakLogo() {
  FakeImage.instances = [];
  const header = new FakeElement();
  const favicon = new FakeLink();
  const head = new FakeElement();
  let resolveFetch: (response: { ok: boolean; json: () => Promise<unknown> }) => void = () => {};
  let rejectFetch: (reason?: unknown) => void = () => {};
  let timeoutCallback: () => void = () => {};
  const setTimeoutMock = vi.fn((callback: () => void) => {
    timeoutCallback = callback;
    return 1;
  });
  const clearTimeoutMock = vi.fn();
  const fetch = vi.fn(
    () =>
      new Promise<{ ok: boolean; json: () => Promise<unknown> }>((resolve, reject) => {
        resolveFetch = resolve;
        rejectFetch = reject;
      }),
  );

  vm.runInNewContext(script, {
    Error,
    Image: FakeImage,
    Promise,
    document: {
      head,
      readyState: "complete",
      querySelector: (selector: string) => (selector.startsWith("header") ? header : favicon),
      createElement: (tag: string) => (tag === "link" ? new FakeLink() : new FakeElement()),
      createElementNS: () => new FakeElement(),
    },
    fetch,
    setTimeout: setTimeoutMock,
    clearTimeout: clearTimeoutMock,
  });

  return {
    clearTimeout: clearTimeoutMock,
    favicon,
    fetch,
    header,
    rejectRequest() {
      rejectFetch(new Error("Boot config request failed"));
    },
    resolveProfile(profile: string) {
      resolveFetch({
        ok: true,
        json: () => Promise.resolve({ distributionAssets: { profile } }),
      });
    },
    setTimeout: setTimeoutMock,
    triggerTimeout() {
      timeoutCallback();
    },
  };
}

async function flushPromises() {
  await new Promise((resolvePromise) => setTimeout(resolvePromise, 0));
}

describe("Keycloak distribution assets", () => {
  it("keeps an empty fixed logo slot until boot config resolves", () => {
    const { favicon, fetch, header, setTimeout } = renderKeycloakLogo();

    expect(fetch).toHaveBeenCalledOnce();
    expect(setTimeout).toHaveBeenCalledWith(expect.any(Function), 4500);
    expect(FakeImage.instances).toHaveLength(0);
    expect(favicon.href).toBe("");
    expect(header.children).toHaveLength(0);
    expect(styles).toMatch(/header\.login-pf-header\s*\{[\s\S]*?background-image:\s*none;/);
    expect(styles).not.toContain("../img/");
    expect(existsSync("../keycloak-extensions/runai-theme/login/resources/img/nvidia-login-logo.svg")).toBe(false);
    expect(existsSync("../keycloak-extensions/runai-theme/login/resources/img/favicon.ico")).toBe(false);
    expect(styles).toMatch(
      /distribution-assets-vendor-logo-viewport:has\(\.distribution-assets-vendor-logo-profile-cc\)\s*\{[\s\S]*?flex:\s*0 0 auto;/,
    );
  });

  it("loads the CC logo and favicon after boot config resolves", async () => {
    const { favicon, header, resolveProfile } = renderKeycloakLogo();

    resolveProfile("cc");
    await flushPromises();
    const [vendorLogo, productLogo] = FakeImage.instances;
    expect(vendorLogo.src).toBe("/distribution-assets/cc/vendor-logo-on-light.svg");
    expect(productLogo.src).toBe("/distribution-assets/product-logo.svg");
    expect(favicon.href).toBe("/distribution-assets/cc/favicon.ico");

    vendorLogo.onload?.();
    productLogo.onload?.();
    await flushPromises();

    expect(vendorLogo.className).toContain("distribution-assets-vendor-logo-profile-cc");
    expect(header.children).toHaveLength(1);
    expect(header.classList.contains("distribution-assets-profile-cc")).toBe(true);
    expect(header.style.backgroundImage).toBe("none");
  });

  it("loads a split CE logo and falls back once for the vendor image", async () => {
    const { favicon, header, resolveProfile } = renderKeycloakLogo();

    resolveProfile("ce");
    await flushPromises();
    const [vendorLogo, productLogo] = FakeImage.instances;
    expect(vendorLogo.src).toBe("/distribution-assets/ce/vendor-logo-on-light.svg");
    expect(productLogo.src).toBe("/distribution-assets/product-logo.svg");

    vendorLogo.onerror?.();
    expect(vendorLogo.src).toBe("/distribution-assets/cc/vendor-logo-on-light.svg");
    vendorLogo.onload?.();
    productLogo.onload?.();
    await flushPromises();

    expect(header.children).toHaveLength(1);
    expect(header.classList.contains("distribution-assets-profile-ce")).toBe(true);
    expect(vendorLogo.className).toContain("distribution-assets-vendor-logo-profile-cc");
    expect(header.style.backgroundImage).toBe("none");

    favicon.onerror?.();
    expect(favicon.href).toBe("/distribution-assets/cc/favicon.ico");
    expect(favicon.onerror).toBeNull();
  });

  it("loads the CC default immediately when boot config fails", async () => {
    const { favicon, header, rejectRequest } = renderKeycloakLogo();

    rejectRequest();
    await flushPromises();
    const [vendorLogo, productLogo] = FakeImage.instances;

    expect(vendorLogo.src).toBe("/distribution-assets/cc/vendor-logo-on-light.svg");
    expect(productLogo.src).toBe("/distribution-assets/product-logo.svg");
    expect(favicon.href).toBe("/distribution-assets/cc/favicon.ico");

    vendorLogo.onload?.();
    productLogo.onload?.();
    await flushPromises();

    expect(header.children).toHaveLength(1);
    expect(header.classList.contains("distribution-assets-profile-cc")).toBe(true);
  });

  it("loads the CC default after the timeout and ignores a late response", async () => {
    const { favicon, resolveProfile, triggerTimeout } = renderKeycloakLogo();

    triggerTimeout();
    const [vendorLogo, productLogo] = FakeImage.instances;
    expect(vendorLogo.src).toBe("/distribution-assets/cc/vendor-logo-on-light.svg");
    expect(productLogo.src).toBe("/distribution-assets/product-logo.svg");
    expect(favicon.href).toBe("/distribution-assets/cc/favicon.ico");

    resolveProfile("ce");
    await flushPromises();

    expect(FakeImage.instances).toHaveLength(2);
    expect(favicon.href).toBe("/distribution-assets/cc/favicon.ico");
  });

  it("keeps the fixed logo slot empty if the product image fails", async () => {
    const { header, resolveProfile } = renderKeycloakLogo();

    resolveProfile("ce");
    await flushPromises();
    const [vendorLogo, productLogo] = FakeImage.instances;
    vendorLogo.onload?.();
    productLogo.onerror?.();
    await flushPromises();

    expect(header.children).toHaveLength(0);
    expect(header.className).toBe("");
  });
});
