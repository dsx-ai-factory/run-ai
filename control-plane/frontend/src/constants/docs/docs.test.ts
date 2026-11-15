/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

import { beforeEach, describe, expect, it, vi } from "vitest";

const bootConfigState = vi.hoisted(() => ({
  docsBaseUrl: "https://docs.example.test/product/2.26",
  catalogVersion: "v1" as "v1" | "v2" | undefined,
  profile: "cc" as "cc" | "ce",
}));

vi.mock("@/services/infra/boot-config.service/boot-config.service", () => ({
  getBootConfig: () => ({
    distributionAssets: { profile: bootConfigState.profile },
    version: { display: "release-42" },
    docs: {
      baseUrl: bootConfigState.docsBaseUrl,
      catalogVersion: bootConfigState.catalogVersion,
      whatsNew: { label: "Product updates", url: "https://updates.example.test/configured" },
    },
    support: { url: "", label: "Get help" },
    capabilities: { showsClusterLocationStep: true },
    unleashConfig: null,
  }),
}));

import { DOCS, getDocsLink, getDocsUrl, resolveDocsLinksInHtml, resolveDocsPath } from "./docs";
import { DOCS_CATALOG_V1, DOCS_CATALOG_V2, type DocsKey } from "./catalogs";

function resolvedAnchor(html: string): HTMLAnchorElement {
  const template = document.createElement("template");
  template.innerHTML = resolveDocsLinksInHtml(html);
  const anchor = template.content.querySelector("a");
  if (!anchor) throw new Error("Expected a documentation anchor");
  return anchor;
}

describe("documentation URL resolution", () => {
  beforeEach(() => {
    bootConfigState.docsBaseUrl = "https://docs.example.test/product/2.26";
    bootConfigState.catalogVersion = "v1";
    bootConfigState.profile = "cc";
  });

  it("joins paths to the configured base without duplicate slashes", () => {
    bootConfigState.docsBaseUrl += "/";
    expect(getDocsUrl("/reference/cli")).toBe("https://docs.example.test/product/2.26/reference/cli");
    expect(getDocsUrl("")).toBe("https://docs.example.test/product/2.26");
  });

  it("uses v1 when an older server omits the catalog", () => {
    bootConfigState.catalogVersion = undefined;
    expect(getDocsLink("authenticationOverview")).toBe(
      "https://docs.example.test/product/2.26/infrastructure-setup/authentication/overview",
    );
  });

  it.each(["cc", "ce"] as const)("selects v2 independently of asset profile %s", (profile) => {
    bootConfigState.profile = profile;
    bootConfigState.catalogVersion = "v2";
    expect(getDocsLink("authenticationOverview")).toBe(
      "https://docs.example.test/product/2.26/documentation/platform-management/authentication/overview",
    );
  });

  it.each(["https://mirror.example.test/content/v3.0/", "/internal-docs/"])(
    "uses the selected catalog with custom base %s",
    (baseUrl) => {
      bootConfigState.docsBaseUrl = baseUrl;
      bootConfigState.catalogVersion = "v2";
      expect(getDocsLink("cliInstallation")).toBe(`${baseUrl}reference/cli/install-and-configure-cli`);
    },
  );

  it("maps renamed anchors and historical release-note links explicitly", () => {
    bootConfigState.catalogVersion = "v2";
    expect(resolveDocsPath("fqdnRequirements")).toBe(
      "documentation/installation-and-configuration/system-requirements/networking-setup#fqdn-and-dns",
    );
    expect(resolveDocsPath("flexibleSubmission")).toBe(
      "documentation/workloads-in-nvidia-run-ai/using-workspaces/running-workspace#submission-form-options",
    );
  });

  it("preserves the configured root and complete updates URL", () => {
    expect(DOCS.ROOT).toBe("https://docs.example.test/product/2.26");
    expect(DOCS.WHATS_NEW).toEqual({ URL: "https://updates.example.test/configured", LABEL: "Product updates" });
  });

  it("keeps both catalogs complete and their paths relative", () => {
    expect(Object.keys(DOCS_CATALOG_V2)).toEqual(Object.keys(DOCS_CATALOG_V1));
    for (const catalog of [DOCS_CATALOG_V1, DOCS_CATALOG_V2]) {
      for (const path of Object.values(catalog)) {
        if (path !== null) expect(path).not.toMatch(/^(?:\/|[a-z]+:)/i);
      }
    }
  });

  it.each(["networkRequirements", "customInstallation"] as const)(
    "leaves unconfirmed v2 destination %s unavailable",
    (key) => {
      expect(getDocsLink(key)).toBeTruthy();
      bootConfigState.catalogVersion = "v2";
      expect(getDocsLink(key)).toBeUndefined();
    },
  );
});

describe("embedded documentation references", () => {
  beforeEach(() => {
    bootConfigState.docsBaseUrl = "/docs/v3.0";
    bootConfigState.catalogVersion = "v2";
  });

  it.each(Object.keys(DOCS_CATALOG_V2) as DocsKey[])("resolves the shared key %s", (key) => {
    const anchor = resolvedAnchor(`<a data-doc-key='${key}' target='_blank'><strong>Guide</strong></a>`);
    expect(anchor.getAttribute("href")).toBe(getDocsLink(key) ?? null);
    expect(anchor.hasAttribute("data-doc-key")).toBe(false);
    expect(anchor.getAttribute("target")).toBe("_blank");
    expect(anchor.innerHTML).toBe("<strong>Guide</strong>");
  });

  it("escapes attribute characters in the configured base", () => {
    bootConfigState.docsBaseUrl = '/docs/"quoted"';
    const anchor = resolvedAnchor('<a data-doc-key="projects">Guide</a>');
    expect(anchor.getAttribute("href")).toBe(getDocsLink("projects"));
    expect(anchor.attributes).toHaveLength(1);
  });

  it.each(["missing", "constructor", "__proto__", "networkRequirements"])(
    "removes stale href values for unavailable key %s",
    (key) => {
      const anchor = resolvedAnchor(`<a href="/stale" data-doc-key="${key}">Guide</a>`);
      expect(anchor.hasAttribute("href")).toBe(false);
      expect(anchor.textContent).toBe("Guide");
    },
  );

  it("leaves unrelated links unchanged", () => {
    const html = '<a href="https://example.test/guide?q=1#section">External guide</a>';
    expect(resolveDocsLinksInHtml(html)).toBe(html);
  });
});
